package majiangserver

import (
	conn "centerclient"
	//gp "code.google.com/p/goprotobuf/proto"
	cmn "common"
	"logger"
	"math"
	"math/rand"
	"rpc"
	"strconv"
	"sync"
	"time"
	"timer"
)

const (
	RoomStayTimeName  = "RoomStayTime"
	StartGameName     = "StartGame"
	MainTimerName     = "MainTimer"
	GamingMaxTimeName = "GamingMaxTime"
)

type MaJiangRoom struct {
	uid                  int32                               //房间的唯一标示
	rtype                int32                               //房间类型(即配置表的ID)
	state                int32                               //房间的状态
	players              [RoomMaxPlayerAmount]*MaJiangPlayer //房间的所有玩家
	activePlayerIndex    int                                 //当前的玩家
	lc                   *LicensingController                //发牌控制器
	curAroundState       *RoomAroundState                    //当前这一轮的状态信息
	activeCard           *MaJiangCard                        //当前活动的卡牌
	gamingPlayerIDs      [RoomMaxPlayerAmount]string         //进行游戏时的玩家ID列表，主要用于定庄
	nextBankerPlayerID   string                              //胡牌的玩家ID -也用于定庄
	Difen                int32                               //底注
	IsDaigui             bool                                //是否带归
	MaxMultiple          int32                               //倍数上限
	RakeRate             int32                               //抽成比率
	QiHuKeAmount         int32                               //房间的起胡颗数
	IsAntiCheating       bool                                //是否防作弊
	TotalHongZhongAmount int32                               //红中数量
	TimerInterval        int64                               //给玩家考虑的时间
	RoomStayTime         int64                               //房间等待的最大时间
	GameStartDelay       int64                               //延迟开始游戏
	DoActionDelay        int64                               //执行动作的延迟
	OpenCardDelay        int64                               //开一张牌的延迟
	GamingMaxTime        int64                               //游戏最大进行时间

	rs       cmn.RoomSelector   //房间选择器（分支器）
	timeMgr  *cmn.TimerMgr      //定时器管理器
	t        *timer.Timer       //定时器
	msgQueue []cmn.RoomMsgQueue //消息队列
	rcv      chan cmn.RoomMsgQueue
	exit     chan bool
	ql       sync.Mutex
}

//创建函数
//type CreateMsgFun func() interface{}

//var msgReflect map[string]CreateMsgFun

func init() {
	// msgReflect = make(map[string]CreateMsgFun)

	// msgReflect["ActionREQ"] = func() interface{} { return &rpc.ActionREQ{} }

}

//新建一个麻将房间,注意使用的都是大贰的配置表所有使用的是GetDaerRoomConfig
func NewMajiangRoom(uid, rtype int32) *MaJiangRoom {
	r := new(MaJiangRoom)

	//初始化
	r.Init(uid, rtype)

	//创建发牌器
	r.lc = NewLicensingController(r.TotalHongZhongAmount)

	return r
}

//初始化房间
func (room *MaJiangRoom) Init(uid, rtype int32) {

	room.rtype = rtype
	room.state = RSReady
	room.uid = uid

	room.timeMgr = cmn.NewTimerMgr()
	room.timeMgr.SetUpdateInterval(time.Millisecond * 1000)
	room.curAroundState = NewRoomAroundState()
	room.SetSelector(room)
	room.msgQueue = []cmn.RoomMsgQueue{}
	room.rcv = make(chan cmn.RoomMsgQueue, 10)

	room.InitByConfig(rtype)

	timerInterval := time.Millisecond * 50
	room.t = timer.NewTimer(timerInterval)
	room.t.Start(
		func() {
			room.UpdateTimer(timerInterval)
		},
	)
	go room.process()
}

//初始化配置表中提供的数据
func (room *MaJiangRoom) InitByConfig(rtype int32) {
	//获取底注
	room.Difen = 100
	cfg := cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		room.Difen = cfg.Difen
	} else {
		logger.Error("读取房间配置表出错ID：%s", room.rtype)
	}

	//获取是否带归
	room.IsDaigui = true
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		if cfg.Type == cmn.RTDaerHight {
			room.IsDaigui = true
		} else {
			room.IsDaigui = false
		}
	} else {
		logger.Error("读取房间配置表出错ID：%s", rtype)
	}

	//获取最高倍数
	room.MaxMultiple = 10
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		room.MaxMultiple = cfg.MaxMultiple
	} else {
		logger.Error("读取房间配置表出错ID：%s", room.rtype)
	}

	//获取抽成比利
	room.RakeRate = 10
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		room.RakeRate = cfg.RakeRate
	} else {
		logger.Error("读取房间配置表出错ID：%s", room.rtype)
	}

	//获取是否防作弊
	room.IsAntiCheating = false
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		if cfg.AntiCheating == 0 {
			room.IsAntiCheating = false
		} else {
			room.IsAntiCheating = true
		}
	} else {
		logger.Error("读取房间配置表出错ID：%s", room.rtype)
	}

	//获取起胡颗数
	room.QiHuKeAmount = 1
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		room.QiHuKeAmount = cfg.QiHuKeAmount
	} else {
		logger.Error("读取房间配置表出错ID：%s", room.rtype)
	}

	//获取红中的数量
	room.TotalHongZhongAmount = 8
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		room.TotalHongZhongAmount = cfg.HongZhongAmount
	} else {
		logger.Error("读取房间配置表出错ID：%s", room.rtype)
	}

	//初始化全局变量--基础
	//给玩家考虑的时间
	room.TimerInterval = 180
	gcfg := cmn.GetDaerGlobalConfig("501")
	if gcfg != nil {
		room.TimerInterval = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//延迟开始游戏
	room.GameStartDelay = 6
	gcfg = cmn.GetDaerGlobalConfig("502")
	if gcfg != nil {
		room.GameStartDelay = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//执行动作的延迟
	room.DoActionDelay = 1
	gcfg = cmn.GetDaerGlobalConfig("503")
	if gcfg != nil {
		room.DoActionDelay = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//开一张牌的延迟
	room.OpenCardDelay = 2
	gcfg = cmn.GetDaerGlobalConfig("504")
	if gcfg != nil {
		room.OpenCardDelay = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//房间等待的最大时间
	room.RoomStayTime = 120
	gcfg = cmn.GetDaerGlobalConfig("505")
	if gcfg != nil {
		room.RoomStayTime = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//游戏最大进行时间
	room.GamingMaxTime = 1200
	gcfg = cmn.GetDaerGlobalConfig("506")
	if gcfg != nil {
		room.GamingMaxTime = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}
}

func (room *MaJiangRoom) handleMsg() {
	room.ql.Lock()
	defer room.ql.Unlock()

	if len(room.msgQueue) <= 0 {
		return
	}

	bLoop := true
	for bLoop {
		bLoop = false
		logger.Info("handleMsg 中处理消息， 消息数量为：", len(room.msgQueue))
		for index, e := range room.msgQueue {
			if e.Func == "ActionREQ" {
				msg := e.Msg
				s, ok := msg.(*rpc.ActionREQ)
				if !ok {
					logger.Error("handleMsg  msg.(rpc.ActionREQ) error")
					return
				}

				//room.OnPlayerDoAction(s)
				room.rs.OnPlayerDoAction(s)
				room.msgQueue = append(room.msgQueue[:index], room.msgQueue[index+1:]...)
				bLoop = true
				break
			} else if e.Func == "TieGuiREQ" {
				msg := e.Msg
				s, ok := msg.(*rpc.MJTieGuiREQ)
				if !ok {
					logger.Error("handleMsg  msg.(rpc.MJTieGuiREQ) error")
					return
				}

				room.OnTieGuiREQ(s)
				room.msgQueue = append(room.msgQueue[:index], room.msgQueue[index+1:]...)
				bLoop = true
				break
			} else if e.Func == "Enter" {
				msg := e.Msg
				p, ok := msg.(*MaJiangPlayer)
				if !ok {
					logger.Error("handleMsg  msg.(rpc.PlayerBaseInfo) error")
					return
				}

				//room.Enter(p)
				room.rs.Enter(p)
				room.msgQueue = append(room.msgQueue[:index], room.msgQueue[index+1:]...)

				bLoop = true
				break
			} else if e.Func == "ReEnter" {
				msg := e.Msg
				s, ok := msg.(*rpc.PlayerBaseInfo)
				if !ok {
					logger.Error("handleMsg  msg.(rpc.PlayerBaseInfo) error")
					return
				}

				//room.ReEnter(s.GetUid(), s)
				room.rs.ReEnter(s.GetUid(), s)
				room.msgQueue = append(room.msgQueue[:index], room.msgQueue[index+1:]...)

				bLoop = true
				break
			} else if e.Func == "Leave" || e.Func == "Kick" {
				msg := e.Msg
				uid, uiOk := msg.(string)
				if !uiOk {
					logger.Error("handleMsg  msg.(string) error")
					return
				}

				isChangeDesk, changeOk := e.Msg2.(bool)
				if !changeOk {
					logger.Error("handleMsg  msg.(bool) error")
					return
				}

				//room.Leave(uid, isChangeDesk)
				room.rs.Leave(uid, isChangeDesk)

				if e.Func == "Kick" {
					//通知gameserver道具
					if err := conn.SendCostResourceMsg(uid, strconv.Itoa(cmn.KickCardID), "majiang", -1); err != nil {
						logger.Error("发送扣取踢人卡出错：", err)
					}
				}

				room.msgQueue = append(room.msgQueue[:index], room.msgQueue[index+1:]...)

				bLoop = true

				break
			} else {
				logger.Error("未处理的命令")
			}
		}
	}
}

func (room *MaJiangRoom) process() {
	for {
		select {
		case r := <-room.rcv:
			logger.Info("process 收到消息")
			room.ql.Lock()
			room.msgQueue = append(room.msgQueue, r)
			room.ql.Unlock()
		case <-room.exit:
			room.t.Stop()
			logger.Info("退出房间的接受消息线程：", room.UID())
			return
		}
	}

}

//重置房间
func (room *MaJiangRoom) ResetRoom() {
	room.state = RSReady
	room.activeCard = nil
	room.activePlayerIndex = 0
	//room.ownCards = make([]*MaJiangCard, 0)

	room.timeMgr.Clear()
	room.curAroundState.ClearAll()
	room.lc = NewLicensingController(room.TotalHongZhongAmount)

	// //初始化牌的状态
	// for _, v := range room.cards {
	// 	v.flag = cmn.CBack
	// 	v.owner = nil
	// }

	//重置玩家状态
	room.ResetForAllPlayer()
}

//进入房间
func (room *MaJiangRoom) Enter(p cmn.Player) {
	//检查输入参数
	if p == nil {
		logger.Error("player is nil.")
		return
	}

	logger.Info("调用进入房间：MaJiangRoom。Enter:", p.ID())

	player := p.(*MaJiangPlayer)
	if player == nil {
		logger.Error("接口转换失败")
		return
	}

	//检查能否进入房间
	if room.IsFull() {
		logger.Error("MaJiangRoom.Enter: room is full")
		return
	}

	//修改玩家的房间相关的信息
	player.room = room
	for i := 0; i < RoomMaxPlayerAmount; i++ {
		if room.players[i] == nil {
			room.players[i] = player
			room.SendEnterRoomACK(player)
			logger.Info("MaJiangRoom.Enter: player:%s enter room(%s):", player.id, i)
			break
		}
	}

	//开启一个房间停留计时器
	room.StartDelayCallback(RoomStayTimeName, room.RoomStayTime, func(data interface{}) {
		room.rs.ForceAllPlayerLeave()
	}, nil)
}

//重新进入房间
func (room *MaJiangRoom) ReEnter(playerID string, playerInfo *rpc.PlayerBaseInfo) {
	//TODO:重新建立玩家连接，并下发当前玩家的数据
	if player := room.GetPlayerByID(playerID); player != nil {
		player.client = playerInfo
		room.SendEnterRoomACK(player)
		if room.IsGaming() {
			player.SendGameStartACK(true)
		} else {
			//开启一个房间停留计时器
			room.StartDelayCallback(RoomStayTimeName, room.RoomStayTime, func(data interface{}) {
				room.rs.ForceAllPlayerLeave()
			}, nil)
		}
	} else {
		logger.Error("MaJiangRoom:player not in the room")
	}
}

// 离开房间
func (room *MaJiangRoom) Leave(uid string, isChangeDesk bool) bool {
	logger.Info("离开房间：", uid)
	//检查输入参数
	if uid == "" {
		logger.Error("player is nil.")
		return false
	}

	leavePlayer := room.GetPlayerByID(uid)
	if leavePlayer == nil {
		logger.Error("在房间里没有查找到制定的玩家：", uid)
		return false
	}

	if room.IsGaming() {
		return false
	}

	//重置一下当前胡牌的玩家
	room.nextBankerPlayerID = ""

	for i := 0; i < RoomMaxPlayerAmount; i++ {
		tempPlayer := room.players[i]
		if tempPlayer != nil && uid == tempPlayer.id {
			room.SendLeaveRoomACK(tempPlayer, isChangeDesk)
			tempPlayer.room = nil
			room.players[i] = nil
			tempPlayer.Reset()
			logger.Info("MaJiangRoom.Leave: player:%s Leave room:", tempPlayer.id)
			break
		}
	}

	//开启一个房间停留计时器
	room.StartDelayCallback(RoomStayTimeName, room.RoomStayTime, func(data interface{}) {
		room.rs.ForceAllPlayerLeave()
	}, nil)

	//删除roommgr中的引用关系
	maJiangRoomMgr.DeleteLeavePlayerInfo(room.rtype, uid)

	//检查是否是换桌
	if isChangeDesk {
		maJiangRoomMgr.EnterGame(room.rtype, leavePlayer.client, isChangeDesk)
	}

	//检查房间是否为空了，如果为空就结算房间
	if room.IsEmpty() {
		logger.Error("停止房间的Timer和消息接受线程")
		//room.ResetRoom()
		//room.t.Stop()
		room.exit <- true
	}

	return true
}

//指定时间玩家都还没有开始游戏，那么将所有人提出房间
func (room *MaJiangRoom) ForceAllPlayerLeave() {
	if maJiangRoomMgr == nil {
		return
	}

	for _, p := range room.players {
		if p != nil {
			maJiangRoomMgr.LeaveGame(p.id, false)
		}
	}
}

//强制玩家离开房间
func (room *MaJiangRoom) ForcePlayerLeave(uid string) {
	if maJiangRoomMgr == nil {
		return
	}

	for _, p := range room.players {
		if p != nil && p.id == uid {
			maJiangRoomMgr.LeaveGame(uid, false)
		}
	}
}

//开始游戏
func (room *MaJiangRoom) StartGame() {
	logger.Info("=======开始游戏-=======")

	//重置房间
	//room.ResetRoom()

	//缓存游戏开始时的玩家
	room.CachePlayerID()

	//开始洗牌
	room.lc.Shuffle()

	//开始发牌
	room.lc.Licensing(room.ConvertAllPlayersToSlice(false, nil))
	for _, p := range room.players {
		PrintCardsS("发完牌后玩家的手牌："+p.id, p.cards)
	}

	//定庄
	room.DecideBanker()

	//更新计算胡控制器
	room.UpdateHuControllerForAll()

	//向所有玩家发送开始游戏消息
	room.sendGameStartForAll()

	//切换到报牌阶段
	room.SwitchRoomState(RSBankerTianHuStage)

	//延迟触发庄家天胡阶段的检测
	room.StartDelayCallback(StartGameName, room.GameStartDelay, func(data interface{}) {
		logger.Info("进入庄家天胡检查...，并启动第一个检查....")
		banker := room.GetBanker()
		openCard := room.OpenOneCard(banker)
		if openCard == nil {
			logger.Error("开第一张牌就为空了")
			return
		}

		PrintCardS("开的第一张牌是：", openCard)

		//缓存庄家的第一张牌
		banker.aroundState.moCard = openCard

		//通知庄家摸第14张牌
		banker.SendActionACK(AMo, openCard, nil, ACSuccess)

		//检查有没得动作
		room.CheckDoAction(openCard, banker, nil, true)
	}, nil)

	//开启游戏后执行的一些操作
	room.rs.OnStartGameAfter()

}

//缓存房间玩家ID--用于定庄检查
func (room *MaJiangRoom) CachePlayerID() {
	if !room.IsFull() {
		logger.Error("开始游戏的时候，房间人数没有满")
		return
	}

	for i := 0; i < RoomMaxPlayerAmount; i++ {
		room.gamingPlayerIDs[i] = room.players[i].id
	}
}

//定庄
func (room *MaJiangRoom) DecideBanker() {

	bankerIndex := 0

	if room.IsPlayerHaveChange() || room.nextBankerPlayerID == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		bankerIndex = r.Intn(RoomMaxPlayerAmount)
	} else {
		p := room.GetPlayerByID(room.nextBankerPlayerID)
		if p != nil {
			bankerIndex = room.GetPlayerIndex(p)
		} else {
			logger.Error("在定庄时，没有获取到胡牌玩家：%s 的玩家信息", room.nextBankerPlayerID)
		}
	}

	if FixedBankerIndex >= 0 {
		bankerIndex = FixedBankerIndex
	}

	room.activePlayerIndex = bankerIndex
	ap := room.players[bankerIndex]
	if ap == nil {
		logger.Error("定庄的索引(%s)没有玩家", bankerIndex)
		return
	}

	ap.ptype = cmn.PTBanker

	room.nextBankerPlayerID = ap.id
}

//更新胡控制器
func (room *MaJiangRoom) UpdateHuControllerForAll() {
	for _, p := range room.players {
		if p == nil {
			continue
		}
		p.controller.huController.UpdateData(p.cards)
	}
}

//开启游戏后
func (room *MaJiangRoom) OnStartGameAfter() {
	//停止房间停留计时器
	room.StopDelayCallback(RoomStayTimeName)

	//开启最大游戏进行时间的延迟，防止房间被卡死，最大时间到了，房间自动解散
	room.StartDelayCallback(GamingMaxTimeName, room.GamingMaxTime, func(data interface{}) {
		room.SwitchRoomState(RSReady)
		room.rs.ForceAllPlayerLeave()
	}, nil)
}

//房间的玩家是否变化
func (room *MaJiangRoom) IsPlayerHaveChange() (result bool) {

	for _, p := range room.players {
		if p == nil {
			result = true
			break
		}

		find := false
		for _, id := range room.gamingPlayerIDs {
			if id == p.id {
				find = true
				break
			}
		}

		if !find {
			result = true
			break
		}
	}

	return
}

//重置所有玩家状态
func (room *MaJiangRoom) ResetForAllPlayer() {
	for _, player := range room.players {
		if player == nil {
			continue
		}

		player.Reset()
	}
}

//检测一个玩家时候是活动玩家
func (room *MaJiangRoom) IsActivePlayer(player *MaJiangPlayer) bool {
	if player == nil {
		return false
	}

	activePalyer := room.players[room.activePlayerIndex]
	return activePalyer != nil && activePalyer.id == player.id
}

//获取活动玩家
func (room *MaJiangRoom) GetActivePlayer() *MaJiangPlayer {
	return room.players[room.activePlayerIndex]
}

//改变活动玩家
func (room *MaJiangRoom) ChangeActivePlayerByIndex(index int) {
	if index < 0 || index > RoomMaxPlayerAmount {
		logger.Error("index out of")
		return
	}

	player := room.players[index]
	if player == nil {
		logger.Error("设置的活动玩家是一个nil值")
		return
	}

	if player.IsHu() {
		logger.Error("已经胡牌的玩家是不能作为活动玩家的！")
		return
	}

	room.activePlayerIndex = index
}

//改变活动玩家
func (room *MaJiangRoom) ChangeActivePlayerToNext() {

	for i := 1; i <= RoomMaxPlayerAmount; i++ {
		api := (room.activePlayerIndex + i) % RoomMaxPlayerAmount
		player := room.players[api]
		if player == nil {
			continue
		}

		if player.IsHu() {
			continue
		}

		logger.Info("改变活动玩家到：", api)
		room.activePlayerIndex = api
		break
	}

}

//改变活动玩家
func (room *MaJiangRoom) ChangeActivePlayerTo(p cmn.Player) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*MaJiangPlayer)
	if player == nil {
		logger.Error("接口转换失败")
		return
	}

	if player.IsHu() {
		logger.Error("已经胡牌的玩家是不能够成为活动玩家的！")
		return
	}

	for i, v := range room.players {
		if v != nil && player.id == v.id {
			room.activePlayerIndex = i
			return
		}
	}

	logger.Error("在切换活动玩家的时候，没有找到指定的玩家", player.id)
}

//获取一个玩家所在的索引
func (room *MaJiangRoom) GetPlayerIndex(player *MaJiangPlayer) int {
	if player == nil {
		logger.Error("player is nil.")
		return -1
	}

	for i, v := range room.players {
		if v != nil && player.id == v.id {
			return i
		}
	}

	return -1

}

//获取所有玩家
func (room *MaJiangRoom) GetAllPlayer() *[RoomMaxPlayerAmount]*MaJiangPlayer {
	return &room.players
}

//转换所有晚间到slice
func (room *MaJiangRoom) ConvertAllPlayersToSlice(isIncludeAlreadyHu bool, excludePlayers []*MaJiangPlayer) []*MaJiangPlayer {

	result := make([]*MaJiangPlayer, 0)

	for _, p := range room.players {
		if p == nil || (!isIncludeAlreadyHu && p.IsHu()) {
			continue
		}

		if !IsExistPlayer(excludePlayers, p) {
			result = append(result, p)
		}
	}

	return result
}

//通过胡牌的玩家，确定下个活动玩家
func (room *MaJiangRoom) GetNextActivePlayerByHuPlayers(huPlayers []*MaJiangPlayer) *MaJiangPlayer {
	//检查输入参数的合法性
	if huPlayers == nil || len(huPlayers) <= 0 {
		logger.Error("huPlayers is nil.")
		return nil
	}

	activePlayer := room.GetActivePlayer()
	if activePlayer == nil {
		logger.Error("activePlayer is null.")
		return nil
	}

	checkEndPlayer := activePlayer

	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {
		for i, hp := range huPlayers {
			if hp.id == player.id {
				huPlayers = append(huPlayers[:i], huPlayers[i+1:]...)
				break
			}
		}

		player = player.GetXiaJia()
		isInit = false

		if len(huPlayers) <= 0 {
			if player.IsHu() {
				continue
			}
			return player
		}
	}

	//如果找这里到还没找到下个活动玩家，那么再检查第一个而玩家（checkEndPlayer：在进入函数后就标记成了第一个玩家）
	if len(huPlayers) <= 0 && checkEndPlayer != nil && !checkEndPlayer.IsHu() {
		return checkEndPlayer
	}

	logger.Error("胡牌后没有获取到下一个活动玩家!  胡牌的玩家数量剩余：%d，最有一个玩家是否胡牌：%s", len(huPlayers), checkEndPlayer.IsHu())
	return nil
}

//获取胡牌的玩家及数量
func (room *MaJiangRoom) GetPlayerOfHu() (playerAmount int32, result []*MaJiangPlayer) {
	result = make([]*MaJiangPlayer, 0)

	for _, player := range room.players {
		if player != nil && player.IsHu() {
			result = append(result, player)
		}
	}

	return int32(len(result)), result
}

//获取庄家
func (room *MaJiangRoom) GetBanker() *MaJiangPlayer {
	for _, player := range room.players {
		if player != nil && player.ptype == cmn.PTBanker {
			return player
		}
	}
	return nil
}

//获取等待执行动作和准备执行动作的玩家
func (room *MaJiangRoom) GetHaveSpecificActionPlayers(action int32) (result []*MaJiangPlayer) {

	result = make([]*MaJiangPlayer, 0)
	for _, p := range room.players {
		if p == nil {
			continue
		}

		if Exist(p.watingAction, action) || p.readyDoAction == action {
			result = append(result, p)
		}
	}

	return
}

//检查一个玩家是否在房间
func (room *MaJiangRoom) IsInRoom(uid string) bool {
	for _, v := range room.players {
		if v != nil && uid == v.id {
			return true
		}
	}

	return false
}

//获取报牌的人数
func (room *MaJiangRoom) GetAmountOfBaoPai() int32 {
	var result int32 = 0
	for _, p := range room.players {
		if p == nil {
			continue
		}

		if p.HaveBao() {
			result++
		}
	}

	return result
}

//通过ID获取一个玩家
func (room *MaJiangRoom) GetPlayerByID(id string) *MaJiangPlayer {
	for _, player := range room.players {
		if player != nil && player.id == id {
			return player
		}
	}

	return nil
}

//获取房间人数
func (room *MaJiangRoom) GetPlayerAmount() int32 {
	var result int32 = 0
	for _, v := range room.players {
		if v != nil {
			result += 1
		}
	}
	return result
}

//玩家数量上限
func (room *MaJiangRoom) MaxPlayerAmount() int32 {
	return RoomMaxPlayerAmount
}

//房间是否满员
func (room *MaJiangRoom) IsFull() bool {
	return room.GetPlayerAmount() >= RoomMaxPlayerAmount
}

//房间是否为空
func (room *MaJiangRoom) IsEmpty() bool {
	return room.GetPlayerAmount() <= 0
}

//获取有游戏类型
func (room *MaJiangRoom) GetGameType() int32 {
	return cmn.MaJiang
}

//是否正在游戏
func (room *MaJiangRoom) IsGaming() bool {
	return room.state != RSReady && room.state != RSSettlement
}

//所有玩家都准备好了吗
func (room *MaJiangRoom) IsReadyForAll() bool {
	result := true
	for _, v := range room.players {
		if v != nil && !v.isReady {
			result = false
			break
		}
	}

	return result
}

//是否是黄庄
func (room *MaJiangRoom) IsHuangZhuang() bool {

	result := true

	for _, p := range room.players {
		if p != nil {
			if p.IsHu() || p.HaveJiao() || p.isChaJiaoHu {
				result = false
				break
			}
		}
	}

	return result
}

//能够开始游戏了吗
func (room *MaJiangRoom) CanStartGame() bool {
	if room.state != RSReady {
		return false
	}

	if !room.IsFull() {
		return false
	}

	if !room.IsReadyForAll() {
		return false
	}

	return true
}

//从座面上开一张牌
func (room *MaJiangRoom) OpenOneCard(player *MaJiangPlayer) *MaJiangCard {
	card := room.lc.OpenOneCard(player)
	if card != nil {
		card.owner = player
	}

	room.activeCard = card
	return card
}

//当前这一把是否完成(没牌或者只剩下一家)
func (room *MaJiangRoom) IsOverForAround() bool {
	huPaiPlayerAmount, _ := room.GetPlayerOfHu()
	notHuPalyerAmount := RoomMaxPlayerAmount - huPaiPlayerAmount
	return room.lc.RemainCardAmount() <= 0 || notHuPalyerAmount <= 1
}

//切换房间状态
func (room *MaJiangRoom) SwitchRoomState(state int32) {
	room.state = state
}

//检查所有玩家能执行的动作,并下发给玩家
func (room *MaJiangRoom) CheckCanDoActionAndNotifyPlayer(card *MaJiangCard, onlyCheckPlayer *MaJiangPlayer, excludePlayers []*MaJiangPlayer, onlyCheckHuAction bool) (result []int32) {

	result = make([]int32, 0)

	//获取需要检查的玩家
	checkPlayers := make([]*MaJiangPlayer, 0)
	if onlyCheckPlayer != nil {
		if !onlyCheckPlayer.IsHu() {
			checkPlayers = append(checkPlayers, onlyCheckPlayer)
		} else {
			logger.Error("这个玩家已经胡牌了，不能再进行动作的检查了！")
		}
	} else {
		checkPlayers = append(checkPlayers, room.ConvertAllPlayersToSlice(false, excludePlayers)...)
	}

	//检查能执行的动作
	switch room.state {
	case RSBankerTianHuStage:
		//检查庄家出完第一张后还能不能报牌
		banker := room.GetBanker()
		if banker == nil {
			logger.Error("竟然没有庄家")
			break
		}

		if canHu, _ := banker.controller.CheckHuSpecific(card); canHu {
			result = append(result, AHu)
			//cards := make(map[int32][]*MaJiangCard, 0)
			//cards[AHu] = []*MaJiangCard{card}
			banker.SendActionNotifyACK(card, []int32{AHu}, nil)
			logger.Info("CheckCanDoActionAndNotifyPlayer:玩家%s,检查到:%s", banker.id, actionName[AHu])
		}

	case RSNotBankerBaoPaiStage:
		for _, p := range checkPlayers {
			if p == nil || p.ptype == cmn.PTBanker || p.HaveBao() {
				continue
			}

			if canBao, _ := p.controller.CheckBao(); canBao {
				result = append(result, ABao)
				//cards := make(map[int32][]*MaJiangCard, 0)
				//cards[ABao] = []*MaJiangCard{card}
				p.SendActionNotifyACK(card, []int32{ABao}, nil)
				logger.Info("CheckCanDoActionAndNotifyPlayer:玩家%s,检查到:%s", p.id, actionName[ABao])
			}
		}
	case RSBankerChuPaiStage:
		break
	case RSBankerBaoPaiStage:
		//检查庄家出完第一张后还能不能报牌
		banker := room.GetBanker()
		if banker == nil {
			logger.Error("竟然没有庄家")
			break
		}

		//胡了牌就不用检查报了
		if banker.IsHu() {
			break
		}

		//检查庄家报牌
		if canBao, _ := banker.controller.CheckBao(); canBao {
			result = append(result, ABao)
			//cards := make(map[int32][]*MaJiangCard, 0)
			//cards[ABao] = []*MaJiangCard{card}
			banker.SendActionNotifyACK(card, []int32{ABao}, nil)
			logger.Info("CheckCanDoActionAndNotifyPlayer:玩家%s,检查到:%s", banker.id, actionName[ABao])
		}
	case RSLoopWorkStage:
		//检测胡，碰，杠。
		doActions := make(map[string]map[int32][]*MaJiangCard, 0)
		//胡, 暗杠，补杠,明杠
		if card == nil {
			break
		}

		//胡，暗杠，补杠,明杠,碰
		for _, p := range checkPlayers {
			//红中牌只有自己能够进行胡
			if card.IsHongZhong() && card.owner != nil && card.owner.id != p.id {
				continue
			}

			//胡
			if canHu, _ := p.controller.CheckHuSpecific(card); canHu {
				result = append(result, AHu)
				doActions = AddActionAgrs(doActions, p.id, AHu, []*MaJiangCard{card})
			}

			if onlyCheckHuAction {
				continue
			}

			//自己的牌或者自己莫得牌
			if card.owner == nil || card.owner.id == p.id {
				if !room.IsActivePlayer(p) {
					continue
				}

				if canAnGang, cards := p.controller.CheckAnGang(card); canAnGang {
					result = append(result, AAnGang)
					doActions = AddActionAgrs(doActions, p.id, AAnGang, cards)
				}

				if p.HaveBao() {
					continue
				}

				if canBuGang, cards := p.controller.CheckBuGang(card); canBuGang {
					result = append(result, ABuGang)
					doActions = AddActionAgrs(doActions, p.id, ABuGang, cards)
				}
			} else {
				if canMingGang, isNeedHongZhong := p.controller.CheckMingGang(card); canMingGang {
					if isNeedHongZhong {
						result = append(result, ATieMingGang)
						doActions = AddActionAgrs(doActions, p.id, ATieMingGang, []*MaJiangCard{card})
					} else {
						result = append(result, AMingGang)
						doActions = AddActionAgrs(doActions, p.id, AMingGang, []*MaJiangCard{card})
					}
				}

				if p.HaveBao() {
					continue
				}

				if canPeng, isNeedHongZhong := p.controller.CheckPeng(card); canPeng {

					if isNeedHongZhong {
						result = append(result, ATiePeng)
						doActions = AddActionAgrs(doActions, p.id, ATiePeng, []*MaJiangCard{card})
					} else {
						result = append(result, APeng)
						doActions = AddActionAgrs(doActions, p.id, APeng, []*MaJiangCard{card})
					}
				}
			}

		}

		//发送动作通知给客服端
		for pid, _ := range doActions {
			p := room.GetPlayerByID(pid)
			if p == nil {
				logger.Error("不能够找到ID为：%s 的玩家", pid)
				continue
			}

			actions, cards := GetActionsArgs(doActions, pid)

			if actions != nil && len(actions) > 0 {
				p.SendActionNotifyACK(card, actions, cards)
				logger.Info("CheckCanDoActionAndNotifyPlayer:玩家%s,检查到:%s", p.id, actionName[AHu])
			}
		}
	}

	//通知倒计时
	if len(result) > 0 {
		room.sendCountdownNotifyACK()
	}

	return result
}

func AddActionAgrs(doActions map[string]map[int32][]*MaJiangCard, uid string, action int32, cards []*MaJiangCard) map[string]map[int32][]*MaJiangCard {

	//检查输入参数
	if doActions == nil {
		logger.Error("doActions is empty.")
		return nil
	}

	if cards == nil || len(cards) <= 0 {
		logger.Error("cards is empty.")
		return nil
	}

	//添加数据的列表
	if _, exist := doActions[uid]; !exist {
		doActions[uid] = make(map[int32][]*MaJiangCard, 0)
	}

	if _, exist := doActions[uid][action]; !exist {
		doActions[uid][action] = []*MaJiangCard{}
	}

	doActions[uid][action] = append(doActions[uid][action], cards...)

	return doActions
}

func GetActionsArgs(doActions map[string]map[int32][]*MaJiangCard, uid string) (actions []int32, cards map[int32][]*MaJiangCard) {
	//检查输入参数
	if doActions == nil {
		logger.Error("doActions is empty.")
		return
	}

	//查找数据
	if _, exist := doActions[uid]; !exist {
		return
	}

	cards = make(map[int32][]*MaJiangCard, 0)
	for k, v := range doActions[uid] {
		actions = append(actions, k)
		cards[k] = v
	}

	return
}

//执行一个动作，根据当前等待的动作
func (room *MaJiangRoom) CheckDoAction(card *MaJiangCard, onlyCheckPlayer *MaJiangPlayer, excludePlayers []*MaJiangPlayer, onlyCheckHuAction bool) {
	if !room.CanCheckDoAction() {
		return
	}

	if card != nil {
		logger.Info("CheckDoAction:开始执行一次检测==============当前的房间状态：%s 检查的牌card:%s", roomTypeName[room.state], ConvertToWord(card))
	} else {
		logger.Info("CheckDoAction:开始执行一次检测=====当前的房间状态：%s", roomTypeName[room.state])
	}

	canDoActions := room.CheckCanDoActionAndNotifyPlayer(card, onlyCheckPlayer, excludePlayers, onlyCheckHuAction)
	haveAction := len(canDoActions) > 0
	if card != nil {
		if haveAction {
			logger.Info("CheckDoAction:执行一次检测=====当前的房间状态：%s 准备执行的动作：%s card:%s", roomTypeName[room.state], CnvtActsToStr(canDoActions), ConvertToWord(card))
		} else {
			logger.Info("CheckDoAction:执行一次检测=====当前的房间状态：%s 准备执行的动作:%s card:%s", roomTypeName[room.state], "无", ConvertToWord(card))
		}
	} else {
		if haveAction {
			logger.Info("CheckDoAction:执行一次检测=====当前的房间状态：%s 准备执行的动作：%s", roomTypeName[room.state], CnvtActsToStr(canDoActions))
		} else {
			logger.Info("CheckDoAction:执行一次检测=====当前的房间状态：%s 准备执行的动作:%s", roomTypeName[room.state], "无")
		}
	}

	switch room.state {
	case RSBankerTianHuStage:
		if !haveAction {
			room.DoMoByCache(false)
			PrintCardsS("庄家天胡阶段没有天胡是，把牌摸到手上后，手上的牌是：", room.GetBanker().cards)
			room.SwitchRoomState(RSNotBankerBaoPaiStage)
			room.CheckDoAction(nil, nil, nil, false)
		}
	case RSNotBankerBaoPaiStage:
		if !haveAction {
			banker := room.GetBanker()
			if banker == nil {
				logger.Error("竟然没有庄家！太不可思议了")
				return
			}

			//庄家天胡了
			if banker.IsHu() {
				room.SwitchRoomState(RSLoopWorkStage)
				room.ChangeActivePlayerToNext()
				room.CheckDoAction(nil, nil, nil, false)
			} else {
				room.SwitchRoomState(RSBankerChuPaiStage)
				banker.SendActionNotifyACK(nil, []int32{AChu}, nil)
			}
		}
	case RSBankerChuPaiStage:
		if !haveAction {
			room.SwitchRoomState(RSBankerBaoPaiStage)
			room.CheckDoAction(nil, nil, nil, false)
		}
	case RSBankerBaoPaiStage:
		if !haveAction {
			room.SwitchRoomState(RSLoopWorkStage)
			room.ChangeActivePlayerToNext()
			room.CheckDoAction(room.activeCard, nil, []*MaJiangPlayer{room.GetBanker()}, false)
		}
	case RSLoopWorkStage:
		if !haveAction {

			ap := room.GetActivePlayer()
			if ap == nil {
				return
			}

			openCard := room.OpenOneCard(ap)
			if openCard == nil {
				room.ChaJiao()
				room.SwitchRoomState(RSSettlement)
				room.CheckDoAction(nil, nil, nil, false)
			} else {
				delayCallId := ap.id + strconv.Itoa(AMo)
				room.StartDelayCallback(delayCallId, room.OpenCardDelay, func(data interface{}) {
					ap.SendActionACK(AMo, openCard, nil, ACSuccess)

					//这个玩家摸牌了，那么以前的杠牌就不可能会导致杠上炮了
					ap.aroundState.checkGangShangPaoCard = nil

					//logger.Error("玩家：%s 清除过水和升值前，过水和升值的情况: 是否仅自摸：%s, 是否过水了：%s, 以前的颗数是：%d, 过的碰杠牌：%s", ap.client.GetName(), ap.aroundState.IsOnlyZiMo(), ap.aroundState.IsGuoShuiHu(), ap.aroundState.huKe, ConvertToWord(ap.aroundState.guoPengGangCard))
					//清除过水，升值等标志
					ap.aroundState.ClearGuoShuiAndShengZhiFlag(ap.HaveBao())

					//检测当前玩家有没的需要执行的动作，没有则通知玩家出牌
					apCanDoActions := room.CheckCanDoActionAndNotifyPlayer(openCard, ap, nil, false)
					if len(apCanDoActions) <= 0 {
						ap.ObtainCard(openCard)

						ap.SendActionNotifyACK(nil, []int32{AChu}, nil)
					} else {
						//缓存发送的牌
						ap.aroundState.moCard = openCard
					}
				}, nil)
			}
		}
	case RSSettlement:
		logger.Info("进入了结算阶段.............")
		room.rs.DoJieSuan()
		//room.DoJieSuan()
	}
}

//把缓存的牌放入手里
func (room *MaJiangRoom) DoMoByCache(clearCache bool) {

	for _, player := range room.players {
		if player == nil {
			continue
		}

		moCard := player.aroundState.moCard

		if moCard != nil {
			if !clearCache && !player.IsHu() {
				player.ObtainCard(moCard)
				PrintCardsS("MaJiangRoom.DoMoByCache--把牌摸到手上后，手上的牌是：", player.cards)
			}
			player.aroundState.moCard = nil

		}
	}
}

//清除补杠标志
func (room *MaJiangRoom) ClearBuGangFlag() {
	for _, player := range room.players {
		if player == nil {
			continue
		}

		player.aroundState.buGangCard = nil
		player.aroundState.buGangCardRemoved = false
	}
}

//检查是否有补杠标志
func (room *MaJiangRoom) HaveBuGangFlag() (have bool, buGangPlayer *MaJiangPlayer) {
	for _, player := range room.players {
		if player == nil {
			continue
		}

		if player.aroundState.buGangCard != nil {
			return true, player
		}
	}

	return
}

//查叫
func (room *MaJiangRoom) ChaJiao() {
	for _, player := range room.players {
		if player == nil || player.IsHu() {
			continue
		}

		player.controller.ChaJiao()
	}
}

//結算
func (room *MaJiangRoom) DoJieSuan() {

	logger.Info("清扫房间准备下一场")

	//结算金币
	jieSuanCoinInfo := room.CalcFinalJieSuanCoin()

	//发送金币扣取通知-通知gameserver扣取金币
	room.SendJieSuanCoinNotifyForAll(jieSuanCoinInfo)

	//发送结算数据
	room.SendJieSuanACKForAll(jieSuanCoinInfo)

	//重置数据
	room.ResetRoom()

	//启动房间停留计时器
	room.StartDelayCallback(RoomStayTimeName, room.RoomStayTime, func(data interface{}) {
		room.rs.ForceAllPlayerLeave()
	}, nil)

	//检查所有玩家的金币信息是否还能继续进行下去
	room.CheckCoinForAll()
}

//检查所有玩家的金币信息是否还能继续进行下去
func (room *MaJiangRoom) CheckCoinForAll() {

	for _, p := range room.players {
		if p == nil {
			continue
		}

		if ok, code := cmn.CheckCoin(room.rtype, p.client); !ok {
			SendEnterRoomErrorACK(p.id, room.rtype, code, false)
			//room.ForcePlayerLeave(p.id)
		}
	}
}

//能否检查下一个动作的执行
func (room *MaJiangRoom) CanCheckDoAction() bool {
	if room.state == RSReady {
		return false
	}

	canDoAction := true
	for _, player := range room.players {
		if player == nil || player.HaveWaitingDoAction() || player.readyDoAction != ANone {
			canDoAction = false
			break
		}
	}

	return canDoAction
}

//获取等待此动作的玩家
func (room *MaJiangRoom) GetWatingActionPlayer(watingAction []int32) (players []*MaJiangPlayer, have bool) {
	if len(watingAction) <= 0 {
		return
	}

	players = make([]*MaJiangPlayer, 0)

	for _, player := range room.players {
		if player == nil {
			continue
		}

		for _, act := range watingAction {
			if Exist(player.watingAction, act) {
				players = append(players, player)
			}
		}

	}

	have = false
	if len(players) > 0 {
		have = true
	}

	return
}

//获取底注
func (room *MaJiangRoom) GetDifen() int32 {
	return room.Difen
}

//是否带归
func (room *MaJiangRoom) GetIsDaiGui() bool {
	return room.IsDaigui
}

//倍数上限
func (room *MaJiangRoom) GetMaxMultiple() int32 {
	return room.MaxMultiple
}

//抽成比率
func (room *MaJiangRoom) GetRakeRate() int32 {
	return room.RakeRate
}

//获取替用的数量
func (room *MaJiangRoom) GetTiYongAmount() int32 {
	return room.TotalHongZhongAmount
}

//获取起胡颗数
func (room *MaJiangRoom) GetQiHuKeAmount() int32 {
	return room.QiHuKeAmount
}

//获取ID
func (room *MaJiangRoom) ID() int32 {
	return room.rtype
}

//获取房间的唯一标示
func (room *MaJiangRoom) UID() int32 {
	return room.uid
}

//获取房间名字
func (room *MaJiangRoom) Name() string {
	return ""
}

//设置房间选择器
func (room *MaJiangRoom) SetSelector(rs cmn.RoomSelector) {
	room.rs = rs
}

//设置发牌器
func (room *MaJiangRoom) SetLicensingController(lc *LicensingController) {
	room.lc = lc
}

//线程接受chan
func (room *MaJiangRoom) GetRcvThreadHandle() *chan cmn.RoomMsgQueue {
	return &room.rcv
}

//线程退出chan
func (room *MaJiangRoom) GetExitThreadHandle() *chan bool {
	return &room.exit
}

//设置当前胡牌的玩家
func (room *MaJiangRoom) SetNextBankerPlayerID(uid string) {
	room.nextBankerPlayerID = uid
}

//获取准备执行此动作的玩家
func (room *MaJiangRoom) GetPlayersForReadyDoAction(doAction int32) (players []*MaJiangPlayer, have bool) {
	if doAction == ANone {
		return
	}

	players = make([]*MaJiangPlayer, 0)

	for _, player := range room.players {
		if player != nil && doAction == player.readyDoAction {
			players = append(players, player)
		}
	}

	have = false
	if len(players) > 0 {
		have = true
	}

	return
}

//重置玩家准备执行的动作为ANone
func (room *MaJiangRoom) ResetReadyDoAction(resetAll bool) {
	for _, player := range room.players {
		if player == nil {
			continue
		}

		if resetAll || player.readyDoAction != AHu {
			player.SwitchReadyDoAction(ANone)
		}
	}
}

//重置玩家等待执行的动作
func (room *MaJiangRoom) ResetWaitingDoAction(resetAll bool) {
	for _, player := range room.players {
		if player == nil {
			continue
		}

		if player.HaveWaitingDoAction() && (resetAll || !Exist(player.watingAction, AHu)) {
			player.SendActionACK(player.watingAction[0], nil, nil, ACAbandon)
			player.SwitchWatingAction([]int32{})
		}

		//只保留胡的动作
		if !resetAll && Exist(player.watingAction, AHu) {
			player.SwitchWatingAction([]int32{AHu})
		}
	}
}

//重置玩家的等待和准备执行动作
func (room *MaJiangRoom) ResetAllAction(resetAll bool) {
	room.ResetWaitingDoAction(resetAll)
	room.ResetReadyDoAction(resetAll)
}

//胡->硬明杠/硬碰->贴鬼明杠/贴鬼碰
func (room *MaJiangRoom) DoReadyActionByOrder() (success, end bool, successPlayer *MaJiangPlayer) {

	//获取玩家
	huPlayers := room.GetHaveSpecificActionPlayers(AHu)
	mingGangPlayers := room.GetHaveSpecificActionPlayers(AMingGang)
	//tieMingGangPlayers := room.GetHaveSpecificActionPlayers(ATieMingGang)
	pengPlayers := room.GetHaveSpecificActionPlayers(APeng)
	//tiePengPlayers := room.GetHaveSpecificActionPlayers(ATiePeng)

	//检查有胡牌动作的玩家
	for _, p := range huPlayers {
		if !p.HaveWaitingDoAction() && p.readyDoAction == AHu {
			success = true
			successPlayer = p
		}
	}

	if success {
		end = len(huPlayers) <= 1
		return
	} else {
		if len(huPlayers) > 0 {
			logger.Info("还有胡的玩家没有执行动作：")
			return
		} else {
			if room.curAroundState.HaveHuPlayer() {
				end = true
				return
			}
		}
	}

	//检查有硬杠动作的玩家
	for _, p := range mingGangPlayers {
		if !p.HaveWaitingDoAction() && p.readyDoAction == AMingGang {
			success = true
			successPlayer = p
		}
	}

	if success {
		end = true
		return
	} else {
		if len(mingGangPlayers) > 0 {
			logger.Info("还有明杠的玩家没有执行动作：")
			return
		}
	}

	//检查有硬碰动作的玩家
	for _, p := range pengPlayers {
		if !p.HaveWaitingDoAction() && p.readyDoAction == APeng {
			success = true
			successPlayer = p
		}
	}

	if success {
		end = true
		return
	} else {
		if len(pengPlayers) > 0 {
			logger.Info("还有硬碰的玩家没有执行动作：")
			return
		}
	}

	//检查有贴鬼明杠碰动作的玩家
	activePlayer := room.GetActivePlayer()
	checkEndPlayer := activePlayer
	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {

		if !player.HaveWaitingDoAction() {
			if player.readyDoAction != ANone {
				success = true
				successPlayer = player
				end = true
				break
			}
		} else {
			break
		}

		player = player.GetXiaJia()
		isInit = false
	}

	return
}

//获取倒计时的玩家
func (room *MaJiangRoom) GetCountdownPlayer(checkAllPlayer bool) *MaJiangPlayer {

	for _, p := range room.players {
		logger.Info("倒计时获取活动玩家  ---------:", CnvtActsToStr(p.watingAction))
	}
	activePlayer := room.GetActivePlayer()

	checkEndPlayer := activePlayer.GetShangJia()
	if checkAllPlayer {
		checkEndPlayer = activePlayer
	}

	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {
		if player.HaveWaitingDoAction() {
			return player
		}

		player = player.GetXiaJia()
		isInit = false
	}
	return nil
}

//启动计时
func (room *MaJiangRoom) StartTimer(second int64) {

	room.timeMgr.StartTimer(MainTimerName, second, room.OnMainTimer, nil)

	logger.Info("开启一个主计时器")
}

//获取剩余倒计时
func (room *MaJiangRoom) GetRemainTime() int32 {
	return room.timeMgr.GetRemainTime(MainTimerName)
}

//是否在计时
func (room *MaJiangRoom) IsTiming() bool {
	return room.timeMgr.IsTiming(MainTimerName)
}

//停止计时器
func (room *MaJiangRoom) StopTimer() {
	room.timeMgr.StopTimer(MainTimerName)
}

//计时器的update
func (room *MaJiangRoom) UpdateTimer(ft time.Duration) {
	//logger.Info("倒计时中..", room.GetPlayerAmount())
	room.handleMsg()        //处理网络消息
	room.timeMgr.Update(ft) //处理内部的延迟消息
}

//到时间的回调
func (room *MaJiangRoom) OnMainTimer(data interface{}) {

	room.StopTimer()
	logger.Info("桌面时间到。==============")

	for _, p := range room.players {
		if p == nil || !p.HaveWaitingDoAction() {
			continue
		}

		//超时就把它设置为自动
		p.PlayerDoAction(ATuoGuan, room.activeCard)
	}
}

//开启一个定时器
func (room *MaJiangRoom) StartDelayCallback(name string, delay int64, call cmn.TimerCallback, data interface{}) {
	room.timeMgr.StartTimer(name, delay, call, data)
}

//停止一个定时器
func (room *MaJiangRoom) StopDelayCallback(name string) {
	room.timeMgr.StopTimer(name)
}

//计算最终输赢金币
func (room *MaJiangRoom) CalcFinalJieSuanCoin() (jieSuanCoinInfo map[string]int32) {
	//计算原始的结算信息
	orginJieSuanCoin := room.CalcOrginJieSuanCoin()

	//修正原始结算金币信息（如果手上的金币不够时，就赢家就不能赢那么多）
	jieSuanTotalCoinInfo := room.CorrectionJieSuanCoin(orginJieSuanCoin)

	//对金币进行抽成
	jieSuanCoinInfo = room.Rake(jieSuanTotalCoinInfo)

	return
}

//计算结算金币
func (room *MaJiangRoom) CalcOrginJieSuanCoin() (jieSuanCoinInfo map[string]int32) {
	jieSuanCoinInfo = make(map[string]int32, 0)

	//初始化玩家金币的输赢信息
	for _, p := range room.players {
		if p == nil {
			continue
		}
		jieSuanCoinInfo[p.id] = 0
	}

	//计算每个玩家的输赢信息
	for _, p := range room.players {
		if p == nil {
			continue
		}

		if p.IsHu() && len(p.beiHuPlayers) > 0 {
			var totalWinCoin int32 = 0
			for _, bhp := range p.beiHuPlayers {
				ke := p.GetKeAmountOfHu(bhp)
				winCoin := ke * room.Difen

				jieSuanCoinInfo[bhp.id] -= winCoin
				totalWinCoin += winCoin
			}

			jieSuanCoinInfo[p.id] += totalWinCoin
		}
	}
	return
}

//修正结算
func (room *MaJiangRoom) CorrectionJieSuanCoin(jieSuanCoinInfo map[string]int32) map[string]int32 {
	//检查输入参数
	if jieSuanCoinInfo == nil {
		logger.Error("jieSuanCoinInfo is nil")
		return nil
	}

	//切分输家也赢家
	losePlayers := make([]*MaJiangPlayer, 0)
	for pid, coin := range jieSuanCoinInfo {
		if coin < 0 {
			losePlayers = append(losePlayers, room.GetPlayerByID(pid))
		}
	}

	//检查输家的金币是否够输，如果不够输需要让赢他的玩家少赢点（平分）
	for _, lp := range losePlayers {
		if loseCoin, exist := jieSuanCoinInfo[lp.id]; exist {
			differCoin := loseCoin - lp.client.GetCoin()
			if differCoin <= 0 {
				continue
			}

			winPlayers := room.GetWinPlayersForSpecificPlayer(lp)
			winPlayersAmount := int32(len(winPlayers))
			if winPlayersAmount <= 0 {
				logger.Error("竟然没有赢家，那他们是怎么输的呢！！")
				continue
			}

			offsetCoin := differCoin / winPlayersAmount
			for _, wp := range winPlayers {
				if _, wExist := jieSuanCoinInfo[wp.id]; wExist {
					jieSuanCoinInfo[wp.id] = int32(math.Max(float64(jieSuanCoinInfo[wp.id]-offsetCoin), 0))
				} else {
					logger.Error("怎么可能没有玩家的结算信息呢")
				}
			}

			jieSuanCoinInfo[lp.id] = -lp.client.GetCoin()

		} else {
			logger.Error("没有这个玩加的结算信息：", lp.id)
		}
	}

	return jieSuanCoinInfo
}

//抽成
func (room *MaJiangRoom) Rake(jieSuanCoinInfo map[string]int32) map[string]int32 {

	for pid, coin := range jieSuanCoinInfo {
		if coin > 0 {
			coin = int32(float32(coin) * (1.0 - float32(room.RakeRate)/100.0))
			jieSuanCoinInfo[pid] = coin
		}
	}

	return jieSuanCoinInfo
}

//获取赢指定的玩家的玩家
func (room *MaJiangRoom) GetWinPlayersForSpecificPlayer(specificPlayer *MaJiangPlayer) (winPlayers []*MaJiangPlayer) {
	winPlayers = make([]*MaJiangPlayer, 0)

	if specificPlayer == nil {
		logger.Error("specificPlayer is nil")
		return
	}

	for _, p := range room.players {
		if p == nil {
			continue
		}

		for _, bhp := range p.beiHuPlayers {
			if bhp.id == specificPlayer.id {
				winPlayers = append(winPlayers, p)
			}
		}
	}

	return
}

func (room *MaJiangRoom) GetAllPlayerIDs() []string {

	result := make([]string, 0)
	for _, p := range room.players {
		if p != nil {
			result = append(result, p.id)
		}
	}

	return result
}

/////////////
//网络消息相关
////////////
//发送进入房间ACK
func (room *MaJiangRoom) SendEnterRoomACK(p cmn.Player) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*MaJiangPlayer)

	logger.Info("发送进入房间：", player.id)

	//打印房间信息
	logger.Info("房间情况：", room.GetPlayerAmount())

	//给自己发送ACK
	msg := &rpc.MJEnterRoomACK{}
	msg.SetRoomId(room.rtype)
	msg.SetLocation(int32(room.GetPlayerIndex(player)))
	msg.SetBReady(player.isReady)
	msg.SetPlayerInfo(player.client)
	msg.SetCode(0)
	if err := conn.SendCommonNotify2S(room.GetAllPlayerIDs(), msg, "MJEnterRoomACK"); err != nil {
		logger.Error("发送进入房间时出错：", err, msg)
		return
	}

	//发送在房间的其他玩家给自己
	for _, rp := range room.players {
		if rp == nil || rp.id == player.id {
			continue
		}

		msg := &rpc.MJEnterRoomACK{}
		msg.SetRoomId(room.rtype)
		msg.SetLocation(int32(room.GetPlayerIndex(rp)))
		msg.SetBReady(rp.isReady)
		msg.SetPlayerInfo(rp.client)
		msg.SetCode(0)

		if err := conn.SendCommonNotify2S([]string{player.id}, msg, "MJEnterRoomACK"); err != nil {
			logger.Error("发送进入房间时出错：", err, msg)
			continue
		}
	}
}

//向客户端发送玩家离开房间的消息
func (room *MaJiangRoom) SendLeaveRoomACK(p cmn.Player, isChangeDesk bool) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*MaJiangPlayer)

	//给房间里的每个发送离开房间
	msg := &rpc.MJLeaveRoomACK{}
	msg.SetPlayerID(player.id)
	msg.SetIsChangeDesk(isChangeDesk)
	if err := conn.SendCommonNotify2S(room.GetAllPlayerIDs(), msg, "MJLeaveRoomACK"); err != nil {
		logger.Error("发送离开房间出错：", err, msg)
		return
	}
}

//发送进入房间错误ACK
func SendEnterRoomErrorACK(uid string, roomType, code int32, isNormalReqEnterRoom bool) {
	rmMsg := &rpc.MJEnterRoomACK{}
	rmMsg.SetRoomId(roomType)
	rmMsg.SetCode(code)
	rmMsg.SetIsNormalReqEnterRoom(isNormalReqEnterRoom)
	//logger.Info("进入房间错消息：", rmMsg)
	if err := conn.SendCommonNotify2S([]string{uid}, rmMsg, "MJEnterRoomACK"); err != nil {
		logger.Error("发送进入房间出错：", err)
	}
}

//发送聊天通知
func (room *MaJiangRoom) SendCommonMsg2Others(msg *rpc.FightRoomChatNotify) {
	logger.Info("SendCommonMsg2Others has been called")

	//给房间里的每个发送离开房间
	conn.SendCommonNotify2S(room.GetAllPlayerIDs(), msg, "FightRoomChatNotify")
}

//发送游戏开始
func (room *MaJiangRoom) sendGameStartForAll() {
	for _, p := range room.players {
		if p == nil {
			continue
		}
		p.SendGameStartACK(false)
	}

	logger.Info("——————————房间状态：", roomTypeName[room.state])
}

//发送桌面通知信息
func (room *MaJiangRoom) sendCountdownNotifyACK() {

	cp := room.GetCountdownPlayer(true)
	if cp != nil {
		for _, p := range room.players {
			if p == nil {
				continue
			}

			if p.HaveWaitingDoAction() {
				p.sendCountdownNotifyACK(p)
			} else {
				p.sendCountdownNotifyACK(cp)
			}
		}
	} else {
		logger.Error("没有活动玩家")
	}

}

//通知结算金币
func (room *MaJiangRoom) SendJieSuanCoinNotifyForAll(jieSuanCoin map[string]int32) {
	if jieSuanCoin == nil {
		logger.Error("结算玩家的金币信息有误")
		return
	}

	for _, p := range room.players {
		if p == nil {
			continue
		}

		if coin, exist := jieSuanCoin[p.id]; exist {
			p.SendJieSuanCoinNotify(coin)
		}
	}
}

//发送游戏结算
func (room *MaJiangRoom) SendJieSuanACKForAll(jieSuanCoin map[string]int32) {

	logger.Info("向所有发送结算信息")

	addiData := &rpc.JieSuanAdditionData{}
	addiData.SetSysType(cmn.PiPeiFang)
	for _, p := range room.players {
		if p == nil {
			continue
		}

		//发送数据
		p.SendJieSuanACK(jieSuanCoin, addiData)
	}

	return
}

//接收客户端发来的消息
func (room *MaJiangRoom) OnPlayerDoAction(msg *rpc.ActionREQ) {
	if msg == nil {
		logger.Error("MaJiangRoom.OnPlayerDoAction:客户端发送来的数据为空！")
		return
	}

	logger.Info("receive client msg:(Action:%s, playerID:%s)", msg.GetAction(), msg.GetPlayerID())

	player := room.GetPlayerByID(msg.GetPlayerID())
	if player == nil {
		logger.Error("MaJiangRoom.OnPlayerDoAction:此房间没有这个人%s", msg.GetPlayerID())
		return
	}

	//玩家吃牌
	action := int32(msg.GetAction())

	curTime := time.Now()

	if action == AChu || action == AAnGang || action == ABuGang {
		if pai := msg.GetCardArgs(); pai != nil {
			cPai := convertCardToMaJiangCard(pai)
			logger.Info("客户端发送要出的牌上来:", ConvertToWord(cPai))

			if action != AChu && cPai.IsHongZhong() {
				logger.Error("客户端发送上来的牌:%s 不应该是红中", ConvertToWord(cPai))
			}
			cPai.owner = player
			player.PlayerDoAction(action, cPai)
		} else {
			logger.Error("客户端没有发送要出的牌上来")
		}
	} else {
		logger.Info("客服端发送来的的动作：", actionName[action])
		player.PlayerDoAction(action, room.activeCard)
	}

	logger.Info("服务器执行一个动作需要的时间:", time.Now().Sub(curTime))
}

//解散房间的接口
func (room *MaJiangRoom) OnJieSanRoom(uid string, msg *rpc.JieSanRoomREQ) {

}

func (room *MaJiangRoom) OnTieGuiREQ(msg *rpc.MJTieGuiREQ) {
	if msg == nil {
		logger.Error("MaJiangRoom.OnTieGuiREQ:客户端发送来的数据为空！")
		return
	}

	p := room.GetPlayerByID(msg.GetPlayerID())
	if p == nil {
		logger.Error("没有获取到玩家：", msg.GetPlayerID())
		return
	}

	p.IsOpenHongZhongCheck = msg.GetBTieGui()

	logger.Info("是否开启贴鬼检查：", p.IsOpenHongZhongCheck)
}
