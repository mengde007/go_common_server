package matchclient

import (
	"common"
	"logger"
	"proto"
	"rpc"
	"rpcplusclientpool"
)

var pPoll *rpcplusclientpool.ClientPool

//初始化加锁客户端
func init() {
	aServerHost := common.ReadServerClientConfig("matchserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "matchserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}
	return
}

//请求房间列表
func MatchListREQ() (err error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	err = conn.Call("MatchServer.MatchListREQ", nil, nil)
	if err != nil {
		return err
	}
	return nil
}

//报名
func EnrollREQ(base *rpc.PlayerBaseInfo, client *rpc.EnrollREQ) (err error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(base)
	if err != nil {
		return err
	}

	bufc, err := common.EncodeMessage(client)
	if err != nil {
		return err
	}

	req := &proto.ReqDaerRoom{
		Base:   bufb,
		Client: bufc,
	}
	rst := &proto.OperRst{}

	err = conn.Call("MatchServer.EnrollREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//退赛
func WithdrawREQ(uid string, client *rpc.WithdrawREQ) (err error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufc, err := common.EncodeMessage(client)
	if err != nil {
		return err
	}

	req := &proto.ReqCreateCustomRoom{
		Uid: uid,
		Msg: bufc,
	}
	rst := &proto.OperRst{}

	err = conn.Call("MatchServer.WithdrawREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

// //是否在游戏中
// func PlayerInRoom(uid string) (bool, error) {
// 	err, conn := pPoll.RandomGetConn()
// 	if err != nil {
// 		return false, err
// 	}

// 	req := &proto.ReqIsInRoom{
// 		Uid: uid,
// 	}
// 	rst := &proto.OperRst{}

// 	err = conn.Call("RoomServer.PlayerIsInRoom", req, rst)
// 	if err != nil {
// 		return false, err
// 	}

// 	if rst.Ok == "OK" {
// 		return true, nil
// 	}
// 	return false, nil
// }
