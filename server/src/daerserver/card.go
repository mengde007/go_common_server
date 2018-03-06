package daerserver

import (
	cmn "common"
)

type DaerCard struct {
	id    int32
	value int32
	big   bool
	flag  int32
	owner *DaerPlayer
}

//新建一个卡牌
func NewCard(id int32, value int32, isBig bool) *DaerCard {
	return &DaerCard{id: id, value: value, big: isBig}
}

//是否是红色牌
func (card *DaerCard) IsRed() bool {
	return card.value == 2 || card.value == 7 || card.value == 10
}

//两卡牌是否相等
func (card *DaerCard) IsEqual(c *DaerCard) bool {
	if c == nil {
		return false
	}
	return card.value == c.value && card.big == c.big
}

//是进牌还是出牌
func (card *DaerCard) IsIncomeCard() bool {
	return card.owner == nil
}

//是否是锁定的
func (card *DaerCard) IsLock() bool {
	return card.flag&cmn.CLock > 0
}
