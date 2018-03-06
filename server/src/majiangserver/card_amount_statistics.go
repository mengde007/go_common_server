package majiangserver

import (
	cmn "common"
	"logger"
)

const CardMaxValue = 9

//此类是不能统计红中数量
type CardAmountStatistics struct {
	tiaoCardsAmount [CardMaxValue + 1]int32
	tongCardsAmount [CardMaxValue + 1]int32
	wanCardsAmount  [CardMaxValue + 1]int32
	hongZhongAmount int32
}

func NewCardAmountStatisticsByCards(cards []*MaJiangCard, replaceMode bool) *CardAmountStatistics {
	o := &CardAmountStatistics{}

	o.CalcCardAmountByCards(cards, replaceMode)
	return o
}

func NewCardAmountStatisticsByPatternGroup(patternGroup *MaJiangPatternGroup, replaceMode bool) *CardAmountStatistics {
	o := &CardAmountStatistics{}

	o.CalcCardAmountByPatternGroup(patternGroup, replaceMode)
	return o
}

func (self *CardAmountStatistics) CalcCardAmountByCards(cards []*MaJiangCard, replaceMode bool) {
	if cards == nil {
		logger.Error("cards is nil.")
	}

	self.Reset()

	if replaceMode {
		for _, card := range cards {
			cType, val := card.CurValue()

			switch cType {
			case Tiao:
				self.tiaoCardsAmount[val]++
			case Tong:
				self.tongCardsAmount[val]++
			case Wan:
				self.wanCardsAmount[val]++
			default:
				logger.Error("未知类型的牌", cType, val)
			}
		}
	} else {
		for _, card := range cards {
			switch card.cType {
			case Tiao:
				self.tiaoCardsAmount[card.value]++
			case Tong:
				self.tongCardsAmount[card.value]++
			case Wan:
				self.wanCardsAmount[card.value]++
			case HongZhong:
				self.hongZhongAmount++
			default:
				logger.Error("未知类型的牌", card.cType, card.value)
			}
		}
	}
}

func (self *CardAmountStatistics) CalcCardAmountByPatternGroup(patternGroup *MaJiangPatternGroup, replaceMode bool) {
	if patternGroup == nil {
		logger.Error("patternGroup is nil.")
		return
	}

	self.CalcCardAmountByCards(patternGroup.GetCards(), replaceMode)
}

func (self *CardAmountStatistics) Reset() {
	//条
	for i := 0; i <= CardMaxValue; i++ {
		self.tiaoCardsAmount[i] = 0
	}

	//筒
	for i := 0; i <= CardMaxValue; i++ {
		self.tongCardsAmount[i] = 0
	}

	//万
	for i := 0; i <= CardMaxValue; i++ {
		self.wanCardsAmount[i] = 0
	}

	//红中
	self.hongZhongAmount = 0

}

func (self *CardAmountStatistics) GetAmountInfo() (result [HongZhong + 1][CardMaxValue + 1]int32) {
	result[Tiao] = self.tiaoCardsAmount
	result[Tong] = self.tongCardsAmount
	result[Wan] = self.wanCardsAmount
	result[HongZhong] = [CardMaxValue + 1]int32{self.hongZhongAmount}

	return result
}

func (self *CardAmountStatistics) GetAmountInfoByType(cType int32) (result [10]int32) {

	switch cType {
	case Tiao:
		result = self.tiaoCardsAmount
	case Tong:
		result = self.tongCardsAmount
	case Wan:
		result = self.wanCardsAmount
	case HongZhong:
		result[0] = self.hongZhongAmount
	default:
		logger.Error("未知类型的牌", cType)
	}

	return result
}

func (self *CardAmountStatistics) GetCardAmountByType(cType int32) (result int32) {

	switch cType {
	case Tiao:
		for i := 0; i <= CardMaxValue; i++ {
			result += self.tiaoCardsAmount[i]
		}
	case Tong:
		for i := 0; i <= CardMaxValue; i++ {
			result += self.tongCardsAmount[i]
		}
	case Wan:
		for i := 0; i <= CardMaxValue; i++ {
			result += self.wanCardsAmount[i]
		}
	case HongZhong:
		result = self.hongZhongAmount
	default:
		logger.Error("未知类型的牌", cType)
	}

	return
}

func (self *CardAmountStatistics) GetCardAmount(cType, val int32) int32 {
	if val <= 0 || val > CardMaxValue {
		logger.Error("val 的值是错误的:", val)
		return 0
	}

	switch cType {
	case Tiao:
		return self.tiaoCardsAmount[val]
	case Tong:
		return self.tongCardsAmount[val]
	case Wan:
		return self.wanCardsAmount[val]
	case HongZhong:
		return self.hongZhongAmount
	default:
		logger.Error("未知类型的牌", cType, val)
		return 0
	}
}

func (self *CardAmountStatistics) GetTypeAmount(includeHongZhong bool, excludeType []int32) (result int32) {

	if !Exist(excludeType, Tiao) {
		amount := self.GetCardAmountByType(Tiao)
		if amount > 0 {
			result++
		}
	}

	if !Exist(excludeType, Tong) {
		amount := self.GetCardAmountByType(Tong)
		if amount > 0 {
			result++
		}
	}

	if !Exist(excludeType, Wan) {
		amount := self.GetCardAmountByType(Wan)
		if amount > 0 {
			result++
		}
	}

	if !Exist(excludeType, HongZhong) {
		if includeHongZhong && self.hongZhongAmount > 0 {
			result++
		}
	}

	return
}

//获取达到指定数量的牌有多少
func (self *CardAmountStatistics) GetAmountBySpecificAmount(amount int32) (result int32) {

	return int32(len(self.GetCardsBySpecificAmount(amount, nil)))
}

//获取达到指定数量的牌
func (self *CardAmountStatistics) GetCardsBySpecificAmount(amount int32, excludeCards []*MaJiangCard) (result []*MaJiangCard) {

	result = make([]*MaJiangCard, 0)

	for v, a := range self.tiaoCardsAmount {
		if IsExistCardByTypeAndValue(excludeCards, Tiao, int32(v)) {
			continue
		}

		if a >= amount {
			result = append(result, &MaJiangCard{value: int32(v), cType: Tiao, flag: cmn.CUnknown})
		}
	}

	for v, a := range self.tongCardsAmount {
		if IsExistCardByTypeAndValue(excludeCards, Tong, int32(v)) {
			continue
		}

		if a >= amount {
			result = append(result, &MaJiangCard{value: int32(v), cType: Tong, flag: cmn.CUnknown})
		}
	}

	for v, a := range self.wanCardsAmount {
		if IsExistCardByTypeAndValue(excludeCards, Wan, int32(v)) {
			continue
		}

		if a >= amount {
			result = append(result, &MaJiangCard{value: int32(v), cType: Wan, flag: cmn.CUnknown})
		}
	}

	if !IsExistCardByTypeAndValue(excludeCards, HongZhong, 0) {
		if self.hongZhongAmount >= amount {
			result = append(result, &MaJiangCard{value: 0, cType: HongZhong, flag: cmn.CUnknown})
		}
	}

	return
}

//是否存在
func IsExistCard(cards []*MaJiangCard, card *MaJiangCard) bool {
	if cards == nil || card == nil {
		return false
	}

	for _, c := range cards {
		if c.IsEqual(card) {
			return true
		}
	}

	return false
}

//是否存在
func IsExistCardByTypeAndValue(cards []*MaJiangCard, cType, cVal int32) bool {
	if cards == nil {
		return false
	}

	for _, c := range cards {
		if c.IsEqualByTypeAndValue(cType, cVal) {
			return true
		}
	}

	return false
}

func Exist(list []int32, val int32) bool {
	if list == nil {
		return false
	}

	for _, v := range list {
		if v == val {
			return true
		}
	}

	return false
}

func GetTypeInfoByPatternList(patternList []*MaJiangPattern, excludeType []int32) (result []int32) {
	result = make([]int32, 0)
	if patternList == nil || len(patternList) <= 0 {
		return
	}

	haveTiao := false
	haveTong := false
	haveWan := false

	excludeTiao := Exist(excludeType, Tiao)
	excludeTong := Exist(excludeType, Tong)
	excludeWan := Exist(excludeType, Wan)

	for _, p := range patternList {
		//check over
		if (excludeTiao || haveTiao) && (excludeTong || haveTong) && (excludeWan || haveWan) {
			break
		}

		//check param
		if p == nil || p.cards == nil || len(p.cards) <= 0 {
			continue
		}

		for _, c := range p.cards {
			if c == nil {
				continue
			}

			cType, _ := c.CurValue()

			if !excludeTiao && cType == Tiao && !haveTiao {
				haveTiao = true
			}

			if !excludeTong && cType == Tong && !haveTong {
				haveTong = true
			}

			if !excludeWan && cType == Wan && !haveWan {
				haveWan = true
			}
		}
	}

	if haveTiao {
		result = append(result, Tiao)
	}

	if haveTong {
		result = append(result, Tong)
	}

	if haveWan {
		result = append(result, Wan)
	}

	return
}

func GetTypeInfoByCardList(cards []*MaJiangCard, excludeType []int32) (result []int32) {
	result = make([]int32, 0)
	if cards == nil || len(cards) <= 0 {
		return
	}

	haveTiao := false
	haveTong := false
	haveWan := false

	excludeTiao := Exist(excludeType, Tiao)
	excludeTong := Exist(excludeType, Tong)
	excludeWan := Exist(excludeType, Wan)

	for _, c := range cards {
		if c == nil {
			continue
		}

		if (excludeTiao || haveTiao) && (excludeTong || haveTong) && (excludeWan || haveWan) {
			break
		}

		cType, _ := c.CurValue()
		if !excludeTiao && cType == Tiao && !haveTiao {
			haveTiao = true
		}
		if !excludeTong && cType == Tong && !haveTong {
			haveTong = true
		}
		if !excludeWan && cType == Wan && !haveWan {
			haveWan = true
		}
	}

	if haveTiao {
		result = append(result, Tiao)
	}

	if haveTong {
		result = append(result, Tong)
	}

	if haveWan {
		result = append(result, Wan)
	}

	return
}
