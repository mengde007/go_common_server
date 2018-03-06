package roomclient

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
	aServerHost := common.ReadServerClientConfig("roomserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "roomserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}
	return
}

//进入房间
func EnterRoom(base *rpc.PlayerBaseInfo, client *rpc.EnterCustomRoomREQ) error {
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

	err = conn.Call("RoomServer.EnterRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//离开房间
func LeaveRoom(msg *rpc.LeaveCustomRoomREQ) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqLeaveDaerRoom{
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.LeaveRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//请求动作
func ActionREQ(msg *rpc.ActionREQ) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqActionREQ{
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.ActionREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//请求是否贴鬼
func MaJiangTieGuiREQ(msg *rpc.MJTieGuiREQ) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqActionREQ{
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.TieGuiREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//创建房间
func CreateRoomREQ(base *rpc.PlayerBaseInfo, client *rpc.CreateRoomREQ, itemAmount int32) error {
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

	req := &proto.ReqDaerRoomWithItem{
		Base:   bufb,
		Client: bufc,
		Amount: itemAmount,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.CreateRoomREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//获取房间列表
func RoomListREQ(uid string, msg *rpc.RoomListREQ) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqCreateCustomRoom{
		Uid: uid,
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.RoomListREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

func FindRoomREQ(uid string, msg *rpc.FindRoomREQ) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqCreateCustomRoom{
		Uid: uid,
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.FindRoomREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

func JieSanRoomREQ(uid string, msg *rpc.JieSanRoomREQ) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqCreateCustomRoom{
		Uid: uid,
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.JieSanRoomREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

func ForceLeaveRoomREQ(uid string, msg *rpc.ForceLeaveRoomREQ) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqCreateCustomRoom{
		Uid: uid,
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.ForceLeaveRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//是否在游戏中
func PlayerInRoom(uid string) (bool, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return false, err
	}

	req := &proto.ReqIsInRoom{
		Uid: uid,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.PlayerIsInRoom", req, rst)
	if err != nil {
		return false, err
	}

	if rst.Ok == "OK" {
		return true, nil
	}
	return false, nil
}

//发送消息给其它玩家
func ReqSendDeskChat(msg *rpc.FightRoomChatNotify) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqActionREQ{
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("RoomServer.SendDeskMsg", req, rst)
	if err != nil {
		return err
	}
	return nil
}
