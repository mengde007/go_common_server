package majiangserver

import (
	cmn "common"
	"logger"
	"math"
	//"rpc"
	//"strconv"
)

type MaJiangController struct {
	player       *MaJiangPlayer
	huController *HuController
}

func NewController(player *MaJiangPlayer) *MaJiangController {
	controller := new(MaJiangController)
	controller.player = player
	controller.huController = NewHuController(player)
	return controller
}

//检查能报吗
func (controller *MaJiangController) CheckBao() (bool, []*MaJiangCard) {
	huPais := controller.CheckHu(false)
	PrintCardsS("检查报时，能胡的牌：", huPais)
	return huPais != nil && len(huPais) > 0, huPais
}

//报牌
func (controller *MaJiangController) Bao() {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//锁定所有牌
	for _, card := range player.cards {
		card.flag = cmn.CBack | cmn.CLock
	}

	player.multipleCount[MTBao] = MinTangFanShu[MTBao]
}

//检查能否胡牌
func (controller *MaJiangController) CheckHu(isCheckQiHuKeAmount bool) []*MaJiangCard {
	//检查参数的合法性
	player := controller.player
	if player == nil || player.cards == nil || len(player.cards) <= 0 {
		logger.Error("data of player isn't init.")
		return nil
	}

	//通过huController进行胡的计算
	PrintCardsS("更新胡时的手牌：", player.cards)
	PrintPatternsS("更新胡时的已碰或杠的牌：", player.showPatterns)

	controller.huController.UpdateData(player.cards)

	PrintPatternGroupsS("更新胡牌后的模式组：", controller.huController.patternGroups, false)

	return player.GetHuCards(isCheckQiHuKeAmount)
}

//检查能否胡指定的牌
func (controller *MaJiangController) CheckHuSpecific(card *MaJiangCard) (result bool, ke int32) {
	//检查参数
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	player := controller.player
	if player == nil || player.room == nil {
		logger.Error("MaJiangController.player or player room is nil")
		return
	}

	huPais := controller.CheckHu(true)
	if huPais == nil || len(huPais) <= 0 {
		return
	}

	PrintCardS("CheckHuSpecific：检查能胡指定牌：", card)
	PrintCardsS("全部能胡的牌：", huPais)

	if card.cType == HongZhong {
		// controller.huController.UpdateData(player.cards)

		// PrintPatternGroupsS("更新胡牌后的模式组：", controller.huController.patternGroups, false)

		// return player.GetHuCards(isCheckQiHuKeAmount)

		var maxKe int32 = 0
		tempCard := *card
		for _, v := range huPais {
			tempCard.SetHZReplaceValue(v.cType, v.value)
			ke, _ = player.GetMaxHuOfPatternGroupByCard(&tempCard)
			if ke > maxKe {
				maxKe = ke
				result = true
			}
		}

		ke = maxKe
	} else {
		for _, v := range huPais {
			if v.IsEqual(card) {
				ke, _ = player.GetMaxHuOfPatternGroupByCard(card)
				if ke > 0 {
					result = true
				}
			}
		}
	}

	//检查过水没有或升值
	isZiMo := card.owner == nil || card.owner.id == player.id
	if player.aroundState.IsOnlyZiMo() && !isZiMo {
		result = false
	}

	if !isZiMo && !player.aroundState.IsGuoShuiHu() && !player.aroundState.IsShengZhiHu(ke) {
		result = false
	}

	//logger.Error("玩家:%s 检查胡时，过水和升值的情况：是否是自摸：%s, 是否仅自摸：%s, 是否过水了：%s, 是否是升值：%s, 以前的颗数是：%d", player.client.GetName(), isZiMo, player.aroundState.IsOnlyZiMo(), player.aroundState.IsGuoShuiHu(), player.aroundState.IsShengZhiHu(ke), player.aroundState.huKe)

	return
}

//胡牌
func (controller *MaJiangController) Hu(card *MaJiangCard) bool {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return false
	}

	room := player.room
	if room == nil {
		logger.Error("MaJiangController.player.room is nil")
		return false
	}

	//获取最大的胡的模式组，并修改player中的fixedpatterns和cards
	var maxPatternGroup *MaJiangPatternGroup = nil
	if card.IsHongZhong() {
		var maxKe int32 = 0
		huPais := controller.CheckHu(true)
		if huPais == nil || len(huPais) <= 0 {
			logger.Error("不能会牌啊！，怎么会执行胡操作呢！")
			return false
		}

		tempCard := *card
		for _, v := range huPais {
			tempCard.SetHZReplaceValue(v.cType, v.value)
			ke, patternGroup := player.GetMaxHuOfPatternGroupByCard(&tempCard)
			if ke > maxKe {
				maxKe = ke
				maxPatternGroup = patternGroup
				card.SetHZReplaceValue(v.cType, v.value)
			}
		}

	} else {
		_, maxPatternGroup = player.GetMaxHuOfPatternGroupByCard(card)
	}
	if maxPatternGroup == nil {
		logger.Error("没哟胡的模式组哦！", ConvertToWord(card))
		return false
	}

	//如果靠牌是两个，那么检查是否为对子，如果是对子，那么还可以胡碰牌(胡三个的)
	//var lastPattern *MaJiangPattern = nil
	lastPatterns := CalcPatternType(maxPatternGroup.kaoCards, card)
	if len(lastPatterns) <= 0 {
		logger.Error("最后一个/两个模式生成失败！")
		cardList := []*MaJiangCard{card}
		cardList = append(cardList, maxPatternGroup.kaoCards...)
		lastPatterns = append(lastPatterns, NewPattern(PTUknown, cardList, false))
	}
	// if ptype != PTUknown {
	// 	cardList := []*MaJiangCard{card}
	// 	cardList = append(cardList, maxPatternGroup.kaoCards...)
	// 	lastPattern = NewPattern(ptype, cardList, false)
	// } else {
	// 	logger.Error("胡牌的最有一个模式不应该是个未知的模式")
	// 	PrintCardsS("靠牌：", maxPatternGroup.kaoCards)
	// 	PrintCardS("胡牌：", card)
	// 	//异常兼容代码
	// 	cardList := []*MaJiangCard{card}
	// 	cardList = append(cardList, maxPatternGroup.kaoCards...)
	// 	lastPattern = NewPattern(ptype, cardList, false)
	// }

	//修改胡牌的flag,加一个
	card.flag |= cmn.CHu
	if room.state == RSBankerTianHuStage {
		card.flag |= cmn.CTianHu
	} else {
		isZiMo := card.owner == nil || card.owner.id == player.id
		if isZiMo {
			card.flag |= cmn.CZiMoHu
			//检查是否是杠上花
			if player.aroundState.HaveGangShangHuaFlag() {
				card.flag |= cmn.CGangShangHu
			}

		} else {
			card.flag |= cmn.CDianPaoHu

			if card.owner != nil && card.owner.aroundState.HaveGangShangPaoFlag() {
				card.flag |= cmn.CGangShangPaoHu
			}
		}
	}

	//统计最终的名堂和翻数
	_, _, remainMingTang := player.CalcMulitAndKeByPatternGroup(maxPatternGroup, card)
	for k, v := range remainMingTang {
		player.multipleCount[k] = v
	}

	//缓存胡的牌，因为在重登录的时候，需要下发给客户端
	player.huCard = card

	//合并模式
	player.showPatterns = append(player.showPatterns, maxPatternGroup.patterns...)
	if lastPatterns != nil {
		player.showPatterns = append(player.showPatterns, lastPatterns...)
	}

	//清空手牌
	player.cards = []*MaJiangCard{}

	//统计点炮玩家,用于结算统计分数
	if player.isChaJiaoHu {
		for _, p := range room.players {
			if !p.IsHu() && !p.HaveJiao() {
				player.beiHuPlayers = append(player.beiHuPlayers, p)
			}
		}
	} else {
		if player.HaveZiMoFeatureForHu() {
			for _, p := range room.players {
				if !p.IsHu() {
					player.beiHuPlayers = append(player.beiHuPlayers, p)
				}
			}
		} else {
			if card.owner != nil && !card.owner.IsHu() {
				player.beiHuPlayers = append(player.beiHuPlayers, card.owner)
			} else {
				logger.Error("竟然都是杠上炮,抢杠或点炮了，这个牌竟然还不知道是谁的！")
			}
		}
	}

	return true

	// PrintPatternsS("胡牌时玩家的显示牌如下：", player.showPatterns)
	// PrintCardsS("当前玩家的手牌：", player.cards)
	// logger.Info("玩家%s 胡牌了：", player.id, ConvertToWord(card))
}

//检查暗杠牌
func (controller *MaJiangController) CheckAnGang(addiCard *MaJiangCard) (canGang bool, result []*MaJiangCard) {
	result = make([]*MaJiangCard, 0)
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	cardAmountInfo := player.cardAmountInfo
	if addiCard != nil {
		cardAmountInfo = NewCardAmountStatisticsByCards(append(player.cards, addiCard), false)
	}

	//能够硬杠的
	yingGang := cardAmountInfo.GetCardsBySpecificAmount(4, nil)

	//需要贴鬼杠的
	var hongZhongAmount int32 = 0
	if player.IsOpenHongZhongCheck {
		hongZhongAmount = cardAmountInfo.GetCardAmountByType(HongZhong)
	}

	canGangCards := cardAmountInfo.GetCardsBySpecificAmount(int32(math.Max(float64(1), float64(4-hongZhongAmount))), yingGang)
	canGangCards = append(canGangCards, yingGang...)

	types := player.GetCurMayOwnTypes()
	for _, c := range canGangCards {
		if c.cType == HongZhong {
			continue
		}

		if Exist(types, c.cType) {
			result = append(result, c)
		}
	}

	//报牌后检查时候还能够进行杠
	canGang = len(result) > 0
	if canGang {
		result = controller.FilterNoGangForBaoPlayer(result, 4)
	}

	return len(result) > 0, result
}

//进行暗杠牌
func (controller *MaJiangController) AnGang(card *MaJiangCard) (result *MaJiangPattern) {
	return controller.Gang(card, 4)
}

//检查明杠牌
func (controller *MaJiangController) CheckMingGang(card *MaJiangCard) (result, isNeedHongZhong bool) {

	//检查参数
	if card == nil {
		logger.Error("MaJiangController.card is nil")
	}

	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	//红中不能被明杠
	if card.IsHongZhong() {
		return
	}

	//检查过水没有
	if !player.aroundState.IsGuoShuiPengGang(card) {
		return
	}

	cardAmount := player.cardAmountInfo.GetCardAmount(card.cType, card.value)
	if cardAmount <= 0 {
		return
	}

	isNeedHongZhong = cardAmount < int32(3)

	var hongZhongAmount int32 = 0
	if player.IsOpenHongZhongCheck && isNeedHongZhong {
		hongZhongAmount = player.cardAmountInfo.GetCardAmountByType(HongZhong)
	}

	if hongZhongAmount+cardAmount >= 3 {
		types := player.GetCurMayOwnTypes()
		if Exist(types, card.cType) {
			filtered := controller.FilterNoGangForBaoPlayer([]*MaJiangCard{card}, 3)
			result = len(filtered) > 0
			return
		}
	}

	return
}

//执行明扛
func (controller *MaJiangController) MingGang(card *MaJiangCard) (result *MaJiangPattern) {
	return controller.Gang(card, 3)
}

//筛查报牌玩家不能够杠的牌
func (controller *MaJiangController) FilterNoGangForBaoPlayer(gangCards []*MaJiangCard, needCardAmount int32) (result []*MaJiangCard) {
	result = make([]*MaJiangCard, 0)
	if gangCards == nil || len(gangCards) <= 0 {
		return
	}

	result = make([]*MaJiangCard, len(gangCards))
	copy(result, gangCards)

	player := controller.player
	if controller.player == nil {
		logger.Error("MaJiangController.FilterNoGangForBaoPlayer: controller.player is nil")
		return
	}

	if player.HaveBao() {

		tempResult := make([]*MaJiangCard, 0)
		for _, gangCard := range gangCards {
			//拷贝一个副本
			tempCards := make([]*MaJiangCard, len(player.cards))
			copy(tempCards, player.cards)
			//从手牌里移除扛牌
			removedCards := make([]*MaJiangCard, 0)
			//removedHongZhongCards := make([]*MaJiangCard, 0)
			tempCards, removedCards = RemoveCardsByType(tempCards, gangCard.cType, gangCard.value, needCardAmount)

			//检查本牌是否足够，不足够则用红中替代
			needRemovedHongZhongAmount := needCardAmount - int32(len(removedCards))
			if needRemovedHongZhongAmount > 0 {
				tempCards, _ = RemoveCardsByType(tempCards, HongZhong, 0, needRemovedHongZhongAmount)
			}

			huChecker := NewHuController(player)
			huChecker.UpdateData(tempCards)
			haveHu := huChecker.patternGroups != nil && len(huChecker.patternGroups) > 0
			if haveHu {
				tempResult = append(tempResult, gangCard)
			}
		}

		result = tempResult
	}

	return
}

//进行杠牌
func (controller *MaJiangController) Gang(card *MaJiangCard, needCardAmount int32) (result *MaJiangPattern) {
	//检查参数
	if card == nil {
		logger.Error("MaJiangController.AnGang card is nil")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	//红中不能进行暗杠
	if card.IsHongZhong() {
		logger.Error("红中不能够暗杠红中这张牌", ConvertToWord(card))
	}

	//从手牌里移除扛牌
	removedCards := make([]*MaJiangCard, 0)
	removedHongZhongCards := make([]*MaJiangCard, 0)
	player.cards, removedCards = RemoveCardsByType(player.cards, card.cType, card.value, needCardAmount)

	//检查本牌是否足够，不足够则用红中替代
	needRemovedHongZhongAmount := needCardAmount - int32(len(removedCards))
	if needRemovedHongZhongAmount > 0 {
		player.cards, removedHongZhongCards = RemoveCardsByType(player.cards, HongZhong, 0, needRemovedHongZhongAmount)
	}

	//设置红中的替换值并锁定替换
	for _, hongZhongCard := range removedHongZhongCards {
		if hongZhongCard == nil {
			continue
		}

		hongZhongCard.SetHZReplaceValue(card.cType, card.value)
		hongZhongCard.flag = cmn.CLockHongZhongValue | cmn.CLock | cmn.CPositive
	}

	//保存杠牌的结果
	resultCards := append(removedCards, removedHongZhongCards...)
	isMingGang := needCardAmount == 3
	if isMingGang {
		resultCards = append(resultCards, card)
		result = NewPattern(PTGang, resultCards, true)
	} else {
		result = NewPattern(PTAnGang, resultCards, true)
	}

	//添加一个显示的模式
	player.showPatterns = append(player.showPatterns, result)

	//清除过水，升值等标志
	player.aroundState.ClearGuoShuiAndShengZhiFlag(player.HaveBao())

	//记录杠，用于确定是否杠上花或杠上炮
	//player.aroundState.gangCard = card
	player.aroundState.AddGangFlag(card)

	//手牌变了后需要从新更新hu控制器
	controller.huController.UpdateData(player.cards)

	//重现计算缓存的卡牌数量
	player.cardAmountInfo.CalcCardAmountByCards(player.cards, false)

	return
}

//检查补杠
func (controller *MaJiangController) CheckBuGang(addiCard *MaJiangCard) (canBuGang bool, result []*MaJiangCard) {
	result = make([]*MaJiangCard, 0)
	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	//获取可以补杠的牌
	pengCards := player.GetPengCardsForAlready()
	if len(pengCards) <= 0 {
		return
	}

	//获取所有手牌
	cards := player.cards
	if addiCard != nil {
		cards = append(player.cards, addiCard)
	}

	haveHongZhongInHand := false
	for _, handCard := range cards {
		for _, c := range pengCards {
			if handCard.IsEqual(c) {
				if !IsExist(result, handCard) {
					result = append(result, handCard)
				}
			}
		}

		if player.IsOpenHongZhongCheck && !haveHongZhongInHand && handCard.cType == HongZhong {
			for _, c := range pengCards {
				if !IsExist(result, c) {
					result = append(result, c)
				}
			}

			haveHongZhongInHand = true
		}
	}

	return len(result) > 0, result
}

//执行补杠
func (controller *MaJiangController) BuGang(card *MaJiangCard) (buCard *MaJiangCard, result *MaJiangPattern) {
	//检查参数
	if card == nil {
		logger.Error("MaJiangController.card is nil.")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	//检查是否使用红中进行补杠的
	removedCards := make([]*MaJiangCard, 0)
	cType, cValue := card.CurValue()
	//先移除本牌，如果没有再移除红中
	player.cards, removedCards = RemoveCardsByType(player.cards, cType, cValue, 1)
	if len(removedCards) <= 0 {
		//移除一张红中
		player.cards, removedCards = RemoveCardsByType(player.cards, HongZhong, 0, 1)
	}

	//添加一张补杠的牌
	if len(removedCards) != 1 {
		logger.Error("没有此牌%s，如何补杠.被移除的牌的数量：%d", ConvertToWord(card), len(removedCards))
		PrintCardsS("此时玩家手上的牌是：", player.cards)
	} else {
		buCard = removedCards[0]
		if buCard.IsHongZhong() {
			buCard.SetHZReplaceValue(cType, cValue)
			buCard.flag = cmn.CLockHongZhongValue | cmn.CLock | cmn.CPositive
		}

		result = player.AddOneBuGangCard(buCard)
	}

	//手牌变了后需要从新更新hu控制器
	controller.huController.UpdateData(player.cards)

	//重现计算缓存的卡牌数量
	player.cardAmountInfo.CalcCardAmountByCards(player.cards, false)

	return
}

//检查碰
func (controller *MaJiangController) CheckPeng(card *MaJiangCard) (canPeng, isNeedHongZhong bool) {
	//检查参数
	if card == nil {
		logger.Error("MaJiangController.card is nil.")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	//红中不去检查碰
	if card.IsHongZhong() {
		return
	}

	//检查是否过水
	if !player.aroundState.IsGuoShuiPengGang(card) {
		return
	}

	types := player.GetCurMayOwnTypes()
	if Exist(types, card.cType) {
		selfAmount := player.cardAmountInfo.GetCardAmount(card.cType, card.value)
		//没有本牌
		if selfAmount <= 0 {
			return false, false
		}

		isNeedHongZhong = selfAmount < 2
		if !isNeedHongZhong {
			return true, false
		}

		if player.IsOpenHongZhongCheck {
			hongZhongAmount := player.cardAmountInfo.GetCardAmountByType(HongZhong)
			return hongZhongAmount+selfAmount >= 2, true
		}
	}

	return
}

//碰牌
func (controller *MaJiangController) Peng(card *MaJiangCard) (result *MaJiangPattern) {
	//检查参数
	if card == nil {
		logger.Error("MaJiangController.card is nil.")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	//移除手上的牌
	removedCards := make([]*MaJiangCard, 0)
	removedHongZhongCards := make([]*MaJiangCard, 0)

	player.cards, removedCards = RemoveCardsByType(player.cards, card.cType, card.value, 2)
	//检查本牌是否足够，不足够则用红中替代
	needRemovedHongZhongAmount := 2 - int32(len(removedCards))
	if needRemovedHongZhongAmount > 0 {
		player.cards, removedHongZhongCards = RemoveCardsByType(player.cards, HongZhong, 0, needRemovedHongZhongAmount)
	}

	//设置红中的替换值并锁定替换
	for _, hongZhongCard := range removedHongZhongCards {
		if hongZhongCard == nil {
			continue
		}

		hongZhongCard.SetHZReplaceValue(card.cType, card.value)
		hongZhongCard.flag = cmn.CLockHongZhongValue | cmn.CLock | cmn.CPositive
	}

	//保存碰牌的结果
	pengCards := append(removedCards, removedHongZhongCards...)
	pengCards = append(pengCards, card)
	result = NewPattern(PTKan, pengCards, true)

	//添加一个显示的模式
	player.showPatterns = append(player.showPatterns, result)

	//清除过水，升值等标志
	player.aroundState.ClearGuoShuiAndShengZhiFlag(player.HaveBao())

	//手牌变了后需要从新更新hu控制器
	controller.huController.UpdateData(player.cards)

	//重现计算缓存的卡牌数量
	player.cardAmountInfo.CalcCardAmountByCards(player.cards, false)

	return
}

//计算胡牌的列模式
func CalcPatternType(kaoCards []*MaJiangCard, huCard *MaJiangCard) (result []*MaJiangPattern) {
	result = []*MaJiangPattern{}
	if huCard == nil {
		logger.Error("MaJiangController.card is nil.")
		return
	}

	kaoCardsAmount := len(kaoCards)
	if kaoCardsAmount <= 0 || kaoCardsAmount > 4 {
		logger.Error("kaoCardsAmount is empty or greater 4.")
		return
	}

	//检查对子
	cardList := []*MaJiangCard{huCard}
	cardList = append(cardList, kaoCards...)
	switch kaoCardsAmount {
	case 1:
		if kaoCards[0].IsEqual(huCard) {
			result = append(result, NewPattern(PTPair, cardList, false))
			return
		}
	case 2:
		kaoCard1 := kaoCards[0]
		kaoCard2 := kaoCards[1]

		//检查是不是碰胡
		if kaoCard1.IsEqual(kaoCard2) {
			//碰胡
			if kaoCard1.IsEqual(huCard) {
				result = append(result, NewPattern(PTKan, cardList, false))
				return
			}
		}

		//检查顺子
		//检查三张牌是否都花色相同
		if !(kaoCard1.IsSameHuaSe(kaoCard2) && kaoCard1.IsSameHuaSe(huCard)) {
			return
		}

		offset := kaoCard1.value - kaoCard2.value
		switch offset {
		case 1, -1, 2, -2:
			result = append(result, NewPattern(PTSZ, cardList, false))
		}
	case 4:
		AAPatterns, _ := SplitToAA_A(kaoCards)
		if len(AAPatterns) == 2 {
			p1 := AAPatterns[0]
			p2 := AAPatterns[1]
			if huCard.IsEqual(p1.cards[0]) {
				result = append(result, NewPattern(PTKan, append(p1.cards, huCard), false))
				result = append(result, NewPattern(PTPair, p2.cards, false))
			} else if huCard.IsEqual(p2.cards[0]) {
				result = append(result, NewPattern(PTKan, append(p2.cards, huCard), false))
				result = append(result, NewPattern(PTPair, p1.cards, false))
			}
		}

	default:
		logger.Error("最后的单牌数量只能是1， 2， 4个！")
	}

	return
}

//获取一个模式的拷牌
func GetKaoCards(pattern *MaJiangPattern, card *MaJiangCard) (kaoCard []*MaJiangCard) {
	kaoCard = make([]*MaJiangCard, 0)
	for i, c := range pattern.cards {
		if c.IsEqual(card) {
			kaoCard = append(kaoCard, pattern.cards[:i]...)
			kaoCard = append(kaoCard, pattern.cards[i+1:]...)
			break
		}
	}

	return
}

//查叫
func (controller *MaJiangController) ChaJiao() {
	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil")
		return
	}

	huPatternGroup, card := player.GetMaxHuOfPatternGroup()
	if huPatternGroup == nil || card == nil {
		return
	}

	player.isChaJiaoHu = true
	controller.Hu(card)
}

//获取一个出牌
func (controller *MaJiangController) GetChuPai() *MaJiangCard {
	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil.")
		return nil
	}

	if player.cards == nil || len(player.cards) <= 0 {
		return nil
	}

	for i := len(player.cards) - 1; i >= 0; i-- {
		if !player.cards[i].IsLockChu() {
			return player.cards[i]
		}
	}

	return nil
}

//检查能出指定牌吗
func (controller *MaJiangController) CheckChu(card *MaJiangCard) bool {
	//检查参数的合法性
	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil.")
		return false
	}

	//查找出能出的牌
	chuPais := FindCards(player.cards, card.cType, card.value)
	if chuPais == nil || len(chuPais) <= 0 {
		logger.Error("没有此牌：", ConvertToWord(card))
		return false
	}

	var finalChuPai *MaJiangCard = nil

	for _, chuPai := range chuPais {
		if !chuPai.IsLockChu() {
			finalChuPai = chuPai
			break
		}
	}

	if finalChuPai == nil {
		logger.Error("没有此牌,或者此牌已经被锁定了！", ConvertToWord(card))
		return false
	}

	return true
}

//出牌
func (controller *MaJiangController) ChuPai(card *MaJiangCard) (finalChuPai *MaJiangCard) {
	//检查参数的合法性
	player := controller.player
	if player == nil {
		logger.Error("MaJiangController.player is nil.")
		return
	}

	//查找出能出的牌
	chuPais := FindCards(player.cards, card.cType, card.value)
	if chuPais == nil || len(chuPais) <= 0 {
		logger.Error("没有此牌：", ConvertToWord(card))
		return
	}

	for _, chuPai := range chuPais {
		if !chuPai.IsLockChu() {
			finalChuPai = chuPai
			break
		}
	}

	if finalChuPai == nil {
		logger.Error("没有此牌,或者此牌已经被锁定了！", ConvertToWord(card))
		return
	}

	//从手里移除此牌
	//PrintCardsS("出了一张牌之前胡更新器里的牌是：", controller.huController.originCards)
	for i, c := range player.cards {
		if c.IsEqual(finalChuPai) && !c.IsLockChu() {
			player.cards = append(player.cards[:i], player.cards[i+1:]...)
			break
		}
	}

	player.room.activeCard = finalChuPai

	player.AddChuCard(finalChuPai)

	//PrintCardsS("出了一张牌后手里的牌是：", player.cards)
	//PrintCardsS("原来胡更新器里的手牌是：", controller.huController.originCards)
	//手牌变了后需要从新更新hu控制器
	controller.huController.UpdateData(player.cards)

	//重现计算缓存的卡牌数量
	player.cardAmountInfo.CalcCardAmountByCards(player.cards, false)

	return
}

//生成最终的胡牌模式组
func (controller *MaJiangController) GenerateFinalPatternGroup() (patternGroup *MaJiangPatternGroup) {
	//检查输入参数是否合法
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil.")
		return
	}

	//产生最终的胡的模式组
	patternsList := make([]*MaJiangPattern, 0)
	patternsList = append(patternsList, player.showPatterns...)

	if player.cards != nil && len(player.cards) > 0 {
		maxPg, huCard := player.GetMaxHuOfPatternGroup()
		if maxPg != nil && huCard != nil {
			patternsList = append(patternsList, maxPg.patterns...)
			huCard.flag |= cmn.CHu
			//ptype := CalcPatternType(maxPg.kaoCards, huCard)
			lastPatterns := CalcPatternType(maxPg.kaoCards, huCard)
			patternsList = append(patternsList, lastPatterns...)
		} else {
			patternsList = append(patternsList, NewPattern(PTSingle, player.cards, false))
		}
	}

	patternGroup = NewPatternGroup(patternsList)

	return
}
