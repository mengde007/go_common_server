package majiangserver

import (
	cmn "common"
	"logger"
	"math"
	"math/rand"
	"time"
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

var MeanwhileMaxHongZhongAmount = 6

type LicensingController struct {
	hongZhongCards []*MaJiangCard //红中牌
	normalCards    []*MaJiangCard // 其他普通牌

	hongZhongAmount int32 //红中的数量
}

//新建一个发牌控制器
func NewLicensingController(hongZhongAmount int32) *LicensingController {

	c := &LicensingController{}
	//hongZhongAmount = 8
	c.hongZhongAmount = hongZhongAmount
	//logger.Error("=============强行设置只有4个红中！！！！！！！！！！")

	//创建普通牌（条，筒，万）
	c.normalCards = make([]*MaJiangCard, 108)
	for typeVal := Tiao; typeVal <= Wan; typeVal++ {
		for cardVal := 0; cardVal < 9; cardVal++ {
			for cardAmount := 1; cardAmount <= 4; cardAmount++ {
				idVal := (typeVal-Tiao)*36 + cardVal*4 + cardAmount
				c.normalCards[idVal-1] = &MaJiangCard{id: int32(idVal), cType: int32(typeVal), value: int32(cardVal + 1), flag: cmn.CBack}
			}
		}
	}

	//创建红中牌
	c.hongZhongCards = make([]*MaJiangCard, hongZhongAmount)
	for i := 1; i <= int(hongZhongAmount); i++ {
		c.hongZhongCards[i-1] = &MaJiangCard{id: int32(108 + i), cType: int32(HongZhong), value: 0, flag: cmn.CBack}
	}

	logger.Info("发牌器构造完成后每种牌的数量信息：普通牌（%d）红中数量（%d）", len(c.normalCards), len(c.hongZhongCards))
	return c
}

//洗牌
func (self *LicensingController) Shuffle() {

	//打乱牌
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	i := 0

	totalCardAmount := self.TotalNormalAmount()
	randomTimes := int(float32(totalCardAmount) * 0.9)
	for i < randomTimes {
		index1 := r.Intn(int(totalCardAmount))
		index2 := r.Intn(int(totalCardAmount))

		temp := self.normalCards[index1]
		self.normalCards[index1] = self.normalCards[index2]
		self.normalCards[index2] = temp
		i++
	}

	//特殊洗牌-打乱最后一张
	// lastReplaceIndex := r.Intn(int(totalCardAmount))
	// temp := self.normalCards[lastReplaceIndex]
	// self.normalCards[lastReplaceIndex] = self.normalCards[len(self.normalCards)-1]
	// self.normalCards[len(self.normalCards)-1] = temp

	return
}

//发牌
func (self *LicensingController) Licensing(players []*MaJiangPlayer) {

	//是否启动了特殊发牌规则
	if SpecificLicensingType != LCNone {
		logger.Error("启用了特殊发牌======")
		self.SpecificLicensing(SpecificLicensingType, players)

		return
	}

	//检测玩家数量是否正确
	curPalyerAmount := len(players)
	if curPalyerAmount <= 0 || curPalyerAmount > RoomMaxPlayerAmount {
		logger.Error("玩家数量不对！")
		return
	}

	//给每个玩家发送常规的牌和有概率的发送红中
	normalCardCount := FirstCardsAmount - MeanwhileMaxHongZhongAmount
	for _, player := range players {
		if player == nil {
			continue
		}

		playerCards := make([]*MaJiangCard, normalCardCount)
		copy(playerCards, self.normalCards[:normalCardCount])
		self.normalCards = self.normalCards[normalCardCount:]

		hongZhongAmount := self.CalcFirstHongZhongAmount()
		playerCards = append(playerCards, self.OpenHongZhongCard(hongZhongAmount)...)
		normalCardAmount := int32(MeanwhileMaxHongZhongAmount) - hongZhongAmount
		playerCards = append(playerCards, self.OpenNormalCard(normalCardAmount)...)

		player.Compose(playerCards)

		logger.Info("发完一个玩家的牌后，发的红中数量:（%d）普通牌的数量：(%d), 座面上还剩下几张牌：(%d)", hongZhongAmount, normalCardAmount, self.RemainCardAmount())
	}

}

//首次红中数量
//var fixedProbability = [4]int32{50, 75, 95, 100}
var fixedProbability = [4]int32{20, 50, 80, 100}

func (self *LicensingController) CalcFirstHongZhongAmount() int32 {
	ownHZAmount := len(self.hongZhongCards)
	if ownHZAmount <= 0 {
		return 0
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randProbability := int32(r.Intn(100))
	for hzAmount, randVal := range fixedProbability {
		if randProbability < randVal {
			return int32(math.Min(float64(hzAmount+1), float64(ownHZAmount)))
		}
	}

	return 0
}

//开一张牌
func (self *LicensingController) OpenOneCard(player *MaJiangPlayer) *MaJiangCard {

	//检查输入参数
	if player == nil {
		logger.Error("player is null.")
		return nil
	}

	//检查是否有机会开出红中
	normalCardAmount := len(self.normalCards)
	hongZhongAmount := len(self.hongZhongCards)
	hongZhongAmountInHand := player.GetHongZhongAmountInHand()
	if hongZhongAmountInHand >= int32(MeanwhileMaxHongZhongAmount) {
		if normalCardAmount <= 0 {
			if hongZhongAmount > 0 {
				return self.OpenHongZhongCard(1)[0]
			}
		} else {
			return self.OpenNormalCard(1)[0]
		}
	} else {
		if normalCardAmount > 0 && hongZhongAmount > 0 {
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			isOpenHongZhong := r.Intn(normalCardAmount+hongZhongAmount) > normalCardAmount
			if isOpenHongZhong {
				return self.OpenHongZhongCard(1)[0]
			} else {
				return self.OpenNormalCard(1)[0]
			}
		} else if normalCardAmount > 0 {
			return self.OpenNormalCard(1)[0]
		} else if hongZhongAmount > 0 {
			return self.OpenHongZhongCard(1)[0]
		} else {
			logger.Info("桌面上已经没有牌了........")
		}
	}

	return nil
}

//开几张普通的牌
func (self *LicensingController) OpenNormalCard(amount int32) []*MaJiangCard {
	result := make([]*MaJiangCard, 0)

	if amount <= 0 {
		return result
	}

	if amount > int32(len(self.normalCards)) {
		logger.Error("剩余牌的数量太少，不能够开这么多张牌出来: ", amount)
		return result
	}

	remainCardAmount := int32(len(self.normalCards))
	result = append(result, self.normalCards[remainCardAmount-amount:remainCardAmount]...)
	self.normalCards = self.normalCards[:remainCardAmount-amount]

	return result
}

//开几张红中牌
func (self *LicensingController) OpenHongZhongCard(amount int32) []*MaJiangCard {

	result := make([]*MaJiangCard, 0)

	if amount <= 0 {
		return result
	}

	if amount > int32(len(self.hongZhongCards)) {
		logger.Error("剩余红中数量太少，不能够开这么多张红中出来: ", amount)
		return result
	}

	result = append(result, self.hongZhongCards[:amount]...)
	self.hongZhongCards = self.hongZhongCards[amount:]

	return result
}

//总的牌的数量
func (self *LicensingController) TotalCardAmount() int32 {
	return self.TotalNormalAmount() + self.hongZhongAmount
}

func (self *LicensingController) TotalNormalAmount() int32 {
	return 108
}

//总的红中数量
func (self *LicensingController) TotalHongZhongAmount() int32 {
	return self.hongZhongAmount
}

//剩余红中数量
func (self *LicensingController) RemainHongZhongAmount() int32 {
	return int32(len(self.hongZhongCards))
}

//剩余的牌数量
func (self *LicensingController) RemainCardAmount() int32 {
	return int32(len(self.hongZhongCards) + len(self.normalCards))
}

//特殊发牌
func (self *LicensingController) SpecificLicensing(ctype int, players []*MaJiangPlayer) {

	switch ctype {
	case HuPengSimultaneously:
		//return self.FHuPengSimultaneously(players)
		logger.Error("未实现：", ctype)
	case TestHu:
		self.FTestHu(players)
	case TestErLongTouYi:
		//return self.FErLongTouYi(players)
		logger.Error("未实现：", ctype)
	default:
		logger.Error("没有此发牌类型：", ctype)
		return
	}

	return
}

//测试模式能胡不
func (self *LicensingController) FTestHu(players []*MaJiangPlayer) {

	if len(players) != 4 {
		logger.Error("玩家必须是4个")
		return
	}

	// self.FillHandCard(players[0], []*MaJiangCard{
	// 	&MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	// 	// &MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	// 	// &MaJiangCard{value: 2, cType: Tiao, rcType: UnknowCardType},
	// 	// &MaJiangCard{value: 2, cType: Tiao, rcType: UnknowCardType},
	// })

	//case1
	// self.FillHandCard(players[0], []*MaJiangCard{
	// 	&MaJiangCard{value: 7, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 3, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 5, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 8, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Tong, rcType: UnknowCardType},
	// })

	// self.FillHandCard(players[1], []*MaJiangCard{
	// 	&MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 6, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 6, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Wan, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Wan, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Wan, rcType: UnknowCardType},
	// })

	// self.FillHandCard(players[2], []*MaJiangCard{
	// 	&MaJiangCard{value: 6, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 5, cType: Wan, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 5, cType: Wan, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 2, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 2, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Wan, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 2, cType: Wan, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 3, cType: Wan, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Wan, rcType: UnknowCardType},
	// })

	// self.FillHandCard(players[0], []*MaJiangCard{
	// 	&MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 2, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 3, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 6, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 6, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
	// })

	//case1 -硬通叫
	// self.FillHandCard(players[0], []*MaJiangCard{
	// 	&MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 2, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 3, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 4, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 5, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 6, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 7, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 8, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tiao, rcType: UnknowCardType},
	// 	&MaJiangCard{value: 9, cType: Tiao, rcType: UnknowCardType},
	// })

	//case1 -带红中通叫
	self.FillHandCard(players[0], []*MaJiangCard{
		&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
		&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
		&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
		&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
		&MaJiangCard{value: 0, cType: HongZhong, rcType: UnknowCardType},
		&MaJiangCard{value: 8, cType: Tiao, rcType: UnknowCardType},
		&MaJiangCard{value: 9, cType: Tiao, rcType: UnknowCardType},
		&MaJiangCard{value: 9, cType: Tiao, rcType: UnknowCardType},
		&MaJiangCard{value: 5, cType: Tiao, rcType: UnknowCardType},
		&MaJiangCard{value: 8, cType: Tiao, rcType: UnknowCardType},
		&MaJiangCard{value: 5, cType: Tiao, rcType: UnknowCardType},
		&MaJiangCard{value: 9, cType: Tiao, rcType: UnknowCardType},
		&MaJiangCard{value: 1, cType: Tiao, rcType: UnknowCardType},
	})

	self.FillHandCard(players[1], []*MaJiangCard{
		&MaJiangCard{value: 6, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 6, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 6, cType: Wan, rcType: UnknowCardType},

		&MaJiangCard{value: 2, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 2, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 2, cType: Wan, rcType: UnknowCardType},

		&MaJiangCard{value: 4, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 4, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 4, cType: Wan, rcType: UnknowCardType},

		&MaJiangCard{value: 5, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 5, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 5, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 5, cType: Tong, rcType: UnknowCardType},
	})

	self.FillHandCard(players[2], []*MaJiangCard{
		&MaJiangCard{value: 6, cType: Wan, rcType: UnknowCardType},

		&MaJiangCard{value: 5, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 5, cType: Wan, rcType: UnknowCardType},

		&MaJiangCard{value: 2, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 2, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 2, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 2, cType: Tong, rcType: UnknowCardType},

		&MaJiangCard{value: 4, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 4, cType: Tong, rcType: UnknowCardType},

		&MaJiangCard{value: 3, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 3, cType: Tong, rcType: UnknowCardType},

		&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},
	})

	self.FillHandCard(players[3], []*MaJiangCard{
		&MaJiangCard{value: 6, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 6, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 6, cType: Tong, rcType: UnknowCardType},

		&MaJiangCard{value: 7, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 7, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 7, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 7, cType: Tong, rcType: UnknowCardType},

		&MaJiangCard{value: 8, cType: Tong, rcType: UnknowCardType},

		&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},
		&MaJiangCard{value: 9, cType: Tong, rcType: UnknowCardType},

		&MaJiangCard{value: 9, cType: Wan, rcType: UnknowCardType},
		&MaJiangCard{value: 9, cType: Wan, rcType: UnknowCardType},
	})

	//self.normalCards[len(self.normalCards)-1] = self.OpenHongZhongCard(1)[0]
	// self.normalCards[len(self.normalCards)-1].cType = Wan
	// self.normalCards[len(self.normalCards)-1].value = 6

	// self.normalCards[len(self.normalCards)-2].cType = Wan
	// self.normalCards[len(self.normalCards)-2].value = 6

	// self.normalCards[len(self.normalCards)-3].cType = Wan
	// self.normalCards[len(self.normalCards)-3].value = 6

	// self.normalCards[len(self.normalCards)-4].cType = Wan
	// self.normalCards[len(self.normalCards)-4].value = 6

	// self.normalCards[len(self.normalCards)-5].cType = Wan
	// self.normalCards[len(self.normalCards)-5].value = 6

	// self.normalCards[len(self.normalCards)-6].cType = Wan
	// self.normalCards[len(self.normalCards)-6].value = 6

	// self.normalCards[len(self.normalCards)-7].cType = Wan
	// self.normalCards[len(self.normalCards)-7].value = 6

	// self.normalCards[len(self.normalCards)-8].cType = Wan
	// self.normalCards[len(self.normalCards)-8].value = 6

	// self.normalCards[len(self.normalCards)-9].cType = Wan
	// self.normalCards[len(self.normalCards)-9].value = 6

	// self.normalCards[len(self.normalCards)-10].cType = Wan
	// self.normalCards[len(self.normalCards)-10].value = 6

	// self.normalCards[len(self.normalCards)-11].cType = Wan
	// self.normalCards[len(self.normalCards)-11].value = 6

	return
}

//填充一个人的手牌
func (self *LicensingController) FillHandCard(player *MaJiangPlayer, cards []*MaJiangCard) {

	if player == nil || cards == nil {
		logger.Error("player or cards is nil.")
		return
	}

	result := make([]*MaJiangCard, 0)

	for _, c := range cards {
		if c.IsHongZhong() {
			if len(self.hongZhongCards) <= 0 {
				logger.Error("没有红中可以发了。。。")
				break
			}
			result = append(result, self.hongZhongCards[len(self.hongZhongCards)-1])
			self.hongZhongCards = self.hongZhongCards[0 : len(self.hongZhongCards)-1]
		} else {
			isFind := false
			for i, nc := range self.normalCards {
				if c.IsEqual(nc) {
					result = append(result, self.normalCards[i])
					self.normalCards = append(self.normalCards[:i], self.normalCards[i+1:]...)
					isFind = true
					break
				}
			}
			if !isFind {
				logger.Error("没有此牌了：", ConvertToWord(c))
			}
		}
	}

	if len(result) != FirstCardsAmount {
		logger.Error("特殊发牌出错！必须要有：%d张", FirstCardsAmount)
		//return
	}

	player.Compose(result)
}
