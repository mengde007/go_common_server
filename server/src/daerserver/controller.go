package daerserver

import (
	cmn "common"
	"fmt"
	"logger"
	"math"
	"rpc"
	//"strconv"
)

type DaerController struct {
	player       *DaerPlayer
	huController *HuController
}

func NewController(player *DaerPlayer) *DaerController {
	controller := new(DaerController)
	controller.player = player
	controller.huController = NewHuController(player)
	return controller
}

//检查能否进行黑摆
func (controller *DaerController) CheckHeiBai(card *DaerCard) (canBaiPai bool, score int32) {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//检查黑牌
	if player.cards == nil || len(player.cards) <= 0 {
		logger.Error("CheckHeiBai:player.cards is nil or empty")
		return
	}

	//检测进的这张是不是红牌
	if card != nil && card.IsRed() {
		return
	}

	//检测拢牌里有没有红色的
	for _, pattern := range player.showPatterns {
		if pattern.cards == nil || len(pattern.cards) <= 0 {
			continue
		}

		if pattern.cards[0].IsRed() {
			return
		}
	}
	//检测坎牌里有没有红色的
	for _, pattern := range player.fixedpatterns {
		if pattern.cards == nil || len(pattern.cards) <= 0 {
			continue
		}

		if pattern.cards[0].IsRed() {
			return
		}
	}

	//检测手里的牌里有没有红色的
	for _, v := range player.cards {
		if v.IsRed() {
			return
		}
	}

	canBaiPai = true
	score, _ = player.GetMaxBaiOfPatternGroup()

	return
}

//进行黑摆
func (controller *DaerController) HeiBai() {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	player.multipleCount[MTHeiBai] = 1
}

//检查能否进行三拢
func (controller *DaerController) CheckSanLongBai() (canBaiPai bool, score int32) {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//检查三拢
	longAmount := player.GetPatternCount(PTLong)
	if longAmount >= ThreeLong {
		canBaiPai = true
		score, _ = player.GetMaxBaiOfPatternGroup()
	}

	return
}

//进行三拢的摆牌
func (controller *DaerController) SanLongBai() {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	player.multipleCount[MTSanLongBai] = 1
}

//检查能否进行四坎摆
func (controller *DaerController) CheckSiKanBai() (canBaiPai bool, score int32) {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//检查四坎
	kanAmount := player.GetPatternCount(PTKan) + player.GetPatternCount(PTLong)
	if kanAmount >= FourKan {
		canBaiPai = true
		score, _ = player.GetMaxBaiOfPatternGroup()
	}

	return
}

//进行四坎的摆牌
func (controller *DaerController) SiKanBai() {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	player.multipleCount[MTSiKanBai] = 1
}

//检测摆（黑，三拢或四坎摆）
func (controller *DaerController) CheckBai(card *DaerCard) (doAction int32, score int32) {

	doAction = ANone

	var maxScore int32 = 0
	canBai, score := controller.CheckSanLongBai()
	if canBai && maxScore < score {
		maxScore = score
		doAction = ASanLongBai
	}

	canBai, score = controller.CheckSiKanBai()
	logger.Info("检查能不能4️坎摆", canBai, score, maxScore)
	if canBai && maxScore < score {
		maxScore = score
		doAction = ASiKanBai
	}

	canBai, score = controller.CheckHeiBai(card)
	if canBai && maxScore < score {
		maxScore = score
		doAction = AHeiBai
	}

	return
}

//检测摆和胡
func (controller *DaerController) CheckBaiOrHu(card *DaerCard) (doAction int32) {
	doAction = ANone
	if card == nil {
		logger.Error("DaerController.CheckBaiOrHu, card is nil.")
		return ANone
	}

	tempDoBai, tempBaiScore := controller.CheckBai(card)
	if tempDoBai != ANone {
		canHu, tempHuScore := controller.CheckHuSpecific(card)
		if canHu {
			if tempBaiScore > tempHuScore {
				doAction = tempDoBai
			} else {
				doAction = AHu
			}
		} else {
			doAction = tempDoBai
		}
	} else {
		canHu, _ := controller.CheckHuSpecific(card)
		if canHu {
			doAction = AHu
		}
	}

	return
}

//检查能报吗
func (controller *DaerController) CheckBao() (bool, []*DaerCard) {
	huPai := controller.CheckHu()
	logger.Info("检查报时，能胡的牌：")
	PrintCards(huPai)
	return huPai != nil && len(huPai) > 0, huPai
}

//报牌
func (controller *DaerController) Bao() {
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

	fmt.Println("bao...")

	player.multipleCount[MTBaoPai] = MinTangFanShu[MTBaoPai]
}

//检查能否胡牌
func (controller *DaerController) CheckHu() []*DaerCard {
	//检查参数的合法性
	player := controller.player
	if player == nil || player.cards == nil || len(player.cards) <= 0 {
		logger.Error("data of player isn't init.")
		return nil
	}

	//通过huController进行胡的计算
	logger.Info("更新胡时的手牌：")
	PrintCards(player.cards)

	controller.huController.UpdateData(player.cards)

	logger.Info("更新胡牌后的模式组：")
	PrintPatternGroups(controller.huController.patternGroups, false)

	return player.GetHuCards()
}

//检查能否胡指定的牌
func (controller *DaerController) CheckHuSpecific(card *DaerCard) (result bool, score int32) {
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	huPai := controller.CheckHu()
	if huPai == nil || len(huPai) <= 0 {
		return
	}

	logger.Info("CheckHuSpecific：检查到能胡指定牌：")
	PrintCard(card)
	logger.Info("全部能胡的牌：")
	PrintCards(huPai)

	for _, v := range huPai {
		if v.IsEqual(card) {
			score, _ = controller.player.GetMaxHuOfPatternGroupByCard(card)
			result = true
			return
		}
	}

	return
}

//胡牌
func (controller *DaerController) Hu(card *DaerCard) {
	//检查参数
	player := controller.player
	if player == nil || player.room == nil {
		logger.Error("controller.player or player room is nil")
		return
	}

	//获取最大的胡的模式组，并修改player中的fixedpatterns和cards
	_, maxPatternGroup := player.GetMaxHuOfPatternGroupByCard(card)
	if maxPatternGroup == nil {
		logger.Error("没哟胡的模式组哦！")
		return
	}

	//如果靠牌是两个，那么检查是否为对子，如果是对子，那么还可以胡坎牌(胡三个的)
	lastPattern := make([]*DaerPattern, 0)
	ptype := controller.CalcPatternType(maxPatternGroup.kaoCards, card, maxPatternGroup)
	if ptype == PTZhao {
		cardList := []*DaerCard{card}
		for i, fixedP := range player.fixedpatterns {
			if fixedP.ptype == PTKan && fixedP.cards != nil && len(fixedP.cards) > 0 && fixedP.cards[0].IsEqual(card) {
				cardList = append(cardList, fixedP.cards...)
				player.fixedpatterns = append(player.fixedpatterns[:i], player.fixedpatterns[i+1:]...)
				break
			}
		}
		lastPattern = append(lastPattern, NewPattern(PTPair, maxPatternGroup.kaoCards))
		lastPattern = append(lastPattern, NewPattern(ptype, cardList))
	} else if ptype != PTUknown {
		cardList := []*DaerCard{card}
		cardList = append(cardList, maxPatternGroup.kaoCards...)
		lastPattern = append(lastPattern, NewPattern(ptype, cardList))
	} else {
		logger.Error("胡牌的最有一个模式不应该是个为止的模式")
	}

	//修改胡牌的flag,加一个
	card.flag |= cmn.CHu

	//合并模式
	player.fixedpatterns = append(player.fixedpatterns, maxPatternGroup.patterns...)
	player.fixedpatterns = append(player.fixedpatterns, lastPattern...)

	//清空手牌
	player.cards = []*DaerCard{}

	//是否是查叫,查叫时，不用统计炸天报 天胡，地胡，水上漂，海底捞，自摸和点炮
	if player.IsChaJiao() {
		return
	}

	//检查是否有报，有则记录报
	isFirstCard := player.room.state == RSBankerJinPaiStage ||
		player.room.state == RSBankerChuPaiAfterStage || player.room.state == RSBankerBaoStage
	if isFirstCard {
		if v, ok := player.multipleCount[MTBaoPai]; ok && v >= 1 {
			if player.ptype == cmn.PTBanker {
				player.multipleCount[MTZhaTianBao] = MinTangFanShu[MTZhaTianBao]
				player.multipleCount[MTBaoPai] = 0
				return
			}
		}

		if player.ptype == cmn.PTBanker {
			player.multipleCount[MTTianHu] = MinTangFanShu[MTTianHu]
		} else {
			player.multipleCount[MTDiHu] = MinTangFanShu[MTDiHu]
		}
	}

	//如果是进牌，则自摸，否则是别人点炮
	if card.IsIncomeCard() {
		//检查是否是水上漂
		remainCardsAmount := len(player.room.ownCards)
		isShuiShangPiao := remainCardsAmount == CardTotalAmount-FirstCardsAmount*RoomMaxPlayerAmount-2
		if isShuiShangPiao {
			player.multipleCount[MTShuiShangPiao] = MinTangFanShu[MTShuiShangPiao]
		}

		//检查是否是海底捞
		isHaiDiLao := remainCardsAmount == 0
		if isHaiDiLao {
			player.multipleCount[MTHaiDiLao] = MinTangFanShu[MTHaiDiLao]
		}

		//检查这张进牌是堂出的牌还是自摸的牌
		if room := player.room; room != nil && room.IsActivePlayer(player) && !player.HaveSpecificMingTang(MTTianHu) {
			player.multipleCount[MTZiMo] = MinTangFanShu[MTZiMo]
		}
	} else {
		player.multipleCount[MTDianPao] = MinTangFanShu[MTDianPao]
	}

	//统计杀报
	baoAmount := player.room.GetAmountOfBaoPai()
	if player.HaveBao() {
		baoAmount--
	}
	player.multipleCount[MTShaBao] = baoAmount * MinTangFanShu[MTShaBao]

	//logger.Error("胡牌时的名堂:%s, 房间状态：%s", player.multipleCount, rootTypeName[player.room.state])

}

//计算胡牌的列模式
func (controller *DaerController) CalcPatternType(kaoCards []*DaerCard, huCard *DaerCard, additionCheckPatternGroup *DaerPatternGroup) uint {
	if huCard == nil {
		logger.Error("card is nil.")
		return PTUknown
	}

	player := controller.player
	if player == nil {
		logger.Error("player is nil.")
		return PTUknown
	}

	kaoCardsAmount := len(kaoCards)
	if kaoCardsAmount <= 0 || kaoCardsAmount >= 3 {
		logger.Error("kaoCardsAmount is empty or greater 3.")
		return PTUknown
	}

	//检查对子
	if kaoCardsAmount == 1 {
		if kaoCards[0].IsEqual(huCard) {
			return PTPair
		} else {
			return PTUknown
		}
	} else {

		kaoCard1 := kaoCards[0]
		kaoCard2 := kaoCards[1]

		//检查是不是招胡和坎胡
		if kaoCard1.IsEqual(kaoCard2) {
			//坎胡
			if kaoCard1.IsEqual(huCard) {
				return PTPeng
			}

			//检查在模式组里面已经有一个对子，有就不能再检查是否是招胡了，招胡后就有两对
			for {
				//检查是否有附加的条件检查
				if additionCheckPatternGroup != nil {
					pairtAmount, _ := additionCheckPatternGroup.GetPairAmount()
					if pairtAmount >= 1 {
						break
					}
				}

				//招胡
				for _, fixedP := range player.fixedpatterns {
					if fixedP == nil || fixedP.ptype != PTKan || len(fixedP.cards) <= 0 {
						continue
					}

					if fixedP.cards[0].IsEqual(huCard) {
						return PTZhao
					}
				}

				break
			}
		}

		//检查二七十
		if StatisticsEQS(kaoCards, huCard) != nil {
			return PTEQSColumn
		}

		//检查AAB和碰
		if StatisticsAAB(kaoCards, huCard) != nil {
			return PTAABColumn
		}

		//检查顺子
		offset := kaoCard1.value - kaoCard2.value

		//检查三张牌是否都是大牌或小牌
		if !(kaoCard1.big == kaoCard2.big && kaoCard1.big == huCard.big) {
			return PTUknown
		}

		patternCards := append(kaoCards, huCard)
		switch offset {
		case 1:
			if kaoCard1.value+1 == huCard.value || kaoCard2.value-1 == huCard.value {
				if IsOneTwoThree(patternCards) {
					return PTOneTwoThree
				}
				return PTSZColumn
			}
		case -1:
			if kaoCard1.value-1 == huCard.value || kaoCard2.value+1 == huCard.value {
				if IsOneTwoThree(patternCards) {
					return PTOneTwoThree
				}
				return PTSZColumn
			}
		case 2, -2:
			if (kaoCard1.value+kaoCard2.value)/2 == huCard.value {
				if IsOneTwoThree(patternCards) {
					return PTOneTwoThree
				}
				return PTSZColumn
			}

		default:
			return PTUknown
		}

	}

	return PTUknown
}

//拢牌
func (controller *DaerController) Long() (isErLongTouYi bool) {
	player := controller.player
	if player == nil || player.cards == nil || len(player.cards) <= 0 {
		logger.Error("data of controller.player is error.")
		return
	}

	room := player.room
	if room == nil {
		logger.Error("room is nil.")
		return
	}

	//缓存拢之前的拢的数量
	longCount := len(player.showPatterns)

	smallAmount, bigAmount := StatisticsCardAmount(player.cards)
	//修改拢小牌
	for value, small := range smallAmount {
		if small == 4 {
			controller.long(int32(value+1), false)
		}
	}

	//修改拢大牌
	for value, big := range bigAmount {
		if big == 4 {
			controller.long(int32(value+1), true)
		}
	}

	//检查手牌上有没有开fixecpatterns列表里的（主要是由于庄家会单独进一张牌，如果手上有一个坎牌，然后在进这一张时，就需要检查fixecpatterns）
	controller.stripKanLong()

	//检测二拢偷一
	//有新的拢吗
	haveNewLong := longCount != len(player.showPatterns)
	if haveNewLong {

		isErLongTouYi = len(player.showPatterns) == ErLongTouYi
		if isErLongTouYi {
			openCard := player.room.OpenOneCard()
			if openCard == nil {
				logger.Error("初始阶段桌面上必须由牌")
				return
			}

			//logger.Info("二拢偷一时，房间的状态：", rootTypeName[room.state])
			player.SendActionACK(AMo, openCard, nil, ACSuccess)
			player.erLongTouYi = append(player.erLongTouYi, openCard)
			player.ObtainCard(openCard)
		}
	}

	return
}

//从坎牌手里再剔除拢(因为摸一张可能和手里的坎牌再形成拢)
func (controller *DaerController) stripKanLong() {
	player := controller.player
	if player == nil {
		logger.Error("player is nil.")
		return
	}

	room := player.room
	if room == nil {
		logger.Error("room is nil.")
		return
	}

	for i, c := range player.cards {
		isFind := false
		for fpi, fp := range player.fixedpatterns {
			if len(fp.cards) <= 0 {
				continue
			}

			if fp.cards[0].IsEqual(c) {
				c.flag = cmn.CPositive | cmn.CLock
				longPattern := NewPattern(PTLong, append(fp.cards, c))
				player.showPatterns = append(player.showPatterns, longPattern)
				player.fixedpatterns = append(player.fixedpatterns[:fpi], player.fixedpatterns[fpi+1:]...)
				player.cards = append(player.cards[:i], player.cards[i+1:]...)

				if room.state == RSBankerJinPaiStage {
					player.SendActionACK(ALong, nil, []*DaerPattern{longPattern}, ACSuccess)
				}

				isFind = true
				break
			}
		}
		if isFind {
			break
		}
	}
}

func (controller *DaerController) long(value int32, isBig bool) {
	//修改需要拢的牌的状态
	cards := FindCards(controller.player.cards, value, isBig)
	for i, card := range cards {
		card.owner = controller.player
		if i == 0 {
			card.flag = cmn.CPositive | cmn.CLock
		} else {
			card.flag = cmn.CBack | cmn.CLock
		}
	}

	//创建一个模式
	pattern := NewPattern(PTLong, cards)
	controller.player.showPatterns = append(controller.player.showPatterns, pattern)

	//从手牌中移除
	controller.player.cards = RemoveCardsByType(controller.player.cards, value, isBig)
}

//剔坎牌
func (controller *DaerController) StripKan() {
	if controller.player == nil || controller.player.cards == nil || len(controller.player.cards) <= 0 {
		logger.Error("data of controller.player is error.")
		return
	}

	smallAmount, bigAmount := StatisticsCardAmount(controller.player.cards)

	//修改小牌坎牌
	for value, small := range smallAmount {
		if small == 3 {
			controller.modifyKanCardStatus(int32(value+1), false)
		}
	}

	//修改大牌坎牌
	for value, big := range bigAmount {
		if big == 3 {
			controller.modifyKanCardStatus(int32(value+1), true)
		}
	}
}

func (controller *DaerController) modifyKanCardStatus(value int32, isBig bool) {
	//修改需要拢的牌的状态
	cards := FindCards(controller.player.cards, value, isBig)
	for _, card := range cards {
		card.owner = controller.player
		card.flag = cmn.CBack | cmn.CLock
	}

	//创建一个模式
	pattern := NewPattern(PTKan, cards)
	controller.player.fixedpatterns = append(controller.player.fixedpatterns, pattern)

	//从手牌中移除
	controller.player.cards = RemoveCardsByType(controller.player.cards, value, isBig)
}

//修改手牌状态
func (controller *DaerController) ModifyOtherCardStatusInHand() {
	//修改其他单的手牌
	for _, card := range controller.player.cards {
		isLock := card.flag&cmn.CLock > 0
		if !isLock {
			card.owner = controller.player
			card.flag = cmn.CBack
		}
	}
}

//检查招牌
func (controller *DaerController) CheckZhao(card *DaerCard) (result bool, isAgainZhao bool) {
	//检查参数
	if card == nil {
		logger.Error("DaerController.CheckZhao, card is nil")
		return
	}
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//检查是否有坎牌
	for _, pattern := range player.fixedpatterns {
		if pattern != nil && pattern.ptype == PTKan && len(pattern.cards) > 0 && card.IsEqual(pattern.cards[0]) {
			isAgainZhao = controller.GetAmountForZhaoAndLong() >= 1
			result = true
			break
		}
	}

	if result && !isAgainZhao && len(player.cards) < 2 {
		result = false
	}

	return
}

//进行招牌
func (controller *DaerController) Zhao(card *DaerCard) (result *DaerPattern, isAgainZhao bool) {
	//检查参数
	if card == nil {
		logger.Error("DaerController.Zhao card is nil")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	isAgainZhao = controller.GetAmountForZhaoAndLong() >= 1

	//进行招牌
	for i, pattern := range player.fixedpatterns {
		if pattern != nil && pattern.ptype == PTKan && len(pattern.cards) > 0 && card.IsEqual(pattern.cards[0]) {

			result = NewPattern(PTZhao, append(pattern.cards, card))
			//修改牌的状态
			for _, c := range result.cards {
				c.flag = cmn.CPositive | cmn.CLock
				c.owner = player
			}

			player.showPatterns = append(player.showPatterns, result)
			player.fixedpatterns = append(player.fixedpatterns[:i], player.fixedpatterns[i+1:]...)

			break
		}
	}

	return
}

//获取当前找和拢的数量
func (controller *DaerController) GetAmountForZhaoAndLong() (result int) {
	//检查参数
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return result
	}

	//获取招牌和拢牌的数量
	for _, pattern := range player.showPatterns {
		if pattern.ptype == PTLong || pattern.ptype == PTZhao {
			result++
		}
	}

	return result
}

//检查碰
func (controller *DaerController) CheckPeng(card *DaerCard) (result bool) {
	//检查参数
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//检查是否有对牌
	if card.big {
		for cardValue, cardAmount := range controller.huController.bigCardAmount {
			if cardValue+1 == int(card.value) {
				result = cardAmount >= 2
				break
			}
		}
	} else {
		for cardValue, cardAmount := range controller.huController.smallCardAmount {
			if cardValue+1 == int(card.value) {
				result = cardAmount >= 2
				break
			}
		}
	}

	if result && len(player.cards) < 4 {
		result = false
	}

	return
}

//碰牌
func (controller *DaerController) Peng(card *DaerCard) (result *DaerPattern) {
	//检查参数
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//进行碰牌
	pendCards := FindCards(player.cards, card.value, card.big)
	pendCards = append(pendCards, card)
	result = NewPattern(PTPeng, pendCards)
	for _, v := range result.cards {
		v.flag = cmn.CPositive | cmn.CLock
		v.owner = player
	}
	player.showPatterns = append(player.showPatterns, result)

	player.cards = RemoveCardsByType(player.cards, card.value, card.big)

	//手牌变了后需要从新更新hu控制器
	controller.huController.UpdateData(player.cards)

	return
}

//检查吃
func (controller *DaerController) CheckChi(card *DaerCard) (result []*DaerPattern, biResult map[uint][]*DaerPattern) {
	//检查参数
	if card == nil {
		logger.Error("card is nil.")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//统计能吃的模式
	//检查能二七十
	biResult = make(map[uint][]*DaerPattern, 0)
	tempResult := make([]*DaerPattern, 0)
	pattern := StatisticsEQS(player.cards, card)
	if pattern != nil {
		tempResult = append(tempResult, pattern)
	}

	//检查能吃AAB模式吗
	patterns := StatisticsAABs(player.cards, card)
	if patterns != nil {
		tempResult = append(tempResult, patterns...)
	}

	//检查能吃顺子吗
	patterns = controller.HaveSZB([]*DaerCard{}, card)
	for _, pattern := range patterns {
		if pattern != nil {
			tempResult = append(tempResult, pattern)
		}
	}

	//检查是否需要比
	needBi := FindCard(player.cards, card.value, card.big) != nil
	if !needBi {
		result = append(result, tempResult...)
	} else {

		for _, v := range tempResult {
			//搜集这个模式的靠牌
			kaoCards := GetKaoCards(v, card)
			if kaoCards == nil {
				continue
			}

			//获取所有比
			tempBiPatterns := make([]*DaerPattern, 0)

			//检查比不比的起
			switch v.ptype {
			case PTEQSColumn:
				//统计二七十比
				biPattern := controller.HaveEQSB(card)
				if biPattern != nil {
					tempBiPatterns = append(tempBiPatterns, biPattern)
				}

				//统计顺子的比
				biPatterns := controller.HaveSZB([]*DaerCard{}, card)
				if biPatterns != nil && len(biPatterns) > 0 {
					tempBiPatterns = append(tempBiPatterns, biPatterns...)
				}

				//统计AABB比
				biPattern = controller.HaveAABB(card)
				if biPattern != nil {
					tempBiPatterns = append(tempBiPatterns, biPattern)
				}
			case PTSZColumn:
				fallthrough
			case PTOneTwoThree:
				//统计二七十比
				biPattern := StatisticsEQS(player.cards, card)
				if biPattern != nil {
					tempBiPatterns = append(tempBiPatterns, biPattern)
				}

				//统计顺子的比
				biPatterns := controller.HaveSZB(kaoCards, card)
				if biPatterns != nil && len(biPatterns) > 0 {
					tempBiPatterns = append(tempBiPatterns, biPatterns...)
				}

				//统计AABB比
				biPattern = controller.HaveAABB(card)
				if biPattern != nil {
					tempBiPatterns = append(tempBiPatterns, biPattern)
				}
			case PTAABColumn:
				//如果在靠牌里存在吃牌，那么那就永远吃不起
				if IsExistCard(kaoCards, card) {
					result = append(result, v)
					break
				}

				//统计二七十比
				biPattern := StatisticsEQS(player.cards, card)
				if biPattern != nil {
					tempBiPatterns = append(tempBiPatterns, biPattern)
				}

				//统计顺子的比
				biPatterns := controller.HaveSZB([]*DaerCard{}, card)
				if biPatterns != nil && len(biPatterns) > 0 {
					tempBiPatterns = append(tempBiPatterns, biPatterns...)
				}
			default:
				logger.Error("除了二七十，顺子，AABB的其他模式不能检查比:", v)
			}

			//检查能比起吗
			if len(tempBiPatterns) > 0 {
				result = append(result, v)
				biResult[v.id] = tempBiPatterns
			}
		}

		//		biPattern := controller.HaveEQSB(card)
		//		if biPattern != nil {
		//			tempBiPatterns = append(tempBiPatterns, biPattern)
		//		}

		//		biPattern = controller.HaveAABB(card)
		//		if biPattern != nil {
		//			tempBiPatterns = append(tempBiPatterns, biPattern)
		//		}

		//		for _, v := range tempResult {
		//			//搜集这个模式的靠牌
		//			kaoCard := GetKaoCards(v, card)
		//			if kaoCard == nil {
		//				continue
		//			}

		//			//统计顺子的比
		//			biPatterns := controller.HaveSZB(kaoCard, card)

		//			//检查能比起吗
		//			if v.ptype == PTAABColumn {
		//				result = append(result, v)
		//			} else if len(tempBiPatterns) > 0 || len(biPatterns) > 0 {
		//				result = append(result, v)
		//				biResult[v.id] = append(tempBiPatterns, biPatterns...)
		//			}
		//		}
	}

	//特殊检查，吃了就没有牌出了
	curHandCount := len(player.cards)
	if curHandCount >= 7 {
		return
	}

	bRst := true
	for bRst {
		bRst = false
		for i, chiPattern := range result {
			if val, exist := biResult[chiPattern.id]; exist && val != nil && len(val) > 0 {
				if curHandCount < 7 {
					result = append(result[:i], result[i+1:]...)
					delete(biResult, chiPattern.id)

					bRst = true
					break
				}
			} else {
				if curHandCount < 4 {
					result = append(result[:i], result[i+1:]...)
					delete(biResult, chiPattern.id)

					bRst = true
					break
				}
			}
		}

		if !bRst {
			break
		}
	}

	return
}

//获取一个模式的拷牌
func GetKaoCards(pattern *DaerPattern, card *DaerCard) (kaoCard []*DaerCard) {
	kaoCard = make([]*DaerCard, 0)
	for i, c := range pattern.cards {
		if c.IsEqual(card) {
			kaoCard = append(kaoCard, pattern.cards[:i]...)
			kaoCard = append(kaoCard, pattern.cards[i+1:]...)
			break
		}
	}

	return
}

//获取一个模式的拷牌
func GetKaoCardsByRPC(pattern *rpc.Pattern, card *DaerCard) (kaoCard []*DaerCard) {
	kaoCard = make([]*DaerCard, 0)
	rpcCard := make([]*rpc.Card, 0)
	for i, c := range pattern.Cards {
		if int32(c.GetValue()) == card.value && c.GetBBig() == card.big {
			rpcCard = append(rpcCard, pattern.Cards[:i]...)
			rpcCard = append(rpcCard, pattern.Cards[i+1:]...)
			break
		}
	}

	//转换成服务器的数据结构（DaerCard）
	for _, c := range rpcCard {
		kaoCard = append(kaoCard, NewCard(0, int32(c.GetValue()), c.GetBBig()))
	}

	return
}

//进行吃牌和比牌
func (controller *DaerController) Chi(kaoCards []*DaerCard, card *DaerCard, biCards []*DaerCard) {
	if kaoCards == nil || len(kaoCards) <= 0 || card == nil {
		logger.Error("Chi:kaoCards or card is nil.")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil")
		return
	}

	//吃
	//添加吃牌标志
	card.flag |= cmn.CChi

	chiPatternType := controller.CalcPatternType(kaoCards, card, nil)
	if chiPatternType == PTUknown {
		logger.Error("吃的牌不能构成一个模式：靠牌和吃牌如下：")
		PrintCards(kaoCards)
		PrintCard(card)
		return
	}

	//生成组模式并放入玩家的showPatterns中
	chiPattern := NewPattern(chiPatternType, append(kaoCards, card))
	for _, v := range chiPattern.cards {
		v.flag |= cmn.CPositive | cmn.CLock
		v.owner = player
	}
	player.showPatterns = append(player.showPatterns, chiPattern)

	//移除卡牌从玩家手牌上
	for _, v := range kaoCards {
		player.cards = RemoveCardByType(player.cards, v.value, v.big)
	}

	//比
	if biCards != nil && len(biCards) > 0 {

		//生成组模式并放入玩家的showPatterns中
		biPatternType := controller.CalcPatternType(biCards[:len(biCards)-1], biCards[len(biCards)-1], nil)
		if biPatternType == PTUknown {
			logger.Error("比的牌不能构成一个模式：靠牌和吃牌如下：")
			PrintCards(biCards[:len(biCards)-1])
			PrintCard(biCards[len(biCards)-1])
			return
		}

		biPattern := NewPattern(biPatternType, biCards)
		for _, v := range biPattern.cards {
			v.flag |= cmn.CPositive | cmn.CLock
			v.owner = player
		}

		//修改吃比牌标记
		controller.ModifyBiFlag(biPattern, card)

		player.showPatterns = append(player.showPatterns, biPattern)

		//移除卡牌从玩家手牌上
		for _, v := range biCards {
			player.cards = RemoveCardByType(player.cards, v.value, v.big)
		}

	}

	//手牌变了后需要从新更新hu控制器
	controller.huController.UpdateData(player.cards)
}

//修改吃比牌标记
func (controller *DaerController) ModifyBiFlag(pattern *DaerPattern, card *DaerCard) {
	if pattern == nil || card == nil {
		logger.Error("pattern or card is nil")
		return
	}

	for _, v := range pattern.cards {
		if card.IsEqual(v) {
			v.flag |= cmn.CChi
			break
		}
	}

	return
}

//检查能否生成二七十的比
func (controller *DaerController) HaveEQSB(card *DaerCard) (result *DaerPattern) {
	if card == nil {
		return
	}

	tempCardAmount := controller.huController.smallCardAmount
	if card.big {
		tempCardAmount = controller.huController.bigCardAmount
	}

	switch card.value {
	case 2:
		if tempCardAmount[7-1] < 2 || tempCardAmount[10-1] < 2 {
			return
		}
	case 7:
		if tempCardAmount[2-1] < 2 || tempCardAmount[10-1] < 2 {
			return
		}
	case 10:
		if tempCardAmount[2-1] < 2 || tempCardAmount[7-1] < 2 {
			return
		}
	default:
		//logger.Error("次函数来检查二七十")
		return
	}

	return NewPattern(PTEQSColumn, []*DaerCard{NewCard(0, 2, card.big), NewCard(0, 7, card.big), NewCard(0, 10, card.big)})
}

//检查是否有顺子比
func (controller *DaerController) HaveSZB(kaoCard []*DaerCard, card *DaerCard) (result []*DaerPattern) {
	//检查输入参数
	if kaoCard == nil || card == nil {
		return
	}

	//剩余能够组比牌的牌的数量
	tempCardAmount := controller.huController.smallCardAmount
	if card.big {
		tempCardAmount = controller.huController.bigCardAmount
	}
	for _, v := range kaoCard {
		tempCardAmount[v.value-1]--
	}

	//检查并生成可以比的模式
	//向下检查
	curCheckValue := card.value - 1
	if curCheckValue > 0 && tempCardAmount[curCheckValue-1] >= 1 {
		curCheckValue = card.value - 2
		if curCheckValue > 0 && tempCardAmount[curCheckValue-1] >= 1 {
			if curCheckValue == 1 {
				result = append(result, NewPattern(PTOneTwoThree, []*DaerCard{NewCard(0, 1, card.big), NewCard(0, 2, card.big), NewCard(0, 3, card.big)}))
			} else {
				result = append(result, NewPattern(PTSZColumn, []*DaerCard{NewCard(0, card.value, card.big), NewCard(0, card.value-1, card.big), NewCard(0, card.value-2, card.big)}))
			}
		}
	}

	//中间检查
	curCheckValue = card.value - 1
	if curCheckValue > 0 && tempCardAmount[curCheckValue-1] >= 1 {
		curCheckValue = card.value + 1
		if curCheckValue <= 10 && tempCardAmount[curCheckValue-1] >= 1 {
			if card.value-1 == 1 {
				result = append(result, NewPattern(PTOneTwoThree, []*DaerCard{NewCard(0, 1, card.big), NewCard(0, 2, card.big), NewCard(0, 3, card.big)}))
			} else {
				result = append(result, NewPattern(PTSZColumn, []*DaerCard{NewCard(0, card.value-1, card.big), NewCard(0, card.value, card.big), NewCard(0, card.value+1, card.big)}))
			}
		}
	}

	//向上检查
	curCheckValue = card.value + 1
	if curCheckValue <= 10 && tempCardAmount[curCheckValue-1] >= 1 {
		curCheckValue = card.value + 2
		if curCheckValue <= 10 && tempCardAmount[curCheckValue-1] >= 1 {
			if card.value == 1 {
				result = append(result, NewPattern(PTOneTwoThree, []*DaerCard{NewCard(0, 1, card.big), NewCard(0, 2, card.big), NewCard(0, 3, card.big)}))
			} else {
				result = append(result, NewPattern(PTSZColumn, []*DaerCard{NewCard(0, card.value, card.big), NewCard(0, card.value+1, card.big), NewCard(0, card.value+2, card.big)}))
			}
		}
	}

	return
}

//检查是否有AAB比
func (controller *DaerController) HaveAABB(card *DaerCard) (result *DaerPattern) {
	//检查输入参数
	if card == nil {
		return
	}

	if card.big {
		if controller.huController.smallCardAmount[card.value-1] >= 2 {
			result = NewPattern(PTAABColumn, []*DaerCard{NewCard(0, card.value, true), NewCard(0, card.value, false), NewCard(0, card.value, false)})
		}
	} else {
		if controller.huController.bigCardAmount[card.value-1] >= 2 {
			result = NewPattern(PTAABColumn, []*DaerCard{NewCard(0, card.value, false), NewCard(0, card.value, true), NewCard(0, card.value, true)})
		}
	}
	return
}

//查叫
func (controller *DaerController) ChaJiao() {
	huPatternGroup, card := controller.player.GetMaxHuOfPatternGroup()
	if huPatternGroup == nil || card == nil {
		return
	}

	controller.player.multipleCount[MTChaJiao] = MinTangFanShu[MTChaJiao]
	controller.Hu(card)
}

//获取一个出牌
func (controller *DaerController) GetChuPai() *DaerCard {
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil.")
		return nil
	}

	if player.cards == nil || len(player.cards) <= 0 {
		return nil
	}

	for i := len(player.cards) - 1; i >= 0; i-- {
		if !player.cards[i].IsLock() {
			return player.cards[i]
		}
	}

	return nil
}

//出牌
func (controller *DaerController) ChuPai(card *DaerCard) (finalChuPai *DaerCard) {
	//检查参数的合法性
	if controller.player == nil {
		logger.Error("controller.player is nil.")
		return
	}

	//查找出能出的牌
	chuPais := FindCards(controller.player.cards, card.value, card.big)
	if chuPais == nil || len(chuPais) <= 0 {
		logger.Error("没有此牌：", card.value, card.big)
		return
	}

	for _, chuPai := range chuPais {
		if !chuPai.IsLock() {
			finalChuPai = chuPai
			break
		}
	}

	if finalChuPai == nil {
		logger.Error("没有此牌,或者此牌已经被锁定了！", finalChuPai.value, finalChuPai.big)
		return
	}

	//从手里移除此牌
	for i, c := range controller.player.cards {
		if c.IsEqual(finalChuPai) && !c.IsLock() {
			controller.player.cards = append(controller.player.cards[:i], controller.player.cards[i+1:]...)
			break
		}
	}

	controller.player.room.activeCard = finalChuPai

	controller.player.AddGuoCard(finalChuPai)

	//手牌变了后需要从新更新hu控制器
	controller.huController.UpdateData(controller.player.cards)

	return
}

//生成最终的胡牌模式组
func (controller *DaerController) GenerateFinalPatternGroup() (patternGroup *DaerPatternGroup) {
	//检查输入参数是否合法
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil.")
		return
	}

	//产生最终的胡的模式组
	patternsList := make([]*DaerPattern, 0)
	patternsList = append(patternsList, player.showPatterns...)
	patternsList = append(patternsList, player.fixedpatterns...)

	logger.Info("产生最终模式时，是否有摆牌", player.HaveBai())

	if player.HaveBai() {

		_, pg := player.GetMaxBaiOfPatternGroup()
		PrintPatternGroupS("在产生最终模式时，的最大摆牌的模式：", pg, false)

		remainPatternGroup := controller.GenerateTempBaiPatternGroup(pg)
		if remainPatternGroup != nil {
			patternsList = append(patternsList, remainPatternGroup.patterns...)
		}
	}

	patternGroup = NewPatternGroup(patternsList)

	return
}

//产生临时的摆拍模式组
func (controller *DaerController) GenerateTempBaiPatternGroup(pg *DaerPatternGroup) (patternGroup *DaerPatternGroup) {

	singleCards := make([]*DaerCard, 0)
	patternsList := make([]*DaerPattern, 0)

	if pg != nil {
		patternsList = append(patternsList, pg.patterns...)
		singleCardInpatternGroups := controller.huController.GetSingleCardInPatternGroup([]*DaerPatternGroup{pg})
		if singleCardInpatternGroups != nil && len(singleCardInpatternGroups) == 1 {
			singleCards = singleCardInpatternGroups[0]
		} else {
			logger.Error("没有获取到单牌")
		}
	} else {
		singleCards = controller.player.cards
	}

	patternsList = append(patternsList, controller.GenerateSinglePartterns(singleCards)...)

	patternGroup = NewPatternGroup(patternsList)

	return
}

//单牌组成3个一组的单牌模式
func (controller *DaerController) GenerateSinglePartterns(singleCards []*DaerCard) []*DaerPattern {
	result := make([]*DaerPattern, 0)
	if singleCards == nil || len(singleCards) <= 0 {
		return result
	}

	singleCardCount := len(singleCards)
	count := singleCardCount / 3
	logger.Info("Count:", count)
	if count == 0 {
		result = append(result, NewPattern(PTSingle, singleCards))
	} else if singleCardCount%3 == 0 {
		PrintCardsS("在产生最终模式时，单牌：", singleCards)
		for i := 0; i < count; i++ {
			result = append(result, NewPattern(PTSingle, singleCards[i*3:(i+1)*3]))
		}
	} else {
		for i := 0; i <= count; i++ {
			upValue := int(math.Min(float64((i+1)*3), float64(singleCardCount)))
			result = append(result, NewPattern(PTSingle, singleCards[i*3:upValue]))
		}
	}

	return result

}

//生成最终的胡牌模式组
func (controller *DaerController) GenerateTempPatternGroup(basePatternGroup *DaerPatternGroup, kaoCards []*DaerCard, huCard *DaerCard) (patternGroup *DaerPatternGroup) {
	//检查输入参数是否合法
	if basePatternGroup == nil {
		logger.Error("basePatternGroup is nil")
		return
	}
	if kaoCards == nil || len(kaoCards) <= 0 || huCard == nil {
		logger.Error("kaoCards or huCard is nil")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil.")
		return
	}

	//把最后的靠牌和胡牌组成一个模式放入模式组中
	ptype := controller.CalcPatternType(kaoCards, huCard, basePatternGroup)
	if ptype == PTUknown {
		logger.Error("最后一个模式无效")
		return
	}
	cardList := []*DaerCard{huCard}
	cardList = append(cardList, kaoCards...)
	lastPattern := NewPattern(ptype, cardList)

	//产生最终的胡的模式组
	patternsList := make([]*DaerPattern, 0)
	patternsList = append(patternsList, player.showPatterns...)
	patternsList = append(patternsList, player.fixedpatterns...)
	patternsList = append(patternsList, basePatternGroup.patterns...)
	patternsList = append(patternsList, lastPattern)

	patternGroup = NewPatternGroup(patternsList)

	return
}

//统计名堂
func (controller *DaerController) StatisticsRemainMinTangAndSave(patternGroup *DaerPatternGroup) {
	//检查输入参数是否合法
	if patternGroup == nil {
		logger.Error("patternGroup is nil")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil.")
		return
	}

	//统计名堂
	//黑摆是不统计胡子和名堂的
	if !player.HaveHeiBai() {
		//检查乱胡
		if patternGroup.Value() <= 0 {
			controller.player.multipleCount[MTLuanHu] = 1
		}

		//检查坤
		haveKun := IsKunMinTang(patternGroup)
		if haveKun {
			controller.player.multipleCount[MTKun] = MinTangFanShu[MTKun]
		}

		//检查红牌
		hongCardAmount := patternGroup.GetRedCardAmount()
		if !player.HaveBaiAndNotHeiBai() {
			if hongCardAmount >= 10 {
				controller.player.multipleCount[MTHongPai] = MinTangFanShu[MTHongPai]
			}
		}

		//检查黑牌
		if hongCardAmount <= 0 {
			controller.player.multipleCount[MTHeiPai] = MinTangFanShu[MTHeiPai]
		}
	}

	//检查归
	if controller.player.room.IsDaigui {
		controller.player.multipleCount[MTGui] = controller.GetGuiAmountByPatternGroup(patternGroup)
	}

	return
}

//统计剩余名堂（乱，红，坤，黑，归）
func (controller *DaerController) StatisticsRemainMinTang(patternGroup *DaerPatternGroup) (multipleCount map[int32]int32, fanCount int32) {
	multipleCount = make(map[int32]int32, 0)
	//检查输入参数是否合法
	if patternGroup == nil {
		logger.Error("patternGroup is nil")
		return
	}

	player := controller.player
	if player == nil {
		logger.Error("player is null.")
		return
	}

	logger.Info("统计剩余名堂是的模式组：")
	PrintPatternGroup(patternGroup, false)

	//统计名堂
	//黑摆是不统计胡子和名堂的
	if !player.HaveHeiBai() {
		//检查乱胡
		if patternGroup.Value() <= 0 {
			multipleCount[MTLuanHu] = 1
		}

		//检查坤
		haveKun := IsKunMinTang(patternGroup)
		if haveKun {
			multipleCount[MTKun] = MinTangFanShu[MTKun]
		}

		//检查红牌
		hongCardAmount := patternGroup.GetRedCardAmount()
		if !player.HaveBaiAndNotHeiBai() {
			if hongCardAmount >= 10 {
				multipleCount[MTHongPai] = MinTangFanShu[MTHongPai]
			}
		}

		//检查黑牌
		if hongCardAmount <= 0 {
			multipleCount[MTHeiPai] = MinTangFanShu[MTHeiPai]
		}
	}
	//检查归
	multipleCount[MTGui] = controller.GetGuiAmountByPatternGroup(patternGroup)

	//统计翻数
	fanCount = 0
	for _, fan := range multipleCount {
		fanCount += fan
	}

	return
}

//统计坤名堂
func IsKunMinTang(patternGroup *DaerPatternGroup) bool {

	haveZhaoOrLong := false
	for _, pattern := range patternGroup.patterns {
		if pattern.ptype == PTLong || pattern.ptype == PTZhao ||
			(pattern.cards != nil && len(pattern.cards) == 4) {
			haveZhaoOrLong = true
		}
	}

	haveKun := true
	for _, pattern := range patternGroup.patterns {
		if haveZhaoOrLong && pattern.ptype == PTPair {
			continue
		}

		if pattern.value() <= 0 {
			haveKun = false
			break
		}
	}

	return haveKun
}

// //获取归的数量
// func (controller *DaerController) GetGuiAmount() (result int32) {

// 	player := controller.player
// 	if player == nil {
// 		logger.Error("Controller.GetGuiAmount:player is nil.")
// 		return
// 	}

// 	cards := controller.GetAllCardsOfPlayer()
// 	if cards != nil {
// 		s, b := StatisticsCardAmount(cards)
// 		for _, amount := range s {
// 			if amount >= 4 {
// 				result++
// 			}
// 		}

// 		for _, amount := range b {
// 			if amount >= 4 {
// 				result++
// 			}
// 		}
// 	} else {
// 		logger.Error("玩家没有牌")
// 	}

// 	return
// }

// //获取玩家的所有卡牌
// func (controller *DaerController) GetAllCardsOfPlayer() (result []*DaerCard) {
// 	player := controller.player
// 	if player == nil {
// 		logger.Error("Controller.GetGuiAmount:player is nil.")
// 		return
// 	}

// 	result = make([]*DaerCard, 0)

// 	for _, p := range player.showPatterns {
// 		result = append(result, p.cards...)
// 	}

// 	for _, p := range player.fixedpatterns {
// 		result = append(result, p.cards...)
// 	}

// 	result = append(result, player.cards...)

// 	return
// }

//获取归的数量
func (controller *DaerController) GetGuiAmountByPatternGroup(patternGroup *DaerPatternGroup) (result int32) {
	if patternGroup == nil {
		logger.Error("模式组为空")
		return
	}

	cards := controller.GetAllCardsOfPatternGroup(patternGroup)

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
		logger.Error("GetGuiAmountByPatternGroup:玩家没有牌")
	}

	return
}

//获取模式组中的所有卡牌
func (controller *DaerController) GetAllCardsOfPatternGroup(patternGroup *DaerPatternGroup) (result []*DaerCard) {

	result = make([]*DaerCard, 0)

	if patternGroup == nil {
		logger.Error("模式组为空")
		return
	}

	for _, pattern := range patternGroup.patterns {
		result = append(result, pattern.cards...)
	}

	return
}

//统计胡数和分数
func (controller *DaerController) StatisticsHuAndScore(patternGroup *DaerPatternGroup, isLimitMul bool) (huAmount int32, huScore int32) {

	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil.")
		return
	}

	room := player.room
	if room == nil {
		logger.Error("player.room is nil.")
		return
	}

	//结算,检查是否是黑摆(黑摆是不算名堂的）
	if value, exist := player.multipleCount[MTHeiBai]; exist && value > 0 {
		huScore = SpecificHuScore[MTHeiBai]
		return
	} else if value, exist := player.multipleCount[MTLuanHu]; exist && value > 0 {
		huScore = SpecificHuScore[MTLuanHu]
	} else {
		//检查输入参数是否合法
		if patternGroup == nil {
			//logger.Error("patternGroup is nil.")
			return
		}

		huAmount = int32(patternGroup.Value())
		if huAmount > MaxHu {
			logger.Error("计算有问题，不能超过最大胡", MaxHu)
			huAmount = MaxHu
		}

		huScore = GetScoreByHu(huAmount)
	}

	//算翻
	logger.Info("算翻之前的分数：", huScore)
	var resultMul int32 = 0
	for mt, mul := range player.multipleCount {
		if mul > 0 && HaveFan(mt) {
			//huScore *= int32(math.Pow(2, float64(mul)))
			resultMul += mul
		}
	}

	huScore *= int32(math.Pow(2, float64(resultMul)))

	//logger.Error("限制倍数之前：", huScore)
	//进行封顶限制
	if isLimitMul {
		//点炮和查叫时分数上限翻一倍
		if player.HaveSpecificMingTang(MTDianPao) || player.IsChaJiao() {
			huScore = int32(math.Min(float64(huScore), float64(room.MaxMultiple*2)))
			//logger.Error("点炮和查叫算翻之后的分数：", huScore, room.MaxMultiple)
		} else {
			huScore = int32(math.Min(float64(huScore), float64(room.MaxMultiple)))
			//logger.Error("算翻之后的分数：", huScore, room.MaxMultiple)
		}
	}

	logger.Info("算翻之后的分数：", huScore)
	return
}

//是否是有翻得名堂
func HaveFan(mingtang int32) bool {
	return mingtang != MTSanLongBai && mingtang != MTSiKanBai && mingtang != MTHeiBai && mingtang != MTLuanHu
}
