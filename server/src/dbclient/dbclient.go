package dbclient

import (
	gp "code.google.com/p/goprotobuf/proto"
	"common"
	"logger"
	"rpcplusclientpool"
)

var pPollBase *rpcplusclientpool.ClientPool
var pPollExtern *rpcplusclientpool.ClientPool

func init() {
	//base
	aServerHost := common.ReadServerClientConfig("dbserverbase")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPollBase = rpcplusclientpool.CreateClientPool(aServerHost, "dbserverbase")
	if pPollBase == nil {
		logger.Fatal("create failed")
		return
	}

	//extern
	aServerHost = common.ReadServerClientConfig("dbserverextern")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPollExtern = rpcplusclientpool.CreateClientPool(aServerHost, "dbserverextern")
	if pPollExtern == nil {
		logger.Fatal("create failed")
		return
	}

	//定时读取
	common.RegisterReloadServerClientCfg(func() {
		hosts := common.ReadServerClientConfig("dbserverbase")
		if len(hosts) == 0 {
			return
		}
		pPollBase.AdaptConnect(hosts)

		hosts = common.ReadServerClientConfig("dbserverextern")
		if len(hosts) == 0 {
			return
		}
		pPollExtern.AdaptConnect(hosts)
	})
}

//基础信息库
func KVQueryBase(table, uid string, value gp.Message) (exist bool, err error) {
	err, conn := pPollBase.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVQuery(conn, table, uid, value)
}

func KVWriteBase(table, uid string, value gp.Message) (result bool, err error) {
	err, conn := pPollBase.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVWrite(conn, table, uid, value)
}

func KVDeleteBase(table, uid string) (exist bool, err error) {
	err, conn := pPollBase.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVDelete(conn, table, uid)
}

//额外信息库
func KVQueryExt(table, uid string, value gp.Message) (exist bool, err error) {
	err, conn := pPollExtern.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVQuery(conn, table, uid, value)
}

func KVWriteExt(table, uid string, value gp.Message) (result bool, err error) {
	err, conn := pPollExtern.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVWrite(conn, table, uid, value)
}

func KVDeleteExt(table, uid string) (exist bool, err error) {
	err, conn := pPollExtern.RandomGetConn()
	if err != nil {
		return
	}

	return common.KVDelete(conn, table, uid)
}
