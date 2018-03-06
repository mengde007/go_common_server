package pockerserver

import (
	"fmt"
	"logger"
	// "math/rand"
	"rpc"
	"strconv"
	// "time"
)

const (
	CARD_NUM   = 7
	TARGET_NUM = 5
	AVALUE     = 14
)

const (
	HEART   = iota //红桃
	SPADE          //黑桃
	CLUB           //梅花
	DIAMOND        //方块
)

const (
	ROYAL_STRAIGHT_FLUSH = 10 //皇家同花顺
	STRAIGHT_FLUSH       = 9  //同花顺
	FOUR_KIND            = 8  //四条
	FULL_HOUSE           = 7  //葫芦
	FLUSH                = 6  //同花
	STRAIGHT             = 5  //顺子
	THREE_KIND           = 4  //三条
	TWO_PAIR             = 3  //两对
	ONE_PAIR             = 2  //一对
	HIGHT_CARD           = 1  //单张大牌
)

const (
	STATUS_READY        = iota //就绪
	STATUS_STAND               //站起
	STATUS_WATTING_JOIN        //加入等入
	STATUS_FOLD                //弃牌
	STATUS_THINKING            //思考中
	STATUS_CALL                //跟注
	STATUS_RAISE               //加注
	STATUS_ALLIN               //allin
	STATUS_CHEDK               //看牌
)

var COUNTDOWN_MAX int32

type pocker struct {
	eType int   //花色
	num   int32 //数字2~14
}

type pockerman struct {
	pockers    []pocker    //原手2张+公共牌
	combines   [][]pocker  //5张组合
	bestPocker int         //最好牌
	bestIndex  int         //最好牌组合索引
	combineNum int         //用于比较的编号
	chips      int32       //筹码总数
	drops      int32       //最近1次下注数量
	status     int         //当前状态
	posIndex   int32       //位置索引
	room       *PockerRoom //所在房间
	baseinfo   *rpc.PlayerBaseInfo
	waitFrom   int32 //等待开始时间
	autofold   bool  //是否自动弃牌
}

func (p *pockerman) rest_data() {
	p.pockers = []pocker{}
	p.combines = [][]pocker{}
	p.bestPocker = 0
	p.bestIndex = 0
	p.combineNum = int(0)
	p.drops = int32(0)
	if p.status != STATUS_STAND {
		p.status = STATUS_READY
	}

	p.waitFrom = int32(0)
	p.autofold = false
}

func (p *pockerman) rest_combine() {
	p.combines = [][]pocker{}
}

func (p *pockerman) Update(now int32) {
	if p.waitFrom == 0 {
		return
	}

	if now-p.waitFrom < COUNTDOWN_MAX {
		logger.Info("玩家：%s 倒计时:%d, 索引:%d,玩家：%s", p.baseinfo.GetUid(), now-p.waitFrom, p.posIndex, p.baseinfo.GetName())
		return
	}
	logger.Info("玩家弃牌：%s, 索引:%d", p.baseinfo.GetUid(), p.posIndex)

	p.status = STATUS_FOLD
	p.waitFrom = 0
	p.room.spkIndex = p.posIndex
	p.room.attends -= int32(1)
	p.autofold = true

	//通知其它玩家，老子弃牌了
	p.room.sync_s2c_status(p, ACT_FOLD)

	//check gameover
	if p.room.forceover() {
		// p.room.forcewinner = p.baseinfo.GetUid()
		p.room.force_over()
		return
	}
	if p.room.nextindex(p.posIndex) == -1 {
		p.room.over_undelay(p.room.rounds)
		return
	}
	p.room.nextplayer()
}

//对所有组合进行排序
func (p *pockerman) Sort() {
	for _, v := range p.combines {
		for i := 0; i < len(v); i++ {
			for j := i + 1; j < len(v); j++ {
				if v[i].num < v[j].num {
					tmp := v[j]
					v[j] = v[i]
					v[i] = tmp
				}
			}
		}
	}
}

//算牌的组合
func (p *pockerman) combine(index int) {
	tmp := []pocker{}
	for i := 0; i < len(p.pockers); i++ {
		if (index>>uint32(i))&1 == 1 {
			tmp = append(tmp, p.pockers[i])
		}
	}
	if len(tmp) == TARGET_NUM {
		p.combines = append(p.combines, tmp)

		//		strRst := ""
		//		for _, v := range tmp {
		//			strRst += fmt.Sprintf("%02d", v.num)

		//		}
		//		fmt.Println(strRst)
	}
}

func (p *pockerman) CalcCombine() {
	for i := 31; i < 1<<uint(len(p.pockers)); i++ {
		p.combine(i)
	}
}

//一组牌中，相同牌的数量
func (p *pockerman) CalcSecond(cards []pocker) (int, int, int) {
	four, three, two, tmp := 0, 0, 0, 1
	for i := 0; i < len(cards)-1; i++ {
		if cards[i].num == cards[i+1].num {
			tmp += 1
			if i+2 != len(cards) {
				continue
			}
		}
		switch tmp {
		case 2:
			two += 1
			break
		case 3:
			three = 1
			break
		case 4:
			four = 1
			break
		}
		tmp = 1
	}
	return four, three, two
}

func (p *pockerman) AddPocker(pk pocker) {
	p.pockers = append(p.pockers, pk)
}

//得到最大牌的值以及在组合中的索引
func (p *pockerman) GetMaxValue() (int, int) {
	maxIndex := 0
	maxValue := 0
	maxWeight := 0
	for index, v := range p.combines {
		value, weight := p.GetCardValue(v)
		if value > maxValue {
			maxIndex = index
			maxValue = value
			maxWeight = weight
		}
	}
	p.bestPocker = maxWeight
	p.bestIndex = maxIndex
	p.combineNum = int(maxValue)
	logger.Info("*******GetMaxValue组合数,combineNum:%d, bestPocker:%d", p.combineNum, p.bestPocker)

	return maxIndex, maxWeight
}

//计算一组牌的类型
func (p *pockerman) GetCardValue(cards []pocker) (int, int) {
	if cards == nil || len(cards) != 5 {
		//		logger.Error("GetCardValue param cards error")
		fmt.Println("GetCardValue param cards error")
		return 0, 0
	}

	weight := 0
	bFlush := p.is_flush(cards)
	bStraight := p.is_straight(cards)
	if bFlush && bStraight && cards[0].num == AVALUE { //皇家同花顺
		weight = ROYAL_STRAIGHT_FLUSH
	} else if bFlush && bStraight { //同花顺
		weight = STRAIGHT_FLUSH
	}
	if weight != 0 {
		return p.toNumbers(weight, cards), weight
	}

	four, three, two := p.CalcSecond(cards)
	//	fmt.Println("four, three, two", four, three, two)
	if four == 1 { //四条
		if cards[0] != cards[1] {
			p.swap(&cards, 0, 4)
		}
		weight = FOUR_KIND
	} else if three == 1 && two == 1 { //葫芦
		if cards[1] != cards[2] {
			p.swap(&cards, 0, 3)
			p.swap(&cards, 1, 4)
		}
		weight = FULL_HOUSE
	} else if bFlush { //同花
		weight = FLUSH
	} else if bStraight { //顺子
		weight = STRAIGHT
	} else if three == 1 { //三条
		if cards[0].num != cards[1].num {
			if cards[1].num == cards[2].num {
				p.swap(&cards, 0, 3)
			} else {
				p.swap(&cards, 0, 3)
				p.swap(&cards, 1, 4)
			}
		}
		weight = THREE_KIND
	} else if two == 2 { //两对
		if cards[1].num != cards[2].num && cards[2].num != cards[3].num {
			p.swap(&cards, 2, 4)
		} else if cards[0].num != cards[1].num {
			tmp := cards[0]
			cards = append(cards[:0], cards[1:]...)
			cards = append(cards, tmp)
		}
		weight = TWO_PAIR
	} else if two == 1 { //一对
		if cards[0].num != cards[1].num {
			index := 0
			for i := 0; i < len(cards)-1; i++ {
				if cards[i].num == cards[i+1].num {
					index = i
					break
				}
			}

			for i := 0; i < index; i++ {
				cards = append(cards, cards[i])
			}
			cards = append(cards[:0], cards[index:]...)
		}
		weight = ONE_PAIR
	} else { //单张大牌
		weight = HIGHT_CARD
	}

	return p.toNumbers(weight, cards), weight
}

func (p *pockerman) swap(cards *[]pocker, i, j int) {
	len := len(*cards)
	if len != TARGET_NUM || i >= len || j >= len {
		//		logger.Error("swap error,len:%d, i:%d, j:%d", len, i, j)
		fmt.Println("swap error,len:%d, i:%d, j:%d", len, i, j)
		return
	}

	tmp := (*cards)[i]
	(*cards)[i] = (*cards)[j]
	(*cards)[j] = tmp
}

func (p *pockerman) toNumbers(weight int, cards []pocker) int {
	strRst := strconv.Itoa(weight)
	for _, v := range cards {
		strRst += fmt.Sprintf("%02d", v.num)
	}
	rst, _ := strconv.Atoi(strRst)
	return rst
}

//true同花
func (p *pockerman) is_flush(cards []pocker) bool {
	for i := 0; i < len(cards)-1; i++ {
		if cards[i].eType != cards[i+1].eType {
			return false
		}
	}
	return true
}

//true顺子
func (p *pockerman) is_straight(cards []pocker) bool {
	for i := 0; i < len(cards)-1; i++ {
		if cards[i].num != cards[i+1].num+1 {
			return false
		}
	}
	return true
}

func (p *pockerman) Show( /*maxValue, index int*/ ) {
	logger.Info("\n所有牌")
	for _, v := range p.pockers {
		logger.Info("花:%d, 值:%d", v.eType, v.num)
	}

	logger.Info("\n组合数")
	for k, v := range p.combines {
		logger.Info("combine idx:%d num:%d", k, v)
	}

	// fmt.Println("max combines:", p.combines[index])
	// fmt.Println("number value:", maxValue)
}

func (p *pockerman) islittle_blind() bool {
	// index := p.room.dIndex
	// for i := 0; i < len(p.room.players); i++ {
	// 	index++
	// 	if index >= int32(len(p.room.players)) {
	// 		index = 0
	// 	}

	// 	cp := p.room.players[index]
	// 	if cp == nil {
	// 		continue
	// 	}

	// 	if cp.status == STATUS_FOLD || cp.status == STATUS_STAND  {
	// 		continue
	// 	}

	// 	if cp.baseinfo.GetUid() != p.baseinfo.GetUid() {
	// 		return false
	// 	}
	// 	return true
	// }
	if p.posIndex == p.room.smallblind {
		return true
	}

	return false
}

func (p *pockerman) call() bool {
	if p.drops > p.room.lastValue {
		logger.Error("玩家：%s跟注，p.drops:%d > p.room.lastValue:%d", p.baseinfo.GetName(), p.drops, p.room.lastValue)
		return false
	}

	need := p.room.lastValue - p.drops
	if p.chips < need {
		logger.Error("玩家：%s跟注，钱不够，has:%d, need:%d", p.baseinfo.GetName(), p.chips, need)
		return false
	}

	p.chips -= need
	p.drops += need

	p.status = STATUS_CALL
	if p.chips == int32(0) {
		p.status = STATUS_ALLIN
	}

	return true
}

func (p *pockerman) raise(value int32) bool {
	// inc := value - p.drops
	// if inc <= 0 {
	// 	logger.Error("加注，value:%d - p.drops:%d  <= 0", value, p.drops)
	// 	return false
	// }

	// if p.islittle_blind() {
	// 	if value

	// 	p.chips -= value
	// 	p.drops += value
	// 	return true
	// }

	if value <= p.drops {
		logger.Error("加注额度不能小于之前的，value:%d， p.drops:%d", value, p.drops)
		return false
	}

	need := value - p.drops
	if p.chips < need {
		logger.Error("钱不够，加注增量：%d,手上筹码：%d", need, p.chips)
		return false
	}

	p.chips -= need
	p.drops += need

	p.status = STATUS_RAISE
	if p.chips == int32(0) {
		p.status = STATUS_ALLIN
	}
	return true
}

func (p *pockerman) allin() bool {
	if p.chips <= 0 {
		logger.Error("钱都没的了，不能allin, chips:%d", p.chips)
		return false
	}

	p.drops += p.chips
	if p.chips > p.room.lastValue {
		p.room.lastValue = p.chips
		p.room.raiseIdx = p.posIndex
	}

	p.chips = 0
	p.status = STATUS_ALLIN
	return true
}
