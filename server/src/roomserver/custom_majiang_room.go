package roomserver

import (
	cmn "common"
	"logger"
	//	"math/rand"
	"rpc"

	conn "centerclient"
	"lockclient"
	mj "majiangserver"
	"strconv"
	"time"
)

const (
	DefaultMaJiangRoomType = 10
)

type CustomMaJiangRoom struct {
	mj.MaJiangRoom
	CustomRoom
}

//新建一个大贰自建房间
func NewCustomMaJiangRoom(uid int32, owner string, roomInfo *rpc.CreateRoomREQ) *CustomMaJiangRoom {
	r := new(CustomMaJiangRoom)

	//创建发牌器
	r.SetLicensingController(mj.NewLicensingController(roomInfo.GetTiYongAmount()))

	//初始化自建房信息
	r.InitCustomRoom( /*id,*/ owner, cmn.MaJiang, roomInfo)

	//初始化麻将放假
	r.Init(uid, DefaultMaJiangRoomType)

	//重置选择器接口
	r.SetSelector(r)

	//红中数量
	r.TotalHongZhongAmount = roomInfo.GetTiYongAmount()
	//最大倍数
	r.MaxMultiple = roomInfo.GetMaxMultiple()
	//起胡颗数
	r.QiHuKeAmount = roomInfo.GetQiHuKeAmount()
	//最大等待玩家执行的动作的时间（无限制等待）
	r.TimerInterval = 1000000000
	//游戏最大的执行时间
	r.GamingMaxTime = 2000000000

	switch r.currencyType {
	case CTCoin:
		cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(cmn.MaJiang)))
		if cfg != nil {
			r.RakeRate = cfg.RakeRate
		} else {
			logger.Error("读取房间配置表出错ID：%s", cmn.MaJiang)
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
func (self *CustomMaJiangRoom) Enter(p cmn.Player) {
	//检查输入参数
	if p == nil {
		logger.Error("player is nil.")
		return
	}

	player := p.(*mj.MaJiangPlayer)
	if player == nil {
		logger.Error("接口转换失败")
		return
	}

	//检查能否进入房间
	if self.IsFull() {
		logger.Error("MaJiangRoom.Enter: room is full")
		return
	}

	//进入房间时，将玩家的货币设置为初始化积分
	self.InitCurCredit(player)

	//修改玩家的房间相关的信息
	player.SetRoom(&self.MaJiangRoom)
	players := self.GetAllPlayer()

	for i := 0; i < mj.RoomMaxPlayerAmount; i++ {
		if players[i] == nil {
			players[i] = player
			self.SendEnterRoomACK(player)
			logger.Info("MaJiangRoom.Enter: player:%s enter room(%s):", player.ID(), i)
			break
		}
	}

	//开启一个房间停留计时器
	//self.StartForceLeave(p.ID())

	//开启解散房间的倒计时
	//self.StartDissolveRoom()
}

//重新进入房间
func (self *CustomMaJiangRoom) ReEnter(playerID string, playerInfo *rpc.PlayerBaseInfo) {
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
			self.SendJieSanRoomNotify(self.MaJiangRoom.GetAllPlayerIDs())
		}
	} else {
		logger.Error("MaJiangRoom:player not in the room")
	}
}

//计算玩家的当前积分
func (self *CustomMaJiangRoom) InitCurCredit(p cmn.Player) {
	if p == nil {
		logger.Error("p is null.")
		return
	}

	if self.currencyType == CTCredits {
		if cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(cmn.MaJiang))); cfg != nil {
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
func (self *CustomMaJiangRoom) Leave(uid string, isChangeDesk bool) bool {
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
	self.SetNextBankerPlayerID("")
	players := self.GetAllPlayer()

	//检查是否要提前最终的结算
	if self.currencyType == CTCredits && self.IsStart() && !self.isAlreadyFinalJieSuan {
		self.DoFinalJieSuan()
		self.ForceAllPlayerLeave()
	}

	for i := 0; i < mj.RoomMaxPlayerAmount; i++ {
		tempPlayer := players[i]
		if tempPlayer != nil && uid == tempPlayer.ID() {
			self.SendLeaveRoomACK(tempPlayer, isChangeDesk)
			tempPlayer.SetRoom(nil)
			players[i] = nil
			tempPlayer.Reset()
			logger.Info("CustomMaJiangRoom.Leave: player:%s Leave room:", tempPlayer.ID())
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
func (self *CustomMaJiangRoom) ChangeRoomOwner() {
	players := self.GetAllPlayer()

	for i := 0; i < mj.RoomMaxPlayerAmount; i++ {
		tempPlayer := players[i]
		if tempPlayer != nil {
			self.owner = tempPlayer.ID()
			logger.Info("CustomMaJiangRoom.ChangeRoomOwner: 改变房主到：%s", self.owner)
			break
		}
	}
}

func (self *CustomMaJiangRoom) SendEnterRoomACK(p cmn.Player) {
	logger.Info("自建房间的发送进入房间ACK")
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*mj.MaJiangPlayer)
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
	msg.SetLocation(int32(self.GetPlayerIndex(player)))
	msg.SetBReady(player.IsReady())
	msg.SetPlayerInfo(player.GetPlayerBasicInfo())
	msg.SetGameType(cmn.MaJiang)
	msg.SetTimes(self.times)
	msg.SetCurTimes(self.curTimes)
	msg.SetIsOwner(player.ID() == self.owner)
	msg.SetCurrencyType(self.currencyType)
	msg.SetDifen(self.GetDifen())
	msg.SetQiHuKeAmount(self.GetQiHuKeAmount())
	msg.SetTiYongAmount(self.GetTiYongAmount())
	if err := conn.SendCommonNotify2S(self.GetAllPlayerIDs(), msg, "EnterCustomRoomACK"); err != nil {
		logger.Error("给自己发送进入房间时出错：", err, msg)
		return
	}

	//发送在房间的其他玩家给自己
	players := self.GetAllPlayer()
	for _, rp := range players {
		if rp == nil || rp.ID() == player.ID() {
			continue
		}

		msg := &rpc.EnterCustomRoomACK{}
		msg.SetRoomId(self.UID())
		msg.SetLocation(int32(self.GetPlayerIndex(rp)))
		msg.SetBReady(rp.IsReady())
		msg.SetPlayerInfo(rp.GetPlayerBasicInfo())
		msg.SetGameType(cmn.MaJiang)
		msg.SetTimes(self.times)
		msg.SetCurTimes(self.curTimes)
		msg.SetIsOwner(rp.ID() == self.owner)
		msg.SetCurrencyType(self.currencyType)
		msg.SetDifen(self.GetDifen())
		msg.SetQiHuKeAmount(self.GetQiHuKeAmount())
		msg.SetTiYongAmount(self.GetTiYongAmount())
		if err := conn.SendCommonNotify2S([]string{player.ID()}, msg, "EnterCustomRoomACK"); err != nil {
			logger.Error("给自己发送进入房间时出错：", err, msg)
			return
		}
	}
}

//向客户端发送玩家离开房间的消息
func (self *CustomMaJiangRoom) SendLeaveRoomACK(p cmn.Player, isChangeDesk bool) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*mj.MaJiangPlayer)
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
func (self *CustomMaJiangRoom) OnStartGameAfter() {
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
func (self *CustomMaJiangRoom) OnPlayerDoAction(msg *rpc.ActionREQ) {
	//检查输入参数
	if msg == nil {
		logger.Error("CustomMaJiangRoom.OnPlayerDoAction:客户端发送来的数据为空！")
		return
	}

	//拦截准备的消息，其他的消息还是流到基类去  --出版本临时添加
	if self.currencyType == CTCoin {
		action := int32(msg.GetAction())
		if action == mj.AReady {
			uid := msg.GetPlayerID()
			p := self.MaJiangRoom.GetPlayerByID(uid)
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

	self.MaJiangRoom.OnPlayerDoAction(msg)
}

//请求解散房间
func (self *CustomMaJiangRoom) OnJieSanRoom(uid string, msg *rpc.JieSanRoomREQ) {
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
		self.InitVoteList(uid, self.MaJiangRoom.GetAllPlayerIDs())
		//通知玩家有人请求解散房间
		self.SendJieSanRoomNotify(self.MaJiangRoom.GetAllPlayerIDs())
		//开启投票倒计时
		self.StartDelayCallback(JieSanRoomName, int64(self.voteDuration), func(data interface{}) {
			self.ForceAllPlayerLeave()
			//self.ClearVoteList()
		}, nil)

	case JSAgree:
		self.UpdateVote(uid, JSAgree)
		self.SendJieSanRoomUpdateStatusNotify(self.MaJiangRoom.GetAllPlayerIDs(), uid, JSAgree)
		if isEnd, isSuccess := self.IsVoteEnd(); isEnd && isSuccess {
			self.StopDelayCallback(JieSanRoomName)
			self.ForceAllPlayerLeave()
			//self.ClearVoteList()
		}
	case JSRefuse:
		self.StopDelayCallback(JieSanRoomName)
		self.SendJieSanRoomUpdateStatusNotify(self.MaJiangRoom.GetAllPlayerIDs(), uid, JSRefuse)
		self.ClearVoteList()
	}
}

//开启玩家不准备时自动踢出房间
func (self *CustomMaJiangRoom) StartForceLeave(uid string) {
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
func (self *CustomMaJiangRoom) StartDissolveRoom() {
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
func (self *CustomMaJiangRoom) DoJieSuan() {

	logger.Info("清扫房间准备下一场")

	//积分房检查是否结束，结束就直接进行最终结算
	if self.currencyType == CTCredits {
		//计算原始的结算信息
		orginJieSuanCoin := self.CalcOrginJieSuanCoin()

		//统计金币（结算）信息
		self.StatisticsCoin(ConvertToJieSuanCoin(orginJieSuanCoin))

		//计算玩家身上的积分信息
		self.CalcCoinInPlayer()

		//发送结算数据
		self.SendJieSuanACKForAll(orginJieSuanCoin)

		//重置数据
		self.ResetRoom()

		//检查是否结束了
		if self.IsEnd() {
			self.DoFinalJieSuan()
			self.ForceAllPlayerLeave()
		}
	} else if self.currencyType == CTCoin {
		//结算金币
		jieSuanCoinInfo := self.CalcFinalJieSuanCoin()

		//发送金币扣取通知-通知gameserver扣取金币
		self.SendJieSuanCoinNotifyForAll(jieSuanCoinInfo)

		//发送结算数据
		self.SendJieSuanACKForAll(jieSuanCoinInfo)

		//重置数据
		self.ResetRoom()

		//检查所有玩家的金币信息是否还能继续进行下去
		self.CheckCoinForAll()
	} else {
		logger.Error("不支持的房间货币计算类型，只能是金币或积分！")
	}

	//当前比赛的场次
	self.curTimes++

	// //启动房间停留计时器
	// room.StartDelayCallback(RoomStayTimeName, room.RoomStayTime, func(data interface{}) {
	// 	room.rs.ForceAllPlayerLeave()
	// }, nil)

}

//终结结束
func (self *CustomMaJiangRoom) DoFinalJieSuan() {
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

//转换map到rpc.JieSuanCoin
func ConvertToJieSuanCoin(jieSuanCoinInfo map[string]int32) (result []*rpc.JieSuanCoin) {
	result = make([]*rpc.JieSuanCoin, 0)

	if jieSuanCoinInfo == nil {
		logger.Error("jieSuanCoinInfo is nil.")
		return
	}

	for id, coin := range jieSuanCoinInfo {
		temp := &rpc.JieSuanCoin{}
		temp.SetPlayerID(id)
		temp.SetCoin(coin)
		result = append(result, temp)
	}

	return
}

//计算并设置玩家身上的积分信息
func (self *CustomMaJiangRoom) CalcCoinInPlayer() {
	for uid, coin := range self.playerTotalCoin {
		p := self.GetPlayerByID(uid)
		if p != nil {
			p.GetPlayerBasicInfo().SetCoin(coin)
		} else {
			logger.Error("没有获取到指定玩家（%s）", uid)
		}
	}
}

//检查是否有玩家已经离线了
func (self *CustomMaJiangRoom) IsOnlineForAll() (result bool) {
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
func (self *CustomMaJiangRoom) KickOfflinePlayer() {
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

//发送游戏结算
func (self *CustomMaJiangRoom) SendJieSuanACKForAll(jieSuanCoin map[string]int32) {

	logger.Info("向所有发送结算信息")

	players := self.GetAllPlayer()
	if players == nil {
		return
	}

	addiData := self.GetAddiData()
	for _, p := range players {
		if p == nil {
			continue
		}

		//发送数据
		p.SendJieSuanACK(jieSuanCoin, addiData)
	}

	return
}

//获取结算的附加数据
func (self *CustomMaJiangRoom) GetAddiData() *rpc.JieSuanAdditionData {
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
func (self *CustomMaJiangRoom) IsStart() bool {
	return self.IsGaming() || self.curTimes > 1
}

//是否最终结束了
func (self *CustomMaJiangRoom) IsEnd() bool {
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
func (self *CustomMaJiangRoom) ForceAllPlayerLeave() {
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
func (self *CustomMaJiangRoom) ForcePlayerLeave(uid string) {
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
