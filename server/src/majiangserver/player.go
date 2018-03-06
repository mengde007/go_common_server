package majiangserver

import (
	conn "centerclient"
	cmn "common"
	"connector"
	//"fmt"
	"logger"
	"math"
	"rpc"
	"runtime/debug"
	//"strconv"
)

type MaJiangPlayer struct {
	id                   string
	ptype                int32 //庄家，闲家
	cards                []*MaJiangCard
	chuCards             []*MaJiangCard
	huCard               *MaJiangCard //胡的牌
	showPatterns         []*MaJiangPattern
	controller           *MaJiangController
	aroundState          *PlayerAroundState
	cardAmountInfo       *CardAmountStatistics
	isReady              bool
	isChaJiaoHu          bool //是不是查叫时候才胡的牌
	beiHuPlayers         []*MaJiangPlayer
	room                 *MaJiangRoom
	multipleCount        map[int32]int32
	watingAction         []int32
	readyDoAction        int32 //准备执行的动作，玩家点击了吃，但是他的上家还在准备吃，当上家放弃吃的时候，用于表示自己可以吃
	delayDoAction        int32 //需要延迟到一下个阶段才生效的动作（现在只有非庄家的报或摆）
	mode                 int32 //自动/手动模式
	IsOpenHongZhongCheck bool  //是否开启了贴鬼碰

	client *rpc.PlayerBaseInfo

	//记录下发给玩家的动作通知
	sendedActionMsg *rpc.MJActionNotifyACK
}

func NewMaJiangPlayer(id string, selfInfo *rpc.PlayerBaseInfo) *MaJiangPlayer {
	p := new(MaJiangPlayer)
	p.id = id
	p.client = selfInfo
	p.IsOpenHongZhongCheck = true
	p.controller = NewController(p)
	p.aroundState = NewPlayerAroundState()
	p.cardAmountInfo = NewCardAmountStatisticsByCards([]*MaJiangCard{}, true)

	// p.ptype = cmn.PTNormal
	// self.mode = cmn.Manual
	// p.watingAction = []int32{}
	// p.readyDoAction = ANone
	// p.delayDoAction = ANone
	// p.isChaJiaoHu = false
	// p.beiHuPlayers = make([]*MaJiangPlayer, 0)
	p.Reset()

	if p.client == nil {
		logger.Error("self:Newself, selfInfo is nil.")
	}

	return p
}

//重置玩家
func (self *MaJiangPlayer) Reset() {
	self.ptype = cmn.PTNormal
	self.cards = make([]*MaJiangCard, 0)
	self.chuCards = make([]*MaJiangCard, 0)
	self.huCard = nil
	self.showPatterns = make([]*MaJiangPattern, 0)
	self.isReady = false
	self.isChaJiaoHu = false
	self.beiHuPlayers = make([]*MaJiangPlayer, 0)
	self.multipleCount = make(map[int32]int32, 0)
	self.aroundState.ClearAll()
	self.cardAmountInfo.Reset()
	self.watingAction = []int32{}
	self.readyDoAction = ANone
	self.delayDoAction = ANone
	self.mode = cmn.Manual
	self.sendedActionMsg = nil
}

//组牌
func (self *MaJiangPlayer) Compose(cards []*MaJiangCard) {
	//检查参数的合法性
	if cards == nil || len(cards) <= 0 {
		logger.Error("Compose:cards is nil or empty")
		return
	}

	if self.controller == nil {
		logger.Error("self.controller is nil.")
		return
	}

	//保存发的牌
	self.cards = cards

	//初始化手牌状态
	self.InitCardStatus()

	//统计并缓存卡牌数量
	self.cardAmountInfo.CalcCardAmountByCards(self.cards, false)
}

//设置牌的初始化状态
func (self *MaJiangPlayer) InitCardStatus() {
	for _, card := range self.cards {
		card.owner = self
		card.flag = cmn.CBack
	}
}

//指定玩家执行动作
func (self *MaJiangPlayer) PlayerDoAction(action int32, card *MaJiangCard) {

	//检查能否执行这个动作
	if !self.CanDoAction(action) {
		logger.Info("等待的动作:%s 和 执行的动作:%s不相同", CnvtActsToStr(self.watingAction), action)
		return
	}

	room := self.room

	//执行动作时检查是否需要清除杠上花和杠上炮的标志,注意：自动过相当于执行，因为在自动状态下，过牌是直接执行默认操作的
	self.ClearGangShangHuaAndGangShangPaoFlag(action)

	switch action {
	case AReady: //准备
		if room.state != RSReady || self.isReady {
			self.SendActionACK(action, nil, nil, AOccursError)
			return
		}

		//准备的时候检查一下金币，结算后金币会变化
		if maJiangRoomMgr != nil {
			if ok, _ := cmn.CheckCoin(room.rtype, self.client); !ok {
				maJiangRoomMgr.LeaveGame(self.id, false)
				return
			}
		}

		self.isReady = true
		self.SendActionACK(action, nil, nil, ACSuccess)
		if room.CanStartGame() {
			room.StartGame()
		}
		logger.Info("PlayerDoAction: 准备:", self.client.GetName())

	case ACancelReady:
		if room.state != RSReady || !self.isReady {
			self.SendActionACK(action, nil, nil, AOccursError)
			return
		}

		self.isReady = false
		self.SendActionACK(action, nil, nil, ACSuccess)
		logger.Info("PlayerDoAction: 取消准备:", self.client.GetName())

	case ATuoGuan: //托管
		if !room.IsGaming() || self.mode == cmn.Auto {
			self.SendActionACK(action, nil, nil, AOccursError)
			return
		}
		self.SwitchControllMode(cmn.Auto)
		self.SendActionACK(action, nil, nil, ACSuccess)

		//有等待的动作就执行了
		if room.IsGaming() && self.HaveWaitingDoAction() && IsWaitingAction(self.watingAction) {
			self.PlayerDoAction(AGuo, room.activeCard)
		}
		// //如果在出牌时发生错误，那么就要检查是否客服端已经执行了这个动作及self.readyDoAction不能ANone
		// if self.readyDoAction != ANone {
		// 	if self.readyDoAction == AChu {
		// 		self.PlayerDoAction(AGuo, room.activeCard)
		// 	} else {
		// 		logger.Error("托管时，还有准备执行的动作：", actionName[self.readyDoAction])
		// 	}
		// }

		logger.Info("PlayerDoAction: 托管:", self.client.GetName())

	case ACancelTuoGuan: //取消托管
		if !room.IsGaming() || self.mode == cmn.Manual {
			self.SendActionACK(action, nil, nil, AOccursError)
			return
		}

		self.SwitchControllMode(cmn.Manual)
		self.SendActionACK(action, nil, nil, ACSuccess)

		logger.Info("取消托管时的准备执行的动作：", CnvtActsToStr(self.watingAction))
		if self.HaveWaitingDoAction() && IsWaitingAction(self.watingAction) {
			delayCallId := self.id + CnvtActsToStr(self.watingAction)
			room.StopDelayCallback(delayCallId)
			logger.Info("取消托管时停止自动执行：", delayCallId)
			room.StartTimer(room.TimerInterval)

			room.sendCountdownNotifyACK()
		}

		logger.Info("PlayerDoAction: 取消托管:", self.client.GetName())

	case AHu:
		fallthrough
	case AMingGang:
		fallthrough
	case ATieMingGang:
		fallthrough
	case APeng:
		fallthrough
	case ATiePeng:
		self.SwitchReadyDoAction(action)
		self.DoHuPengGangAfter(card, false)
		logger.Info("PlayerDoAction: 胡,明杠，碰:", self.client.GetName(), card.value)

	case ABao:
		self.SwitchReadyDoAction(ABao)
		self.DoBaoAfter(false)
		logger.Info("PlayerDoAction: 报牌:", self.client.GetName())
	case AAnGang:
		self.SwitchReadyDoAction(AAnGang)
		self.DoAnGangAfter(card, false)
		logger.Info("PlayerDoAction: 暗杠:", self.client.GetName(), card.value)
	case ABuGang:
		self.SwitchReadyDoAction(ABuGang)
		self.DoBuGangAfter(card, false)
		logger.Info("PlayerDoAction: 补杠:", self.client.GetName(), card.value)
	case AChu:
		self.DoChuAfter(card)

		logger.Info("PlayerDoAction: 出牌:", self.client.GetName(), card.value)
	case AGuo:
		self.DoGuo(card)
		logger.Info("PlayerDoAction: 过:", self.client.GetName())
	default:
		logger.Error("执行的动作是否有问题：", action)
	}

	if action != AReady && action != ACancelReady && action != AReady && action != ACancelReady {
		logger.Info("===================执行动作后的情况===========================")
		PrintRoom(room)
	}
}

func (self *MaJiangPlayer) ClearGangShangHuaAndGangShangPaoFlag(curDoAction int32) {
	//执行动作时检查是否需要清除杠上花和杠上炮的标志,注意：自动过相当于执行，因为在自动状态下，过牌是直接执行默认操作的
	if IsActionInFight(curDoAction) {
		if self.mode == cmn.Auto {
			if curDoAction != AHu && (curDoAction == AGuo &&
				self.HaveWaitingDoAction() && self.watingAction[0] != AHu) {
				self.aroundState.checkGangShangHuaCard = nil
			}

			if curDoAction != AChu && (curDoAction == AGuo &&
				self.HaveWaitingDoAction() && self.watingAction[0] != AChu) {
				self.aroundState.checkGangShangPaoCard = nil
			}
		} else {
			if curDoAction != AHu {
				self.aroundState.checkGangShangHuaCard = nil
			}

			if curDoAction != AChu {
				self.aroundState.checkGangShangPaoCard = nil
			}
		}
	}
}

// func (self *MaJiangPlayer) ClearBuGanggFlag(curDoAction int32) {

// 	//检查参数
// 	room := self.room
// 	if room == nil {
// 		logger.Error("room is nil.")
// 		return
// 	}

// 	//执行动作时检查是否需要清除补杠的标志,注意：自动过相当于执行，因为在自动状态下，过牌是直接执行默认操作的
// 	if IsActionInFight(curDoAction) {

// 		//补杠玩家自己执行任何动作都会清除补杠标志
// 		haveBuGang, buGangPlayer := room.HaveBuGangFlag()
// 		if haveBuGang && buGangPlayer != nil && buGangPlayer.id == self.id {
// 			room.ClearBuGangFlag()
// 			return
// 		}

// 		if self.mode == cmn.Auto {
// 			if curDoAction != AHu && (curDoAction == AGuo &&
// 				self.HaveWaitingDoAction() && self.watingAction[0] != AHu) {
// 				room.ClearBuGangFlag()
// 			}
// 		} else {
// 			if curDoAction != AHu {
// 				room.ClearBuGangFlag()
// 			}
// 		}
// 	}
// }

func (self *MaJiangPlayer) CanDoAction(action int32) bool {

	//检查参数的正确性
	room := self.room
	if room == nil {
		logger.Error("玩家没有所属的房间。")
		return false
	}

	if IsWaitingAction([]int32{action}) && !Exist(self.watingAction, action) {
		logger.Error("等待的动作和执行动作不相同！W:%s   E:%s", CnvtActsToStr(self.watingAction), actionName[action])
		return false
	}

	//玩家报了牌只能胡
	if self.HaveBao() {
		if action == APeng || action == ATiePeng || action == ABuGang || action == ABao {
			return false
		}

		if action == AGuo {
			//只能过这些动作
			if !(Exist(self.watingAction, AHu) || Exist(self.watingAction, AAnGang) ||
				Exist(self.watingAction, AMingGang) || Exist(self.watingAction, ATieMingGang) ||
				Exist(self.watingAction, AChu)) {
				return false
			}
		}
	}

	return true
}

//玩家对胡,碰，明杠动作选择对应操作后执行
func (self *MaJiangPlayer) DoHuPengGangAfter(card *MaJiangCard, isGuo bool) bool {

	//检测是不是在一个房间里
	room := self.room
	if room == nil {
		return false
	}

	if card == nil {
		logger.Error("不能胡一个空牌")
		return false
	}

	//1 如果是过牌,缓存
	if isGuo {
		//记录当前过牌信息（用于检测升值，过水等）
		self.CacheGuoPaiInfo(card)

		//通知客服单过牌成功
		self.SendActionACK(AGuo, nil, nil, ACSuccess)

		//重置所有动作
		self.ResetDoAction()

		//清除补杠玩家的readyDoAction,如果有的话
		if haveBuGang, buGangPlayer := room.HaveBuGangFlag(); haveBuGang && buGangPlayer != nil {
			if buGangPlayer.readyDoAction == ABuGang {
				buGangPlayer.readyDoAction = ANone
			}
		}

	}

	//2. 检测并执行胡,碰，明杠
	var success bool = false
	var end bool = false
	var py *MaJiangPlayer = nil
	if success, end, py = room.DoReadyActionByOrder(); success {

		//2.1 执行特定动作
		readyDoAction := py.readyDoAction
		switch readyDoAction {
		case AHu:
			py.controller.Hu(card)

			room.curAroundState.AddPlayerOfHu(py)

			py.ResetDoAction()

			py.SendActionACK(AHu, card, nil, ACSuccess)

			//被胡的牌，要从出牌的玩家的出牌队列里移除
			if card.owner != nil && card.owner.id != py.id {
				card.owner.RemoveChuCard(card)
			}

			//检查是是否是补杠的牌，如果是且🈶没有从手上移除，那么从手上移除（只是通知玩家，真正的移除是在所有玩家胡牌后）
			if haveBuGang, buGangPlayer := room.HaveBuGangFlag(); haveBuGang && buGangPlayer != nil {
				if !buGangPlayer.aroundState.buGangCardRemoved {
					buGangPlayer.aroundState.buGangCardRemoved = true
					buGangPlayer.SendRemoveCardNotifyACK(buGangPlayer.aroundState.buGangCard, true)
				}
			}

		case AMingGang:
			fallthrough
		case ATieMingGang:
			gangPattern := py.controller.MingGang(card)

			room.ChangeActivePlayerTo(py)

			py.SendActionACK(py.readyDoAction, card, gangPattern, ACSuccess)

			room.ResetAllAction(true)

			//被杠的牌，要从出牌的玩家的出牌队列里移除
			if card.owner != nil {
				card.owner.RemoveChuCard(card)
			}

			room.CheckDoAction(nil, nil, nil, false)

		case APeng:
			fallthrough
		case ATiePeng:
			pengPattern := py.controller.Peng(card)

			room.ChangeActivePlayerTo(py)

			py.SendActionACK(py.readyDoAction, card, pengPattern, ACSuccess)

			room.ResetAllAction(true)

			py.SendActionNotifyACK(card, []int32{AChu}, nil)

			//被碰的牌，要从出牌的玩家的出牌队列里移除
			if card.owner != nil {
				card.owner.RemoveChuCard(card)
			}
		default:
			logger.Error("不能在这个函数里执行其他动作")
		}

		//2.2 动作执行完了后，清场
		if readyDoAction == AHu {

			logger.Info("执行动作后，是否结束了这轮动作执行：%s, 有胡牌玩家：%s", end, room.curAroundState.HaveHuPlayer())
			//2. 是不是胡牌导致这轮检查结束
			if room.curAroundState.HaveHuPlayer() && end {
				//是否是第一轮胡牌，并确定下一轮庄家
				self.CheckFirstAroundHuAndDecideBanker(card)

				//如果有缓存的摸牌，那么清除
				if haveBuGang, buGangPlayer := room.HaveBuGangFlag(); haveBuGang && buGangPlayer != nil {
					//补杠的牌是以前在手上的牌，还是刚摸上来的牌
					if buGangPlayer.aroundState.buGangCard.IsEqual(buGangPlayer.aroundState.moCard) {
						room.DoMoByCache(true)
					} else {
						cType, value := buGangPlayer.aroundState.buGangCard.CurValue()
						if _, removedCards := buGangPlayer.RemoveHandCards(cType, value, 1); len(removedCards) != 1 {
							logger.Error("移除补杠玩家(%s)手上的牌失败！补杠的牌：%s  摸的牌：%s",
								buGangPlayer.id, ConvertToWord(buGangPlayer.aroundState.buGangCard),
								buGangPlayer.aroundState.moCard)
						}
					}

					room.ClearBuGangFlag()
				}

				//被胡的牌，要从出牌的玩家的出牌队列里移除
				if card.owner != nil {
					card.owner.RemoveChuCard(card)
				}

				//检查这一把是否结束
				if room.IsOverForAround() {
					room.ResetAllAction(true)

					room.SwitchRoomState(RSSettlement)

					room.curAroundState.ClearHuPlayers()

					room.CheckDoAction(nil, nil, nil, false)
				} else {
					//确定下一个活动玩家
					nextActivePalyer := room.GetNextActivePlayerByHuPlayers(room.curAroundState.huPlayers)
					room.ChangeActivePlayerTo(nextActivePalyer)

					//重置所有玩家的动作
					room.ResetAllAction(true)

					//清除这一轮胡牌的玩家列表
					room.curAroundState.ClearHuPlayers()

					room.CheckDoAction(nil, nil, nil, false)
				}
			}

		} else {
			//如果有缓存的摸牌，那么清除
			room.DoMoByCache(true)
		}

		return true
	}

	//3. 没有人执行动作
	if _, have := room.GetWatingActionPlayer([]int32{AHu, AMingGang, ATieMingGang, APeng, ATiePeng}); !have {

		logger.Info("没有人执行动作!")
		//如果有缓存的摸牌，那么将牌摸到手上
		room.DoMoByCache(false)

		room.ResetAllAction(true)

		if room.state == RSBankerTianHuStage {
			room.SwitchRoomState(RSNotBankerBaoPaiStage)

			room.CheckDoAction(nil, nil, nil, false)
		} else if room.state == RSLoopWorkStage {
			if haveBuGang, buGangPlayer := room.HaveBuGangFlag(); haveBuGang && buGangPlayer != nil {
				logger.Info("没有人执行动作时的卡牌是：%s, 检查补杠：%s , 补杠的玩家是：%s", ConvertToWord(card), haveBuGang, buGangPlayer.client.GetName())
				buGangPlayer.ModifyDataForBuGang(card)

				room.ClearBuGangFlag()
			} else {
				logger.Info("没有人执行动作时的卡牌是：%s, 这牌的拥有者是：%s", ConvertToWord(card), card.owner.client.GetName())
				ap := room.GetActivePlayer()
				if (ap != nil && ap.id == card.owner.id) || card.IsIncomeCard() {
					ap.SendActionNotifyACK(card, []int32{AChu}, nil)
				} else {
					//room.ChangeActivePlayerToNext()

					room.CheckDoAction(nil, nil, nil, false)
				}
			}

		} else {
			logger.Error("胡,明杠，碰时不应该处理其他状态的情况")
		}

	} else {
		if !isGuo && self.readyDoAction != ANone {
			self.SendActionACK(self.readyDoAction, nil, nil, ACWaitingOtherPlayer)
		}
	}

	return false
}

// func (self *MaJiangPlayer) DoHuPengGangAfter(card *MaJiangCard, isGuo bool) bool {

// 	//检测是不是在一个房间里
// 	room := self.room
// 	if room == nil {
// 		return false
// 	}

// 	if card == nil {
// 		logger.Error("不能胡一个空牌")
// 		return false
// 	}

// 	//1. 检测并执行胡,碰，明杠
// 	var success bool = false
// 	var end bool = false
// 	var py *MaJiangPlayer = nil
// 	if success, end, py = room.DoReadyActionByOrder(); success {

// 		readyDoAction := py.readyDoAction
// 		switch readyDoAction {
// 		case AHu:
// 			py.controller.Hu(card)

// 			room.curAroundState.AddPlayerOfHu(py)

// 			py.ResetDoAction()

// 			py.SendActionACK(AHu, card, nil, ACSuccess)

// 			//被胡的牌，要从出牌的玩家的出牌队列里移除
// 			if card.owner != nil && card.owner.id != py.id {
// 				card.owner.RemoveChuCard(card)
// 			}

// 		case AMingGang:
// 			fallthrough
// 		case ATieMingGang:
// 			gangPattern := py.controller.MingGang(card)

// 			room.ChangeActivePlayerTo(py)

// 			py.SendActionACK(py.readyDoAction, card, gangPattern, ACSuccess)

// 			room.ResetAllAction(true)

// 			//被杠的牌，要从出牌的玩家的出牌队列里移除
// 			if card.owner != nil {
// 				card.owner.RemoveChuCard(card)
// 			}

// 			room.CheckDoAction(nil, nil, nil, false)

// 		case APeng:
// 			fallthrough
// 		case ATiePeng:
// 			pengPattern := py.controller.Peng(card)

// 			room.ChangeActivePlayerTo(py)

// 			py.SendActionACK(py.readyDoAction, card, pengPattern, ACSuccess)

// 			room.ResetAllAction(true)

// 			py.SendActionNotifyACK(card, []int32{AChu}, nil)

// 			//被碰的牌，要从出牌的玩家的出牌队列里移除
// 			if card.owner != nil {
// 				card.owner.RemoveChuCard(card)
// 			}
// 		}

// 		//执行了其他动作
// 		if readyDoAction != AHu {
// 			//如果有缓存的摸牌，那么清除
// 			room.DoMoByCache(true)
// 			return true
// 		}
// 	}

// 	logger.Info("执行动作后，是否结束了这轮动作执行：%s, 有胡牌玩家：%s", end, room.curAroundState.HaveHuPlayer())
// 	//2. 是不是胡牌导致这轮检查结束
// 	if room.curAroundState.HaveHuPlayer() && end {
// 		//是否是第一轮胡牌，并确定下一轮庄家
// 		self.CheckFirstAroundHuAndDecideBanker(card)

// 		//如果有缓存的摸牌，那么清除
// 		room.DoMoByCache(true)

// 		//被胡的牌，要从出牌的玩家的出牌队列里移除
// 		if card.owner != nil {
// 			card.owner.RemoveChuCard(card)
// 		}

// 		//检查这一把是否结束
// 		if room.IsOverForAround() {
// 			room.ResetAllAction(true)

// 			room.SwitchRoomState(RSSettlement)

// 			room.curAroundState.ClearHuPlayers()

// 			room.CheckDoAction(nil, nil, nil, false)
// 		} else {
// 			//确定下一个活动玩家
// 			nextActivePalyer := room.GetNextActivePlayerByHuPlayers(room.curAroundState.huPlayers)
// 			room.ChangeActivePlayerTo(nextActivePalyer)

// 			//重置所有玩家的动作
// 			room.ResetAllAction(true)

// 			//清除这一轮胡牌的玩家列表
// 			room.curAroundState.ClearHuPlayers()

// 			room.CheckDoAction(nil, nil, nil, false)
// 		}

// 		return true
// 	}

// 	//3. 是不是过
// 	if isGuo {
// 		//记录当前过牌信息（用于检测升值，过水等）
// 		self.CacheGuoPaiInfo(card)

// 		//通知客服单过牌成功
// 		self.SendActionACK(AGuo, nil, nil, ACSuccess)

// 		//重置所有动作
// 		self.ResetDoAction()

// 	}

// 	//4. 没有人执行动作
// 	if _, have := room.GetWatingActionPlayer([]int32{AHu, AMingGang, ATieMingGang, APeng, ATiePeng}); !have {

// 		//如果有缓存的摸牌，那么将牌摸到手上
// 		room.DoMoByCache(false)

// 		room.ResetAllAction(true)

// 		if room.state == RSBankerTianHuStage {
// 			room.SwitchRoomState(RSNotBankerBaoPaiStage)

// 			room.CheckDoAction(nil, nil, nil, false)
// 		} else if room.state == RSLoopWorkStage {
// 			if haveBuGang, buGangPlayer := room.HaveBuGangFlag(); haveBuGang && buGangPlayer != nil {
// 				buGangPlayer.ModifyDataForBuGang(card)

// 				room.ClearBuGangFlag()
// 			} else {
// 				if card.IsIncomeCard() {
// 					ap := room.GetActivePlayer()
// 					if ap != nil {
// 						ap.SendActionNotifyACK(card, []int32{AChu}, nil)
// 					} else {
// 						logger.Error("竟然没有活动玩家")
// 					}
// 				} else {
// 					//room.ChangeActivePlayerToNext()

// 					room.CheckDoAction(nil, nil, nil, false)
// 				}
// 			}

// 		} else {
// 			logger.Error("胡,明杠，碰时不应该处理其他状态的情况")
// 		}

// 	} else {
// 		if !isGuo && self.readyDoAction != ANone {
// 			self.SendActionACK(self.readyDoAction, nil, nil, ACWaitingOtherPlayer)
// 		}
// 	}

// 	return false
// }

//缓存过牌信息
func (self *MaJiangPlayer) CacheGuoPaiInfo(card *MaJiangCard) {
	//检查输入参数
	if card == nil {
		logger.Error("MjiangPlayer.CacheGuoPaiInfo: card is nil.")
		return
	}

	//红中不能过
	if card.IsHongZhong() {
		return
	}

	//缓存数据
	for _, wa := range self.watingAction {
		switch wa {
		case AHu:
			//isZiMo := card.owner == nil || card.owner.id == self.id
			//if !isZiMo {
			if self.HaveBao() {
				self.aroundState.huKe = -1
			} else {
				self.aroundState.huKe, _ = self.GetMaxHuOfPatternGroupByCard(card)
			}
			//}
		case AMingGang:
			fallthrough
		case ATieMingGang:
			fallthrough
		case APeng:
			fallthrough
		case ATiePeng:
			self.aroundState.AddGuoShuiPengGangCard(card)
		}
	}

	//logger.Error("玩家：%s 过牌后，缓存的过水和升值胡情况：等待动作:%s, 当前过的升值颗数：%d, 当前过的碰杠牌是：%s", self.client.GetName(), CnvtActsToStr(self.watingAction), self.aroundState.huKe, ConvertToWord(self.aroundState.guoPengGangCard))
}

//检查是否是第一局胡牌， 并确定下把的庄家
func (self *MaJiangPlayer) CheckFirstAroundHuAndDecideBanker(card *MaJiangCard) {
	//检测是不是在一个房间里
	room := self.room
	if room == nil {
		return
	}

	_, huPlayers := room.GetPlayerOfHu()
	curAroundhHuAmount, curAroundHuPlayers := room.curAroundState.GetPlayerOfHu()
	isFirstAround := IsSamePlayersList(huPlayers, curAroundHuPlayers)
	if isFirstAround {
		//一炮单响
		if curAroundhHuAmount == 1 {
			room.SetNextBankerPlayerID(curAroundHuPlayers[0].id)
		}

		//一炮多响
		if curAroundhHuAmount > 1 {
			if card.owner == nil {
				logger.Error("此张胡牌不知道是谁打的，所以不能点炮！")
				return
			}

			room.SetNextBankerPlayerID(card.owner.id)
		}
	}
}

//报阶段完成后要执行的操作
func (self *MaJiangPlayer) DoBaoAfter(isGuo bool) {
	//检测是不是在一个房间里
	room := self.room
	if room == nil {
		logger.Error("room is nil")
		return
	}

	if self.mode == cmn.Manual && isGuo {
		self.SendActionACK(AGuo, nil, nil, ACSuccess)

		self.ResetDoAction()
	} else {
		//执行报
		if self.readyDoAction == ABao {
			self.controller.Bao()

			self.SendActionACK(ABao, nil, nil, ACSuccess)

			self.ResetDoAction()
		}
	}

	//没有等待报牌的玩家吗
	if _, have := room.GetWatingActionPlayer([]int32{ABao}); !have {

		room.ResetAllAction(true)

		if room.state == RSNotBankerBaoPaiStage {
			banker := room.GetBanker()
			if banker == nil {
				logger.Error("DoBaoAfter:竟然没有庄家！太不可思议了")
				return
			}
			//庄家天胡了
			logger.Info("在非庄家报阶段，庄家是否胡牌了：", banker.IsHu())
			PrintCardS("胡的牌：", banker.huCard)
			PrintCardsS("手牌：", banker.cards)
			PrintPatternsS("显示的模式组：", banker.showPatterns)
			if banker.IsHu() {
				room.SwitchRoomState(RSLoopWorkStage)
				room.CheckDoAction(nil, nil, nil, false)
			} else {
				room.SwitchRoomState(RSBankerChuPaiStage)
				banker.SendActionNotifyACK(nil, []int32{AChu}, nil)
			}
		} else if room.state == RSBankerBaoPaiStage {
			room.SwitchRoomState(RSLoopWorkStage)

			room.ChangeActivePlayerToNext()

			room.CheckDoAction(room.activeCard, nil, []*MaJiangPlayer{room.GetBanker()}, false)
		} else {
			logger.Error("报牌时不应该处理其他状态的情况")
		}
	}
}

//执行暗杠
func (self *MaJiangPlayer) DoAnGangAfter(card *MaJiangCard, isGuo bool) {
	//检测是不是在一个房间里
	room := self.room
	if room == nil {
		return
	}

	if card == nil {
		logger.Error("不能招一个空牌")
		return
	}

	//如果有缓存的摸牌，那么将牌摸到手上
	room.DoMoByCache(false)

	if self.mode == cmn.Manual && isGuo {
		self.SendActionACK(AGuo, nil, nil, ACSuccess)

		self.ResetDoAction()

		self.SendActionNotifyACK(card, []int32{AChu}, nil)
	} else {
		anGangPattern := self.controller.AnGang(card)

		room.ChangeActivePlayerTo(self)

		room.ResetAllAction(true)

		self.SendActionACK(AAnGang, card, anGangPattern, ACSuccess)

		//在CheckDoAction中进行的摸牌
		room.CheckDoAction(nil, self, nil, false)

	}
}

//执行补杠
func (self *MaJiangPlayer) DoBuGangAfter(card *MaJiangCard, isGuo bool) {
	//检测是不是在一个房间里
	room := self.room
	if room == nil {
		return
	}

	if card == nil {
		logger.Error("不能招一个空牌")
		return
	}

	//如果有缓存的摸牌，那么将牌摸到手上
	room.DoMoByCache(false)

	if self.mode == cmn.Manual && isGuo {

		self.SendActionACK(AGuo, nil, nil, ACSuccess)

		self.ResetDoAction()

		self.SendActionNotifyACK(card, []int32{AChu}, nil)
	} else {

		//检查能否抢杠
		actions := room.CheckCanDoActionAndNotifyPlayer(card, nil, []*MaJiangPlayer{self}, true)
		canQiangGang := Exist(actions, AHu)
		if canQiangGang {
			logger.Error("抢杠：", ConvertToWord(card))
			card.owner = self
			self.aroundState.buGangCard = card
			room.activeCard = card

			self.SendActionACK(ABuGang, card, nil, ACWaitingOtherPlayer)

		} else {
			self.ModifyDataForBuGang(card)
		}
	}
}

//执行补杠
func (self *MaJiangPlayer) ModifyDataForBuGang(card *MaJiangCard) {
	//检测是不是在一个房间里
	room := self.room
	if room == nil {
		return
	}

	if card == nil {
		logger.Error("不能招一个空牌")
		return
	}

	buCard, buGangPattern := self.controller.BuGang(card)

	room.ChangeActivePlayerTo(self)

	room.ResetAllAction(true)

	self.SendActionACK(ABuGang, buCard, buGangPattern, ACSuccess)

	room.CheckDoAction(nil, self, nil, false)
	//self.SendActionNotifyACK(nil, []int32{AChu}, nil)
}

//执行出牌
func (self *MaJiangPlayer) DoChuAfter(card *MaJiangCard) {
	//检测是不是在一个房间里
	room := self.room
	if room == nil {
		logger.Error("room is nil")
		return
	}

	if card == nil {
		logger.Error("不能出一个空牌")
		return
	}

	if !room.IsActivePlayer(self) {
		logger.Error("不是活动玩家不能出牌")
		return
	}

	//如果有缓存的摸牌，那么将牌摸到手上
	room.DoMoByCache(false)

	//执行出
	if chuPai := self.controller.ChuPai(card); chuPai != nil {
		self.SwitchReadyDoAction(AChu)

		room.ResetAllAction(true)

		self.SendActionACK(AChu, chuPai, nil, ACSuccess)

		if room.state == RSBankerChuPaiStage && card.owner != nil && card.owner.IsBanker() {
			room.SwitchRoomState(RSBankerBaoPaiStage)
			room.CheckDoAction(nil, nil, nil, false)
		} else {
			room.ChangeActivePlayerToNext()
			room.CheckDoAction(chuPai, nil, []*MaJiangPlayer{self}, false)
		}

	} else {

		self.SendActionACK(AChu, card, nil, AOccursError)

		self.PlayerDoAction(ATuoGuan, nil)
	}
}

//过牌
func (self *MaJiangPlayer) DoGuo(card *MaJiangCard) {

	if !self.HaveWaitingDoAction() {
		logger.Error("等待执行的动作为空。所以不能过任何动作！")
		return
	}

	wa := self.watingAction[0]
	isGuo := self.mode == cmn.Manual
	if isGuo {
		//self.ResetDoAction()

		logger.Info("DoGuo：手动过的动作：", actionName[wa])
	} else {

		self.SwitchReadyDoAction(wa)

		logger.Info("DoGuo：自动过的动作：", actionName[wa])
	}

	switch wa {
	case AHu:
		fallthrough
	case AMingGang:
		fallthrough
	case ATieMingGang:
		fallthrough
	case APeng:
		fallthrough
	case ATiePeng:
		self.DoHuPengGangAfter(card, isGuo)
		logger.Info("DoGuo: 过胡，明杠和碰:", self.client.GetName())
	case ABao:
		self.DoBaoAfter(isGuo)
		logger.Info("DoGuo: 过报:", self.client.GetName())
	case AAnGang:
		if self.mode == cmn.Auto {
			if canAnGang, gangCards := self.controller.CheckAnGang(card); canAnGang && len(gangCards) > 0 {
				self.DoAnGangAfter(gangCards[0], isGuo)
			} else {
				logger.Error("在自动过牌时，没有检查到暗杠，这个动作暗杠的动作是怎么来的！！！")
			}
		} else {
			self.DoAnGangAfter(card, isGuo)
		}

	case ABuGang:
		if self.mode == cmn.Auto {
			if canBuGang, gangCards := self.controller.CheckBuGang(card); canBuGang && len(gangCards) > 0 {
				self.DoBuGangAfter(gangCards[0], isGuo)
			} else {
				logger.Error("在自动过牌时，没有检查到补杠，这个动作补杠的动作是怎么来的！！！")
			}
		} else {
			self.DoBuGangAfter(card, isGuo)
		}
	case AChu:
		autoChu := self.controller.GetChuPai()
		if autoChu == nil {
			logger.Error("DoGuo:玩家手里竟然没有牌了:", self.client.GetName())
			return
		}

		self.DoChuAfter(autoChu)
		logger.Info("DoGuo: 过出牌:", self.client.GetName(), ConvertToWord(autoChu))

	default:
		logger.Error("DoGuo:其他状态不能过", actionName[wa])
		debug.PrintStack()

	}
}

//获取要胡的牌
func (self *MaJiangPlayer) GetHuCards(isCheckQiHuKeAmount bool) []*MaJiangCard {

	controller := self.controller
	//检查又没有胡的模式组
	if len(controller.huController.patternGroups) <= 0 {
		return nil
	}

	//统计胡的牌
	result := []*MaJiangCard{}

	if !isCheckQiHuKeAmount {
		for _, patternGroup := range controller.huController.patternGroups {
			for j := 0; patternGroup.huCards != nil && j < len(patternGroup.huCards); j++ {

				huCard := patternGroup.huCards[j]
				if !IsExist(result, huCard) {
					result = append(result, huCard)
				}
			}
		}
	} else {
		for _, patternGroup := range controller.huController.patternGroups {
			for j := 0; patternGroup.huCards != nil && j < len(patternGroup.huCards); j++ {

				huCard := patternGroup.huCards[j]
				_, ke, _ := self.CalcMulitAndKeByPatternGroup(patternGroup, huCard)

				if self.room == nil || ke < self.room.QiHuKeAmount {
					continue
				}

				if !IsExist(result, huCard) {
					result = append(result, huCard)
				}
			}
		}

	}

	return result
}

//获取指定牌最大的胡的模式组
func (self *MaJiangPlayer) GetMaxHuOfPatternGroupByCard(card *MaJiangCard) (maxKe int32, result *MaJiangPatternGroup) {
	//检查参数的合法性
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	controller := self.controller
	//获取最大组模式
	if len(controller.huController.patternGroups) <= 0 {
		logger.Info("没有胡牌的模式组！", ConvertToWord(card))
		return
	}

	maxKe = 0
	for _, patternGroup := range controller.huController.patternGroups {

		if !patternGroup.CanHuSpecificCard(card) {
			continue
		}

		_, ke, _ := self.CalcMulitAndKeByPatternGroup(patternGroup, card)

		if self.room == nil || ke < self.room.QiHuKeAmount {
			logger.Info("胡这张牌%s 的起胡颗数不够！当前颗数：%d，起胡颗数：%d", ConvertToWord(card), ke, self.room.QiHuKeAmount)
			continue
		}

		if maxKe < ke {
			maxKe = ke
			result = patternGroup
		}
	}

	return
}

//获取最大胡的模式组
func (self *MaJiangPlayer) GetMaxHuOfPatternGroup() (result *MaJiangPatternGroup, huCard *MaJiangCard) {
	//获取最大组模式
	controller := self.controller
	if len(controller.huController.patternGroups) <= 0 {
		return
	}

	var maxKe int32 = 0
	for _, patternGroup := range controller.huController.patternGroups {
		for _, card := range patternGroup.huCards {
			_, ke, _ := self.CalcMulitAndKeByPatternGroup(patternGroup, card)

			if self.room == nil || ke < self.room.QiHuKeAmount {
				continue
			}

			if maxKe < ke {
				maxKe = ke
				result = patternGroup
				huCard = card
			}
		}
	}

	return
}

//计算模式组的翻数和颗数
func (self *MaJiangPlayer) CalcMulitAndKeByPatternGroup(patternGroup *MaJiangPatternGroup, huCard *MaJiangCard) (multi, ke int32, multipleResult map[int32]int32) {

	multipleResult = make(map[int32]int32, 0)
	//检查输入参数的合法性
	if patternGroup == nil {
		logger.Error("MaJinagPlayer.CalcMulitAndKeByPatternGroup: patternGroup is nil")
		return
	}

	if huCard == nil {
		logger.Error("MaJinagPlayer.CalcMulitAndKeByPatternGroup: huCard is nil.")
		return
	}

	if !patternGroup.CanHuSpecificCard(huCard) {
		PrintCardS("patternGroup can't hu ", huCard)
	}

	//计算临时倍数
	tempMultipleCount := make(map[int32]int32, 0)
	//归
	tempMultipleCount[MTGui] = self.GetGuiAmountByPatternGroup(patternGroup, huCard)
	//清一色
	if patternGroup.IsQingYiSe(self.showPatterns, huCard) {
		tempMultipleCount[MTQingYiSe] = MinTangFanShu[MTQingYiSe]
	}
	//无鬼(不用检查胡的牌)
	if patternGroup.IsNoneHongZhong(self.showPatterns) {
		//tempMultipleCount[MTNoneHongZhong] = MinTangFanShu[MTNoneHongZhong]
		tempMultipleCount[MTNoneHongZhong] = 3
		if self.room != nil && self.room.TotalHongZhongAmount == 8 {
			tempMultipleCount[MTNoneHongZhong] = 2
		}
	}
	//七对
	if patternGroup.IsQiDui(huCard) {
		tempMultipleCount[MTQiDui] = MinTangFanShu[MTQiDui]
	} else {
		//大对子
		if patternGroup.IsDaDuiZi(huCard) {
			tempMultipleCount[MTDaDuiZi] = MinTangFanShu[MTDaDuiZi]
		}
	}
	//顶报 (不是自己的牌且自己报牌，对方也报牌才能叫顶报)
	isSelfCard := huCard.owner != nil && huCard.owner.id == self.id
	oppositeHaveBao := huCard.owner != nil && huCard.owner.HaveBao()
	if !isSelfCard && self.HaveBao() && oppositeHaveBao {
		tempMultipleCount[MTDingBao] = MinTangFanShu[MTDingBao]
	}

	//计算胡牌时的翻数
	huMultipleCount := make(map[int32]int32, 0)
	if !self.isChaJiaoHu {
		//自摸
		if huCard.owner == nil || huCard.owner.id == self.id {
			huMultipleCount[MTZiMo] = MinTangFanShu[MTZiMo]
		}
		//杠上花
		if (huCard.owner == nil || huCard.owner.id == self.id) && self.aroundState.HaveGangShangHuaFlag() {
			huMultipleCount[MTGangShangHua] = MinTangFanShu[MTGangShangHua]
		}
		//杠上炮
		if huCard.owner != nil && huCard.owner.id != self.id && huCard.owner.aroundState.HaveGangShangPaoFlag() {
			huMultipleCount[MTGangShangPao] = MinTangFanShu[MTGangShangPao]
		}

		//抢杠
		if huCard.owner != nil && huCard.owner.id != self.id && huCard.owner.aroundState.HaveBuGang() {
			huMultipleCount[MTQiangGang] = MinTangFanShu[MTQiangGang]
		}
		//天胡
		if self.IsBanker() && self.room != nil &&
			self.room.lc.TotalCardAmount()-self.room.lc.RemainCardAmount() == 53 {
			huMultipleCount[MTTianHu] = MinTangFanShu[MTTianHu]
		}
	}

	//统计翻数
	multi = self.StatisticsMultipleCount(tempMultipleCount, huMultipleCount)

	if multi > int32(len(KeAmount)) {
		logger.Error("玩家（%s）超过了最大倍数(%d) 那么此时的名堂有哪些：", self.ID(), multi, self.multipleCount, tempMultipleCount, huMultipleCount)
		PrintCardsS("此时的手牌是", self.cards)

		multi = int32(len(KeAmount))
	}

	if multi > 0 {
		ke = KeAmount[multi-1]
	}

	//统计现在总的名堂信息
	for k, v := range tempMultipleCount {
		multipleResult[k] = v
	}
	for k, v := range huMultipleCount {
		multipleResult[k] = v
	}

	return

}

//统计当前的翻数
func (self *MaJiangPlayer) StatisticsMultipleCount(tempMultipleCount map[int32]int32, huMultipleCount map[int32]int32) int32 {

	//统计翻数
	//统计已经确定的翻数（报牌，如果是最后计算时，self.multipleCount将包含所有的翻）
	var fanCount int32 = 0
	for _, fan := range self.multipleCount {
		fanCount += fan
	}

	//统计临时的翻数（归，大对子，清一色，无鬼，七对,顶报）
	if tempMultipleCount != nil {
		for _, fan := range tempMultipleCount {
			fanCount += fan
		}
	}

	//统计胡牌时的翻数（自摸，杠上花，杠上炮，抢杠，天胡）
	if huMultipleCount != nil {
		for _, fan := range huMultipleCount {
			fanCount += fan
		}
	}

	return fanCount
}

//获取归的数量
func (self *MaJiangPlayer) GetGuiAmountByPatternGroup(patternGroup *MaJiangPatternGroup, huCard *MaJiangCard) (result int32) {

	if patternGroup == nil {
		logger.Error("MaJiangPlayer.GetGuiAmountByPatternGroup:patternGroup is nil.")
		return
	}

	if huCard == nil {
		logger.Error("MaJiangPlayer.GetGuiAmountByPatternGroup:huCard is nil.")
		return
	}

	cards := self.GetAllCardsByPatternGroup(patternGroup, huCard)
	if cards != nil {
		amountInfo := NewCardAmountStatisticsByCards(cards, true)
		result = amountInfo.GetAmountBySpecificAmount(4)
	} else {
		logger.Error("玩家没有牌")
	}

	return
}

//获取玩家的所有卡牌
func (self *MaJiangPlayer) GetAllCardsByPatternGroup(patternGroup *MaJiangPatternGroup, huCard *MaJiangCard) (result []*MaJiangCard) {
	if patternGroup == nil {
		logger.Error("MaJiangPlayer.GetAllCardsByPatternGroup: patternGroup is nil.")
		return
	}

	if huCard == nil {
		logger.Error("")
	}

	result = []*MaJiangCard{}

	for _, p := range self.showPatterns {
		result = append(result, p.cards...)
	}

	result = append(result, patternGroup.GetCards()...)

	result = append(result, huCard)

	return
}

//获取显示的牌的花色数量
func (self *MaJiangPlayer) GetCurMayOwnTypes() (result []int32) {
	result = []int32{Tiao, Tong, Wan}
	if self.showPatterns == nil || len(self.showPatterns) <= 0 {
		return
	}

	showTypes := GetTypeInfoByPatternList(self.showPatterns, nil)
	if showTypes != nil && len(showTypes) >= 2 {
		return showTypes
	}

	return
}

//获取已经碰了的牌
func (self *MaJiangPlayer) GetPengCardsForAlready() (result []*MaJiangCard) {
	result = make([]*MaJiangCard, 0)
	if self.showPatterns == nil || len(self.showPatterns) <= 0 {
		return
	}

	for _, p := range self.showPatterns {
		if p.ptype == PTKan && p.cards != nil && len(p.cards) > 0 {
			result = append(result, p.cards[0])
		}
	}

	return
}

//添加一张补杠牌
func (self *MaJiangPlayer) AddOneBuGangCard(card *MaJiangCard) *MaJiangPattern {
	//检查输入参数
	if card == nil {
		logger.Error("card is nil.")
		return nil
	}
	if self.showPatterns == nil || len(self.showPatterns) <= 0 {
		logger.Error("没有显示的牌，不能进行补杠")
		return nil
	}

	for _, p := range self.showPatterns {
		if p.ptype == PTKan && p.cards != nil && len(p.cards) > 0 {
			if p.cards[0].IsEqual(card) {
				p.ptype = PTGang
				p.Init(append(p.cards, card))
				return p
			}
		}
	}

	return nil
}

//获取手上的红中数量
func (self *MaJiangPlayer) GetHongZhongAmountInHand() (result int32) {

	for _, c := range self.cards {
		if c.IsHongZhong() {
			result++
		}
	}
	return
}

//切换手动或自动模式
func (self *MaJiangPlayer) SwitchControllMode(mode int) {
	self.mode = int32(mode)
	//debug.PrintStack()
	//logger.Error("设置自动模式：", mode)
}

//切换等待动作
func (self *MaJiangPlayer) SwitchWatingAction(watingAction []int32) {
	logger.Info("切换等待动作到：%s  玩家的位置：%d", CnvtActsToStr(watingAction), self.room.GetPlayerIndex(self))

	if self.HaveWaitingDoAction() {
		self.readyDoAction = ANone
	}
	self.watingAction = watingAction
}

//有等待执行的动作吗
func (self *MaJiangPlayer) HaveWaitingDoAction() bool {
	return self.watingAction != nil && len(self.watingAction) > 0 && !Exist(self.watingAction, ANone)
}

//切换准备执行的动作
func (self *MaJiangPlayer) SwitchReadyDoAction(readyDoAction int32) {
	//logger.Info("切换准备执行动作到：", actionName[readyDoAction])
	if readyDoAction != ANone {
		self.watingAction = []int32{}
	}
	self.readyDoAction = readyDoAction
}

//设置延迟执行的动作
func (self *MaJiangPlayer) SetDelayDoAction(action int32) {
	self.delayDoAction = action
}

//重置动作状态
func (self *MaJiangPlayer) ResetDoAction() {
	self.SwitchWatingAction([]int32{})
	self.SwitchReadyDoAction(ANone)
}

//是否已经胡牌了
func (self *MaJiangPlayer) IsHu() (ok bool) {
	return (self.cards == nil || len(self.cards) <= 0) &&
		(self.showPatterns != nil && len(self.showPatterns) > 0) && self.huCard != nil
}

//是否已经胡牌了
func (self *MaJiangPlayer) GetKeAmountOfHu(beiHuPlayer *MaJiangPlayer) (ke int32) {

	//检查参数是否合法
	room := self.room
	if room == nil {
		logger.Error("room is nil")
		return
	}

	//胡
	if self.IsHu() {
		multiAmount := self.StatisticsMultipleCount(nil, nil)

		if !self.HaveSpecificMingTang(MTDingBao) && beiHuPlayer != nil {
			isDingBao := self.HaveBao() && beiHuPlayer.HaveBao()
			if isDingBao {
				multiAmount += MinTangFanShu[MTDingBao]
			}
		}

		maxMultiAmount := int32(math.Min(float64(multiAmount), float64(len(KeAmount))))

		ke = KeAmount[maxMultiAmount-1]

		ke = int32(math.Min(float64(ke), float64(room.MaxMultiple)))
	}

	return
}

//获取ID
func (self *MaJiangPlayer) ID() string {
	return self.id
}

//获取基本
func (self *MaJiangPlayer) GetPlayerBasicInfo() *rpc.PlayerBaseInfo {
	return self.client
}

//设置基础信息
func (self *MaJiangPlayer) SetPlayerBasicInfo(client *rpc.PlayerBaseInfo) {
	self.client = client
}

//是否是准备
func (self *MaJiangPlayer) IsReady() bool {
	return self.isReady
}

//是否是战斗中的动作
func IsActionInFight(action int32) bool {
	return action == AGuo || action == AChu || action == AMo ||
		action == APeng || action == ATiePeng || action == AAnGang || action == AMingGang ||
		action == ATieMingGang || action == ABuGang || action == AHu || action == ABao
}

//检测一个动作时候需要等待,
func IsWaitingAction(actions []int32) bool {
	for _, action := range actions {
		if action == AChu || action == APeng || action == ATiePeng || action == AAnGang ||
			action == AMingGang || action == ATieMingGang || action == ABuGang || action == AHu || action == ABao {
			return true
		}
	}

	return false
}

//设置room
func (self *MaJiangPlayer) SetRoom(room *MaJiangRoom) {
	self.room = room
}

//有叫吗
func (self *MaJiangPlayer) HaveJiao() bool {
	huC := self.controller.huController
	return huC.patternGroups != nil && len(huC.patternGroups) > 0
}

//是否是庄家
func (self *MaJiangPlayer) IsBanker() bool {
	return self.ptype == cmn.PTBanker
}

//有报吗
func (self *MaJiangPlayer) HaveBao() bool {
	return self.HaveSpecificMingTang(MTBao)
}

//自摸吗？此函数自后在胡了牌的晚间才有效
func (self *MaJiangPlayer) IsZiMo() bool {
	return self.HaveSpecificMingTang(MTZiMo)
}

//是否具有自摸特性(需要家家给钱的)
func (self *MaJiangPlayer) HaveZiMoFeatureForHu() bool {
	return self.IsZiMo() || self.HaveSpecificMingTang(MTGangShangHua) || self.HaveSpecificMingTang(MTTianHu)
}

//有指定名堂吗
func (self *MaJiangPlayer) HaveSpecificMingTang(mingtang int32) bool {
	val, exist := self.multipleCount[mingtang]
	if exist && val > 0 {
		return true
	}

	return false
}

//获得牌
func (self *MaJiangPlayer) ObtainCard(card *MaJiangCard) {
	//检查输入参数的合法性
	if card == nil {
		logger.Error("庄家进的第一张是nil.")
		return
	}

	if self.controller == nil {
		logger.Error("self.controller is nil.")
		return
	}

	//在手牌里添加一张新拍
	card.owner = self
	self.cards = append(self.cards, card)

	//手牌变了后需要从新更新hu控制器
	//self.controller.huController.UpdateData(self.cards)

	//重新计算缓存的卡牌数量
	self.cardAmountInfo.CalcCardAmountByCards(self.cards, false)
}

//获取上家
func (self *MaJiangPlayer) GetShangJia() *MaJiangPlayer {
	room := self.room
	if room == nil {
		logger.Error("self.room is nil.")
		return nil
	}

	curPlayerIndex := room.GetPlayerIndex(self)
	if curPlayerIndex >= 0 {
		curPlayerIndex--
		shangJiaIndex := (curPlayerIndex + RoomMaxPlayerAmount) % RoomMaxPlayerAmount
		//logger.Info("self.GetShangJia Index:.", shangJiaIndex)
		return room.players[shangJiaIndex]
	}

	return nil
}

//获取下家
func (self *MaJiangPlayer) GetXiaJia() *MaJiangPlayer {
	room := self.room
	if room == nil {
		logger.Error("self.room is nil.")
		return nil
	}

	curPlayerIndex := room.GetPlayerIndex(self)
	if curPlayerIndex >= 0 {
		curPlayerIndex++
		xiaJiaIndex := curPlayerIndex % RoomMaxPlayerAmount
		logger.Info("self.GetXiaJia Index:.", xiaJiaIndex)
		return room.players[xiaJiaIndex]
	}

	return nil
}

//增加一张出牌
func (self *MaJiangPlayer) AddChuCard(card *MaJiangCard) {
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	//cloneCard := *card

	//cloneCard.flag = cmn.CLock | cmn.CPositive | cmn.CLockHongZhongValue

	card.flag = cmn.CLock | cmn.CPositive | cmn.CLockHongZhongValue

	self.chuCards = append(self.chuCards, card)
}

//移除手牌
func (self *MaJiangPlayer) RemoveHandCards(cType, value, wantRemovedAmount int32) (result []*MaJiangCard, outRemovedCards []*MaJiangCard) {

	//移除手上的牌
	removedCards := make([]*MaJiangCard, 0)
	removedHongZhongCards := make([]*MaJiangCard, 0)

	self.cards, removedCards = RemoveCardsByType(self.cards, cType, value, wantRemovedAmount)
	//检查本牌是否足够，不足够则用红中替代
	needRemovedHongZhongAmount := wantRemovedAmount - int32(len(removedCards))
	if needRemovedHongZhongAmount > 0 {
		self.cards, removedHongZhongCards = RemoveCardsByType(self.cards, HongZhong, 0, needRemovedHongZhongAmount)
	}

	//设置红中的替换值并锁定替换
	for _, hongZhongCard := range removedHongZhongCards {
		if hongZhongCard == nil {
			continue
		}

		hongZhongCard.SetHZReplaceValue(cType, value)
		hongZhongCard.flag = cmn.CLockHongZhongValue | cmn.CLock | cmn.CPositive
	}

	return self.cards, append(removedCards, removedHongZhongCards...)
}

//移除一张出牌
func (self *MaJiangPlayer) RemoveChuCard(card *MaJiangCard) {
	if card == nil {
		logger.Error("card is null.")
		return
	}

	if len(self.chuCards) <= 0 {
		return
	}

	lastCard := self.chuCards[len(self.chuCards)-1]
	if lastCard.IsFullEqual(card) {
		logger.Info("通知玩家(%s)出移一张牌：%s", self.client.GetName(), ConvertToWord(card))
		self.chuCards = self.chuCards[:len(self.chuCards)-1]
		self.SendRemoveCardNotifyACK(card, false)
	} else {
		//logger.Error("最后一张牌不是：", ConvertToWord(card))
	}

	// for i := len(self.chuCards) - 1; i >= 0; i-- {
	// 	c := self.chuCards[i]
	// 	if c != nil && c.IsFullEqual(card) {
	// 		self.chuCards = append(self.chuCards[:i], self.chuCards[i+1:]...)
	// 		self.SendRemoveCardNotifyACK(card)
	// 		return
	// 	}
	// }

}

//检查两个玩家列表是否相同
func IsSamePlayersList(p1 []*MaJiangPlayer, p2 []*MaJiangPlayer) bool {
	if p1 == nil && p2 == nil {
		return true
	}

	if p1 == nil && p2 != nil {
		return false
	}

	if p1 != nil && p2 == nil {
		return false
	}

	if len(p1) != len(p2) {
		return false
	}

	tempP2 := make([]*MaJiangPlayer, len(p2))
	copy(tempP2, p2)

	for _, p1p := range p1 {
		for i, p2p := range tempP2 {
			if p1p.id == p2p.id {
				tempP2 = append(tempP2[:i], tempP2[i+1:]...)
				break
			}
		}
	}

	return len(tempP2) <= 0
}

//是否有这个玩家，在这个列表里
func IsExistPlayer(players []*MaJiangPlayer, player *MaJiangPlayer) bool {
	if players == nil {
		return false
	}

	for _, p := range players {
		if p == nil {
			continue
		}

		if p.id == player.id {
			return true
		}
	}

	return false
}

//获取固定模式(碰和杠后显示在桌面上的牌)中牌的类型数量
func (self *MaJiangPlayer) GetTypeInfoInShowPattern() (result []int32) {

	return GetTypeInfoByPatternList(self.showPatterns, nil)
}

//获取红中可替换的类型（条，筒，万）
func (self *MaJiangPlayer) GetCanReplaceType() (result [][]int32) {
	result = make([][]int32, 0)

	fixedType := GetTypeInfoByPatternList(self.showPatterns, nil)

	handCardsType := GetTypeInfoByCardList(self.cards, fixedType)

	tempType := []int32{}
	tempType = append(tempType, fixedType...)
	tempType = append(tempType, handCardsType...)
	switch len(tempType) {
	case 1:
		result = append(result, tempType)

		switch tempType[0] {
		case Tiao:
			result = append(result, []int32{Tiao, Tong})
			result = append(result, []int32{Tiao, Wan})
		case Tong:
			result = append(result, []int32{Tong, Tiao})
			result = append(result, []int32{Tong, Wan})
		case Wan:
			result = append(result, []int32{Wan, Tiao})
			result = append(result, []int32{Wan, Tong})
		default:
			logger.Error("不能是其他类型，只能是条，筒，万")
		}
	case 2:
		result = append(result, tempType)
	}

	return

}

//
//网络消息相关函数
//

//发送战斗开始
func (self *MaJiangPlayer) SendGameStartACK(reEnter bool) {
	msg := &rpc.MJGameStartACK{}

	//设置战斗状态
	room := self.room
	if room == nil {
		return
	}

	//确定当前房间的状态
	if room.state == RSReady {
		msg.SetFightState(cmn.FSReady)
	} else if room.state == RSSettlement {
		msg.SetFightState(cmn.FSSettlement)
	} else {
		msg.SetFightState(cmn.FSFighting)
	}

	logger.Info("发送房间的状态：：：：：：：：：", msg.GetFightState())

	//组织FightselfInfo结构
	for _, p := range self.room.players {
		if p != nil {
			fgtPlayersInfo := fillFightPlayerInfoMsg(p, self.id == p.id)
			msg.FightPlayersInfo = append(msg.FightPlayersInfo, fgtPlayersInfo)
			logger.Info("玩家的战斗信息:Name:%s, Banker:%s", p.GetPlayerBasicInfo().GetName(), fgtPlayersInfo.GetBZhuang())
		}
	}

	//组织MJFightCurrentStateInfo结构
	msgc := &rpc.MJFightCurrentStateInfo{}

	//填充倒计时
	// for _, p := range room.players {
	// 	if p != nil && p.HaveWaitingDoAction() {
	// 		countDown := &rpc.MJCountDown{}
	// 		countDown.SetPlayerID(p.id)
	// 		countDown.SetCurrentCountDown(room.GetRemainTime())
	// 		msgc.CurrentCountDownInfo = append(msgc.CurrentCountDownInfo, countDown)
	// 	}
	// }

	//填充当前活动的玩家
	ap := room.GetActivePlayer()
	if ap != nil {
		msgc.SetActivePlayerID(ap.id)
		msgc.SetCurrentCountDown(room.GetRemainTime())
	} else {
		logger.Error("竟然没有活动玩家")
	}

	//填充上一个活动玩家
	if room.activeCard != nil && room.activeCard.owner != nil {
		msgc.SetLastActivePlayerID(room.activeCard.owner.id)
	}

	//填充当前座面上剩余的卡牌数
	msgc.SetCurrentDeskRemainCard(room.lc.RemainCardAmount())

	logger.Info("当前座面的牌数：", msgc.GetCurrentDeskRemainCard())

	msg.SetCurrentFightState(msgc)

	if err := conn.SendCommonNotify2S([]string{self.id}, msg, "MJGameStartACK"); err != nil {
		logger.Error("发送游戏开始出错：", err, msg)
	}

	//如果是重登并且当前玩家有等待执行的动作，需要把这个动作通知给客服端
	if reEnter {
		logger.Info("重登录时，检查是否有摸牌动作：", self.aroundState.moCard != nil)
		if !self.IsHu() && self.aroundState.moCard != nil {
			self.SendActionACK(AMo, self.aroundState.moCard, nil, ACSuccess)
		}

		logger.Info("重登录时，玩家等待的动作：", CnvtActsToStr(self.watingAction), self.sendedActionMsg)
		if self.HaveWaitingDoAction() && self.sendedActionMsg != nil {
			if err := conn.SendCommonNotify2S([]string{self.id}, self.sendedActionMsg, "MJActionNotifyACK"); err != nil {
				logger.Error("发送恢复动作出错：", err, msg)
			}
		}
	}
}

//填充战斗开始信息
func fillFightPlayerInfoMsg(p *MaJiangPlayer, isSelf bool) *rpc.MJFightPlayerInfo {
	//组织MJFightPlayerInfo结构
	msgc := &rpc.MJFightPlayerInfo{}
	msgc.SetPlayerID(p.id)
	msgc.SetBZhuang(p.IsBanker())
	msgc.SetBBao(p.HaveBao())
	msgc.SetBTuoGuan(p.mode == cmn.Auto)

	msgc.ChuCards = convertCards(p.chuCards)
	msgc.ShowPatterns = convertPatterns(p.showPatterns)

	//如果已经胡牌的玩家手牌放在showPatterns中的，查看controller.Hu函数
	// handCards := make([]*MaJiangCard, 0)
	// if p.IsHu() {
	// 	for _, pattern := range p.showPatterns {
	// 		isInHandPattern := !pattern.isShowPattern
	// 		//logger.Error("是不是已显示模式:", pattern.isShowPattern)
	// 		if !isInHandPattern {
	// 			msgc.ShowPatterns = append(msgc.ShowPatterns, convertPattern(pattern))
	// 		} else {
	// 			handCards = append(handCards, pattern.cards...)
	// 		}
	// 	}

	// 	msgc.SetAlreadyCardArg(convertCard(p.huCard))
	// } else {

	// 	handCards = append(handCards, p.cards...)
	// }

	if p.huCard != nil {
		msgc.SetAlreadyCardArg(convertCard(p.huCard))
	}

	if isSelf {
		msgc.HandCards = convertCards(p.cards)
	} else {
		msgc.SetHandCardCount(int32(len(p.cards)))
	}

	return msgc
}

//发送可以执行动作通知到客服端
func (self *MaJiangPlayer) SendActionNotifyACK(curCard *MaJiangCard, actions []int32, cards map[int32][]*MaJiangCard) {
	room := self.room
	if room == nil {
		logger.Error("room is nil.")
		return
	}

	if len(actions) <= 0 {
		logger.Error("actions is empty.")
		return
	}

	//需要等待的动作
	self.SwitchWatingAction(actions)

	logger.Info("向（%s）发送准备执行的动作为：%s 是不是等待动作:%s", self.ID(), CnvtActsToStr(actions), IsWaitingAction(actions))
	if IsWaitingAction(actions) {
		if self.mode == cmn.Manual {
			self.room.StartTimer(room.TimerInterval)
		} else {

			//延迟执行这个动作
			delayCallId := self.id + CnvtActsToStr(actions)
			room.StartDelayCallback(delayCallId, room.DoActionDelay, func(data interface{}) {
				self.PlayerDoAction(AGuo, curCard)
			}, nil)
		}

	} else {
		logger.Info("自动执行的动作：%s,", CnvtActsToStr(actions))
		delayCallId := self.id + CnvtActsToStr(actions)
		room.StartDelayCallback(delayCallId, room.DoActionDelay, func(data interface{}) {
			logger.Info("延迟后自动执行的动作：%s, card:%s", CnvtActsToStr(actions), ConvertToWord(curCard))
			self.PlayerDoAction(actions[0], curCard)
		}, nil)
	}

	//向客户端发送消息
	//fmt.Println("向客户端发送触发动作：", actionName[action])
	msg := &rpc.MJActionNotifyACK{}
	for _, a := range actions {
		action := &rpc.MJActionArgs{}
		action.SetAction(a)
		if cards != nil {
			action.CardsArgs = convertCards(cards[a])
		} else {
			action.CardsArgs = make([]*rpc.MJCard, 0)
		}

		msg.Actions = append(msg.Actions, action)
	}
	if len(actions) <= 0 {
		logger.Error("到这里actions怎么可能是空呢！，进入函数就挡回去了")
	}

	if err := conn.SendCommonNotify2S([]string{self.id}, msg, "MJActionNotifyACK"); err != nil {
		logger.Error("发送触发动作通知出错：", err, *msg)
	}
	//缓存发送的消息一遍，重进入后恢复
	self.sendedActionMsg = msg

	//如果是通知出牌动作，那么发送倒计时的通知消息,因为出牌是单独的没有走CheckCanDoAction
	if actions[0] == AChu {
		logger.Info("==______给玩家发送出。。并通知玩家倒计时。。。。", CnvtActsToStr(self.watingAction))
		room.sendCountdownNotifyACK()
	}
}

//发送倒计时通知信息
func (self *MaJiangPlayer) sendCountdownNotifyACK(cp *MaJiangPlayer) {
	if cp != nil {
		timerInfo := &rpc.MJCountDown{}
		timerInfo.SetPlayerID(cp.id)
		timerInfo.SetCurrentCountDown(int32(cp.room.TimerInterval))

		msg := &rpc.MJCountdownNotifyACK{}
		msg.SetCountDown(timerInfo)
		if err := conn.SendCommonNotify2S([]string{self.id}, msg, "MJCountdownNotifyACK"); err != nil {
			logger.Error("发送倒计时出错：", err, msg)
		}
	}
}

//发送移除牌的通知
func (self MaJiangPlayer) SendRemoveCardNotifyACK(card *MaJiangCard, isRemoveHandCard bool) {
	if self.room == nil {
		logger.Error("player.room is nil.")
		return
	}

	msg := &rpc.MJRemoveCardNotifyACK{}
	msg.SetPlayerID(self.id)
	msg.SetIsRemoveHandCard(isRemoveHandCard)
	msg.SetCard(convertCard(card))

	if err := conn.SendCommonNotify2S(self.room.GetAllPlayerIDs(), msg, "MJRemoveCardNotifyACK"); err != nil {
		logger.Error("发送移除玩家出牌的牌是出错：", err, msg)
	}

}

//发送动作执行回复ACK
func (self *MaJiangPlayer) SendActionACK(action int32, card *MaJiangCard, pattern *MaJiangPattern, code int32) {
	//向客户端发送消息
	if card != nil {
		logger.Info("向客户端(%s)发送动作执行结果：%s,   Card:%s, Code:%s", self.ID(), actionName[action], ConvertToWord(card), code)
	} else {
		logger.Info("向客户端(%s)发送动作执行结果：%s Code:%s", self.ID(), actionName[action], code)
	}

	room := self.room
	if room == nil {
		logger.Error("room is nil.")
		return
	}

	for _, p := range room.players {
		if p != nil {
			msg := &rpc.MJActionACK{}
			msg.SetAction(int32(action))
			msg.SetPlayerID(self.id)

			if card != nil {
				msg.SetCardArgs(convertCard(card))
			}

			if pattern != nil {
				msg.PengGangPattern = append(msg.PengGangPattern, convertPattern(pattern))
			}

			msg.SetCurrenDeskRemianCard(room.lc.RemainCardAmount())
			msg.SetResult(int32(code))

			if err := conn.SendCommonNotify2S([]string{p.id}, msg, "MJActionACK"); err != nil {
				logger.Error("发送动作执行结果出错：", err, msg)
			}

			logger.Info("动作执行后 ID:%s, DeskRemianCard:", msg.GetPlayerID(), msg.GetCurrenDeskRemianCard())
		}
	}
}

//发送扣取金币
func (self *MaJiangPlayer) SendJieSuanCoinNotify(coin int32) {

	//通知gameserver扣钱
	if err := conn.SendCostResourceMsg(self.id, connector.RES_COIN, "majiang", coin); err != nil {
		logger.Error("发送扣取金币出错：", err, self.id, coin)
		return
	}

	self.client.SetCoin(self.client.GetCoin() + coin)
}

//发送结算
func (self *MaJiangPlayer) SendJieSuanACK(jieSuanCoin map[string]int32, addiData *rpc.JieSuanAdditionData) {

	//检查输入擦书
	if jieSuanCoin == nil {
		logger.Error("jieSuanCoin is nil.")
		return
	}

	room := self.room
	if room == nil {
		logger.Error("room is nil.")
		return
	}

	//发送结算信息
	msg := &rpc.MJJieSuanNotifyACK{}

	// addiData := &rpc.JieSuanAdditionData{}
	// addiData.SetSysType(cmn.PiPeiFang)
	msg.SetAddi(addiData)

	huangZhuang := room.IsHuangZhuang()
	msg.SetHuangZhuang(huangZhuang)

	//填充每个玩家的结算信息
	for _, p := range room.players {
		if p == nil {
			continue
		}

		pJieSuan := &rpc.MJPlayerJieSuanInfo{}
		pJieSuan.SetPlayerID(p.id)
		pJieSuan.Patterns = append(pJieSuan.Patterns, fillPatterns(p)...)
		pJieSuan.MingTang = append(pJieSuan.MingTang, fillMingTang(p)...)
		if val, exist := jieSuanCoin[p.id]; exist {
			pJieSuan.SetCoin(val)
		} else {
			logger.Error("在结算的金币信息中，竟然没有玩家：(%s),  金币信息:", p.id, jieSuanCoin)
		}

		msg.PlayerJieSuanInfo = append(msg.PlayerJieSuanInfo, pJieSuan)

	}

	//通知客服端结算
	if err := conn.SendCommonNotify2S([]string{self.id}, msg, "MJJieSuanNotifyACK"); err != nil {
		logger.Error("发送结算信息错误", err)
	}

	logger.Info("黄庄：%s   结算消息：", huangZhuang, msg)

	return

}

//填充最终的Patterns
func fillPatterns(player *MaJiangPlayer) (patterns []*rpc.MJPattern) {
	patterns = make([]*rpc.MJPattern, 0)

	if player == nil {
		logger.Error("player is nil.")
		return
	}

	if player.controller == nil {
		logger.Error("player.controller is nil")
		return
	}

	patternGroup := player.controller.GenerateFinalPatternGroup()
	if patternGroup == nil {
		logger.Error("patternGroup is nil.")
		return
	}

	patterns = convertPatterns(patternGroup.patterns)
	return
}

//填充名堂
func fillMingTang(player *MaJiangPlayer) (mingtang []*rpc.MJMingTang) {

	mingtang = make([]*rpc.MJMingTang, 0)

	if player == nil {
		logger.Error("player is nil.")
		return
	}

	for mt, mtVal := range player.multipleCount {
		if mtVal > 0 && mt != MTDingBao {
			rpcMt := &rpc.MJMingTang{}
			rpcMt.SetMingTang(int32(mt))
			rpcMt.SetValue(int32(mtVal))
			mingtang = append(mingtang, rpcMt)
		}
	}

	return mingtang
}

//转换daer.Card 到protobuff的Card
func convertCard(src *MaJiangCard) *rpc.MJCard {
	if src == nil {
		logger.Error("src is nil.")
		return nil
	}
	rpcCard := &rpc.MJCard{}
	rpcCard.SetValue(int32(src.value))
	rpcCard.SetCType(src.cType)
	rpcCard.SetRcType(src.rcType)
	rpcCard.SetFlag(src.flag)

	return rpcCard
}

//转换daer.Card 到protobuff的Card
func convertCards(src []*MaJiangCard) (dest []*rpc.MJCard) {
	if src == nil {
		return make([]*rpc.MJCard, 0)
	}
	dest = make([]*rpc.MJCard, len(src))
	for i, card := range src {
		if card == nil {
			logger.Error("转换的牌竟然是个nil.")
			continue
		}

		dest[i] = convertCard(card)
	}
	return
}

//转换protobuff的card到daer.Card
func convertCardToMaJiangCard(src *rpc.MJCard) *MaJiangCard {
	if src == nil {
		logger.Error("src is nil.")
		return nil
	}

	card := NewCard(0, src.GetCType(), src.GetValue())
	card.rcType = src.GetRcType()
	card.flag = src.GetFlag()

	return card
}

func convertCardsToMaJiangCards(src []*rpc.MJCard) (dest []*MaJiangCard) {
	if src == nil {
		return make([]*MaJiangCard, 0)
	}
	dest = make([]*MaJiangCard, len(src))
	for i, card := range src {
		dest[i] = convertCardToMaJiangCard(card)
	}
	return
}

//转换daer.Pattern 到protobuff的Pattern
func convertPattern(src *MaJiangPattern) *rpc.MJPattern {

	if src == nil {
		logger.Error("src is nil.")
		return nil
	}
	rpcPattern := &rpc.MJPattern{}
	rpcPattern.SetPtype(src.ptype)
	rpcPattern.SetCType(src.cType)
	rpcPattern.SetIsShow(src.isShowPattern)
	rpcPattern.Cards = convertCards(src.cards)

	return rpcPattern
}

//转换daer.Pattern 到protobuff的Pattern
func convertPatterns(src []*MaJiangPattern) (dest []*rpc.MJPattern) {
	if src == nil {
		return make([]*rpc.MJPattern, 0)
	}
	dest = make([]*rpc.MJPattern, len(src))
	for i, pattern := range src {
		dest[i] = convertPattern(pattern)
	}

	return
}
