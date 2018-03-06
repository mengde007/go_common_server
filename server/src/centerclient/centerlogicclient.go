package centerclient

import (
	gp "code.google.com/p/goprotobuf/proto"
	"common"
	"logger"
	"proto"
	"rpc"
)

//参数如果有proto结构的，格式为 rpc.StructName 例如: msg *rpc.FormatedMsg
//给玩家发消息要用到哪些函数你就在这个文件里定义,函数体我来写,参数必须有玩家的uid,玩家登陆到大二服务器会传过来的

//发送城池数据给玩家
func SendActionNotify(uid string, msg *rpc.ActionNotifyACK) error {
	logger.Info("SendActionNotify called uid:%s", uid)

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "ActionNotifyACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendAction(uid string, msg *rpc.ActionACK) error {
	logger.Info("SendAction called uid:%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "ActionACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendGameStart(uid string, msg *rpc.GameStartACK) error {
	logger.Info("SendGameStart called uid:%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "GameStartACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendEnterRoom(uid string, msg *rpc.EnterRoomACK) error {
	logger.Info("SendEnterRoom called uid:%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "EnterRoomACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendLeaveRoom(uid string, msg *rpc.LeaveRoomACK) error {
	logger.Info("SendLeaveRoom called uid :%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "LeaveRoomACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendJieSuan(uid string, msg *rpc.JieSuanNotifyACK) error {
	logger.Info("SendJieSuan called uid:%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "JieSuanNotifyACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendCountdownNotify(uid string, msg *rpc.CountdownNotifyACK) error {
	logger.Info("SendCountdownNotify called uid:%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "CountdownNotifyACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendPassCardNotify(uid string, msg *rpc.PassCardNotifyACK) error {
	logger.Info("SendPassCardNotify called uid:%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "PassCardNotifyACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendCostResourceMsg(uid, rstType, gameType string, resValue int32) error {
	logger.Info("SendCostResourceMsg called uid:%s, resType:%s, resValue:%s", uid, rstType, resValue)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.ReqCostRes{
		PlayerList: uids,
		ResName:    rstType,
		ResValue:   resValue,
		GameType:   gameType,
	}
	rst := &proto.CommonRst{}
	conn.Call("Center.CostPlayerRes", req, rst)
	return nil
}

func SendPassedNotify(uid string, msg *rpc.PassedNotifyACK) error {
	logger.Info("SendPassedNotify called uid:%s", uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}
	uids := []string{}
	uids = append(uids, uid)

	req := &proto.PushDaerMsg2Player{
		Func:       "PassedNotifyACK",
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendCommonNotify2S(uids []string, msg gp.Message, Func string) error {
	logger.Info("==================SendCommonNotify2S called uid, msg:%v", msg)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.PushDaerMsg2Player{
		Func:       Func,
		PlayerUids: uids,
		Value:      buf,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Call("Center.SendActionsNotify", req, rst)
	return nil
}

func SendPayResult2Player(uids []string, msg gp.Message) error {
	logger.Info("==================SendPayResult2Player called Func:")
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ReqRechargeNofity{
		PlayerList: uids,
		Buf:        buf,
	}
	rst := &proto.CommonRst{}
	conn.Call("Center.SendPayResult2Player", req, rst)
	return nil
}
