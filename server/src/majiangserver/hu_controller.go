package majiangserver

import (
	//cmn "common"
	//"fmt"
	//"debug"
	"logger"
	"math"
	"sort"
	"time"
)

const (
	ESuccess = iota
	ECardNull
	ETypeAmountMuch
	ECardFullSame
)

//胡牌类型
const (
	DanDiaoHu = iota //单调胡
	ShunZiHu         //顺子胡
	DuiChuHu         //对处胡
)

//模式类型
const (
	NormalPattern    = iota //普通模式
	DaDuiZiPattern          //大对子
	XiaoQiDuiPattern        //小七对
)

type HuController struct {
	patternGroups     []*MaJiangPatternGroup
	originCards       []*MaJiangCard
	cards             []*MaJiangCard
	rmCardsAmountInfo *CardAmountStatistics //替换模式下的卡牌数量

	player *MaJiangPlayer
}

func NewHuController(p *MaJiangPlayer) *HuController {
	huC := &HuController{player: p}

	return huC
}

//初始化函数 ，调用完次函数后，就可以直接获取成员数据了
func (self *HuController) UpdateData(cards []*MaJiangCard) {
	//1.check input param
	if !self.CheckHandCardsAmount(cards) {
		//logger.Error("手牌数量有问题, 当前的手牌数量为：%d", len(cards))
		return
	}

	player := self.player
	if player == nil {
		logger.Error("HuController.player is nil.")
		return
	}

	//2.need check hu
	logger.Info("更新胡：")
	if eReason := self.needUpdate(cards); eReason != ESuccess {
		logger.Info("不需要更新:", eReason)
		//当以前是能够胡牌的，但是摸了一张不同花色的牌导致现在有不能胡了要把以前的胡的模式组给清除掉
		if eReason == ETypeAmountMuch {
			self.patternGroups = make([]*MaJiangPatternGroup, 0)
			self.originCards = make([]*MaJiangCard, len(cards))
			copy(self.originCards, cards)
		}

		PrintPatternGroupsS("在进行胡控制器更新时，坚持到不需要更新，此时以前可以胡的模式组如下：", self.patternGroups, false)

		return
	}

	logger.Info("开始计算胡数")

	//3.cache card
	self.originCards = make([]*MaJiangCard, len(cards))
	copy(self.originCards, cards)

	self.patternGroups = make([]*MaJiangPatternGroup, 0)

	//4.克隆一副手牌
	clonedCards := CloneMaJiangCards(cards)

	//5.拆分出本牌和红中
	normalCards, hongZhongCards := SplitCards(clonedCards)

	//6.根据手牌数量，生成素胡，打对子和小七对
	self.GeneratePatternGroup(normalCards, len(hongZhongCards))

	//7.打印最终的结果
	PrintPatternGroupsS("最终结果：", self.patternGroups, true)

}

//检查是否需要重新算胡牌，牌没变化就用算了
func (self *HuController) needUpdate(cards []*MaJiangCard) int32 {
	//1. check input param
	if cards == nil {
		logger.Error("needUpdate:cards is nil.")
		return ECardNull
	}

	//2. check type amount
	if self.player != nil {
		fixedTypeList := self.player.GetTypeInfoInShowPattern()
		amountInfo := NewCardAmountStatisticsByCards(cards, false)
		typeAmount := amountInfo.GetTypeAmount(false, fixedTypeList)
		if typeAmount > int32(2-len(fixedTypeList)) {
			return ETypeAmountMuch
		}
	}

	//3. check card amount is same
	if len(self.originCards) != 0 && len(self.originCards) != len(cards) {
		return ESuccess
	}

	//4. check is same
	tempCards := make([]*MaJiangCard, len(cards))
	copy(tempCards, cards)

	for _, v := range self.originCards {
		cType, cVal := v.CurValue()
		removedSuccess := true
		removedSuccess, tempCards = RemoveCardByType(tempCards, cType, cVal)
		if !removedSuccess {
			return ESuccess
		}
	}

	isSame := len(tempCards) <= 0
	if isSame {
		return ECardFullSame
	}

	return ESuccess
}

//检查手牌数量是否正确
func (self *HuController) CheckHandCardsAmount(cards []*MaJiangCard) bool {
	if cards == nil {
		return false
	}

	cardAmount := len(cards)
	return !(cardAmount != 13 && cardAmount != 10 && cardAmount != 7 && cardAmount != 4 && cardAmount != 1)
}

//产生模式组
func (self *HuController) GeneratePatternGroup(normalCards []*MaJiangCard, hzAmount int) {

	cardAmount := len(normalCards)

	if self.IsOnlyGenerateDaDuiZi(int32(cardAmount), int32(hzAmount)) {

		self.GenerateDaDuiZiPatternGroup(normalCards, hzAmount)
	} else if self.IsOnlyGenerateXiaoQiDui(int32(cardAmount), int32(hzAmount)) {

		self.GenerateXiaoQiDuiPatternGroup(normalCards, hzAmount)
	} else {
		//1.产生普通的模式组
		self.GenerateNormalPatternGroup(normalCards, hzAmount)

		//2.产生大对子的模式组
		self.GenerateDaDuiZiPatternGroup(normalCards, hzAmount)

		//3.产生小七对的模式组
		self.GenerateXiaoQiDuiPatternGroup(normalCards, hzAmount)
	}

	//4.打印所有胡的模式组
	//PrintPatternGroupsS("所有能胡的模式组:", self.patternGroups, true)
}

//产生素(普通)的模式组
func (self *HuController) GenerateNormalPatternGroup(normalCards []*MaJiangCard, hzAmount int) {

	logger.Info("开始计算普通胡")
	//根据胡牌方式进行枚举
	curTime := time.Now()
	//单调胡
	logger.Info("=====================普通胡牌下的单调胡=====================")
	self.GenerateNormalPatternGroupByHuType(DanDiaoHu, normalCards, hzAmount)

	//顺子胡
	logger.Info("=====================普通胡牌下的顺子胡=====================")
	self.GenerateNormalPatternGroupByHuType(ShunZiHu, normalCards, hzAmount)

	//对处胡
	logger.Info("=====================普通胡牌下的对处胡=====================")
	self.GenerateNormalPatternGroupByHuType(DuiChuHu, normalCards, hzAmount)

	logger.Info("产生普通模式组用时：", time.Now().Sub(curTime))
}

//产生素(普通)的单调胡模式组
func (self *HuController) GenerateNormalPatternGroupByHuType(huType int32, normalCards []*MaJiangCard, hzAmount int) {

	//1.检查输入参数
	if normalCards == nil || len(normalCards) <= 0 {
		logger.Error("本牌的数量是nil或0")
		return
	}

	//2.产生单调胡的模式列表,并计算最终的红中替换后的胡牌模式组
	totalCardAmount := len(normalCards) + hzAmount
	minCardAmount := 4
	if huType == DuiChuHu {
		minCardAmount = 7
	}
	if totalCardAmount < minCardAmount {
		return
	}

	maxKanAmount := (totalCardAmount - minCardAmount) / 3

	//计算每种组合
	PrintCardsS("========+++++++++普通胡牌，开始统计坎前，剩余的牌：", normalCards)
	AAAPatterns, remainAAANormalCards := GetAAAPatterns(normalCards, hzAmount)
	PrintPatternsS("========+++++++++普通胡牌，所有坎模式的列表：", AAAPatterns)
	PrintCardsS("========+++++++++普通胡牌，所有坎模式的列表时，剩余本牌：", remainAAANormalCards)

	for needKanAmount := maxKanAmount; needKanAmount >= 0; needKanAmount-- {
		//坎
		validAAAPatterns, validAAARNCCards, remainAAAHZAmounts := GetValidAAAPatterns(needKanAmount, AAAPatterns, remainAAANormalCards, hzAmount)

		if len(validAAAPatterns) != len(validAAARNCCards) || len(validAAAPatterns) != len(remainAAAHZAmounts) {
			logger.Error("三个的数量必须相同！")
			continue
		}

		//dump info
		logger.Info("普通胡牌，统计(%d)个坎时，所有组合：", needKanAmount)
		for kanCIndex, kanPatternC := range validAAAPatterns {
			PrintPatternsS("组合的坎：", kanPatternC)
			PrintCardsS("组合的坎后=====剩余的本牌：", validAAARNCCards[kanCIndex])
			logger.Info("剩余的红中数量:%d", remainAAAHZAmounts[kanCIndex])
		}

		for kanCIndex, kanPatternC := range validAAAPatterns {

			//顺子
			needSZAmount := maxKanAmount - needKanAmount + 1
			if huType == ShunZiHu {
				needSZAmount = maxKanAmount - needKanAmount
			}

			//计算顺子的时候，剩余的牌
			// remainKanCards := GetCardsOfPatternList(validAAARNCCards[kanCIndex])
			// remainKanNormalCards, remainKanHZCards := SplitCards(remainKanCards)
			// remainKanNormalCards = append(remainKanNormalCards, remainNormalCards...)

			remainKanNormalCards := validAAARNCCards[kanCIndex]
			remainKanHZAmount := remainAAAHZAmounts[kanCIndex]
			PrintCardsS("普通胡牌，开始顺子统计前，剩余的牌：", remainKanNormalCards)
			ABCPatterns := GetABCPatterns(remainKanNormalCards, remainKanHZAmount)
			if len(ABCPatterns) < needSZAmount {
				continue
			}

			PrintPatternsS("普通胡牌，统计的顺子：", ABCPatterns)

			validABCPatterns, validABCRNCCards, remainABCHZAmounts := GetValidABCPatterns(needSZAmount, ABCPatterns, remainKanNormalCards, remainKanHZAmount)
			if len(validABCPatterns) != len(validABCRNCCards) || len(validABCPatterns) != len(remainABCHZAmounts) {
				logger.Error("三个的数量必须相同！")
				continue
			}

			logger.Info("普通胡牌，统计(%d)个顺子时，所有组合：", needSZAmount)
			for szCIndex, szPatternC := range validABCPatterns {
				PrintPatternsS("组合的顺子：", szPatternC)
				PrintCardsS("组合的顺子后=====剩余的本牌：", validABCRNCCards[szCIndex])
				logger.Info("剩余的红中数量:%d", remainABCHZAmounts[szCIndex])
			}

			//生成红中替换后的最终模式组
			for szCIndex, szPatternC := range validABCPatterns {
				//计算生成最终模式前，剩余的牌
				// remainABCCards := GetCardsOfPatternList(otherSZPatternsC[szCIndex])
				// remainABCNormalCards, remainABCHZCards := SplitCards(remainABCCards)
				// remainABCNormalCards = append(remainABCNormalCards, ABCRemainCards...)
				// tempABCRemainHZAmount := ABCRemainHZAmount + len(remainABCHZCards)
				remainABCNormalCards := validABCRNCCards[szCIndex]
				remainABCHZAmount := remainABCHZAmounts[szCIndex]

				PrintCardsS("普通胡牌，顺子统计完后，剩余的牌：", remainABCNormalCards)
				if huType == ShunZiHu {
					needDuiZiAmount := 1

					if len(remainABCNormalCards)+remainABCHZAmount != 4 {
						logger.Error("到这一步时，必须是4张牌！")
						return
					}

					AAPatterns, remainAANormalCards := GetAAPatterns(remainABCNormalCards, remainABCHZAmount)
					if len(AAPatterns) < needDuiZiAmount {
						continue
					}

					validAAPatterns, validAARNCCards, remainAAHZAmounts := GetValidAAPatterns(needDuiZiAmount, AAPatterns, remainAANormalCards, remainABCHZAmount)
					if len(validAAPatterns) != len(validAARNCCards) || len(validAAPatterns) != len(remainAAHZAmounts) {
						logger.Error("三个的数量必须相同！")
						continue
					}

					for aaCIndex, aaPatternC := range validAAPatterns {
						//计算剩余的牌
						remainAANormalCards := validAARNCCards[aaCIndex]
						remainAAHZAmount := remainAAHZAmounts[aaCIndex]

						// singlRemainCards := []*MaJiangCard{}
						// singlRemainCards = append(singlRemainCards, singleCards...)

						for i := 0; i < remainAAHZAmount; i++ {
							remainAANormalCards = append(remainAANormalCards, NewHongZhong())
						}

						patterns := []*MaJiangPattern{}
						patterns = append(patterns, kanPatternC...)
						patterns = append(patterns, szPatternC...)
						patterns = append(patterns, aaPatternC[:needDuiZiAmount]...)
						self.GenerateFinalPatternGroups(NormalPattern, patterns, remainAANormalCards, hzAmount)
					}

				} else {
					for i := 0; i < remainABCHZAmount; i++ {
						remainABCNormalCards = append(remainABCNormalCards, NewHongZhong())
					}

					patterns := []*MaJiangPattern{}
					patterns = append(patterns, kanPatternC...)
					patterns = append(patterns, szPatternC...)
					self.GenerateFinalPatternGroups(NormalPattern, patterns, remainABCNormalCards, hzAmount)
				}
			}
		}

	}

	PrintCardsS("GenerateNormalPatternGroupByHuType函数退出后，normalCards的情况：", normalCards)
}

//产生大对子模式组
func (self *HuController) GenerateDaDuiZiPatternGroup(normalCards []*MaJiangCard, hzAmount int) {
	//根据胡牌方式进行枚举
	curTime := time.Now()
	//单调胡
	logger.Info("=====================大对子胡牌下的单调胡=====================")
	self.GenerateDaDuiZiPatternGroupByHuType(DanDiaoHu, normalCards, hzAmount)
	//对处胡
	logger.Info("=====================大对子胡牌下的对处胡=====================")
	self.GenerateDaDuiZiPatternGroupByHuType(DuiChuHu, normalCards, hzAmount)

	logger.Info("产生大对子模式组用时：", time.Now().Sub(curTime))
}

func (self *HuController) GenerateDaDuiZiPatternGroupByHuType(huType int32, normalCards []*MaJiangCard, hzAmount int) {

	//1.检查输入参数
	if normalCards == nil || len(normalCards) <= 0 {
		logger.Error("本牌的数量是nil或0")
		return
	}

	//2.产生单调胡的模式列表,并计算最终的红中替换后的胡牌模式组
	totalCardAmount := len(normalCards) + hzAmount
	minCardAmount := 1
	if huType == DuiChuHu {
		minCardAmount = 4
	}
	if totalCardAmount < minCardAmount {
		return
	}

	needKanAmount := (totalCardAmount - minCardAmount) / 3

	//坎
	AAAPatterns, remainAAANormalCards := GetAAAPatterns(normalCards, hzAmount)
	if len(AAAPatterns) < needKanAmount {
		return
	}

	PrintPatternsS("大对子胡牌，统计的坎：", AAAPatterns)
	PrintCardsS("大对子胡牌，统计坎时，剩余本牌：", remainAAANormalCards)

	//坎
	validAAAPatterns, validAAARNCCards, remainAAAHZAmounts := GetValidAAAPatterns(needKanAmount, AAAPatterns, remainAAANormalCards, hzAmount)

	if len(validAAAPatterns) != len(validAAARNCCards) || len(validAAAPatterns) != len(remainAAAHZAmounts) {
		logger.Error("三个的数量必须相同！")
		return
	}

	//dump info
	logger.Info("大对子胡牌，统计(%d)个坎时，所有组合：", needKanAmount)
	for kanCIndex, kanPatternC := range validAAAPatterns {
		PrintPatternsS("组合的坎：", kanPatternC)
		PrintCardsS("组合的坎后=====剩余的本牌：", validAAARNCCards[kanCIndex])
		logger.Info("剩余的红中数量:%d", remainAAAHZAmounts[kanCIndex])
	}

	for kanCIndex, kanPatternC := range validAAAPatterns {
		// remainKanCards := GetCardsOfPatternList(otherKanPatternsC[kanCIndex])
		// remainKanNormalCards, remainKanHZCards := SplitCards(remainKanCards)
		// remainKanNormalCards = append(remainKanNormalCards, remainNormalCards...)
		//tempRemainHZAmount := remainHZAmount + len(remainKanHZCards)

		remainKanNormalCards := validAAARNCCards[kanCIndex]
		remainKanHZAmount := remainAAAHZAmounts[kanCIndex]

		for i := 0; i < remainKanHZAmount; i++ {
			remainKanNormalCards = append(remainKanNormalCards, NewHongZhong())
		}

		PrintCardsS("大对子胡牌，坎统计完后，剩余的牌：", remainKanNormalCards)
		patterns := []*MaJiangPattern{}
		patterns = append(patterns, kanPatternC...)
		self.GenerateFinalPatternGroups(DaDuiZiPattern, patterns, remainKanNormalCards, hzAmount)
	}
}

//产生小七对模式组
func (self *HuController) GenerateXiaoQiDuiPatternGroup(normalCards []*MaJiangCard, hzAmount int) {

	//1.检查输入参数
	if normalCards == nil || len(normalCards) <= 0 {
		logger.Error("本牌的数量是nil或0")
		return
	}

	curTime := time.Now()

	totalCardAmount := len(normalCards) + hzAmount
	if totalCardAmount != 13 {
		return
	}
	logger.Info("=====================小七对胡牌下的单调胡=====================")
	PrintCardsS("小七对胡牌，开始统计坎时，剩余的牌：", normalCards)

	needDuiZiAmount := (totalCardAmount - 1) / 2

	normalAAPatterns, normalSingleCards := SplitToAA_A(normalCards)

	needDuiZiAmount -= len(normalAAPatterns)

	AAPatterns, remainSingleCards := GetAAPatterns(normalSingleCards, hzAmount)
	if len(AAPatterns) < needDuiZiAmount {
		return
	}

	validAAPatterns, validAARNCCards, remainAAHZAmounts := GetValidAAPatterns(needDuiZiAmount, AAPatterns, remainSingleCards, hzAmount)

	if len(validAAPatterns) != len(validAARNCCards) || len(validAAPatterns) != len(remainAAHZAmounts) {
		logger.Error("三个的数量必须相同！")
		return
	}

	//dump info
	logger.Info("大对子胡牌，统计(%d)个坎时，所有组合：", needDuiZiAmount)
	for aaCIndex, aaPatternC := range validAAPatterns {
		PrintPatternsS("组合的坎：", aaPatternC)
		PrintCardsS("组合的坎后=====剩余的本牌：", validAARNCCards[aaCIndex])
		logger.Info("剩余的红中数量:%d", remainAAHZAmounts[aaCIndex])
	}

	for aaCIndex, aaPatternC := range validAAPatterns {
		// remainKanCards := GetCardsOfPatternList(otherKanPatternsC[kanCIndex])
		// remainKanNormalCards, remainKanHZCards := SplitCards(remainKanCards)
		// remainKanNormalCards = append(remainKanNormalCards, remainNormalCards...)
		//tempRemainHZAmount := remainHZAmount + len(remainKanHZCards)

		remainAANormalCards := validAARNCCards[aaCIndex]
		remainAAHZAmount := remainAAHZAmounts[aaCIndex]

		logger.Info("小七对胡牌，统计坎后，红中的数量：", remainAAHZAmount)
		PrintPatternsS("小七对胡牌，统计的对：", aaPatternC)
		PrintCardsS("小七对胡牌，统计对时，剩余本牌：", remainAANormalCards)

		//计算剩余的牌
		// singlRemainCards := GetCardsOfPatternList(AAPatterns[needDuiZiAmount:])
		// singlRemainCards = append(singlRemainCards, remainSingleCards...)

		for i := 0; i < remainAAHZAmount; i++ {
			remainAANormalCards = append(remainAANormalCards, NewHongZhong())
		}

		PrintCardsS("小七对胡牌，对子统计完后，剩余的牌：", remainAANormalCards)
		patterns := []*MaJiangPattern{}
		patterns = append(patterns, normalAAPatterns...)
		patterns = append(patterns, AAPatterns...)
		self.GenerateFinalPatternGroups(XiaoQiDuiPattern, patterns, remainAANormalCards, hzAmount)
	}

	logger.Info("产生小七对模式组,用时：", time.Now().Sub(curTime))

}

//获取模式列表里的牌
func GetCardsOfPatternList(patterns []*MaJiangPattern) []*MaJiangCard {
	result := make([]*MaJiangCard, 0)

	if len(patterns) <= 0 {
		return result
	}

	for _, p := range patterns {
		result = append(result, p.cards...)
	}

	return result
}

//获取AAA的所有模式(包括红中的替代的模式)
func GetAAAPatterns(normalCards []*MaJiangCard, hzAmount int) (patterns []*MaJiangPattern, remainNormalCards []*MaJiangCard) {
	//初始话返回参数
	patterns = []*MaJiangPattern{}
	remainNormalCards = []*MaJiangCard{}
	if len(normalCards) <= 0 {
		return
	}

	//拆分本牌
	AAAPatterns, AAPatterns, AAASingleCards := SplitToAAA_AA_A(normalCards)

	patterns = append(patterns, AAAPatterns...)

	//根据红中的数量来组合AAA的模式
	if hzAmount <= 0 {
		remainNormalCards = append(remainNormalCards, GetCardsOfPatternList(AAPatterns)...)
		remainNormalCards = append(remainNormalCards, AAASingleCards...)
		return
	}

	if hzAmount >= 1 {
		for _, aaP := range AAPatterns {
			temp := make([]*MaJiangCard, 0, 3)
			temp = append(temp, aaP.cards...)
			temp = append(temp, NewCard(0, HongZhong, 0))
			patterns = append(patterns, NewPattern(PTKan, temp, false))
		}
	}

	if hzAmount >= 2 {
		for _, c := range AAASingleCards {
			temp := make([]*MaJiangCard, 0, 3)
			temp = append(temp, c)
			temp = append(temp, NewCard(0, HongZhong, 0))
			temp = append(temp, NewCard(0, HongZhong, 0))
			patterns = append(patterns, NewPattern(PTKan, temp, false))
		}
	} else {
		remainNormalCards = append(remainNormalCards, AAASingleCards...)
	}

	return
}

//获取有效的AAA
func GetValidAAAPatterns(needKanAmount int, patterns []*MaJiangPattern, normalCards []*MaJiangCard, hzAmount int) (
	result [][]*MaJiangPattern, remainNormalCards [][]*MaJiangCard, remainHZAmounts []int) {

	//初始化返回值
	result = [][]*MaJiangPattern{}
	remainNormalCards = [][]*MaJiangCard{}
	remainHZAmounts = []int{}

	//检查参数的合法性
	if needKanAmount > len(patterns) {
		return
	}

	kanPCs, otherKanPCs := GetAllPatternsCombination(needKanAmount, patterns, true)
	if len(kanPCs) != len(otherKanPCs) {
		logger.Error("两个的数量必须相同！")
		return
	}

	for kanPCIndex, kanPC := range kanPCs {
		sucess, remainHZAmount := VerifyPatterns(kanPC, hzAmount)
		if !sucess {
			continue
		}

		otherKanPCCards := GetCardsOfPatternList(otherKanPCs[kanPCIndex])
		nCards, _ := SplitCards(otherKanPCCards)

		result = append(result, kanPC)
		tempCards := []*MaJiangCard{}
		tempCards = append(tempCards, nCards...)
		if normalCards != nil {
			tempCards = append(tempCards, normalCards...)
		}
		remainNormalCards = append(remainNormalCards, tempCards)
		remainHZAmounts = append(remainHZAmounts, remainHZAmount)
	}
	return
}

//是否是有效的组
func VerifyPatterns(patterns []*MaJiangPattern, hzAmount int) (isValid bool, remainHZAmount int) {
	isValid = true
	remainHZAmount = hzAmount
	if len(patterns) <= 0 {
		return
	}

	for _, p := range patterns {
		for _, c := range p.cards {
			if c.IsHongZhong() {
				remainHZAmount--
				if remainHZAmount < 0 {
					isValid = false
					return
				}
			}
		}
	}

	return

}

//获取AA的所有模式(包括红中的替代的模式)
func GetAAPatterns(normalCards []*MaJiangCard, hzAmount int) (patterns []*MaJiangPattern, remainNormalCards []*MaJiangCard) {
	//初始话返回参数
	patterns = []*MaJiangPattern{}
	remainNormalCards = []*MaJiangCard{}
	if len(normalCards) <= 0 {
		return
	}

	//拆分本牌
	AAPatterns, ABCSingleCards := SplitToAA_A(normalCards)

	patterns = append(patterns, AAPatterns...)

	//根据红中的数量来组合AAA的模式
	if hzAmount <= 0 {
		remainNormalCards = append(remainNormalCards, ABCSingleCards...)
		return
	}

	if hzAmount >= 1 {
		for _, c := range ABCSingleCards {
			temp := []*MaJiangCard{c, NewHongZhong()}
			patterns = append(patterns, NewPattern(PTPair, temp, false))
		}
	}

	return
}

//获取有效的AA
func GetValidAAPatterns(needPairAmount int, patterns []*MaJiangPattern, normalCards []*MaJiangCard, hzAmount int) (
	result [][]*MaJiangPattern, remainNormalCards [][]*MaJiangCard, remainHZAmounts []int) {

	//初始化返回值
	result = [][]*MaJiangPattern{}
	remainNormalCards = [][]*MaJiangCard{}
	remainHZAmounts = []int{}

	//检查参数的合法性
	if needPairAmount > len(patterns) {
		return
	}

	pairPCs, otherPairPCs := GetAllPatternsCombination(needPairAmount, patterns, true)
	if len(pairPCs) != len(otherPairPCs) {
		logger.Error("两个的数量必须相同！")
		return
	}

	for pairPCIndex, pairPC := range pairPCs {
		sucess, remainHZAmount := VerifyPatterns(pairPC, hzAmount)
		if !sucess {
			continue
		}

		otherPairPCCards := GetCardsOfPatternList(otherPairPCs[pairPCIndex])
		nCards, _ := SplitCards(otherPairPCCards)

		result = append(result, pairPC)
		tempCards := []*MaJiangCard{}
		tempCards = append(tempCards, nCards...)
		if normalCards != nil {
			tempCards = append(tempCards, normalCards...)
		}

		remainNormalCards = append(remainNormalCards, tempCards)
		remainHZAmounts = append(remainHZAmounts, remainHZAmount)
	}

	return
}

//获取AAA的所有模式(包括红中的替代的模式)
func GetABCPatterns(normalCards []*MaJiangCard, hzAmount int) (patterns []*MaJiangPattern) {
	//1.初始话返回参数
	patterns = []*MaJiangPattern{}
	if len(normalCards) <= 0 {
		return
	}

	//2.组合可能的顺子
	//2.1组合3个本牌的顺子
	tempPatterns := []*MaJiangPattern{}
	for _, c := range normalCards {
		findCard1 := FindCard(normalCards, c.cType, c.value-1)
		if findCard1 != nil {
			findCard2 := FindCard(normalCards, c.cType, c.value-2)
			if findCard2 != nil {
				cards := []*MaJiangCard{c, findCard1, findCard2}
				tempPatterns = append(tempPatterns, NewPattern(PTSZ, cards, false))
			}
		}
	}

	//2.2组合2个本牌的顺子
	//2.2.1组合2个本牌连着的顺子
	if hzAmount >= 1 {
		for _, c := range normalCards {
			findCard1 := FindCard(normalCards, c.cType, c.value-1)
			if findCard1 != nil {
				cards := []*MaJiangCard{c, findCard1, NewHongZhong()}
				tempPatterns = append(tempPatterns, NewPattern(PTSZ, cards, false))
			}
		}
	}

	//2.2.2组合2个本牌间隔的顺子
	if hzAmount >= 1 {
		for _, c := range normalCards {
			findCard1 := FindCard(normalCards, c.cType, c.value-2)
			if findCard1 != nil {
				cards := []*MaJiangCard{c, findCard1, NewHongZhong()}
				tempPatterns = append(tempPatterns, NewPattern(PTSZ, cards, false))
			}
		}
	}
	//2.3组合1个本牌的顺子
	if hzAmount >= 2 {
		for _, c := range normalCards {
			cards := []*MaJiangCard{c, NewHongZhong(), NewHongZhong()}
			tempPatterns = append(tempPatterns, NewPattern(PTSZ, cards, false))
		}
	}
	//2.4组合全红中的顺子
	hzSZAmount := hzAmount / 3
	for i := 0; i < hzSZAmount; i++ {
		cards := []*MaJiangCard{NewHongZhong(), NewHongZhong(), NewHongZhong()}
		tempPatterns = append(tempPatterns, NewPattern(PTSZ, cards, false))
	}

	//3.踢出多余的顺子
	//3.1拆分相同的模式
	//PrintPatternsS("统计后的所有模式:", tempPatterns)
	tempMap := SplitPatternsToSameList(tempPatterns)

	//3.2计算相同的顺子需要保留几个
	for _, ps := range tempMap {

		//PrintPatternsS("统计后的每一组模式:", ps)
		if len(ps) <= 0 {
			continue
		}

		if ps[0].IsAllHZ() {
			patterns = append(patterns, ps...)
		} else {
			curPsAmount := len(ps)
			maxPsAmount := CalcMaxPatternsAmount(normalCards, ps[0])
			reservePsAmount := int32(math.Min(float64(curPsAmount), float64(maxPsAmount)))
			//移除多余的
			patterns = append(patterns, ps[:reservePsAmount]...)
		}
	}

	//PrintPatternsS("统计后的最终模式组:", patterns)
	return
}

//拆分相同的模式组
func SplitPatternsToSameList(patterns []*MaJiangPattern) (result [][]*MaJiangPattern) {
	result = [][]*MaJiangPattern{}

	if len(patterns) <= 0 {
		return
	}

	for _, p := range patterns {
		find := false
		for i, rp := range result {
			if len(rp) <= 0 {
				continue
			}

			if rp[0].IsEqual(p) {
				result[i] = append(result[i], p)
				find = true
				break
			}
		}

		if !find {
			result = append(result, []*MaJiangPattern{p})
		}
	}

	return
}

//获取能够通过本牌组成的最大数量的模式
func CalcMaxPatternsAmount(normalCards []*MaJiangCard, pattern *MaJiangPattern) (result int) {

	if len(normalCards) <= 0 || pattern == nil {
		return 0
	}

	result = 999999
	for _, c := range pattern.cards {
		if c.IsHongZhong() {
			continue
		}
		findCards := FindCards(normalCards, c.cType, c.value)
		findCardsAmount := len(findCards)
		if findCardsAmount < result {
			result = findCardsAmount
		}
	}

	return
}

//获取有效的ABC
func GetValidABCPatterns(needSZAmount int, patterns []*MaJiangPattern, normalCards []*MaJiangCard, hzAmount int) (
	result [][]*MaJiangPattern, remainNormalCards [][]*MaJiangCard, remainHZAmounts []int) {

	//初始化返回值
	result = [][]*MaJiangPattern{}
	remainNormalCards = [][]*MaJiangCard{}
	remainHZAmounts = []int{}

	//检查参数的合法性
	if needSZAmount > len(patterns) {
		return
	}

	szPatterns := GetSZPatternsCombination(needSZAmount, patterns)

	for _, szPtns := range szPatterns {
		sucess, vfyRemainNormalCards, remainHZAmount := GetRemainNormalCardsAndHZAmountAndVerify(szPtns, normalCards, hzAmount)
		if !sucess {
			continue
		}

		result = append(result, szPtns)
		remainNormalCards = append(remainNormalCards, vfyRemainNormalCards)
		remainHZAmounts = append(remainHZAmounts, remainHZAmount)
	}
	return
}

//获取顺子组合的模式组后剩余的本牌和红中数量
func GetRemainNormalCardsAndHZAmountAndVerify(patterns []*MaJiangPattern, normalCards []*MaJiangCard, hzAmount int) (
	success bool, remainNormalCards []*MaJiangCard, remainHZAmount int) {

	//1.设置返回的初始化值
	remainNormalCards = make([]*MaJiangCard, len(normalCards))
	copy(remainNormalCards, normalCards)
	remainHZAmount = hzAmount
	success = true

	if len(patterns) <= 0 {
		return
	}

	//2.获取剩余的本牌和红中数量
	for _, p := range patterns {
		for _, c := range p.cards {
			if c.IsHongZhong() {
				if remainHZAmount <= 0 {
					success = false
					break
				}
				remainHZAmount--
			} else {
				if len(remainNormalCards) <= 0 {
					success = false
					break
				}
				removedSuccess := true
				removedSuccess, remainNormalCards = RemoveCardByType(remainNormalCards, c.cType, c.value)
				if !removedSuccess {
					success = false
					break
				}
			}
		}
		if !success {
			break
		}
	}

	return
}

//将牌拆分成AAA, AA或A模式
func SplitToAAA_AA_A(cards []*MaJiangCard) (AAAPatterns, AAPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {

	AAAPatterns = []*MaJiangPattern{}
	AAPatterns = []*MaJiangPattern{}
	singleCards = []*MaJiangCard{}

	isFirst := true
	var firstCard *MaJiangCard
	var ptnCards []*MaJiangCard

	tempCards := make([]*MaJiangCard, len(cards))
	copy(tempCards, cards)

	for len(tempCards) > 0 {
		if isFirst {
			firstCard = tempCards[0]
			ptnCards = []*MaJiangCard{}
			ptnCards = append(ptnCards, firstCard)
			tempCards = tempCards[1:]
			isFirst = false
		}

		isEndInnerLoop := len(tempCards) <= 0
		for i, c := range tempCards {
			isEndInnerLoop = i+1 >= len(tempCards)
			if firstCard.IsFullEqual(c) {
				ptnCards = append(ptnCards, c)
				tempCards = append(tempCards[:i], tempCards[i+1:]...)
				break
			}
		}

		ptnAmount := len(ptnCards)
		if isEndInnerLoop || ptnAmount >= 3 {
			isFirst = true
			switch ptnAmount {
			case 3:
				AAAPatterns = append(AAAPatterns, NewPattern(PTKan, ptnCards, false))
			case 2:
				AAPatterns = append(AAPatterns, NewPattern(PTPair, ptnCards, false))
			case 1:
				singleCards = append(singleCards, ptnCards...)
			default:
				logger.Error("不能超过3个！")
			}
		}
	}

	return
}

////将牌拆分成AA或A模式
func SplitToAA_A(cards []*MaJiangCard) (AAPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {
	AAPatterns = []*MaJiangPattern{}
	singleCards = []*MaJiangCard{}

	isFirst := true
	var firstCard *MaJiangCard
	var ptnCards []*MaJiangCard

	tempCards := make([]*MaJiangCard, len(cards))
	copy(tempCards, cards)

	for len(tempCards) > 0 {
		if isFirst {
			firstCard = tempCards[0]
			ptnCards = []*MaJiangCard{}
			ptnCards = append(ptnCards, firstCard)
			tempCards = tempCards[1:]
			isFirst = false
		}

		isEndInnerLoop := len(tempCards) <= 0
		for i, c := range tempCards {
			isEndInnerLoop = i+1 >= len(tempCards)
			if firstCard.IsFullEqual(c) {
				ptnCards = append(ptnCards, c)
				tempCards = append(tempCards[:i], tempCards[i+1:]...)
				break
			}
		}

		ptnAmount := len(ptnCards)
		if isEndInnerLoop || ptnAmount >= 2 {
			isFirst = true

			switch ptnAmount {
			case 2:
				AAPatterns = append(AAPatterns, NewPattern(PTPair, ptnCards, false))
			case 1:
				singleCards = append(singleCards, ptnCards...)
			default:
				logger.Error("不能超过2个！")
			}
		}
	}

	return
}

//组合指定数量的顺子
func SplitToABC_AB_A(cards []*MaJiangCard) (ABCPatterns, ABPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {
	//1.检查输入参数
	if cards == nil || len(cards) <= 0 {
		logger.Error("cards is null.")
		return
	}

	tempCards := make([]*MaJiangCard, len(cards))
	copy(tempCards, cards)

	//2.排序
	sort.Sort(CardList(tempCards))

	//3.统计3张牌的顺子
	ABCPatterns = make([]*MaJiangPattern, 0)
	tagIndex := 0
	for tagIndex <= len(tempCards)-3 {
		tagCard := tempCards[tagIndex]
		ptnCards := []*MaJiangCard{tagCard}

		if tempCards[tagIndex+1].IsFullEqualByTypeAndValue(tagCard.cType, tagCard.value+1, tagCard.rcType) {
			ptnCards = append(ptnCards, tempCards[tagIndex+1])
		}

		if tempCards[tagIndex+2].IsFullEqualByTypeAndValue(tagCard.cType, tagCard.value+2, tagCard.rcType) {
			ptnCards = append(ptnCards, tempCards[tagIndex+2])
		}

		if len(ptnCards) >= 3 {
			ABCPatterns = append(ABCPatterns, NewPattern(PTSZ, ptnCards, false))
			tempCards = append(tempCards[:tagIndex], tempCards[tagIndex+3:]...)
		} else {
			tagIndex++
		}
	}

	//4.统计2张牌的间隔模式
	ABPatterns = make([]*MaJiangPattern, 0)
	tagIndex = 0
	for tagIndex <= len(tempCards)-2 {
		tagCard := tempCards[tagIndex]
		ptnCards := []*MaJiangCard{tagCard}

		if tempCards[tagIndex+1].IsFullEqualByTypeAndValue(tagCard.cType, tagCard.value+2, tagCard.rcType) {
			ptnCards = append(ptnCards, tempCards[tagIndex+1])
			ABPatterns = append(ABPatterns, NewPattern(PTSZ, ptnCards, false))
			tempCards = append(tempCards[:tagIndex], tempCards[tagIndex+2:]...)
		} else {
			tagIndex++
		}

	}

	//5.统计2张牌的相邻模式
	tagIndex = 0
	for tagIndex <= len(tempCards)-2 {
		tagCard := tempCards[tagIndex]
		ptnCards := []*MaJiangCard{tagCard}

		if tempCards[tagIndex+1].IsFullEqualByTypeAndValue(tagCard.cType, tagCard.value+1, tagCard.rcType) {
			ptnCards = append(ptnCards, tempCards[tagIndex+1])
			ABPatterns = append(ABPatterns, NewPattern(PTSZ, ptnCards, false))
			tempCards = append(tempCards[:tagIndex], tempCards[tagIndex+2:]...)
		} else {
			tagIndex++
		}

	}

	//6.统计剩余的单牌
	singleCards = make([]*MaJiangCard, 0)
	singleCards = append(singleCards, tempCards...)

	return
}

//牌排序
type CardList []*MaJiangCard

func (self CardList) Len() int {
	return len(self)
}
func (self CardList) Less(i, j int) bool {
	iCurType, iCurValue := self[i].CurValue()
	jCurType, jCurValue := self[j].CurValue()

	if iCurType == jCurType {
		return iCurValue < jCurValue
	}

	return iCurType < jCurType
}
func (self CardList) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

//计算胡和红中替换
func (self *HuController) GenerateFinalPatternGroups(pType int32, patterns []*MaJiangPattern, remainCards []*MaJiangCard, hzAmount int) (result []*MaJiangPatternGroup) {
	result = make([]*MaJiangPatternGroup, 0)

	//debug info
	switch pType {
	case NormalPattern:
		PrintPatternsS("GenerateFinalPatternGroups=====普通胡牌时生成的模式：", patterns)
		PrintCardsS("GenerateFinalPatternGroups=====剩余的单牌是：", remainCards)
	case DaDuiZiPattern:
		PrintPatternsS("GenerateFinalPatternGroups=====大对子胡牌时生成的模式：", patterns)
		PrintCardsS("GenerateFinalPatternGroups=====剩余的单牌是：", remainCards)
	case XiaoQiDuiPattern:
		PrintPatternsS("GenerateFinalPatternGroups=====小七对胡牌时生成的模式：", patterns)
		PrintCardsS("GenerateFinalPatternGroups=====剩余的单牌是：", remainCards)
	}

	logger.Info("GenerateFinalPatternGroups: enter!")
	//1.检查剩余的牌有没得机会可以胡
	if !CanHu(remainCards) {
		PrintCardsS("不能胡牌，因为剩余的单牌：", remainCards)
		return
	}

	rpTypess := self.player.GetCanReplaceType()
	logger.Info("可以替换的类型是：", rpTypess)

	if hzAmount > 0 {
		logger.Info("GenerateFinalPatternGroups: 分割输入模式组(非全红中/全红中)!")
		//2.分割固定替换(有本牌的替换)和任意替换(没有本牌的替换)
		notAllHZPatterns, allHZPatterns := SplitPatterns(patterns)
		PrintPatternsS("分割产生最终的模式列表：非全红中", notAllHZPatterns)
		PrintPatternsS("分割产生最终的模式列表：全红中", allHZPatterns)

		//3.计算模式列表的红中固定替换列表
		logger.Info("非全红中时==============")
		notAllHZRPPatterns := self.GenerateNotAllHZ(notAllHZPatterns)
		PrintPatternsS("非全红中时，", notAllHZPatterns)

		//4.根据替换类型，产生每一组替换类型对应的替换模式列表
		for _, rpTypes := range rpTypess {
			allHZRPPatterns := self.GenerateAllHZ(rpTypes, allHZPatterns)
			for _, pnts := range allHZRPPatterns {
				PrintPatternsS("全红中时，", pnts)
			}

			outAllRPList := make([][]*MaJiangPattern, 0)
			tempResult := make([]*MaJiangPattern, 0)
			self.GenerateCandidatePattern(0, append(notAllHZRPPatterns, allHZRPPatterns[:]...), &tempResult, &outAllRPList)

			for _, pnts := range outAllRPList {
				logger.Info("===========================rpTypes:", rpTypes)
				PrintPatternsS("所有的后备模式，", pnts)
			}
			for _, rp := range outAllRPList {
				self.GenerateOneFinalPatternGroups(rpTypes, rp, remainCards)
			}
		}
	} else {
		self.GenerateOneFinalPatternGroups(rpTypess[0], patterns, remainCards)
	}

	logger.Info("完成一次最终模式组的生成-GenerateFinalPatternGroups!")
	PrintPatternGroupsS("完成一次最终模式组的生成:", self.patternGroups, true)

	return result
}

//产生所有非全红中的模式的替换列表
func (self *HuController) GenerateNotAllHZ(notAllHZPatterns []*MaJiangPattern) (notAllHZRPPatterns [][]*MaJiangPattern) {
	notAllHZRPPatterns = make([][]*MaJiangPattern, len(notAllHZPatterns))
	if len(notAllHZPatterns) <= 0 {
		return
	}

	for i, hzp := range notAllHZPatterns {
		normalCards, hongZhongCards := SplitCards(hzp.cards)
		switch hzp.ptype {
		case PTKan:
			fallthrough
		case PTPair:
			notAllHZRPPatterns[i] = ComposeKanAnDZPatternsForNotAllHZ(normalCards, hongZhongCards, hzp.ptype)
		case PTSZ:
			notAllHZRPPatterns[i] = ComposeSZPatternsForNotAllHZ(normalCards, hongZhongCards)
		default:
			logger.Error("不能有其他类型的模式，只能是坎，顺子，对子。")
		}
	}

	return
}

//组合坎和对子模式通过本牌和红中牌
func ComposeKanAnDZPatternsForNotAllHZ(normalCards, hongZhongCards []*MaJiangCard, ptype int32) (result []*MaJiangPattern) {
	result = make([]*MaJiangPattern, 0)
	if len(normalCards) <= 0 {
		logger.Error("非所有红中，那么应该至少有一个是本派！")
		return
	}

	PrintCardsS("组合坎和对子模式通过本牌和红中牌时，本牌有：", normalCards)

	tempCards := []*MaJiangCard{}
	tempCards = append(tempCards, normalCards...)
	for _, hzc := range hongZhongCards {
		tempHZC := *hzc //copy
		tempHZC.SetHZReplaceValue(normalCards[0].cType, normalCards[0].value)
		tempCards = append(tempCards, &tempHZC)
	}

	result = append(result, NewPattern(ptype, tempCards, false))
	return
}

//组合顺子模式通过本牌和红中牌
func ComposeSZPatternsForNotAllHZ(normalCards, hongZhongCards []*MaJiangCard) (result []*MaJiangPattern) {
	result = make([]*MaJiangPattern, 0)
	ncAmount := len(normalCards)
	hzAmount := len(hongZhongCards)
	if ncAmount+hzAmount != 3 {
		logger.Error("顺子模式必须是3张牌！")
		return
	}

	//PrintCardsS("ComposeSZPatternsForNotAllHZ:普通牌：", normalCards)
	switch ncAmount {
	case 3:
		result = append(result, NewPattern(PTSZ, normalCards, false))
	case 2:
		firstCard := normalCards[0]
		secondCard := normalCards[1]
		valOffset := firstCard.value - secondCard.value
		if valOffset == 2 || valOffset == -2 {
			tempHZCard := *hongZhongCards[0]
			tempHZCard.SetHZReplaceValue(firstCard.cType, (firstCard.value+secondCard.value)/2)

			result = append(result, NewPattern(PTSZ, append(normalCards, &tempHZCard), false))
		}
		if valOffset == -1 {
			if firstCard.value-1 > 0 {
				tempHZCard := *hongZhongCards[0]
				tempHZCard.SetHZReplaceValue(firstCard.cType, firstCard.value-1)
				result = append(result, NewPattern(PTSZ, append(normalCards, &tempHZCard), false))
			}

			if secondCard.value+1 <= 9 {
				tempHZCard := *hongZhongCards[0]
				tempHZCard.SetHZReplaceValue(secondCard.cType, secondCard.value+1)
				result = append(result, NewPattern(PTSZ, append(normalCards, &tempHZCard), false))
			}
		}

		if valOffset == 1 {
			if secondCard.value-1 > 0 {
				tempHZCard := *hongZhongCards[0]
				tempHZCard.SetHZReplaceValue(secondCard.cType, secondCard.value-1)
				result = append(result, NewPattern(PTSZ, append(normalCards, &tempHZCard), false))
			}

			if firstCard.value+1 <= 9 {
				tempHZCard := *hongZhongCards[0]
				tempHZCard.SetHZReplaceValue(firstCard.cType, firstCard.value+1)
				result = append(result, NewPattern(PTSZ, append(normalCards, &tempHZCard), false))
			}
		}
	case 1:
		firstCard := normalCards[0]

		if firstCard.value-2 > 0 {
			tempHZCard1 := *hongZhongCards[0]
			tempHZCard1.SetHZReplaceValue(firstCard.cType, firstCard.value-1)
			tempHZCard2 := *hongZhongCards[1]
			tempHZCard2.SetHZReplaceValue(firstCard.cType, firstCard.value-2)

			result = append(result, NewPattern(PTSZ, append(append(normalCards, &tempHZCard1), &tempHZCard2), false))
		}

		if firstCard.value-1 > 0 && firstCard.value+1 <= 9 {
			tempHZCard1 := *hongZhongCards[0]
			tempHZCard1.SetHZReplaceValue(firstCard.cType, firstCard.value-1)
			tempHZCard2 := *hongZhongCards[1]
			tempHZCard2.SetHZReplaceValue(firstCard.cType, firstCard.value+1)

			result = append(result, NewPattern(PTSZ, append(append(normalCards, &tempHZCard1), &tempHZCard2), false))
		}

		if firstCard.value+2 <= 9 {
			tempHZCard1 := *hongZhongCards[0]
			tempHZCard1.SetHZReplaceValue(firstCard.cType, firstCard.value+1)
			tempHZCard2 := *hongZhongCards[1]
			tempHZCard2.SetHZReplaceValue(firstCard.cType, firstCard.value+2)

			result = append(result, NewPattern(PTSZ, append(append(normalCards, &tempHZCard1), &tempHZCard2), false))
		}

	default:
		logger.Error("本牌不能是其他数量！")
	}

	return
}

//产生所有全红中的模式下的指定替换类型下的的替换列表
func (self *HuController) GenerateAllHZ(rpTypes []int32, allHZPatterns []*MaJiangPattern) (allHZRPPatterns [][]*MaJiangPattern) {
	allHZRPPatterns = make([][]*MaJiangPattern, len(allHZPatterns))
	if len(allHZRPPatterns) <= 0 {
		return
	}

	for i, hzp := range allHZPatterns {
		hongZhongCards := hzp.cards
		switch hzp.ptype {
		case PTKan:
			fallthrough
		case PTPair:
			allHZRPPatterns[i] = ComposeKanAndDZPatternsForAllHZ(rpTypes, hongZhongCards, hzp.ptype)
		case PTSZ:
			allHZRPPatterns[i] = ComposeSZPatternsForAllHZ(rpTypes, hongZhongCards)
		default:
			logger.Error("不能有其他类型的模式，只能是坎，顺子，对子。")
		}
	}

	return
}

//组合坎和对子模式通过红中牌
func ComposeKanAndDZPatternsForAllHZ(rpTypes []int32, hongZhongCards []*MaJiangCard, ptype int32) (result []*MaJiangPattern) {
	result = make([]*MaJiangPattern, 0)
	if len(hongZhongCards) <= 0 {
		logger.Error("所有红中，红中的数量不应该小于等于0！")
		return
	}

	for _, rpType := range rpTypes {
		for i := 1; i <= 9; i++ {
			tempCards := []*MaJiangCard{}
			for _, hzc := range hongZhongCards {
				tempHZC := *hzc //copy
				tempHZC.SetHZReplaceValue(rpType, int32(i))
				tempCards = append(tempCards, &tempHZC)
			}

			result = append(result, NewPattern(ptype, tempCards, false))
		}
	}

	return
}

//组合顺子模式通过红中牌
func ComposeSZPatternsForAllHZ(rpTypes []int32, hongZhongCards []*MaJiangCard) (result []*MaJiangPattern) {
	result = make([]*MaJiangPattern, 0)

	if len(hongZhongCards) != 3 {
		logger.Error("顺子模式必须是3张牌！")
		return
	}

	for _, rpType := range rpTypes {
		for i := 1; i <= 7; i++ {
			tempCards := []*MaJiangCard{}
			for hi, hzc := range hongZhongCards {
				tempHZC := *hzc //copy
				tempHZC.SetHZReplaceValue(rpType, int32(i+hi))
				tempCards = append(tempCards, &tempHZC)
			}

			result = append(result, NewPattern(PTSZ, tempCards, false))
		}
	}

	return
}

//排列生成用于生成最终模式的模式的后备模式列表
func (self *HuController) GenerateCandidatePattern(curIndex int, hzRPPatterns [][]*MaJiangPattern, tempResult *[]*MaJiangPattern, outAllRPList *[][]*MaJiangPattern) {
	//1. 检查参数
	if outAllRPList == nil || tempResult == nil {
		logger.Error("参数错误！", outAllRPList, tempResult)
		return
	}

	if curIndex >= len(hzRPPatterns) {
		PrintPatternsS("生成所有后备队列时的其中一组：", *tempResult)
		temp := make([]*MaJiangPattern, len(*tempResult))
		copy(temp, *tempResult)
		*outAllRPList = append(*outAllRPList, temp)
		return
	}

	loopList := hzRPPatterns[curIndex]
	curIndex++
	for _, p := range loopList {
		*tempResult = append(*tempResult, p)
		self.GenerateCandidatePattern(curIndex, hzRPPatterns, tempResult, outAllRPList)
		oldTempResult := *tempResult
		*tempResult = oldTempResult[:len(oldTempResult)-1]
	}

}

//产生一组最终的模式
func (self *HuController) GenerateOneFinalPatternGroups(rpTypes []int32, hzRPPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {
	if IsAllHongZhong(singleCards) {
		self.GenerateOneFinalPatternGroupsForKaoAllHZ(rpTypes, hzRPPatterns, singleCards)
	} else {
		self.GenerateOneFinalPatternGroupsForKaoNotAllHZ(rpTypes, hzRPPatterns, singleCards)
	}
}

//产生一组最终的模式，对于靠牌全是红中的情况
func (self *HuController) GenerateOneFinalPatternGroupsForKaoAllHZ(rpTypes []int32, hzRPPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {
	//1.检查输入参数
	if len(rpTypes) <= 0 || len(singleCards) <= 0 {
		logger.Error("替换类型为空。")
		return
	}

	switch len(singleCards) {
	//单调
	case 1:
		for _, rpType := range rpTypes {
			for i := 1; i <= 9; i++ {
				tempHZC := *singleCards[0] //copy
				tempHZC.SetHZReplaceValue(rpType, int32(i))
				kaoCards := []*MaJiangCard{&tempHZC}

				pg := NewPatternGroup(hzRPPatterns)
				pg.kaoCards = kaoCards

				huCards := []*MaJiangCard{NewCard(0, rpType, int32(i))}
				pg.huCards = huCards
				self.AddOnePatternGroup(pg)
			}
		}
	//顺子
	case 2:
		//连续的
		for _, rpType := range rpTypes {
			for i := 2; i <= 7; i++ {
				tempHZC1 := *singleCards[0] //copy
				tempHZC1.SetHZReplaceValue(rpType, int32(i))

				tempHZC2 := *singleCards[1] //copy
				tempHZC2.SetHZReplaceValue(rpType, int32(i+1))

				kaoCards := []*MaJiangCard{&tempHZC1, &tempHZC2}

				pg := NewPatternGroup(hzRPPatterns)
				pg.kaoCards = kaoCards

				huCards := []*MaJiangCard{NewCard(0, rpType, int32(i-1)), NewCard(0, rpType, int32(i+2))}
				pg.huCards = huCards
				self.AddOnePatternGroup(pg)
			}
		}

		//间隔的
		for _, rpType := range rpTypes {
			for i := 1; i <= 7; i++ {
				tempHZC1 := *singleCards[0] //copy
				tempHZC1.SetHZReplaceValue(rpType, int32(i))

				tempHZC2 := *singleCards[1] //copy
				tempHZC2.SetHZReplaceValue(rpType, int32(i+2))

				kaoCards := []*MaJiangCard{&tempHZC1, &tempHZC2}

				pg := NewPatternGroup(hzRPPatterns)
				pg.kaoCards = kaoCards

				huCards := []*MaJiangCard{NewCard(0, rpType, int32(i+1))}
				pg.huCards = huCards
				self.AddOnePatternGroup(pg)
			}
		}
	//对处
	case 4:
		rpTypeAmount := len(rpTypes)
		for i := 1; i <= 9*rpTypeAmount; i++ {
			iRPType := rpTypes[i/10]
			iRPValue := int32(((i - 1) % 9) + 1)

			for j := 1; j <= 9*rpTypeAmount; j++ {
				tempHZC1 := *singleCards[0] //copy
				tempHZC1.SetHZReplaceValue(iRPType, iRPValue)

				tempHZC2 := *singleCards[1] //copy
				tempHZC2.SetHZReplaceValue(iRPType, iRPValue)

				jRPType := rpTypes[j/10]
				jRPValue := int32(((j - 1) % 9) + 1)
				tempHZC3 := *singleCards[2] //copy
				tempHZC3.SetHZReplaceValue(jRPType, jRPValue)

				tempHZC4 := *singleCards[3] //copy
				tempHZC4.SetHZReplaceValue(jRPType, jRPValue)

				kaoCards := []*MaJiangCard{&tempHZC1, &tempHZC2, &tempHZC3, &tempHZC4}

				pg := NewPatternGroup(hzRPPatterns)
				pg.kaoCards = kaoCards

				huCards := []*MaJiangCard{NewCard(0, iRPType, iRPValue), NewCard(0, jRPValue, jRPValue)}
				pg.huCards = huCards
				self.AddOnePatternGroup(pg)
			}
		}
	default:
		logger.Error("剩余的单牌数量只能是：1, 2, 4")
	}

}

//产生一组最终的模式，对于靠牌非全红中的情况
func (self *HuController) GenerateOneFinalPatternGroupsForKaoNotAllHZ(rpTypes []int32, hzRPPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {
	//1.检查输入参数
	if len(rpTypes) <= 0 || len(singleCards) <= 0 {
		logger.Error("替换类型为空。")
		return
	}

	logger.Info("产生一个最终的模式组(靠牌非全红中)：替换类型：", rpTypes)
	PrintPatternsS("产生一个最终的模式组(靠牌非全红中)：模式列表：", hzRPPatterns)
	PrintCardsS("产生一个最终的模式组(靠牌非全红中)：单牌：", singleCards)

	switch len(singleCards) {
	//单调
	case 1:
		fallthrough
	//顺子
	case 2:
		self.GenerateOneFinalPatternGroupsForKaoNotAllHZ_DD_SZ(rpTypes, hzRPPatterns, singleCards)
	//对处
	case 4:
		self.GenerateOneFinalPatternGroupsForKaoNotAllHZ_DC(rpTypes, hzRPPatterns, singleCards)
	default:
		logger.Error("剩余的单牌数量只能是：1, 2, 4")
	}
}

//产生一组最终的模式，对于靠牌非全红中的情况下的单调
func (self *HuController) GenerateOneFinalPatternGroupsForKaoNotAllHZ_DD_SZ(rpTypes []int32, hzRPPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {
	//1.检查输入参数
	if len(rpTypes) <= 0 || len(singleCards) <= 0 {
		logger.Error("替换类型为空。")
		return
	}

	//排序单牌
	sort.Sort(CardList(singleCards))

	normalCards, hongZhongCards := SplitCards(singleCards)
	if len(normalCards) <= 0 {
		logger.Error("在非所有红中的模式下，不可能没有本牌")
		return
	}

	// szPatterns, otherSZPatterns := GetSZPatternsByType(hzRPPatterns, normalCards[0].cType)
	// szPatternsC, otherSZPatternsC := GetAllPatternsCombination(2, szPatterns, false)
	// if len(szPatternsC) != len(otherSZPatternsC) {
	// 	logger.Error("获取所有模式的组合情况不正确！")
	// 	for _, szP := range szPatternsC {
	// 		PrintPatternsS("szPatternsC：", szP)
	// 	}
	// 	for _, szP := range otherSZPatternsC {
	// 		PrintPatternsS("otherSZPatternsC：", szP)
	// 	}

	// 	return
	// }

	//PrintPatternsS("本牌靠牌对应的顺子：", szPatterns)
	// for i, szP := range szPatternsC {
	// 	PrintPatternsS("本牌靠牌对应的顺子可能组合情况：", szP)
	// 	PrintPatternsS("本牌靠牌对应的顺子可能组合情况_剩余的模式组：", otherSZPatternsC[i])
	// }

	switch len(singleCards) {
	//在非全红中的情况下，如果只有一个牌的话，那么这个牌一定不是红中，所以不用考虑singleCards的替换情况
	case 1:
		// if len(szPatternsC) > 0 {
		// 	for i, szC := range szPatternsC {
		// 		fixedPatterns := append(otherSZPatterns, otherSZPatternsC[i]...)
		// 		addType, sortedPatterns := self.CalcAddPosForSZ(szC, singleCards)
		// 		logger.Info("顺子的替换未知（前，中，后等）AddType:%d", addType)
		// 		PrintPatternsS("本牌靠牌对应的顺子可能组合情况：", sortedPatterns)
		// 		if addType != UnknowAdd {
		// 			self.CalcAddForSZ(addType, sortedPatterns, singleCards, fixedPatterns)
		// 		} else {
		// 			pg := NewPatternGroup(append(fixedPatterns, ClonePatterns(szC)...))
		// 			tempCard := *singleCards[0]
		// 			pg.kaoCards = []*MaJiangCard{&tempCard}
		// 			curCType, curValue := tempCard.CurValue()
		// 			pg.huCards = []*MaJiangCard{NewCard(0, curCType, curValue)}
		// 			self.AddOnePatternGroup(pg)
		// 		}
		// 	}
		// } else {
		pg := NewPatternGroup(hzRPPatterns)
		tempCard := *singleCards[0]
		pg.kaoCards = []*MaJiangCard{&tempCard}
		curCType, curValue := tempCard.CurValue()
		pg.huCards = []*MaJiangCard{NewCard(0, curCType, curValue)}
		self.AddOnePatternGroup(pg)
		//}

	case 2:

		//计算可以替换顺子的列表
		rpCardsList := make([][]*MaJiangCard, 0)
		if len(hongZhongCards) == 1 {
			rpType := normalCards[0].cType
			nVal := normalCards[0].value
			if nVal-1 > 0 {
				clonedCard := *hongZhongCards[0]
				clonedCard.SetHZReplaceValue(rpType, nVal-1)
				rpCardsList = append(rpCardsList, []*MaJiangCard{&clonedCard, normalCards[0]})
			}
			if nVal+1 <= 9 {
				clonedCard := *hongZhongCards[0]
				clonedCard.SetHZReplaceValue(rpType, nVal+1)
				rpCardsList = append(rpCardsList, []*MaJiangCard{normalCards[0], &clonedCard})
			}
		} else {
			rpCardsList = append(rpCardsList, normalCards)
		}

		//生成模式组
		for _, rpSingleCards := range rpCardsList {
			//for i, szC := range szPatternsC {
			// fixedPatterns := append(otherSZPatterns, otherSZPatternsC[i]...)
			// addType, sortedPatterns := self.CalcAddPosForSZ(szC, rpSingleCards)
			// if addType != UnknowAdd {
			// 	self.CalcAddForSZ(addType, sortedPatterns, rpSingleCards, fixedPatterns)
			// } else {
			pg := NewPatternGroup(hzRPPatterns)

			firstCard := *rpSingleCards[0]
			secondCard := *rpSingleCards[1]
			pg.kaoCards = []*MaJiangCard{&firstCard, &secondCard}
			pg.huCards = []*MaJiangCard{}

			valOffset := firstCard.value - secondCard.value
			switch valOffset {
			case -2, 2:
				pg.huCards = append(pg.huCards, NewCard(0, normalCards[0].cType, (firstCard.value+secondCard.value)/2))
			case -1:
				if firstCard.value-1 > 0 {
					pg.huCards = append(pg.huCards, NewCard(0, normalCards[0].cType, firstCard.value-1))
				}

				if secondCard.value+1 <= 9 {
					pg.huCards = append(pg.huCards, NewCard(0, normalCards[0].cType, secondCard.value+1))
				}
			case 1:
				if secondCard.value-1 > 0 {
					pg.huCards = append(pg.huCards, NewCard(0, normalCards[0].cType, secondCard.value-1))
				}

				if firstCard.value+1 <= 9 {
					pg.huCards = append(pg.huCards, NewCard(0, normalCards[0].cType, firstCard.value+1))
				}
			default:
				logger.Error("值有问题！first:%d, second:%d", firstCard.value, secondCard.value)
			}

			self.AddOnePatternGroup(pg)
			//}
			//}
		}
	default:
		logger.Error("只能1或2单牌")
	}

}

//产生一组最终的模式，对于靠牌非全红中的情况下的对处
func (self *HuController) GenerateOneFinalPatternGroupsForKaoNotAllHZ_DC(rpTypes []int32, hzRPPatterns []*MaJiangPattern, singleCards []*MaJiangCard) {
	if len(singleCards) != 4 {
		logger.Error("对处的单牌必须是4张！")
		return
	}

	normalCards, hongZhongCards := SplitCards(singleCards)
	if len(normalCards) <= 0 {
		logger.Error("必须要有个本牌！")
		return
	}

	AAPatterns, tempSingleCards := SplitToAA_A(normalCards)

	switch len(hongZhongCards) {
	case 0:
		if len(AAPatterns) != 2 {
			logger.Error("必须是两个对子模式！")
			break
		}
		pg := NewPatternGroup(hzRPPatterns)

		pg.kaoCards = CloneMaJiangCards(singleCards)
		pg.huCards = []*MaJiangCard{
			NewCard(0, AAPatterns[0].cards[0].cType, AAPatterns[0].cards[0].value),
			NewCard(0, AAPatterns[1].cards[0].cType, AAPatterns[1].cards[0].value)}

		self.AddOnePatternGroup(pg)

	case 1:
		if len(AAPatterns) != 1 {
			logger.Error("必须有一个本牌的对子模式！")
			break
		}

		pg := NewPatternGroup(hzRPPatterns)

		tempHongZhong := *hongZhongCards[0]
		curCType, curValue := tempSingleCards[0].CurValue()
		tempHongZhong.SetHZReplaceValue(curCType, curValue)
		pg.kaoCards = append(normalCards, &tempHongZhong)

		pg.huCards = []*MaJiangCard{NewCard(0, AAPatterns[0].cards[0].cType, AAPatterns[0].cards[0].value), NewCard(0, curCType, curValue)}

		self.AddOnePatternGroup(pg)
	case 2:
		if len(AAPatterns) == 1 {
			for _, rpType := range rpTypes {
				for i := 1; i <= 9; i++ {
					pg := NewPatternGroup(hzRPPatterns)

					tempHongZhong1 := *hongZhongCards[0]
					tempHongZhong2 := *hongZhongCards[1]
					tempHongZhong1.SetHZReplaceValue(rpType, int32(i))
					tempHongZhong2.SetHZReplaceValue(rpType, int32(i))

					pg.kaoCards = append(normalCards, &tempHongZhong1)
					pg.kaoCards = append(pg.kaoCards, &tempHongZhong2)

					pg.huCards = []*MaJiangCard{NewCard(0, AAPatterns[0].cards[0].cType, AAPatterns[0].cards[0].value)}
					pg.huCards = append(pg.huCards, NewCard(0, rpType, int32(i)))

					self.AddOnePatternGroup(pg)
				}
			}
		} else {
			pg := NewPatternGroup(hzRPPatterns)

			tempHongZhong1 := *hongZhongCards[0]
			tempHongZhong2 := *hongZhongCards[1]
			tempHongZhong1.SetHZReplaceValue(normalCards[0].cType, normalCards[0].value)
			tempHongZhong2.SetHZReplaceValue(normalCards[1].cType, normalCards[1].value)

			pg.kaoCards = append(normalCards, &tempHongZhong1)
			pg.kaoCards = append(pg.kaoCards, &tempHongZhong2)

			pg.huCards = []*MaJiangCard{
				NewCard(0, normalCards[0].cType, normalCards[0].value),
				NewCard(0, normalCards[1].cType, normalCards[1].value)}

			self.AddOnePatternGroup(pg)
		}
	case 3:
		for _, rpType := range rpTypes {
			for i := 1; i <= 9; i++ {
				pg := NewPatternGroup(hzRPPatterns)

				tempHongZhong1 := *hongZhongCards[0]
				tempHongZhong2 := *hongZhongCards[1]
				tempHongZhong1.SetHZReplaceValue(rpType, int32(i))
				tempHongZhong2.SetHZReplaceValue(rpType, int32(i))

				tempHongZhong3 := *hongZhongCards[2]
				tempHongZhong3.SetHZReplaceValue(normalCards[0].cType, normalCards[0].value)

				pg.kaoCards = append(normalCards, &tempHongZhong1)
				pg.kaoCards = append(pg.kaoCards, &tempHongZhong2)
				pg.kaoCards = append(pg.kaoCards, &tempHongZhong3)

				pg.huCards = []*MaJiangCard{NewCard(0, rpType, int32(i))}
				pg.huCards = append(pg.huCards, NewCard(0, normalCards[0].cType, normalCards[0].value))

				self.AddOnePatternGroup(pg)
			}
		}
	default:
		logger.Error("红只可能是0， 1， 2和3个的情况！")
	}

}

//叠加情况
const (
	UnknowAdd = iota
	FrontAdd
	BackAdd
	MiddleAdd
)

//计算叠加位置
func (self *HuController) CalcAddPosForSZ(patterns []*MaJiangPattern, singleCards []*MaJiangCard) (addType int32, sortedPatterns []*MaJiangPattern) {

	addType = UnknowAdd
	sortedPatterns = patterns
	if len(patterns) <= 0 {
		return
	}

	switch len(patterns) {
	case 1:

		//确定addType
		switch len(singleCards) {
		case 1:
			if patterns[0].cards[0].value-singleCards[0].value == 1 {
				addType = FrontAdd
			}

			if singleCards[0].value-patterns[0].cards[2].value == 1 {
				addType = BackAdd
			}

		case 2:
			if singleCards[1].value-singleCards[0].value == 1 {
				if patterns[0].cards[0].value-singleCards[1].value == 1 {
					addType = FrontAdd
				}

				if singleCards[0].value-patterns[0].cards[2].value == 1 {
					addType = BackAdd
				}
			}
		default:
			logger.Error("不能支持超过2个的单牌数量！")
		}
	case 2:
		//对patterns进行排序
		firstPattern := patterns[0]
		secondPattern := patterns[1]
		if firstPattern.cards[0].value > secondPattern.cards[0].value {
			firstPattern = patterns[1]
			secondPattern = patterns[0]
		}

		sortedPatterns = []*MaJiangPattern{firstPattern, secondPattern}
		//确定addType
		switch len(singleCards) {
		case 1:
			patternInterval := secondPattern.cards[0].value - firstPattern.cards[2].value
			if patternInterval == 1 {
				if firstPattern.cards[0].value-singleCards[0].value == 1 {
					addType = FrontAdd
				}

				if singleCards[0].value-secondPattern.cards[2].value == 1 {
					addType = BackAdd
				}
			} else if patternInterval == 2 {
				if singleCards[0].value-firstPattern.cards[2].value == 1 {
					addType = MiddleAdd
				}
			}
		case 2:

			patternInterval := secondPattern.cards[0].value - firstPattern.cards[2].value
			if patternInterval == 1 {
				if singleCards[1].value-singleCards[0].value == 1 {
					if firstPattern.cards[0].value-singleCards[1].value == 1 {
						addType = FrontAdd
					}

					if singleCards[0].value-secondPattern.cards[2].value == 1 {
						addType = BackAdd
					}
				}
			} else if patternInterval == 3 {
				if singleCards[0].value-firstPattern.cards[2].value == 1 {
					addType = MiddleAdd
				}
			}

		default:
			logger.Error("不能支持超过2个的单牌数量！")
		}
	default:
		logger.Error("不支持超过两个的pattern！")
		PrintPatternsS("不支持超过两个的pattern:", patterns)
	}

	return
}

//计算单牌和已有顺子模式的叠加情况
func (self *HuController) CalcAddForSZ(addType int32, patterns []*MaJiangPattern, singleCards []*MaJiangCard, fixedPatterns []*MaJiangPattern) {
	//检查输入参数
	if len(patterns) <= 0 || len(singleCards) <= 0 {
		logger.Error("patterns or singleCards is nil:", patterns, singleCards)
		return
	}

	singleAmount := len(singleCards)
	if !(singleAmount == 1 || singleAmount == 2) {
		logger.Error("不是顺子的方式")
		return
	}

	tempCards := []*MaJiangCard{}
	switch addType {
	case FrontAdd:
		tempCards = append(tempCards, singleCards...)
		for _, p := range patterns {
			tempCards = append(tempCards, p.cards...)
		}

	case BackAdd:
		for _, p := range patterns {
			tempCards = append(tempCards, p.cards...)
		}
		tempCards = append(tempCards, singleCards...)
	case MiddleAdd:
		if len(patterns) != 2 {
			logger.Error("中间添加的话，必须是两个模式，才能被放在中间！")
			return
		}

		firstP := patterns[0]
		tempCards = append(tempCards, firstP.cards...)
		tempCards = append(tempCards, singleCards...)
		secondP := patterns[1]
		tempCards = append(tempCards, secondP.cards...)
	default:
		logger.Error("添加类型错误！")
	}

	//检查牌的最低限制
	clonedTempCards := CloneMaJiangCards(tempCards)
	clonedTCAmount := len(clonedTempCards)

	PrintCardsS("CalcAddForSZ.clonedTempCards:", clonedTempCards)
	tempSingleCardAmount := clonedTCAmount % 3
	switch tempSingleCardAmount {
	case 1:
		if clonedTCAmount < 4 {
			logger.Error("至少是4张牌")
			return
		}
		for i := 0; i < clonedTCAmount; i += 3 {
			remainPtns := []*MaJiangPattern{}
			if i == 0 {
				remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i+1:i+4], false))
				if clonedTCAmount >= 7 {
					remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i+4:i+7], false))
				}

			} else if i == 3 {
				remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i-3:i], false))
				if clonedTCAmount >= 7 {
					remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i+1:i+4], false))
				}
			} else if i == 6 {
				remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i-6:i-3], false))
				if clonedTCAmount >= 7 {
					remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i-3:i], false))
				}
			} else {
				logger.Error("不应该出现其他的情况")
			}

			// PrintPatternsS("CalcAddForSZ.fixedPatterns:", fixedPatterns)
			// PrintPatternsS("CalcAddForSZ.remainPtns:", remainPtns)
			resultPtns := []*MaJiangPattern{}
			resultPtns = append(resultPtns, fixedPatterns...)
			pg := NewPatternGroup(append(resultPtns, remainPtns...))
			pg.kaoCards = []*MaJiangCard{clonedTempCards[i]}
			curCType, curValue := clonedTempCards[i].CurValue()
			pg.huCards = []*MaJiangCard{NewCard(0, curCType, curValue)}
			self.AddOnePatternGroup(pg)
			//PrintPatternGroupsS("CalcAddForSZ.patternGroups", self.patternGroups, true)
		}
	case 2:
		if clonedTCAmount < 5 {
			logger.Error("至少是5张牌")
			return
		}
		for i := 0; i < clonedTCAmount; i += 3 {
			remainPtns := []*MaJiangPattern{}
			if i == 0 {
				remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i+2:i+5], false))
				if clonedTCAmount >= 8 {
					remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i+5:i+8], false))
				}

			} else if i == 3 {
				remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i-3:i], false))
				if clonedTCAmount >= 8 {
					remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i+2:i+5], false))
				}
			} else if i == 6 {
				remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i-6:i-3], false))
				if clonedTCAmount >= 8 {
					remainPtns = append(remainPtns, NewPattern(PTSZ, clonedTempCards[i-3:i], false))
				}
			} else {
				logger.Error("不应该出现其他的情况")
			}

			resultPtns := []*MaJiangPattern{}
			resultPtns = append(resultPtns, fixedPatterns...)
			pg := NewPatternGroup(append(resultPtns, remainPtns...))
			pg.kaoCards = []*MaJiangCard{clonedTempCards[i], clonedTempCards[i+1]}
			prevCType, prevValue := clonedTempCards[i].CurValue()
			backCType, backValue := clonedTempCards[i+1].CurValue()
			pg.huCards = []*MaJiangCard{}
			if prevValue-1 > 0 {
				pg.huCards = append(pg.huCards, NewCard(0, prevCType, prevValue-1))
			}
			if backValue+1 <= 9 {
				pg.huCards = append(pg.huCards, NewCard(0, backCType, backValue+1))
			}
			self.AddOnePatternGroup(pg)
		}

	default:
		logger.Error("只能是1或2！")
	}
}

//添加一个模式组到最终的模式组里
func (self *HuController) AddOnePatternGroup(pg *MaJiangPatternGroup) {
	if pg == nil {
		return
	}

	pg.GenerateID()
	if !IsExistPatternGroup(self.patternGroups, pg) {
		logger.Info("在添加到最终模式组列表时检查是否重复：ID:", pg.id)
		self.patternGroups = append(self.patternGroups, pg)
	}
}

//仅需要产生大对子胡模式组吗
func (self *HuController) IsOnlyGenerateDaDuiZi(cardAmount, hzCardAmount int32) bool {
	if cardAmount == 13 {
		return false
	}

	return cardAmount*2 < hzCardAmount
}

//仅需要产生小七对胡模式组吗
func (self *HuController) IsOnlyGenerateXiaoQiDui(cardAmount, hzCardAmount int32) bool {
	if cardAmount != 13 {
		return false
	}

	return cardAmount < hzCardAmount
}

//检查是否存在相同的PatternGroup
func IsExistPatternGroup(pgs []*MaJiangPatternGroup, pg *MaJiangPatternGroup) bool {
	if len(pgs) <= 0 {
		return false
	}

	for _, p := range pgs {
		if p.IsEqual(pg) {
			return true
		}
	}

	return false
}

//检查靠能否胡牌
func CanHu(cards []*MaJiangCard) bool {
	//检查数量
	cardAmount := len(cards)
	if !(cardAmount == 1 || cardAmount == 2 || cardAmount == 4) {
		logger.Error("剩余检查胡的牌的数量只能是1， 2， 4")
		return false
	}

	//检查两张牌时
	if cardAmount == 2 && GetHongZhongAmount(cards) <= 0 {
		//检查花色
		if !IsSameHuaSe(cards, false) {
			return false
		}

		valOffset := cards[0].value - cards[1].value
		if valOffset > 2 || valOffset < -2 || valOffset == 0 {
			return false
		}
	}

	//检查4张牌时
	if cardAmount == 4 {
		normalcards, hongZhongCards := SplitCards(cards)
		AAPatterns, _ := SplitToAA_A(normalcards)
		hzAmount := len(hongZhongCards)
		AAAmount := len(AAPatterns)
		if AAAmount == 0 && hzAmount <= 1 {
			return false
		}

		if AAAmount == 1 && hzAmount <= 0 {
			return false
		}
	}

	return true
}

//分开本牌和红中
func SplitCards(cards []*MaJiangCard) (normalCards, hongZhongCards []*MaJiangCard) {
	normalCards = make([]*MaJiangCard, 0)
	hongZhongCards = make([]*MaJiangCard, 0)

	if len(cards) <= 0 {
		//logger.Error("cards is empty!")
		return
	}

	for _, c := range cards {
		if c.IsHongZhong() {
			hongZhongCards = append(hongZhongCards, c)
		} else {
			normalCards = append(normalCards, c)
		}
	}
	return
}

//分开本牌和红中
func SplitCardsToHuaSe(normalCards []*MaJiangCard) (splitedCards map[int32][]*MaJiangCard) {
	splitedCards = make(map[int32][]*MaJiangCard, 0)

	if len(normalCards) <= 0 {
		return
	}

	for _, c := range normalCards {
		_, exist := splitedCards[c.cType]
		if !exist {
			splitedCards[c.cType] = []*MaJiangCard{}
		}
		splitedCards[c.cType] = append(splitedCards[c.cType], c)
	}
	return
}

//分割全红中和非全红中的模式
func SplitPatterns(patterns []*MaJiangPattern) (notAllHZPatterns, allHZPatterns []*MaJiangPattern) {
	notAllHZPatterns = make([]*MaJiangPattern, 0)
	allHZPatterns = make([]*MaJiangPattern, 0)
	if len(patterns) <= 0 {
		return
	}

	for _, p := range patterns {
		if p.IsAllHZ() {
			allHZPatterns = append(allHZPatterns, p)
		} else {
			notAllHZPatterns = append(notAllHZPatterns, p)
		}
	}

	return
}

//分割不同花色的模式
func SplitHuaSePatterns(patterns []*MaJiangPattern) (result map[int32][]*MaJiangPattern) {
	result = make(map[int32][]*MaJiangPattern, 0)
	if len(patterns) <= 0 {
		return
	}

	notAllHZPatterns, allHZPatterns := SplitPatterns(patterns)
	result[HongZhong] = allHZPatterns
	result[Tiao] = []*MaJiangPattern{}
	result[Tong] = []*MaJiangPattern{}
	result[Wan] = []*MaJiangPattern{}

	for _, p := range notAllHZPatterns {
		for _, c := range p.cards {
			if !c.IsHongZhong() {
				result[c.cType] = append(result[c.cType], p)
				break
			}
		}
	}

	return
}

//获取指定花色的顺子
func GetSZPatternsByType(patterns []*MaJiangPattern, cType int32) (result, other []*MaJiangPattern) {
	result = make([]*MaJiangPattern, 0)
	other = make([]*MaJiangPattern, 0)
	if len(patterns) <= 0 {
		return
	}

	for _, p := range patterns {
		isFind := false
		if p.ptype == PTSZ {
			for _, c := range p.cards {
				if cType == c.cType {
					result = append(result, p)
					isFind = true
					break
				}
			}
		}

		if !isFind {
			other = append(other, p)
		}
	}

	return
}

//获取1或2个模式的所有组合情况
func GetAllPatternsCombination(n int, patterns []*MaJiangPattern, isOnlyN bool) (result, other [][]*MaJiangPattern) {
	result = make([][]*MaJiangPattern, 0)
	other = make([][]*MaJiangPattern, 0)
	patternAmount := len(patterns)
	if patternAmount <= 0 {
		result = append(result, []*MaJiangPattern{})
		other = append(other, []*MaJiangPattern{})
		return
	}

	if n <= 0 {
		result = append(result, []*MaJiangPattern{})
		temp := make([]*MaJiangPattern, len(patterns))
		copy(temp, patterns)
		other = append(other, temp)
		return
	}

	//生成用于进行排列组合的下标
	indexes := make([]int32, patternAmount)
	for i := 0; i < patternAmount; i++ {
		indexes[i] = int32(i)
	}

	//进行组合
	maxCAmount := int(math.Min(float64(patternAmount), float64(n)))
	i := 1
	if isOnlyN {
		i = maxCAmount
	}
	for ; i <= maxCAmount; i++ {
		cIndexes := C(int32(i), int32(patternAmount))
		for _, cIndex := range cIndexes {

			//没有选择中的添加到列表里
			tempOther := []*MaJiangPattern{}
			remainIndexes := Exclude(indexes, cIndex)
			for _, ri := range remainIndexes {
				tempOther = append(tempOther, ClonePattern(patterns[ri]))
			}
			other = append(other, tempOther)

			//把选择中的添加到列表里
			temp := []*MaJiangPattern{}
			for _, index := range cIndex {
				temp = append(temp, ClonePattern(patterns[index]))
			}
			result = append(result, temp)
		}
	}

	return
}

//获取顺子的组合情况
func GetSZPatternsCombination(n int, patterns []*MaJiangPattern) (result [][]*MaJiangPattern) {
	patternAmount := len(patterns)
	cIndexes := C(int32(n), int32(patternAmount))
	for _, cIndex := range cIndexes {
		temp := []*MaJiangPattern{}

		for _, index := range cIndex {
			temp = append(temp, patterns[index])
		}
		result = append(result, temp)
	}
	return
}

//克隆Patterns列表
func ClonePatterns(src []*MaJiangPattern) (dst []*MaJiangPattern) {
	if src == nil {
		return
	}

	dst = make([]*MaJiangPattern, len(src))
	for i, v := range src {
		dst[i] = ClonePattern(v)
	}

	return

}

//克隆一个Pattern
func ClonePattern(src *MaJiangPattern) (dst *MaJiangPattern) {
	if src == nil {
		return
	}

	dst = &MaJiangPattern{}

	dst.id = src.id
	dst.ptype = src.ptype
	dst.cType = src.cType
	dst.cards = CloneMaJiangCards(src.cards)
	dst.isShowPattern = src.isShowPattern

	return

}

//复制牌
func CloneMaJiangCards(src []*MaJiangCard) (dst []*MaJiangCard) {
	if src == nil {
		return nil
	}

	dst = make([]*MaJiangCard, len(src))
	for i, v := range src {
		temp := *v
		dst[i] = &temp
	}

	return
}

//获取除自定数字以外的其他数字
func Exclude(ids []int32, excludeIds []int32) []int32 {
	if len(ids) <= 0 || len(excludeIds) <= 0 {
		return ids
	}

	result := []int32{}
	for _, id := range ids {
		if !Exist(excludeIds, id) {
			result = append(result, id)
		}
	}

	return result
}

//生成模式的组合下标
func C(n int32, m int32) (result [][]int32) {
	if n > m {
		logger.Error("n must be less than m", n, m)
		return nil
	}

	//	fmt.Println("组合数量：", n, m, size)
	result = make([][]int32, 0)

	index := make([]int32, m)
	//fmt.Println(m, index)
	for i := 0; int32(i) < m; i++ {
		index[i] = 0
	}

	for i := 0; int32(i) < n; i++ {
		index[i] = 1
	}

	result = append(result, GetC(index))

	for true {
		for i := 0; int32(i) < m-1; i++ {
			if index[i] == 1 && index[i+1] == 0 {
				oneIndex := 0
				for j := 0; j < i; j++ {
					if index[j] == 1 {
						index[j] = 0
						index[oneIndex] = 1
						oneIndex++
					}
				}

				index[i] = 0
				index[i+1] = 1

				result = append(result, GetC(index))
				break
			}
		}

		//check is end
		isEnd := true
		for k := m - n; k < m; k++ {
			if index[k] != 1 {
				isEnd = false
				break
			}
		}

		if isEnd {
			break
		}
	}

	return
}

func GetC(index []int32) (result []int32) {
	result = make([]int32, 0)
	for i := 0; i < len(index); i++ {
		if index[i] == 1 {
			result = append(result, int32(i))
		}
	}

	return
}

//检查是否是全红中的牌
func IsAllHongZhong(cards []*MaJiangCard) bool {
	return int32(len(cards)) == GetHongZhongAmount(cards)
}

//获取红中的数量
func GetHongZhongAmount(cards []*MaJiangCard) (result int32) {
	result = 0
	if len(cards) <= 0 {
		return result
	}

	for _, c := range cards {
		if c.IsHongZhong() {
			result++
		}
	}

	return result
}

//检查是否是同一种牌类型
func IsSameHuaSe(cards []*MaJiangCard, checkHZ bool) bool {
	if len(cards) <= 0 {
		return true
	}

	tempCards := make([]*MaJiangCard, len(cards))
	copy(tempCards, cards)

	//移除红中
	for !checkHZ {
		curIndex := 0
		for i, c := range tempCards {
			curIndex = i
			if c.IsHongZhong() {
				tempCards = append(tempCards[:i], tempCards[i+1:]...)
				break
			}
		}
		//全部检查完了
		if curIndex+1 >= len(tempCards) {
			break
		}
	}

	//检查是否花色相同
	if len(tempCards) > 0 {
		firstCard := tempCards[0]
		for _, c := range tempCards {
			if firstCard.cType != c.cType {
				return false
			}
		}
	}

	return true
}

//从一个切片中移除指定类型的Card
func RemoveCardByType(cards []*MaJiangCard, cType, value int32) (success bool, result []*MaJiangCard) {
	if cards == nil {
		logger.Error("RemoveCardByType:cards is nil.")
		return false, nil
	}

	for i, v := range cards {
		curCType, curVal := v.CurValue()
		if curVal == value && curCType == cType {
			cards = append(cards[:i], cards[i+1:]...)
			return true, cards
		}
	}

	return false, cards
}

//从一个切片中移除指定类型的多个Card
func RemoveCardsByType(cards []*MaJiangCard, cType, value, wantRemovedAmount int32) (result []*MaJiangCard, outRemovedCards []*MaJiangCard) {

	if cards == nil {
		logger.Error("RemoveCardsByType:cards is nil.")
		return
	}

	//removedAmount := 0
	result = make([]*MaJiangCard, 0)
	outRemovedCards = make([]*MaJiangCard, 0)
	for i, v := range cards {
		if int32(len(outRemovedCards)) >= wantRemovedAmount {
			result = append(result, cards[i:]...)
			break
		}

		if !v.IsEqualByTypeAndValue(cType, value) {
			result = append(result, v)
		} else {
			outRemovedCards = append(outRemovedCards, v)
		}
	}

	return
}

//在列表中查找指定的Card–
func FindCard(cards []*MaJiangCard, cType, val int32) *MaJiangCard {
	if cards == nil {
		logger.Error("FindCard:cards is nil.")
		return nil
	}

	for i, v := range cards {
		if v.IsEqualByTypeAndValue(cType, val) {
			return cards[i]
		}
	}

	return nil
}

//在列表中查找指定的Card
func FindCards(cards []*MaJiangCard, cType, val int32) []*MaJiangCard {
	if cards == nil {
		logger.Error("FindCards:cards is nil.")
		return nil
	}

	result := []*MaJiangCard{}
	for i, v := range cards {
		if v.IsEqualByTypeAndValue(cType, val) {
			result = append(result, cards[i])
		}
	}

	return result
}

// package majiangserver

// import (
// 	cmn "common"
// 	//"fmt"
// 	"logger"
// 	"math"
// 	"time"
// )

// const (
// 	ESuccess = iota
// 	ECardNull
// 	ETypeAmountMuch
// 	ECardFullSame
// )

// type HuController struct {
// 	patternGroups []*MaJiangPatternGroup
// 	//allpatternGroups    []*MaJiangPatternGroup
// 	originCards       []*MaJiangCard
// 	cards             []*MaJiangCard
// 	rmCardsAmountInfo *CardAmountStatistics //替换模式下的卡牌数量
// 	//notrmCardAmountInfo *CardAmountStatistics //非替换模式下的卡牌数量
// 	player *MaJiangPlayer
// }

// func NewHuController(p *MaJiangPlayer) *HuController {
// 	huC := &HuController{player: p}

// 	return huC
// }

// //初始化函数 ，调用完次函数后，就可以直接获取成员数据了
// func (self *HuController) UpdateData(cards []*MaJiangCard) {
// 	//check input param
// 	if cards == nil || len(cards) <= 0 {
// 		logger.Error("UpdateData:cards is nil.")
// 		return
// 	}

// 	//need check hu
// 	logger.Info("更新胡：")
// 	if eReason := self.needUpdate(cards); eReason != ESuccess {
// 		logger.Info("不需要更新:", eReason)
// 		//当以前是能够胡牌的，但是摸了一张不同花色的牌导致现在有不能胡了要把以前的胡的模式组给清除掉
// 		if eReason == ETypeAmountMuch {
// 			self.patternGroups = make([]*MaJiangPatternGroup, 0)
// 		}

// 		return
// 	}

// 	logger.Info("开始计算胡数")

// 	//cache card && statistics card amount
// 	self.originCards = make([]*MaJiangCard, len(cards))
// 	copy(self.originCards, cards)

// 	//self.notrmCardAmountInfo = NewCardAmountStatisticsByCards(cards, false)

// 	self.patternGroups = make([]*MaJiangPatternGroup, 0)
// 	//self.allpatternGroups = make([]*MaJiangPatternGroup, 0)

// 	// replace hongzhong value && statistics card amount
// 	player := self.player
// 	if player == nil {
// 		logger.Error("HuController.player is nil.")
// 		return
// 	}

// 	self.cards = CloneMaJiangCards(cards)
// 	hongZhongCards := GetSpecificTypeCardsByCardsList(self.cards, HongZhong, false)
// 	hongZhongAmount := len(hongZhongCards)
// 	hongZhongCardsInReplaceMode := GetSpecificTypeCardsByCardsList(self.cards, HongZhong, true)
// 	//没有红中不需要进行替换,或者红中是已经被替换掉的
// 	if hongZhongAmount <= 0 || len(hongZhongCardsInReplaceMode) <= 0 {
// 		//统计这一组替换的牌的数量信息
// 		self.rmCardsAmountInfo = NewCardAmountStatisticsByCards(self.cards, true)

// 		PrintCardsS("这一组替换后的手牌:", cards)

// 		curTime := time.Now()
// 		//生成特殊的胡牌模式组
// 		self.GenerateSpecificPatternGroup()

// 		//生成可以胡牌的模式组,将结果保存在成员中
// 		self.GeneratePatternGroup()

// 		logger.Info("一次完整的计算胡用时：", time.Now().Sub(curTime))
// 	} else {
// 		curTime := time.Now()
// 		canReplaceTypes := player.GetCanReplaceType()
// 		replaceList := GetReplaceList(int32(hongZhongAmount), canReplaceTypes)
// 		logger.Info("红中数量：%d, replaceListLen:%d, 能替换的类型：", hongZhongAmount, len(replaceList), canReplaceTypes, " 计算红中替换列表耗时：", time.Now().Sub(curTime))
// 		for _, replaceGroup := range replaceList {
// 			if replaceGroup == nil || len(replaceGroup) <= 0 {
// 				continue
// 			}

// 			if len(replaceGroup) != hongZhongAmount {
// 				logger.Error("替代牌的数量和红中的数量不匹配")
// 				continue
// 			}

// 			//将红中的牌复制一份，再进行替换
// 			tempCards, clonedHongZhong := CloneHongZhong(self.cards)
// 			self.cards = tempCards

// 			//设置一组替换值并计算胡
// 			for i, replace := range replaceGroup {
// 				hongZhong := clonedHongZhong[i]
// 				if hongZhong == nil {
// 					logger.Error("红中牌竟然是nil.")
// 					continue
// 				}

// 				hongZhong.SetHZReplaceValue(replace.rType, replace.rValue)
// 			}

// 			//统计这一组替换的牌的数量信息
// 			self.rmCardsAmountInfo = NewCardAmountStatisticsByCards(self.cards, true)

// 			PrintCardsS("这一组替换后的手牌:", cards)

// 			//生成特殊的胡牌模式组
// 			self.GenerateSpecificPatternGroup()

// 			//生成可以胡牌的模式组,将结果保存在成员中
// 			self.GeneratePatternGroup()
// 		}

// 		logger.Info("一次完整的计算胡用时：", time.Now().Sub(curTime))
// 		//打印最终的结果
// 		PrintPatternGroupsS("最终结果：", self.patternGroups, true)
// 	}

// }

// //检查是否需要重新算胡牌，牌没变化就用算了
// func (self *HuController) needUpdate(cards []*MaJiangCard) int32 {
// 	//check input param
// 	if cards == nil {
// 		logger.Error("needUpdate:cards is nil.")
// 		return ECardNull
// 	}

// 	//check type amount
// 	if self.player != nil {
// 		fixedTypeList := self.player.GetTypeInfoInShowPattern()
// 		amountInfo := NewCardAmountStatisticsByCards(cards, false)
// 		typeAmount := amountInfo.GetTypeAmount(false, fixedTypeList)
// 		if typeAmount > int32(2-len(fixedTypeList)) {
// 			return ETypeAmountMuch
// 		}
// 	}

// 	//check card amount is same
// 	if len(self.originCards) != 0 && len(self.originCards) != len(cards) {
// 		return ESuccess
// 	}

// 	//is same
// 	tempCards := make([]*MaJiangCard, len(cards))
// 	copy(tempCards, cards)

// 	for _, v := range self.originCards {
// 		cType, cVal := v.CurValue()
// 		tempCards = RemoveCardByType(tempCards, cType, cVal)
// 	}

// 	isSame := len(tempCards) <= 0
// 	if isSame {
// 		return ECardFullSame
// 	}

// 	return ESuccess
// }

// //产生模式组
// func (self *HuController) GeneratePatternGroup() {

// 	curTime := time.Now()
// 	//获取去重后的所有模式
// 	patterns := self.StatisticsAllPattern()
// 	logger.Info("获取所有的准模式:", len(patterns), "用时：", time.Now().Sub(curTime))
// 	PrintPatterns(patterns)

// 	//然后对这些模式进行组合，形成模式组（一套可胡牌的模式列表）
// 	curTime = time.Now()

// 	n := math.Min(float64(len(self.cards)/3), float64(len(patterns)))
// 	patternGroups := self.CalcPatternGroup(int(n), patterns)
// 	logger.Info("zuhe:", n, len(patterns))
// 	PrintPatterns(patterns)
// 	logger.Info("进行模式的组合:", len(patternGroups), "用时:", time.Now().Sub(curTime))
// 	PrintPatternGroups(patternGroups, false)

// 	//没有模式组可以生成时，检查手牌，是否只剩下1或2张牌了
// 	if patternGroups == nil || len(patternGroups) <= 0 {
// 		cardCount := len(self.cards)
// 		switch cardCount {
// 		case 0:
// 			fallthrough
// 		case 1:
// 			fallthrough
// 		case 2:
// 			patternGroups = []*MaJiangPatternGroup{NewPatternGroup([]*MaJiangPattern{})}
// 		default:
// 			// self.patternGroups = make([]*MaJiangPatternGroup, 0)
// 			// self.allpatternGroups = make([]*MaJiangPatternGroup, 0)
// 			return
// 		}
// 	}

// 	//计算出每种模式组中的单牌
// 	curTime = time.Now()
// 	singleCardInpatternGroups := self.GetSingleCardInPatternGroup(patternGroups)
// 	logger.Info("计算每组模式的单牌：", len(singleCardInpatternGroups), "用时：", time.Now().Sub(curTime))

// 	//最后通过每个模式组中的单牌来计算胡的牌
// 	curTime = time.Now()
// 	patternGroups = self.CalcHu(singleCardInpatternGroups, patternGroups)
// 	logger.Info("计算每组模式的胡：", "用时：", time.Now().Sub(curTime))

// 	//剔除不能胡的牌
// 	curTime = time.Now()
// 	patternGroups = self.StripNoHuPatternGroup(patternGroups)
// 	logger.Info("剔除不能胡的牌：", len(patternGroups), "用时：", time.Now().Sub(curTime))

// 	//生成模式组的ID
// 	curTime = time.Now()
// 	self.GeneratePatternGroupID(patternGroups)
// 	logger.Info("生成模式组的ID 用时：", time.Now().Sub(curTime))

// 	//剔除当前的重复模式组
// 	curTime = time.Now()
// 	patternGroups = self.StripSamePatternGroup(patternGroups)
// 	logger.Info("剔除当前的重复模式组 用时：", time.Now().Sub(curTime))
// 	PrintPatternGroups(patternGroups, false)

// 	//剔除重复的模式组
// 	curTime = time.Now()
// 	self.patternGroups = append(self.patternGroups, patternGroups...)
// 	//self.patternGroups = self.StripSamePatternGroup(self.patternGroups)

// 	logger.Info("剔除总的重复的模式组 用时：", time.Now().Sub(curTime))
// 	PrintPatternGroupsS("剔除总的重复的模式组 用时：", self.patternGroups, false)
// }

// //产生特殊模式组
// func (self *HuController) GenerateSpecificPatternGroup() {
// 	logger.Info("HuController.GenerateSpecificPatternGroup: 手牌数量：", len(self.cards))
// 	//小七对
// 	if len(self.cards) == 13 {
// 		tempCards := make([]*MaJiangCard, len(self.cards))
// 		copy(tempCards, self.cards)

// 		partternList := make([]*MaJiangPattern, 0)
// 		singleCards := make([]*MaJiangCard, 0)
// 		for len(tempCards) > 0 && len(singleCards) <= 1 {

// 			for _, c := range tempCards {
// 				cType, cVal := c.CurValue()

// 				tempCards = RemoveCardByType(tempCards, cType, cVal)
// 				findCard := FindCard(tempCards, cType, cVal)
// 				//没有此牌的对子，那么此牌为单牌
// 				if findCard == nil {
// 					singleCards = append(singleCards, c)
// 				} else {
// 					tempCards = RemoveCardByType(tempCards, cType, cVal)
// 					partternList = append(partternList, NewPattern(PTPair, []*MaJiangCard{c, findCard}, false))
// 				}

// 				break
// 			}

// 			PrintCardsS("HuController.GenerateSpecificPatternGroup:当前的手牌情况", tempCards)
// 			PrintCardsS("HuController.GenerateSpecificPatternGroup:当前的单牌情况", singleCards)

// 		}

// 		PrintCardsS("HuController.GenerateSpecificPatternGroup:最后单牌的情况：", singleCards)
// 		if len(singleCards) == 1 {
// 			patternGroup := NewPatternGroup(partternList)
// 			patternGroup.GenerateID()
// 			cType, cVal := singleCards[0].CurValue()

// 			patternGroup.kaoCards = append(patternGroup.kaoCards, singleCards[0])
// 			patternGroup.huCards = append(patternGroup.huCards, &MaJiangCard{value: cVal, cType: cType, flag: cmn.CUnknown})

// 			self.patternGroups = append(self.patternGroups, patternGroup)
// 		}
// 	}

// 	//2个或3个四张的牌
// 	fourCards := self.rmCardsAmountInfo.GetCardsBySpecificAmount(4, nil)
// 	if len(fourCards) >= 2 {
// 		tempCards := make([]*MaJiangCard, len(self.cards))
// 		copy(tempCards, self.cards)
// 		partternList := make([]*MaJiangPattern, len(fourCards))

// 		//统计四个的模式
// 		for i, c := range fourCards {
// 			cType, cVal := c.CurValue()
// 			var removedCards []*MaJiangCard = nil
// 			tempCards, removedCards = RemoveCardsByType(tempCards, cType, cVal, 4)
// 			partternList[i] = NewPattern(PTGang, removedCards, false)
// 		}

// 		//计算胡
// 		residueCardAmount := len(tempCards)
// 		if residueCardAmount < 3 && residueCardAmount > 0 {
// 			patternGroup := NewPatternGroup(partternList)
// 			patternGroup.GenerateID()
// 			calcHuPatterGroups := self.CalcHu([][]*MaJiangCard{tempCards}, []*MaJiangPatternGroup{patternGroup})
// 			calcHuPatterGroups = self.StripNoHuPatternGroup(calcHuPatterGroups)
// 			if calcHuPatterGroups != nil {
// 				self.patternGroups = append(self.patternGroups, calcHuPatterGroups...)
// 			}

// 		} else if residueCardAmount > 3 {
// 			tempHuController := NewHuController(self.player)
// 			tempHuController.UpdateData(tempCards)

// 			for _, p := range tempHuController.patternGroups {
// 				if p == nil {
// 					continue
// 				}

// 				p.patterns = append(p.patterns, partternList...)
// 				p.GenerateID()
// 			}

// 			self.patternGroups = append(self.patternGroups, tempHuController.patternGroups...)
// 		} else {
// 			logger.Error("没有这种胡牌，剩余牌的数量是:", residueCardAmount)
// 		}
// 	}
// }

// //统计出所有模式
// func (self *HuController) StatisticsAllPattern() (result []*MaJiangPattern) {
// 	result = make([]*MaJiangPattern, 0)

// 	for _, v := range self.cards {
// 		patterns := StatisticsPattern(self.cards, v)

// 		PrintCard(v)
// 		PrintPatterns(patterns)

// 		result = append(result, patterns...)
// 	}

// 	result = self.RemoveUselessPattern(result)

// 	return
// }

// //统计单张牌的所有模式
// func StatisticsPattern(cards []*MaJiangCard, card *MaJiangCard) []*MaJiangPattern {
// 	if cards == nil {
// 		logger.Error("StatisticsPattern:cards is nil.")
// 		return nil
// 	}

// 	if card == nil {
// 		logger.Error("card is nil.")
// 		return nil
// 	}

// 	result := make([]*MaJiangPattern, 0)

// 	tempCards := make([]*MaJiangCard, 0)
// 	tempCards = append(tempCards, cards...)
// 	cType, cVal := card.CurValue()
// 	tempCards = RemoveCardByType(tempCards, cType, cVal)

// 	pattern := StatisticsAAAA(tempCards, card)
// 	if pattern != nil {
// 		result = append(result, pattern)
// 	}

// 	pattern = StatisticsAAA(tempCards, card)
// 	if pattern != nil {
// 		result = append(result, pattern)
// 	}

// 	pattern = StatisticsAA(tempCards, card)
// 	if pattern != nil {
// 		result = append(result, pattern)
// 	}

// 	pattern = StatisticsSZ(tempCards, card)
// 	if pattern != nil {
// 		result = append(result, pattern)
// 	}

// 	return result
// }

// //统计AAAA
// func StatisticsAAAA(cards []*MaJiangCard, card *MaJiangCard) *MaJiangPattern {
// 	if cards == nil {
// 		logger.Error("StatisticsEQS:cards is nil.")
// 		return nil
// 	}

// 	if card == nil {
// 		logger.Error("card is nil.")
// 		return nil
// 	}

// 	cType, cVal := card.CurValue()
// 	findCards := FindCards(cards, cType, cVal)
// 	if len(findCards) >= 3 {
// 		return NewPattern(PTGang, []*MaJiangCard{findCards[0], findCards[1], findCards[2], card}, false)
// 	}

// 	return nil
// }

// //统计AAA
// func StatisticsAAA(cards []*MaJiangCard, card *MaJiangCard) *MaJiangPattern {
// 	if cards == nil {
// 		logger.Error("StatisticsEQS:cards is nil.")
// 		return nil
// 	}

// 	if card == nil {
// 		logger.Error("card is nil.")
// 		return nil
// 	}

// 	cType, cVal := card.CurValue()
// 	findCards := FindCards(cards, cType, cVal)
// 	if len(findCards) >= 2 {
// 		return NewPattern(PTKan, []*MaJiangCard{findCards[0], findCards[1], card}, false)
// 	}

// 	return nil
// }

// //统计AA
// func StatisticsAA(cards []*MaJiangCard, card *MaJiangCard) *MaJiangPattern {
// 	if cards == nil {
// 		logger.Error("StatisticsAA:cards is nil.")
// 		return nil
// 	}

// 	if card == nil {
// 		logger.Error("card is nil.")
// 		return nil
// 	}

// 	cType, cVal := card.CurValue()
// 	findCard := FindCard(cards, cType, cVal)
// 	if findCard == nil {
// 		return nil
// 	}

// 	return NewPattern(PTPair, []*MaJiangCard{card, findCard}, false)
// }

// //统计顺子
// func StatisticsSZ(cards []*MaJiangCard, card *MaJiangCard) *MaJiangPattern {
// 	if cards == nil {
// 		logger.Error("StatisticsSZ:cards is nil.")
// 		return nil
// 	}

// 	if card == nil {
// 		logger.Error("card is nil.")
// 		return nil
// 	}

// 	cType, val := card.CurValue()
// 	curNum := val - 1
// 	if curNum > 0 {
// 		findCard := FindCard(cards, cType, curNum)
// 		curNum--
// 		if findCard != nil && curNum > 0 {
// 			secondFindCard := FindCard(cards, cType, curNum)
// 			if secondFindCard != nil {
// 				return NewPattern(PTSZ, []*MaJiangCard{card, findCard, secondFindCard}, false)
// 			}
// 		}
// 	}

// 	return nil
// }

// //移除多余的模式
// func (self *HuController) RemoveUselessPattern(patterns []*MaJiangPattern) (result []*MaJiangPattern) {

// 	result = append(result, patterns...)

// 	patternAmount := StatisticsPatternAmount(patterns)

// 	for _, v := range patternAmount {
// 		minAmount := int32(255)
// 		if v[0].ptype == PTPair {
// 			minAmount = 1
// 		} else {
// 			for _, card := range v[0].cards {
// 				cType, val := card.CurValue()
// 				curCardAmount := self.rmCardsAmountInfo.GetCardAmount(cType, val)
// 				if minAmount > curCardAmount {
// 					minAmount = curCardAmount
// 				}

// 			}

// 		}

// 		for i := minAmount; i < int32(len(v)); i++ {
// 			for k, r := range result {
// 				if r.id == v[0].id {
// 					result = append(result[:k], result[k+1:]...)
// 					break
// 				}
// 			}
// 		}
// 	}

// 	return
// }

// //统计相同模式的数量
// func StatisticsPatternAmount(patterns []*MaJiangPattern) (patternAmount map[int32][]*MaJiangPattern) {
// 	if patterns == nil {
// 		logger.Error("pattern is nil.")
// 		return nil
// 	}

// 	patternAmount = make(map[int32][]*MaJiangPattern, 0)

// 	if patterns == nil {
// 		return patternAmount
// 	}

// 	for _, v := range patterns {
// 		if patternAmount[v.id] == nil {
// 			patternAmount[v.id] = make([]*MaJiangPattern, 0)
// 		}

// 		patternAmount[v.id] = append(patternAmount[v.id], v)
// 	}
// 	return
// }

// //从一个切片中移除指定类型的Card
// func RemoveCardByType(cards []*MaJiangCard, cType, value int32) []*MaJiangCard {
// 	if cards == nil {
// 		logger.Error("RemoveCardByType:cards is nil.")
// 		return nil
// 	}

// 	for i, v := range cards {
// 		curCType, curVal := v.CurValue()
// 		if curVal == value && curCType == cType {
// 			cards = append(cards[:i], cards[i+1:]...)
// 			break
// 		}
// 	}

// 	return cards
// }

// //从一个切片中移除指定类型的多个Card
// func RemoveCardsByType(cards []*MaJiangCard, cType, value, wantRemovedAmount int32) (result []*MaJiangCard, outRemovedCards []*MaJiangCard) {

// 	if cards == nil {
// 		logger.Error("RemoveCardsByType:cards is nil.")
// 		return
// 	}

// 	//removedAmount := 0
// 	result = make([]*MaJiangCard, 0)
// 	outRemovedCards = make([]*MaJiangCard, 0)
// 	for i, v := range cards {
// 		if int32(len(outRemovedCards)) >= wantRemovedAmount {
// 			result = append(result, cards[i:]...)
// 			break
// 		}

// 		if !v.IsEqualByTypeAndValue(cType, value) {
// 			result = append(result, v)
// 		} else {
// 			outRemovedCards = append(outRemovedCards, v)
// 		}
// 	}

// 	return
// }

// //在列表中查找指定的Card–
// func FindCard(cards []*MaJiangCard, cType, val int32) *MaJiangCard {
// 	if cards == nil {
// 		logger.Error("FindCard:cards is nil.")
// 		return nil
// 	}

// 	for i, v := range cards {
// 		if v.IsEqualByTypeAndValue(cType, val) {
// 			return cards[i]
// 		}
// 	}

// 	return nil
// }

// //在列表中查找指定的Card
// func FindCards(cards []*MaJiangCard, cType, val int32) []*MaJiangCard {
// 	if cards == nil {
// 		logger.Error("FindCards:cards is nil.")
// 		return nil
// 	}

// 	result := []*MaJiangCard{}
// 	for i, v := range cards {
// 		if v.IsEqualByTypeAndValue(cType, val) {
// 			result = append(result, cards[i])
// 		}
// 	}

// 	return result
// }

// //测试性能统计时间用的
// var cloneCardTime float64 = 0
// var checkTime float64 = 0
// var appendToContainerTime float64 = 0
// var stripSameTime float64 = 0
// var stripSameCompareTime float64 = 0

// //计算所有的模式组
// func (self *HuController) CalcPatternGroup(n int, patterns []*MaJiangPattern) (patternGroups []*MaJiangPatternGroup) {
// 	//patternGroups = make([]*MaJiangPatternGroup, 0)
// 	tempPatternGroups := make([]*MaJiangPatternGroup, 0)

// 	if n <= 0 || len(patterns) <= 0 {
// 		return
// 	}

// 	if n > len(patterns) {
// 		logger.Error("n must be less than patterns's length. N:(%s) Patterns_Length:%s", n, len(patterns))
// 		PrintCards(self.cards)
// 		return
// 	}

// 	curTime := time.Now()
// 	result := C(n, len(patterns))
// 	//fmt.Println("生成排列组合数 用时：", time.Now().Sub(curTime).Seconds())

// 	//fmt.Println("模式组合数：", n, len(patterns), len(result))
// 	//	for k, v := range tempSmallCardCount {
// 	//		fmt.Print(k, v)
// 	//		fmt.Print(", ")
// 	//	}
// 	//	fmt.Println("=========")
// 	//	for k, v := range tempBigCardCount {
// 	//		fmt.Print(k, v)
// 	//		fmt.Print(", ")
// 	//	}

// 	curCardCountForPatternList := int32(len(self.cards) - 3)

// 	for i := 0; i < len(result); i++ {
// 		curTime = time.Now()

// 		tempCardsCount := self.rmCardsAmountInfo.GetAmountInfo()

// 		cloneCardTime += time.Now().Sub(curTime).Seconds()

// 		//		fmt.Println("小：", self.smallCardAmount)
// 		//		fmt.Println("大：", self.bigCardAmount)

// 		isFailure := false

// 		tempPatternList := make([]*MaJiangPattern, 0)

// 		for j := 0; j < len(result[i]); j++ {

// 			curTime = time.Now()

// 			index := result[i][j]
// 			pattern := patterns[index]
// 			if pattern.ptype == PTPair {
// 				for k := 0; k < len(tempPatternList); k++ {
// 					if tempPatternList[k].ptype == PTPair {
// 						isFailure = true
// 						break
// 					}
// 				}
// 			}

// 			if isFailure {
// 				break
// 			}

// 			for k := 0; k < len(pattern.cards); k++ {
// 				cType, cVal := pattern.cards[k].CurValue()

// 				if tempCardsCount[cType][cVal] <= 0 {
// 					isFailure = true
// 					break
// 				}

// 				tempCardsCount[cType][cVal]--
// 			}

// 			if isFailure {
// 				break
// 			}

// 			checkTime += time.Now().Sub(curTime).Seconds()

// 			curTime = time.Now()
// 			tempPatternList = append(tempPatternList, pattern)
// 			if GetCardCountByPatternList(tempPatternList) > curCardCountForPatternList {
// 				break
// 			}

// 			appendToContainerTime += time.Now().Sub(curTime).Seconds()
// 		}

// 		if !isFailure {
// 			//fmt.Println("生成的模式组：")
// 			//PrintPatterns(tempPatternList)
// 			curTime = time.Now()
// 			patternGroup := NewPatternGroup(tempPatternList)
// 			tempPatternGroups = append(tempPatternGroups, patternGroup)
// 			appendToContainerTime += time.Now().Sub(curTime).Seconds()
// 		}

// 		// fmt.Println("生成的模式组：")
// 		// PrintPatternGroups(tempPatternGroups)
// 		curTimeS := time.Now()
// 		//fmt.Println("可行的模式：", len(tempPatternGroups))
// 		//		for i := 0; i < len(tempPatternGroups); i++ {
// 		//			isExist := false

// 		//			curTime = time.Now()
// 		//			for j := 0; j < len(patternGroups); j++ {

// 		//				if tempPatternGroups[i].IsEqual(patternGroups[j]) {
// 		//					isExist = true
// 		//					break
// 		//				}
// 		//			}

// 		//			stripSameCompareTime += time.Now().Sub(curTime).Seconds()

// 		//			if !isExist {
// 		//				patternGroups = append(patternGroups, tempPatternGroups[i])
// 		//			}
// 		//		}

// 		stripSameTime += time.Now().Sub(curTimeS).Seconds()

// 		// fmt.Println("剔除重复后的模式组：")
// 		// PrintPatternGroups(patternGroups)

// 	}

// 	patternGroups = tempPatternGroups
// 	//patternGroups = append(patternGroups, tempPatternGroups...)
// 	//	fmt.Println("生成后的模式组：", len(tempPatternGroups))
// 	//	PrintPatternGroups(tempPatternGroups, false)
// 	//	fmt.Println("生成后的模式组：==============")

// 	//fmt.Println("克隆的时间：", cloneCardTime)
// 	//fmt.Println("检测的时间：", checkTime)
// 	//fmt.Println("追加的时间：", appendToContainerTime)
// 	//fmt.Println("剔除重复的时间：", stripSameTime)
// 	//fmt.Println("剔除重复的比较的时间：", stripSameCompareTime)

// 	return
// }

// //生成模式的组合下标
// func C(n int, m int) (result [][]int) {
// 	if n > m {
// 		logger.Error("n must be less than m", n, m)
// 		return nil
// 	}

// 	//	size := Factorial(m) / (Factorial(m-n) * Factorial(n))
// 	//	fmt.Println("组合数量：", n, m, size)
// 	result = make([][]int, 0)

// 	index := make([]int, m)
// 	//fmt.Println(m, index)
// 	for i := 0; i < m; i++ {
// 		index[i] = 0
// 	}

// 	for i := 0; i < n; i++ {
// 		index[i] = 1
// 	}

// 	result = append(result, GetC(index))

// 	for true {
// 		for i := 0; i < m-1; i++ {
// 			if index[i] == 1 && index[i+1] == 0 {
// 				oneIndex := 0
// 				for j := 0; j < i; j++ {
// 					if index[j] == 1 {
// 						index[j] = 0
// 						index[oneIndex] = 1
// 						oneIndex++
// 					}
// 				}

// 				index[i] = 0
// 				index[i+1] = 1

// 				result = append(result, GetC(index))
// 				break
// 			}
// 		}

// 		//check is end
// 		isEnd := true
// 		for k := m - n; k < m; k++ {
// 			if index[k] != 1 {
// 				isEnd = false
// 				break
// 			}
// 		}

// 		if isEnd {
// 			break
// 		}
// 	}

// 	return
// }

// func GetC(index []int) (result []int) {
// 	result = make([]int, 0)
// 	for i := 0; i < len(index); i++ {
// 		if index[i] == 1 {
// 			result = append(result, i)
// 		}
// 	}

// 	return
// }

// //获取模式列表卡牌的数量
// func GetCardCountByPatternList(patternList []*MaJiangPattern) (result int32) {

// 	for _, p := range patternList {
// 		if p == nil {
// 			continue
// 		}

// 		result += int32(len(p.cards))
// 	}

// 	return result
// }

// //获取每个组模式中的单牌
// func (self *HuController) GetSingleCardInPatternGroup(patternGroup []*MaJiangPatternGroup) (result [][]*MaJiangCard) {
// 	result = make([][]*MaJiangCard, 0)
// 	for _, v := range patternGroup {
// 		tempCards := []*MaJiangCard{}
// 		tempCards = append(tempCards, self.cards...)
// 		for i := 0; i < len(v.patterns); i++ {
// 			for j := 0; j < len(v.patterns[i].cards); j++ {
// 				card := v.patterns[i].cards[j]
// 				cType, cVal := card.CurValue()
// 				tempCards = RemoveCardByType(tempCards, cType, cVal)
// 			}
// 		}
// 		result = append(result, tempCards)
// 	}
// 	return
// }

// //计算每个组模式中胡的牌
// func (self *HuController) CalcHu(singleCards [][]*MaJiangCard, patternGroups []*MaJiangPatternGroup) []*MaJiangPatternGroup {
// 	for i, v := range singleCards {

// 		singleCardCount := len(v)
// 		pairCards := patternGroups[i].GetPairCard()
// 		pairCount := len(pairCards)
// 		if singleCardCount >= 3 || singleCardCount <= 0 {
// 			continue
// 		}

// 		//logger.Info("当前的模式组：")
// 		//PrintPatternGroup(patternGroups[i], false)
// 		//logger.Info("一个模式组中的单牌数%s, 对子数：%s", singleCardCount, pairCount)
// 		//logger.Info("单张牌：")
// 		//PrintCards(v)
// 		//logger.Info("对子：")
// 		//PrintCards(pairCards)

// 		result := []*MaJiangCard{}
// 		if singleCardCount == 1 {
// 			if pairCount == 0 {
// 				curType, curVal := v[0].CurValue()
// 				result = []*MaJiangCard{&MaJiangCard{value: curVal, cType: curType, flag: cmn.CUnknown}}
// 			}
// 		} else if singleCardCount == 2 {
// 			firstCard := v[0]
// 			secondCard := v[1]
// 			firstCType, _ := firstCard.CurValue()
// 			secondCType, secondVal := secondCard.CurValue()

// 			//AA mode
// 			if firstCard.IsEqualByTypeAndValue(secondCType, secondVal) {
// 				if pairCount == 1 {
// 					curType, curVal := pairCards[0].CurValue()
// 					result = []*MaJiangCard{&MaJiangCard{value: firstCard.value, cType: firstCType, flag: cmn.CUnknown},
// 						&MaJiangCard{value: curVal, cType: curType, flag: cmn.CUnknown}}
// 				} else if pairCount == 0 {
// 					result = []*MaJiangCard{&MaJiangCard{value: firstCard.value, cType: firstCType, flag: cmn.CUnknown}}
// 				} else {
// 					logger.Error("不可能存在此种情况")
// 				}
// 				//SZ
// 			} else if firstCType == secondCType {
// 				offset := int(secondCard.value) - int(firstCard.value)

// 				switch offset {
// 				case 1:
// 					result = []*MaJiangCard{}
// 					if firstCard.value-1 > 0 {
// 						result = append(result, &MaJiangCard{value: firstCard.value - 1, cType: firstCType, flag: cmn.CUnknown})
// 					}
// 					if secondCard.value+1 < 10 {
// 						result = append(result, &MaJiangCard{value: secondCard.value + 1, cType: secondCType, flag: cmn.CUnknown})
// 					}
// 				case -1:
// 					result = []*MaJiangCard{}
// 					if secondCard.value-1 > 0 {
// 						result = append(result, &MaJiangCard{value: secondCard.value - 1, cType: secondCType, flag: cmn.CUnknown})
// 					}
// 					if firstCard.value+1 < 10 {
// 						result = append(result, &MaJiangCard{value: firstCard.value + 1, cType: firstCType, flag: cmn.CUnknown})
// 					}
// 				case 2, -2:
// 					result = []*MaJiangCard{&MaJiangCard{value: (firstCard.value + secondCard.value) / 2, cType: firstCType, flag: cmn.CUnknown}}

// 				}
// 			}

// 		}

// 		//检查是否有胡
// 		if result != nil && len(result) > 0 {
// 			logger.Info("特殊情况下的胡牌：")
// 			PrintCards(result)

// 			patternGroups[i].kaoCards = append(patternGroups[i].kaoCards, singleCards[i]...)
// 			patternGroups[i].huCards = append(patternGroups[i].huCards, result...)
// 		}
// 	}

// 	return patternGroups
// }

// //剔除不能胡的组模式
// func (self *HuController) StripNoHuPatternGroup(groupPatterns []*MaJiangPatternGroup) (result []*MaJiangPatternGroup) {
// 	//剔除不能胡的牌
// 	for _, v := range groupPatterns {
// 		if v.CanHu() {
// 			result = append(result, v)
// 		}
// 	}

// 	return result
// }

// //通过胡数和胡的牌产生唯一ID
// func (self *HuController) GeneratePatternGroupID(groupPatterns []*MaJiangPatternGroup) {

// 	for _, v := range groupPatterns {
// 		v.GenerateID()
// 	}
// }

// //剔除相同的模式组
// func (self *HuController) StripSamePatternGroup(groupPatterns []*MaJiangPatternGroup) (result []*MaJiangPatternGroup) {
// 	result = make([]*MaJiangPatternGroup, 0)

// 	//剔除重复的牌
// 	isExist := false
// 	for _, v := range groupPatterns {
// 		isExist = false
// 		for _, rv := range result {
// 			if v.id == rv.id {
// 				isExist = true
// 			}
// 		}

// 		if !isExist {
// 			result = append(result, v)
// 		}
// 	}

// 	return
// }

// //复制牌
// func CloneMaJiangCards(src []*MaJiangCard) (dst []*MaJiangCard) {
// 	if src == nil {
// 		return nil
// 	}

// 	dst = make([]*MaJiangCard, len(src))
// 	for i, v := range src {
// 		dst[i] = CloneMaJiangCard(v)
// 	}

// 	return
// }

// //复制红中,防止每次都是替代的同一组红中
// func CloneHongZhong(src []*MaJiangCard) (dst []*MaJiangCard, hongCards []*MaJiangCard) {
// 	dst = make([]*MaJiangCard, len(src))
// 	hongCards = make([]*MaJiangCard, 0)

// 	if src == nil {
// 		return
// 	}

// 	for i, v := range src {
// 		if v.IsHongZhong() {
// 			temp := CloneMaJiangCard(v)
// 			dst[i] = temp
// 			hongCards = append(hongCards, temp)
// 		} else {
// 			dst[i] = v
// 		}
// 	}
// 	return
// }

// func CloneMaJiangCard(src *MaJiangCard) (dst *MaJiangCard) {
// 	if src == nil {
// 		return nil
// 	}

// 	dst = &MaJiangCard{}

// 	dst.id = src.id
// 	dst.value = src.value
// 	dst.cType = src.cType
// 	dst.rcType = src.rcType
// 	dst.flag = src.flag
// 	dst.owner = src.owner

// 	return dst
// }

// //获取指定类型的牌
// func GetSpecificTypeCardsByCardsList(cardsList []*MaJiangCard, cType int32, replaceMode bool) (result []*MaJiangCard) {
// 	result = make([]*MaJiangCard, 0)

// 	if cardsList == nil || len(cardsList) <= 0 {
// 		return
// 	}

// 	if !replaceMode {
// 		for _, card := range cardsList {
// 			if card.cType == cType {
// 				result = append(result, card)
// 			}
// 		}
// 	} else {
// 		for _, card := range cardsList {
// 			tempCType, _ := card.CurValue()
// 			if tempCType == cType {
// 				result = append(result, card)
// 			}
// 		}
// 	}

// 	return result
// }

// type HongZhongReplaceInfo struct {
// 	id     int32
// 	rType  int32
// 	rValue int32
// }

// func GetReplaceList(hongZhongAmount int32, canReplaceTypeList []int32) (result [][]*HongZhongReplaceInfo) {
// 	//可以替换的牌的数量
// 	canReplaceTypeAmount := len(canReplaceTypeList)
// 	canReplaceAmount := hongZhongAmount * int32(canReplaceTypeAmount) * 9

// 	//生成可以替换的牌的列表
// 	hongZhongReplaceList := make([]*HongZhongReplaceInfo, canReplaceAmount)
// 	for h := 0; h < int(hongZhongAmount); h++ {
// 		for typeIndex, rType := range canReplaceTypeList {
// 			for i := 0; i < 9; i++ {
// 				hongZhongReplaceList[h*(canReplaceTypeAmount*9)+typeIndex*9+i] = &HongZhongReplaceInfo{id: rType*10 + int32(i+1), rType: rType, rValue: int32(i + 1)}
// 			}
// 		}
// 	}

// 	//对列表进行组合
// 	result = make([][]*HongZhongReplaceInfo, 0)
// 	cIndexList := C(int(hongZhongAmount), int(canReplaceAmount))
// 	logger.Info("替换的组合列表：", cIndexList)
// 	for _, indexGroup := range cIndexList {
// 		tempReplaceGroup := make([]*HongZhongReplaceInfo, len(indexGroup))
// 		for index, tempReplaceIndex := range indexGroup {
// 			tempReplaceGroup[index] = hongZhongReplaceList[tempReplaceIndex]
// 		}

// 		if !ExistSameReplace(result, tempReplaceGroup) {
// 			result = append(result, tempReplaceGroup)
// 		}
// 	}

// 	return
// }

// func ExistSameReplace(replaceList [][]*HongZhongReplaceInfo, replaceInfo []*HongZhongReplaceInfo) bool {
// 	for _, replaceGroup := range replaceList {
// 		tempReplaceInfo := make([]*HongZhongReplaceInfo, len(replaceInfo))
// 		copy(tempReplaceInfo, replaceInfo)

// 		for _, replace := range replaceGroup {
// 			tempReplaceInfo = RemoveReplaceCard(tempReplaceInfo, replace.id)
// 		}

// 		isSame := len(tempReplaceInfo) <= 0
// 		if isSame {
// 			return true
// 		}
// 	}

// 	return false
// }

// func RemoveReplaceCard(replaceList []*HongZhongReplaceInfo, id int32) []*HongZhongReplaceInfo {
// 	for i, val := range replaceList {
// 		if val.id == id {
// 			return append(replaceList[:i], replaceList[i+1:]...)
// 		}
// 	}

// 	return replaceList
// }
