package main

import (
	"common"
	"gateserver"
	"logger"
	"net"
	"os"
	"syscall"
)

var ipcfg common.GateServerCfg

func main() {

	common.ReadGateServerConfig(&ipcfg)
	common.DebugInit(ipcfg.GcTime, ipcfg.DebugHost, "gateserver", ipcfg.CpuProfile)

	quitChan := make(chan int)

	listenerForClient, err := net.Listen("tcp", ipcfg.GsIpForClient)
	defer listenerForClient.Close()
	if err != nil {
		logger.Error("Listening to : %s %s", ipcfg.GsIpForClient, " failed !!")
		return
	}

	listenerForServer, err := net.Listen("tcp", ipcfg.GsIpForServer)
	defer listenerForServer.Close()
	if err != nil {
		logger.Error("Listening to : %s", listenerForServer.Addr().String())
		return
	}

	go gateserver.CreateGateServicesForCnserver(listenerForServer)
	go gateserver.CreateGateServicesForClient(listenerForClient, ipcfg.VersionOld, ipcfg.VersionNew, ipcfg.DownloadUrl)

	handler := func(s os.Signal, arg interface{}) {
		logger.Info("gateserver close handle signal : %v", s)
		common.DebugEnd(ipcfg.CpuProfile)
		os.Exit(0)
	}

	handlerArray := []os.Signal{syscall.SIGINT,
		syscall.SIGILL,
		syscall.SIGFPE,
		syscall.SIGSEGV,
		syscall.SIGTERM,
		syscall.SIGABRT}

	common.WatchSystemSignal(&handlerArray, handler)

	nQuitCount := 0
	for {
		select {
		case <-quitChan:
			nQuitCount = nQuitCount + 1
		}
		if nQuitCount == 2 {
			break
		}
	}
	logger.Info("gateserver close")

}
