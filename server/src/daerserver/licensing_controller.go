package daerserver

import (
	"logger"
)

const (
	LCNone = iota
	HuPengSimultaneously
	TestHu
	TestErLongTouYi
)

var SpecificLicensingType int = LCNone
var FixedBankerIndex int = -1

// var SpecificLicensingType int = TestHu
// var FixedBankerIndex int = 0

func Licensing(ctype int, room *DaerRoom) []*DaerCard {
	if room == nil {
		return nil
	}

	switch ctype {
	case HuPengSimultaneously:
		return FHuPengSimultaneously(room)
	case TestHu:
		return FTestHu(room)
	case TestErLongTouYi:
		return FErLongTouYi(room)
	default:
		logger.Error("没有此发牌类型：", ctype)
		return nil
	}

	return nil
}

func FHuPengSimultaneously(room *DaerRoom) []*DaerCard {
	cards := make([]*DaerCard, CardTotalAmount)
	copy(cards, room.cards)

	logger.Info("特殊发牌之前======")
	PrintCards(cards)
	//修改第0个玩家的牌
	FillHandCard(cards, 0, []int{
		20, 20, 20, 20,
		17, 17, 17, 17,
		11, 11, 11, 11,
		2, 7, 10,
		3, 4, 5,
		12, 2})

	//修改第1个玩家的牌
	FillHandCard(cards, 1, []int{
		1, 1, 1,
		10, 10, 10,
		7, 7, 7,
		19, 19, 19,
		2, 3, 4,
		2, 3, 4,
		12, 12})

	//修改第2个玩家的牌
	FillHandCard(cards, 2, []int{
		13, 13, 13,
		3, 4, 5,
		14, 15, 16,
		8, 18, 18,
		14, 15, 16,
		9, 9, 9,
		6, 6})

	cards[79] = NewCard(0, 2, true)
	//cards[78] = NewCard(0, 8, true)
	return cards
}

func FErLongTouYi(room *DaerRoom) []*DaerCard {
	cards := make([]*DaerCard, CardTotalAmount)
	copy(cards, room.cards)

	logger.Info("特殊发牌之前======")
	PrintCards(cards)
	//修改第0个玩家的牌
	FillHandCard(cards, 0, []int{
		1, 2, 3,
		4, 4, 4,
		7, 8, 9,
		10, 10, 10, 5,
		20, 20, 20, 20,
		16, 17, 18})

	//修改第1个玩家的牌
	FillHandCard(cards, 1, []int{
		1, 2, 3,
		2,
		5, 5, 5,
		11, 12, 13,
		14, 15, 16,
		16, 17, 18,
		19, 19, 19, 19})

	//修改第2个玩家的牌
	FillHandCard(cards, 2, []int{
		1, 2, 3,
		13, 13, 13,
		8, 9, 7,
		16, 17, 18,
		14, 15, 17,
		14, 15, 11,
		12, 4})

	cards[78] = NewCard(0, 6, false)
	cards[79] = NewCard(0, 10, false)
	return cards
}

//测试模式能胡不
func FTestHu(room *DaerRoom) []*DaerCard {
	cards := make([]*DaerCard, CardTotalAmount)
	copy(cards, room.cards)

	FillHandCard(cards, 0, []int{
		1, 1, 1,
		11, 11, 11, 11,
		13, 13, 3,
		14, 14, 14,
		5, 6, 7,
		4,
		12, 17, 20})

	//case1
	// FillHandCard(cards, 0, []int{
	// 	14, 14, 14,
	// 	19, 19, 19,
	// 	17, 17, 17,
	// 	12, 2, 12,
	// 	5, 6, 7,
	// 	1, 1,
	// 	7, 8, 9})

	// FillHandCard(cards, 0, []int{
	// 	1, 11, 11,
	// 	2, 12, 12,
	// 	4, 5, 6,
	// 	6, 16, 16,
	// 	7, 17, 7,
	// 	18, 1,
	// 	9, 9, 9})

	//case2
	// FillHandCard(cards, 1, []int{
	// 	2, 2, 2,
	// 	3, 3, 3,
	// 	8, 8, 8,
	// 	13, 13, 13,
	// 	1, 5, 6,
	// 	15, 11, 6,
	// 	12, 12})

	// //case3
	// FillHandCard(cards, 2, []int{
	// 	1, 11, 11,
	// 	10, 10, 10,
	// 	4, 4, 14,
	// 	13, 15, 3,
	// 	4, 5,
	// 	7, 9, 9,
	// 	17, 20, 19})

	//case4
	// FillHandCard(cards, 1, []int{
	// 	1, 11,
	// 	2, 2,
	// 	13, 13, 13,
	// 	14, 14, 14,
	// 	7, 7, 7,
	// 	10, 10, 6,
	// 	8, 8, 6, 16})

	//case6
	// FillHandCard(cards, 0, []int{
	// 	6, 6,
	// 	13, 3,
	// 	19, 19, 19, 19,
	// 	20, 20, 20,
	// 	15, 15, 5,
	// 	18, 18, 18,
	// 	1, 1, 11})

	// //case5
	// FillHandCard(cards, 1, []int{
	// 	11, 11, 11,
	// 	15, 15, 5,
	// 	12, 12, 12, 12,
	// 	7, 8, 9,
	// 	7, 8, 9,
	// 	20, 18,
	// 	10, 10})

	// //case7
	// FillHandCard(cards, 2, []int{
	// 	2, 2, 2, 2,
	// 	9, 9, 17,
	// 	14, 14, 14,
	// 	1, 1, 5,
	// 	5, 7, 7,
	// 	8, 8, 10,
	// 	10})

	//case8
	//	FillHandCard(cards, 0, []int{
	//		12, 12, 12,
	//		19, 19, 19,
	//		4, 4, 4,
	//		13, 13,
	//		2, 7, 10,
	//		8, 18, 18,
	//		8, 9, 10})

	// cards[73] = NewCard(0, 3, true)
	// cards[74] = NewCard(0, 3, true)
	// cards[75] = NewCard(0, 3, true)
	// cards[76] = NewCard(0, 6, true)
	// cards[77] = NewCard(0, 4, false)
	cards[78] = NewCard(0, 1, true)
	cards[79] = NewCard(0, 1, true)
	cards[60] = NewCard(0, 1, true)

	return cards
}

//填充一个人的手牌
func FillHandCard(cards []*DaerCard, playerIndex int, cardsInt []int) {
	if playerIndex >= 3 || playerIndex < 0 {
		logger.Error("玩家索引有误（0-2）")
		return
	}

	if len(cardsInt) != FirstCardsAmount {
		logger.Error("手牌必须是：%s）", FirstCardsAmount)
		return
	}

	if len(cards) != CardTotalAmount {
		logger.Error("牌的数量不对：%）", len(cards))
		return
	}

	logger.Info("提取的牌：", cardsInt)
	for j, v := range cardsInt {

		isBig := v > 10
		value := v
		if isBig {
			value = v - 10
		}

		baseIndex := playerIndex*FirstCardsAmount + j
		//logger.Error("BaseIndex:%s,  J:%s", baseIndex, j)
		find := false
		vaildRange := cards[baseIndex:]
		PrintCards(vaildRange)
		for i, c := range vaildRange {
			if c.big == isBig && int32(value) == c.value {
				find = true
				temp := cards[baseIndex]
				cards[baseIndex] = cards[baseIndex+i]
				cards[baseIndex+i] = temp
				break
			}
		}

		if !find {
			logger.Error("没有此牌了：", v)
		}
	}

}
