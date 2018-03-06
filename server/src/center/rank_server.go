package center

// import (
// 	gp "code.google.com/p/goprotobuf/proto"
// 	"logger"
// 	"proto"
// 	"rpc"
// 	"strconv"
// 	//add for update rankplayers
// 	// "clanclient"
// 	"common"
// 	"dbclient"
// 	"rpcplus"
// 	//"superleagueclient"
// 	"herobattleclient"
// 	"superleagueconfig"
// 	"time"
// 	"timer"
// 	"tttclient"
// )

// type SaveRankResult struct {
// 	RankPlayers         proto.SaveRankPlayer
// 	RankLocationPlayers proto.SaveRankLocationPlayers
// 	RankClans           proto.SaveRankClan
// 	RankTTTplayers      proto.SaveTTTRank
// 	RankMyself          proto.SaveMyself
// 	RankHeroBattle      proto.SaveHeroBattleRank
// }

// var SaveAllRankPlayers SaveRankResult

// //第一次推送，看看那个cns连接到我了,就推送一次
// func (self *Center) theFirstUpdate(conn *rpcplus.Client) {
// 	RankPlayerReply := &proto.SaveRankPlayerResult{}
// 	RankLocationPlayersReply := &proto.SaveRankLocationPlayersResult{}
// 	RankClansReply := &proto.SaveRankClanResult{}
// 	RankTTTplayersReply := &proto.SaveTTTRankResult{}
// 	RankHeroBattleReply := &proto.SaveHeroBattleRankResult{}

// 	conn.Call("CenterService.SaveRankPlayers", SaveAllRankPlayers.RankPlayers, RankPlayerReply)
// 	if !RankPlayerReply.OK {
// 		logger.Error("CenterService.SaveRankPlayers error!")
// 	}

// 	conn.Call("CenterService.SaveRankLocationPlayers", SaveAllRankPlayers.RankLocationPlayers, RankLocationPlayersReply)
// 	if !RankLocationPlayersReply.OK {
// 		logger.Error("CenterService.SaveRankLocationPlayers error!")
// 	}

// 	conn.Call("CenterService.SaveRankClans", SaveAllRankPlayers.RankClans, RankClansReply)
// 	if !RankClansReply.OK {
// 		logger.Error("CenterService.SaveRankClans error!")
// 	}

// 	conn.Call("CenterService.SaveTTTPlayers", SaveAllRankPlayers.RankTTTplayers, RankTTTplayersReply)
// 	if !RankTTTplayersReply.OK {
// 		logger.Error("CenterService.SaveTTTPlayers error!")
// 	}

// 	conn.Call("CenterService.SaveRankHeroBattle", SaveAllRankPlayers.RankHeroBattle, RankHeroBattleReply)
// 	if !RankHeroBattleReply.OK {
// 		logger.Error("CenterService.SaveRankHeroBattle error!!!")
// 	}
// }

// func (self *Center) initUpdateRankPlayers() {
// 	//开始定时器
// 	var centercfg common.CenterConfig
// 	if err := common.ReadCenterConfig(&centercfg); err != nil {
// 		logger.Error("read common.CenterConfig error", err)
// 		return
// 	}

// 	updatetime := centercfg.UpdateTime

// 	self.updatetime = timer.NewTimer(time.Duration(updatetime) * time.Minute)
// 	self.updatetime.Start(
// 		func() {
// 			self.updateAllRankplayers(false)
// 		},
// 	)

// 	//立即调用一次更新缓存
// 	self.updateAllRankplayers(true)
// }

// func (self *Center) updateAllRankplayers(bIsFirst bool) {
// 	self.updateRankplayers(bIsFirst)
// 	//暂时注释掉区域排行榜
// 	//self.updateRankplayersLocation(bIsFirst)
// 	self.updateRankClans(bIsFirst)
// 	self.updateTTTplayers(bIsFirst)
// 	// 无双排行榜
// 	self.updateRankHeroBattle(bIsFirst)
// }

// func (self *Center) updateRankplayers(bIsFirst bool) {
// 	//加入最大的检测
// 	start := common.GetGlobalConfig("NUMBER_OF_RANK_START")
// 	stop := common.GetGlobalConfig("NUMBER_OF_RANK_END")
// 	//数据强制修正，如果超过200，则强制修正为200
// 	if stop > uint32(200) {
// 		stop = uint32(200)
// 	}
// 	var req proto.GetRankPlayers
// 	var reply proto.GetRankPlayersResult
// 	req.Start = int(start)
// 	req.Stop = int(stop)
// 	self.getRankPlayers(&req, &reply, bIsFirst)
// }

// func (self *Center) updateRankplayersLocation(bIsFirst bool) {
// 	start := common.GetGlobalConfig("NUMBER_OF_RANK_START")
// 	stop := common.GetGlobalConfig("NUMBER_OF_RANK_END")
// 	var req proto.GetRankPlayersLocation
// 	var reply proto.GetRankPlayersLocationResult
// 	req.Start = int(start)
// 	req.Stop = int(stop)

// 	for _, value := range rpc.GameLocation_value {

// 		req.Location = int64(value)
// 		self.getRankPlayersLocation(&req, &reply, bIsFirst)
// 	}
// }

// func (self *Center) updateRankClans(bIsFirst bool) {
// 	err, buf := clanclient.QueryRankClans(0, 99)
// 	if err != nil {
// 		logger.Error("QueryRankClans failed", err)
// 		return
// 	}
// 	logger.Info("QueryRankClans ok", buf)

// 	//排名上升或者下降
// 	rankOld := &rpc.ClanInfos{}
// 	if len(SaveAllRankPlayers.RankClans.Value) > 0 {
// 		if err := gp.Unmarshal(SaveAllRankPlayers.RankClans.Value, rankOld); err != nil {
// 			rankOld = nil
// 		}
// 	}

// 	f := func(clan string, index int) int32 {
// 		if rankOld == nil {
// 			return 0
// 		}

// 		for i, c := range rankOld.Infos {
// 			if c.GetName() == clan {
// 				return int32(i - index)
// 			}
// 		}

// 		return 0
// 	}

// 	cis := &rpc.ClanInfos{}
// 	if err := gp.Unmarshal(buf, cis); err == nil {
// 		for index, ci := range cis.Infos {
// 			ci.SetRankupdown(f(ci.GetName(), index))
// 		}

// 		if bufTemp, err := gp.Marshal(cis); err == nil {
// 			buf = bufTemp
// 		}
// 	}

// 	req := &proto.SaveRankClan{}
// 	req.Value = buf
// 	rst := &proto.SaveRankClanResult{}

// 	if !bIsFirst {
// 		for _, conn := range centerServer.cnss {
// 			conn.Call("CenterService.SaveRankClans", req, rst)

// 			if !rst.OK {
// 				logger.Error("CenterService.SaveRankPlayers error")
// 			}
// 		}
// 	}

// 	//这里保存一下第一次查到的结果
// 	SaveAllRankPlayers.RankClans = *req

// 	logger.Info("come out getRankClans")

// 	return
// }

// func (self *Center) updateTTTplayers(bIsFirst bool) {
// 	req := proto.GetRankPlayerTTTScore{
// 		Start: 0,
// 		Stop:  99,
// 	}
// 	reply := proto.GetRankPlayerTTTScoreResult{}

// 	////需要通天塔数据
// 	if err := tttclient.GetRankPlayerTTTScore(&req, &reply); err != nil {
// 		logger.Error("updateTTTplayers failed", err)
// 		return
// 	}

// 	rps := &rpc.TTTRankPlayers{}
// 	for _, PlayerTTTScoreStructinfo := range reply.Value {
// 		var TTTp rpc.PlayerBaseInfo
// 		exist, err := dbclient.KVQueryBase(common.TB_t_base_playerbase, PlayerTTTScoreStructinfo.Id, &TTTp)
// 		if err != nil {
// 			continue
// 		}

// 		if exist {
// 			rp := rpc.TTTPlayer{}
// 			rp.SetType(rpc.TTTPlayer_Rank)
// 			rp.SetName(TTTp.GetName())
// 			rp.SetUid(TTTp.GetUid())
// 			rp.SetLevel(TTTp.GetLevel())
// 			rp.SetTrophy(TTTp.GetTrophy())
// 			rp.SetClanName(TTTp.GetClan())
// 			rp.SetClanSymbol(TTTp.GetClanSymbol())
// 			rp.SetTttSCoreQuery(PlayerTTTScoreStructinfo.Score)
// 			rp.SetRanknumberQuery(0)
// 			rp.SetVipLevel(TTTp.GetGameVipLevel())
// 			logger.Info("player vip level is %v, %v", rp.GetVipLevel(), TTTp.GetGameVipLevel())

// 			rps.RpsTop = append(rps.RpsTop, &rp)
// 		}

// 	}

// 	buff, err := gp.Marshal(rps)
// 	if err != nil {
// 		logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), buff)
// 		return
// 	}

// 	tttResult := &proto.SaveTTTRank{}
// 	tttResult.Value = buff
// 	myreply := &proto.SaveTTTRankResult{}
// 	if !bIsFirst {
// 		for _, conn := range centerServer.cnss {
// 			conn.Call("CenterService.SaveTTTPlayers", tttResult, myreply)

// 			if !myreply.OK {
// 				logger.Error("CenterService.SaveTTTPlayers error")
// 			}
// 		}
// 	}

// 	//这里保存一下第一次查到的结果
// 	SaveAllRankPlayers.RankTTTplayers = *tttResult
// }

// func (self *Center) updateRankHeroBattle(bFirst bool) {
// 	stop := common.GetGlobalConfig("WS_RANKDISPLAY")
// 	var req proto.GetRankHeroBattle
// 	var reply proto.GetRankHeroBattleResult
// 	req.Start = 0
// 	req.Stop = int(stop - 1)

// 	if err := herobattleclient.GetHeroBattleRankPlayers(&req, &reply); err != nil {
// 		logger.Error("updateRankHeroBattle error.", err)
// 		return
// 	}

// 	rps := &rpc.WuShuangRanks{}
// 	for index, uid := range reply.Value {
// 		logger.Info("updateRankHeroBattle: ", index, uid)
// 		rp := &rpc.WuShuangPlayer{}
// 		if common.IsWuShuangRobot(uid) {
// 			// robot
// 			rp.SetRank(uint32(index + 1))
// 			rp.SetUid(uid)
// 			rp.SetName(common.GetWuShuangRobotName(uid))
// 			attackHeros := common.GetWuShuangRobotAttackHero(uid)
// 			rp.Heros = attackHeros
// 		} else {
// 			// real player
// 			var base rpc.PlayerBaseInfo
// 			OK, err := dbclient.KVQueryBase(common.TB_t_base_playerbase, uid, &base)
// 			if err != nil || !OK {
// 				logger.Error("updateRankHeroBattle: querybase err", uid)
// 				continue
// 			}
// 			var v rpc.VillageInfo
// 			exist, err := dbclient.KVQueryExt(common.TB_t_ext_village, strconv.FormatUint(base.GetVillageId(), 16), &v)
// 			if err != nil || !exist {
// 				logger.Error("updateRankHeroBattle: queryvillage err", uid)
// 				continue
// 			}

// 			rp.SetRank(uint32(index + 1))
// 			rp.SetUid(uid)
// 			rp.SetName(base.GetName())
// 			heroContainer := v.Center.GetHeroContainer()
// 			attackHeros := common.GetWuShuangAttackHero(heroContainer)
// 			rp.Heros = attackHeros
// 		}

// 		rps.RpsTop = append(rps.RpsTop, rp)
// 	}

// 	buff, err := gp.Marshal(rps)
// 	if err != nil {
// 		logger.Error("get herobattle rank error on marshal: ", err.Error(), buff)
// 		return
// 	}
// 	wuShuangRanks := &proto.SaveHeroBattleRank{}
// 	myReply := &proto.SaveHeroBattleRankResult{}
// 	wuShuangRanks.Value = buff

// 	if !bFirst {
// 		for _, conn := range centerServer.cnss {
// 			conn.Call("CenterService.SaveRankHeroBattle", wuShuangRanks, myReply)
// 			if !myReply.OK {
// 				logger.Error("CenterService.SaveRankHeroBattle error!")
// 			}
// 		}
// 	}

// 	SaveAllRankPlayers.RankHeroBattle = *wuShuangRanks
// 	return
// }

// func (self *Center) getRankPlayers(req *proto.GetRankPlayers, reply *proto.GetRankPlayersResult, bIsFirst bool) (err error) {
// 	buf, err := self.zrevrange("rank", "player", req.Start, req.Stop)
// 	if err != nil {
// 		logger.Error("GetRankPlayers Error On zrevrange (%s, %v)", err.Error(), buf)
// 		return
// 	}

// 	//排名上升或者下降
// 	rankOld := &rpc.RankPlayers{}
// 	if len(SaveAllRankPlayers.RankPlayers.Value) > 0 {
// 		if err := gp.Unmarshal(SaveAllRankPlayers.RankPlayers.Value, rankOld); err != nil {
// 			rankOld = nil
// 		}
// 	}

// 	f := func(uid string, index int) int32 {
// 		if rankOld == nil {
// 			return 0
// 		}

// 		for i, p := range rankOld.RpsTop {
// 			if p.GetUid() == uid {
// 				return int32(i - index)
// 			}
// 		}

// 		return 0
// 	}

// 	reply.Code = proto.GetRankPlayerOk
// 	reply.Value = buf

// 	rps := &rpc.RankPlayers{}
// 	myreply := &proto.SaveRankPlayerResult{}

// 	//timeNow := uint32(time.Now().Unix())
// 	for index, uid := range reply.Value {
// 		var p rpc.PlayerBaseInfo

// 		exist, err := dbclient.KVQueryBase(common.TB_t_base_playerbase, uid, &p)
// 		if err != nil {
// 			continue
// 		}

// 		if exist {
// 			rp := rpc.Player{}
// 			rp.SetType(rpc.Player_Rank)
// 			rp.SetName(p.GetName())
// 			rp.SetRank(uint32(index))
// 			rp.SetUid(p.GetUid())
// 			rp.SetTrophy(p.GetTrophy())
// 			rp.SetLevel(p.GetLevel())
// 			rp.SetClanName(p.GetClan())
// 			rp.SetClanSymbol(p.GetClanSymbol())
// 			rp.SetRankupdown(f(p.GetUid(), len(rps.RpsTop)))
// 			rp.SetSuperleagueseg(superleagueconfig.GetRealSeg(p.GetSuperleagueseg(), p.GetSuperleaguesoverdue()))
// 			rp.SetGameVipLevel(p.GetGameVipLevel())

// 			// //check last rank in personal league
// 			// if p.GetLastPersonalOverdue() > timeNow && p.GetLastPersonalRank() > 0 {
// 			// 	rp.SetLastRankIcon(p.GetLastPersonalRank())
// 			// 	rp.SetRankIconOverdue(p.GetLastPersonalOverdue())
// 			// } else {
// 			// 	if rank, overdue, wrong := superleagueclient.QueryPersonalTop100(uid); wrong == nil {
// 			// 		if rank > 0 {
// 			// 			rp.SetLastRankIcon(rank)
// 			// 			rp.SetRankIconOverdue(overdue)
// 			// 		}
// 			// 	}
// 			// }

// 			rps.RpsTop = append(rps.RpsTop, &rp)
// 		}
// 	}

// 	buff, err := gp.Marshal(rps)
// 	if err != nil {
// 		logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), buff)
// 		return
// 	}

// 	globalPlayer := &proto.SaveRankPlayer{}
// 	globalPlayer.Value = buff

// 	if !bIsFirst {
// 		for _, conn := range centerServer.cnss {
// 			conn.Call("CenterService.SaveRankPlayers", globalPlayer, myreply)

// 			if !myreply.OK {
// 				logger.Error("CenterService.SaveRankPlayers error!")
// 			}
// 		}
// 	}

// 	//这里保存一下第一次查到的结果
// 	SaveAllRankPlayers.RankPlayers = *globalPlayer
// 	return nil
// }

// //add for location player 查询区域玩家杯数
// func (self *Center) getRankPlayersLocation(req *proto.GetRankPlayersLocation, reply *proto.GetRankPlayersLocationResult, bIsFirst bool) (err error) {
// 	buf, err := self.zrevrange("rank", "PlayerLocation"+strconv.Itoa(int(req.Location)), req.Start, req.Stop)
// 	if err != nil {
// 		logger.Error("GetRankPlayersLocation Error On zrevrange (%s, %v)", err.Error(), buf)
// 		return
// 	}

// 	reply.Code = proto.GetRankPlayersLocationResultOK
// 	reply.Value = buf
// 	var location int64

// 	rps := &rpc.RankPlayers{}
// 	myreply := &proto.SaveRankLocationPlayersResult{}

// 	for index, uid := range reply.Value {
// 		var p rpc.PlayerBaseInfo

// 		exist, err := dbclient.KVQueryBase(common.TB_t_base_playerbase, uid, &p)
// 		if err != nil {
// 			continue
// 		}

// 		if exist {
// 			rp := rpc.Player{}
// 			rp.SetType(rpc.Player_Rank)
// 			rp.SetName(p.GetName())
// 			rp.SetRank(uint32(index))
// 			rp.SetUid(p.GetUid())
// 			rp.SetTrophy(p.GetTrophy())
// 			rp.SetLevel(p.GetLevel())
// 			rp.SetClanName(p.GetClan())
// 			rp.SetClanSymbol(p.GetClanSymbol())
// 			rp.SetGameVipLevel(p.GetGameVipLevel())

// 			rps.RpsTop = append(rps.RpsTop, &rp)
// 		}

// 		location = int64(p.GetGamelocation())
// 	}

// 	buff, err := gp.Marshal(rps)
// 	if err != nil {
// 		logger.Error("SearchClan Error On Marshal (%s, %v)", err.Error(), buff)
// 		return
// 	}

// 	locationPlayer := &proto.SaveRankLocationPlayers{}
// 	locationPlayer.Value = buff
// 	locationPlayer.Location = req.Location

// 	if !bIsFirst {
// 		for _, conn := range centerServer.cnss {
// 			conn.Call("CenterService.SaveRankLocationPlayers", locationPlayer, myreply)

// 			if !myreply.OK {
// 				logger.Error("CenterService.SaveRankPlayers error!")
// 			}
// 		}
// 	}

// 	//这里保存一下第一次查到的结果
// 	if locationPlayer.Location == location {

// 		SaveAllRankPlayers.RankLocationPlayers = *locationPlayer
// 	}

// 	return nil
// }

// func (self *Center) getRankMyselfLocation(req *proto.GetMyself, reply *proto.GetMyselfResult) error {
// 	location, err := self.zrevrank("rank", "PlayerLocation"+req.Location, req.Uid)

// 	if err != nil {
// 		logger.Error("GetRankMyself Error On zrank ", err)
// 		return nil
// 	}

// 	reply.Code = proto.GetRankPlayerOk
// 	reply.Rank = int(location)

// 	return nil
// }

// func (self *Center) GetRankMyselfGlobal(req *proto.GetMyself, reply *proto.GetMyselfResult) error {
// 	global, err := self.zrevrank("rank", "player", req.Uid)

// 	if err != nil {
// 		logger.Error("GetRankMyself Error On zrank ", err)
// 		return nil
// 	}

// 	reply.Code = proto.GetRankPlayerOk
// 	reply.Rank = int(global)

// 	return nil
// }

// func (self *Center) GetRankPlayerNum(req *proto.GetMyself, reply *proto.GetMyselfResult) error {
// 	global, err := self.zcard("rank", "player")

// 	if err != nil {
// 		logger.Error("GetRankMyself Error On zrank ", err)
// 		return nil
// 	}

// 	reply.Code = proto.GetRankPlayerOk
// 	reply.Rank = int(global)

// 	return nil
// }

// //*RPC*////////////////////////////////////////////////////////////////////////////
// func (self *Center) SetPlayerTrophy(req *proto.SetTrophy, reply *proto.SetTrophyResult) (err error) {
// 	ts("Center:SetPlayerTrophy", req.Uid, req.Trophy)

// 	//更新玩家杯数
// 	if err = self.zadd("rank", "player", req.Uid, req.Trophy); err != nil {
// 		te("Center:SetPlayerTrophy", req.Uid, req.Trophy)
// 		return
// 	}

// 	// 增加玩家区域字段，方便查找区域玩家排名
// 	if err = self.zadd("rank", "PlayerLocation"+strconv.Itoa(int(req.Location)), req.Uid, req.Trophy); err != nil {
// 		te("Center:SetPlayerTrophy", req.Uid, req.Trophy)
// 		return
// 	}

// 	//这里更新玩家杯数的时候，顺便更新一下玩家当前最高杯数
// 	number, _ := self.getCurPlayerValue(req.Uid)
// 	// 如果玩家身上的最高大于读取到的，就保存一下
// 	if uint32(number) < req.Trophy {
// 		err := self.setValue(req.Uid, strconv.Itoa(int(req.Trophy)))
// 		if err != nil {
// 			logger.Error("Set Player Cur Highest Trophy error : %s", err.Error())
// 			te("Center:SetPlayerTrophy", req.Uid, req.Trophy)
// 			return err
// 		}
// 	}

// 	te("Center:SetPlayerTrophy", req.Uid, req.Trophy)
// 	return
// }

// func (self *Center) DelPlayerTrophy(req *proto.SetTrophy, reply *proto.SetTrophyResult) (err error) {
// 	ts("Center:DelPlayerTrophy", req.Uid, req.Trophy)

// 	//更新玩家杯数
// 	if err = self.zrem("rank", "player", req.Uid); err != nil {
// 		te("Center:DelPlayerTrophy", req.Uid, req.Trophy)
// 		return
// 	}

// 	te("Center:DelPlayerTrophy", req.Uid, req.Trophy)
// 	return
// }

// func (self *Center) GetRankPlayers(req *proto.SLQueryRankPlayers, rst *proto.SLQueryRankPlayersRst) error {
// 	if SaveAllRankPlayers.RankPlayers.Value == nil {
// 		self.updateAllRankplayers(true)
// 	}

// 	rst.Value = SaveAllRankPlayers.RankPlayers.Value
// 	return nil
// }

// func (self *Center) GetRankClans(req *proto.SLQueryRankPlayers, rst *proto.SLQueryRankPlayersRst) error {
// 	if SaveAllRankPlayers.RankClans.Value == nil {
// 		self.updateAllRankplayers(true)
// 	} else {
// 		self.updateAllRankplayers(false)
// 	}

// 	rst.Value = SaveAllRankPlayers.RankClans.Value
// 	return nil
// }
