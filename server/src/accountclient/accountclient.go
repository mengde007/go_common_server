package accountclient

import (
	"common"
	"errors"
	"logger"
	"proto"
	// "rpc"
	"rpcplusclientpool"
	"strconv"
	"strings"
)

var pPoll *rpcplusclientpool.ClientPool

var FailedError error = errors.New("operate failed, please try again !")

func init() {
	aServerHost := common.ReadServerClientConfig("accountserver")
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
		hosts := common.ReadServerClientConfig("accountserver")
		if len(hosts) == 0 {
			return
		}
		pPoll.AdaptConnect(hosts)
	})
}

func genPartnerId(partnerid string, pf int) string {
	return partnerid + ":" + strconv.Itoa(pf)
}

func parsePartnerId(fullid string) (string, int, error) {
	index := strings.LastIndex(fullid, ":")
	if index <= 0 {
		return "", 0, errors.New("wrong format !")
	}

	i, err := strconv.Atoi(fullid[index+1:])
	if err != nil {
		return "", 0, err
	}

	return fullid[:index], i, nil
}

//关联
func SetPartnerIdToPlayerId(partnername, partnerid string, pf int, playerid string) error {
	try := &proto.AccountDbWrite{
		Table: partnername,
		Key:   genPartnerId(partnerid, pf),
		Value: playerid,
	}
	rst := &proto.AccountDbWriteResult{}

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	if err = conn.Call("AccountServer.Write", try, rst); err != nil {
		return err
	}

	if rst.Code != proto.Ok {
		return FailedError
	}

	return nil
}

//删除关联
func DelPartnerIdToPlayerId(partnername, partnerid string, pf int) error {
	try := &proto.AccountDbQuery{
		Table: partnername,
		Key:   genPartnerId(partnerid, pf),
	}
	rst := &proto.AccountDbQueryResult{}

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	if err = conn.Call("AccountServer.Delete", try, rst); err != nil {
		return nil
	}

	if rst.Code != proto.Ok {
		return FailedError
	}

	return nil
}

//查询关联的玩家id
func QueryPlayerIdByPartnerId(partnername, partnerid string) (playerid string, err error) {
	////直接传入了uid，特殊用途
	//if common.CheckUUID(partnerid) {
	//	return partnerid, nil
	//}

	try := &proto.AccountDbQuery{
		Table: partnername,
		Key:   genPartnerId(partnerid, 10),
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

//反向查询关联的玩家id
func QueryPartnerIdByPlayerId(partnername, playerid string) (partnerid string, pf int, err error) {
	try := &proto.AccountDbQuery{
		Table: partnername,
		Key:   playerid,
	}
	rst := &proto.AccountDbQueryResult{}

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return
	}

	if err = conn.Call("AccountServer.ReQuery", try, rst); err != nil {
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

	partnerid, pf, err = parsePartnerId(rst.Value)

	return
}
