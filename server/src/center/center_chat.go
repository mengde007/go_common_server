/*
本文件从原chatserver移植过来，去掉原chatserver
*/

package center

import (
	"common"
	"logger"
	"proto"
	"rpc"
	"time"
)

type MapGasList map[uint8][]string

func (self *Center) getOnGasList(uids []string) MapGasList {
	mapRst := make(MapGasList)

	if len(uids) == 0 {
		for sid, _ := range self.cnss {
			mapRst[sid] = make([]string, 0)
		}
	} else {
		for _, uid := range uids {
			sid, ok := self.GetPlayerGasId(uid)
			if ok {
				if lists, ok := mapRst[sid]; ok {
					mapRst[sid] = append(lists, uid)
				} else {
					mapRst[sid] = append([]string{}, uid)
				}
			}
		}
	}

	return mapRst
}

func (self *Center) PlayerChatToPlayer(msg *proto.PlayerChatToPlayer, result *proto.PlayerChatToPlayerResult) error {
	cmd := &rpc.S2CChatP2P{}
	cmd.SetFromPlayerId(msg.FromPlayerId)
	cmd.SetFromPlayerName(msg.FromPlayerName)
	cmd.SetFromPlayerLevel(msg.FromPlayerLevel)
	cmd.SetChatContent(msg.Content)

	buf, err := common.EncodeMessage(cmd)
	if err != nil {
		return err
	}

	reqGas := &proto.ChatSendMsg2Player{
		MsgName:    "S2CChatP2P",
		PlayerList: append([]string{}, msg.ToPlayerId),
		Buf:        buf,
	}

	rstGas := &proto.ChatSendMsg2PlayerResult{}

	return self.SendMsg2Player(reqGas, rstGas)
}

func (self *Center) PlayerWorldChat(msg *proto.PlayerWorldChat, result *proto.PlayerWorldChatResult) error {
	cmd := &rpc.S2CChatWorld{}
	cmd.SetFromPlayerId(msg.FromPlayerId)
	cmd.SetFromPlayerName(msg.FromPlayerName)
	cmd.SetFromPlayerLevel(msg.FromPlayerLevel)
	cmd.SetChatContent(msg.Content)
	cmd.SetChatTime(time.Now().Unix())
	// cmd.SetLastRank(msg.LastLeagueRank)
	// cmd.SetVipLevel(msg.Viplevel)
	// cmd.SetUseIM(msg.UseIM)
	// cmd.SetVoiceTime(msg.VoiceTime)
	// if msg.CName != "" {
	// 	cmd.SetAllianceName(msg.CName)
	// 	cmd.SetAllianceSymbol(msg.CSymbol)
	// }

	buf, err := common.EncodeMessage(cmd)
	if err != nil {
		return err
	}

	reqGas := &proto.ChatSendMsg2Player{
		MsgName:    "S2CChatWorld",
		PlayerList: make([]string, 0),
		Buf:        buf,
	}

	rstGas := &proto.ChatSendMsg2PlayerResult{}

	return self.SendMsg2Player(reqGas, rstGas)
}

func (self *Center) SendMsg2Player(req *proto.ChatSendMsg2Player, rst *proto.ChatSendMsg2PlayerResult) error {
	for sid, uids := range self.getOnGasList(req.PlayerList) {
		if conn, ok := self.cnss[sid]; ok {
			reqGas := &proto.ChatSendMsg2Player{
				MsgName:    req.MsgName,
				PlayerList: uids,
				Buf:        req.Buf,
			}
			rstGas := &proto.ChatSendMsg2PlayerResult{}
			conn.Go("CenterService.SendMsg2Player", reqGas, rstGas, nil)
		} else {
			logger.Error("SendMsg2Player can't find sid:%d, uid:", sid, uids)
		}
	}
	return nil
}

func (self *Center) CostPlayerRes(req *proto.ReqCostRes, rst *proto.CommonRst) error {
	if len(req.PlayerList) == 0 {
		return nil
	}

	//扣钱不在线先放缓存
	lst := self.getOnGasList(req.PlayerList)
	if len(lst) == 0 {
		self.saveCost2Cache(req)
		return nil
	}

	for sid, uids := range lst {
		if conn, ok := self.cnss[sid]; ok {
			req.PlayerList = uids
			conn.Go("CenterService.C2SCostResource", req, rst, nil)
		}
	}
	return nil
}

//通知玩家支付结果
func (self *Center) SendPayResult2Player(req *proto.ReqRechargeNofity, rst *proto.CommonRst) error {
	if len(req.PlayerList) == 0 {
		return nil
	}

	lst := self.getOnGasList(req.PlayerList)
	if len(lst) == 0 {
		// self.saveCost2Cache(req)//暂时不需要缓存
		return nil
	}

	for sid, uids := range lst {
		if conn, ok := self.cnss[sid]; ok {
			req.PlayerList = uids
			conn.Go("CenterService.SendPayResult2Player", req, rst, nil)
		}
	}
	return nil
}

func (self *Center) SendMsg2PlayerCall(req *proto.ChatSendMsg2Player, rst *proto.ChatSendMsg2PlayerResult) error {
	for sid, uids := range self.getOnGasList(req.PlayerList) {
		if conn, ok := self.cnss[sid]; ok {
			reqGas := &proto.ChatSendMsg2Player{
				MsgName:    req.MsgName,
				PlayerList: uids,
				Buf:        req.Buf,
			}
			rstGas := &proto.ChatSendMsg2PlayerResult{}
			conn.Call("CenterService.SendMsg2Player", reqGas, rstGas)
		}
	}

	return nil
}

func (self *Center) SendMsg2LocationPlayer(req *proto.ChatSendMsg2LPlayer, rst *proto.ChatSendMsg2LPlayerResult) error {
	for _, conn := range self.cnss {
		conn.Go("CenterService.SendMsg2LocalPlayer", req, rst, nil)
	}

	return nil
}

func (self *Center) SendActionsNotify(msg *proto.PushDaerMsg2Player, result *proto.PlayerChatToPlayerResult) error {
	logger.Info("SendActionsNotify begin, uid:%s, msg.Func:%s", msg.PlayerUids, msg.Func)
	defer logger.Info("SendActionsNotify end")
	reqGas := &proto.ChatSendMsg2Player{
		MsgName:    msg.Func,
		PlayerList: msg.PlayerUids,
		Buf:        msg.Value,
	}
	rstGas := &proto.ChatSendMsg2PlayerResult{}

	return self.SendMsg2PlayerCall(reqGas, rstGas)
}
