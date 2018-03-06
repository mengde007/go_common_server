package matchserver

import (
	conn "centerclient"
	cmn "common"
	"connector"
	"fmt"
	"logger"
	"math"
	"strconv"
	"strings"
	"time"
)

//奖励项
type RewardItem struct {
	rank   int32  //排名
	id     string //道具ID
	amount int32  //道具数量
}

type Match struct {
	id            int32        //比赛ID
	startEntrance bool         //开始入场了吗
	countdown     int64        //倒计时
	watingplayers []cmn.Player //等待比赛开始的玩家
	enrollPlayers []cmn.Player //报名的玩家
	isValid       bool         //是否有效的
	isRemoved     bool         //是否被移除
	cmn.MatchCfg               //配置表数据

	startTime *time.Time            //缓存比赛的开始时间-不用每次都去解析时间
	endTime   *time.Time            //缓存比赛的结算时间
	rewards   map[int32]*RewardItem //奖励的
}

func NewMatch(cfg cmn.MatchCfg) *Match {
	match := &Match{}

	match.Init(cfg)

	return match

}

func (self *Match) Init(cfg cmn.MatchCfg) {
	self.id = cfg.ID

	self.MatchCfg = cfg
	self.isValid = true
	self.isRemoved = false

	self.InitTime()

	self.InitRewards()

	self.Reset()
}

func (self *Match) InitTime() {
	switch self.MatchCfg.StartMatchMode {
	case EverydayFixedTimeMode:
		fallthrough
	case EverydayIntervalMode:
		curTime := time.Now()
		//获取开始时间
		startTimeStr := fmt.Sprintf("%d-%d-%d %s", curTime.Year(), curTime.Month(), curTime.Day(), self.MatchCfg.StartTime)
		startTime, serr := time.Parse("2006-1-2 15:04:05", startTimeStr)
		if serr != nil {
			logger.Error("读取比赛配置表中的开始时间配置错误,格式必须是15:04:05：", serr)
			return
		}

		if self.MatchCfg.EndTime != "" {
			endTimeStr := fmt.Sprintf("%d-%d-%d %s", curTime.Year(), curTime.Month(), curTime.Day(), self.MatchCfg.EndTime)
			endTime, eerr := time.Parse("2006-1-2 15:04:05", endTimeStr)
			if eerr != nil {
				logger.Error("读取比赛配置表中的开始时间配置错误,格式必须是15:04:05：", eerr)
				return
			}

			self.endTime = &endTime
		}

		self.startTime = &startTime
	case FixedTimeMode:

		startTime, serr := time.Parse("2006-1-2 15:04:05", self.MatchCfg.StartTime)
		if serr != nil {
			logger.Error("读取比赛配置表中的开始时间配置错误,格式必须是15:04:05：", serr)
			return
		}

		self.startTime = &startTime

	case FullStartMode:

	default:
		logger.Error("未知的比赛开始类型")
	}
}

func (self *Match) InitRewards() {
	self.rewards = make(map[int32]*RewardItem, 0)

	rewardStrs := strings.Split(self.MatchCfg.Reward, "#")
	for _, rewardStr := range rewardStrs {
		rewardInfo := strings.Split(rewardStr, "_")
		if len(rewardInfo) != 3 {
			logger.Error("在读取比赛配置表时，发现奖励字段（Reward）的格式不对，必须是：名次_道具ID_道具数量#名次_道具ID_道具数量")
			continue
		}

		rank, err := strconv.Atoi(rewardInfo[0])
		if err != nil {
			logger.Error("装换奖励信息出错: ", err)
			continue
		}

		amount, err := strconv.Atoi(rewardInfo[2])
		if err != nil {
			logger.Error("装换奖励信息出错: ", err)
			continue
		}

		itemInfo := &RewardItem{}
		itemInfo.rank = int32(rank)
		itemInfo.id = rewardInfo[1]
		itemInfo.amount = int32(amount)

		self.rewards[itemInfo.rank] = itemInfo
	}
}

func (self *Match) Reset() {
	self.startEntrance = false
	self.countdown = math.MaxInt64

	self.watingplayers = make([]cmn.Player, 0)
	self.enrollPlayers = make([]cmn.Player, 0)
}

func (self *Match) AddPlayer(p cmn.Player) {
	if p == nil {
		logger.Error("p is nil")
		return
	}

	pInfo := p.GetPlayerBasicInfo()
	if pInfo == nil {
		logger.Error("Don't exist PlayerInfo.")
		return
	}

	pInfo.SetCoin(self.MatchCfg.InitCredit)

	if self.startEntrance {
		self.watingplayers = append(self.watingplayers, p)
		self.enrollPlayers = self.DeletePlayer(self.enrollPlayers, p.ID())
	} else {
		self.enrollPlayers = append(self.enrollPlayers, p)
	}

	//扣钱
	self.ModifyMoneny(p.ID(), true)

}

func (self *Match) DeletePlayer(list []cmn.Player, uid string) []cmn.Player {
	if list == nil {
		return nil
	}

	for i, p := range list {
		if p == nil {
			continue
		}

		if p.ID() == uid {
			return append(list[:i], list[i+1:]...)
		}
	}

	return list
}

func (self *Match) RemovePlayer(uid string) {
	if self.startEntrance {
		self.DeletePlayer(self.watingplayers, uid)
	} else {
		self.DeletePlayer(self.enrollPlayers, uid)
	}
}

func (self *Match) GetWatingAmount() int32 {
	return int32(len(self.watingplayers))
}

func (self *Match) GetEnrollAmount() int32 {
	return int32(len(self.enrollPlayers))
}

//一秒更新一下
func (self *Match) Update() {
	//检查是否已经被移除
	if self.isRemoved {
		return
	}

	//如果准备入场时，倒计时减
	if self.startEntrance {
		self.countdown -= 1

		if self.CanStartMatch() {
			self.StartMatch()
			self.isValid = !self.IsOnceMatchPerDay()
			self.isRemoved = self.IsOnceMatch()
		}

	} else {
		self.CheckIsStartEntrance()
	}
}

func (self *Match) CheckIsStartEntrance() {

	switch self.MatchCfg.StartMatchMode {
	case EverydayIntervalMode:
		//检查是否要更新开始时间（隔天）
		curTime := time.Now()
		if !IsSameDay(self.startTime, &curTime) {
			self.InitTime()
		}

		//检查是否可以开启一轮新的比赛报名
		if curTime.Unix() >= self.startTime.Unix() && curTime.Unix() < self.endTime.Unix() {
			//距离下一场开始还剩多少秒
			remainTime := int64(self.MatchCfg.StartMatchInterval*60) -
				((curTime.Unix() - self.startTime.Unix()) % (int64(self.MatchCfg.StartMatchInterval) * 60))

			//如果距离开始的时间小于入场的时间（s）
			if remainTime <= int64(self.MatchCfg.EntryThreshold)*60 {
				self.startEntrance = true
				self.countdown = int64(self.MatchCfg.EntryThreshold * 60)
			}
		}

	case FixedTimeMode:
		curTime := time.Now()
		if curTime.Unix() >= self.startTime.Unix() {
			self.startEntrance = true
		}

	case EverydayFixedTimeMode:
		//检查是否要更新开始时间（隔天）
		curTime := time.Now()
		if !IsSameDay(self.startTime, &curTime) {
			self.InitTime()
			self.isValid = true

		}

		//每天只有一场
		if curTime.Unix() >= self.startTime.Unix() {
			self.startEntrance = true
		}

	case FullStartMode:

		self.startEntrance = true
	default:
		logger.Error("未知的比赛开始类型")
	}
}

//是不是一次性的比赛
func (self *Match) IsOnceMatch() bool {
	return self.MatchCfg.StartMatchMode == FixedTimeMode
}

func (self *Match) IsOnceMatchPerDay() bool {
	return self.MatchCfg.StartMatchMode == EverydayFixedTimeMode
}

func IsSameDay(t1 *time.Time, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}

	if t1 == nil || t2 == nil {
		return false
	}

	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}

//修改玩家钱
func (self *Match) ModifyMoneny(uid string, isDeduct bool) {
	//确定是扣钱还是归还钱
	var symbol int32 = 1
	if isDeduct {
		symbol = -1
	}

	//通知gameserver扣钱
	switch self.MatchCfg.EntryFeeCurrencyType {
	case EntranceCurrencyFree:
		return
	case EntranceCurrencyCoin:
		if err := conn.SendCostResourceMsg(uid, connector.RES_COIN, cmn.GameTypeName[self.MatchCfg.GameType], symbol*self.MatchCfg.EntryFee); err != nil {
			logger.Error("发送扣取金币出错：", err, uid)
			return
		}
	case EntranceCurrencyGem:
		if err := conn.SendCostResourceMsg(uid, connector.RES_GEM, cmn.GameTypeName[self.MatchCfg.GameType], symbol*self.MatchCfg.EntryFee); err != nil {
			logger.Error("发送扣取金币出错：", err, uid)
			return
		}
	default:
		logger.Error("未知的入场货币类型")
	}
}

func (self *Match) CanStartMatch() bool {
	if self.MatchCfg.StartMatchMode == FullStartMode {
		if self.GetWatingAmount() >= self.MatchCfg.FullStartCount {
			return true
		}
		return false
	} else {
		return self.countdown <= 0
	}
}

func (self *Match) StartMatch() {
	//扣除未进入游戏等待区的玩家的金币
	if self.MatchCfg.IsGiveBackCoin > 0 {
		for _, p := range self.enrollPlayers {
			if p == nil {
				continue
			}

			self.ModifyMoneny(p.ID(), false)
		}
	}

	switch self.MatchCfg.StartMatchMode {
	case EverydayIntervalMode:
		fallthrough
	case FixedTimeMode:
		fallthrough
	case EverydayFixedTimeMode:
		self.Reset()

	case FullStartMode:
		//self.watingplayers = append(self.watingplayers[0:3])

	default:
		logger.Error("未知的比赛开始类型")
	}

}
