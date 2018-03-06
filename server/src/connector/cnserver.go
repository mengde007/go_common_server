package connector

import (
	"centerclient"
	"common"
	"csvcfg"
	// "errors"
	// "jfclient"
	"lockclient"
	"logger"
	"math/rand"
	"net"
	"os"
	"path"
	"proto"
	"rpc"
	"rpcplus"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"timer"
)

var cns *CNServer

var gTaskCfg map[string]*[]TaskCfg
var gUplevelCfg map[uint32]*[]UplevelCfg
var gItemCfg map[string]*[]ItemCfg

var Cfg common.CnsConfig

type CenterService struct {
}

type CNServer struct {
	serverForClient *rpc.Server
	// FsMgr             FServerConnMgr
	players           []*stPlayerConnIdSeg
	otherplayers      []*stPlayerConnIdSeg
	playersbyid       []*stPlayerUidSeg
	centerService     *CenterService
	exit              chan bool
	id                uint8
	listenIp          string
	titleInfo         *common.GlobalInfo
	listener          net.Listener
	maxPlayerCount    int32
	curPlayerCount    int32
	version           *common.VersionCfg
	profile           *common.GasProfileCfg
	versionCfgTick    *timer.Timer
	gamestateTick     *timer.Timer
	limitlogin        *common.LimitLoginConfig
	limitloginCfgTick *timer.Timer
	androidplayer     int32
	iosplayer         int32
	chatLock          sync.RWMutex
	chatMsgs          []*stChatMsg
	quitTick          *timer.Timer
	// heroLottery       map[string]*stHeroLottery
	lotteryLock    sync.RWMutex
	rankMgr        *RankMgr
	lastUpdateTime int
}

func (self *CNServer) GetServerIdStr() string {
	return strconv.Itoa(int(self.GetServerId()))
}

func (self *CNServer) GetServerId() uint8 {
	return self.id
}

func LoadConfigFiles(cfgDir string) {
	rand.Seed(time.Now().UnixNano())

	task := path.Join(cfgDir, "任务.csv")
	csvcfg.LoadCSVConfig(task, &gTaskCfg)

	uplevel := path.Join(cfgDir, "玩家经验表.csv")
	csvcfg.LoadCSVConfig(uplevel, &gUplevelCfg)

	item := path.Join(cfgDir, "道具.csv")
	csvcfg.LoadCSVConfig(item, &gItemCfg)

	common.LoadDaerGlobalConfig()
	common.LoadCustomRoomConfig()
}

func (self *CNServer) Quit() {
	self.quitTick = timer.NewTimer(time.Minute * 3) // 3分钟强制关服，如果有人卡住的话
	self.quitTick.Start(self.ForceEndService)

	self.listener.Close()
	self.serverForClient.Quit()
	//logger.Error("fs quit begin!!!!!!")
	//self.FsMgr.Quit()
}

func (self *CNServer) ForceEndService() {
	logger.Error("ForceEndService!!!!!!")
	time.Sleep(time.Millisecond * 5)
	os.Exit(1)
}

func (self *CNServer) EndService() {
}

func (self *CNServer) ReloadCfg() {
	// tempversion := &common.VersionCfg{}
	// common.ReadVersionConfig(tempversion)
	// self.version = tempversion

	tempprofile := &common.GasProfileCfg{}
	common.ReadProfileConfig(tempprofile)
	self.profile = tempprofile

	if self.serverForClient != nil {
		if 1 == self.profile.RpcProfile {
			self.serverForClient.OpenProfile()
		} else {
			self.serverForClient.CloseProfile()
		}
	}

}

func (self *CNServer) LogGameServerState() {
	// _, zoneId := common.GetPayUrlAndZoneId(false)
	// jfclient.LogGameServerState(getGetQQAppIdStr(), uint64(time.Now().Unix()/60),
	// 	uint32(self.GetServerId()), uint32(zoneId), uint32(self.iosplayer), uint32(self.androidplayer))
}

func (self *CNServer) LoadLimitLoginCfg() {
	templist := &common.LimitLoginConfig{}
	common.ReadLimitLoginConfig(templist)
	self.limitlogin = templist
}

func (self *CNServer) StartClientService(cfg *common.CnsConfig, wg *sync.WaitGroup) {
	self.ReloadCfg()
	self.versionCfgTick = timer.NewTimer(time.Second * 5)
	self.versionCfgTick.Start(self.ReloadCfg)

	self.gamestateTick = timer.NewTimer(time.Minute)
	self.gamestateTick.Start(self.LogGameServerState)

	self.LoadLimitLoginCfg()
	self.limitloginCfgTick = timer.NewTimer(time.Minute)
	self.limitloginCfgTick.Start(self.LoadLimitLoginCfg)

	rpcServer := rpc.NewServer()
	self.serverForClient = rpcServer

	rpcServer.Register(cns)
	rpcServer.RegCallBackOnConn(
		func(conn rpc.RpcConn) {
			self.onConn(conn)
		},
	)

	rpcServer.RegCallBackOnDisConn(
		func(conn rpc.RpcConn) {
			self.onDisConn(conn)
		},
	)

	rpcServer.RegCallBackOnCallBefore(
		func(conn rpc.RpcConn) {
			conn.Lock()
		},
	)

	rpcServer.RegCallBackOnCallAfter(
		func(conn rpc.RpcConn) {
			conn.Unlock()
		},
	)

	//开始对fightserver的RPC服务
	// self.FsMgr.Init(rpcServer, cfg)
	listener, err := net.Listen("tcp", cfg.CnsHost)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	self.listener = listener
	self.listenIp = cfg.CnsHostForClient

	//self.openPlayerCountToGate()

	wg.Add(1) //监听client要算一个
	go func() {
		for {
			//For Client/////////////////////////////
			time.Sleep(time.Millisecond * 5)
			conn, err := self.listener.Accept()

			if err != nil {
				logger.Error("cns StartServices %s", err.Error())
				wg.Done() // 退出监听就要减去一个
				break
			}

			newCount := atomic.AddInt32(&self.curPlayerCount, 1)
			if newCount > self.maxPlayerCount {
				atomic.AddInt32(&self.curPlayerCount, -1)

				lr := &rpc.LoginResult{}
				lr.SetResult("max player ! wail moment")
				lr.SetServerTime(time.Now().UnixNano() / 1e6)
				common.SimpleWriteResult(conn, lr)

				conn.Close()
				continue
			}

			wg.Add(1) // 这里是给客户端增加计数
			go func() {
				rpcConn := rpc.NewProtoBufConn(rpcServer, conn, 192, 150) //
				defer func() {
					if r := recover(); r != nil {
						logger.Error("player rpc runtime error begin:", r)

						rpcConn.Unlock()
						logger.Error("player rpc runtime error rpcConn.Unlock")
						debug.PrintStack()
						logger.Error("player rpc runtime error printstack")
						self.onDisConn(rpcConn)
						logger.Error("player rpc runtime error onDisConn")
						rpcConn.Close()

						logger.Error("player rpc runtime error end ")
					}
					atomic.AddInt32(&self.curPlayerCount, -1)
					wg.Done() //todo 客户端退出减去计数，放到延迟下线之后
				}()

				rpcServer.ServeConn(rpcConn)
			}()
		}
	}()
}

func StartCenterService(self *CNServer, listener net.Listener, cfg *common.CnsConfig) {
	//连接center
	rpcCenterServer := rpcplus.NewServer()
	rpcCenterServer.Register(self.centerService)

	req := &proto.CenterConnCns{
		ServerId: self.GetServerId(),
		Addr:     listener.Addr().String(),
	}
	rst := &proto.CenterConnCnsResult{}
	centerclient.Go("Center.CenterConnCns", req, rst)

	connCenter, err := listener.Accept()
	if err != nil {
		logger.Fatal("StartCenterServices %s", err.Error())
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("StartCenterService runtime error: %s", r)
				debug.PrintStack()
			}
		}()
		rpcCenterServer.ServeConn(connCenter)
		connCenter.Close()
	}()

}

func NewCNServer(cfg *common.CnsConfig) (server *CNServer) {
	server = &CNServer{
		players:       make([]*stPlayerConnIdSeg, PLAYERS_SEG),
		otherplayers:  make([]*stPlayerConnIdSeg, PLAYERS_SEG),
		playersbyid:   make([]*stPlayerUidSeg, PLAYERS_SEG),
		centerService: &CenterService{},
		// rankMgr:        CreateRankMgr(),
		id:             cfg.ServerId,
		curPlayerCount: 0,
		maxPlayerCount: cfg.MaxPlayerCount,
		chatMsgs:       make([]*stChatMsg, 0, CHAT_MSG_NUM),
		// heroLottery:    make(map[string]*stHeroLottery, 0)
		rankMgr: NewRankMgr(),
	}
	cns = server
	server.initSegs()
	// server.initHeroLottery()

	LoadConfigFiles(common.GetDesignerDir())

	// //心跳
	// tm := timer.NewTimer(time.Minute)
	// tm.Start(func() {
	// 	TLogGameSvrState(cfg.CnsHostForClient)
	// })

	return
}

func (self *CNServer) onConn(conn rpc.RpcConn) {
	rep := rpc.LoginCnsInfo{}
	// curVersion := self.version // 因为配置表reload有用锁，所以采用local变量拷贝，防止被切换
	// rep.SetVersionOld(curVersion.VersionOld)
	// rep.SetVersionNew(curVersion.VersionNew)
	// rep.SetVersionMid(curVersion.VersionMid)
	// rep.SetDownloadUrl(curVersion.DownloadUrl)
	temp := ""
	rep.CnsIp = &temp
	rep.GsInfo = &temp

	WriteResult(conn, &rep)
}

//login途中出错的disconnect流程，别乱调用
func (self *CNServer) DisConnOnLogin(conn rpc.RpcConn, uid string, lid uint64, gl int32, trophy uint32, bOnlined, bSendMsg bool) {
	lockclient.TryUnlock(common.LockName_Player, uid, lid)
	//玩家上线从离线表离拿走，不能再被匹配出来战斗
	// if bOnlined {
	// 	NotifyOffline(uid, false, int(gl), trophy, lid, true)
	// }

	if bSendMsg {
		// WriteLoginResult(conn, rpc.LoginResult_SERVERERROR)
	}
}

//正常下线流程出严重问题时才走此流程，禁止手动调用
func (self *CNServer) forceOnDisConn(connId uint64) {
	if p, ok := self.getPlayerByConnId(connId); ok {
		p.Unlock()

		if p.t != nil {
			p.t.Stop()
		}

		// NotifyOffline(p.GetUid(), p.GetIsUserguideFinish(), int(p.GetGamelocation()), p.GetTrophy(), p.lid, true)

		self.delMapPlayer(connId)
		self.delPlayerByUid(p.GetUid())
	}

	if op, ok := self.getOtherPlayerByConnId(connId); ok {
		op.Unlock()

		// NotifyOffline(op.GetUid(), op.GetIsUserguideFinish(), int(op.GetGamelocation()), op.GetTrophy(), op.lid, false)

		self.delMapOtherPlayer(connId)
		self.delPlayerByUid(op.GetUid())
	}
}

func (self *CNServer) ondisconn(conn rpc.RpcConn) {
	connId := conn.GetId()

	bCloseServer := self.serverForClient.IsClose()

	func() {
		conn.Lock()
		defer func() {
			conn.Unlock()
			if r := recover(); r != nil {
				debug.PrintStack()
				if p, exist := self.getPlayerByConnId(connId); exist {
					p.LogError("forceOnDisConn begin", connId, p.GetUid(), r)
					self.forceOnDisConn(connId)
					p.LogError("forceOnDisConn end", connId, p.GetUid(), r)
				} else {
					uid := ""
					logger.Error("else forceOnDisConn begin", connId, uid, r)
					self.forceOnDisConn(connId)
					logger.Error("else forceOnDisConn end", connId, uid, r)
				}
			}
		}()

		//如果是关服就不向fightserver发送了，直接退兵
		if !bCloseServer {

		}

		//正常流程走这里
		self.delPlayer(connId)
		self.delOtherPlayer(connId)
	}()
}

func (self *CNServer) onDisConn(conn rpc.RpcConn) {
	logger.Info("+CNServer:onDisConn connId = %d", conn.GetId())
	self.ondisconn(conn)
	logger.Info("-CNServer:onDisConn connId = %d", conn.GetId())
}

func (self *CNServer) Ping(conn rpc.RpcConn, login rpc.Ping) error {
	// if p, exist := self.getPlayerByConnId(conn.GetId()); exist {
	// p.UpdateClientTime(login.GetClientTime())
	// }

	rep := rpc.PingResult{}
	rep.SetServerTime(int32(time.Now().Unix()))

	WriteResult(conn, &rep)
	return nil
}
