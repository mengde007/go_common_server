package majiangserver

import (
	cmn "common"
	"logger"
)

const (
	UnknowCardType = iota
	Tiao
	Tong
	Wan
	HongZhong
)

type MaJiangCard struct {
	id     int32
	value  int32
	cType  int32 // Tiao, Tong, Wan, HongZhong
	rcType int32 // Tiao, Tong, Wan
	flag   int32
	owner  *MaJiangPlayer
}

//新建一个卡牌
func NewCard(id int32, cType int32, value int32) *MaJiangCard {
	return &MaJiangCard{id: id, cType: cType, value: value}
}

func NewHongZhong() *MaJiangCard {
	return &MaJiangCard{id: 0, cType: HongZhong, value: 0}
}

func (card *MaJiangCard) CurValue() (cType, value int32) {
	if card.cType == HongZhong {
		if card.rcType == UnknowCardType {
			return HongZhong, 0
		} else {
			return card.rcType, card.value
		}
	} else {
		return card.cType, card.value
	}
}

//设置红中替代值
func (card *MaJiangCard) SetHZReplaceValue(cType, value int32) {
	if !card.IsHongZhong() {
		logger.Error("只有红中才能设置替代值")
		return
	}

	if card.flag&cmn.CLockHongZhongValue > 0 {
		logger.Error("红中值已经被锁定不能再进行更改了")
		return
	}

	card.rcType = cType
	card.value = value
}

//重置红中替代
func (card *MaJiangCard) ResetHZ() {
	if !card.IsHongZhong() {
		logger.Error("不是红中。")
		return
	}

	card.rcType = UnknowCardType
	card.value = 0
}

//是否是鬼牌
func (card *MaJiangCard) IsHongZhong() bool {
	return card.cType == HongZhong
}

//两张卡牌是否完全相同
func (card *MaJiangCard) IsFullEqual(c *MaJiangCard) bool {
	if c == nil {
		return false
	}

	return c.value == card.value && c.cType == card.cType && c.rcType == card.rcType
}

//两张卡牌是否完全相同
func (card *MaJiangCard) IsFullEqualByTypeAndValue(cType, val, rcType int32) bool {

	return val == card.value && cType == card.cType && rcType == card.rcType
}

//两卡牌是否相等
func (card *MaJiangCard) IsEqual(c *MaJiangCard) bool {
	if c == nil {
		return false
	}

	selfCType, selfVal := card.CurValue()
	cCType, cVal := c.CurValue()

	return selfCType == cCType && selfVal == cVal
}

//是否是等于这张牌
func (card *MaJiangCard) IsEqualByTypeAndValue(cType, val int32) bool {
	selfCType, selfVal := card.CurValue()

	return selfCType == cType && selfVal == val
}

func (card *MaJiangCard) IsSameHuaSe(c *MaJiangCard) bool {
	if c == nil {
		return false
	}

	selfCType, _ := card.CurValue()
	cCType, _ := c.CurValue()

	return selfCType == cCType
}

//是进牌还是出牌
func (card *MaJiangCard) IsIncomeCard() bool {
	return card.owner == nil
}

//是否是锁定出牌
func (card *MaJiangCard) IsLockChu() bool {
	return card.flag&cmn.CLock > 0
}

//是否是锁定值
func (card *MaJiangCard) IsLockValue() bool {
	return card.flag&cmn.CLockHongZhongValue > 0
}

//是否存在此牌
func IsExist(cards []*MaJiangCard, card *MaJiangCard) bool {
	for i := 0; i < len(cards); i++ {
		if cards[i].IsEqual(card) {
			return true
		}
	}

	return false
}
