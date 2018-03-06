package main

import (
	"common"
	"logger"
	"net"
	"os"
	"roleserver"
	"syscall"
)

func main() {
	var cfg common.RoleConfig
	if err := common.ReadRoleServerConfig(&cfg); err != nil {
		logger.Error("load config failed, error is: %v", err)
		return
	}

	common.DebugInit(cfg.GcTime, cfg.DebugHost, "roleserver", cfg.CpuProfile)

	quitChan := make(chan int)

	listener, err := net.Listen("tcp", cfg.Host)
	if err != nil {
		logger.Fatal("Listening to: %s failed !!", cfg.Host)
		return
	}
	logger.Info("Listening to: %s Success !!", cfg.Host)
	defer listener.Close()

	go roleserver.CreateServices(cfg, listener)

	handler := func(s os.Signal, arg interface{}) {
		logger.Info("handle signal: %v", s)
		logger.Info("logserver close")
		common.DebugEnd(cfg.CpuProfile)
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

	logger.Info("roleserver close")
}
