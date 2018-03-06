package connector

import (
	"common"
	"logger"
	"sync"
	"sync/atomic"
)

const (
	PLAYERS_SEG = 1024
)

type stPlayerConnIdSeg struct {
	l       sync.RWMutex
	players map[uint64]*player
}

type stPlayerUidSeg struct {
	l       sync.RWMutex
	players map[string]*player
}
type stPlayerOpenIdSeg stPlayerUidSeg //与uid结构一样

func getConnIdSeg(connId uint64) int {
	return int(connId % uint64(PLAYERS_SEG))
}

func getUidSeg(uid string) int {
	return int(common.MakeHash(uid) % uint32(PLAYERS_SEG))
}

//初始化
func (self *CNServer) initSegs() {
	for i := 0; i < PLAYERS_SEG; i++ {
		self.players[i] = &stPlayerConnIdSeg{
			players: make(map[uint64]*player),
		}
		self.otherplayers[i] = &stPlayerConnIdSeg{
			players: make(map[uint64]*player),
		}
		self.playersbyid[i] = &stPlayerUidSeg{
			players: make(map[string]*player),
		}
	}
}

//添加玩家到全局表中
func (self *CNServer) addPlayer(connId uint64, p *player) {
	// pts(p, "CNServer:addPlayer")
	//连接id
	segConn := self.players[getConnIdSeg(connId)]
	segConn.l.Lock()
	segConn.players[connId] = p
	segConn.l.Unlock()

	//uid
	self.addPlayerByUid(p)

	if 0 == p.mobileqqinfo.PlatId {
		atomic.AddInt32(&self.iosplayer, 1)
	} else if 1 == p.mobileqqinfo.PlatId {
		atomic.AddInt32(&self.androidplayer, 1)
	}

	// pte(p, "CNServer:addPlayer")
}

//从全局表中取玩家
func (self *CNServer) getPlayerByConnId(connId uint64) (*player, bool) {
	segConn := self.players[getConnIdSeg(connId)]
	segConn.l.RLock()
	p, ok := segConn.players[connId]
	segConn.l.RUnlock()

	if ok && p != nil {
		return p, ok
	}

	return nil, false
}

//销毁玩家
func (self *CNServer) delPlayer(connId uint64) {
	ts("CNServer:delPlayer connId = ", connId)
	segConn := self.players[getConnIdSeg(connId)]
	segConn.l.Lock()
	p, exist := segConn.players[connId]
	delete(segConn.players, connId)
	segConn.l.Unlock()

	if exist {
		if 0 == p.mobileqqinfo.PlatId {
			atomic.AddInt32(&self.iosplayer, -1)
		} else if 1 == p.mobileqqinfo.PlatId {
			atomic.AddInt32(&self.androidplayer, -1)
		}
		self.delPlayerByUid(p.GetUid())
		p.OnQuit(true)
	}
	te("CNServer:delPlayer connId = ", connId)
}

/************************************
           	otherplayer
************************************/
//添加被攻击玩家到全局表中
func (self *CNServer) addOtherPlayer(connId uint64, p *player) {
	ts("CNServer:addOtherPlayer connId = ", connId, p.GetUid())

	//上一个玩家还在的情况
	self.delOtherPlayer(connId)

	//连接id
	segConn := self.otherplayers[getConnIdSeg(connId)]
	segConn.l.Lock()
	segConn.players[connId] = p
	segConn.l.Unlock()

	//uid
	self.addPlayerByUid(p)

	te("CNServer:addOtherPlayer connId = ", connId)
}

//从全局表中取其他玩家
func (self *CNServer) getOtherPlayerByConnId(connId uint64) (*player, bool) {
	segConn := self.otherplayers[getConnIdSeg(connId)]
	segConn.l.RLock()
	p, ok := segConn.players[connId]
	segConn.l.RUnlock()

	if ok && p != nil {
		return p, ok
	}

	return nil, false
}

//销毁被攻击的玩家
func (self *CNServer) delOtherPlayer(connId uint64) {
	ts("CNServer:delOtherPlayer connId = ", connId)

	segConn := self.otherplayers[getConnIdSeg(connId)]
	segConn.l.Lock()
	p, exist := segConn.players[connId]
	if exist {
		delete(segConn.players, connId)
	}
	segConn.l.Unlock()

	if exist {
		self.delPlayerByUid(p.GetUid())
		p.OnQuit(false)
	}

	te("CNServer:delOtherPlayer connId = %v", connId)
}

/************************************
        uid,包括otherplayer
************************************/
func (self *CNServer) addPlayerByUid(p *player) {
	logger.Info("**********addPlayerByUid:%d", getUidSeg(p.GetUid()))

	segUid := self.playersbyid[getUidSeg(p.GetUid())]
	segUid.l.Lock()
	segUid.players[p.GetUid()] = p
	segUid.l.Unlock()
}

func (self *CNServer) getPlayerByUid(uid string) (*player, bool) {
	segUid := self.playersbyid[getUidSeg(uid)]
	segUid.l.RLock()
	p, ok := segUid.players[uid]
	segUid.l.RUnlock()

	if ok && p != nil {
		return p, ok
	}

	return nil, false
}

func (self *CNServer) delPlayerByUid(uid string) {
	segUid := self.playersbyid[getUidSeg(uid)]
	segUid.l.Lock()
	delete(segUid.players, uid)
	segUid.l.Unlock()
}

/************************************
        特殊流程，只从map中删除
************************************/
func (self *CNServer) delMapPlayer(connId uint64) {
	segConn := self.players[getConnIdSeg(connId)]
	segConn.l.Lock()
	delete(segConn.players, connId)
	segConn.l.Unlock()
}

func (self *CNServer) delMapOtherPlayer(connId uint64) {
	segConn := self.otherplayers[getConnIdSeg(connId)]
	segConn.l.Lock()
	delete(segConn.players, connId)
	segConn.l.Unlock()
}

/************************************
             取在线人数
************************************/
func (self *CNServer) getOnlineNumbers() uint32 {
	numbers := uint32(0)

	for _, segUid := range self.playersbyid {
		numbers += uint32(len(segUid.players))
	}

	return numbers
}
