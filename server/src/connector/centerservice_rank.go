package connector

import (
	"common"
	"logger"
	"proto"
	"rpc"
	"sync"
	"time"
)

// 每一种排行榜都在RankMgr里面有对应的值和锁
type RankMgr struct {
	Exps           rpc.RankList
	Coins          rpc.RankList
	Profits        rpc.RankList
	lastUpdateTime int64

	explock    sync.RWMutex
	coinlock   sync.RWMutex
	profitlock sync.RWMutex
}

func NewRankMgr() *RankMgr {
	return &RankMgr{}
}

func (c *CenterService) SaveRankings(req *proto.SaveRankPlayers, rst *proto.SaveRankPlayersRst) error {
	logger.Info("SaveRankings begin")
	defer logger.Info("SaveRankings end")

	//exp ranking
	cns.rankMgr.explock.Lock()
	if err := common.DecodeMessage(req.Exps, &cns.rankMgr.Exps); err != nil {
		logger.Error("SaveRankings DecodeMessage proto.Exps error:", err)
		rst.OK = false
		return err
	}
	rst.OK = true
	cns.rankMgr.explock.Unlock()

	//coin ranking
	cns.rankMgr.coinlock.Lock()
	if err := common.DecodeMessage(req.Coins, &cns.rankMgr.Coins); err != nil {
		logger.Error("SaveRankings DecodeMessage proto.Coins error:", err)
		rst.OK = false
		return err
	}
	rst.OK = true
	cns.rankMgr.coinlock.Unlock()

	//profit ranking
	cns.rankMgr.profitlock.Lock()
	if err := common.DecodeMessage(req.Profits, &cns.rankMgr.Profits); err != nil {
		logger.Error("SaveRankings DecodeMessage proto.Profits error:", err)
		rst.OK = false
		return err
	}
	rst.OK = true
	cns.rankMgr.profitlock.Unlock()

	cns.rankMgr.lastUpdateTime = time.Now().Unix()
	return nil
}
