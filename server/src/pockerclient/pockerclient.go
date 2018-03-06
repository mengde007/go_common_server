package pockerclient

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
	aServerHost := common.ReadServerClientConfig("pockerserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "pockerserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}
	return
}

//进入房间
func EnterPockerRoom(base *rpc.PlayerBaseInfo, gameType string, eType int32) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(base)
	if err != nil {
		return err
	}

	req := &proto.ReqEnterPockerRoom{
		GameType: gameType,
		EType:    eType,
		Msg:      bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("PockerServer.PlayerEnterGame", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//进入房间
func QuicklySeatdown(base *rpc.PlayerBaseInfo) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(base)
	if err != nil {
		return err
	}

	req := &proto.ReqEnterPockerRoom{
		Msg: bufb,
	}
	rst := &proto.OperRst{}

	err = conn.Call("PockerServer.QuicklySeatdown", req, rst)
	if err != nil {
		return err
	}
	return nil
}

//创建自寻房
func CreateCustomRoom(limId, blindId int32, base *rpc.PlayerBaseInfo) (error, int32) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err, -1
	}

	bufb, err := common.EncodeMessage(base)
	if err != nil {
		return err, -1
	}

	req := &proto.ReqPockerCustomRoom{
		Msg:     bufb,
		BlindId: limId,
		LimId:   blindId,
	}
	rst := &proto.RstPockerCustomRoom{}
	err = conn.Call("PockerServer.CreateCustomRoom", req, rst)
	if err != nil {
		return err, -1
	}
	return nil, rst.RoomNo
}

//请求动作
func ReqAction(msg *rpc.C2SAction) error {
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

	err = conn.Call("PockerServer.PlayerAction", req, rst)
	if err != nil {
		return err
	}
	return nil
}

func GetPockerOnlineNum() (*rpc.OnlineInfo, error) {
	conns := pPoll.GetAllConn()
	if conns == nil {
		logger.Error("GetPockerOnlineNum pPoll.GetAllConn() return nil")
		return nil, nil
	}

	sendMsg := &rpc.OnlineInfo{}
	for _, conn := range conns {
		req := &proto.ReqOnlineNum{}
		rst := &proto.RstOnlineNum{}
		err := conn.Call("PockerServer.GetOnlineNums", req, rst)
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

	err = conn.Call("PockerServer.PlayerIsInRoom", req, rst)
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

	err = conn.Call("PockerServer.SendDeskMsg", req, rst)
	if err != nil {
		return err
	}
	return nil
}
