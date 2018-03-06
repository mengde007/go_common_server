package center

import (
	"common"
	"dbclient"
	"logger"
	"proto"
	"rankclient"
	"rpc"
	"rpcplus"
	"time"
	"timer"
)

type RanksResult struct {
	RankingByts proto.SaveRankPlayers
}

var AllRank RanksResult

// 第一次推送 启服时gas链接center时 center第一次推送
func (self *Center) theFirstPushRank(conn *rpcplus.Client) {
	logger.Info("theFirstPushRank!!!!!")
	rst := &proto.SaveRankPlayersRst{}

	if AllRank.RankingByts.Exps != nil {
		conn.Call("CenterService.SaveRankings", AllRank.RankingByts, rst)
		if !rst.OK {
			logger.Error("CenterService.SaveRankings error!")
		}
	}
}

func (self *Center) initUpdateAllRank() {
	self.updateTick = timer.NewTimer(time.Duration(5) * time.Minute)
	self.updateTick.Start(
		func() {
			self.updateAllRanks(false)
		},
	)

	// 先调用一次更新
	self.updateAllRanks(true)
}

func (self *Center) updateAllRanks(bIsFirst bool) {
	go self.updateFightPowerTotallRanking(50, bIsFirst)
}

func (self *Center) updateFightPowerTotallRanking(showMax int, bIsFirst bool) {
	logger.Info("updateFightPowerTotallRanking called showMax:", showMax)
	defer logger.Info("updateFightPowerTotallRanking end")
	rets, err := rankclient.GetRankingInfo(showMax)
	if err != nil {
		logger.Error("rankclient.GetRankingInfo error:", err)
		return
	}

	//exp ranking
	msg := &rpc.RankList{}
	msg.SetRankType(int32(3))
	self.fillRankInfo(rets.Exps, msg, 3)
	bytes, err := common.EncodeMessage(msg)
	if err != nil {
		logger.Error("updateFightPowerTotallRanking EncodeMessage error:", err)
		return
	}
	AllRank.RankingByts.Exps = bytes

	//coin ranking
	msg = &rpc.RankList{}
	msg.SetRankType(int32(2))
	self.fillRankInfo(rets.Coins, msg, 2)
	bytes, err = common.EncodeMessage(msg)
	if err != nil {
		logger.Error("updateFightPowerTotallRanking EncodeMessage error:", err)
		return
	}
	AllRank.RankingByts.Coins = bytes

	//profits ranking
	msg = &rpc.RankList{}
	msg.SetRankType(int32(1))
	self.fillRankInfo(rets.Profits, msg, 1)
	bytes, err = common.EncodeMessage(msg)
	if err != nil {
		logger.Error("updateFightPowerTotallRanking EncodeMessage error:", err)
		return
	}
	AllRank.RankingByts.Profits = bytes

	if bIsFirst {
		return
	}

	req := &proto.SaveRankPlayers{
		Exps:    AllRank.RankingByts.Exps,
		Coins:   AllRank.RankingByts.Coins,
		Profits: AllRank.RankingByts.Profits,
	}
	rst := &proto.SaveRankPlayersRst{}
	for i, v := range centerServer.cnss {
		v.Call("CenterService.SaveRankings", req, rst)
		if !rst.OK {
			logger.Info("CenterService.SaveRankings error! serverId:", i)
		}
	}
}

func (self *Center) fillRankInfo(ranks []string, msg *rpc.RankList, eType int) {
	logger.Info("fillRankInfo called ranks:%d", len(ranks))
	if len(ranks) == 0 {
		return
	}

	var base rpc.PlayerBaseInfo
	for k, v := range ranks {
		if err := self.getPlayerBase(v, &base); err != nil {
			continue
		}

		info := &rpc.RankInfo{}
		info.SetUid(base.GetUid())
		info.SetRoleId(base.GetRoleId())
		info.SetSex(base.GetSex())
		info.SetName(base.GetName())
		info.SetLevel(base.GetLevel())
		info.SetExp(base.GetExp())
		info.SetBVip(false)
		if base.GetVipLeftDay() > int32(0) {
			info.SetBVip(true)
		}
		info.SetCoin(base.GetCoin())
		info.SetGem(base.GetGem())
		info.SetHeaderUrl(base.GetHeaderUrl())

		if eType == 3 {
			info.SetRankValue(int64(base.GetExpTotal()))
			logger.Info("fillRankInfo eType:%d, name:%s, rank:%d, exp:%d, k:%s", eType, base.GetName(), k+1, base.GetExpTotal(), v)
		} else if eType == 2 {
			info.SetRankValue(int64(base.GetCoin() + base.GetInsurCoin()))
			logger.Info("fillRankInfo eType:%d, name:%s, rank:%d, coin:%d, k:%s", eType, base.GetName(), k+1, base.GetCoin()+base.GetInsurCoin(), v)
		} else {
			profits := int64(base.GetProfits())
			if profits < int64(0) {
				profits = 0
			}
			info.SetRankValue(profits)
			logger.Info("fillRankInfo eType:%d, name:%s, rank:%d, profits:%d, k:%s", eType, base.GetName(), k+1, profits, v)
		}
		info.SetRankNum(int32(k + 1))
		msg.RankList = append(msg.RankList, info)

	}
}

func (self *Center) getPlayerBase(uid string, base *rpc.PlayerBaseInfo) error {
	exist, err := dbclient.KVQueryBase(common.TB_t_base_playerbase, uid, base)
	if err != nil || !exist {
		logger.Error("getPlayerBase query PlayerBaes failed!:", err, uid)
		return err
	}
	return nil
}

// TODO: 1. RPC调用对应的排行榜服务器得到排行榜数据
// TODO: 2. 根据1中的排行榜uid读取数据库填充排行榜玩家信息
// TODO: 3. 把填充好的排行榜信息发送到gameserver(由于是gp结构 序列化发送[]byte)centerServer.
// TODO: 4. 更新本地缓存
