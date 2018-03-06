package roomserver

import (
	cmn "common"
	"logger"
	//	"math/rand"
	"rpc"

	conn "centerclient"
	ds "daerserver"
	"lockclient"
	"strconv"
	"time"
)

const (
	DefaultRoomType         = 4
	DefaultIsDaiGui         = true
	WaitingDissolveTimeName = "WaitingDissolveTime"
)

type CustomDaerRoom struct {
	ds.DaerRoom
	CustomRoom
}

//新建一个大贰自建房间
func NewCustomDaerRoom(uid int32, owner string, roomInfo *rpc.CreateRoomREQ) *CustomDaerRoom {
	r := new(CustomDaerRoom)
	r.InitCustomRoom( /*id,*/ owner, cmn.DaerGame, roomInfo)
	r.Init(uid, DefaultRoomType)
	r.SetSelector(r)

	//最大倍数
	r.MaxMultiple = roomInfo.GetMaxMultiple()
	//最大等待玩家执行的动作的时间（无限制等待）
	r.TimerInterval = 1000000000
	//游戏最大的执行时间
	r.GamingMaxTime = 2000000000

	switch r.currencyType {
	case CTCoin:
		cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(cmn.DaerGame)))
		if cfg != nil {
			r.RakeRate = cfg.RakeRate
		} else {
			logger.Error("读取房间配置表出错ID：%s", cmn.DaerGame)
		}

		r.Difen = roomInfo.GetDifen()
		r.IsDaigui = roomInfo.GetIsDaiGui()
		r.limitCoin = roomInfo.GetLimitCoin()

	case CTCredits:
		r.RakeRate = 0
		r.Difen = 1
		r.IsDaigui = roomInfo.GetIsDaiGui()
		r.limitCoin = 0
	default:
		logger.Error("不能识别的获取结算类型")
	}

	return r
}

//进入房间
func (self *CustomDaerRoom) Enter(p cmn.Player) {
	//检查输入参数
	if p == nil {
		logger.Error("player is nil.")
		return
	}

	player := p.(*ds.DaerPlayer)
	if player == nil {
		logger.Error("接口转换失败")
		return
	}

	//检查能否进入房间
	if self.IsFull() {
		logger.Error("DaerRoom.Enter: room is full")
		return
	}

	//进入房间时，将玩家的货币设置为初始化积分
	self.InitCurCredit(player)

	//修改玩家的房间相关的信息
	player.SetRoom(&self.DaerRoom)
	players := self.GetAllPlayer()

	for i := 0; i < ds.RoomMaxPlayerAmount; i++ {
		if players[i] == nil {
			players[i] = player
			self.SendEnterRoomACK(player)
			logger.Info("DaerRoom.Enter: player:%s enter room(%s):", player.ID(), i)
			break
		}
	}

	//开启一个房间停留计时器
	//self.StartForceLeave(p.ID())

	//开启解散房间的倒计时
	//self.StartDissolveRoom()
}

//重新进入房间
func (self *CustomDaerRoom) ReEnter(playerID string, playerInfo *rpc.PlayerBaseInfo) {
	//TODO:重新建立玩家连接，并下发当前玩家的数据
	if player := self.GetPlayerByID(playerID); player != nil {
		player.SetPlayerBasicInfo(playerInfo)
		self.InitCurCredit(player)
		self.SendEnterRoomACK(player)
		if self.IsGaming() {
			player.SendGameStartACK(true)
		} else {
			//开启解散房间的倒计时
			//self.StartDissolveRoom()
		}
		//检查是否有投票
		if self.IsVoting() {
			self.SendJieSanRoomNotify(self.DaerRoom.GetAllPlayerIDs())
		}
	} else {
		logger.Error("DaerRoom:player not in the room")
	}
}

//计算玩家的当前积分
func (self *CustomDaerRoom) InitCurCredit(p cmn.Player) {
	if p == nil {
		logger.Error("p is null.")
		return
	}

	if self.currencyType == CTCredits {
		if cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(cmn.DaerGame))); cfg != nil {
			var ownCredit int32 = 0
			if credit, exist := self.playerTotalCoin[p.ID()]; exist {
				ownCredit = credit
			}

			p.GetPlayerBasicInfo().SetCoin(cfg.InitCredit + ownCredit)

			logger.Info("玩家进入房间获得初始化积分%d", cfg.InitCredit)
		} else {
			p.GetPlayerBasicInfo().SetCoin(0)
			logger.Error("没读取到配置表！")
		}
	}
}

// //离开房间
func (self *CustomDaerRoom) Leave(uid string, isChangeDesk bool) bool {
	logger.Info("离开房间：", uid)
	//检查输入参数
	if uid == "" {
		logger.Error("player is nil.")
		return false
	}

	//只有在有投票的情况下，才能让玩家在游戏中就可以退出
	if !self.IsVoting() {
		//logger.Error("============没有投票的时候离开房间============是否在游戏中：", self.IsGaming())
		if self.IsGaming() {
			return false
		}
	}

	//重置一下当前胡牌的玩家
	self.SetHuPaiPlayerID("")
	players := self.GetAllPlayer()

	//检查是否要提前最终的结算
	if self.currencyType == CTCredits && self.IsStart() && !self.isAlreadyFinalJieSuan {
		self.DoFinalJieSuan()
		self.ForceAllPlayerLeave()
	}

	for i := 0; i < ds.RoomMaxPlayerAmount; i++ {
		tempPlayer := players[i]
		if tempPlayer != nil && uid == tempPlayer.ID() {
			self.SendLeaveRoomACK(tempPlayer, isChangeDesk)
			tempPlayer.SetRoom(nil)
			players[i] = nil
			tempPlayer.Reset()
			logger.Info("CustomDaerRoom.Leave: player:%s Leave room:", tempPlayer.ID())
			break
		}
	}

	//如果是房主离开，且还有人在房间时，则切换房主
	if uid == self.owner && !self.IsEmpty() {
		self.ChangeRoomOwner()
	}

	//删除roommgr中的引用关系
	customRoomMgr.DeleteLeavePlayerInfo(uid)

	//如果房间为空就结算房间线程了
	if self.IsEmpty() {
		logger.Error("停止房间的Timer和消息接受线程")
		self.ClearVoteList()
		*self.GetExitThreadHandle() <- true
	}

	return true
}

//改变房主
func (self *CustomDaerRoom) ChangeRoomOwner() {
	players := self.GetAllPlayer()

	for i := 0; i < ds.RoomMaxPlayerAmount; i++ {
		tempPlayer := players[i]
		if tempPlayer != nil {
			self.owner = tempPlayer.ID()
			logger.Info("CustomDaerRoom.ChangeRoomOwner: 改变房主到：%s", self.owner)
			break
		}
	}
}

func (self *CustomDaerRoom) SendEnterRoomACK(p cmn.Player) {
	logger.Info("自建房间的发送进入房间ACK")
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*ds.DaerPlayer)
	if player == nil {
		logger.Error("player is nil.")
		return
	}

	//打印房间信息
	logger.Info("发送进入自建房间：", player.ID())
	logger.Info("自建房间情况：", self.GetPlayerAmount())

	//给自己发送ACK
	msg := &rpc.EnterCustomRoomACK{}
	msg.SetRoomId(self.UID())
	msg.SetShangjiaType(3)
	msg.SetBReady(player.IsReady())
	msg.SetPlayerInfo(player.GetPlayerBasicInfo())
	msg.SetGameType(cmn.DaerGame)
	msg.SetTimes(self.times)
	msg.SetCurTimes(self.curTimes)
	msg.SetIsOwner(player.ID() == self.owner)
	msg.SetCurrencyType(self.currencyType)
	msg.SetDifen(self.GetDifen())
	if err := conn.SendCommonNotify2S([]string{player.ID()}, msg, "EnterCustomRoomACK"); err != nil {
		logger.Error("给自己发送进入房间时出错：", err, msg)
		return
	}

	//给房间里的每个发送进入房间
	shangJia := player.GetShangJia()
	if shangJia != nil {
		msg := &rpc.EnterCustomRoomACK{}
		msg.SetRoomId(self.UID())
		msg.SetShangjiaType(2)
		msg.SetBReady(player.IsReady())
		msg.SetPlayerInfo(player.GetPlayerBasicInfo())
		msg.SetGameType(cmn.DaerGame)
		msg.SetTimes(self.times)
		msg.SetCurTimes(self.curTimes)
		msg.SetIsOwner(player.ID() == self.owner)
		msg.SetCurrencyType(self.currencyType)
		msg.SetDifen(self.GetDifen())
		if err := conn.SendCommonNotify2S([]string{shangJia.ID()}, msg, "EnterCustomRoomACK"); err != nil {
			logger.Error("给上家发送自己进入房间时出错：", err, msg)
			return
		}

		msg = &rpc.EnterCustomRoomACK{}
		msg.SetRoomId(self.UID())
		msg.SetShangjiaType(1)
		msg.SetBReady(shangJia.IsReady())
		msg.SetPlayerInfo(shangJia.GetPlayerBasicInfo())
		msg.SetGameType(cmn.DaerGame)
		msg.SetTimes(self.times)
		msg.SetCurTimes(self.curTimes)
		msg.SetIsOwner(shangJia.ID() == self.owner)
		msg.SetCurrencyType(self.currencyType)
		msg.SetDifen(self.GetDifen())
		if err := conn.SendCommonNotify2S([]string{player.ID()}, msg, "EnterCustomRoomACK"); err != nil {
			logger.Error("给自己发送上家进入房间时出错：", err, msg)
			return
		}

		logger.Info("sned EnterCustomRoomACK to shangJia(%s).", shangJia.ID())
	}

	xiaJia := player.GetXiaJia()
	if xiaJia != nil {
		msg := &rpc.EnterCustomRoomACK{}
		msg.SetRoomId(self.UID())
		msg.SetShangjiaType(1)
		msg.SetBReady(player.IsReady())
		msg.SetPlayerInfo(player.GetPlayerBasicInfo())
		msg.SetGameType(cmn.DaerGame)
		msg.SetTimes(self.times)
		msg.SetCurTimes(self.curTimes)
		msg.SetIsOwner(player.ID() == self.owner)
		msg.SetCurrencyType(self.currencyType)
		msg.SetDifen(self.GetDifen())
		if err := conn.SendCommonNotify2S([]string{xiaJia.ID()}, msg, "EnterCustomRoomACK"); err != nil {
			logger.Error("给下家发送自己进入房间时出错：", err, msg)
			return
		}

		msg = &rpc.EnterCustomRoomACK{}
		msg.SetRoomId(self.UID())
		msg.SetShangjiaType(2)
		msg.SetBReady(xiaJia.IsReady())
		msg.SetPlayerInfo(xiaJia.GetPlayerBasicInfo())
		msg.SetGameType(cmn.DaerGame)
		msg.SetTimes(self.times)
		msg.SetCurTimes(self.curTimes)
		msg.SetIsOwner(xiaJia.ID() == self.owner)
		msg.SetCurrencyType(self.currencyType)
		msg.SetDifen(self.GetDifen())
		if err := conn.SendCommonNotify2S([]string{player.ID()}, msg, "EnterCustomRoomACK"); err != nil {
			logger.Error("给自己发送下家进入房间时出错：", err, msg)
			return
		}

		logger.Info("sned EnterCustomRoomACK to xiaJia(%s).", xiaJia.ID())
	}
}

//向客户端发送玩家离开房间的消息
func (self *CustomDaerRoom) SendLeaveRoomACK(p cmn.Player, isChangeDesk bool) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*ds.DaerPlayer)
	if player == nil {
		logger.Error("player is nil.")
		return
	}

	logger.Info("发送离开房间：", player.ID())

	//给房间里的每个发送离开房间
	players := self.GetAllPlayer()
	for _, p := range players {
		if p != nil {
			msg := &rpc.LeaveCustomRoomACK{}
			msg.SetPlayerID(player.ID())
			if err := conn.SendCommonNotify2S([]string{p.ID()}, msg, "LeaveCustomRoomACK"); err != nil {
				logger.Error("发送离开房间出错：", err, msg)
				continue
			}
		}
	}
}

//开始游戏后
func (self *CustomDaerRoom) OnStartGameAfter() {
	//积分房在第一场开始后，才扣取房卡
	logger.Info("当前自建房状态：", self.currencyType, self.curTimes, self.creatingPlayer)
	if self.currencyType == CTCredits && self.curTimes <= 1 {
		costRoomCardCount := GetCostRoomCardCount(self.gameType, self.times)
		if costRoomCardCount > 0 {
			logger.Info("发送房卡扣取：", costRoomCardCount)
			if err := conn.SendCostResourceMsg(self.creatingPlayer, strconv.Itoa(cmn.CustomRoomCardID), cmn.GameTypeName[self.gameType], -costRoomCardCount); err != nil {
				logger.Error("发送扣取房卡出错：", err, self.creatingPlayer, costRoomCardCount)
			}
		}
	}
}

//玩家执行动作
//接收客户端发来的消息
func (self *CustomDaerRoom) OnPlayerDoAction(msg *rpc.ActionREQ) {
	//检查输入参数
	if msg == nil {
		logger.Error("CustomDaerRoom.OnPlayerDoAction:客户端发送来的数据为空！")
		return
	}

	//拦截准备的消息，其他的消息还是流到基类去  --出版本临时添加
	if self.currencyType == CTCoin {
		action := int32(msg.GetAction())
		if action == ds.AReady {
			uid := msg.GetPlayerID()
			p := self.DaerRoom.GetPlayerByID(uid)
			if p != nil && customRoomMgr != nil {
				if p.GetPlayerBasicInfo().GetCoin() <= 0 {
					customRoomMgr.OnLeaveGame(uid)
					return
				}
			} else {
				logger.Error("没有获取到到指定的玩家或则customRoomMgr is nil！")
			}
		}
	}

	self.DaerRoom.OnPlayerDoAction(msg)

	//检查是否是准备动作，如果是，则停止等待准备超时踢出玩家的倒计时
	// action := int32(msg.GetAction())
	// uid := msg.GetPlayerID()
	// if action == ds.AReady {
	// 	self.StopDelayCallback(uid)
	// } else if action == ds.ACancelReady {
	// 	self.StartForceLeave(uid)
	// }
}

//请求解散房间
func (self *CustomDaerRoom) OnJieSanRoom(uid string, msg *rpc.JieSanRoomREQ) {
	if msg == nil {
		logger.Error("参数为空！")
		return
	}

	switch msg.GetOperatorStatus() {
	case JSClaimer:
		if self.IsVoting() {
			logger.Error("已经有人提交解散房间申请了！")
			return
		}

		//如果只有发起者，则直接解散房间
		if self.GetPlayerAmount() == 1 && self.GetAllPlayer()[0].ID() == uid {
			self.ForceAllPlayerLeave()
			return
		}

		//初始化解散房间投票列表
		self.InitVoteList(uid, self.DaerRoom.GetAllPlayerIDs())
		//通知玩家有人请求解散房间
		self.SendJieSanRoomNotify(self.DaerRoom.GetAllPlayerIDs())
		//开启投票倒计时
		self.StartDelayCallback(JieSanRoomName, int64(self.voteDuration), func(data interface{}) {
			self.ForceAllPlayerLeave()
			//self.ClearVoteList()
		}, nil)

	case JSAgree:
		self.UpdateVote(uid, JSAgree)
		self.SendJieSanRoomUpdateStatusNotify(self.DaerRoom.GetAllPlayerIDs(), uid, JSAgree)
		if isEnd, isSuccess := self.IsVoteEnd(); isEnd && isSuccess {
			self.StopDelayCallback(JieSanRoomName)
			self.ForceAllPlayerLeave()
			//self.ClearVoteList()
		}
	case JSRefuse:
		self.StopDelayCallback(JieSanRoomName)
		self.SendJieSanRoomUpdateStatusNotify(self.DaerRoom.GetAllPlayerIDs(), uid, JSRefuse)
		self.ClearVoteList()
	}
}

//开启玩家不准备时自动踢出房间
func (self *CustomDaerRoom) StartForceLeave(uid string) {
	//没有等待准备的限制，无线等待
	if self.waitingReadyTime <= 0 {
		logger.Info("没有开启强制玩家离开的倒计时。。。。。。")
		return
	}

	self.StartDelayCallback(uid, int64(self.waitingReadyTime), func(data interface{}) {
		self.ForcePlayerLeave(data.(string))
	}, uid)
}

//开启房间解散倒计时
func (self *CustomDaerRoom) StartDissolveRoom() {
	if !self.IsStart() {
		if self.startWaitingDissolveTime <= 0 {
			logger.Info("没有开启解散房间的倒计时。。。。。。:", self.currencyType)
			return
		}

		self.StartDelayCallback(WaitingDissolveTimeName, int64(self.startWaitingDissolveTime), func(data interface{}) {
			self.ForceAllPlayerLeave()
		}, nil)
	} else {
		if self.middleWaitingDissolveTime <= 0 {
			logger.Info("没有开启解散房间的倒计时。。。。。。:", self.currencyType)
			return
		}

		self.StartDelayCallback(WaitingDissolveTimeName, int64(self.middleWaitingDissolveTime), func(data interface{}) {
			self.DoFinalJieSuan()
			self.ForceAllPlayerLeave()
		}, nil)
	}

}

//結算
func (self *CustomDaerRoom) DoJieSuan() {

	logger.Info("清扫自建房间准备下一场")

	//统计最后的名堂
	self.StatisticsMinTangForAll()

	//发送结算数据
	jieSuanPlayer, isHuangZhuang := self.SendJieSuanACKAndJieSuanCoinForAll()
	if !isHuangZhuang && !self.HavePlayerChaJiao() {
		if jieSuanPlayer != nil {
			self.SetHuPaiPlayerID(jieSuanPlayer.ID())
			logger.Info("设置胡牌的玩家：ID:%d  Name:%s", jieSuanPlayer.ID(), jieSuanPlayer.GetPlayerBasicInfo().GetName())
		} else {
			logger.Error("没有黄庄，怎么可能没有结算的玩家（赢家）")
		}
	}

	//重置数据
	self.ResetRoom()

	//开启解散房间的倒计时
	//self.StartDissolveRoom()

	//积分房检查是否结束，结束就直接进行最终结算
	if self.currencyType == CTCredits {
		//if self.IsEnd() || !self.IsOnlineForAll() {
		if self.IsEnd() {
			self.DoFinalJieSuan()
			self.ForceAllPlayerLeave()
		}
	}

	//当前比赛的场次
	self.curTimes++

	// else if self.currencyType == CTCoin {
	// 	self.KickOfflinePlayer()
	// } else {
	// 	logger.Error("为定义的结算货币类型")
	// }

	logger.Info("第%d场，结算！", self.curTimes)

	//启动房间停留计时器
	// self.StartDelayCallback(ds.RoomStayTimeName, ds.RoomStayTime, func() {
	// 	self.ForceAllPlayerLeave()
	// })
}

//终结结束
func (self *CustomDaerRoom) DoFinalJieSuan() {
	logger.Info("进行最终结算：=========")
	//便利所有玩家发送终结结算
	players := self.GetAllPlayer()
	if players == nil {
		logger.Error("最终结算时，房间里竟然没有玩家")
		return
	}

	//如果有投票的，那么最终结算时，最后一局是不能算在已经打的局数上的，因为最有一局没有打完
	if self.IsVoting() {
		self.curTimes--
	}

	//通知客服端
	for _, p := range players {
		if p == nil {
			continue
		}

		//发送数据
		logger.Info("最终结算时，向客服端发送最终结算消息")
		msg := &rpc.FinalJieSuanNotifyACK{}
		msg.SetJieSuanInfo(self.GetAddiData())
		if err := conn.SendCommonNotify2S([]string{p.ID()}, msg, "FinalJieSuanNotifyACK"); err != nil {
			logger.Error("发送自建房终极结算时出错：", err, msg)
			continue
		}
	}

	//设置标记
	self.isAlreadyFinalJieSuan = true
}

//检查是否有玩家已经离线了
func (self *CustomDaerRoom) IsOnlineForAll() (result bool) {
	//便利所有玩家
	players := self.GetAllPlayer()
	if players == nil {
		return false
	}

	//检查是否有离线的
	result = true
	for _, p := range players {
		if p == nil {
			return false
		}

		result = result && lockclient.IsOnline(p.ID())
	}

	return
}

//踢出不在线的玩家
func (self *CustomDaerRoom) KickOfflinePlayer() {
	//便利所有玩家
	players := self.GetAllPlayer()
	if players == nil {
		return
	}

	//检查是否有离线的
	for _, p := range players {
		if p == nil {
			continue
		}

		if !lockclient.IsOnline(p.ID()) {
			self.ForcePlayerLeave(p.ID())
		}
	}

}

//发送结算
func (self *CustomDaerRoom) SendJieSuanACKAndJieSuanCoinForAll() (jieSuanPlayer cmn.Player, isHuangZhuang bool) {

	mjp, isAllHu := self.GetMaxJieSuanPlayer()

	isHuangZhuang = mjp == nil || isAllHu
	jieSuanPlayer = mjp

	logger.Info("大贰自定义房间向所有发送结算信息: 结算Player：%s, AllHu:%s", mjp == nil, isAllHu)

	var coins []*rpc.JieSuanCoin = nil

	players := self.GetAllPlayer()

	//通知客服端
	for _, p := range players {
		if p == nil {
			continue
		}

		//发送数据
		addiData := self.GetAddiData()

		isCorrection := self.currencyType == CTCoin
		if hu, _ := p.IsHu(true); hu {
			coins = self.SendJieSuanACK(p.ID(), p, isHuangZhuang, isCorrection, addiData)
		} else {
			coins = self.SendJieSuanACK(p.ID(), jieSuanPlayer, isHuangZhuang, isCorrection, addiData)
		}
	}

	//扣取金币和统计金币
	if coins != nil {

		//统计金币信息
		self.StatisticsCoin(coins)

		//如果是金币结算，才需要扣取玩家身上的金币
		if self.currencyType == CTCoin {
			for _, p := range players {
				if p == nil {
					continue
				}
				p.JieSuanCoin(coins)
			}
		} else if self.currencyType == CTCredits {
			for uid, coin := range self.playerTotalCoin {
				p := self.GetPlayerByID(uid)
				if p != nil {
					p.GetPlayerBasicInfo().SetCoin(coin)
				} else {
					logger.Error("没有获取到指定玩家（%s）", uid)
				}
			}
		} else {
			logger.Error("未定义的货币结算类型")
		}
	}

	return
}

//获取结算的附加数据
func (self *CustomDaerRoom) GetAddiData() *rpc.JieSuanAdditionData {
	result := &rpc.JieSuanAdditionData{}
	result.SetSysType(cmn.ZiJianFang)
	result.SetStageEnd(self.IsEnd())
	//result.SetSuccess()
	result.Coin = append(result.Coin, self.GetTotalCoin()...)
	result.SetJieSuanTime(time.Now().Unix())
	result.SetCurTimes(self.curTimes)
	return result
}

//是否已经开场了
func (self *CustomDaerRoom) IsStart() bool {
	return self.IsGaming() || self.curTimes > 1
}

//是否最终结束了
func (self *CustomDaerRoom) IsEnd() bool {
	if self.isAlreadyFinalJieSuan {
		return true
	}

	if self.currencyType == CTCoin {
		return false
	}

	logger.Info("当前局数：%s  总局数：%s", self.curTimes, self.times)
	return self.curTimes >= self.times
}

//指定时间玩家都还没有开始游戏，那么将所有人提出房间
func (self *CustomDaerRoom) ForceAllPlayerLeave() {
	if customRoomMgr == nil {
		return
	}

	players := self.GetAllPlayer()

	for _, p := range players {
		if p != nil {
			customRoomMgr.OnLeaveGame(p.ID())
		}
	}
}

//强制玩家离开房间
func (self *CustomDaerRoom) ForcePlayerLeave(uid string) {
	if customRoomMgr == nil {
		return
	}

	players := self.GetAllPlayer()

	for _, p := range players {
		if p != nil && p.ID() == uid {
			customRoomMgr.OnLeaveGame(p.ID())
		}
	}
}
