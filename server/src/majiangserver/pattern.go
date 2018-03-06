package majiangserver

import (
	"logger"
	//"sort"
)

type MaJiangPattern struct {
	id            int32
	ptype         int32
	cType         int32
	cards         []*MaJiangCard
	isShowPattern bool
}

//新建一个模式
func NewPattern(ptype int32, cards []*MaJiangCard, isShowPattern bool) *MaJiangPattern {
	if cards == nil || len(cards) <= 0 {
		logger.Error("NewPattern:selfs is nil.")
		return nil
	}

	o := new(MaJiangPattern)
	o.ptype = ptype
	o.isShowPattern = isShowPattern
	o.Init(cards)
	return o
}

//初始化一个模式
func (self *MaJiangPattern) Init(cards []*MaJiangCard) {
	// if self.ptype == PTSZ {
	// 	sort.Sort(CardList(cards))
	// }

	self.cards = cards
	self.checkCardType()
	self.calcID()
}

//计算模式的ID
func (self *MaJiangPattern) calcID() {
	var id int32 = 0
	for _, v := range self.cards {
		id += v.value
	}

	self.id = self.ptype*1000 + self.cType*100 + id
}

//检查牌的花色
func (self *MaJiangPattern) checkCardType() {
	if self.cards == nil || len(self.cards) <= 0 {
		logger.Error("checkBigCard:using before must init.")
	}

	self.cType, _ = self.cards[0].CurValue()

	for _, v := range self.cards {
		cType, _ := v.CurValue()
		if cType != self.cType {
			self.cType = UnknowCardType
			break
		}
	}
}

//是否相等
func (self *MaJiangPattern) IsEqual(p *MaJiangPattern) bool {
	if p == nil {
		return false
	}

	if self.ptype != p.ptype {
		return false
	}

	if self.cType != p.cType {
		return false
	}

	nSelfCards, hzSelfCards := SplitCards(self.cards)
	nPCards, hzPCards := SplitCards(p.cards)
	if len(nSelfCards) != len(nPCards) || len(hzSelfCards) != len(hzPCards) {
		return false
	}

	temp := make([]*MaJiangCard, len(nSelfCards))
	copy(temp, nSelfCards)
	for _, c := range nPCards {
		removedSuccess := true
		removedSuccess, temp = RemoveCardByType(temp, c.cType, c.value)
		if !removedSuccess {
			return false
		}
	}

	if len(temp) > 0 {
		return false
	}

	return true
}

//是否是全红中
func (self *MaJiangPattern) IsAllHZ() bool {
	for _, c := range self.cards {
		if !c.IsHongZhong() {
			return false
		}
	}

	return true
}
