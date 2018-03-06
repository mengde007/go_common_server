package roomserver

import (
	conn "centerclient"
	cmn "common"
	"logger"
	"rpc"
	"strconv"
	"sync"
	"time"
)

const (
	JieSanRoomName = "JieSanRoomName"
)

type CustomRoom struct {
	//id             int32  //ID
	owner          string //房间的拥有者(当前的房主)
	creatingPlayer string //房间的创建者（扣他的房卡）
	//name      string //名字
	//pwd       string //密码
	gameType     int32 //游戏类型（大贰，麻将，德州扑克）
	limitCoin    int32 //进入房间的限制
	currencyType int32 //货币类型
	//maxMultiple  int32 //最大倍数
	tiYongAmount int32 //替用数量 -麻将有效

	times                 int32 //游戏场数
	curTimes              int32 //当前场次数
	isAlreadyFinalJieSuan bool  //是否已经最终结算了

	waitingReadyTime          int32 //等待准备的时间(小于等于0：无尽等待)
	startWaitingDissolveTime  int32 //等待解散的时间
	middleWaitingDissolveTime int32 //开场后的等待房间时间

	voteList      []*rpc.JieSanPlayerInfo //玩家解散房间的投票列表
	voteRl        sync.RWMutex            //voteList的读写锁
	voteStartTime int64                   //投票开始时间
	voteDuration  int32                   //投票持续时间

	playerTotalCoin map[string]int32 //玩家总的金币结算信息

}

func (self *CustomRoom) InitCustomRoom( /*id int32,*/ owner string, gameType int32, roomInfo *rpc.CreateRoomREQ) {
	//self.id = id
	self.owner = owner
	self.creatingPlayer = owner
	//r.name = roomInfo.GetName()
	self.gameType = gameType
	//r.pwd = roomInfo.GetPwd()
	//结算货币类型
	self.currencyType = roomInfo.GetCurrencyType()
	//当前结算次数
	self.curTimes = 1
	//结算场次
	self.times = roomInfo.GetTimes()
	//是否已经最终结算了
	self.isAlreadyFinalJieSuan = false
	//统计玩家最后的结算信息
	self.playerTotalCoin = make(map[string]int32, 0)
	//玩家解散房间的投票列表
	self.voteList = []*rpc.JieSanPlayerInfo{}
	//初始化金币限制
	self.limitCoin = 0
	if self.currencyType == CTCoin {
		self.limitCoin = roomInfo.GetLimitCoin()
	}

	//通过配置初始化数据
	self.InitCustomRoomByConfig(gameType, self.currencyType)
}

func (self *CustomRoom) InitCustomRoomByConfig(gameType, currencyType int32) {
	switch currencyType {
	case CTCoin:
		cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(gameType)))
		if cfg != nil {
			self.waitingReadyTime = cfg.CoinWaitingReadyTime
			self.startWaitingDissolveTime = cfg.CoinRoomDissolveTime
			self.middleWaitingDissolveTime = cfg.CoinRoomDissolveTime
		} else {
			logger.Error("读取房间配置表出错ID：%s", gameType)
		}

	case CTCredits:
		cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(gameType)))
		if cfg != nil {
			self.waitingReadyTime = cfg.CreditsWaitingReadyTime
			self.startWaitingDissolveTime = cfg.CreditsStartDissolveTime
			self.middleWaitingDissolveTime = cfg.CreditsMiddleDissolveTime
		} else {
			logger.Error("读取房间配置表出错ID：%s", gameType)
		}

	default:
		logger.Error("不能识别的获取结算类型")
	}

	//投票持续时间
	self.voteDuration = 60
	gcfg := cmn.GetDaerGlobalConfig("507")
	if gcfg != nil {
		self.voteDuration = gcfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}
}

//初始化结算房间投票列表
func (self *CustomRoom) InitVoteList(claimerID string, playerIDs []string) {
	if len(playerIDs) <= 0 {
		logger.Error("players is empty.")
		return
	}

	self.ClearVoteList()

	claimer := &rpc.JieSanPlayerInfo{}
	claimer.SetPlayerID(claimerID)
	claimer.SetStatus(JSClaimer)

	self.voteRl.Lock()
	defer self.voteRl.Unlock()
	self.voteList = append(self.voteList, claimer)

	for _, pid := range playerIDs {
		if pid != claimerID {
			jsp := &rpc.JieSanPlayerInfo{}
			jsp.SetPlayerID(pid)
			jsp.SetStatus(JSWatingDispose)
			self.voteList = append(self.voteList, jsp)
		}
	}
}

func (self *CustomRoom) ClearVoteList() {
	self.voteRl.Lock()
	defer self.voteRl.Unlock()
	self.voteList = []*rpc.JieSanPlayerInfo{}
	self.voteStartTime = time.Now().Unix()
}

func (self *CustomRoom) IsVoting() bool {
	self.voteRl.RLock()
	defer self.voteRl.RUnlock()
	return len(self.voteList) > 0
}

//更新投票列表
func (self *CustomRoom) UpdateVote(uid string, result int32) {
	self.voteRl.Lock()
	defer self.voteRl.Unlock()

	if len(self.voteList) <= 0 {
		logger.Error("还没有人发起解散房间投票！")
		return
	}

	for i, vote := range self.voteList {
		if vote.GetPlayerID() == uid {
			self.voteList[i].SetStatus(result)
			return
		}
	}
}

//获取投票是否结束和是否成功
func (self *CustomRoom) IsVoteEnd() (isEnd, isSuccess bool) {
	self.voteRl.RLock()
	defer self.voteRl.RUnlock()

	if len(self.voteList) <= 0 {
		logger.Error("还没有人发起解散房间投票！")
		return true, false
	}

	//检查是否有人拒绝
	for _, vote := range self.voteList {
		if vote.GetStatus() == JSRefuse {
			return true, false
		}
	}

	//检查是否全部投同意了
	for _, vote := range self.voteList {
		if vote.GetStatus() == JSWatingDispose {
			return false, false
		}
	}

	return true, true
}

func (self *CustomRoom) StatisticsCoin(coins []*rpc.JieSuanCoin) {
	if coins == nil {
		logger.Error("coin is nil.")
		return
	}

	for _, coin := range coins {
		uid := coin.GetPlayerID()
		coin := coin.GetCoin()
		self.playerTotalCoin[uid] += coin
	}
}

//转换最后的结算金币信息
func (self *CustomRoom) GetTotalCoin() []*rpc.JieSuanCoin {
	result := make([]*rpc.JieSuanCoin, 0)
	for uid, coin := range self.playerTotalCoin {
		jieSuanCoin := &rpc.JieSuanCoin{}
		jieSuanCoin.SetPlayerID(uid)
		jieSuanCoin.SetCoin(coin)
		result = append(result, jieSuanCoin)
	}

	return result
}

func ConvertToCustomRoom(gameType int32, room cmn.GameRoom) *CustomRoom {
	if room == nil {
		logger.Error("room is nil.")
		return nil
	}

	switch gameType {
	case cmn.DaerGame:
		cr := room.(*CustomDaerRoom)
		return &cr.CustomRoom
	case cmn.MaJiang:
		cr := room.(*CustomMaJiangRoom)
		return &cr.CustomRoom
	case cmn.DeZhouPuker:
	default:
		logger.Error("不能识别的游戏类型！")
	}

	return nil
}

func GenerateRoomInfo(gameType int32, room cmn.GameRoom) *rpc.RoomInfo {
	//装换为自建房间
	cr := ConvertToCustomRoom(gameType, room)
	if cr == nil {
		logger.Error("转换房间类型错误")
		return nil
	}

	roomInfo := &rpc.RoomInfo{}

	roomInfo.SetId(room.UID())
	//roomInfo.SetName(cr.name)
	roomInfo.SetCurrencyType(cr.currencyType)
	roomInfo.SetGameType(cr.gameType)
	roomInfo.SetDifen(room.GetDifen())
	roomInfo.SetTimes(cr.times)
	//roomInfo.SetHavePwd(cr.pwd != "")
	roomInfo.SetLimitCoin(cr.limitCoin)
	roomInfo.SetMaxMultiple(room.GetMaxMultiple())
	roomInfo.SetIsDaiGui(room.GetIsDaiGui())
	roomInfo.SetTiYongAmount(room.GetTiYongAmount())
	roomInfo.SetQiHuKeAmount(room.GetQiHuKeAmount())

	//roomInfo.SetPlayerCount(room.GetPlayerAmount())

	return roomInfo
}

//网路相关的
//通知投票开启
func (self *CustomRoom) SendJieSanRoomNotify(palyerIDs []string) {
	if len(self.voteList) <= 0 {
		logger.Error("投票列表等于空！")
		return
	}

	msg := &rpc.JieSanRoomNotify{}
	msg.SetRemainTime(int32(int64(self.voteDuration) - (time.Now().Unix() - self.voteStartTime)))
	msg.JieSanPlayerInfo = append(msg.JieSanPlayerInfo, self.voteList...)

	if err := conn.SendCommonNotify2S(palyerIDs, msg, "JieSanRoomNotify"); err != nil {
		logger.Error("发送结束投票通知出错：", err, msg)
	}
}

//通知投票结果改变
func (self *CustomRoom) SendJieSanRoomUpdateStatusNotify(palyerIDs []string, uid string, status int32) {
	if len(palyerIDs) <= 0 {
		logger.Error("palyerIDs is null.")
		return
	}

	msg := &rpc.JieSanRoomUpdateStatusNotify{}
	vote := &rpc.JieSanPlayerInfo{}
	vote.SetPlayerID(uid)
	vote.SetStatus(status)
	msg.SetJieSanPlayerInfo(vote)

	if err := conn.SendCommonNotify2S(palyerIDs, msg, "JieSanRoomUpdateStatusNotify"); err != nil {
		logger.Error("发送结束投票更新通知出错：", err, msg)
	}
}
