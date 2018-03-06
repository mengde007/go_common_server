package main

import (
	"center"
	"common"
	"logger"
	"net"
	"os"
	"os/signal"
	"syscall"
)

import (
	"sync/atomic"
	"time"
)

var nid uint32 = 0

func GetUUID(sid int8, tid int8, value int8) uint64 {

	tmpid := int8(atomic.AddUint32(&nid, 1))

	return uint64(time.Now().Unix()) | uint64(tmpid)<<32 | uint64(value)<<40 | uint64(tid)<<48 | uint64(sid)<<56
}

var (
//laddr = flag.String("l", "127.0.0.1:8810", "The address to bind to.")
//dbg_addr = flag.String("d", "127.0.0.1:8811", "The address to bind to.(for debug)")
)

var centercfg common.CenterConfig

func main() {
	logger.Info("center start")
	//flag.Parse()

	if err := common.ReadCenterConfig(&centercfg); err != nil {
		return
	}

	common.DebugInit(centercfg.GcTime, centercfg.DebugHost, "center", centercfg.CpuProfile)

	centerServer := center.NewCenterServer(centercfg)

	tsock, err := net.Listen("tcp", centercfg.Host)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	//dbg_sock, err := net.Listen("tcp", *dbg_addr)

	//if err != nil {
	//	logger.Fatal("net.Listen: %s", err.Error())
	//}

	listenner, err := net.Listen("tcp", centercfg.HostForGm)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}
	go center.CreateCenterServiceForGM(listenner)

	signalChan := make(chan os.Signal, 1)
	exitChan := make(chan int)
	go func() {
		<-signalChan
		exitChan <- 1
	}()

	signal.Notify(signalChan, os.Interrupt)

	//go http.Serve(dbg_sock, nil)

	// 开启对支付宝充值的服务
	//center.StartTaobaoServices(centercfg.HostForTaobao)
	go center.StartServices(centerServer, tsock)

	handler := func(s os.Signal, arg interface{}) {
		logger.Info("center close handle signal : %v", s)
		common.DebugEnd(centercfg.CpuProfile)
		os.Exit(0)
	}

	handlerArray := []os.Signal{syscall.SIGINT,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGSEGV,
		syscall.SIGTERM,
		syscall.SIGABRT}

	common.WatchSystemSignal(&handlerArray, handler)

	<-exitChan

	tsock.Close()

	logger.Info("center end")
}
