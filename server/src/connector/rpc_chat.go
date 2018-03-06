package connector

// import (
// 	"centerclient"
// 	"chatclient"
// 	// "clanclient"
// 	"common"
// 	// "csvcfg"
// 	// "fmt"
// 	// "herobattleclient"
// 	"logger"
// 	// "mailclient"
// 	// "math/rand"
// 	"os"
// 	"proto"
// 	"rpc"
// 	"strconv"
// 	"strings"
// 	"time"
// )

//单点聊天关闭，暂时不开放了
/*func (self *CNServer) ChatP2P(conn rpc.RpcConn, msg rpc.C2SChatP2P) error {

	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}
	canChat, _ := p.beChat()
	if !canChat {
		p.LogError("IDIP limit this player chat!!!")
		return nil
	}

	return chatclient.P2PChat(p.GetUid(), p.GetName(), p.GetSuperLeagueSeg(), msg.GetToPlayerId(), msg.GetChatContent())
}*/

// //世界聊天
// func (self *CNServer) ChatWorld(conn rpc.RpcConn, msg rpc.C2SChatWorld) error {
// 	p, exist := self.getPlayerByConnId(conn.GetId())
// 	if !exist {
// 		return nil
// 	}
// 	canChat, _ := p.beChat()
// 	if !canChat {
// 		p.LogError("IDIP limit this player chat!!!")
// 		return nil
// 	}

// 	//gm指令判断
// 	if self.checkGmCommand(conn, p, msg.GetChatContent()) {
// 		return nil
// 	}

// 	// 防止刷子
// 	if !p.CanWorldChat() {
// 		return nil
// 	}

// 	//// 扣元宝
// 	//chatCost := GetGlocalChatCost()

// 	//if p.GetPlayerTotalGem() < chatCost {
// 	//	return nil
// 	//}

// 	////先确定扣除成功再做后面的操作
// 	//if !p.CostResource(chatCost, proto.ResType_Gem, proto.Lose_WorldChat) {
// 	//	return nil
// 	//}

// 	// cfgCostTiLi := uint32(GetGlobalCfg("WORLD_CHAT_COST_TILI"))

// 	// tiliInfo := p.GetTiliinfo()
// 	// if cfgCostTiLi > tiliInfo.GetTiliNum() {
// 	// 	logger.Error("chat world, have no enough tili", conn.GetId())
// 	// 	return nil
// 	// }
// 	// p.CostResource(cfgCostTiLi, proto.ResType_TiLi, proto.Lose_WorldChat)

// 	timeNow := uint32(time.Now().Unix())
// 	lastRank := uint32(0)
// 	if timeNow < p.GetLastPersonalOverdue() {
// 		lastRank = p.GetLastPersonalRank()
// 	}

// 	//以前是全服务器广播，现在改成本服务器广播
// 	err := chatclient.P2WChat(
// 		p.GetUid(),
// 		p.GetName(),
// 		p.GetSuperLeagueSeg(),
// 		msg.GetChatContent(),
// 		p.GetClan(),
// 		p.GetClanSymbol(),
// 		lastRank,
// 		p.GetGameVipLevel(),
// 		msg.GetUseIM(),
// 		msg.GetVoiceTime(),
// 	)
// 	if err != nil {
// 		p.LogError("ChatWorld error. ", err)
// 		return nil
// 	}

// 	// cmd := &rpc.S2CChatWorld{}
// 	// cmd.SetFromPlayerId(p.GetUid())
// 	// cmd.SetFromPlayerName(p.GetName())
// 	// cmd.SetFromPlayerLevel(p.GetSuperLeagueSeg())
// 	// cmd.SetChatContent(msg.GetChatContent())
// 	// cmd.SetChatTime(time.Now().Unix())
// 	// cmd.SetLastRank(lastRank)
// 	// if p.GetClan() != "" {
// 	// 	cmd.SetAllianceName(p.GetClan())
// 	// 	cmd.SetAllianceSymbol(p.GetClanSymbol())
// 	// }
// 	// cns.serverForClient.ServerBroadcast(cmd)

// 	go TlogSecTalkFlow(p, 1, "", msg.GetChatContent())

// 	return nil
// }

// func (self *CNServer) getPlayerNum(conn rpc.RpcConn, p *player, content string) bool {
// 	if len(content) < 20 || content[:] != "$$@%Wangch_TTdsg_pn$" {
// 		return false
// 	}
// 	mytry := &proto.GetMyself{}
// 	myret := &proto.GetMyselfResult{}

// 	if err := centerclient.Call("Center.GetRankPlayerNum", mytry, myret); err != nil {
// 		return true
// 	}

// 	req := &rpc.ClanChatMessage{}
// 	req.SetType(rpc.ClanChatMessage_Chat)
// 	req.SetUid(p.GetUid())
// 	req.SetName(p.GetName())
// 	req.SetLevel(p.GetSuperLeagueSeg())

// 	selfPower := rpc.Player_ClanPower(0)
// 	req.SetPower(selfPower)
// 	req.SetTime(time.Now().Unix())
// 	timeNow := uint32(time.Now().Unix())
// 	lastRank := uint32(0)
// 	if timeNow < p.GetLastPersonalOverdue() {
// 		lastRank = p.GetLastPersonalRank()
// 	}
// 	req.SetLastRank(lastRank)
// 	req.Args = append(req.Args, "player num:"+strconv.Itoa(myret.Rank))

// 	WriteResult(conn, req)

// 	return true
// }

// //gm指定
// func (self *CNServer) checkGmCommand(conn rpc.RpcConn, p *player, content string) bool {
// 	//检查开关
// 	if !common.IsOpenGm() {
// 		return false
// 	}

// 	logger.Info("content is %s", content)

// 	if len(content) < 2 || content[:2] != "$$" {
// 		return false
// 	}

// 	content = strings.Trim(content[2:], " ")
// 	pos := strings.Index(content, " ")
// 	if pos == -1 {
// 		return false
// 	}

// 	cmd, args := strings.ToLower(content[:pos]), content[pos+1:]
// 	intarg, err := strconv.Atoi(args)

// 	if err != nil && cmd != "pvp" && cmd != "ra" && cmd != "sysn" && cmd != "mailch" && cmd != "rechargeret" && cmd != "ul" && cmd != "vipshop" && cmd != "getvipshop" {
// 		return false
// 	}

// 	logger.Info("cmd is %s", cmd)

// 	switch cmd {
// 	//加钱
// 	case "am":
// 		{
// 			// if intarg > 0 {
// 			// 	p.GainResource(uint32(intarg), proto.ResType_Gold, proto.Gain_GM)
// 			// } else {
// 			// 	p.CostResource(uint32(-intarg), proto.ResType_Gold, proto.Lose_GM)
// 			// }
// 		}

// 	default:
// 		return false
// 	}

// 	return true
// }
// func GetReplayInfo(conn rpc.RpcConn) {
// 	fileName := "fight.data"
// 	file, err := os.Open(fileName)
// 	defer file.Close()
// 	if err != nil {
// 		return
// 	}
// 	buf := make([]byte, 1024*1024)
// 	n, _ := file.Read(buf)
// 	if n == 0 {
// 		return
// 	}
// 	data := buf[:n]
// 	msg := &rpc.ReplayTest{}
// 	dataInfo := string(data)
// 	msg.SetData(dataInfo)

// 	WriteResult(conn, msg)
// }
