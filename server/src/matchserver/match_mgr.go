package matchserver

import (
	//conn "centerclient"
	cmn "common"
	//	"fmt"
	"logger"
	"rpc"
	"runtime/debug"
	//	"strconv"
	"time"
	//  ds "daerserver"
	//"errors"
	//"strconv"
	//"strings"
	"timer"
)

var matchMgr *MatchMgr

type MatchMgr struct {
	t *timer.Timer

	matches []*Match //比赛列表
	//playerInRoom map[string]cmn.GameRoom //保存玩家所在的游戏房间
}

func (self *MatchMgr) init() {
	self.initMatchList()
	self.createTimer()
}

func (self *MatchMgr) initMatchList() {
	matchesCfg := cmn.GetMatchConfigForAll()
	if matchesCfg == nil {
		logger.Error("matchesCfg is nil.")
	}

	for _, cfg := range matchesCfg {
		if cfg == nil {
			continue
		}

		self.matches = append(self.matches, NewMatch(*cfg))
	}
}

func (self *MatchMgr) createTimer() {
	self.t = timer.NewTimer(time.Second)
	self.t.Start(
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("player tick runtime error :", r)
					debug.PrintStack()
				}
			}()

			self.OnTick()
		},
	)
}

func (self *MatchMgr) OnTick() {
	//fmt.Println("OnTick time:%d", time.Now().Unix())

	for i, match := range self.matches {

		if !match.isRemoved {
			match.Update()
		} else {
			self.matches = append(self.matches[:i], self.matches[i+1:]...)
			break
		}
	}

}

func (self *MatchMgr) MatchListREQ() {

	matchListACK := &rpc.MatchListACK{}
	for _, match := range self.matches {
		if match == nil || !match.isValid {
			continue
		}

		m := &rpc.Match{}
		m.SetId(match.id)
		matchListACK.Matches = append(matchListACK.Matches, m)

	}

}

func (self *MatchMgr) EnrollREQ(playerBasic *rpc.PlayerBaseInfo, msg *rpc.EnrollREQ) {

}

func (self *MatchMgr) WithdrawREQ(uid string, msg *rpc.WithdrawREQ) {

}
