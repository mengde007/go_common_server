package common

import (
	"logger"
	"rpc"
	"strconv"
	"strings"
	"time"
)

type TimerCallback func(data interface{})

type RoomTimer struct {
	timerStartTime int64         //定时器的开始时间
	timerTime      int64         //定时的时间，时间到后执行回调
	callback       TimerCallback //回调
	data           interface{}   //附加数据
}

type TimerMgr struct {
	timerList    map[string]*RoomTimer //定时器列表
	curTime      time.Duration         //当前的时间
	timeInterval time.Duration         //更新间隔
}

func NewTimerMgr() *TimerMgr {
	return &TimerMgr{make(map[string]*RoomTimer, 0), 0, 0}
}

//启动计时
func (self *TimerMgr) StartTimer(name string, delay int64, call TimerCallback, data interface{}) {
	if name == "" {
		return
	}

	if timer, exist := self.timerList[name]; !exist {
		timerS := &RoomTimer{
			timerStartTime: time.Now().Unix(),
			timerTime:      delay,
			callback:       call,
			data:           data,
		}

		self.timerList[name] = timerS
	} else {
		timer.timerStartTime = time.Now().Unix()
		timer.timerTime = delay
		timer.callback = call
		timer.data = data
	}
}

//获取剩余倒计时
func (self *TimerMgr) GetRemainTime(name string) int32 {
	if name == "" {
		return 0
	}

	if !self.IsTiming(name) {
		return 0
	}

	timer, exist := self.timerList[name]
	if !exist {
		logger.Error("不存在定时器：", name)
		return 0
	}

	return int32(timer.timerTime - (time.Now().Unix() - timer.timerStartTime))
}

//是否在计时
func (self *TimerMgr) IsTiming(name string) bool {
	if name == "" {
		return false
	}

	timer, exist := self.timerList[name]
	if !exist {
		return false
	}

	return timer.timerStartTime > 0
}

//停止计时器
func (self *TimerMgr) StopTimer(name string) {
	if name == "" {
		return
	}

	timer, exist := self.timerList[name]
	if !exist {
		//logger.Error("不存在此定时器：", name)
		return
	}

	timer.timerStartTime = 0
	timer.timerTime = 0
	timer.callback = nil
	timer.data = nil
}

//停止所有的计时器
func (self *TimerMgr) StopAllTimer() {
	if self.timerList == nil {
		return
	}

	for _, timer := range self.timerList {
		if timer == nil {
			continue
		}

		timer.timerStartTime = 0
		timer.timerTime = 0
		timer.callback = nil
		timer.data = nil
	}
}

//清除计时器
func (self *TimerMgr) Clear() {
	self.timerList = make(map[string]*RoomTimer, 0) //清空定时器列表
}

//设置更新间隔
func (self *TimerMgr) SetUpdateInterval(interval time.Duration) {
	self.timeInterval = interval
	self.curTime = 0
}

//更新函数
func (self *TimerMgr) Update(ft time.Duration) {

	//检查更新间隔是否到
	if self.timeInterval > 0 {
		self.curTime += ft
		if self.curTime < self.timeInterval {
			return
		}

		self.curTime = 0
	}

	//更新定时器列表
	for name, timer := range self.timerList {
		if timer == nil || timer.timerStartTime <= 0 {
			continue
		}

		elapsedTime := time.Now().Unix() - timer.timerStartTime
		logger.Info("倒计时名称：%s  延迟倒计时:%s", name, elapsedTime)
		if elapsedTime >= timer.timerTime && timer.callback != nil {
			timer.timerStartTime = 0
			timer.timerTime = 0
			timer.callback(timer.data)
			//deletedTimers = append(deletedTimers, name)
		}
	}
}

func CheckCoin(roomType int32, info *rpc.PlayerBaseInfo) (ok bool, code int32) {
	//检查输入参数
	if info == nil {
		return
	}

	//检测金币是否足够
	cfg := GetDaerRoomConfig(strconv.Itoa(int(roomType)))
	if cfg == nil {
		logger.Error("读取配置表出错")
		return
	}

	curCoin := info.GetCoin()
	logger.Info("进入游戏检查金币,房间ID：%d 拥有的%s, 下限%s, 上限%s", roomType, curCoin, cfg.MinLimit, cfg.MaxLimit)
	if curCoin < cfg.MinLimit {
		ok = false
		code = ERLessCoin
		return
	}

	if cfg.MaxLimit > 0 && curCoin > cfg.MaxLimit {
		ok = false
		code = ERReachUpLimit
		return
	}

	return true, 0
}

func CheckPockerCoin(roomType int32, info *rpc.PlayerBaseInfo) (ok bool, code int32) {
	//检查输入参数
	if info == nil {
		return
	}

	//检测金币是否足够
	cfg := GetDaerRoomConfig(strconv.Itoa(int(roomType)))
	if cfg == nil {
		logger.Error("读取配置表出错")
		return
	}

	curCoin := info.GetCoin()
	logger.Info("进入游戏检查金币,房间ID：%d 拥有的%s, 下限%s, 上限%s", roomType, curCoin, cfg.MinLimit, cfg.MaxLimit)
	if curCoin < cfg.MinLimit {
		ok = false
		code = int32(1)
		return
	}

	if cfg.MaxLimit > 0 && curCoin > cfg.MaxLimit {
		ok = false
		code = int32(2)
		return
	}
	return true, 0
}

func CheckCustomPockerCoin(blindId, limId int32, info *rpc.PlayerBaseInfo) (bool, int32) {
	logger.Info("********************CheckCustomPockerCoin0")
	//检查输入参数
	if info == nil {
		return false, 0
	}

	//进入上下限
	cfg := GetDaerGlobalConfig(strconv.Itoa(int(limId)))
	if cfg == nil {
		logger.Fatal("GetDaerGlobalConfig(:%d) return nil", limId)
		return false, 0
	}
	arrs := strings.Split(cfg.StringValue, "_")
	if len(arrs) != 2 {
		logger.Error("NewCustomPockerRoom cfg.StringValue:%s , limId:%d err", cfg.StringValue, limId)
		return false, 0
	}
	sbValue, _ := strconv.Atoi(arrs[0])
	bigValue, _ := strconv.Atoi(arrs[1])

	curCoin := int(info.GetCoin())
	logger.Info("进入游戏检查金币,拥有的%s, 下限%s, 上限%s", curCoin, sbValue, bigValue)
	if curCoin < sbValue {
		return false, int32(1)
	}
	logger.Info("****************CheckCustomPockerCoin1")
	if bigValue > 0 && curCoin > bigValue {
		return false, int32(2)
	}
	return true, 0
}
