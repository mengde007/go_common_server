package gmclient

import (
	"common"
	"logger"
	"proto"
	"rpc"
	"rpcplusclientpool"
)

var pPoll *rpcplusclientpool.ClientPool

func init() {
	aServerHost := common.ReadServerClientConfig("gmserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "gmserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	return
}

//关联uid与名字，first表示是否第一次，要删除老的关联
func SetOpenId2Name(openid string, namelast, name string) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	req := &proto.GmUpdateOpenId2Name{
		OpenId:   openid,
		Name:     name,
		NameLast: namelast,
	}
	rst := &proto.GmUpdateOpenId2NameRst{}

	if err = conn.Call("GmService.UpdateOpenId2Name", req, rst); err != nil {
		return err
	}

	return nil
}

//玩家发送跑马灯
func PlayerSendNotice(msg *rpc.ReqBroadCast) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.GmPlayerSend{
		Msg: bufb,
	}
	rst := &proto.GmUpdateOpenId2NameRst{}

	conn.Go("GmService.AddPlayerNotice", req, rst, nil)
	return nil
}
