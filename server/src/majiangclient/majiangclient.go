package majiangclient

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
	aServerHost := common.ReadServerClientConfig("majiangserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "majiangserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}
	return
}

//快熟进入房间
func QuickEnterMaJiangRoom(base *rpc.PlayerBaseInfo, client *rpc.MJQuickEnterRoomREQ) error {
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

	err = conn.Call("MaJiangServer.QuickEnterRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//进入房间
func EnterMaJiangRoom(base *rpc.PlayerBaseInfo, client *rpc.EnterRoomREQ) error {
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

	err = conn.Call("MaJiangServer.EnterRoom", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//离开房间
func LeaveMaJiangRoom(msg *rpc.MJLeaveRoomREQ) error {
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

	err = conn.Call("MaJiangServer.LeaveRoom", req, rst)
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

	err = conn.Call("MaJiangServer.ForceLeaveRoom", req, rst)
	if err != nil {
		return err
	}
	return nil

}

//请求动作
func MaJiangActionREQ(msg *rpc.ActionREQ) error {
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

	err = conn.Call("MaJiangServer.ActionsREQ", req, rst)
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

	err = conn.Call("MaJiangServer.TieGuiREQ", req, rst)
	if err != nil {
		return err
	}
	return nil
}

func GetMaJiangServrOnlineNum() (*rpc.OnlineInfo, error) {
	conns := pPoll.GetAllConn()
	if conns == nil {
		logger.Error("GetDaerServrOnlineNum pPoll.GetAllConn() return nil")
		return nil, nil
	}

	sendMsg := &rpc.OnlineInfo{}
	for _, conn := range conns {
		req := &proto.ReqOnlineNum{}
		rst := &proto.RstOnlineNum{}
		err := conn.Call("MaJiangServer.GetOnlineNums", req, rst)
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
func PlayerInMaJiangGame(uid string) (bool, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return false, err
	}

	req := &proto.ReqIsInRoom{
		Uid: uid,
	}
	rst := &proto.OperRst{}

	err = conn.Call("MaJiangServer.PlayerIsInRoom", req, rst)
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

	err = conn.Call("MaJiangServer.SendDeskMsg", req, rst)
	if err != nil {
		return err
	}
	return nil
}
