package lockserver

import (
	"net"
	"rpcplus"
	//"time"
	"common"
	"logger"
	"proto"
	"runtime/debug"
)

type cacheGroup map[uint32]*common.CachePool

type LockServerServices struct {
	exit        chan bool
	cacheGroups map[string]cacheGroup
	cacheNodes  map[string][]uint32
	tables      map[string]*LockServer
}

var pLockServices *LockServerServices

func CreateServices(cfg common.LockServerCfg, listener net.Listener) *LockServerServices {

	pLockServices = &LockServerServices{}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pLockServices)

	//初始化所有的cache
	for key, pools := range cfg.CacheProfiles {

		logger.Info("Init Cache Profile %s", key)
		pLockServices.cacheGroups = make(map[string]cacheGroup)
		pLockServices.cacheNodes = make(map[string][]uint32)
		pLockServices.tables = make(map[string]*LockServer)

		temGroups := make(cacheGroup)
		temDbInt := []uint32{}
		for _, poolCfg := range pools {
			logger.Info("Init Cache %v", poolCfg)
			leng := poolCfg.NodeName
			temGroups[leng] = common.NewCachePool(poolCfg)
			temDbInt = append(temDbInt, leng)
		}

		pLockServices.cacheGroups[key] = temGroups
		common.BubbleSort(temDbInt) //排序节点
		pLockServices.cacheNodes[key] = temDbInt

	}

	//初始化table
	for key, table := range cfg.Tables {
		logger.Info("Init Table: %s %v", key, table)
		pLockServices.tables[key] = NewLockServer(key, table, pLockServices)
	}

	for {

		conn, err := listener.Accept()
		if err != nil {
			logger.Error("StartServices %s", err.Error())
			break
		}

		//开始对cns的RPC服务
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("lockServer Rpc Runtime Error: %s", r)
					debug.PrintStack()
				}
			}()
			rpcServer.ServeConn(conn)
		}()
	}

	return pLockServices
}

func (self *LockServerServices) TryGetLock(req *proto.TryGetLock, rep *proto.GetLockResult) (err error) {

	logger.Info("+++ LockServices TryGetLock", req)
	//defer logger.Info("--- LockServices TryGetLock", rep)
	if table, exist := self.tables[req.Service]; exist {
		rep.Result, rep.OldValue = table.GetLock(req.Name, req.Value, req.ValidTime)
		return nil
	}
	logger.Info("*********lockServices GetLockFailed, req.Service:%s", req.Service)
	rep.Result = false
	rep.OldValue = 0
	return nil
}

func (self *LockServerServices) UnLock(req *proto.FreeLock, rep *proto.FreeLockResult) (err error) {

	//logger.Info("+++ LockServices UnLock", req)
	//defer logger.Info("--- LockServices UnLock", rep)
	if table, exist := self.tables[req.Service]; exist {
		rep.Result = table.UnLock(req.Name, req.Value)
		return nil
	}

	rep.Result = false
	return nil
}

func (self *LockServerServices) ForceUnLock(req *proto.ForceUnLock, rep *proto.ForceUnLockResult) (err error) {

	//logger.Info("+++ LockServices ForceUnLock", req)
	//defer logger.Info("--- LockServices ForceUnLock", rep)
	if table, exist := self.tables[req.Service]; exist {
		rep.Result = table.ForceUnLock(req.Name)
		return nil
	}

	rep.Result = false
	return nil
}

func (self *LockServerServices) TryRenewLock(req *proto.RenewLock, rep *proto.RenewLockResult) (err error) {

	//logger.Info("+++ LockServices TryRenewLock", req)
	//defer logger.Info("--- LockServices TryRenewLock", rep)
	if table, exist := self.tables[req.Service]; exist {
		rep.Result = table.RenewLock(req.Name, req.Value)
		return nil
	}

	rep.Result = false
	return nil
}

func (self *LockServerServices) QueryPlayer(req *proto.QueryPlayer, rep *proto.QueryPlayerResult) (err error) {
	if table, exist := self.tables[req.Service]; exist {
		rep.Value = table.QueryPlayer(req.Name)
		return nil
	}
	rep.Value = 0
	return nil
}
