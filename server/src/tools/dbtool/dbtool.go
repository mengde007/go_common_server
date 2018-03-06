package main

import (
	"common"
	db "dbtool"
	"flag"
	"fmt"
	"io/ioutil"
	"logger"
	"net/http"
	"time"
	//"os"
	//"proto"
)

/*var (
		dbConfigFileOld = flag.String("o", "", "old config file name for dbserver")
		dbConfigFileNew = flag.String("n", "", "new config file name for dbserver")
    )

var dbServer *db.DBServer
var dbServerNew *db.DBServer

func main() {
	logger.Info("dbtoolserver start...")
	flag.Parse()

	var dbcfg common.DBConfig
	var dbcfgNew common.DBConfig
	if err := common.ReadDbConfig(*dbConfigFileOld, &dbcfg); err != nil {
		logger.Fatal("load old config failed,error is:%v",err)
		return
	}

	if err := common.ReadDbConfig(*dbConfigFileNew, &dbcfgNew); err != nil {
		logger.Fatal("load new config failed,error is:%v",err)
		return
	}

	dbServer = db.NewDBServerTool(dbcfg)
	dbServerNew = db.NewDBServerTool(dbcfgNew)

	count := 1
	//move
	for key,_ := range dbcfg.Tables {
		query := proto.DBQuery{key, "1"}
		//负责移动数据库的数据
		err := dbServer.MoveData(dbServerNew, &query)
    	if err != nil {
	    	logger.Info("query err\n")
		}
		//添加hash_id列
//		err = dbServerNew.AddHashId(&query,3)
		fmt.Fprintf(os.Stdin, "move data of %d over!!\n" ,count)
		count += 1
	}
}*/

func getRespose(host string) string {
	client := &http.Client{}
	reqest, _ := http.NewRequest("GET", "http://"+host+"/reload", nil)

	reqest.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Set("Accept-Charset", "GBK,utf-8;q=0.7,*;q=0.3")
	reqest.Header.Set("Accept-Encoding", "gzip,deflate,sdch")
	reqest.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	reqest.Header.Set("Cache-Control", "max-age=0")
	reqest.Header.Set("Connection", "keep-alive")

	response, _ := client.Do(reqest)
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		bodystr := string(body)
		fmt.Println(bodystr)

		return bodystr
	} else {
		return "error"
	}
}

var (
	csvDir = flag.String("c", "config", "config dir")
)

var dbAccount *db.DBServer
var dbDest *db.DBServer

func main() {
	logger.Info("dbtoolserver start...")

	flag.Parse()

	var cfgAccount common.DBConfig
	var cfgDest common.DBConfig
	if err := common.ReadAccountConfig(*csvDir, &cfgAccount); err != nil {
		logger.Fatal("load config failed, error is: %v", err)
		return
	}

	if err := common.ReadAccountConfig("../cfg/dbtool.json", &cfgDest); err != nil {
		logger.Fatal("load dbtool cfg failed, error is: %v", err)
		return
	}

	common.DebugInit(cfgAccount.GcTime, cfgAccount.DebugHost, "dbtool", cfgAccount.CpuProfile)
	db.InitDb()

	time.Sleep(time.Millisecond * 5)

	dbAccount = db.NewDBServerTool(cfgAccount)
	dbDest = db.NewDBServerTool(cfgDest)
	err := dbAccount.ProccessData(dbDest)
	if err != nil {
		logger.Info("query err")
	}

	logger.Info("query end")
}
