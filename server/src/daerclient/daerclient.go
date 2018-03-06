package daerclient

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
	aServerHost := common.ReadServerClientConfig("daerserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "daerserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}
	return
}

//快熟进入房间
func QuickEnterRoom(base *rpc.PlayerBaseInfo, client *rpc.QuickEnterRoomREQ) error {
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

	err = conn.Call("DaerServer.QuickEnteredRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//进入房间
func EnterDaerRoom(base *rpc.PlayerBaseInfo, client *rpc.EnterRoomREQ) error {
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

	err = conn.Call("DaerServer.EnteredDaerRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//离开大二房间
func LeaveDaerRoom(msg *rpc.LeaveRoomREQ) error {
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

	err = conn.Call("DaerServer.LeavedDaerRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//踢人
func ForceLeaveRoom(uid string, msg *rpc.ForceLeaveRoomREQ) error {
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

	err = conn.Call("DaerServer.ForceLeaveRoom", req, rst)
	if err != nil {
		return err
	}
	return nil

}

//请求动作
func DaerActionREQ(msg *rpc.ActionREQ) error {
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

	err = conn.Call("DaerServer.DaerActionsREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

func GetDaerServrOnlineNum() (*rpc.OnlineInfo, error) {
	conns := pPoll.GetAllConn()
	if conns == nil {
		logger.Error("GetDaerServrOnlineNum pPoll.GetAllConn() return nil")
		return nil, nil
	}

	sendMsg := &rpc.OnlineInfo{}
	for _, conn := range conns {
		req := &proto.ReqOnlineNum{}
		rst := &proto.RstOnlineNum{}
		err := conn.Call("DaerServer.GetOnlineNums", req, rst)
		if err != nil {
			return nil, err
		}
		msg := &rpc.OnlineInfo{}
		err = common.DecodeMessage(rst.RoomInfo, msg)
		if err != nil {
			return nil, err
		}
		sendMsg.Info = append(sendMsg.Info, msg.Info...)
	}

	return sendMsg, nil
}

//是否在游戏中
func PlayerInDaerGame(uid string) (bool, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return false, err
	}

	req := &proto.ReqIsInRoom{
		Uid: uid,
	}
	rst := &proto.OperRst{}

	err = conn.Call("DaerServer.PlayerIsInRoom", req, rst)
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

	err = conn.Call("DaerServer.SendDeskMsg", req, rst)
	if err != nil {
		return err
	}
	return nil
}
