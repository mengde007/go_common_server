package center

import (
	"common"
	//"github.com/garyburd/redigo/redis"
	"csvcfg"
	"logger"
	"path"
	"proto"
	// "pushmsg"
	"strconv"
	"strings"
	"time"
	"timer"
)

//读表结构
type PushCfg struct {
	TID        string
	StartTime  string
	EndTime    string
	StartHour  uint32
	StartMin   uint32
	StartYear  uint32
	StartMonth uint32
	StartDay   uint32
	WeekDay    string
	MsgTitle   string
	MsgContent string
}

//活动结构
type PushMsgConfig struct {
	TID        string
	StartTime  uint32
	EndTime    uint32
	StartHour  uint32
	StartMin   uint32
	StartYear  uint32
	StartMonth uint32
	StartDay   uint32
	WeekDay    [7]bool
	MsgTitle   string
	MsgContent string
}

var mapPushCfg map[uint32]*[]PushCfg
var MapPushMsgConfig map[uint32]*PushMsgConfig

func (self *Center) LoadPushConfig() {
	filename := path.Join(common.GetDesignerDir(), "PushMsg.csv")
	csvcfg.LoadCSVConfig(filename, &mapPushCfg)
}

func (self *Center) initMinTick() {
	//加载表
	self.LoadPushConfig()
	MapPushMsgConfig = make(map[uint32]*PushMsgConfig, 0)

	loc, _ := time.LoadLocation("Asia/Shanghai")
	const TimeFormat = "2006.1.2 15:4:5"

	for index, info := range mapPushCfg {
		cfg := &(*info)[0]
		pushCfg := &PushMsgConfig{
			TID:        cfg.TID,
			MsgTitle:   cfg.MsgTitle,
			MsgContent: cfg.MsgContent,
			StartHour:  cfg.StartHour,
			StartMin:   cfg.StartMin,
			StartYear:  cfg.StartYear,
			StartMonth: cfg.StartMonth,
			StartDay:   cfg.StartDay,
		}

		tcs, _ := time.ParseInLocation(TimeFormat, cfg.StartTime, loc)
		pushCfg.StartTime = uint32(tcs.Unix())

		tce, _ := time.ParseInLocation(TimeFormat, cfg.EndTime, loc)
		pushCfg.EndTime = uint32(tce.Unix())

		WeekDayArrStr := strings.Split(cfg.WeekDay, ",")
		for nindex := 0; nindex < 7; nindex++ {
			pushCfg.WeekDay[nindex] = false
		}
		if cfg.WeekDay != "" {
			for _, day := range WeekDayArrStr {
				nday, _ := strconv.Atoi(day) //0-6
				if nday > -1 && nday < 7 {
					pushCfg.WeekDay[nday] = true //设置成等于具体日子
				}
			}
		}
		MapPushMsgConfig[index] = pushCfg
	}

	//开始定时器
	self.everyMinTick = timer.NewTimer(time.Duration(1) * time.Second)
	self.everyMinTick.Start(
		func() {
			self.onMinrTick()
		},
	)
}

func (self *Center) onMinrTick() {
	self.everyMinTick.Stop()
	self.everyMinTick = nil

	//具体逻辑检查，是否需要推送
	for _, cfg := range MapPushMsgConfig {
		bIsNeedPush := self.CheakPushTime(cfg)
		if bIsNeedPush {
			logger.Info("Need to push")
			// go pushmsg.PushMsg2All(cfg.MsgTitle, cfg.MsgContent)
		}
	}

	//继续下次tick调用
	self.everyMinTick = timer.NewTimer(time.Duration(1) * time.Minute)
	self.everyMinTick.Start(
		func() {
			self.onMinrTick()
		},
	)
}

func (self *Center) CheakPushTime(info *PushMsgConfig) bool {
	gotime := time.Now()
	tnow := uint32(gotime.Unix())
	hourtemp, mintemp, _ := gotime.Clock()
	hour := uint32(hourtemp)
	min := uint32(mintemp)
	wdnow := time.Now().Weekday()
	year, month, day := gotime.Date()

	bIsOpend := false

	if tnow >= info.StartTime && tnow <= info.EndTime {
		//判断是否星期某日开放
		busewd := false
		bIsTheDay := false
		for nwd := 0; nwd < 7; nwd++ {
			if info.WeekDay[nwd] {
				busewd = true
				if int(wdnow) == nwd {
					bIsTheDay = true
				}
			}
		}
		if !busewd {
			bIsOpend = true
		}

		if busewd && !bIsTheDay {
			bIsOpend = false
		}

		if bIsTheDay {
			bIsOpend = true
		}
	}

	if !bIsOpend {
		return false
	}

	// logger.Info("info.StartYear = %d, info.StartMonth = %d, info.StartDay = %d, info.StartHour = %d, info.StartMin = %d", info.StartYear, info.StartMonth, info.StartDay, info.StartHour, info.StartMin)

	//这个是直发一次的，必须要有具体时间
	if year == int(info.StartYear) && int(month) == int(info.StartMonth) && day == int(info.StartDay) && hour == info.StartHour && min == info.StartMin {
		return true
	}

	//在某段时间内循环去发
	if info.StartYear == 0 && info.StartMonth == 0 && info.StartDay == 0 && hour == info.StartHour && min == info.StartMin {
		return true
	}

	return false
}

func (self *Center) PushMsg(req *proto.PushMsg, rst *proto.PushMsgResult) error {
	info := self.getPushCfg(req.ID)
	if info == nil {
		return nil
	}
	// go pushmsg.PushMsg2All(info.MsgTitle, info.MsgContent)
	return nil
}

func (self *Center) getPushCfg(ID int) *PushMsgConfig {
	if info, ok := MapPushMsgConfig[uint32(ID)]; ok {
		return info
	}
	return nil
}
