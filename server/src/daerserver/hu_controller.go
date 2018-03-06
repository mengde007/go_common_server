package daerserver

import (
	cmn "common"
	"fmt"
	"logger"
	"math"
	"time"
)

type HuController struct {
	patternGroups    []*DaerPatternGroup
	allpatternGroups []*DaerPatternGroup
	cards            []*DaerCard
	smallCardAmount  [10]int32
	bigCardAmount    [10]int32
	player           *DaerPlayer
}

func NewHuController(p *DaerPlayer) *HuController {
	huC := &HuController{player: p}
	return huC
}

//初始化函数 ，调用完次函数后，就可以直接获取成员数据了
func (controller *HuController) UpdateData(cards []*DaerCard) {
	if cards == nil || len(cards) <= 0 {
		logger.Error("UpdateData:cards is nil.")
		return
	}

	logger.Info("更新胡：")
	if !controller.needUpdate(cards) {
		logger.Info("不需要更新")
		return
	}

	logger.Info("开始计算胡数")

	//保存牌
	controller.cards = cards

	//统计出牌的数量
	controller.smallCardAmount, controller.bigCardAmount = StatisticsCardAmount(controller.cards)

	fmt.Println("牌列表：")
	// PrintCards(controller.cards)
	//生成可以胡牌的模式组,将结果保存在成员中
	controller.GeneratePatternGroup()
}

//检查是否需要重新算胡牌，牌没变化就用算了
func (controller *HuController) needUpdate(cards []*DaerCard) bool {
	if cards == nil {
		logger.Error("needUpdate:cards is nil.")
		return false
	}

	if len(controller.cards) != len(cards) {
		return true
	}

	tempCards := make([]*DaerCard, len(cards))
	copy(tempCards, cards)

	for _, v := range controller.cards {
		tempCards = RemoveCardByType(tempCards, v.value, v.big)
	}

	return len(tempCards) > 0
}

//产生模式组
func (controller *HuController) GeneratePatternGroup() {

	curTime := time.Now()
	//获取去重后的所有模式
	patterns := controller.StatisticsAllPattern()
	PrintPatternsS(fmt.Sprintf("获取所有的准模式：%d 用时：%d", len(patterns), time.Now().Sub(curTime)), patterns)

	//然后对这些模式进行组合，形成模式组（一套可胡牌的模式列表）
	curTime = time.Now()

	n := math.Min(float64(len(controller.cards)/3), float64(len(patterns)))
	patternGroups := controller.CalcPatternGroup(int(n), patterns)
	logger.Info("zuhe:", n, len(patterns))
	//PrintPatterns(patterns)

	PrintPatternGroupsS(fmt.Sprintf("进行模式的组合：%d  用时：%d", len(patternGroups), time.Now().Sub(curTime)), patternGroups, false)

	//没有模式组可以生成时，检查手牌，是否只剩下1或2张牌了
	if patternGroups == nil || len(patternGroups) <= 0 {
		cardCount := len(controller.cards)
		switch cardCount {
		case 0:
			fallthrough
		case 1:
			fallthrough
		case 2:
			patternGroups = []*DaerPatternGroup{NewPatternGroup([]*DaerPattern{})}
		default:
			controller.patternGroups = make([]*DaerPatternGroup, 0)
			controller.allpatternGroups = make([]*DaerPatternGroup, 0)
			return
		}
	}

	//计算出每种模式组中的单牌
	curTime = time.Now()
	singleCardInpatternGroups := controller.GetSingleCardInPatternGroup(patternGroups)
	logger.Info("计算每组模式的单牌：", len(singleCardInpatternGroups), "用时：", time.Now().Sub(curTime))
	//	for _, v := range singleCardInpatternGroups {
	//		PrintCards(v)
	//		fmt.Println("")
	//	}

	//最后通过每个模式组中的单牌来计算胡的牌
	curTime = time.Now()
	patternGroups = controller.CalcHu(singleCardInpatternGroups, patternGroups)
	logger.Info("计算每组模式的胡：", "用时：", time.Now().Sub(curTime))

	//缓存所有的模式-用于三拢或四坎的胡子的计算
	controller.allpatternGroups = make([]*DaerPatternGroup, len(patternGroups))
	copy(controller.allpatternGroups, patternGroups)

	//剔除不能胡的牌
	curTime = time.Now()
	patternGroups = controller.StripNoHuPatternGroup(patternGroups)
	logger.Info("剔除不能胡的牌：", len(patternGroups), "用时：", time.Now().Sub(curTime))

	// //生成模式组的ID
	// curTime = time.Now()
	// controller.GeneratePatternGroupID(patternGroups)
	// logger.Info("生成模式组的ID 用时：", time.Now().Sub(curTime))

	// //剔除重复的模式组
	// curTime = time.Now()
	// controller.patternGroups = controller.StripSamePatternGroup(patternGroups)
	// logger.Info("剔除重复的模式组 用时：", time.Now().Sub(curTime))
	// PrintPatternGroups(controller.patternGroups, false)

	controller.patternGroups = patternGroups
	logger.Info("最终结果：", len(controller.patternGroups))
	// PrintPatternGroups(controller.patternGroups, true)
}

//统计出所有模式
func (controller *HuController) StatisticsAllPattern() (result []*DaerPattern) {
	result = make([]*DaerPattern, 0)

	//	fmt.Println("Cards：")
	//	PrintCards(controller.cards)
	for _, v := range controller.cards {
		patterns := StatisticsPattern(controller.cards, v)
		//		fmt.Println(v.value, "的组合")
		//		PrintPatterns(patterns)

		//		fmt.Println("Cards：")
		//		PrintCards(controller.cards)

		result = append(result, patterns...)
	}

	result = controller.RemoveUselessPattern(result)

	return
}

//统计单张牌的所有模式
func StatisticsPattern(cards []*DaerCard, card *DaerCard) []*DaerPattern {
	if cards == nil {
		logger.Error("StatisticsPattern:cards is nil.")
		return nil
	}

	if card == nil {
		logger.Error("card is nil.")
		return nil
	}

	result := make([]*DaerPattern, 0)

	//	fmt.Println("-=============")
	//	PrintCards(cards)

	tempCards := make([]*DaerCard, 0)
	tempCards = append(tempCards, cards...)

	//	PrintCards(tempCards)
	//	fmt.Println(tempCards[0])

	//tempCards = RemoveCardByID(tempCards, card.id)

	tempCards = RemoveCardByType(tempCards, card.value, card.big)

	//	if card.value == 7 && card.big == true {
	//		fmt.Println("输出大七的组合：")
	//		PrintCards(tempCards)
	//	}

	pattern := StatisticsEQS(tempCards, card)
	if pattern != nil {
		result = append(result, pattern)
	}

	pattern = StatisticsSZ(tempCards, card)
	if pattern != nil {
		result = append(result, pattern)
	}

	pattern = StatisticsAA(tempCards, card)
	if pattern != nil {
		result = append(result, pattern)
	}

	pattern = StatisticsAAB(tempCards, card)
	if pattern != nil {
		result = append(result, pattern)
	}

	return result
}

//统计2，7，10
func StatisticsEQS(cards []*DaerCard, card *DaerCard) *DaerPattern {
	if cards == nil {
		logger.Error("StatisticsEQS:cards is nil.")
		return nil
	}

	if card == nil {
		logger.Error("card is nil.")
		return nil
	}

	if !card.IsRed() {
		return nil
	}

	if card.value == 2 {
		findCard := FindCard(cards, 7, card.big)
		if findCard != nil {
			secondFindCard := FindCard(cards, 10, card.big)
			if secondFindCard != nil {
				return NewPattern(PTEQSColumn, []*DaerCard{card, findCard, secondFindCard})
			}
		}
	} else if card.value == 7 {
		findCard := FindCard(cards, 2, card.big)
		if findCard != nil {
			secondFindCard := FindCard(cards, 10, card.big)
			if secondFindCard != nil {
				return NewPattern(PTEQSColumn, []*DaerCard{card, findCard, secondFindCard})
			}
		}
	} else if card.value == 10 {
		findCard := FindCard(cards, 2, card.big)
		if findCard != nil {
			secondFindCard := FindCard(cards, 7, card.big)
			if secondFindCard != nil {
				return NewPattern(PTEQSColumn, []*DaerCard{card, findCard, secondFindCard})
			}
		}
	} else {
		logger.Error("only 2, 7, 10")
	}

	return nil
}

//统计顺子
func StatisticsSZ(cards []*DaerCard, card *DaerCard) *DaerPattern {
	if cards == nil {
		logger.Error("StatisticsSZ:cards is nil.")
		return nil
	}

	if card == nil {
		logger.Error("card is nil.")
		return nil
	}

	curNum := card.value - 1
	if curNum > 0 {
		findCard := FindCard(cards, curNum, card.big)
		curNum--
		if findCard != nil && curNum > 0 {
			secondFindCard := FindCard(cards, curNum, card.big)
			if secondFindCard != nil {
				cardsList := []*DaerCard{card, findCard, secondFindCard}
				if IsOneTwoThree(cardsList) {
					return NewPattern(PTOneTwoThree, cardsList)
				} else {
					return NewPattern(PTSZColumn, cardsList)
				}
			}
		}
	}

	return nil
}

//检测是否1,2,3
func IsOneTwoThree(cards []*DaerCard) bool {
	if cards == nil || len(cards) != 3 {
		return false
	}

	var minValue int32 = 100
	for _, card := range cards {
		if minValue > card.value {
			minValue = card.value
		}
	}

	return minValue <= 1
}

//统计AA
func StatisticsAA(cards []*DaerCard, card *DaerCard) *DaerPattern {
	if cards == nil {
		logger.Error("StatisticsAA:cards is nil.")
		return nil
	}

	if card == nil {
		logger.Error("card is nil.")
		return nil
	}

	findCard := FindCard(cards, card.value, card.big)
	if findCard == nil {
		return nil
	}

	return NewPattern(PTPair, []*DaerCard{card, findCard})
}

//统计AAB
func StatisticsAAB(cards []*DaerCard, card *DaerCard) *DaerPattern {
	if cards == nil {
		logger.Error("StatisticsAAB:cards is nil.")
		return nil
	}

	if card == nil {
		logger.Error("card is nil.")
		return nil
	}

	findCard := FindCard(cards, card.value, card.big)
	if findCard != nil {
		secondFindCard := FindCard(cards, card.value, !card.big)
		if secondFindCard != nil {
			return NewPattern(PTAABColumn, []*DaerCard{card, findCard, secondFindCard})
		}
	} else {
		findCards := FindCards(cards, card.value, !card.big)
		if findCards != nil && len(findCards) == 2 {
			return NewPattern(PTAABColumn, append(findCards, card))
		}
	}

	return nil
}

//统计AAB
func StatisticsAABs(cards []*DaerCard, card *DaerCard) (result []*DaerPattern) {
	result = make([]*DaerPattern, 0)
	if cards == nil {
		logger.Error("StatisticsAAB:cards is nil.")
		return result
	}

	if card == nil {
		logger.Error("card is nil.")
		return result
	}

	findCard := FindCard(cards, card.value, card.big)
	if findCard != nil {
		findCards := FindCards(cards, card.value, !card.big)
		if findCards != nil && len(findCards) >= 1 {
			result = append(result, NewPattern(PTAABColumn, []*DaerCard{card, findCard, findCards[0]}))
		}

		if findCards != nil && len(findCards) >= 2 {
			result = append(result, NewPattern(PTAABColumn, append(findCards, card)))
		}
	} else {
		findCards := FindCards(cards, card.value, !card.big)
		if findCards != nil && len(findCards) == 2 {
			result = append(result, NewPattern(PTAABColumn, append(findCards, card)))
		}
	}

	return
}

//移除多余的模式
func (controller *HuController) RemoveUselessPattern(patterns []*DaerPattern) (result []*DaerPattern) {

	result = append(result, patterns...)

	patternAmount := StatisticsPatternAmount(patterns)

	// for id, v := range patternAmount {
	// 	logger.Error("ID:", id)
	// 	PrintPatterns(v)
	// }

	for _, v := range patternAmount {
		minAmount := int32(255)
		if v[0].ptype == PTPair {
			minAmount = 1
		} else {
			for _, card := range v[0].cards {
				if card.big {
					curCardAmount := controller.bigCardAmount[card.value-1]
					if minAmount > curCardAmount {
						minAmount = curCardAmount
					}
				} else {
					curCardAmount := controller.smallCardAmount[card.value-1]
					if minAmount > curCardAmount {
						minAmount = curCardAmount
					}
				}
			}

		}

		for i := minAmount; i < int32(len(v)); i++ {
			for k, r := range result {
				if r.id == v[0].id {
					result = append(result[:k], result[k+1:]...)
					break
				}
			}
		}
	}

	return
}

//统计每种牌的数量
func StatisticsCardAmount(cards []*DaerCard) (smallCardAmount [10]int32, bigCardAmount [10]int32) {
	if cards == nil {
		logger.Error("StatisticsCardAmount:cards is nil.")
		return
	}

	for _, v := range cards {
		if v.value <= 0 || v.value > 10 {
			logger.Error("错误的牌，牌的值不在指定范围内。手上的牌如下：")
			PrintCards(cards)
			continue
		}

		if v.big {
			bigCardAmount[v.value-1]++
		} else {
			smallCardAmount[v.value-1]++
		}
	}

	return
}

//统计模式的数量
func StatisticsPatternAmount(patterns []*DaerPattern) (patternAmount map[uint][]*DaerPattern) {
	if patterns == nil {
		logger.Error("pattern is nil.")
		return nil
	}

	patternAmount = make(map[uint][]*DaerPattern, 0)

	if patterns == nil {
		return patternAmount
	}

	for _, v := range patterns {
		if patternAmount[v.id] == nil {
			patternAmount[v.id] = make([]*DaerPattern, 0)
		}

		patternAmount[v.id] = append(patternAmount[v.id], v)
	}
	return
}

//从一个切片中移除指定类型的所有Card
func RemoveCardsByType(cards []*DaerCard, value int32, isBig bool) (result []*DaerCard) {
	if cards == nil {
		logger.Error("RemoveCardsByType:cards is nil.")
		return
	}

	result = make([]*DaerCard, 0)
	for _, v := range cards {
		if v.value != value || v.big != isBig {
			result = append(result, v)
		}
	}

	return
}

//从一个切片中移除指定类型的Card
func RemoveCardByType(cards []*DaerCard, value int32, isBig bool) []*DaerCard {
	if cards == nil {
		logger.Error("RemoveCardByType:cards is nil.")
		return nil
	}

	for i, v := range cards {
		if v.value == value && v.big == isBig {
			cards = append(cards[:i], cards[i+1:]...)
			break
		}
	}

	return cards
}

//在列表中查找指定的Card
func FindCard(cards []*DaerCard, value int32, isBig bool) *DaerCard {
	if cards == nil {
		logger.Error("FindCard:cards is nil.")
		return nil
	}

	for i, v := range cards {
		if v.value == value && v.big == isBig {
			return cards[i]
		}
	}

	return nil
}

//在列表中查找指定的Card
func FindCards(cards []*DaerCard, value int32, isBig bool) []*DaerCard {
	if cards == nil {
		logger.Error("FindCards:cards is nil.")
		return nil
	}

	result := []*DaerCard{}
	for i, v := range cards {
		if v.value == value && v.big == isBig {
			result = append(result, cards[i])
		}
	}

	return result
}

//测试性能统计时间用的
var cloneCardTime float64 = 0
var checkTime float64 = 0
var appendToContainerTime float64 = 0
var stripSameTime float64 = 0
var stripSameCompareTime float64 = 0

//计算所有的模式组
func (controller *HuController) CalcPatternGroup(n int, patterns []*DaerPattern) (patternGroups []*DaerPatternGroup) {
	//patternGroups = make([]*DaerPatternGroup, 0)
	tempPatternGroups := make([]*DaerPatternGroup, 0)

	if n <= 0 || len(patterns) <= 0 {
		return
	}

	//fmt.Println("N:", n, len(patterns))
	if n > len(patterns) {
		logger.Error("n must be less than patterns's length. N:(%s) Patterns_Length:%s", n, len(patterns))
		PrintCards(controller.cards)
		return
	}

	curTime := time.Now()
	result := C(n, len(patterns))
	//fmt.Println("生成排列组合数 用时：", time.Now().Sub(curTime).Seconds())

	//fmt.Println("模式组合数：", n, len(patterns), len(result))
	//	for k, v := range tempSmallCardCount {
	//		fmt.Print(k, v)
	//		fmt.Print(", ")
	//	}
	//	fmt.Println("=========")
	//	for k, v := range tempBigCardCount {
	//		fmt.Print(k, v)
	//		fmt.Print(", ")
	//	}

	for i := 0; i < len(result); i++ {
		curTime = time.Now()
		tempSmallCardCount := controller.smallCardAmount
		tempBigCardCount := controller.bigCardAmount
		cloneCardTime += time.Now().Sub(curTime).Seconds()

		//		fmt.Println("小：", controller.smallCardAmount)
		//		fmt.Println("大：", controller.bigCardAmount)

		isFailure := false

		tempPatternList := make([]*DaerPattern, 0)

		for j := 0; j < len(result[i]); j++ {

			curTime = time.Now()

			index := result[i][j]
			pattern := patterns[index]
			if pattern.ptype == PTPair {
				for k := 0; k < len(tempPatternList); k++ {
					if tempPatternList[k].ptype == PTPair {
						isFailure = true
						break
					}
				}
			}

			if isFailure {
				break
			}

			for k := 0; k < len(pattern.cards); k++ {
				patternCard := pattern.cards[k]
				if patternCard.big {
					if tempBigCardCount[patternCard.value-1] <= 0 {
						isFailure = true
						break
					}
					tempBigCardCount[patternCard.value-1]--
				} else {
					if tempSmallCardCount[patternCard.value-1] <= 0 {
						isFailure = true
						break
					}
					tempSmallCardCount[patternCard.value-1]--
				}
			}

			if isFailure {
				break
			}

			checkTime += time.Now().Sub(curTime).Seconds()

			curTime = time.Now()
			tempPatternList = append(tempPatternList, pattern)
			appendToContainerTime += time.Now().Sub(curTime).Seconds()
		}

		//		fmt.Println("小：", tempSmallCardCount)
		//		fmt.Println("大：", tempBigCardCount)

		if !isFailure {
			//fmt.Println("生成的模式组：")
			//PrintPatterns(tempPatternList)
			curTime = time.Now()
			patternGroup := NewPatternGroup(tempPatternList)
			tempPatternGroups = append(tempPatternGroups, patternGroup)
			appendToContainerTime += time.Now().Sub(curTime).Seconds()
		}

		//		fmt.Println("生成的模式组：")
		//		PrintPatternGroups(tempPatternGroups)
		curTimeS := time.Now()
		//fmt.Println("可行的模式：", len(tempPatternGroups))
		//		for i := 0; i < len(tempPatternGroups); i++ {
		//			isExist := false

		//			curTime = time.Now()
		//			for j := 0; j < len(patternGroups); j++ {

		//				if tempPatternGroups[i].IsEqual(patternGroups[j]) {
		//					isExist = true
		//					break
		//				}
		//			}

		//			stripSameCompareTime += time.Now().Sub(curTime).Seconds()

		//			if !isExist {
		//				patternGroups = append(patternGroups, tempPatternGroups[i])
		//			}
		//		}

		stripSameTime += time.Now().Sub(curTimeS).Seconds()

		//		fmt.Println("剔除重复后的模式组：")
		//		PrintPatternGroups(patternGroups)

	}

	patternGroups = tempPatternGroups
	//patternGroups = append(patternGroups, tempPatternGroups...)
	//	fmt.Println("生成后的模式组：", len(tempPatternGroups))
	//	PrintPatternGroups(tempPatternGroups, false)
	//	fmt.Println("生成后的模式组：==============")

	//fmt.Println("克隆的时间：", cloneCardTime)
	//fmt.Println("检测的时间：", checkTime)
	//fmt.Println("追加的时间：", appendToContainerTime)
	//fmt.Println("剔除重复的时间：", stripSameTime)
	//fmt.Println("剔除重复的比较的时间：", stripSameCompareTime)

	return
}

//生成模式的组合下标
func C(n int, m int) (result [][]int) {
	if n > m {
		logger.Error("n must be less than m")
		return nil
	}

	//	size := Factorial(m) / (Factorial(m-n) * Factorial(n))
	//	fmt.Println("组合数量：", n, m, size)
	result = make([][]int, 0)

	index := make([]int, m)
	//fmt.Println(m, index)
	for i := 0; i < m; i++ {
		index[i] = 0
	}

	for i := 0; i < n; i++ {
		index[i] = 1
	}

	result = append(result, GetC(index))

	for true {
		for i := 0; i < m-1; i++ {
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

func GetC(index []int) (result []int) {
	result = make([]int, 0)
	for i := 0; i < len(index); i++ {
		if index[i] == 1 {
			result = append(result, i)
		}
	}

	return
}

//获取每个组模式中的单牌
func (controller *HuController) GetSingleCardInPatternGroup(patternGroup []*DaerPatternGroup) (result [][]*DaerCard) {
	result = make([][]*DaerCard, 0)
	for _, v := range patternGroup {
		tempCards := []*DaerCard{}
		tempCards = append(tempCards, controller.cards...)
		for i := 0; i < len(v.patterns); i++ {
			for j := 0; j < len(v.patterns[i].cards); j++ {
				card := v.patterns[i].cards[j]
				tempCards = RemoveCardByType(tempCards, card.value, card.big)
			}
		}
		result = append(result, tempCards)
	}
	return
}

//计算每个组模式中胡的牌
func (controller *HuController) CalcHu(singleCards [][]*DaerCard, patternGroups []*DaerPatternGroup) []*DaerPatternGroup {
	for i, v := range singleCards {

		singleCardCount := len(v)
		pairCount, pairCards := patternGroups[i].GetPairAmount()
		if singleCardCount >= 3 || pairCount > 2 || (pairCount == 2 && singleCardCount > 0) {
			continue
		}

		logger.Info("当前的模式组：")
		PrintPatternGroup(patternGroups[i], false)
		logger.Info("一个模式组中的单牌数%s, 对子数：%s", singleCardCount, pairCount)
		logger.Info("单张牌：")
		PrintCards(v)
		logger.Info("对子：")
		PrintCards(pairCards)

		result := []*DaerCard{}
		if singleCardCount == 0 {
			if pairCount == 2 {
				for _, c := range pairCards {
					result = append(result, &DaerCard{value: c.value, big: !c.big, flag: cmn.CUnknown})
					result = append(result, &DaerCard{value: c.value, big: c.big, flag: cmn.CUnknown})
				}
			} else if pairCount == 1 {
				result = []*DaerCard{&DaerCard{value: pairCards[0].value, big: !pairCards[0].big, flag: cmn.CUnknown},
					&DaerCard{value: pairCards[0].value, big: pairCards[0].big, flag: cmn.CUnknown}}

				result = append(result, controller.CalcKanHu()...)
			} else {
				result = controller.CalcKanHu()
			}
		} else if singleCardCount == 1 {
			if pairCount == 0 {
				result = []*DaerCard{&DaerCard{value: v[0].value, big: v[0].big, flag: cmn.CUnknown}}
			}
		} else if singleCardCount == 2 {
			firstCard := v[0]
			secondCard := v[1]
			//AA mode
			if firstCard.IsEqual(secondCard) {
				if pairCount == 1 {
					result = []*DaerCard{&DaerCard{value: firstCard.value, big: !firstCard.big, flag: cmn.CUnknown},
						&DaerCard{value: firstCard.value, big: firstCard.big, flag: cmn.CUnknown},
						&DaerCard{value: pairCards[0].value, big: !pairCards[0].big, flag: cmn.CUnknown},
						&DaerCard{value: pairCards[0].value, big: pairCards[0].big, flag: cmn.CUnknown}}
				} else if pairCount == 0 {
					result = []*DaerCard{&DaerCard{value: firstCard.value, big: !firstCard.big, flag: cmn.CUnknown},
						&DaerCard{value: firstCard.value, big: firstCard.big, flag: cmn.CUnknown}}

					result = append(result, controller.CalcKanHu()...)
				} else {
					logger.Error("不可能存在此种情况")
				}
				//AB mode
			} else if firstCard.value == secondCard.value {
				result = []*DaerCard{&DaerCard{value: firstCard.value, big: true, flag: cmn.CUnknown},
					&DaerCard{value: firstCard.value, big: false, flag: cmn.CUnknown}}
				//SZ and EQS mode
			} else if firstCard.big == secondCard.big {
				offset := int(secondCard.value) - int(firstCard.value)
				//SZ
				switch offset {
				case 1:
					result = []*DaerCard{}
					if firstCard.value-1 > 0 {
						result = append(result, &DaerCard{value: firstCard.value - 1, big: firstCard.big, flag: cmn.CUnknown})
					}
					if secondCard.value+1 <= 10 {
						result = append(result, &DaerCard{value: secondCard.value + 1, big: secondCard.big, flag: cmn.CUnknown})
					}
				case -1:
					result = []*DaerCard{}
					if secondCard.value-1 > 0 {
						result = append(result, &DaerCard{value: secondCard.value - 1, big: secondCard.big, flag: cmn.CUnknown})
					}
					if firstCard.value+1 <= 10 {
						result = append(result, &DaerCard{value: firstCard.value + 1, big: firstCard.big, flag: cmn.CUnknown})
					}
				case 2, -2:
					result = []*DaerCard{&DaerCard{value: (firstCard.value + secondCard.value) / 2, big: firstCard.big, flag: cmn.CUnknown}}
					//EQS mode
				default:
					if firstCard.value == 2 {
						if secondCard.value == 7 {
							result = []*DaerCard{&DaerCard{value: 10, big: firstCard.big, flag: cmn.CUnknown}}
						}
						if secondCard.value == 10 {
							result = []*DaerCard{&DaerCard{value: 7, big: firstCard.big, flag: cmn.CUnknown}}
						}
					}

					if firstCard.value == 7 {
						if secondCard.value == 2 {
							result = []*DaerCard{&DaerCard{value: 10, big: firstCard.big, flag: cmn.CUnknown}}
						}
						if secondCard.value == 10 {
							result = []*DaerCard{&DaerCard{value: 2, big: firstCard.big, flag: cmn.CUnknown}}
						}
					}

					if firstCard.value == 10 {
						if secondCard.value == 2 {
							result = []*DaerCard{&DaerCard{value: 7, big: firstCard.big, flag: cmn.CUnknown}}
						}
						if secondCard.value == 7 {
							result = []*DaerCard{&DaerCard{value: 2, big: firstCard.big, flag: cmn.CUnknown}}
						}
					}
				}
			}
		}

		//检查是否有胡
		if result != nil && len(result) > 0 {
			logger.Info("特殊情况下的胡牌：")
			PrintCards(result)

			patternGroups[i].kaoCards = append(patternGroups[i].kaoCards, singleCards[i]...)
			patternGroups[i].huCards = append(patternGroups[i].huCards, result...)
		}
	}

	return patternGroups
}

//统计坎的胡牌
func (controller *HuController) CalcKanHu() (result []*DaerCard) {
	result = make([]*DaerCard, 0)
	if controller.player != nil {
		for _, ptn := range controller.player.fixedpatterns {
			result = append(result, ptn.cards[0])
		}
	} else {
		logger.Error("胡牌控制器没有关联玩家")
	}

	logger.Info("坎的胡牌：")
	PrintCards(result)

	return
}

//剔除不能胡的组模式
func (controller *HuController) StripNoHuPatternGroup(groupPatterns []*DaerPatternGroup) (result []*DaerPatternGroup) {
	//剔除不能胡的牌
	for _, v := range groupPatterns {
		if v.CanHu() {
			result = append(result, v)
		}
	}

	return result
}

//通过胡数和胡的牌产生唯一ID
func (controller *HuController) GeneratePatternGroupID(groupPatterns []*DaerPatternGroup) {

	for _, v := range groupPatterns {
		v.GenerateID()
	}
}

//剔除相同的模式组
func (controller *HuController) StripSamePatternGroup(groupPatterns []*DaerPatternGroup) (result []*DaerPatternGroup) {
	result = make([]*DaerPatternGroup, 0)

	//剔除重复的牌
	isExist := false
	for _, v := range groupPatterns {
		isExist = false
		for _, rv := range result {
			if v.id == rv.id {
				isExist = true
			}
		}

		if !isExist {
			result = append(result, v)
		}
	}

	return
}
