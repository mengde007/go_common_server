package center

import (
	"common"
	"logger"
	"proto"
	// "rpcplus"
	"sync"
)

type StCenterActivityRank struct {
	m common.MapActivityRank
	l sync.RWMutex
}

var goActivityRank *StCenterActivityRank

func (self *Center) initActivityRank() {
	self.activityRank = &StCenterActivityRank{
		m: make(common.MapActivityRank),
	}

	buf, err := common.Resis_getbuf(self.maincache,
		common.SystemTableName,
		common.SystemKeyName_AccRank)

	if buf == nil {
		return
	}

	if err != nil {
		logger.Fatal("Resis_getbuf failed:", err)
		return
	}

	self.activityRank.l.Lock()
	err = common.GobDecode(buf, &self.activityRank.m)
	self.activityRank.l.Unlock()

	if err != nil {
		logger.Fatal("Resis_getbuf failed:", err)
		return
	}
}

func (self *Center) saveActivityRank() {
	buf, err := common.GobEncode(self.activityRank.m)
	if err != nil {
		return
	}

	common.Resis_setbuf(self.maincache,
		common.SystemTableName,
		common.SystemKeyName_AccRank,
		buf)
}

func (self *Center) ActivityRankEnd(req *proto.ActivityRankEnd, rst *proto.ActivityRankEndRst) error {
	self.activityRank.l.Lock()

	ar, ok := self.activityRank.m[req.Id]
	if ok && ar.EndTime == req.EndTime {
		self.activityRank.l.Unlock()
		return nil
	}

	uids, err := self.zrevrange("rank", "player", req.RankBegin, req.RankEnd)
	if err != nil {
		self.activityRank.l.Unlock()
		logger.Error("ActivityRankEnd get rank failed", err, uids)
		return err
	}

	self.activityRank.m[req.Id] = &common.StActivityRank{
		Id:      req.Id,
		EndTime: req.EndTime,
		Uids:    uids,
	}
	self.activityRank.l.Unlock()

	//保存
	self.saveActivityRank()

	//分发到所有的gameserver
	// self.pushActivityRank(nil)

	return nil
}

// func (self *Center) pushActivityRank(conn *rpcplus.Client) error {
// 	self.activityRank.l.RLock()
// 	buf, err := common.GobEncode(self.activityRank.m)
// 	if err != nil {
// 		self.activityRank.l.RUnlock()
// 		logger.Error("pushActivityRank failed 1:", err)
// 		return err
// 	}
// 	self.activityRank.l.RUnlock()

// 	req := &proto.NotifyActivityRank{
// 		Buf: buf,
// 	}

// 	rst := &proto.NotifyActivityRankRst{}

// 	if conn != nil {
// 		if err := conn.Call("CenterService.ActivityRankInfo", req, rst); err != nil {
// 			logger.Error("pushActivityRank failed 2:", err)
// 		}
// 	} else {
// 		for _, rpcc := range self.cnss {
// 			if err := rpcc.Call("CenterService.ActivityRankInfo", req, rst); err != nil {
// 				logger.Error("pushActivityRank failed 3:", err)
// 			}
// 		}
// 	}

// 	return nil
// }
