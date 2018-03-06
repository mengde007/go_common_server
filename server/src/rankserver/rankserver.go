package rankserver

import (
	"common"
	"logger"
	"net"
	"proto"
	"rpcplus"
	"runtime/debug"
	"sync"
	"time"
	"timer"
)

type GeneralRankServer struct {
	fpRank     sync.RWMutex
	pCachePool *common.CachePool
	updateTick *timer.Timer
}

const (
	RANKING_MAIN_TABLE = "ranking_main_talbe"
	EXP                = "exp"
	COIN               = "coin"
	PROFIT             = "profit"
)

var Cfg common.GeneralRankServerCfg
var pServer *GeneralRankServer

func CreateServices(cfg common.GeneralRankServerCfg, listener net.Listener) *GeneralRankServer {
	pServer = &GeneralRankServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)
	common.LoadGlobalConfig()

	pServer.first_tick()
	// pServer.ClearRedis()
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Info("Start GeneralRankServices %s", err.Error())
			break
		}

		//开始对cns的RPC服务
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("GeneralRankServer Rpc Runtime Error: %s", r)
					debug.PrintStack()
				}
			}()
			rpcServer.ServeConn(conn)
		}()
	}
	return pServer
}

func (self *GeneralRankServer) first_tick() {
	tmNow := time.Now()
	tickTtime := time.Date(tmNow.Year(), tmNow.Month(), tmNow.Day()+1, 0, 0, 0, 0, time.Local)
	nextTick := tickTtime.Unix() - tmNow.Unix()

	self.updateTick = timer.NewTimer(time.Duration(nextTick) * time.Second)
	self.updateTick.Start(
		func() {
			self.day_tick()
		},
	)
}

func (self *GeneralRankServer) day_tick() {
	logger.Info("day_tick() excuted")

	self.updateTick.Stop()
	self.updateTick = nil

	//clear profits ranking
	self.fpRank.Lock()
	rets, err := common.Redis_zrevrange(self.pCachePool, RANKING_MAIN_TABLE, PROFIT, 0, 999999)
	if err != nil {
		self.fpRank.Unlock()
		logger.Error("Redis_zrevrange error:", err)
		return
	}
	for _, v := range rets {
		common.Redis_zadd(self.pCachePool, RANKING_MAIN_TABLE, PROFIT, v, uint32(0))
		logger.Info("day_tick :%s", v)
	}
	self.fpRank.Unlock()

	self.updateTick = timer.NewTimer(time.Duration(24*60*60) * time.Second)
	self.updateTick.Start(
		func() {
			self.day_tick()
		},
	)
}

func (self *GeneralRankServer) UpdateRankingInfo(req *proto.SetRankInfo, rst *proto.SetRankInfoRst) error {
	logger.Info("UpdateRankingInfo begain eType:%d, value:%d ", req.EType, req.Value)
	defer logger.Info("UpdateRankingInfo end")

	if req.EType == 3 { //exp
		common.Redis_zadd(self.pCachePool, RANKING_MAIN_TABLE, EXP, req.Uid, uint32(req.Value))
	} else if req.EType == 2 { //coin
		if req.Value < 0 {
			req.Value = int32(0)
		}
		common.Redis_zadd(self.pCachePool, RANKING_MAIN_TABLE, COIN, req.Uid, uint32(req.Value))
	} else { //profit
		self.fpRank.RLock()
		if req.Value < 0 {
			req.Value = int32(0)
		}
		common.Redis_zadd(self.pCachePool, RANKING_MAIN_TABLE, PROFIT, req.Uid, uint32(req.Value))
		self.fpRank.RUnlock()
	}
	return nil
}

func (self *GeneralRankServer) GetRankingInfo(req *proto.GetRankInfo, rst *proto.GetRankInfoRst) error {
	logger.Info("GetRankingInfo begain")
	defer logger.Info("GetRankingInfo end")

	//exp
	rets, err := common.Redis_zrevrange(self.pCachePool, RANKING_MAIN_TABLE, EXP, 0, req.Number)
	if err != nil {
		logger.Error("Redis_zrevrange error:", err)
		return err
	}
	rst.Exps = rets

	//coin
	rets, err = common.Redis_zrevrange(self.pCachePool, RANKING_MAIN_TABLE, COIN, 0, req.Number)
	if err != nil {
		logger.Error("Redis_zrevrange error:", err)
		return err
	}
	rst.Coins = rets

	//profit
	self.fpRank.RLock()
	defer self.fpRank.RUnlock()
	rets, err = common.Redis_zrevrange(self.pCachePool, RANKING_MAIN_TABLE, PROFIT, 0, req.Number)
	if err != nil {
		logger.Error("Redis_zrevrange error:", err)
		return err
	}
	rst.Profits = rets

	return nil
}

func (self *GeneralRankServer) GetMyRankingInfo(req *proto.GetMyRankInfo, rst *proto.GetMyRankInfoRst) error {
	if req.EType == 3 {
		rank, err := common.Redis_zrevrank(self.pCachePool, RANKING_MAIN_TABLE, EXP, req.Uid)
		if err != nil {
			logger.Error("Redis_zrevrank error:", err)
			return err
		}
		rst.Ranking = int32(rank)
	} else if req.EType == 2 {
		rank, err := common.Redis_zrevrank(self.pCachePool, RANKING_MAIN_TABLE, COIN, req.Uid)
		if err != nil {
			logger.Error("Redis_zrevrank error:", err)
			return err
		}
		rst.Ranking = int32(rank)
	} else {
		rank, err := common.Redis_zrevrank(self.pCachePool, RANKING_MAIN_TABLE, PROFIT, req.Uid)
		if err != nil {
			logger.Error("Redis_zrevrank error:", err)
			return err
		}
		rst.Ranking = int32(rank)
	}
	return nil
}
