package gateserver

import (
	"fmt"
	"logger"
	//	"math/rand"
	"net"
	"proto"
	"rpc"
	"rpcplus"
	//	"strconv"
	"common"
	"sync"
	"sync/atomic"
	"time"
	"timer"
)

type serverInfo struct {
	PlayerCount uint32
	ServerIp    string
}

type GateServices struct {
	l            sync.RWMutex
	m            map[uint32]*serverInfo
	stableServer string
	count        uint32
	t            *timer.Timer
}

var pGateServices *GateServices

func CreateGateServicesForCnserver(listener net.Listener) *GateServices {
	pGateServices = &GateServices{m: make(map[uint32]*serverInfo), count: 0}
	rpcServer := rpcplus.NewServer()

	rpcServer.Register(pGateServices)

	//rpcServer.HandleHTTP("/center/rpc", "/debug/rpcdebug/rpc")

	pGateServices.t = timer.NewTimer(time.Second)
	pGateServices.t.Start(
		func() {
			pGateServices.UpdateStableCns()
		},
	)

	var uConnId uint32 = 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gateserver StartServices %s", err.Error())
			break
		}

		uConnId++
		go func(uConnId uint32) {
			rpcServer.ServeConnWithContext(conn, uConnId)
			pGateServices.l.Lock()
			delete(pGateServices.m, uConnId)
			pGateServices.l.Unlock()
		}(uConnId)
	}

	return pGateServices
}

func (self *GateServices) UpdateStableCns() {
	playerCountMax := uint32(0xffffffff) //不会有哪个服务器更大吧
	temp := make(map[uint32]*serverInfo)

	self.l.RLock()
	temp = self.m
	self.l.RUnlock()

	stableServer := ""
	for _, v := range temp {
		if len(v.ServerIp) > 0 && v.PlayerCount < playerCountMax {
			playerCountMax = v.PlayerCount
			stableServer = v.ServerIp
		}
	}

	self.l.Lock()
	self.stableServer = stableServer
	self.l.Unlock()
}

func (self *GateServices) UpdateCnsPlayerCount(uConnId uint32, info *proto.SendCnsInfo, result *proto.SendCnsInfoResult) error {
	server := &serverInfo{uint32(info.PlayerCount), info.ServerIp}
	self.l.Lock()
	self.m[uConnId] = server
	self.l.Unlock()
	return nil
}

func (self *GateServices) getStableCns() (cnsIp string) {
	curCount := atomic.AddUint32(&self.count, 1)
	if curCount > 1000 { //不够严谨这个判断，误差可以接受
		atomic.StoreUint32(&self.count, 0)
		self.UpdateStableCns()
	}
	stableServer := ""
	self.l.RLock()
	stableServer = self.stableServer
	self.l.RUnlock()
	return stableServer
}

type GateServicesForClient struct {
	m           string
	VersionOld  uint32
	VersionNew  uint32
	DownloadUrl string
}

var gateServicesForClient *GateServicesForClient

func CreateGateServicesForClient(listener net.Listener, old uint32, now uint32, downUrl string) *GateServicesForClient {

	gateServicesForClient = &GateServicesForClient{
		VersionOld:  old,
		VersionNew:  now,
		DownloadUrl: downUrl,
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("gateserver StartServices %s", err.Error())
			break
		}
		go func() {
			gateServicesForClient.onConn(conn)
		}()
	}

	return gateServicesForClient
}

func (c *GateServicesForClient) onConn(conn net.Conn) {

	rep := rpc.LoginCnsInfo{}
	rep.SetVersionOld(c.VersionOld)
	rep.SetVersionNew(c.VersionNew)
	rep.SetDownloadUrl(c.DownloadUrl)
	cnsIp := pGateServices.getStableCns()
	rep.CnsIp = &cnsIp

	gasinfo := fmt.Sprintf("%s;%d", conn.RemoteAddr(), time.Now().Unix())
	// encode
	encodeInfo := common.Base64Encode([]byte(gasinfo))
	gasinfo = fmt.Sprintf("%s;%s", gasinfo, encodeInfo)

	rep.GsInfo = &gasinfo

	common.SimpleWriteResult(conn, &rep)
	logger.Info("Client(%s) -> CnServer(%s)", conn.RemoteAddr(), cnsIp)

	time.Sleep(10 * time.Second)
	conn.Close()
}
