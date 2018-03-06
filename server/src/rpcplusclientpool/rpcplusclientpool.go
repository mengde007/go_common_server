package rpcplusclientpool

import (
	"common"
	"errors"
	"logger"
	"math/rand"
	"net"
	"rpcplus"
	"sync"
	"time"
)

var NoServiceError = errors.New("no service, please wait.")

type CALLBACK func(conn *rpcplus.Client)

//加锁服务器的连接
type ServerInfo struct {
	Shost   string
	Conn    *rpcplus.Client
	Breconn bool
}

type ClientPool struct {
	//服务器列表
	aServerList []*ServerInfo
	uServerNum  uint32
	//有效的服务器索引
	aValidServer []int
	l            sync.RWMutex
	//连接成功失败的回调
	okcallback  CALLBACK
	discallback CALLBACK
	//是否关闭
	bClose bool

	//服务名，log用
	serviceName string
}

//创建
func CreateClientPool(aServerHost []string, tag string) *ClientPool {
	pool := &ClientPool{
		aServerList:  make([]*ServerInfo, len(aServerHost)),
		uServerNum:   uint32(len(aServerHost)),
		aValidServer: make([]int, 0),
		bClose:       false,
		serviceName:  tag,
	}

	//由于各种服务器启动关联，修改为第一次启动连接不上也自动重连
	for i, v := range aServerHost {
		info := &ServerInfo{
			Shost:   v,
			Conn:    nil,
			Breconn: false,
		}

		logger.Info("CreateClientPool Server Info : %d, %s", i, v)

		//加入列表下面直接重新连接
		pool.aServerList[i] = info
		//if err := pool.addConnect(v, i); err != nil {
		//	logger.Fatal("dail lockserver failed", err)
		//	return nil
		//}
	}
	//直接重连
	for i, _ := range aServerHost {
		pool.reConnect(i)
	}

	return pool
}

//动态读取配置表后重新加入连接，只提供增加功能
func (self *ClientPool) AdaptConnect(aServerHost []string) {
	addHosts := make([]string, 0)

	bFind := false
	self.l.Lock()
	for _, host := range aServerHost {
		bFind = false
		for _, si := range self.aServerList {
			if host == si.Shost {
				bFind = true
				break
			}
		}
		if !bFind {
			addHosts = append(addHosts, host)
		}
	}

	beginIndex := len(self.aServerList)

	//添加
	for _, host := range addHosts {
		info := &ServerInfo{
			Shost:   host,
			Conn:    nil,
			Breconn: false,
		}

		self.aServerList = append(self.aServerList, info)
	}
	endIndex := len(self.aServerList)
	self.l.Unlock()

	//重新连接
	for i := beginIndex; i < endIndex; i++ {
		self.reConnect(i)
	}
}

//添加一个连接
func (self *ClientPool) addConnect(sHost string, iIndex int) error {
	if self.bClose {
		return nil
	}

	conn, err := net.Dial("tcp", sHost)
	if err != nil {
		logger.Error("ClientPool.addConnect dial failed! tag:%s, error:%v", self.serviceName, err)
		return err
	}

	self.l.Lock()
	defer self.l.Unlock()

	if self.bClose {
		conn.Close()
		return nil
	}

	rpc := rpcplus.NewClient(conn)
	rpc.AddDisCallback(func(err error) {
		logger.Info("disconnected error:", err)

		self.reConnect(iIndex)
	})

	info := &ServerInfo{
		Shost:   sHost,
		Conn:    rpc,
		Breconn: false,
	}

	if self.okcallback != nil {
		go self.okcallback(rpc)
	}

	self.aServerList[iIndex] = info
	self.aValidServer = append(self.aValidServer, iIndex)

	//logger.Info("aServerList numbers:", len(self.aValidServer))

	return nil
}

//重新连接
func (self *ClientPool) reConnect(iIndex int) {
	self.l.Lock()
	defer self.l.Unlock()

	if iIndex >= len(self.aServerList) {
		return
	}

	//重复调用的情况
	info := self.aServerList[iIndex]
	//过期的多次传入
	if info.Breconn {
		return
	}

	if self.discallback != nil {
		go self.discallback(info.Conn)
	}

	logger.Info("reconnect... %v", iIndex)

	info.Breconn = true
	//创建走重建流程连接信息为空
	if info.Conn != nil {
		info.Conn.Close()
	}
	for i, v := range self.aValidServer {
		if v == iIndex {
			self.aValidServer = append(self.aValidServer[:i], self.aValidServer[i+1:]...)
			break
		}
	}

	//重连接
	go func(sHost string) {
		for {
			if err := self.addConnect(sHost, iIndex); err == nil {
				break
			}

			time.Sleep(time.Second * 3)
		}
	}(info.Shost)
}

//随机取一个连接，后面根据负载来处理
func (self *ClientPool) RandomGetConn() (err error, conn *rpcplus.Client) {
	self.l.RLock()

	if len(self.aValidServer) == 0 {
		err = NoServiceError
		self.l.RUnlock()
		return
	}

	index := self.aValidServer[rand.Intn(len(self.aValidServer))]
	info := self.aServerList[index]
	if info.Breconn {
		err = NoServiceError
		self.l.RUnlock()
		return
	}
	conn = info.Conn

	self.l.RUnlock()
	return
}

//根据hash取得连接
func (self *ClientPool) HashGetConn(key string) (err error, conn *rpcplus.Client) {
	self.l.RLock()

	if self.uServerNum == 0 {
		self.l.RUnlock()
		return NoServiceError, nil
	}

	index := common.MakeHash(key) % self.uServerNum
	info := self.aServerList[index]
	if info.Breconn {
		self.l.RUnlock()
		return NoServiceError, nil
	}

	self.l.RUnlock()
	return nil, info.Conn
}

//取得所有的连接
func (self *ClientPool) GetAllConn() []*rpcplus.Client {
	connlist := make([]*rpcplus.Client, 0)

	self.l.RLock()
	for _, index := range self.aValidServer {
		connlist = append(connlist, self.aServerList[index].Conn)
	}

	self.l.RUnlock()
	return connlist
}

//连接成功的回调
func (self *ClientPool) SetConnectedCallback(f CALLBACK) {
	self.okcallback = f
}

//连接断开的回调
func (self *ClientPool) SetDisconnectCallback(f CALLBACK) {
	self.discallback = f
}

//关闭所有连接
func (self *ClientPool) CloseAll() {
	self.l.Lock()

	self.bClose = true
	for _, info := range self.aServerList {
		info.Conn.Close()
	}
	self.aValidServer = make([]int, 0)

	self.l.Unlock()
}
