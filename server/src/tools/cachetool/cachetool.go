package main

import (
	"common"
	"flag"
	"logger"
	"redistool"
)

var (
	fileold1 = flag.String("o1", "config", "configdir")
	fileold2 = flag.String("o2", "config", "configdir")
	fileold3 = flag.String("o3", "config", "configdir")
	filenew  = flag.String("n", "config", "configdir")
)

func main() {
	// flag.Parse()
	// //	var cfgold1,cfgold2,cfgold3,cfgnew common.MatchServerConfig
	// cfg := make([]common.MatchServerConfig, 4, 4)

	// if err := common.ReadMatchServerConfig(*fileold1, &cfg[0]); err != nil {
	// 	logger.Error("load config failed,error is:%v", err)
	// 	return
	// }
	// if err := common.ReadMatchServerConfig(*fileold2, &cfg[1]); err != nil {
	// 	logger.Error("load config failed,error is:%v", err)
	// 	return
	// }
	// if err := common.ReadMatchServerConfig(*fileold3, &cfg[2]); err != nil {
	// 	logger.Error("load config failed,error is:%v", err)
	// 	return
	// }
	// if err := common.ReadMatchServerConfig(*filenew, &cfg[3]); err != nil {
	// 	logger.Error("load config failed,error is:%v", err)
	// 	return
	// }

	// redistool.MoveData(cfg)
}
