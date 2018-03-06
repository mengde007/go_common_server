package center

import (
	"common"
	"github.com/garyburd/redigo/redis"
	"logger"
	"proto"
	"time"
	"timer"
)

func (self *Center) initDayTick() {
	//首先计算现在的时间到晚上0点时间的间隔，注册一个时间tick
	t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 24, 0, 0, 0, time.Local)
	durtime := t.Unix() - time.Now().Unix()
	//到今晚零点的时间加上剩余的时间
	leftHour := common.GetGlobalConfig("FRIEND_HOUR")
	allTime := durtime + int64(leftHour*60*60)
	//logger.Info("现在的秒数   24点得临界秒数   看间隔的秒数 allTime ", time.Now().Unix(), t.Unix(), allTime)

	//开始定时器
	self.everydaytime = timer.NewTimer(time.Duration(allTime) * time.Second)
	self.everydaytime.Start(
		func() {
			self.onDayTick()
		},
	)
}

func (self *Center) onDayTick() {
	self.everydaytime.Stop()
	self.everydaytime = nil

	whichDay := common.GetGlobalConfig("FRIEND_WEEK")

	if time.Now().Weekday() == time.Weekday(whichDay) {
		//这里调用重命名
		self.reNameMap()
	}

	//重新注册每天的tick
	self.everydaytime = timer.NewTimer(time.Duration(24*60*60) * time.Second)
	self.everydaytime.Start(
		func() {
			self.onDayTick()
		},
	)
}

func (self *Center) reNameMap() error {
	err := self.hreName(common.UserTableName, common.PlayerTrophyMapName_New, common.PlayerTrophyMapName_Old)
	if err != nil {
		logger.Error("rename error : %s.(%s to %s)", err.Error(), common.PlayerTrophyMapName_New, common.PlayerTrophyMapName_Old)
	}

	return nil
}

func (self *Center) getCurPlayerValue(playerID string) (number int, err error) {
	number, err = self.hgetValue(common.UserTableName, common.PlayerTrophyMapName_New, playerID)
	if err != nil {
		number = 0

		return
	}

	return
}

func (self *Center) GetOldPlayerValue(req *proto.PlayerLastTrophy, rst *proto.PlayerLastTrophyResult) error {
	cache := self.maincache.Get()
	defer cache.Recycle()

	rst.M = make(map[string]uint32)
	tb := common.UserTableName + common.DbTableKeySplit + common.PlayerTrophyMapName_Old
	for _, uid := range req.Uids {
		number, err := redis.Int(cache.Do("HGET", tb, uid))
		if err != nil && err != redis.ErrNil {
			return err
		} else {
			rst.M[uid] = uint32(number)
		}
	}

	return nil
}

func (self *Center) setValue(playerID string, Value string) error {
	err := self.hsetValue(common.UserTableName, common.PlayerTrophyMapName_New, playerID, Value)
	if err != nil {
		logger.Error("SetValue error : %s (set %s(%d) to %s)", err.Error(), playerID, Value, common.UserTableName+common.PlayerTrophyMapName_New)
	}

	return nil
}

// 添加删除好友通知
func (self *Center) NotifyAddDelFriend(req *proto.FriendNoticeUpdate, rst *proto.FriendNoticeUpdateRst) error {
	if rpcc := self.getOnlineGas(req.Uid); rpcc != nil {
		rpcc.Go("CenterService.NotifyAddDelFriend", req, rst, nil)
	} else {

	}

	return nil
}
