package main

import (
	db "accountserver"
	"common"
	"flag"
	"logger"
	"net"
	_ "net/http/pprof"
)

var (
	csvDir = flag.String("c", "config", "config dir")
)

var dbServer *db.AccountServer

func main() {
	logger.Info("accountsserver start")

	flag.Parse()

	var dbcfg common.DBConfig
	if err := common.ReadAccountConfig(*csvDir, &dbcfg); err != nil {
		logger.Fatal("load config failed, error is: %v", err)
		return
	}

	common.DebugInit(dbcfg.GcTime, dbcfg.DebugHost, "accountserver", dbcfg.CpuProfile)

	dbServer = db.NewAccountServer(dbcfg)

	tsock, err := net.Listen("tcp", dbcfg.DBHost)
	if err != nil {
		logger.Fatal("net.Listen: %s", err.Error())
	}

	go db.StartServices(dbServer, tsock)

	db.WaitForExit(dbServer)

	tsock.Close()

	logger.Info("accountsserver end")

	common.DebugEnd(dbcfg.CpuProfile)
}
