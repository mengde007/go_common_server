package majiangserver

import (
	cmn "common"
	"logger"
)

type MaJiangPatternGroup struct {
	id       int32
	patterns []*MaJiangPattern
	kaoCards []*MaJiangCard
	huCards  []*MaJiangCard
}

//新建一个模式组
func NewPatternGroup(patterns []*MaJiangPattern) *MaJiangPatternGroup {
	if patterns == nil {
		logger.Error("NewPatternGroup:patterns is nil.")
		return nil
	}

	o := new(MaJiangPatternGroup)
	o.Init(patterns)

	return o
}

//初始化函数
func (self *MaJiangPatternGroup) Init(patterns []*MaJiangPattern) {
	if patterns == nil {
		logger.Error("Init:patterns is nil.")
		return
	}

	//赋值模式
	self.patterns = patterns
}

//产生模式组的ID
func (self *MaJiangPatternGroup) GenerateID() {
	if !self.CanHu() {
		logger.Error("ID必须由胡数和胡的牌产生")
		return
	}

	self.id = 0
	for _, v := range self.patterns {
		self.id += v.id
	}

	self.id *= 10000

	for _, v := range self.kaoCards {
		ctype, val := v.CurValue()
		self.id += ctype*1000 + val*100
	}

	for _, v := range self.huCards {
		ctype, val := v.CurValue()
		self.id += ctype*10 + val
	}
}

//比较两个模式是否相等，此函数可以再优化
func (self *MaJiangPatternGroup) IsEqual(g *MaJiangPatternGroup) bool {
	if g == nil {
		logger.Error("g is nil")
		return false
	}

	return self.id == g.id
}

//获取对子牌
func (self *MaJiangPatternGroup) GetPairCard() (card []*MaJiangCard) {
	card = make([]*MaJiangCard, 0)
	if self.patterns == nil {
		logger.Error("HavePair:self.patterns is nil")
		return
	}

	for _, v := range self.patterns {
		if v == nil {
			continue
		}

		if v.ptype == PTPair {
			card = append(card, v.cards[0])
		}
	}

	return
}

//指定牌时候被包含在hu牌列表里
func (self *MaJiangPatternGroup) CanHuSpecificCard(card *MaJiangCard) bool {
	for _, c := range self.huCards {
		if c.IsEqual(card) {
			return true
		}
	}
	return false
}

//检查能否胡牌
func (self *MaJiangPatternGroup) CanHu() bool {
	return self.kaoCards != nil && len(self.kaoCards) > 0 && self.huCards != nil && len(self.huCards) > 0
}

//是否是打对子的牌型
func (self *MaJiangPatternGroup) IsDaDuiZi(huCard *MaJiangCard) bool {
	if huCard == nil {
		logger.Error("huCard is nil.")
		return false
	}

	if !self.CanHuSpecificCard(huCard) {
		return false
	}

	for _, p := range self.patterns {
		if p.ptype == PTSZ || p.ptype == PTSingle || p.ptype == PTUknown {
			return false
		}
	}

	if len(self.kaoCards) >= 2 {
		if !self.kaoCards[0].IsEqual(self.kaoCards[1]) {
			return false
		}
	}

	return true

}

//是否是清一色
func (self *MaJiangPatternGroup) IsQingYiSe(showPatterns []*MaJiangPattern, huCard *MaJiangCard) bool {
	if showPatterns == nil {
		logger.Error("showPatterns is nil.")
	}
	if huCard == nil {
		logger.Error("huCard is nil.")
		return false
	}

	if !self.CanHuSpecificCard(huCard) {
		return false
	}

	allPatterns := make([]*MaJiangPattern, 0)
	allPatterns = append(allPatterns, showPatterns...)
	allPatterns = append(allPatterns, self.patterns...)

	showpatternTypeAmount := GetTypeInfoByPatternList(allPatterns, nil)
	if len(showpatternTypeAmount) > 1 {
		return false
	} else if len(showpatternTypeAmount) <= 0 {
		return true
	} else {
		cType, _ := huCard.CurValue()
		return showpatternTypeAmount[0] == cType
	}
}

//是否是无鬼
func (self *MaJiangPatternGroup) IsNoneHongZhong(showPatterns []*MaJiangPattern) bool {
	if showPatterns == nil {
		logger.Error("showPatterns is nil.")
	}

	for _, sp := range showPatterns {
		for _, c := range sp.cards {
			isHuCard := c.flag&cmn.CHu > 0
			if !isHuCard && c.cType == HongZhong {
				return false
			}
		}
	}

	for _, pg := range self.patterns {
		for _, c := range pg.cards {
			isHuCard := c.flag&cmn.CHu > 0
			if !isHuCard && c.cType == HongZhong {
				return false
			}
		}
	}

	return true

}

//是否是小七对
func (self *MaJiangPatternGroup) IsQiDui(huCard *MaJiangCard) bool {

	if huCard == nil {
		logger.Error("huCard is nil.")
		return false
	}

	if !self.CanHuSpecificCard(huCard) {
		return false
	}

	if len(self.patterns) != 6 {
		return false
	}

	for _, pg := range self.patterns {
		if pg.ptype != PTPair {
			return false
		}
	}

	return true

}

//获取指定模式的数量
func (self *MaJiangPatternGroup) GetPatternAmount(ptype int32) int32 {
	var patternAmount int32 = 0
	for _, pattern := range self.patterns {
		if ptype == pattern.ptype {
			patternAmount++
		}
	}

	return patternAmount
}

//获取所有牌
func (self *MaJiangPatternGroup) GetCards() (result []*MaJiangCard) {
	result = make([]*MaJiangCard, 0)
	//模式的牌
	for _, pattern := range self.patterns {
		result = append(result, pattern.cards...)
	}

	//靠牌
	result = append(result, self.kaoCards...)

	return
}

//获取指定类型的牌
func (self *MaJiangPatternGroup) GetSpecificTypeCards(cType int32) (result []*MaJiangCard) {
	result = make([]*MaJiangCard, 0)
	//模式的牌
	for _, pattern := range self.patterns {
		if pattern.cards == nil {
			continue
		}

		for _, card := range pattern.cards {
			if card.cType == cType {
				result = append(result, card)
			}
		}

	}

	//靠牌
	if self.kaoCards == nil {
		return
	}

	for _, card := range self.kaoCards {
		if card.cType == cType {
			result = append(result, card)
		}
	}

	return
}
