package dbtool

import (
	gp "code.google.com/p/goprotobuf/proto"
	"common"
	"errors"
	"logger"
	"proto"
	"rpc"
	"rpcplusclientpool"
	"strconv"
	"strings"
)

var FailedError error = errors.New("operate failed, please try again !")
var pPollBase *rpcplusclientpool.ClientPool
var pPollExtern *rpcplusclientpool.ClientPool
var pPoll *rpcplusclientpool.ClientPool

func InitDb() {
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

	aServerHost = common.ReadServerClientConfig("accountserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "accountserver")
	if pPoll == nil {
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

		hosts = common.ReadServerClientConfig("accountserver")
		if len(hosts) == 0 {
			return
		}
		pPoll.AdaptConnect(hosts)
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

func genPartnerId(partnerid string, pf rpc.Login_Platform) string {
	return partnerid + ":" + strconv.Itoa(int(pf))
}

func parsePartnerId(fullid string) (string, rpc.Login_Platform, error) {
	index := strings.LastIndex(fullid, ":")
	if index <= 0 {
		return "", 0, errors.New("wrong format !")
	}

	i, err := strconv.Atoi(fullid[index+1:])
	if err != nil {
		return "", 0, err
	}

	return fullid[:index], rpc.Login_Platform(i), nil
}

//查询关联的玩家id
func QueryPlayerIdByPartnerId(partnername, partnerid string, pf rpc.Login_Platform) (playerid string, err error) {
	////直接传入了uid，特殊用途
	//if common.CheckUUID(partnerid) {
	//	return partnerid, nil
	//}

	try := &proto.AccountDbQuery{
		Table: partnername,
		Key:   genPartnerId(partnerid, pf),
	}
	rst := &proto.AccountDbQueryResult{}

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return
	}

	if err = conn.Call("AccountServer.Query", try, rst); err != nil {
		return
	}

	//查询要关心存不存在
	if rst.Code == proto.NoExist {
		playerid = ""
		return
	}

	if rst.Code != proto.Ok {
		err = FailedError
		return
	}

	playerid = rst.Value
	err = nil

	return
}
