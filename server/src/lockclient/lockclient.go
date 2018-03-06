package lockclient

import (
	"common"
	"logger"
	"proto"
	"rpcplusclientpool"
	"runtime/debug"
	"sync"
	"time"
	"timer"
)

type FailedCallback func()

type LockClient struct {
	serviceName     string
	key             string
	tickRenew       *timer.Timer
	lockInfo        *proto.TryGetLock
	callback        FailedCallback
	lastSuccessTime uint32
}
type OneLockService struct {
	mapKeys map[string]*LockClient
	l       sync.RWMutex
}

var mapLockServices map[string]*OneLockService

var pPoll *rpcplusclientpool.ClientPool

//初始化加锁客户端
func init() {
	aServerHost := common.ReadServerClientConfig("lockserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "lockserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	mapLockServices = make(map[string]*OneLockService)
	mapLockServices[common.LockName_Player] = &OneLockService{mapKeys: make(map[string]*LockClient)}
	mapLockServices[common.LockName_Donate] = &OneLockService{mapKeys: make(map[string]*LockClient)}
}

//尝试加锁
func TryLock(service, key string, lid uint64, validTime uint32, f FailedCallback) (result bool, old_value uint64, err error) {
	mService, ok := mapLockServices[service]
	if !ok {
		logger.Error("TryLock wrong service:%s", service)
		return
	}

	try := &proto.TryGetLock{Service: service, Name: key, Value: lid, ValidTime: validTime}
	rst := &proto.GetLockResult{}

	err, conn := pPoll.HashGetConn(key)
	if err != nil {
		return
	}

	err = conn.Call("LockServerServices.TryGetLock", try, rst)
	if err != nil {
		return
	}

	result = rst.Result
	old_value = rst.OldValue

	//成功则加入表
	if result {
		curTime := uint32(time.Now().Unix())

		pTimer := timer.NewTimer(time.Duration(validTime/2) * time.Second)
		pLockClient := &LockClient{
			serviceName:     service,
			key:             key,
			tickRenew:       pTimer,
			lockInfo:        try,
			lastSuccessTime: curTime,
			callback:        f,
		}
		pTimer.Start(func() {
			//自动续期
			timeRenew := uint32(time.Now().Unix())
			if renew(service, key, lid) {
				pLockClient.lastSuccessTime = timeRenew
			} else {
				//回调
				if timeRenew-pLockClient.lastSuccessTime > pLockClient.lockInfo.ValidTime {
					delLock(service, key, lid, false, true)
				}
			}
		})

		mService.l.Lock()
		mService.mapKeys[key] = pLockClient
		mService.l.Unlock()
	}

	return
}

//续期
func renew(service, key string, lid uint64) bool {
	err, conn := pPoll.HashGetConn(key)
	if err != nil {
		return false
	}

	renewReq := &proto.RenewLock{
		Service: service,
		Name:    key,
		Value:   lid,
	}
	renewRst := &proto.RenewLockResult{}

	err = conn.Call("LockServerServices.TryRenewLock", renewReq, renewRst)
	if err != nil {
		return false
	}

	return renewRst.Result
}

//删除锁
func delLock(service, key string, lid uint64, bforce, bcallback bool) {
	mService, ok := mapLockServices[service]
	if !ok {
		logger.Error("delLock wrong service:%s", service)
	}

	//回调
	var bacFunc FailedCallback = nil

	mService.l.Lock()
	if pLockClient, ok := mService.mapKeys[key]; ok && (bforce || pLockClient.lockInfo.Value == lid) {
		pLockClient.tickRenew.Stop()
		delete(mService.mapKeys, key)
		bacFunc = pLockClient.callback
	}
	mService.l.Unlock()

	//回调
	if bcallback && bacFunc != nil {
		bacFunc()
	}
}

//取消加锁
func TryUnlock(service, key string, lid uint64) (result bool, err error) {
	//删除
	delLock(service, key, lid, false, false)

	req := &proto.FreeLock{Service: service, Name: key, Value: lid}
	rst := &proto.FreeLockResult{}

	err, conn := pPoll.HashGetConn(key)
	if err != nil {
		return
	}

	err = conn.Call("LockServerServices.UnLock", req, rst)
	if err != nil {
		return
	}

	result = rst.Result

	return
}

//取消加锁
func ForceUnLock(service, key string) (result bool, err error) {
	//删除
	delLock(service, key, 0, true, false)

	req := &proto.ForceUnLock{Service: service, Name: key}
	rst := &proto.ForceUnLockResult{}

	err, conn := pPoll.HashGetConn(key)
	if err != nil {
		return
	}

	err = conn.Call("LockServerServices.ForceUnLock", req, rst)
	if err != nil {
		return
	}

	result = rst.Result

	return
}

func WaitLockGet(service, key string, lid uint64, validTime uint32, f FailedCallback) bool {
	timeLimit := 0

	for {
		successed, _, err := TryLock(service, key, lid, validTime, f)

		if err != nil {
			return false
		}

		if successed {
			return true
		}

		if timeLimit > 10 {
			return false
		}

		timeLimit += 1

		time.Sleep(time.Millisecond * 200)
	}

	panic("WaitLockGet unreachable")
}

//锁是否有效
func IsLockValid(service, key string, lid uint64) bool {
	mService, ok := mapLockServices[service]
	if !ok {
		logger.Error("IsLockValid: wrong service", service, key, lid)
		debug.PrintStack()
		return false
	}

	mService.l.RLock()
	pLockClient, ok := mService.mapKeys[key]
	mService.l.RUnlock()
	if !ok {
		logger.Error("IsLockValid: wrong key", service, key, lid)
		debug.PrintStack()
		return false
	}

	bOk := pLockClient.lockInfo.Value == lid
	if !bOk {
		logger.Error("IsLockValid: wrong lid", service, key, lid, pLockClient.lockInfo.Value)
	}

	return bOk
}

func queryPlayerGasId(key string) (uint8, uint8, bool) {
	_, ok := mapLockServices[common.LockName_Player]
	if !ok {
		logger.Error("TryLock wrong service:%s", common.LockName_Player)
		return 0, 0, false
	}

	try := &proto.QueryPlayer{Service: common.LockName_Player, Name: key}
	rst := &proto.QueryPlayerResult{}

	err, conn := pPoll.HashGetConn(key)
	if err != nil {
		return 0, 0, false
	}

	err = conn.Call("LockServerServices.QueryPlayer", try, rst)
	if err != nil {
		return 0, 0, false
	}

	if rst.Value == uint64(0) {
		return 0, 0, false
	}

	serverId, tid, _, _, _ := common.ParseLockMessage(rst.Value)

	return serverId, tid, true
}

//query player 信息
func QueryPlayerGasId(key string) (uint8, bool) {
	serverId, tid, ok := queryPlayerGasId(key)
	if ok && proto.MethodPlayerLogin == tid {
		return serverId, true
	}

	return 0, false
}

func IsOnline(uid string) bool {
	if _, _, ok := queryPlayerGasId(uid); ok {
		return true
	}

	return false
}
