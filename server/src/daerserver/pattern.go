package daerserver

import (
	"logger"
)

type DaerPattern struct {
	id     uint
	ptype  uint
	big    bool
	red    bool
	cards  []*DaerCard
	weight int8
}

//新建一个模式
func NewPattern(ptype uint, cards []*DaerCard) *DaerPattern {
	if cards == nil || len(cards) <= 0 {
		logger.Error("NewPattern:patterns is nil.")
		return nil
	}

	o := new(DaerPattern)
	o.ptype = ptype
	o.Init(cards)
	return o
}

//初始化一个模式
func (pattern *DaerPattern) Init(cards []*DaerCard) {
	pattern.cards = cards
	pattern.checkBigCard()
	pattern.calcID()
	pattern.checkRed()
	pattern.calcWeight()
}

//计算模式的ID
func (pattern *DaerPattern) calcID() {
	var id uint = 0
	for _, v := range pattern.cards {
		if v.big {
			id += uint(v.value) * 100
		} else {
			id += uint(v.value)
		}
	}

	pattern.id = pattern.ptype*10000 + id
}

//检查是否是大牌
func (pattern *DaerPattern) checkBigCard() {
	if pattern.cards == nil || len(pattern.cards) <= 0 {
		logger.Error("checkBigCard:using before must init.")
	}

	var isBig = true
	for _, v := range pattern.cards {
		if !v.big {
			isBig = false
			break
		}
	}
	pattern.big = isBig
}

//检查是否是红牌
func (pattern *DaerPattern) checkRed() {
	if pattern.cards == nil || len(pattern.cards) <= 0 {
		logger.Error("checkRed:using before must init.")
	}

	var isRed = true
	for _, v := range pattern.cards {
		if !v.IsRed() {
			isRed = false
			break
		}
	}
	pattern.red = isRed
}

//计算此模式的权重
func (pattern *DaerPattern) calcWeight() {
	if pattern.cards == nil || len(pattern.cards) <= 0 {
		logger.Error("calcWeight:using before must init.")
	}

	pattern.weight = 1
}

//查询此模式的胡数
func (pattern *DaerPattern) value() int32 {
	if pattern.ptype == PTSingle || pattern.ptype == PTPair || pattern.ptype == PTUknown {
		return 0
	}

	if pattern.big {
		if pattern.red {
			return BigWordHuValue[0][pattern.ptype]
		} else {
			return BigWordHuValue[1][pattern.ptype]
		}
	} else {
		if pattern.red {
			return SmallWordHuValue[0][pattern.ptype]
		} else {
			return SmallWordHuValue[1][pattern.ptype]
		}
	}
}
