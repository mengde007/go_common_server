package daerserver

import (
	conn "centerclient"
	cmn "common"
	"connector"
	//"fmt"
	"logger"
	"math"
	"rpc"
	"runtime/debug"
	"strconv"
)

type DaerPlayer struct {
	id            string
	ptype         int32 //庄家，闲家
	cards         []*DaerCard
	fixedpatterns []*DaerPattern
	showCards     []*DaerCard
	guoCards      []*DaerCard
	showPatterns  []*DaerPattern
	controller    *DaerController
	//autoController *AutoController
	isReady       bool
	room          *DaerRoom
	multipleCount map[int32]int32
	watingAction  int32
	readyDoAction int32 //准备执行的动作，玩家点击了吃，但是他的上家还在准备吃，当上家放弃吃的时候，用于表示自己可以吃
	delayDoAction int32 //需要延迟到一下个阶段才生效的动作（现在只有非庄家的报或摆）
	mode          int32 //自动/手动模式
	//erLongTouyiThreshold int   //二拢偷一的阀值
	erLongTouYi []*DaerCard //二拢偷一偷的牌
	client      *rpc.PlayerBaseInfo

	//记录玩家当前提交的吃和比的模式
	curChiCard  *DaerCard
	curKaoCards []*DaerCard
	curBiCards  []*DaerCard

	//记录下发给玩家的吃比的模式
	sendedChiBiMsg *rpc.ActionNotifyACK
}

func NewDaerPlayer(id string, playerInfo *rpc.PlayerBaseInfo) *DaerPlayer {
	p := new(DaerPlayer)
	p.id = id
	p.ptype = cmn.PTNormal
	p.watingAction = ANone
	p.readyDoAction = ANone
	p.delayDoAction = ANone
	p.client = playerInfo
	//p.erLongTouyiThreshold = 2
	p.erLongTouYi = make([]*DaerCard, 0)
	p.controller = NewController(p)
	//p.autoController = NewAutoController(p)

	if p.client == nil {
		logger.Error("Player:NewPlayer, playerInfo is nil.")
	}

	return p
}

//重置Player
func (player *DaerPlayer) Reset() {
	player.ptype = cmn.PTNormal
	player.cards = make([]*DaerCard, 0)
	player.fixedpatterns = make([]*DaerPattern, 0)
	player.showCards = make([]*DaerCard, 0)
	player.guoCards = make([]*DaerCard, 0)
	player.showPatterns = make([]*DaerPattern, 0)
	player.isReady = false
	player.multipleCount = make(map[int32]int32, 0)
	player.watingAction = ANone
	player.readyDoAction = ANone
	player.delayDoAction = ANone
	player.mode = cmn.Manual
	//player.erLongTouyiThreshold = 2
	player.erLongTouYi = make([]*DaerCard, 0)
	player.curChiCard = nil
	player.curKaoCards = nil
	player.curBiCards = nil
	player.sendedChiBiMsg = nil
}

//组牌
func (player *DaerPlayer) Compose(cards []*DaerCard) {
	//检查参数的合法性
	if cards == nil || len(cards) <= 0 {
		logger.Error("Compose:cards is nil or empty")
		return
	}

	if player.controller == nil {
		logger.Error("player.controller is nil.")
		return
	}

	//保存发的牌
	player.cards = cards

	//先把牌拢了
	player.controller.Long()

	//剔除坎牌
	player.controller.StripKan()

	//更新剩下的手牌状态
	player.controller.ModifyOtherCardStatusInHand()

	//更新手牌模式
	//player.controller.huController.UpdateData(player.cards)

}

//获取ID
func (player *DaerPlayer) ID() string {
	return player.id
}

//获取基本
func (player *DaerPlayer) GetPlayerBasicInfo() *rpc.PlayerBaseInfo {
	return player.client
}

//设置基础信息
func (player *DaerPlayer) SetPlayerBasicInfo(client *rpc.PlayerBaseInfo) {
	player.client = client
}

//是否是准备
func (player *DaerPlayer) IsReady() bool {
	return player.isReady
}

//设置room
func (player *DaerPlayer) SetRoom(room *DaerRoom) {
	player.room = room
}

//指定玩家执行动作
func (player *DaerPlayer) PlayerDoAction(action int32, card *DaerCard, kaoCards []*DaerCard, biCards []*DaerCard) {

	//检查能否执行这个动作
	if !player.CanDoAction(action) {
		logger.Info("等待的动作%s 和 执行的动作%s  不相同", player.watingAction, action)
		return
	}

	room := player.room

	switch action {
	case AReady: //准备
		if room.state != RSReady || player.isReady {
			player.SendActionACK(action, nil, nil, AOccursError)
			return
		}

		//准备的时候检查一下金币，结算后金币会变化
		if daerRoomMgr != nil {
			if ok, _ := cmn.CheckCoin(room.rtype, player.client); !ok {
				daerRoomMgr.LeaveGame(player.id, false)
				return
			}
		}

		player.isReady = true
		player.SendActionACK(action, nil, nil, ACSuccess)
		if room.CanStartGame() {
			room.StartGame()
		}
		logger.Info("PlayerDoAction: 准备:", player.client.GetName())

	case ACancelReady:
		if room.state != RSReady || !player.isReady {
			player.SendActionACK(action, nil, nil, AOccursError)
			return
		}

		player.isReady = false
		player.SendActionACK(action, nil, nil, ACSuccess)
		logger.Info("PlayerDoAction: 取消准备:", player.client.GetName())

	case ATuoGuan: //托管
		if !room.IsGaming() || player.mode == cmn.Auto {
			player.SendActionACK(action, nil, nil, AOccursError)
			return
		}
		player.SwitchControllMode(cmn.Auto)
		player.SendActionACK(action, nil, nil, ACSuccess)

		//有等待的动作就执行了
		if room.IsGaming() && player.HaveWaitingDoAction() && IsWaitingAction(player.watingAction) {
			player.PlayerDoAction(AGuo, room.activeCard, nil, nil)
		}

		logger.Info("PlayerDoAction: 托管:", player.client.GetName())

	case ACancelTuoGuan: //取消托管
		if !room.IsGaming() || player.mode == cmn.Manual {
			player.SendActionACK(action, nil, nil, AOccursError)
			return
		}

		player.SwitchControllMode(cmn.Manual)
		player.SendActionACK(action, nil, nil, ACSuccess)

		logger.Info("取消托管时的准备执行的动作：", player.watingAction)
		if player.HaveWaitingDoAction() && IsWaitingAction(player.watingAction) {
			delayCallId := player.id + strconv.Itoa(int(player.watingAction))
			room.StopDelayCallback(delayCallId)
			logger.Info("取消托管时停止自动执行：", delayCallId)
			room.StartTimer(room.TimerInterval)

			room.sendCountdownNotifyACK()
		}

		logger.Info("PlayerDoAction: 取消托管:", player.client.GetName())

	case AHu:
		player.SwitchReadyDoAction(AHu)
		player.DoHuAfter(card, false)
		logger.Info("PlayerDoAction: 胡牌:", player.client.GetName(), card.value)

	case AHeiBai:
		fallthrough
	case ASanLongBai:
		fallthrough
	case ASiKanBai:
		player.SwitchReadyDoAction(action)
		player.DoBaiAfter(action, false)
		logger.Info("PlayerDoAction: 摆牌:", player.client.GetName())

	case ABao:
		player.SwitchReadyDoAction(ABao)
		player.DoBaoAfter()

		logger.Info("PlayerDoAction: 报牌:", player.client.GetName())

	case AZhao:
		fallthrough
	case AZhongZhao:
		player.DoZhaoAfter(card)
		logger.Info("PlayerDoAction: 招牌:", player.client.GetName(), card.value)

	case APeng:
		player.DoPengAfter(card)
		logger.Info("PlayerDoAction: 碰牌:", player.client.GetName(), card.value)

	case AChi:
		player.SwitchReadyDoAction(AChi)
		player.curChiCard = card
		player.curKaoCards = kaoCards
		player.curBiCards = biCards
		player.sendedChiBiMsg = nil
		player.DoChiAfter(false)
		logger.Info("PlayerDoAction: 吃牌:", player.client.GetName(), card.value)

	case AChu:
		player.SwitchReadyDoAction(AChu)
		player.DoChuAfter(card)
		logger.Info("PlayerDoAction: 出牌:", player.client.GetName(), card.value)

	case AGuo:
		player.DoGuo(card)
		logger.Info("PlayerDoAction: 过:", player.client.GetName())
	default:
		logger.Error("执行的动作是否有问题：", action)
	}

	if action != AReady && action != ACancelReady && action != AReady && action != ACancelReady {
		logger.Info("===================执行动作后的情况===========================")
		PrintRoom(room)
	}
}

func (player *DaerPlayer) CanDoAction(action int32) bool {

	//检查参数的正确性
	room := player.room
	if room == nil {
		logger.Error("玩家没有所属的房间。")
		return false
	}

	if IsWaitingAction(action) && player.watingAction != action {
		logger.Error("等待的动作和执行动作不相同！W:%s   E:%s", player.watingAction, action)
		return false
	}

	//玩家报了牌只能胡
	if player.HaveBao() {
		if action == APeng || action == AChi {
			return false
		}

		if action == AGuo {
			if !IsHuAction(player.watingAction) && player.watingAction != AChu {
				return false
			}

			if player.watingAction == AChu && !player.IsBanker() ||
				player.watingAction == AChu && player.IsBanker() && !(room.state == RSBankerBaoStage || room.state == RSBankerJinPaiStage) {
				return false
			}
		}

		if action == AChu && !player.IsBanker() ||
			action == AChu && player.IsBanker() && !(room.state == RSBankerBaoStage || room.state == RSBankerJinPaiStage) {
			return false
		}
	}

	return true
}

//执胡牌
func (player *DaerPlayer) DoHuAfter(card *DaerCard, isGuo bool) bool {

	//检测是不是在一个房间里
	room := player.room
	if room == nil {
		return false
	}

	if card == nil {
		logger.Error("不能胡一个空牌")
		return false
	}

	//检测并执行胡
	if success, py := room.DoReadyActionByOrder(true); success {

		if room.state == RSBankerJinPaiStage && py.IsBanker() {
			room.ResetAllDelayAction()
		}

		py.controller.Hu(card)

		room.ChangeActivePlayerTo(py)

		room.ResetAllAction()

		py.SendActionACK(AHu, card, nil, ACSuccess)

		room.SwitchRoomState(RSSettlement)

		room.CheckDoAction(nil)

		return true
	}

	//没有人执行胡（都放弃了胡）
	if _, have := room.GetWatingActionPlayer([]int32{AHu}); !have {

		room.ResetAllAction()

		if room.state == RSBankerJinPaiStage {
			room.BankerDoMo(room.activeCard)
		} else if room.state == RSBankerChuPaiAfterStage {
			room.SwitchRoomState(RSLoopWorkStage)
			if room.CheckCanDoZhaoAndPengChi(card) == ANone {
				room.CheckDoAction(nil)
			}
		} else if room.state == RSLoopWorkStage {
			if room.CheckCanDoZhaoAndPengChi(card) == ANone {
				room.CheckDoAction(nil)
			}
		} else {
			logger.Error("胡牌时不应该处理其他状态的情况")
		}

	} else {
		if !isGuo {
			player.SendActionACK(AHu, nil, nil, ACWaitingOtherPlayer)
		}
	}

	return false
}

//摆阶段完成后要执行的操作
func (player *DaerPlayer) DoBaiAfter(action int32, isGuo bool) bool {
	//检测是不是在一个房间里
	room := player.room
	if room == nil {
		return false
	}

	//检测摆，并执行摆
	//if success, py := room.DoReadyActionByOrder(true); success {

	if !isGuo || player.mode == cmn.Auto {
		if player.IsBanker() {
			room.ChangeActivePlayerTo(player)

			switch action {
			case ASanLongBai:
				player.controller.SanLongBai()
			case ASiKanBai:
				player.controller.SiKanBai()
			case AHeiBai:
				player.controller.HeiBai()
			default:
				logger.Error("只有两种摆（三拢或四坎， 黑摆）")
			}

			room.ResetAllAction()

			room.ResetAllDelayAction()

			player.SendActionACK(action, nil, nil, ACSuccess)

			room.SwitchRoomState(RSSettlement)

			room.CheckDoAction(nil)
			return true
		} else {
			player.ResetDoAction()
			player.SetDelayDoAction(action)
			player.SendActionACK(action, nil, nil, ACWaitingOtherPlayer)
		}
	} else {
		player.ResetDoAction()
	}
	//}

	//没有人摆（都放弃了摆）
	if _, have := room.GetWatingActionPlayer([]int32{ASanLongBai, ASiKanBai, AHeiBai}); !have {

		room.ResetAllAction()

		if room.state == RSNotBankerBaiStage {

			room.SwitchRoomState(RSBankerJinPaiStage)

			openCard := room.OpenOneCard()
			if openCard == nil {
				logger.Error("在报牌阶段桌面上竟然会没有牌了。")
				return false
			}

			banker := room.GetBanker()
			if banker == nil {
				logger.Error("竟然没有庄家！太不可思议了")
				return false
			}

			//banker.SendActionACK(AMo, openCard, nil, ACSuccess)

			room.CheckDoAction(openCard)
		} else if room.state == RSBankerBaiStage {

			room.SwitchRoomState(RSBankerBaoStage)

			banker := room.GetBanker()
			if banker == nil {
				logger.Error("竟然没有庄家！太不可思议了")
				return false
			}

			banker.SendActionNotifyACK(AChu, nil, nil, nil)
			//room.CheckDoAction(room.activeCard)
		} else {
			room.CheckDoAction(nil)
		}
	}

	return false
}

//报阶段完成后要执行的操作
func (player *DaerPlayer) DoBaoAfter() {
	//检测是不是在一个房间里
	room := player.room
	if room == nil {
		logger.Error("room is nil")
		return
	}

	//执行报
	if player.readyDoAction == ABao {
		if player.IsBanker() {
			player.controller.Bao()
			player.ResetDoAction()
			player.SendActionACK(ABao, nil, nil, ACSuccess)
		} else {
			player.ResetDoAction()
			player.SetDelayDoAction(ABao)
			player.SendActionACK(ABao, nil, nil, ACWaitingOtherPlayer)
		}
	}

	//没有等待报牌的玩家吗
	if _, have := room.GetWatingActionPlayer([]int32{ABao}); !have {

		room.ResetAllAction()

		if room.state == RSBaoStage {
			room.SwitchRoomState(RSNotBankerBaiStage)
			room.CheckDoAction(room.activeCard)

		} else if room.state == RSBankerBaoStage {
			room.SwitchRoomState(RSBankerChuPaiAfterStage)
			doAction := room.DoDelayAction(true)
			if !IsBaiAction(doAction) {
				room.CheckDoAction(room.activeCard)
			}
		}
	}
}

//使延迟动作生效
func (player *DaerPlayer) DoDelayAction(excute bool) {
	delayAction := player.delayDoAction
	if delayAction == ABao {
		if excute {
			player.controller.Bao()
			player.SendActionACK(ABao, nil, nil, ACSuccess)
		} else {
			player.SendActionACK(ABao, nil, nil, ACAbandon)
		}
	} else if IsBaiAction(delayAction) {
		if excute {
			room := player.room
			if room == nil {
				logger.Error("room is nil.")
				return
			}

			if success, py := room.DoDelayActionByOrder(true); success {

				room.ChangeActivePlayerTo(py)

				switch delayAction {
				case ASanLongBai:
					py.controller.SanLongBai()
				case ASiKanBai:
					py.controller.SiKanBai()
				case AHeiBai:
					py.controller.HeiBai()
				default:
					logger.Error("只有两种摆（三拢或四坎， 黑摆）")
				}

				py.SetDelayDoAction(ANone)

				py.SendActionACK(delayAction, nil, nil, ACSuccess)

				room.ResetAllDelayAction()

				room.SwitchRoomState(RSSettlement)

				room.CheckDoAction(nil)
			}
		} else {
			player.SendActionACK(delayAction, nil, nil, ACAbandon)
		}
	} else {
		logger.Error("其他动作不能作为延迟执行的动作：", actionName[delayAction])
	}
}

//执行招牌
func (player *DaerPlayer) DoZhaoAfter(card *DaerCard) {
	//检测是不是在一个房间里
	room := player.room
	if room == nil {
		return
	}

	if card == nil {
		logger.Error("不能招一个空牌")
		return
	}

	//执行招
	if room.state == RSBankerChuPaiAfterStage {
		room.SwitchRoomState(RSLoopWorkStage)
	}

	room.ChangeActivePlayerTo(player)

	room.ResetAllAction()

	if pattern, isBaKuai := player.controller.Zhao(card); pattern != nil && isBaKuai {
		player.SendActionACK(AZhongZhao, card, []*DaerPattern{pattern}, ACSuccess)
		room.CheckDoAction(nil)
	} else if pattern != nil {
		player.SendActionACK(AZhao, card, []*DaerPattern{pattern}, ACSuccess)
		if player.HaveBao() {
			room.CheckDoAction(nil)
		} else {
			player.SendActionNotifyACK(AChu, nil, nil, nil)
		}
	}
}

//执行碰
func (player *DaerPlayer) DoPengAfter(card *DaerCard) {
	//检测是不是在一个房间里
	room := player.room
	if room == nil {
		return
	}

	if card == nil {
		logger.Error("不能碰一个空牌")
		return
	}

	//执行碰
	if room.state == RSBankerChuPaiAfterStage {
		room.SwitchRoomState(RSLoopWorkStage)
	}

	room.ChangeActivePlayerTo(player)

	room.ResetAllAction()

	pattern := player.controller.Peng(card)

	if pattern == nil {
		logger.Error("碰牌失败")
		return
	}
	player.SendActionACK(APeng, card, []*DaerPattern{pattern}, ACSuccess)

	player.SendActionNotifyACK(AChu, nil, nil, nil)
}

//执行吃牌
func (player *DaerPlayer) DoChiAfter(isGuo bool) bool {
	//检测是不是在一个房间里
	room := player.room
	if room == nil {
		logger.Error("room is nil")
		return false
	}

	//检测并执行吃
	if success, py := room.DoReadyActionByOrder(false); success {
		//切换到下一阶段
		if room.state == RSBankerChuPaiAfterStage {
			room.SwitchRoomState(RSLoopWorkStage)
		}

		//改变吃的人为活动玩家
		room.ChangeActivePlayerTo(py)

		//执行吃
		card := py.curChiCard
		kaoCards := py.curKaoCards
		biCards := py.curBiCards

		py.controller.Chi(kaoCards, card, biCards)

		//重置玩家的动作
		room.ResetAllAction()

		//统计吃的模式，并下发给客户端
		chiBiResult := make([]*DaerPattern, 0)
		if kaoCards != nil && len(kaoCards) > 0 && card != nil {
			chiPType := py.controller.CalcPatternType(kaoCards, card, nil)
			if chiPType == PTUknown {
				logger.Error("DoChiAfter:吃的牌不能构成一个模式：靠牌和吃牌如下：")
				PrintCards(kaoCards)
				PrintCard(card)
				return false
			}
			chiBiResult = append(chiBiResult, NewPattern(chiPType, append(kaoCards, card)))
		} else {
			logger.Error("DoChiAfter:kaoCards 不应该是空, 玩家模式：(%s)， ", player.mode)
			return false
		}

		if biCards != nil && len(biCards) > 0 {
			biPType := py.controller.CalcPatternType(biCards[:len(biCards)-1], biCards[len(biCards)-1], nil)
			if biPType == PTUknown {
				logger.Error("DoChiAfter:比的牌不能构成一个模式：靠牌和吃牌如下：")
				PrintCards(biCards[:len(biCards)-1])
				PrintCard(biCards[len(biCards)-1])
				return false
			}
			chiBiResult = append(chiBiResult, NewPattern(biPType, biCards))
		}

		py.SendActionACK(AChi, card, chiBiResult, ACSuccess)

		//让此玩家出牌
		py.SendActionNotifyACK(AChu, nil, nil, nil)

		return true
	}

	//没有玩家吃
	if _, have := room.GetWatingActionPlayer([]int32{AChi}); !have {

		room.ResetAllAction()

		if room.state == RSBankerChuPaiAfterStage {
			room.SwitchRoomState(RSLoopWorkStage)
		}

		room.CheckAndAddShowCard(room.activeCard)

		room.CheckDoAction(nil)
	} else {
		if !isGuo {
			player.SendActionACK(AChi, nil, nil, ACWaitingOtherPlayer)
		}
	}

	return false
}

//执行出牌
func (player *DaerPlayer) DoChuAfter(card *DaerCard) {
	//检测是不是在一个房间里
	room := player.room
	if room == nil {
		logger.Error("room is nil")
		return
	}

	if card == nil {
		logger.Error("不能出一个空牌")
		return
	}

	if !room.IsActivePlayer(player) {
		logger.Error("不是活动玩家不能出牌")
		return
	}

	//执行出
	if chuPai := player.controller.ChuPai(card); chuPai != nil {

		room.ResetAllAction()

		player.SendActionACK(AChu, chuPai, nil, ACSuccess)

		if room.state == RSBankerJinPaiStage {
			room.SwitchRoomState(RSBankerChuPaiAfterStage)
		} else if room.state == RSBankerBaiStage {
			room.SwitchRoomState(RSBankerBaoStage)
		}

		room.CheckDoAction(chuPai)
	} else {

		player.SendActionACK(AChu, card, nil, AOccursError)

		player.PlayerDoAction(ATuoGuan, nil, nil, nil)
	}
}

//过牌
func (player *DaerPlayer) DoGuo(card *DaerCard) {

	if !player.HaveWaitingDoAction() {
		logger.Error("DaerPlayer:等待执行的动作为空。所以不能过任何动作！")
		return
	}

	if player.mode == cmn.Manual {

		wa := player.watingAction

		player.ResetDoAction()

		logger.Info("DoGuo：手动过的动作：", actionName[wa])
		switch wa {
		case AHu:
			player.SendActionACK(AGuo, nil, nil, ACSuccess)

			player.DoHuAfter(card, true)
			logger.Info("DoGuo: 过胡:", player.client.GetName())

		case ABao:
			player.SendActionACK(AGuo, nil, nil, ACSuccess)

			player.DoBaoAfter()
			logger.Info("DoGuo: 过报:", player.client.GetName())

		case ASanLongBai:
			fallthrough
		case ASiKanBai:
			fallthrough
		case AHeiBai:
			player.SendActionACK(AGuo, nil, nil, ACSuccess)

			player.DoBaiAfter(wa, true)
			logger.Info("DoGuo: 过摆:", player.client.GetName())

		case AChu:
			autoChu := player.controller.GetChuPai()
			if autoChu == nil {
				logger.Error("DoGuo:玩家手里竟然没有牌了:", player.client.GetName())
				return
			}

			player.DoChuAfter(autoChu)
			logger.Info("DoGuo: 过出牌:", player.client.GetName(), autoChu.value, autoChu.big)

		case AChi:

			player.SendActionACK(AGuo, nil, nil, ACSuccess)

			player.AddGuoCard(card)

			player.curChiCard = nil
			player.curKaoCards = nil
			player.curBiCards = nil
			player.sendedChiBiMsg = nil

			//logger.Error("===========不应该走手动模式啊！！！")

			player.DoChiAfter(true)

			logger.Info("DoGuo: 过吃:", player.client.GetName())

		default:
			logger.Error("DoGuo:其他状态不能过", actionName[wa])
			debug.PrintStack()

		}

	} else {
		wa := player.watingAction

		player.SwitchReadyDoAction(wa)

		logger.Info("DoGuo：自动过的动作：", actionName[wa])
		switch wa {
		case ABao:
			player.DoBaoAfter()
			logger.Info("DoGuo: 过报:", player.client.GetName())

		case ASanLongBai:
			fallthrough
		case ASiKanBai:
			fallthrough
		case AHeiBai:
			player.DoBaiAfter(wa, true)
			logger.Info("DoGuo: 过四摆:", player.client.GetName())

		case AChu:
			autoChu := player.controller.GetChuPai()
			if autoChu == nil {
				logger.Error("DoGuo:玩家手里竟然没有牌了:", player.client.GetName())
				return
			}

			player.DoChuAfter(autoChu)
			logger.Info("DoGuo: 过出牌:", player.client.GetName(), autoChu.value, autoChu.big)

		case AChi:
			player.DoChiAfter(true)
			logger.Info("DoGuo: 过吃:", player.client.GetName(), card.value, card.big)

		case AHu:
			player.DoHuAfter(card, true)
			logger.Info("DoGuo: 过胡:", player.client.GetName())

		default:
			logger.Error("DoGuo:其他状态不能过", actionName[wa])
			debug.PrintStack()
		}
	}
}

//获取要胡的牌
func (player *DaerPlayer) GetHuCards() []*DaerCard {

	controller := player.controller
	//检查又没有胡的模式组
	if len(controller.huController.patternGroups) <= 0 {
		return nil
	}

	//统计胡的牌
	result := []*DaerCard{}
	for _, patternGroup := range controller.huController.patternGroups {

		patternGroupHu := patternGroup.Value()

		for j := 0; patternGroup.huCards != nil && j < len(patternGroup.huCards); j++ {

			//检查胡的牌是否有效
			huCard := patternGroup.huCards[j]
			huValue, huPatternType, ok := patternGroup.HuValue(huCard, player.controller)
			if !ok {
				continue
			}
			//检查胡数是否超出最大胡数
			var fixePatternHuValue int32 = 0
			if huPatternType == PTZhao {
				fixePatternHuValue = player.GetHuValueForFixedPattern(huCard)
			} else {
				fixePatternHuValue = player.GetHuValueForFixedPattern(nil)
			}
			finalPatternGroupHuValue := fixePatternHuValue + patternGroupHu + huValue

			logger.Info("胡牌：")
			PrintCard(huCard)
			logger.Info("胡的值是：", fixePatternHuValue, patternGroupHu, huValue, finalPatternGroupHuValue)
			if finalPatternGroupHuValue != 0 && finalPatternGroupHuValue < MaxCanHu {
				continue
			}

			isExist := false
			for i := 0; i < len(result); i++ {
				if result[i].IsEqual(huCard) {
					isExist = true
					break
				}
			}

			if !isExist {
				result = append(result, huCard)
			}
		}
	}

	return result
}

//获取指定牌最大的胡的模式组
func (player *DaerPlayer) GetMaxHuOfPatternGroupByCard(card *DaerCard) (score int32, result *DaerPatternGroup) {
	//检查参数的合法性
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	controller := player.controller
	//获取最大组模式
	if len(controller.huController.patternGroups) <= 0 {
		return
	}

	for _, patternGroup := range controller.huController.patternGroups {

		huValue, huPatternType, ok := patternGroup.HuValue(card, controller)
		if !ok {
			continue
		}

		var fixePatternHuValue int32 = 0
		if huPatternType == PTZhao {
			fixePatternHuValue = player.GetHuValueForFixedPattern(card)
		} else {
			fixePatternHuValue = player.GetHuValueForFixedPattern(nil)
		}

		finalPatternGroupHuValue := fixePatternHuValue + patternGroup.Value() + huValue
		if finalPatternGroupHuValue != 0 && finalPatternGroupHuValue < MaxCanHu {
			continue
		}

		tempFinalePatternGroup := controller.GenerateTempPatternGroup(patternGroup, patternGroup.kaoCards, card)
		if tempFinalePatternGroup == nil {
			continue
		}

		//胡算算出来的最大分数
		huMaxScore := GetScoreByHu(finalPatternGroupHuValue)

		//统计番数
		multipleCount, _ := controller.StatisticsRemainMinTang(tempFinalePatternGroup)
		for mt, mul := range multipleCount {
			if mul > 0 && HaveFan(mt) {
				huMaxScore *= int32(math.Pow(2, float64(mul)))
			}
		}

		if huMaxScore > score {
			score = huMaxScore
			result = patternGroup
		}
	}

	return
}

//获取最大胡的模式组
func (player *DaerPlayer) GetMaxHuOfPatternGroup() (result *DaerPatternGroup, huCard *DaerCard) {
	//获取最大组模式
	controller := player.controller
	if len(controller.huController.patternGroups) <= 0 {
		return
	}

	var curHuMaxScore int32 = 0
	for _, patternGroup := range controller.huController.patternGroups {
		for _, card := range patternGroup.huCards {

			huValue, huPatternType, ok := patternGroup.HuValue(card, controller)
			if !ok {
				continue
			}

			var fixePatternHuValue int32 = 0
			if huPatternType == PTZhao {
				fixePatternHuValue = player.GetHuValueForFixedPattern(card)
			} else {
				fixePatternHuValue = player.GetHuValueForFixedPattern(nil)
			}

			finalPatternGroupHuValue := fixePatternHuValue + patternGroup.Value() + huValue
			if finalPatternGroupHuValue != 0 && finalPatternGroupHuValue < MaxCanHu {
				continue
			}

			tempFinalePatternGroup := controller.GenerateTempPatternGroup(patternGroup, patternGroup.kaoCards, card)
			if tempFinalePatternGroup == nil {
				continue
			}

			//胡算算出来的最大分数
			huMaxScore := GetScoreByHu(finalPatternGroupHuValue)

			//统计番数
			multipleCount, _ := controller.StatisticsRemainMinTang(tempFinalePatternGroup)
			for mt, mul := range multipleCount {
				if mul > 0 && HaveFan(mt) {
					huMaxScore *= int32(math.Pow(2, float64(mul)))
				}
			}

			if huMaxScore > curHuMaxScore {
				curHuMaxScore = huMaxScore
				result = patternGroup
				huCard = card
			}
		}
	}

	return
}

//能摆牌，需要统计这组摆牌的最大分数,
func (player *DaerPlayer) GetMaxBaiOfPatternGroup() (score int32, result *DaerPatternGroup) {
	//缓存变量
	huController := player.controller.huController

	//黑摆只算归，分数是固定的，其他的摆牌算胡子和名堂（但出去红牌名堂）
	if player.HaveHeiBai() {
		score = SpecificHuScore[MTHeiBai]
		guiAmount := player.GetGuiAmount()

		score *= int32(math.Pow(2, float64(guiAmount)))
	} else {
		//统计固定牌的胡子
		huValue := player.GetHuValueForFixedPattern(nil)

		score = GetScoreByHu(huValue)

		//更新胡的模式组
		huController.UpdateData(player.cards)

		//通过名堂计算最后的分数
		allPaternG := huController.allpatternGroups
		for _, pg := range allPaternG {
			_, fanCount := player.controller.StatisticsRemainMinTang(pg)
			curPGHuScore := GetScoreByHu(huValue+pg.Value()) * int32(math.Pow(2, float64(fanCount)))
			if curPGHuScore > score {
				score = curPGHuScore
				result = pg
			}
		}
	}

	return
}

//获取归的数量
func (player *DaerPlayer) GetGuiAmount() (result int32) {

	if player == nil {
		logger.Error("Controller.GetGuiAmount:player is nil.")
		return
	}

	cards := player.GetAllCardsOfPlayer()
	if cards != nil {
		s, b := StatisticsCardAmount(cards)
		for _, amount := range s {
			if amount >= 4 {
				result++
			}
		}

		for _, amount := range b {
			if amount >= 4 {
				result++
			}
		}
	} else {
		logger.Error("玩家没有牌")
	}

	return
}

//获取玩家的所有卡牌
func (player *DaerPlayer) GetAllCardsOfPlayer() (result []*DaerCard) {
	if player == nil {
		logger.Error("Controller.GetGuiAmount:player is nil.")
		return
	}

	result = make([]*DaerCard, 0)

	for _, p := range player.showPatterns {
		result = append(result, p.cards...)
	}

	for _, p := range player.fixedpatterns {
		result = append(result, p.cards...)
	}

	result = append(result, player.cards...)

	return
}

//获取固定组合的胡数
func (player *DaerPlayer) GetHuValueForFixedPattern(excludeKan *DaerCard) (huValue int32) {
	//获取显示到桌面的牌的胡数
	huValue = player.GetHuValueForShowPattern()

	//再加上固定牌（坎牌）的胡数
	for _, pattern := range player.fixedpatterns {
		if excludeKan != nil && pattern.cards != nil && len(pattern.cards) > 0 && pattern.cards[0].IsEqual(excludeKan) {
			continue
		}
		huValue += pattern.value()
	}

	return
}

//获取已显示牌的胡数
func (player *DaerPlayer) GetHuValueForShowPattern() (huValue int32) {
	huValue = 0
	for _, pattern := range player.showPatterns {
		huValue += pattern.value()
	}

	return
}

//切换手动或自动模式
func (player *DaerPlayer) SwitchControllMode(mode int) {
	player.mode = int32(mode)
}

//切换等待动作
func (player *DaerPlayer) SwitchWatingAction(watingAction int32) {
	logger.Info("切换等待动作到：", actionName[watingAction])
	if watingAction != ANone {
		player.readyDoAction = ANone
	}
	player.watingAction = watingAction
}

//有等待执行的动作吗
func (player *DaerPlayer) HaveWaitingDoAction() bool {
	return player.watingAction != ANone
}

//切换准备执行的动作
func (player *DaerPlayer) SwitchReadyDoAction(readyDoAction int32) {
	//logger.Info("切换准备执行动作到：", actionName[readyDoAction])
	if readyDoAction != ANone {
		player.watingAction = ANone
	}
	player.readyDoAction = readyDoAction
}

//设置延迟执行的动作
func (player *DaerPlayer) SetDelayDoAction(action int32) {
	player.delayDoAction = action
}

//重置动作状态
func (player *DaerPlayer) ResetDoAction() {
	player.SwitchWatingAction(ANone)
	player.SwitchReadyDoAction(ANone)
}

//是否已经胡牌了
func (player *DaerPlayer) IsHu(isLimitScoreMaxValue bool) (ok bool, huScore int32) {

	//检查参数是否合法
	room := player.room
	if room == nil {
		logger.Error("room is nil")
		return
	}

	controller := player.controller
	if controller == nil {
		logger.Error("controller is nil.")
		return
	}

	finalPatternGroup := controller.GenerateFinalPatternGroup()
	if finalPatternGroup == nil {
		logger.Error("生成最终的模式组出错。")
		return
	}

	//摆
	if player.HaveBai() {
		ok = true
		PrintPatternGroupS("检查是否是摆牌胡：", finalPatternGroup, false)
		_, huScore = player.controller.StatisticsHuAndScore(finalPatternGroup, isLimitScoreMaxValue)
		return
	}

	//正常胡
	if room.state == RSSettlement && (player.cards == nil || len(player.cards) <= 0) {
		ok = true
		_, huScore = controller.StatisticsHuAndScore(finalPatternGroup, isLimitScoreMaxValue)
		return
	}

	return
}

//有叫吗
func (player *DaerPlayer) HaveJiao() bool {
	huC := player.controller.huController
	return huC.patternGroups != nil && len(huC.patternGroups) > 0
}

//是否是查叫
func (player *DaerPlayer) IsChaJiao() bool {
	return player.HaveSpecificMingTang(MTChaJiao)
	// val, exist := player.multipleCount[MTChaJiao]
	// return exist && val > 0
}

//是否有指定的坎牌
func (player *DaerPlayer) HaveSpecificKan(card *DaerCard) bool {
	if card == nil {
		return false
	}

	for _, fp := range player.fixedpatterns {
		if fp == nil {
			continue
		}

		if fp.ptype == PTKan && fp.cards != nil && len(fp.cards) > 0 && fp.cards[0].IsEqual(card) {
			return true
		}
	}

	return false
}

//是否是庄家
func (player *DaerPlayer) IsBanker() bool {
	return player.ptype == cmn.PTBanker
}

//有摆吗
func (player *DaerPlayer) HaveBai() bool {

	result := false
	if result = player.HaveSpecificMingTang(MTSanLongBai); result {
		return true
	} else if result = player.HaveSpecificMingTang(MTSiKanBai); result {
		return true
	} else if result = player.HaveSpecificMingTang(MTHeiBai); result {
		return true
	} else {
		return false
	}

	// val, exist := player.multipleCount[MTSanLongBai]
	// if exist && val > 0 {
	// 	return true
	// }

	// val, exist = player.multipleCount[MTSiKanBai]
	// if exist && val > 0 {
	// 	return true
	// }

	// val, exist = player.multipleCount[MTHeiBai]
	// if exist && val > 0 {
	// 	return true
	// }

	//return false
}

//有黑摆吗
func (player *DaerPlayer) HaveHeiBai() bool {
	return player.HaveSpecificMingTang(MTHeiBai)
	// val, exist := player.multipleCount[MTHeiBai]
	// if exist && val > 0 {
	// 	return true
	// }

	// return false
}

//有非黑摆吗(三拢摆或四坎摆)
func (player *DaerPlayer) HaveBaiAndNotHeiBai() bool {
	return player.HaveBai() && !player.HaveHeiBai()
}

//有报吗
func (player *DaerPlayer) HaveBao() bool {
	return player.HaveSpecificMingTang(MTBaoPai)
	// val, exist := player.multipleCount[MTBaoPai]
	// if exist && val > 0 {
	// 	return true
	// }

	// return false
}

//有指定名堂吗
func (player *DaerPlayer) HaveSpecificMingTang(mingtang int32) bool {
	val, exist := player.multipleCount[mingtang]
	if exist && val > 0 {
		return true
	}

	return false
}

//获得牌（庄家进的第一张）
func (player *DaerPlayer) ObtainCard(card *DaerCard) (resultCanZhao bool) {
	//检查输入参数的合法性
	if card == nil {
		logger.Error("庄家进的第一张是nil.")
		return false
	}

	if player.controller == nil {
		logger.Error("player.controller is nil.")
		return false
	}

	if !player.HaveBao() {
		//在手牌里添加一张新拍
		card.owner = player
		player.cards = append(player.cards, card)

		//在进行一次拢牌和坎牌
		player.controller.Long()
		player.controller.StripKan()
	} else {
		if canZhao, _ := player.controller.CheckZhao(card); canZhao {

			if pattern, _ := player.controller.Zhao(card); pattern != nil {
				player.SendActionACK(AZhao, card, []*DaerPattern{pattern}, ACSuccess)
				resultCanZhao = true
			}

			//player.SendActionNotifyACK(doAction, card, nil, nil)
			//player.controller.Zhao(card)
		} else {
			//在手牌里添加一张新拍
			card.owner = player
			player.cards = append(player.cards, card)
		}
	}

	//手牌变了后需要从新更新hu控制器
	player.controller.huController.UpdateData(player.cards)

	return
}

//获取上家
func (player *DaerPlayer) GetShangJia() *DaerPlayer {
	room := player.room
	if room == nil {
		logger.Error("player.room is nil.")
		return nil
	}

	curPlayerIndex := room.GetPlayerIndex(player)
	if curPlayerIndex >= 0 {
		curPlayerIndex--
		shangJiaIndex := (curPlayerIndex + RoomMaxPlayerAmount) % RoomMaxPlayerAmount
		//logger.Info("player.GetShangJia Index:.", shangJiaIndex)
		return room.players[shangJiaIndex]
	}

	return nil
}

//获取下家
func (player *DaerPlayer) GetXiaJia() *DaerPlayer {
	room := player.room
	if room == nil {
		logger.Error("player.room is nil.")
		return nil
	}

	curPlayerIndex := room.GetPlayerIndex(player)
	if curPlayerIndex >= 0 {
		curPlayerIndex++
		xiaJiaIndex := curPlayerIndex % RoomMaxPlayerAmount
		logger.Info("player.GetXiaJia Index:.", xiaJiaIndex)
		return room.players[xiaJiaIndex]
	}

	return nil
}

//获取拢的数量
func (player *DaerPlayer) GetLongAmount() int32 {
	var result int32 = 0
	for _, pattern := range player.showPatterns {
		if pattern.ptype == PTLong || pattern.ptype == PTZhao {
			result++
		}
	}

	return result
}

//增加一张玩家过牌(没有胡，摆，找，碰，吃的牌需要添加到过的列表中)
func (player *DaerPlayer) AddPassCard(card *DaerCard) {
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	cloneCard := *card
	//	if IsExistCard(player.passCards, &cloneCard) {
	//		return
	//	}

	cloneCard.flag = cmn.CLock | cmn.CPositive

	player.showCards = append(player.showCards, &cloneCard)

	player.sendPassCardNotifyACK(&cloneCard)

}

//增加一张过牌
func (player *DaerPlayer) AddGuoCard(card *DaerCard) {
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	cloneCard := *card
	if IsExistCard(player.guoCards, &cloneCard) {
		return
	}

	cloneCard.flag = cmn.CLock | cmn.CPositive

	player.guoCards = append(player.guoCards, &cloneCard)
}

//检测一个动作时候需要等待、
func IsWaitingAction(action int32) bool {
	return action == AChu || action == AChi ||
		action == AHu || action == ABao || action == ASanLongBai ||
		action == ASiKanBai || action == AHeiBai
}

//是不是胡的动作
func IsHuAction(action int32) bool {
	return action == AHu || action == ASanLongBai || action == ASiKanBai || action == AHeiBai
}

//是不是摆的动作
func IsBaiAction(action int32) bool {
	return action == ASanLongBai || action == ASiKanBai || action == AHeiBai
}

//
//网络消息相关函数
//

//发送战斗开始
func (player *DaerPlayer) SendGameStartACK(reEnter bool) {
	msg := &rpc.GameStartACK{}

	//设置战斗状态
	room := player.room
	if room == nil {
		return
	}

	if room.state == RSReady {
		msg.SetFightState(cmn.FSReady)
	} else if room.state == RSSettlement {
		msg.SetFightState(cmn.FSSettlement)
	} else {
		msg.SetFightState(cmn.FSFighting)
	}

	logger.Info("发送房间的状态：：：：：：：：：", msg.GetFightState())

	//组织FightPlayerInfo结构
	for _, p := range player.room.players {
		if p != nil {
			fgtPlayerInfo := fillPlayerMsg(p, player.id == p.id, reEnter)
			msg.FightPlayersInfo = append(msg.FightPlayersInfo, fgtPlayerInfo)
			logger.Info("玩家的战斗信息:Name:%s, Banker:%s", p.GetPlayerBasicInfo().GetName(), fgtPlayerInfo.GetBZhuang())
		}
	}

	//组织FightCurrentStateInfo结构

	msgc := &rpc.FightCurrentStateInfo{}

	//填充桌面上要显示的牌-
	if room.state == RSBankerChuPaiAfterStage {
		banker := room.GetBanker()
		if banker != nil && !banker.HaveWaitingDoAction() {
			msgc.SetCurrentDeskCard(convertCard(room.activeCard))
			msgc.SetBCurrentDeskCardMo(false)
		}
	} else if room.state == RSLoopWorkStage {
		if room.activeCard != nil {
			msgc.SetCurrentDeskCard(convertCard(room.activeCard))
			if room.activeCard.owner == nil {
				msgc.SetBCurrentDeskCardMo(true)
			}
		}
	}

	//填充倒计时
	if player.HaveWaitingDoAction() {
		countDown := &rpc.CountDown{}
		countDown.SetPlayerID(player.id)
		countDown.SetCurrentCountDown(room.GetRemainTime())
		msgc.CurrentCountDownInfo = append(msgc.CurrentCountDownInfo, countDown)
	} else {
		for _, p := range room.players {
			if p != nil && p.watingAction != ANone {
				countDown := &rpc.CountDown{}
				countDown.SetPlayerID(p.id)
				countDown.SetCurrentCountDown(room.GetRemainTime())
				msgc.CurrentCountDownInfo = append(msgc.CurrentCountDownInfo, countDown)
				break
			}
		}
	}

	//填充当前活动的玩家
	if room.activeCard != nil && room.activeCard.owner != nil {
		msgc.SetCurrentDeskCardPlayerID(room.activeCard.owner.id)
	} else {
		ap := room.GetActivePlayer()
		if ap != nil {
			msgc.SetCurrentDeskCardPlayerID(ap.id)
		}
	}

	msgc.SetCurrentDeskRemainCard(int32(len(room.ownCards)))

	logger.Info("当前做面的牌数：", msgc.GetCurrentDeskRemainCard())

	msg.SetCurrentFightState(msgc)

	if err := conn.SendGameStart(player.id, msg); err != nil {
		logger.Error("发送游戏开始出错：", err, msg)
	}

	logger.Info("玩家名字：", player.GetPlayerBasicInfo().GetName())
	PrintCardsS("开始游戏时，玩家的手牌：", player.cards)

	//如果是重登并且当前玩家有等待执行的动作，需要把这个动作通知给客服端
	if reEnter {
		logger.Info("重登录时，玩家等待的动作：", actionName[player.watingAction], player.sendedChiBiMsg)
		PrintCardsS("重登录时，玩家等待的手牌====：", player.cards)

		if player.watingAction != ANone && player.sendedChiBiMsg != nil {
			if err := conn.SendActionNotify(player.id, player.sendedChiBiMsg); err != nil {
				logger.Error("发送恢复动作出错：", err, msg)
			}
			//player.sendedChiBiMsg = nil
		}
	}
}

//扣取金币
func (player *DaerPlayer) JieSuanCoin(jiesuanCoin []*rpc.JieSuanCoin) {
	if jiesuanCoin == nil || len(jiesuanCoin) <= 0 {
		logger.Error("没有结算金币的玩家")
		return
	}

	for _, jiesuanInfo := range jiesuanCoin {
		uid := jiesuanInfo.GetPlayerID()
		coin := jiesuanInfo.GetCoin()
		if uid == player.id {
			//通知gameserver扣钱
			if err := conn.SendCostResourceMsg(uid, connector.RES_COIN, "daer", coin); err != nil {
				logger.Error("发送扣取金币出错：", err, uid, coin)
				continue
			}
			player.client.SetCoin(player.client.GetCoin() + coin)
		}
	}

}

//填充金币结算新
func fillJieSuanCoin(jieSuanPlayer *DaerPlayer, huScore int32, isCorrection bool) (result []*rpc.JieSuanCoin) {

	//检查参数合法性
	if jieSuanPlayer == nil {
		return nil
	}

	room := jieSuanPlayer.room
	if room == nil {
		return nil
	}

	//修正最大倍数

	//获取底分
	var DiFen int32 = room.Difen
	logger.Info("结算时的底分：%d, Score:%d", DiFen, huScore)
	//	cfg := cmn.GetDaerRoomConfig(strconv.Itoa(int(room.rtype)))
	//	if cfg != nil {
	//		DiFen = cfg.Difen
	//	} else {
	//		logger.Error("读取房间配置表出错ID：%s", room.rtype)
	//	}

	result = []*rpc.JieSuanCoin{&rpc.JieSuanCoin{}, &rpc.JieSuanCoin{}, &rpc.JieSuanCoin{}}

	//是否是查叫
	isChaJiao := len(room.ownCards) <= 0
	if isChaJiao {
		for _, p := range room.players {
			if p == nil {
				continue
			}

			if hu, _ := p.IsHu(true); hu && !p.HaveSpecificMingTang(MTChaJiao) {
				isChaJiao = false
				break
			}

			// if val, exist := p.multipleCount[MTHaiDiLao]; exist && val > 0 {
			// 	isChaJiao = false
			// 	break
			// }
		}
	}

	//如果是查叫
	if isChaJiao {
		for i, p := range room.players {
			result[i].SetPlayerID(p.id)
			result[i].SetCoin(0)
			result[i].SetTag(JSNone)

			pIsHu, pHuScore := p.IsHu(true)
			logger.Info("统计结算 玩家%s:  能胡：%s", p.client.GetName(), pIsHu)

			for _, ip := range room.players {
				//不需计算自己的
				if ip.id == p.id {
					continue
				}

				if pIsHu {
					if hu, _ := ip.IsHu(true); !hu {
						//检查是否有下叫的牌（不满10胡不能胡）
						if !ip.HaveJiao() {
							coin := result[i].GetCoin() + int32(pHuScore)*DiFen
							result[i].SetCoin(coin)
						} else {
							logger.Info("统计结算 玩家%s:  能胡：%s，但是胡数不够", ip.client.GetName(), pIsHu)
						}
					}
				} else {
					if !p.HaveJiao() {
						if hu, score := ip.IsHu(true); hu {
							coin := result[i].GetCoin() - int32(score)*DiFen
							result[i].SetCoin(coin)
						}
					}
				}
			}
		}

		//logger.Error("查叫结算:", result)
	} else {

		//别人点炮
		if val, exist := jieSuanPlayer.multipleCount[MTDianPao]; exist && val > 0 {
			result[0].SetPlayerID(jieSuanPlayer.id)
			result[0].SetCoin(int32(huScore) * DiFen) //点炮,自摸都是算了翻的
			result[0].SetTag(JSNone)

			dianPaoPlayer := jieSuanPlayer.getDianPaoPlayer()
			if dianPaoPlayer != nil {
				result[1].SetPlayerID(dianPaoPlayer.id)
				result[1].SetCoin(-int32(huScore) * DiFen)
				result[1].SetTag(JSDianPao)
			} else {
				logger.Error("点炮尽然没有获取到点炮的玩家")
				return
			}

			for _, op := range room.players {
				if op.id == jieSuanPlayer.id || op.id == dianPaoPlayer.id {
					continue
				}

				result[2].SetPlayerID(op.id)
				result[2].SetCoin(0)
				result[2].SetTag(JSNone)
			}

		} else {
			result[0].SetPlayerID(jieSuanPlayer.id)
			result[0].SetCoin(int32(huScore) * DiFen * 2)
			result[0].SetTag(JSZiMo)

			index := 0
			for _, op := range room.players {
				if op.id != jieSuanPlayer.id {
					index++
					if index >= 3 {
						logger.Error("结算玩家错误：", index)
					}
					result[index].SetPlayerID(op.id)
					result[index].SetCoin(-int32(huScore) * DiFen)
					result[index].SetTag(JSNone)
				}
			}
		}
	}

	//根据拥有的金币修改结算
	if isCorrection {
		jieSuanPlayer.CorrectionJieSuan(result)
	}

	//抽成
	jieSuanPlayer.Rake(result)
	return
}

//修正结算
func (player *DaerPlayer) CorrectionJieSuan(jiesuanInfo []*rpc.JieSuanCoin) {
	room := player.room
	if room == nil {
		logger.Error("room is nil.")
		return
	}

	losePlayers := make([]*rpc.JieSuanCoin, 0)
	winPlayers := make([]*rpc.JieSuanCoin, 0)
	for _, js := range jiesuanInfo {
		coin := js.GetCoin()
		if coin < 0 {
			losePlayers = append(losePlayers, js)
		}

		if coin > 0 {
			winPlayers = append(winPlayers, js)
		}
	}

	loseAmout := len(losePlayers)
	winAmount := len(winPlayers)

	if loseAmout == 1 && winAmount == 1 {
		losePlayer := room.GetPlayerByID(losePlayers[0].GetPlayerID())
		winPlayer := room.GetPlayerByID(winPlayers[0].GetPlayerID())

		chaZi := losePlayer.client.GetCoin() + losePlayers[0].GetCoin()
		if losePlayer != nil && winPlayer != nil && chaZi < 0 {
			winPlayers[0].SetCoin(winPlayers[0].GetCoin() + chaZi)
			losePlayers[0].SetCoin(losePlayers[0].GetCoin() - chaZi)
			losePlayers[0].SetTag(JSPoChan)
		}
	} else if loseAmout == 2 && winAmount == 1 {
		winPlayer := room.GetPlayerByID(winPlayers[0].GetPlayerID())

		var totalChaZi int32 = 0
		for _, lp := range losePlayers {
			losePlayer := room.GetPlayerByID(lp.GetPlayerID())
			chaZi := losePlayer.client.GetCoin() + lp.GetCoin()
			if losePlayer != nil && winPlayer != nil && chaZi < 0 {
				totalChaZi += chaZi
				lp.SetCoin(lp.GetCoin() - chaZi)
				lp.SetTag(JSPoChan)
			}
		}

		winPlayers[0].SetCoin(winPlayers[0].GetCoin() + totalChaZi)

	} else if loseAmout == 1 && winAmount == 2 {
		losePlayer := room.GetPlayerByID(losePlayers[0].GetPlayerID())

		chaZi := losePlayer.client.GetCoin() + losePlayers[0].GetCoin()

		if chaZi < 0 {
			perChaZi := chaZi / int32(winAmount)
			for _, wp := range winPlayers {
				wp.SetCoin(wp.GetCoin() + perChaZi)
			}

			losePlayers[0].SetCoin(losePlayers[0].GetCoin() - chaZi)
			losePlayers[0].SetTag(JSPoChan)
		}
	}
}

//抽成
func (player *DaerPlayer) Rake(jiesuanInfo []*rpc.JieSuanCoin) {

	for _, jieSuanCoin := range jiesuanInfo {
		coin := float32(jieSuanCoin.GetCoin())
		if coin > 0 && player.room != nil {
			coin *= (1.0 - float32(player.room.RakeRate)/100.0)
		}

		jieSuanCoin.SetCoin(int32(coin))
	}

}

//获取指定模式的数量
func (player *DaerPlayer) GetPatternCount(pt int32) int32 {

	var patternAmount int32 = 0

	if player.showPatterns != nil && len(player.showPatterns) > 0 {
		for _, v := range player.showPatterns {
			if v.ptype == uint(pt) {
				patternAmount++
			}
		}
	}

	if player.fixedpatterns != nil && len(player.fixedpatterns) > 0 {
		for _, v := range player.fixedpatterns {
			if v.ptype == uint(pt) {
				patternAmount++
			}
		}
	}

	return patternAmount
}

//获取点炮的玩家
func (player *DaerPlayer) getDianPaoPlayer() *DaerPlayer {
	allCards := player.GetAllCardsOfPlayer()
	for _, c := range allCards {
		if c.owner != nil && c.owner.id != player.id {
			return c.owner
		}
	}

	return nil
}

//填充名堂
func (player *DaerPlayer) fillMingTang() []*rpc.MingTang {

	mingtang := make([]*rpc.MingTang, 0)

	for mt, mtVal := range player.multipleCount {
		if mtVal > 0 {
			rpcMt := &rpc.MingTang{}
			rpcMt.SetMingTang(int32(mt))
			rpcMt.SetValue(int32(mtVal))
			mingtang = append(mingtang, rpcMt)
		}
	}

	return mingtang
}

//填充战斗开始信息
func fillPlayerMsg(player *DaerPlayer, isSelf bool, reEnter bool) *rpc.FightPlayerInfo {
	//组织FightPlayerInfo结构
	msgc := &rpc.FightPlayerInfo{}
	msgc.SetPlayerID(player.id)
	msgc.SetCurrentHu(int32(player.GetHuValueForShowPattern()))
	msgc.SetBZhuang(player.ptype == cmn.PTBanker)
	msgc.SetBBao(player.multipleCount[MTBaoPai] > 0)
	msgc.SetBTuoGuan(player.mode == cmn.Auto)

	if !reEnter {
		msgc.LongPattern = convertPatterns(player.showPatterns)
		if isSelf {
			msgc.HandCards = convertCards(player.cards)
			msgc.KanPattern = convertPatterns(player.fixedpatterns)
			msgc.ErLongTouYi = convertCards(player.erLongTouYi)
		}
	} else {
		msgc.ChuGuoCards = convertCards(player.showCards)
		msgc.ChiPengZhaoLongCards = convertPatterns(player.showPatterns)

		if isSelf {
			msgc.HandCards = convertCards(player.cards)
			msgc.KanPattern = convertPatterns(player.fixedpatterns)
		}
	}
	return msgc
}

//转换daer.Card 到protobuff的Card
func convertCards(src []*DaerCard) (dest []*rpc.Card) {
	if src == nil {
		return make([]*rpc.Card, 0)
	}

	dest = make([]*rpc.Card, len(src))
	for i, card := range src {
		dest[i] = convertCard(card)
	}
	return
}

func convertCardsToDaerCards(src []*rpc.Card) (dest []*DaerCard) {
	if src == nil {
		return make([]*DaerCard, 0)
	}
	dest = make([]*DaerCard, len(src))
	for i, card := range src {
		dest[i] = convertCardToDaerCard(card)
	}
	return
}

//转换daer.Card 到protobuff的Card
func convertCard(src *DaerCard) *rpc.Card {
	if src == nil {
		return nil
	}
	rpcCard := &rpc.Card{}
	rpcCard.SetValue(int32(src.value))
	rpcCard.SetBBig(src.big)
	rpcCard.SetBLock((src.flag & cmn.CLock) > 0)
	rpcCard.SetBChi((src.flag & cmn.CChi) > 0)
	rpcCard.SetBHu((src.flag & cmn.CHu) > 0)

	return rpcCard
}

//转换protobuff的card到daer.Card
func convertCardToDaerCard(src *rpc.Card) *DaerCard {
	if src == nil {
		return nil
	}

	card := NewCard(0, int32(src.GetValue()), src.GetBBig())

	if src.GetBLock() {
		card.flag |= cmn.CLock
	} else {
		card.flag |= cmn.CUnknown
	}

	if src.GetBChi() {
		card.flag |= cmn.CChi
	}

	if src.GetBHu() {
		card.flag |= cmn.CHu
	}

	return card
}

//转换daer.Pattern 到protobuff的Pattern
func convertPatterns(src []*DaerPattern) (dest []*rpc.Pattern) {
	if src == nil {
		return make([]*rpc.Pattern, 0)
	}
	dest = make([]*rpc.Pattern, len(src))
	for i, pattern := range src {
		dest[i] = convertPattern(pattern)
	}

	return
}

//转换daer.Pattern 到protobuff的Pattern
func convertPattern(src *DaerPattern) *rpc.Pattern {

	if src == nil {
		return nil
	}
	rpcPattern := &rpc.Pattern{}
	rpcPattern.SetPtype(int32(src.ptype))
	rpcPattern.Cards = convertCards(src.cards)
	rpcPattern.SetHu(src.value())

	return rpcPattern
}

//发送可以执行动作通知到客服端
func (player *DaerPlayer) SendActionNotifyACK(action int32, card *DaerCard, chiBaoPatterns []*DaerPattern, biPatterns map[uint][]*DaerPattern) {
	room := player.room
	if room == nil {
		logger.Error("room is nil.")
		return
	}

	//需要等待的动作
	player.SwitchWatingAction(action)

	//缓存吃时需要吃的牌
	if action == AChi {
		player.CacheChiData(card, chiBaoPatterns, biPatterns)
	}

	logger.Info("发送执行的动作为：%s 是不是等待动作:%s", actionName[action], IsWaitingAction(action))
	if IsWaitingAction(action) {
		if player.mode == cmn.Manual {
			player.room.StartTimer(room.TimerInterval)
		} else {

			//延迟执行这个动作
			delayCallId := player.id + strconv.Itoa(int(action))
			room.StartDelayCallback(delayCallId, room.DoActionDelay, func(data interface{}) {
				player.PlayerDoAction(AGuo, card, nil, nil)
			}, nil)
		}

	} else {
		logger.Info("自动执行的动作：%s,", actionName[action])
		delayCallId := player.id + strconv.Itoa(int(action))
		room.StartDelayCallback(delayCallId, room.DoActionDelay, func(data interface{}) {
			logger.Info("延迟后自动执行的动作：%s, card:", actionName[action], card)
			player.PlayerDoAction(action, card, nil, nil)
		}, nil)
	}

	//向客户端发送消息
	//fmt.Println("向客户端发送触发动作：", actionName[action])
	msg := &rpc.ActionNotifyACK{}
	msg.SetAction(int32(action))
	if action == AChi && chiBaoPatterns != nil && len(chiBaoPatterns) > 0 {
		for _, chiPattern := range chiBaoPatterns {
			chiActionArgs := fillChiMsg(chiPattern, biPatterns[chiPattern.id])
			chiActionArgs.SetCardArgs(convertCard(card))
			msg.ChiAtionArgs = append(msg.ChiAtionArgs, chiActionArgs)
		}
	}

	if err := conn.SendActionNotify(player.id, msg); err != nil {
		logger.Error("发送触发动作通知出错：", err, msg)
	}
	//缓存发送的消息一遍，重进入后恢复
	player.sendedChiBiMsg = msg

	//如果是通知出牌动作，那么发送倒计时的通知消息,因为出牌是单独的没有走CheckCanDoAction
	if action == AChu {
		logger.Info("==______给玩家发送出。。并通知玩家倒计时。。。。", actionName[player.watingAction])
		room.sendCountdownNotifyACK()
	}
}

//发送倒计时通知信息
func (player *DaerPlayer) sendCountdownNotifyACK(cp *DaerPlayer) {
	if cp != nil {
		timerInfo := &rpc.CountDown{}
		timerInfo.SetPlayerID(cp.id)
		timerInfo.SetCurrentCountDown(int32(cp.room.TimerInterval))

		msg := &rpc.CountdownNotifyACK{}
		msg.SetCountDown(timerInfo)
		if err := conn.SendCountdownNotify(player.id, msg); err != nil {
			logger.Error("发送倒计时出错：", err, msg)
		}
	}
}

//通知玩家的过牌
func (player *DaerPlayer) sendPassCardNotifyACK(card *DaerCard) {
	if card == nil {
		logger.Error("card is nil")
		return
	}

	//组织消息
	msg := &rpc.PassCardNotifyACK{}
	msg.SetPlayerID(player.id)
	msg.SetCard(convertCard(card))

	//发送消息
	for _, p := range player.room.players {
		if p == nil {
			continue
		}

		if err := conn.SendPassCardNotify(p.id, msg); err != nil {
			logger.Error("发送过牌时出错：", err, msg)
			continue
		}
	}

}

//通知玩家此牌已过不能吃牌
func (player *DaerPlayer) sendPassedNotifyACK(card *DaerCard) {
	if card == nil {
		logger.Error("card is nil")
		return
	}

	//组织消息
	msg := &rpc.PassedNotifyACK{}
	msg.SetCard(convertCard(card))
	if err := conn.SendPassedNotify(player.id, msg); err != nil {
		logger.Error("发送过牌通知出错：", err, msg)
	}
}

//组装吃的参数
func fillChiMsg(chiPattern *DaerPattern, biPatterns []*DaerPattern) *rpc.ChiAtionArgs {
	chi := &rpc.ChiAtionArgs{}
	chi.SetCanChiCards(convertPattern(chiPattern))
	chi.NeedBiCards = convertPatterns(biPatterns)
	return chi
}

//发送动作执行回复ACK
func (player *DaerPlayer) SendActionACK(action int32, card *DaerCard, patterns []*DaerPattern, code int32) {
	//向客户端发送消息
	if card != nil {
		logger.Info("向客户端发送动作执行结果：%s,   Card:%s, Code:%s", actionName[action], card, code)
	} else {
		logger.Info("向客户端发送动作执行结果：%s Code:%s", actionName[action], code)
	}

	for _, p := range player.room.players {
		if p != nil {
			msg := &rpc.ActionACK{}
			msg.SetAction(int32(action))
			msg.SetPlayerID(player.id)

			if card != nil {
				msg.SetCardArgs(convertCard(card))
			}

			if patterns != nil {
				msg.ChiPengZhaoLongCards = convertPatterns(patterns)
				logger.Info("向客服端发送吃碰招的模式：", msg.ChiPengZhaoLongCards)
			}

			msg.SetUpdateHu(int32(player.GetHuValueForShowPattern()))
			msg.SetCurrenDeskRemianCard(int32(len(player.room.ownCards)))
			msg.SetResult(int32(code))

			if err := conn.SendAction(p.id, msg); err != nil {
				logger.Error("发送动作执行结果出错：", err, msg)
			}

			logger.Info("更新的胡数：ID:%s  HU:%s, DeskRemianCard:", msg.GetPlayerID(), msg.GetUpdateHu(), msg.GetCurrenDeskRemianCard())
		}
	}

}

//保存吃的牌在player身上
func (player *DaerPlayer) CacheChiData(card *DaerCard, chiBaoPatterns []*DaerPattern, biPatterns map[uint][]*DaerPattern) {
	//先把数据清掉
	player.curChiCard = nil
	player.curKaoCards = nil
	player.curBiCards = nil
	player.sendedChiBiMsg = nil

	//再缓存新的数据
	if chiBaoPatterns != nil && len(chiBaoPatterns) > 0 {
		chiPattern := chiBaoPatterns[0]
		player.curChiCard = card
		player.curKaoCards = GetKaoCards(chiPattern, card)

		if biPatterns != nil && len(biPatterns) > 0 {
			if chiBi, ok := biPatterns[chiPattern.id]; ok && chiBi != nil && len(chiBi) > 0 {
				player.curBiCards = chiBi[0].cards
			}
		}

		player.sendedChiBiMsg = nil
	}
}
