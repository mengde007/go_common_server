package connector

import (
	"accountclient"
	"centerclient"
	"common"
	"daerclient"
	"gmclient"
	"lockclient"
	"logger"
	"mailclient"
	"majiangclient"
	"payclient"
	"pockerclient"
	"roomclient"
	"rpc"
	"strconv"
	"strings"
)

// conn:操作保险箱
func (self *CNServer) OperateInsurence(conn rpc.RpcConn, msg rpc.ReqInsurenceMoney) error {
	logger.Info("client call OperateInsurence begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	if msg.GetBWithdraw() {
		p.Withdraw(msg.GetValue())
	} else {
		p.SaveMoney(msg.GetValue())
	}
	return nil
}

// conn:玩家发送跑马灯
func (self *CNServer) SendBraodCast(conn rpc.RpcConn, msg rpc.ReqBroadCast) error {
	logger.Info("client call SendBraodCast begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	// 临时测试用，正式时，把下面恢复
	// p.playerGm("$$item 7 100")

	// openGm := common.GetDaerGlobalIntValue("51")
	if common.IsOpenGm() && !p.playerGm(msg.GetContent()) {
		return nil
	}

	cost := common.GetDaerGlobalIntValue("44")
	if p.GetCoin() < cost {
		logger.Error("SendBraodCast player not enough money, need:%d, cur:%d", cost, p.GetCoin())
		return nil
	}
	p.SetCoin(p.GetCoin() - cost)
	p.ResourceChangeNotify()

	msg.SetBVip(false)
	if p.GetVipLeftDay() > 0 {
		msg.SetBVip(true)
	}
	msg.SetPlayerName(p.GetName())

	gmclient.PlayerSendNotice(&msg)
	return nil
}

// conn:玩家读取邮件
func (self *CNServer) PlayerReadMail(conn rpc.RpcConn, msg rpc.ReqReadOneMail) error {
	logger.Info("client call PlayerReadMail begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	mailclient.ReadMail(p.GetUid(), msg.GetMailId())
	return nil
}

// conn:取附件
func (self *CNServer) GetMailAttach(conn rpc.RpcConn, msg rpc.ReqReadOneMail) error {
	logger.Info("client call GetMailAttach begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}
	p.GetAttach(msg.GetMailId())
	return nil
}

// conn:游戏中聊天
func (self *CNServer) SendDeskChat(conn rpc.RpcConn, msg rpc.FightRoomChatNotify) error {
	logger.Info("client call SendDeskChat begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	//使用道具
	ftMsg := msg.GetFighChatinfo()
	if ftMsg != nil && ftMsg.GetItemId() != "" {
		if !p.CostItem2Bag(ftMsg.GetItemId(), 1) {
			logger.Error("道具不够，使用毛线啊")
			return nil
		}
	}

	k := ftMsg.GetGameType()

	logger.Info("****************gameType:%s", k)
	if k == "1" {
		daerclient.ReqSendDeskChat(&msg)
	} else if k == "2" {
		majiangclient.ReqSendDeskChat(&msg)
	} else if k == "3" {
		pockerclient.ReqSendDeskChat(&msg)
	} else if k == "4" || k == "5" || k == "6" {
		roomclient.ReqSendDeskChat(&msg)
	}

	return nil
}

func (self *CNServer) Bind3rdAccount(conn rpc.RpcConn, msg rpc.Login) error {
	logger.Info("client call Bind3rdAccount begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	errMsg := &rpc.ErrorCodeNofify{}
	if p.GetAccountType() == 1 {
		logger.Error("Bind3rdAccount error, account already bind")
		errMsg.SetCode(23)
		WriteResult(conn, errMsg)
		return nil
	}

	if msg.GetOpenid() == "" || msg.GetUid() != p.GetUid() {
		logger.Error("Bind3rdAccount err msg.GetOpenid() ==  || msg.GetUid() != p.GetUid()")
		return nil
	}

	binduid, err := accountclient.QueryPlayerIdByPartnerId(
		common.TB_t_account_tencentid2playerid,
		msg.GetOpenid())
	if err != nil {
		logger.Info("connId = %d login QueryPlayerIdByPartnerId :%v", conn.GetId(), err)
		return nil
	}
	if len(binduid) > 0 {
		errMsg.SetCode(24)
		WriteResult(conn, errMsg)
		return nil
	}

	opendId := strconv.Itoa(int(p.GetRoleId()))
	err = accountclient.DelPartnerIdToPlayerId(common.TB_t_account_tencentid2playerid, opendId, 10)
	if err != nil {
		logger.Error("Bind3rdAccount accountclient.DelPartnerIdToPlayerId err, %s", err)
		return nil
	}

	err = accountclient.SetPartnerIdToPlayerId(
		common.TB_t_account_tencentid2playerid,
		msg.GetOpenid(),
		10,
		p.GetUid())
	if err != nil {
		logger.Error("Bind3rdAccount err,", err)
		return nil
	}

	p.SetName(msg.GetNickName())
	p.SetHeaderUrl(msg.GetHeaderUrl())
	p.SetSex(msg.GetSex())
	p.SetAccountType(1)

	playerinfo := &rpc.PlayerInfo{Base: p.PlayerBaseInfo}
	WriteResult(conn, playerinfo)

	return nil
}

func (self *CNServer) GetOtherPlayerInfo(conn rpc.RpcConn, msg rpc.ReqString) error {
	// op, exist := self.getPlayerByConnId(conn.GetId())
	//load otherplayer
	op := LoadOtherPlayer(msg.GetId())
	if op == nil {
		logger.Error("GetOtherPlayerInfo  LoadOtherPlayer return nil, msg:%v", &msg)
		return nil
	}

	opMsg := &rpc.Player{}
	opMsg.SetName(op.GetName())
	opMsg.SetSex(op.GetSex())
	opMsg.SetLevel(op.GetLevel())
	opMsg.SetHeader(op.GetHeader())
	opMsg.SetRoleId(op.GetRoleId())
	opMsg.SetHeaderUrl(op.GetHeaderUrl())
	opMsg.SetUid(op.GetUid())
	opMsg.SetExp(op.GetExp())
	opMsg.SetCoin(op.GetCoin())
	opMsg.SetDiamond(op.GetGem())
	opMsg.SetBOnline(false)
	if lockclient.IsOnline(msg.GetId()) {
		opMsg.SetBOnline(true)
	}
	opMsg.Scores = op.Scores

	WriteResult(conn, opMsg)
	return nil
}

func (self *CNServer) ForceLeaveRoom(conn rpc.RpcConn, msg rpc.ForceLeaveRoomREQ) error {
	logger.Info("client call ForceLeaveRoom begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	sysType := msg.GetSysType()
	switch sysType {
	case common.PiPeiFang:
		itemAmount := p.GetItemNum(strconv.Itoa(common.KickCardID))
		logger.Error("%s道具数量:%s", common.KickCardID, itemAmount)
		if itemAmount > 0 {
			switch msg.GetGameType() {
			case common.DaerGame:
				daerclient.ForceLeaveRoom(p.PlayerBaseInfo.GetUid(), &msg)
			case common.MaJiang:
				majiangclient.ForceLeaveRoom(p.PlayerBaseInfo.GetUid(), &msg)
			case common.DeZhouPuker:
				//pockerclient.EnterPockerRoom(p.PlayerBaseInfo, msg.GetRoomType())
			default:
				logger.Error("未知的游戏类型")
			}

		}
		break
	case common.ZiJianFang:
		roomclient.ForceLeaveRoomREQ(p.PlayerBaseInfo.GetUid(), &msg)
		break
	case common.BiSaiFang:
		break
	default:
		logger.Error("未知的系统类型")
		break
	}

	return nil
}

func (self *CNServer) Signitures(conn rpc.RpcConn, msg rpc.ReqInt) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	p.Signature()
	return nil
}

func (self *CNServer) SignatureBefore(conn rpc.RpcConn, msg rpc.ReqInt) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	p.SignatureBefore(msg.GetId())
	return nil
}

func (self *CNServer) TaskShare(conn rpc.RpcConn, msg rpc.ReqTaskShare) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	p.TaskShares(msg.GetBShare2Friend())
	return nil
}

func (self *CNServer) InvateFriends(conn rpc.RpcConn, msg rpc.InviteFirendsJionCustomRoomREQ) error {
	logger.Info("InvateFriends has been called")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	send := rpc.InviteFirendsJionCustomRoomNotify{}
	for _, v := range msg.PlayerID {
		send.SetCode(int32(0))
		send.SetGameType(p.GetGameType())
		send.SetRoomID(p.GetRoomType())
		send.SetInvitePlayerName(p.GetName())
		send.SetCurrencyType(msg.GetCurrencyType())

		if lockclient.IsOnline(v) {
			centerclient.SendCommonNotify2S([]string{v}, &send, "InviteFirendsJionCustomRoomNotify")
		} else {
			logger.Error("玩家不在线，还邀请个毛线啊")
		}
	}
	return nil
}

func (self *CNServer) GetTaskRewards(conn rpc.RpcConn, msg rpc.ReqInt) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	p.GetTaskRewards(msg.GetId())
	return nil
}

func (self *CNServer) UpdateRoleInfo(conn rpc.RpcConn, msg rpc.RoleInfo) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	if msg.GetName() != "" {
		p.SetName(msg.GetName())
		p.SetBModifyName(true)
	} else if msg.GetPhone() != "" {
		p.SetPhone(msg.GetPhone())
	} else if msg.GetSex() != 0 { //这个判断要放在最后，不然要出妖饿子
		p.SetSex(msg.GetSex())
		p.SetBModifySex(true)
	} else if msg.GetSex() == 0 {
		p.SetSex(0)
		msg.SetSex(0)
	}

	WriteResult(conn, &msg)
	return nil
}

func (self *CNServer) GetBankruptRewards(conn rpc.RpcConn, msg rpc.ReqInt) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	p.BankruptRewards()
	return nil
}

func (self *CNServer) GetSignExra(conn rpc.RpcConn, msg rpc.ReqInt) error {
	return nil
}

//生成支付订单
func (self *CNServer) ReqPrePay(conn rpc.RpcConn, msg rpc.ReqString) error {
	logger.Info("ReqPrePay called...id:%s", msg.GetId())
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}
	idnum := strings.Split(msg.GetId(), "_")
	if len(idnum) == 2 {
		p.Shopping(idnum)
		return nil
	}

	id := idnum[0]
	cfg := GetItemCfg(id)
	if cfg == nil {
		logger.Error("ReqPrePay common.GetItemCfg(:%d) return nil", id)
		return nil
	}

	arrs := strings.Split(cfg.BuyPrice, "_")
	if len(arrs) != 2 {
		logger.Error("ReqPrePay err cfg.BuyPrice:%s", cfg.BuyPrice)
		return nil
	}
	price, _ := strconv.Atoi(arrs[1])

	openId := p.mobileqqinfo.Openid
	ourips := conn.GetRemoteIp()
	ips := strings.Split(ourips, ":")
	logger.Info("Ips:%s, ip0:%s		openId:%s", ourips, ips[0], openId)

	info, err := payclient.CreatePayOrder(p.GetUid(), openId, id, ips[0], uint32(price))
	if err != nil {
		logger.Error("ReqPrePay payclient.CreatePayOrder uid:%s, name:%s, err:%s", p.GetUid(), p.GetName(), err)
		return nil
	}

	WriteResult(conn, info)
	return nil
}
