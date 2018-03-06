package daerserver

import (
	"logger"
)

type DaerPatternGroup struct {
	id       uint
	patterns []*DaerPattern
	kaoCards []*DaerCard
	huCards  []*DaerCard
}

//新建一个模式组
func NewPatternGroup(patterns []*DaerPattern) *DaerPatternGroup {
	if patterns == nil {
		logger.Error("NewPatternGroup:patterns is nil.")
		return nil
	}

	o := new(DaerPatternGroup)
	o.Init(patterns)

	return o
}

//初始化函数
func (group *DaerPatternGroup) Init(patterns []*DaerPattern) {
	if patterns == nil {
		logger.Error("Init:patterns is nil.")
		return
	}

	//赋值模式
	group.patterns = patterns
}

//产生模式组的ID
func (group *DaerPatternGroup) GenerateID() {
	if !group.CanHu() {
		logger.Error("ID必须由胡数和胡的牌产生")
		return
	}

	group.id = 0
	for _, v := range group.patterns {
		group.id += uint(v.value())
	}

	group.id *= 100000

	for _, v := range group.kaoCards {
		if v.big {
			group.id += 1001 * uint(v.value)
		} else {
			group.id += 101 * uint(v.value)
		}
	}

	for _, v := range group.huCards {
		if v.big {
			group.id += 10 * uint(v.value)
		} else {
			group.id += uint(v.value)
		}
	}
}

//比较两个模式是否相等，此函数可以再优化
func (group *DaerPatternGroup) IsEqual(g *DaerPatternGroup) bool {
	if g == nil {
		logger.Error("g is nil")
		return false
	}

	return group.id == g.id
}

//有没有对子
func (group *DaerPatternGroup) GetPairAmount() (result int32, card []*DaerCard) {
	card = make([]*DaerCard, 0)
	if group.patterns == nil {
		logger.Error("HavePair:group.patterns is nil")
		return
	}

	for _, v := range group.patterns {
		if v == nil {
			continue
		}

		if v.ptype == PTPair {
			result++
			card = append(card, v.cards[0])
		}
	}

	return
}

//模式组的胡数
func (group *DaerPatternGroup) Value() (sum int32) {
	if group.patterns == nil {
		logger.Error("Value:group.patterns is nil")
		return
	}

	sum = 0
	for _, v := range group.patterns {
		sum += v.value()
	}

	return
}

//胡指定牌的胡数
func (group *DaerPatternGroup) HuValue(huCard *DaerCard, controller *DaerController) (hu int32, huPatternType uint, ok bool) {
	if controller == nil {
		logger.Error("没有控制器不能计算当前的胡的值")
		return
	}

	if !group.CanHuSpecificCard(huCard) {
		//logger.Error("这个模式组不能胡此张牌：", huCard.value)
		return
	}

	kaoCards := group.kaoCards

	huPatternType = controller.CalcPatternType(kaoCards, huCard, group)
	if huPatternType == PTUknown {
		return
	}

	if huPatternType == PTZhao {
		kaoCards = []*DaerCard{huCard, huCard, huCard}
		// for i := 0; i < 4; i++ {
		// 	kaoCards = append(kaoCards, huCard)
		// }
	}

	huPattern := NewPattern(huPatternType, append(kaoCards, huCard))
	if huPattern == nil {
		return
	}

	ok = true
	hu = huPattern.value()

	return
}

//指定牌时候被包含在hu牌列表里
func (group *DaerPatternGroup) CanHuSpecificCard(card *DaerCard) bool {
	for _, c := range group.huCards {
		if c.IsEqual(card) {
			return true
		}
	}
	return false
}

//检查能否胡牌
func (group *DaerPatternGroup) CanHu() bool {
	return group.kaoCards != nil && len(group.kaoCards) > 0 && group.huCards != nil && len(group.huCards) > 0
}

//获取此模式组的红牌数量
func (group *DaerPatternGroup) GetRedCardAmount() int32 {
	var redCardAmount int32 = 0
	for _, pattern := range group.patterns {
		for _, card := range pattern.cards {
			if card.IsRed() {
				redCardAmount++
			}
		}
	}

	return redCardAmount
}

//获取指定模式的数量
func (group *DaerPatternGroup) GetPatternAmount(ptype uint) int32 {
	var patternAmount int32 = 0
	for _, pattern := range group.patterns {
		if ptype == pattern.ptype {
			patternAmount++
		}
	}

	return patternAmount
}
