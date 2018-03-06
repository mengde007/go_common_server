package connector

import (
	"common"
	"logger"
	"rankclient"
	"rpc"
)

const (
	RANK_PROFIT = 1
	RANK_COIN   = 2
	RANK_EXP    = 3
)

const (
	RANK_CNT = 50
)

func (cn *CNServer) GetRanking(conn rpc.RpcConn, msg rpc.ReqRankList) error {
	logger.Info("GetRanking called")
	p, exist := cn.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	eType := msg.GetRankType()
	sendMsg := &rpc.RankList{}
	sendMsg.SetRankType(eType)
	if eType == RANK_EXP {
		cns.rankMgr.explock.RLock()
		sendMsg.RankList = cns.rankMgr.Exps.RankList
		cns.rankMgr.explock.RUnlock()
	} else if eType == RANK_COIN {
		cns.rankMgr.explock.RLock()
		sendMsg.RankList = cns.rankMgr.Coins.RankList
		cns.rankMgr.explock.RUnlock()
		common.WriteResult(conn, sendMsg)
	} else {
		cns.rankMgr.explock.RLock()
		sendMsg.RankList = cns.rankMgr.Profits.RankList
		cns.rankMgr.explock.RUnlock()
	}
	cns.getMyRanking(sendMsg, p, eType)
	common.WriteResult(conn, sendMsg)
	return nil
}

func (cn *CNServer) getMyRanking(ranking *rpc.RankList, p *player, rankType int32) {
	logger.Info("getMyRanking rankType:", rankType)
	if len(ranking.RankList) < RANK_CNT {
		return
	}

	//在排行榜中
	uid := p.GetUid()
	for _, rank := range ranking.RankList {
		if rank.GetUid() == uid {
			return
		}
	}

	info := &rpc.RankInfo{}
	info.SetUid(p.GetUid())
	info.SetRoleId(p.GetRoleId())
	info.SetSex(p.GetSex())
	info.SetName(p.GetName())
	info.SetLevel(p.GetLevel())
	info.SetExp(p.GetExp())
	if p.GetVipLeftDay() > 0 {
		info.SetBVip(true)
	} else {
		info.SetBVip(false)
	}
	info.SetCoin(p.GetCoin())
	info.SetGem(p.GetGem())
	info.SetHeaderUrl(p.GetHeaderUrl())
	rank, err := rankclient.GetMyRankingInfo(int(rankType), uid)
	if err != nil {
		logger.Error("rankclient.GetRankingInfo error, eType:%d, uid:%s", rankType, uid, err)
		return
	}
	info.SetRankNum(int32(rank + 1))

	if rankType == RANK_EXP {
		info.SetRankValue(int64(p.GetExp()))
	} else if rankType == RANK_COIN {
		info.SetRankValue(int64(p.GetCoin()))
	} else if rankType == RANK_PROFIT {
		info.SetRankValue(int64(p.GetProfits()))
	}
}
