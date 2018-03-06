package main

import (
	"common"
	"flag"
	"lockserver"
	"logger"
	"net"
	"os"
	"syscall"
)

var (
	csvDir = flag.String("c", "config", "config dir")
)

func main() {

	flag.Parse()

	var lsConfig common.LockServerCfg
	if err := common.ReadLockServerConfig(*csvDir, &lsConfig); err != nil {
		logger.Error("load config failed, error is: %v", err)
		return
	}

	common.DebugInit(lsConfig.GcTime, lsConfig.DebugHost, "lockserver", lsConfig.CpuProfile)

	quitChan := make(chan int)

	listener, err := net.Listen("tcp", lsConfig.LockHost)
	if err != nil {
		logger.Error("Listening to: %s %s", lsConfig.LockHost, " failed !!")
		return
	}
	logger.Info("Listening to: %s %s", lsConfig.LockHost, " Success !!")
	defer listener.Close()

	go lockserver.CreateServices(lsConfig, listener)

	handler := func(s os.Signal, arg interface{}) {
		logger.Info("logserver close, handle signal: %v", s)
		common.DebugEnd(lsConfig.CpuProfile)
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

	logger.Info("lockserver close")

}
