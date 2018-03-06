package majiangserver

type PlayerAroundState struct {
	checkGangShangHuaCard *MaJiangCard   //用于识别是否杠上花
	checkGangShangPaoCard *MaJiangCard   //用于识别是否杠上炮
	buGangCard            *MaJiangCard   //用于识别是否抢杠
	buGangCardRemoved     bool           //补杠的牌是否从手上移除了
	huKe                  int32          //用于过水和升值胡（负数标示只能自摸胡牌了）
	guoPengGangCards      []*MaJiangCard //用于过碰的过水检查
	moCard                *MaJiangCard   //当前摸得牌，当玩家摸了一张牌后这个牌有动作需要执行，那么这个牌暂时不能放到玩家的手里，必须等玩家放弃执行动作
}

//新建一个卡牌
func NewPlayerAroundState() *PlayerAroundState {
	return &PlayerAroundState{}
}

func (self *PlayerAroundState) ClearAll() {
	self.checkGangShangHuaCard = nil
	self.checkGangShangPaoCard = nil
	self.buGangCard = nil
	self.buGangCardRemoved = false
	self.moCard = nil
	self.ClearGuoShuiAndShengZhiFlag(false)
}

func (self *PlayerAroundState) ClearGuoShuiAndShengZhiFlag(isBaoPai bool) {
	//报牌玩家是不能清除他胡的可数的
	if !isBaoPai {
		self.huKe = 0
	}

	self.guoPengGangCards = make([]*MaJiangCard, 0)
}

func (self *PlayerAroundState) IsOnlyZiMo() bool {
	return self.huKe < 0
}

func (self *PlayerAroundState) IsGuoShuiHu() bool {
	return self.huKe == 0
}

func (self *PlayerAroundState) IsShengZhiHu(newKe int32) bool {
	return self.huKe > 0 && self.huKe < newKe
}

func (self *PlayerAroundState) AddGuoShuiPengGangCard(card *MaJiangCard) {
	if card == nil {
		return
	}

	if self.guoPengGangCards == nil {
		self.guoPengGangCards = make([]*MaJiangCard, 0)
	}

	self.guoPengGangCards = append(self.guoPengGangCards, card)
}

func (self *PlayerAroundState) IsGuoShuiPengGang(card *MaJiangCard) bool {
	if self.guoPengGangCards == nil || len(self.guoPengGangCards) <= 0 || card == nil {
		return true
	} else {
		return !IsExist(self.guoPengGangCards, card)
	}
}

func (self *PlayerAroundState) AddGangFlag(card *MaJiangCard) {
	self.checkGangShangHuaCard = card
	self.checkGangShangPaoCard = card
}

func (self *PlayerAroundState) HaveGangShangHuaFlag() bool {
	return self.checkGangShangHuaCard != nil
}

func (self *PlayerAroundState) HaveGangShangPaoFlag() bool {
	return self.checkGangShangPaoCard != nil
}

func (self *PlayerAroundState) HaveBuGang() bool {
	return self.buGangCard != nil
}
