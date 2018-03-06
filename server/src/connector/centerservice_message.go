package connector

import (
	gp "code.google.com/p/goprotobuf/proto"
	"common"
	"errors"
	"logger"
	"pockerclient"
	"proto"
	"rpc"
	"time"
)

//创建函数
type CreateMsgFun func() gp.Message

var mapRpc map[string]CreateMsgFun //消息回调用

func init() {
	mapRpc = make(map[string]CreateMsgFun)

	// mapRpc["Notice"] = func() gp.Message { return &rpc.Notice{} }
	// mapRpc["ClanChatMessage"] = func() gp.Message { return &rpc.ClanChatMessage{} }
	mapRpc["Msg"] = func() gp.Message { return &rpc.Msg{} }
	// mapRpc["PlayerMail"] = func() gp.Message { return &rpc.PlayerMail{} }
	mapRpc["S2CChatWorld"] = func() gp.Message { return &rpc.S2CChatWorld{} }
	mapRpc["S2CChatP2P"] = func() gp.Message { return &rpc.S2CChatP2P{} }
	mapRpc["ActionNotifyACK"] = func() gp.Message { return &rpc.ActionNotifyACK{} }
	mapRpc["ActionACK"] = func() gp.Message { return &rpc.ActionACK{} }
	mapRpc["GameStartACK"] = func() gp.Message { return &rpc.GameStartACK{} }
	mapRpc["LeaveRoomACK"] = func() gp.Message { return &rpc.LeaveRoomACK{} }
	mapRpc["EnterRoomACK"] = func() gp.Message { return &rpc.EnterRoomACK{} }
	mapRpc["JieSuanNotifyACK"] = func() gp.Message { return &rpc.JieSuanNotifyACK{} }
	mapRpc["CountdownNotifyACK"] = func() gp.Message { return &rpc.CountdownNotifyACK{} }
	mapRpc["PassCardNotifyACK"] = func() gp.Message { return &rpc.PassCardNotifyACK{} }
	mapRpc["PassedNotifyACK"] = func() gp.Message { return &rpc.PassedNotifyACK{} }
	mapRpc["BroadCastNotify"] = func() gp.Message { return &rpc.BroadCastNotify{} }
	mapRpc["AddMailNotify"] = func() gp.Message { return &rpc.AddMailNotify{} }
	mapRpc["FightRoomChatNotify"] = func() gp.Message { return &rpc.FightRoomChatNotify{} }
	mapRpc["EnterCustomRoomACK"] = func() gp.Message { return &rpc.EnterCustomRoomACK{} }
	mapRpc["LeaveCustomRoomACK"] = func() gp.Message { return &rpc.LeaveCustomRoomACK{} }
	mapRpc["RoomListACK"] = func() gp.Message { return &rpc.RoomListACK{} }
	mapRpc["CreateRoomACK"] = func() gp.Message { return &rpc.CreateRoomACK{} }
	mapRpc["FindRoomACK"] = func() gp.Message { return &rpc.FindRoomACK{} }
	mapRpc["JieSanRoomNotify"] = func() gp.Message { return &rpc.JieSanRoomNotify{} }
	mapRpc["JieSanRoomUpdateStatusNotify"] = func() gp.Message { return &rpc.JieSanRoomUpdateStatusNotify{} }
	mapRpc["SendFriendChat"] = func() gp.Message { return &rpc.SendFriendChat{} }
	mapRpc["FinalJieSuanNotifyACK"] = func() gp.Message { return &rpc.FinalJieSuanNotifyACK{} }
	mapRpc["FriendStatusNofify"] = func() gp.Message { return &rpc.FriendStatusNofify{} }
	mapRpc["DelFriendNofity"] = func() gp.Message { return &rpc.DelFriendNofity{} }
	mapRpc["InviteFirendsJionCustomRoomNotify"] = func() gp.Message { return &rpc.InviteFirendsJionCustomRoomNotify{} }
	mapRpc["PockerRoomInfo"] = func() gp.Message { return &rpc.PockerRoomInfo{} }
	mapRpc["PockerManBase"] = func() gp.Message { return &rpc.PockerManBase{} }
	mapRpc["S2CAction"] = func() gp.Message { return &rpc.S2CAction{} }

	mapRpc["MJEnterRoomACK"] = func() gp.Message { return &rpc.MJEnterRoomACK{} }
	mapRpc["MJLeaveRoomACK"] = func() gp.Message { return &rpc.MJLeaveRoomACK{} }
	mapRpc["MJGameStartACK"] = func() gp.Message { return &rpc.MJGameStartACK{} }
	mapRpc["MJActionACK"] = func() gp.Message { return &rpc.MJActionACK{} }
	mapRpc["MJActionNotifyACK"] = func() gp.Message { return &rpc.MJActionNotifyACK{} }
	mapRpc["MJCountdownNotifyACK"] = func() gp.Message { return &rpc.MJCountdownNotifyACK{} }
	mapRpc["MJJieSuanNotifyACK"] = func() gp.Message { return &rpc.MJJieSuanNotifyACK{} }
	mapRpc["MJRemoveCardNotifyACK"] = func() gp.Message { return &rpc.MJRemoveCardNotifyACK{} }
	mapRpc["LeavePockerRoom"] = func() gp.Message { return &rpc.LeavePockerRoom{} }
	mapRpc["PayResultNotify"] = func() gp.Message { return &rpc.PayResultNotify{} }

}

//来自Center的消息
func (self *CenterService) SendMsg2Player(req *proto.ChatSendMsg2Player, reply *proto.ChatSendMsg2PlayerResult) (err error) {
	logger.Info("Come into SendMsg2Player", req.MsgName)
	f, ok := mapRpc[req.MsgName]
	if !ok {
		logger.Error("CenterService.SendMsg2Player wrong msgname", req.MsgName)
		return errors.New("wrong msgname:" + req.MsgName)
	}

	msg := f()
	if err := common.DecodeMessage(req.Buf, msg); err != nil {
		//logger.Info("CenterService.DecodeMessage error")
		return err
	}

	//所有玩家
	if len(req.PlayerList) == 0 {
		self.saveChatMsg(req.MsgName, msg)
	} else {
		for _, uid := range req.PlayerList {
			if p, ok := cns.getPlayerByUid(uid); ok {
				if req.MsgName == "LeaveRoomACK" || req.MsgName == "MJLeaveRoomACK" || req.MsgName == "LeaveCustomRoomACK" {
					p.SetGameType("")
					p.SetRoomType(int32(0))
					logger.Info("*************离开房间:%s", req.MsgName)
				} else if req.MsgName == "LeavePockerRoom" {
					logger.Info("================扑克离开房间")
					p.SetGameType("")
					p.SetRoomType(int32(0))
					return nil
				} else if req.MsgName == "S2CAction" {
					s, ok := msg.(*rpc.S2CAction)
					if !ok {
						logger.Error("SendMsg2Player S2CAction err")
						return
					}

					if p.GetUid() == uid && s.GetAct() == int32(1) {
						logger.Info("================扑克玩家离开 正常发送请求")
						pkMsg := &rpc.C2SAction{}
						pkMsg.SetAct(int32(1))
						pkMsg.SetUid(s.GetOperater())
						pockerclient.ReqAction(pkMsg)
					}
				}

				WriteResult(p.conn, msg)
			}
		}
	}
	return nil
}

func (self *CenterService) saveChatMsg(method string, msg gp.Message) {
	if method != "S2CChatWorld" {
		cns.serverForClient.ServerBroadcast(msg)
		return
	}
	// 世界聊天不主动推送 存储起来等玩家来取
	cns.chatLock.Lock()
	defer cns.chatLock.Unlock()

	timeNow := uint32(time.Now().Unix())
	index := uint32(1)
	if len(cns.chatMsgs) > 0 {
		info := cns.chatMsgs[len(cns.chatMsgs)-1]
		index = info.msgIndex + 1
		if index == 0 {
			index = 1
		}
	}
	info := &stChatMsg{
		sendTime: timeNow,
		msgIndex: index,
		msg:      msg,
	}

	maxChatNum := uint32(GetGlobalCfg("WORLD_CHAT_SAVE_NUMBER"))
	if len(cns.chatMsgs) < int(maxChatNum) {
		cns.chatMsgs = append(cns.chatMsgs, info)
	} else {
		cns.chatMsgs = cns.chatMsgs[1:]
		cns.chatMsgs = append(cns.chatMsgs, info)
	}
}

func (self *CenterService) SendMsg2LocalPlayer(req *proto.ChatSendMsg2LPlayer, reply *proto.ChatSendMsg2LPlayerResult) (err error) {
	f, ok := mapRpc[req.MsgName]
	if !ok {
		logger.Error("CenterService.SendMsg2LocalPlayer wrong msgname", req.MsgName)
		return errors.New("wrong msgname l:" + req.MsgName)
	}

	msg := f()
	if err := common.DecodeMessage(req.Buf, msg); err != nil {
		return err
	}

	for _, seg := range cns.players {
		seg.l.RLock()
		for _, p := range seg.players {
			// logger.Info("", p.GetLevel())
			//判断各种条件
			// if req.Channel == rpc.Login_All ||
			// req.Channel == common.GetPlatformByGamelocation(p.GetGamelocation()) {
			// if p.GetLevel() >= req.LevelMin &&
			// 	p.GetLevel() <= req.LevelMax {
			WriteResult(p.conn, msg)
			// }
			// }
		}
		seg.l.RUnlock()
	}

	return nil
}

func (self *CenterService) C2SCostResource(req *proto.ReqCostRes, rst *proto.CommonRst) error {
	for _, uid := range req.PlayerList {
		p, ok := cns.getPlayerByUid(uid)
		if !ok {
			logger.Error("C2SCostResource cns.getPlayerByUid return nil, uid:%s", uid)
			continue
		}
		p.Billing(req)
	}
	return nil
}

//支付结果通知
func (self *CenterService) SendPayResult2Player(req *proto.ReqRechargeNofity, rst *proto.CommonRst) error {
	for _, uid := range req.PlayerList {
		p, ok := cns.getPlayerByUid(uid)
		if !ok {
			logger.Error("SendPayResult2Player cns.getPlayerByUid return nil, uid:%s", uid)
			continue
		}
		msg := &rpc.PayResultNotify{}
		if err := common.DecodeMessage(req.Buf, msg); err != nil {
			return err
		}

		p.OnRecharged(msg)
	}
	return nil
}

//公共方法
func (self *CenterService) CallCnserverFunc(req *proto.CallCnserverMsg, rst *proto.CommonRst) error {
	for _, uid := range req.Uids {
		p, ok := cns.getPlayerByUid(uid)
		if !ok {
			logger.Error("CallCnserverFunc cns.getPlayerByUid return nil, uid:%s", uid)
			continue
		}

		if req.Param1 == "PockerEnd" {
			p.TaskTrigger(SIG_PLAY_POCKER, false)
		}
	}

	return nil
}

//统计在线人数
func (self *CenterService) GetOnlineNumber(req *proto.GetOnlineNumber, rst *proto.GetOnlineNumberRst) error {
	rst.Numbers = cns.getOnlineNumbers()

	return nil
}

func (self *CenterService) LogOnlineNumber(req *proto.GetOnlineNumberRst, rst *proto.GetOnlineNumber) error {
	//tlog
	// go TLogOnlineNumbers(req.Numbers)

	return nil
}
