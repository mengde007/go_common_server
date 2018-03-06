package main

import (
	"common"
	"flag"
	"logger"
	"net"
	"os"
	"rankserver"
	"syscall"
)

var (
	cfgDir = flag.String("c", "config", "config dir")
)

func main() {
	logger.Info("power server start")
	flag.Parse()

	var cfg common.GeneralRankServerCfg
	if err := common.ReadGeneralRankServerCfg(&cfg); err != nil {
		logger.Error("load config failed, error is: %v", err)
		return
	}

	common.DebugInit(cfg.GcTime, cfg.DebugHost, "rankserver", cfg.CpuProfile)

	listener, err := net.Listen("tcp", cfg.GeneralRankHost)
	if err != nil {
		logger.Fatal("rankserver listen failed !!", cfg.GeneralRankHost, err)
		return
	}
	logger.Info("rankserver server listening to: %s Success !!", cfg.GeneralRankHost)
	defer listener.Close()

	go rankserver.CreateServices(cfg, listener)

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

	logger.Info("generalrankserver close")
}
