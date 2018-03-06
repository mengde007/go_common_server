package center

import (
	"common"
	"lockclient"
	"logger"
	"net"
	"proto"
	//"pushmsg"
	"errors"
	// "rpc"
	"rpcplus"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"timer"
)

const TRACE = true

func ts(name string, items ...interface{}) {
	if TRACE {
		logger.Info("+%s %v\n", name, items)
	}
}
func te(name string, items ...interface{}) {
	if TRACE {
		logger.Info("-%s %v\n", name, items)
	}
}

var nid uint32 = 0

func GenLockMessage(sid uint8, tid uint8, value uint8) uint64 {

	tmpid := uint8(atomic.AddUint32(&nid, 1))

	return uint64(time.Now().Unix()) | uint64(tmpid)<<32 | uint64(value)<<40 | uint64(tid)<<48 | uint64(sid)<<56
}

var cnsConnId uint32 = 0

const (
	HASH_SEG = 10240
)

func getSeg(uid string) int {
	return int(common.MakeHash(uid) % uint32(HASH_SEG))
}

func (self *Center) getShieldSeg(uid string) *segShield {
	nSeg := getSeg(uid)
	return self.vecHashShield[nSeg]
}

func (self *Center) GetPlayerGasId(uid string) (uint8, bool) {
	return lockclient.QueryPlayerGasId(uid)
}

type segShield struct {
	l      sync.RWMutex
	shield map[string]*timer.Timer
}

func (self *segShield) AddShield(uid string, t *timer.Timer) {
	self.l.Lock()
	self.shield[uid] = t
	self.l.Unlock()
}

func (self *segShield) GetShield(uid string) (*timer.Timer, bool) {
	self.l.RLock()
	gasid, exist := self.shield[uid]
	self.l.RUnlock()
	return gasid, exist
}

func (self *segShield) DelShield(uid string) {
	self.l.Lock()
	delete(self.shield, uid)
	self.l.Unlock()
}

func (self *Center) AddShield(uid string, t *timer.Timer) {
	self.getShieldSeg(uid).AddShield(uid, t)
}

func (self *Center) DelShield(uid string) {
	self.getShieldSeg(uid).DelShield(uid)
}

func (self *Center) GetShield(uid string) (*timer.Timer, bool) {
	return self.getShieldSeg(uid).GetShield(uid)
}

type Center struct {
	l    sync.RWMutex
	cnss map[uint8]*rpcplus.Client
	//shields              map[string]*timer.Timer
	vecHashShield []*segShield
	maincache     *common.CachePool
	updatetime    *timer.Timer //add for update rankplayers
	everydaytime  *timer.Timer
	everyMinTick  *timer.Timer
	pCachePool    *common.CachePool
	onlineNumbers uint32 //当前在线人数
	//活动相关放这里了
	activityRank *StCenterActivityRank
	updateTick   *timer.Timer // 排行榜更新时间
}

var centerServer *Center

func StartServices(self *Center, listener net.Listener) {
	rpcServer := rpcplus.NewServer()

	rpcServer.Register(self)

	rpcServer.HandleHTTP("/center/rpc", "/debug/rpc")

	//add for save rankplayers
	// self.initUpdateRankPlayers()

	// 初始化排行榜
	self.initUpdateAllRank()

	//初始化周排行
	self.initDayTick()

	//创建一张哈希map
	// self.setValue("0", "0")

	//初始化推送tick
	self.initMinTick()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Center StartServices %s", err.Error())
			break
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("center runtime error: %s", r)
					debug.PrintStack()
				}
			}()

			logger.Info("Center: OnCns Connected")
			rpcServer.ServeConn(conn)

			logger.Info("Center: OnCns DisConnected")
			conn.Close()
		}()
	}
}

func NewCenterServer(cfg common.CenterConfig) (server *Center) {
	//加载配置表
	common.LoadGlobalConfig()
	// common.LoadWuShuangRobotConfig()

	server = &Center{
		cnss: make(map[uint8]*rpcplus.Client),
		//shields:              make(map[string]*timer.Timer),
		vecHashShield: make([]*segShield, HASH_SEG),
		pCachePool:    common.NewCachePool(cfg.Maincache),
	}

	for i := 0; i < HASH_SEG; i++ {
		server.vecHashShield[i] = &segShield{
			shield: make(map[string]*timer.Timer),
		}
	}

	//初始化cache
	logger.Info("Init Cache %v", cfg.Maincache)
	server.maincache = common.NewCachePool(cfg.Maincache)

	centerServer = server

	//在线统计
	server.regtickOnlineNum()

	//活动初始化
	server.initActivityRank()

	return server
}

func (self *Center) regtickOnlineNum() {
	timeNow := time.Now()
	d := time.Second * time.Duration(300-(timeNow.Minute()%5*60+timeNow.Second()))
	tm := timer.NewTimer(d)
	tm.Start(func() {
		//重新注册5分钟后的
		tm.Stop()
		self.regtickOnlineNum()

		numbers := uint32(0)

		req := &proto.GetOnlineNumber{}
		rst := &proto.GetOnlineNumberRst{}
		for _, conn := range self.cnss {
			if err := conn.Call("CenterService.GetOnlineNumber", req, rst); err == nil {
				numbers += rst.Numbers
			}
		}

		rst.Numbers = numbers
		self.onlineNumbers = numbers
		for _, conn := range self.cnss {
			if err := conn.Call("CenterService.LogOnlineNumber", rst, req); err == nil {
				break
			}
		}
	})
}

func (self *Center) GetOnlineNumbers(req *proto.ReqOnlinneNumber, reply *proto.RstOnlineNumber) (err error) {
	logger.Info("GetOnlineNumbers 人数:%d", self.onlineNumbers)
	reply.Number = self.onlineNumbers
	return nil
}

func (self *Center) CenterConnCns(req *proto.CenterConnCns, reply *proto.CenterConnCnsResult) (err error) {
	logger.Info("Center:CenterConnCns:%s", req.Addr)

	conn, err := net.Dial("tcp", req.Addr)
	if err != nil {
		logger.Error("Center Connect GameServer Failed :%s", err.Error())
		reply.Ret = false
		return
	}

	tmp := rpcplus.NewClient(conn)
	self.l.Lock()
	if _, ok := self.cnss[req.ServerId]; ok {
		logger.Fatal("the same serverid ", req.ServerId)
		return nil
	}
	self.cnss[req.ServerId] = tmp
	self.l.Unlock()

	self.theFirstPushRank(tmp)
	// self.pushActivityRank(tmp)

	reply.Ret = true

	//断线回调
	tmp.AddDisCallback(func(err error) {
		logger.Error("one gas disconnected", err)
		self.l.Lock()
		delete(self.cnss, req.ServerId)
		self.l.Unlock()
	})

	return nil
}

func (self *Center) KickCnsPlayer(req *proto.LoginKickPlayer, rst *proto.LoginKickPlayerResult) error {
	logger.Info("KickCnsPlayer: Begin!!!", req.Id)
	defer logger.Info("KickCnsPlayer: End!!!", req.Id)

	rst.Success = false

	if rpcc := self.getOnlineGas(req.Id); rpcc != nil {
		if err := rpcc.Call("CenterService.LoginKickPlayer", req, rst); err == nil && rst.Success {
			rst.Success = true
			return nil
		} else {
			return err
		}
	} else {
		return errors.New("getOnlineGas failed !")
	}
}

//取得玩家所在的gas
func (self *Center) getOnlineGas(uid string) *rpcplus.Client {
	serverid, ok := self.GetPlayerGasId(uid)
	if !ok || serverid == 0 {
		return nil
	}

	self.l.RLock()
	rpcc, ok := self.cnss[serverid]
	self.l.RUnlock()

	if ok {
		return rpcc
	}

	return nil
}

//设置最大在线人数
func (self *Center) SetMaxOnlineNumbers(req *proto.SetMaxOnlinePlayers, rst *proto.SetMaxOnlinePlayersRst) error {
	if req.ServerId == uint8(0) {
		for _, rpcc := range self.cnss {
			rpcc.Call("CenterService.SetMaxOnlineNumbers", req, rst)
		}
	} else {
		rpcc, ok := self.cnss[req.ServerId]
		if !ok {
			return errors.New("wrong ServerId")
		}

		rpcc.Call("CenterService.SetMaxOnlineNumbers", req, rst)
	}

	return nil
}

//重新加载表
func (self *Center) ReloadCfg(req *proto.GmCheckActivityConfig, rst *proto.GmCheckActivityConfigResult) error {
	logger.Info("ComeInto CenterGmServices.ReloadCfg")
	for _, conn := range self.cnss {
		conn.Call("CenterService.ReloadCfg", req, rst)
	}

	return nil
}

//取得活动配置表
func (self *Center) GetActivityConfig(req *proto.GetActivityConfig, rst *proto.GetActivityConfigRst) error {
	for _, conn := range self.cnss {
		if err := conn.Call("CenterService.GetActivityConfig", req, rst); err == nil {
			return nil
		}
	}

	return errors.New("no valid server")
}

//玩家支付通知
func (self *Center) NotifyPlayerGetPayInfo(req *proto.NotifyPlayerGetPayInfo, rst *proto.NotifyPlayerGetPayInfo) error {
	if rpcc := self.getOnlineGas(req.Uid); rpcc != nil {
		rpcc.Go("CenterService.NotifyPlayerGetPayInfo", req, rst, nil)
	} else {

	}
	return nil
}

func (self *Center) CallCnserverFunc(req *proto.CallCnserverMsg, rst *proto.CommonRst) error {
	for _, uid := range req.Uids {
		if rpcc := self.getOnlineGas(uid); rpcc != nil {
			rpcc.Go("CenterService.CallCnserverFunc", req, rst, nil)
		} else {

		}
	}

	return nil
}
