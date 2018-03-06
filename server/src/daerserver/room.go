package daerserver

import (
	// gp "code.google.com/p/goprotobuf/proto"
	conn "centerclient"
	cmn "common"
	"logger"
	//"math"
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

type DaerRoom struct {
	uid               int32                            //房间的唯一标示
	rtype             int32                            //房间类型(即配置表的ID)
	state             int32                            //房间的状态
	players           [RoomMaxPlayerAmount]*DaerPlayer //房间的所有玩家
	activePlayerIndex int                              //当前的玩家
	cards             []*DaerCard                      //所有的牌
	ownCards          []*DaerCard                      //桌子上未开的牌
	activeCard        *DaerCard                        //当前活动的卡牌
	gamingPlayerIDs   [RoomMaxPlayerAmount]string      //进行游戏时的玩家ID列表，主要用于定庄
	huPaiPlayerID     string                           //胡牌的玩家ID -也用于定庄
	Difen             int32                            //底注
	IsDaigui          bool                             //是否带归
	MaxMultiple       int32                            //倍数上限
	RakeRate          int32                            //抽成比率
	TimerInterval     int64                            //给玩家考虑的时间
	GameStartDelay    int64                            //延迟开始游戏
	DoActionDelay     int64                            //执行动作的延迟
	OpenCardDelay     int64                            //开一张牌的延迟
	RoomStayTime      int64                            //房间等待的最大时间
	GamingMaxTime     int64                            //游戏最大进行时间
	rs                cmn.RoomSelector                 //房间选择器（分支器）

	timeMgr  *cmn.TimerMgr      //定时器管理器
	t        *timer.Timer       //定时器
	msgQueue []cmn.RoomMsgQueue //消息队列
	rcv      chan cmn.RoomMsgQueue
	exit     chan bool
	ql       sync.Mutex
}

//创建函数
// type CreateMsgFun func() interface{}

// var msgReflect map[string]CreateMsgFun

// func init() {
// 	msgReflect = make(map[string]CreateMsgFun)

// 	msgReflect["ActionREQ"] = func() interface{} { return &rpc.ActionREQ{} }

// }

//新建一个大贰 房间
func NewDaerRoom(uid, rtype int32) *DaerRoom {
	r := new(DaerRoom)
	//r.rtype = rtype
	//r.uid = uid
	//r.SetSelector(r)

	r.Init(uid, rtype)

	// r.SetSelector(r)
	// r.msgQueue = []cmn.RoomMsgQueue{}
	// r.rcv = make(chan cmn.RoomMsgQueue, 10)

	// timerInterval := time.Millisecond * 50
	// r.t = timer.NewTimer(timerInterval)
	// r.t.Start(
	// 	func() {
	// 		r.UpdateTimer(timerInterval)
	// 	},
	// )

	// go r.process()

	return r
}

//初始化房间
func (room *DaerRoom) Init(uid, rtype int32) {
	room.uid = uid
	room.rtype = rtype
	room.state = RSReady
	room.SetSelector(room)
	room.msgQueue = []cmn.RoomMsgQueue{}
	room.rcv = make(chan cmn.RoomMsgQueue, 10)
	room.timeMgr = cmn.NewTimerMgr()
	room.timeMgr.SetUpdateInterval(time.Millisecond * 1000)

	room.InitByConfig(rtype)

	room.cards = []*DaerCard{
		//小牌 1-10
		&DaerCard{id: 1, value: 1, big: false, flag: cmn.CBack},
		&DaerCard{id: 2, value: 2, big: false, flag: cmn.CBack},
		&DaerCard{id: 3, value: 3, big: false, flag: cmn.CBack},
		&DaerCard{id: 4, value: 4, big: false, flag: cmn.CBack},
		&DaerCard{id: 5, value: 5, big: false, flag: cmn.CBack},
		&DaerCard{id: 6, value: 6, big: false, flag: cmn.CBack},
		&DaerCard{id: 7, value: 7, big: false, flag: cmn.CBack},
		&DaerCard{id: 8, value: 8, big: false, flag: cmn.CBack},
		&DaerCard{id: 9, value: 9, big: false, flag: cmn.CBack},
		&DaerCard{id: 10, value: 10, big: false, flag: cmn.CBack},
		//小牌 1-10
		&DaerCard{id: 11, value: 1, big: false, flag: cmn.CBack},
		&DaerCard{id: 12, value: 2, big: false, flag: cmn.CBack},
		&DaerCard{id: 13, value: 3, big: false, flag: cmn.CBack},
		&DaerCard{id: 14, value: 4, big: false, flag: cmn.CBack},
		&DaerCard{id: 15, value: 5, big: false, flag: cmn.CBack},
		&DaerCard{id: 16, value: 6, big: false, flag: cmn.CBack},
		&DaerCard{id: 17, value: 7, big: false, flag: cmn.CBack},
		&DaerCard{id: 18, value: 8, big: false, flag: cmn.CBack},
		&DaerCard{id: 19, value: 9, big: false, flag: cmn.CBack},
		&DaerCard{id: 20, value: 10, big: false, flag: cmn.CBack},
		//小牌 1-10
		&DaerCard{id: 21, value: 1, big: false, flag: cmn.CBack},
		&DaerCard{id: 22, value: 2, big: false, flag: cmn.CBack},
		&DaerCard{id: 23, value: 3, big: false, flag: cmn.CBack},
		&DaerCard{id: 24, value: 4, big: false, flag: cmn.CBack},
		&DaerCard{id: 25, value: 5, big: false, flag: cmn.CBack},
		&DaerCard{id: 26, value: 6, big: false, flag: cmn.CBack},
		&DaerCard{id: 27, value: 7, big: false, flag: cmn.CBack},
		&DaerCard{id: 28, value: 8, big: false, flag: cmn.CBack},
		&DaerCard{id: 29, value: 9, big: false, flag: cmn.CBack},
		&DaerCard{id: 30, value: 10, big: false, flag: cmn.CBack},
		//小牌 1-10
		&DaerCard{id: 31, value: 1, big: false, flag: cmn.CBack},
		&DaerCard{id: 32, value: 2, big: false, flag: cmn.CBack},
		&DaerCard{id: 33, value: 3, big: false, flag: cmn.CBack},
		&DaerCard{id: 34, value: 4, big: false, flag: cmn.CBack},
		&DaerCard{id: 35, value: 5, big: false, flag: cmn.CBack},
		&DaerCard{id: 36, value: 6, big: false, flag: cmn.CBack},
		&DaerCard{id: 37, value: 7, big: false, flag: cmn.CBack},
		&DaerCard{id: 38, value: 8, big: false, flag: cmn.CBack},
		&DaerCard{id: 39, value: 9, big: false, flag: cmn.CBack},
		&DaerCard{id: 40, value: 10, big: false, flag: cmn.CBack},

		//大牌 1-10
		&DaerCard{id: 41, value: 1, big: true, flag: cmn.CBack},
		&DaerCard{id: 42, value: 2, big: true, flag: cmn.CBack},
		&DaerCard{id: 43, value: 3, big: true, flag: cmn.CBack},
		&DaerCard{id: 44, value: 4, big: true, flag: cmn.CBack},
		&DaerCard{id: 45, value: 5, big: true, flag: cmn.CBack},
		&DaerCard{id: 46, value: 6, big: true, flag: cmn.CBack},
		&DaerCard{id: 47, value: 7, big: true, flag: cmn.CBack},
		&DaerCard{id: 48, value: 8, big: true, flag: cmn.CBack},
		&DaerCard{id: 49, value: 9, big: true, flag: cmn.CBack},
		&DaerCard{id: 50, value: 10, big: true, flag: cmn.CBack},

		//大牌 1-10
		&DaerCard{id: 51, value: 1, big: true, flag: cmn.CBack},
		&DaerCard{id: 52, value: 2, big: true, flag: cmn.CBack},
		&DaerCard{id: 53, value: 3, big: true, flag: cmn.CBack},
		&DaerCard{id: 54, value: 4, big: true, flag: cmn.CBack},
		&DaerCard{id: 55, value: 5, big: true, flag: cmn.CBack},
		&DaerCard{id: 56, value: 6, big: true, flag: cmn.CBack},
		&DaerCard{id: 57, value: 7, big: true, flag: cmn.CBack},
		&DaerCard{id: 58, value: 8, big: true, flag: cmn.CBack},
		&DaerCard{id: 59, value: 9, big: true, flag: cmn.CBack},
		&DaerCard{id: 60, value: 10, big: true, flag: cmn.CBack},

		//大牌 1-10
		&DaerCard{id: 61, value: 1, big: true, flag: cmn.CBack},
		&DaerCard{id: 62, value: 2, big: true, flag: cmn.CBack},
		&DaerCard{id: 63, value: 3, big: true, flag: cmn.CBack},
		&DaerCard{id: 64, value: 4, big: true, flag: cmn.CBack},
		&DaerCard{id: 65, value: 5, big: true, flag: cmn.CBack},
		&DaerCard{id: 66, value: 6, big: true, flag: cmn.CBack},
		&DaerCard{id: 67, value: 7, big: true, flag: cmn.CBack},
		&DaerCard{id: 68, value: 8, big: true, flag: cmn.CBack},
		&DaerCard{id: 69, value: 9, big: true, flag: cmn.CBack},
		&DaerCard{id: 70, value: 10, big: true, flag: cmn.CBack},

		//大牌 1-10
		&DaerCard{id: 71, value: 1, big: true, flag: cmn.CBack},
		&DaerCard{id: 72, value: 2, big: true, flag: cmn.CBack},
		&DaerCard{id: 73, value: 3, big: true, flag: cmn.CBack},
		&DaerCard{id: 74, value: 4, big: true, flag: cmn.CBack},
		&DaerCard{id: 75, value: 5, big: true, flag: cmn.CBack},
		&DaerCard{id: 76, value: 6, big: true, flag: cmn.CBack},
		&DaerCard{id: 77, value: 7, big: true, flag: cmn.CBack},
		&DaerCard{id: 78, value: 8, big: true, flag: cmn.CBack},
		&DaerCard{id: 79, value: 9, big: true, flag: cmn.CBack},
		&DaerCard{id: 80, value: 10, big: true, flag: cmn.CBack}}

	timerInterval := time.Millisecond * 50
	room.t = timer.NewTimer(timerInterval)
	room.t.Start(
		func() {
			room.UpdateTimer(timerInterval)
		},
	)

	go room.process()
}

//初始化通过配置表
func (r *DaerRoom) InitByConfig(rtype int32) {
	//获取底注
	r.Difen = 100
	cfg := cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		r.Difen = cfg.Difen
	} else {
		logger.Error("读取房间配置表出错ID：%s", r.rtype)
	}

	//获取是否带归
	r.IsDaigui = true
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		if cfg.Type == cmn.RTDaerHight {
			r.IsDaigui = true
		} else {
			r.IsDaigui = false
		}
	} else {
		logger.Error("读取房间配置表出错ID：%s", rtype)
	}

	//获取最高倍数
	r.MaxMultiple = 10
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		r.MaxMultiple = cfg.MaxMultiple
	} else {
		logger.Error("读取房间配置表出错ID：%s", r.rtype)
	}

	//获取抽成比利
	r.RakeRate = 10
	cfg = cmn.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg != nil {
		r.RakeRate = cfg.RakeRate
	} else {
		logger.Error("读取房间配置表出错ID：%s", r.rtype)
	}

	//给玩家的考虑时间
	r.TimerInterval = 180
	gcfg := cmn.GetDaerGlobalConfig("1")
	if gcfg != nil {
		r.TimerInterval = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//延迟开始游戏
	r.GameStartDelay = 6
	gcfg = cmn.GetDaerGlobalConfig("2")
	if gcfg != nil {
		r.GameStartDelay = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//执行动作的延迟
	r.DoActionDelay = 1
	gcfg = cmn.GetDaerGlobalConfig("3")
	if gcfg != nil {
		r.DoActionDelay = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//开一张牌的延迟
	r.OpenCardDelay = 2
	gcfg = cmn.GetDaerGlobalConfig("4")
	if gcfg != nil {
		r.OpenCardDelay = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//房间等待的最大时间
	r.RoomStayTime = 120
	gcfg = cmn.GetDaerGlobalConfig("5")
	if gcfg != nil {
		r.RoomStayTime = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//游戏最大进行时间
	r.GamingMaxTime = 1200
	gcfg = cmn.GetDaerGlobalConfig("40")
	if gcfg != nil {
		r.GamingMaxTime = int64(gcfg.IntValue)
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

}

func (room *DaerRoom) handleMsg() {
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
			} else if e.Func == "Enter" {
				msg := e.Msg
				p, ok := msg.(*DaerPlayer)
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

				room.rs.Leave(uid, isChangeDesk)
				//room.Leave(uid, isChangeDesk)

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

func (room *DaerRoom) process() {
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

//进入房间
func (room *DaerRoom) Enter(p cmn.Player) {
	//检查输入参数
	if p == nil {
		logger.Error("player is nil.")
		return
	}

	player := p.(*DaerPlayer)
	if player == nil {
		logger.Error("接口转换失败")
		return
	}

	//检查能否进入房间
	if room.IsFull() {
		logger.Error("DaerRoom.Enter: room is full")
		return
	}

	//修改玩家的房间相关的信息
	player.room = room
	for i := 0; i < RoomMaxPlayerAmount; i++ {
		if room.players[i] == nil {
			room.players[i] = player
			room.SendEnterRoomACK(player)
			logger.Info("DaerRoom.Enter: player:%s enter room(%s):", player.id, i)
			break
		}
	}

	//开启一个房间停留计时器
	room.StartDelayCallback(RoomStayTimeName, room.RoomStayTime, func(data interface{}) {
		room.rs.ForceAllPlayerLeave()
	}, nil)
}

//重新进入房间
func (room *DaerRoom) ReEnter(playerID string, playerInfo *rpc.PlayerBaseInfo) {
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
		logger.Error("DaerRoom:player not in the room")
	}
}

//离开房间
func (room *DaerRoom) Leave(uid string, isChangeDesk bool) bool {
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
	room.huPaiPlayerID = ""

	for i := 0; i < RoomMaxPlayerAmount; i++ {
		tempPlayer := room.players[i]
		if tempPlayer != nil && uid == tempPlayer.id {
			room.SendLeaveRoomACK(tempPlayer, isChangeDesk)
			tempPlayer.room = nil
			room.players[i] = nil
			tempPlayer.Reset()
			logger.Info("DaerRoom.Enter: player:%s Leave room:", tempPlayer.id)
			break
		}
	}

	//开启一个房间停留计时器
	room.StartDelayCallback(RoomStayTimeName, room.RoomStayTime, func(data interface{}) {
		room.rs.ForceAllPlayerLeave()
	}, nil)

	//删除roommgr中的引用关系
	daerRoomMgr.DeleteLeavePlayerInfo(room.rtype, uid)

	//检查是否是换桌
	if isChangeDesk {
		daerRoomMgr.EnterGame(room.rtype, leavePlayer.client, isChangeDesk)
	}

	if room.IsEmpty() {
		logger.Error("停止房间的Timer和消息接受线程")
		//room.ResetRoom()
		//room.t.Stop()
		room.exit <- true

	}

	return true
}

//指定时间玩家都还没有开始游戏，那么将所有人提出房间
func (room *DaerRoom) ForceAllPlayerLeave() {
	if daerRoomMgr == nil {
		return
	}

	for _, p := range room.players {
		if p != nil {
			daerRoomMgr.LeaveGame(p.id, false)
		}
	}
}

//强制玩家离开房间
func (room *DaerRoom) ForcePlayerLeave(uid string) {
	if daerRoomMgr == nil {
		return
	}

	for _, p := range room.players {
		if p != nil && p.id == uid {
			daerRoomMgr.LeaveGame(uid, false)
		}
	}
}

//开始游戏
func (room *DaerRoom) StartGame() {
	logger.Info("=======开始游戏-=======")
	//缓存游戏开始时的玩家
	room.CachePlayerID()

	//开始洗牌
	cards := room.Shuffle()
	if cards == nil {
		logger.Error("洗牌失败！")
		return
	}

	//开始发牌
	room.Licensing(cards)

	//定庄
	room.DecideBanker()

	//更新计算胡控制器
	room.UpdateHuControllerForAll()

	//向所有玩家发送开始游戏消息
	room.sendGameStartForAll()

	//切换到报牌阶段
	room.SwitchRoomState(RSBaoStage)

	//延迟触发报牌阶段的检测
	room.StartDelayCallback(StartGameName, room.GameStartDelay, func(data interface{}) {
		logger.Info("进入摆流程....，并启动第一个检查....")
		room.CheckDoAction(nil)
	}, nil)

	//开启游戏后执行的一些操作
	room.rs.OnStartGameAfter()

}

//开启游戏后
func (room *DaerRoom) OnStartGameAfter() {
	//停止房间停留计时器
	room.StopDelayCallback(RoomStayTimeName)

	//开启最大游戏进行时间的延迟
	room.StartDelayCallback(GamingMaxTimeName, room.GamingMaxTime, func(data interface{}) {
		room.SwitchRoomState(RSReady)
		room.rs.ForceAllPlayerLeave()
	}, nil)

	//test
	logger.Info("================开始游戏================")
	PrintRoom(room)
}

//洗牌
func (room *DaerRoom) Shuffle() (cards []*DaerCard) {

	//重置房间的状态
	room.ResetRoom()

	//打乱牌
	cards = make([]*DaerCard, CardTotalAmount)
	copy(cards, room.cards)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	curRound := 1
	i := CardTotalAmount
	for i > 0 {
		index := r.Intn(i)

		temp := cards[index]
		cards[index] = cards[i-1]
		cards[i-1] = temp
		i--

		if i <= 0 && curRound < 2 {
			// logger.Error("第%d轮洗牌后：", curRound)
			// PrintCardsS("洗牌后：", cards)
			i = CardTotalAmount
			curRound++
		}
	}

	return
}

//发牌
func (room *DaerRoom) Licensing(cards []*DaerCard) {

	//是否启动了特殊发牌规则
	if SpecificLicensingType != LCNone {
		logger.Error("启用了特殊发牌======")
		tempCards := Licensing(SpecificLicensingType, room)
		if tempCards == nil {
			logger.Error("特殊发牌出问题了！！！！")
		} else {
			cards = tempCards
		}
	}

	PrintCards(cards)
	//检测输入参数的合法性
	if cards == nil || len(cards) != CardTotalAmount {
		logger.Error("cards is empty or amount is error.")
		return
	}

	//检测玩家数量是否正确
	curPalyerAmount := room.GetPlayerAmount()
	if curPalyerAmount <= 0 || curPalyerAmount > RoomMaxPlayerAmount {
		logger.Error("玩家数量不对！")
		return
	}

	room.ownCards = cards[3*20:]

	for i, player := range room.players {
		if player == nil {
			continue
		}

		playerCards := make([]*DaerCard, FirstCardsAmount)
		copy(playerCards, cards[i*20:(i+1)*20])
		player.Compose(playerCards)
	}
}

//定庄
func (room *DaerRoom) DecideBanker() {

	bankerIndex := 0

	if room.IsPlayerHaveChange() || room.huPaiPlayerID == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		bankerIndex = r.Intn(RoomMaxPlayerAmount)
	} else {
		p := room.GetPlayerByID(room.huPaiPlayerID)
		if p != nil {
			bankerIndex = room.GetPlayerIndex(p)
		} else {
			logger.Error("在定庄时，没有获取到胡牌玩家：%s 的玩家信息", room.huPaiPlayerID)
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

	room.huPaiPlayerID = ap.id
}

//更新胡控制器
func (room *DaerRoom) UpdateHuControllerForAll() {
	for _, p := range room.players {
		if p == nil {
			continue
		}
		p.controller.huController.UpdateData(p.cards)
	}
}

//缓存房间玩家ID--用于定庄检查
func (room *DaerRoom) CachePlayerID() {
	if !room.IsFull() {
		logger.Error("开始游戏的时候，房间人数没有满")
		return
	}

	for i := 0; i < RoomMaxPlayerAmount; i++ {
		room.gamingPlayerIDs[i] = room.players[i].id
	}
}

//房间的玩家是否变化
func (room *DaerRoom) IsPlayerHaveChange() (result bool) {

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

//重置房间
func (room *DaerRoom) ResetRoom() {
	room.state = RSReady
	//room.passCards = make([]*DaerCard, 0)
	room.activeCard = nil
	room.activePlayerIndex = 0
	room.ownCards = make([]*DaerCard, 0)

	//room.timeMgr.StopTimer(MainTimerName)
	//room.timeMgr.StopAllTimer()
	room.timeMgr.Clear()

	//初始化牌的状态
	for _, v := range room.cards {
		v.flag = cmn.CBack
		v.owner = nil
	}

	//重置玩家状态
	room.ResetForAllPlayer()
}

//增加一张桌面过牌(没有胡，摆，找，碰，吃的牌需要添加到过的列表中)
//func (room *DaerRoom) AddPassCard(card *DaerCard) {
//	if card == nil {
//		logger.Error("card is nil.")
//		return
//	}

//	card.flag = cmn.CLock | cmn.CPositive

//	room.passCards = append(room.passCards, card)
//}

//检测添加并添加到过牌(桌面或玩家)列表中去
func (room *DaerRoom) CheckAndAddShowCard(card *DaerCard) {
	if card == nil {
		return
	}

	isDeskOpen := card.owner == nil
	if isDeskOpen {
		ap := room.GetActivePlayer()
		if ap != nil {
			ap.AddPassCard(card)
		}
	} else {
		card.owner.AddPassCard(card)
	}
}

//重置所有玩家状态
func (room *DaerRoom) ResetForAllPlayer() {
	for _, player := range room.players {
		if player == nil {
			continue
		}

		player.Reset()
	}
}

//改变活动玩家
func (room *DaerRoom) ChangeActivePlayerByIndex(index int) {
	if index < 0 || index > RoomMaxPlayerAmount {
		logger.Error("index out of")
		return
	}

	room.activePlayerIndex = index
}

//改变活动玩家
func (room *DaerRoom) ChangeActivePlayerToNext() {
	room.activePlayerIndex = (room.activePlayerIndex + 1) % RoomMaxPlayerAmount
}

//改变活动玩家
func (room *DaerRoom) ChangeActivePlayerTo(p cmn.Player) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*DaerPlayer)
	if player == nil {
		logger.Error("接口转换失败")
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
func (room *DaerRoom) GetPlayerIndex(player *DaerPlayer) int {
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

//检测一个玩家时候是活动玩家
func (room *DaerRoom) IsActivePlayer(player *DaerPlayer) bool {
	if player == nil {
		return false
	}

	activePalyer := room.players[room.activePlayerIndex]
	return activePalyer != nil && activePalyer.id == player.id
}

//检查一个玩家是否在房间
func (room *DaerRoom) IsInRoom(uid string) bool {
	for _, v := range room.players {
		if v != nil && uid == v.id {
			return true
		}
	}

	return false
}

//获取活动玩家
func (room *DaerRoom) GetActivePlayer() *DaerPlayer {
	return room.players[room.activePlayerIndex]
}

//获取所有玩家
func (room *DaerRoom) GetAllPlayer() *[RoomMaxPlayerAmount]*DaerPlayer {
	return &room.players
}

//获取庄家
func (room *DaerRoom) GetBanker() *DaerPlayer {
	for _, player := range room.players {
		if player != nil && player.ptype == cmn.PTBanker {
			return player
		}
	}
	return nil
}

//获取报牌的人数
func (room *DaerRoom) GetAmountOfBaoPai() int32 {
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
func (room *DaerRoom) GetPlayerByID(id string) *DaerPlayer {
	for _, player := range room.players {
		if player != nil && player.id == id {
			return player
		}
	}

	return nil
}

//获取房间人数
func (room *DaerRoom) GetPlayerAmount() int32 {
	var result int32 = 0
	for _, v := range room.players {
		if v != nil {
			result += 1
		}
	}
	return result
}

//玩家数量上限
func (room *DaerRoom) MaxPlayerAmount() int32 {
	return RoomMaxPlayerAmount
}

//房间是否满员
func (room *DaerRoom) IsFull() bool {
	return room.GetPlayerAmount() >= RoomMaxPlayerAmount
}

//房间是否为空
func (room *DaerRoom) IsEmpty() bool {
	return room.GetPlayerAmount() <= 0
}

//获取有游戏类型
func (room *DaerRoom) GetGameType() int32 {
	return cmn.DaerGame
}

//是否正在游戏
func (room *DaerRoom) IsGaming() bool {
	return room.state != RSReady && room.state != RSSettlement
}

//所有玩家都准备好了吗
func (room *DaerRoom) IsReadyForAll() bool {
	result := true
	for _, v := range room.players {
		if v != nil && !v.isReady {
			result = false
			break
		}
	}

	return result
}

//能够开始游戏了吗
func (room *DaerRoom) CanStartGame() bool {
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

//开一张牌
func (room *DaerRoom) OpenOneCard() *DaerCard {
	remainCards := room.ownCards
	if remainCards == nil || len(remainCards) <= 0 {
		return nil
	}

	openCard := remainCards[len(remainCards)-1]
	room.ownCards = room.ownCards[:len(remainCards)-1]
	room.activeCard = openCard
	return openCard
}

//切换房间状态
func (room *DaerRoom) SwitchRoomState(state int32) {
	room.state = state
}

//是否存在
func IsExistCard(cards []*DaerCard, card *DaerCard) bool {
	if cards == nil || card == nil {
		return false
	}

	for _, c := range cards {
		if c.IsEqual(card) {
			return true
		}
	}

	return false
}

//检查所有玩家能执行的动作,并下发给玩家
func (room *DaerRoom) CheckCanDoAction(card *DaerCard) int32 {
	//定义默认的动作
	var doAction int32 = ANone

	switch room.state {
	case RSBaoStage:
		for _, player := range room.players {
			if player == nil {
				continue
			}

			if canBao, _ := player.controller.CheckBao(); canBao {
				doAction = ABao
				player.SendActionNotifyACK(doAction, nil, nil, nil)
				logger.Info("CheckCanDoAction:玩家%s,检查到:%s", player.id, actionName[doAction])
			}
		}
	case RSNotBankerBaiStage:
		for _, p := range room.players {
			//非庄家和报牌的玩家不用检查摆
			if p == nil || p.ptype == cmn.PTBanker || p.delayDoAction == ABao {
				continue
			}

			//logger.Error("在非庄家摆阶段检测到摆，延迟执行的动作为：", actionName[p.delayDoAction])

			if action, _ := p.controller.CheckBai(nil); action != ANone {
				doAction = action
				p.SendActionNotifyACK(doAction, nil, nil, nil)
				logger.Info("CheckCanDoAction:玩家%s,检查到:%s", p.id, actionName[doAction])
			}
		}
	case RSBankerJinPaiStage:
		//检测摆牌（三拢，四坎，黑）
		player := room.GetBanker()
		if player == nil {
			logger.Error("竟然没有庄家")
			break
		}

		//检测胡
		if canHu, _ := player.controller.CheckHuSpecific(card); canHu {
			doAction = AHu
			player.SendActionNotifyACK(doAction, card, nil, nil)
			logger.Info("CheckCanDoAction:玩家%s,检查到:%s", player.id, actionName[doAction])
		}
	case RSBankerBaiStage:
		//检查庄家出完第一张后还能不能摆牌
		banker := room.GetBanker()
		if banker == nil {
			logger.Error("竟然没有庄家")
			break
		}

		//已经报了，就不用检查摆了
		if banker.HaveBao() {
			break
		}

		//检查庄家摆牌
		if action, _ := banker.controller.CheckBai(nil); action != ANone {
			doAction = action
			banker.SendActionNotifyACK(doAction, nil, nil, nil)
			logger.Info("CheckCanDoAction:玩家%s,检查到:%s", banker.id, actionName[doAction])
		}
	case RSBankerBaoStage:
		//检查庄家出完第一张后还能不能报
		banker := room.GetBanker()
		if banker == nil {
			logger.Error("竟然没有庄家")
			break
		}

		//已经报了，这个阶段就不用报了
		if banker.HaveBao() {
			break
		}

		//检查庄家报牌
		if canBao, _ := banker.controller.CheckBao(); canBao {
			doAction = ABao
			banker.SendActionNotifyACK(doAction, nil, nil, nil)
			logger.Info("CheckCanDoAction:玩家%s,检查到:%s", banker.id, actionName[doAction])
		}
	case RSBankerChuPaiAfterStage:
		fallthrough

	case RSLoopWorkStage:
		if card == nil {
			break
		}

		//检查胡，多个玩家可以同时检测到胡
		for _, player := range room.players {
			if player == nil || (card.owner != nil && card.owner.id == player.id) ||
				(room.state == RSBankerChuPaiAfterStage && player.ptype == cmn.PTBanker) {
				continue
			}

			logger.Info("玩家%s  检查胡牌,牌:%s", player.GetPlayerBasicInfo().GetName(), card)
			if canHu, _ := player.controller.CheckHuSpecific(card); canHu {
				doAction = AHu
				player.SendActionNotifyACK(doAction, card, nil, nil)
				logger.Info("CheckCanDoAction:玩家%s,检查到:%s", player.id, actionName[doAction])
			}
		}

		if doAction != ANone {
			break
		}

		//检测招和碰
		doAction = room.CheckCanDoZhaoAndPengChi(card)

	}

	if doAction != ANone {
		room.sendCountdownNotifyACK()
	}

	return doAction
}

//检查能否触发招和碰并发送通知
func (room *DaerRoom) CheckCanDoZhaoAndPengChi(card *DaerCard) (doAction int32) {

	doAction = ANone

	if !room.CanCheckDoAction() {
		return
	}

	//检测招和碰
	for _, player := range room.players {
		if player == nil ||
			(card.owner != nil && card.owner.id == player.id) ||
			(room.state == RSBankerChuPaiAfterStage && player.ptype == cmn.PTBanker) {
			continue
		}

		//检查招
		canZhao, isAgainZhao := player.controller.CheckZhao(card)
		if canZhao {
			action := AZhao
			if isAgainZhao {
				action = AZhongZhao
			}

			doAction = int32(action)
			player.SendActionNotifyACK(doAction, card, nil, nil)
			logger.Info("CheckCanDoAction:玩家%s,检查到:%s", player.id, actionName[doAction])
			break
		}
	}

	if doAction != ANone {
		return doAction
	}

	for _, player := range room.players {
		if player == nil ||
			(card.owner != nil && card.owner.id == player.id) ||
			(room.state == RSBankerChuPaiAfterStage && player.ptype == cmn.PTBanker) ||
			player.HaveBao() {
			continue
		}

		//检查碰
		if canPeng := player.controller.CheckPeng(card); canPeng {
			doAction = APeng
			player.SendActionNotifyACK(APeng, card, nil, nil)
			logger.Info("CheckCanDoAction:玩家%s,检查到:%s", player.id, actionName[APeng])
			break
		}
	}

	if doAction != ANone {
		return doAction
	}

	//检查吃,多个玩家可以同时检测到吃
	activePlayer := room.GetActivePlayer()
	checkEndPlayer := activePlayer.GetShangJia()

	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {

		if player != nil &&
			(card.owner == nil || card.owner.id != player.id) &&
			(room.state != RSBankerChuPaiAfterStage || player.ptype != cmn.PTBanker) &&
			!player.HaveBao() {

			chiPatterns, biPatterns := player.controller.CheckChi(card)
			logger.Info("检查吃：", card)
			PrintPatterns(chiPatterns)
			logger.Info("检查吃时，的手牌:")
			PrintCards(player.cards)

			canChi := chiPatterns != nil && len(chiPatterns) > 0
			if canChi {
				if IsExistCard(player.guoCards, card) {
					player.sendPassedNotifyACK(card)
				} else {
					doAction = AChi
					player.SendActionNotifyACK(AChi, card, chiPatterns, biPatterns)
					logger.Info("CheckCanDoAction:玩家%s,检查到:%s, 有报吗：%s", player.id, actionName[AChi], player.HaveBao())
				}
			}
		}

		player = player.GetXiaJia()
		isInit = false
	}

	return doAction
}

//执行一个动作，根据当前等待的动作
func (room *DaerRoom) CheckDoAction(card *DaerCard) {
	if !room.CanCheckDoAction() {
		return
	}

	if card != nil {
		logger.Info("CheckDoAction:开始执行一次检测==============当前的房间状态：%s 检查的牌card:%s", rootTypeName[room.state], card.value)
	} else {
		logger.Info("CheckDoAction:开始执行一次检测=====当前的房间状态：%s", rootTypeName[room.state])
	}

	action := room.CheckCanDoAction(card)
	if card != nil {
		logger.Info("CheckDoAction:执行一次检测=====当前的房间状态：%s 准备执行的动作:%s card:%s", rootTypeName[room.state], actionName[action], card.value)
	} else {
		logger.Info("CheckDoAction:执行一次检测=====当前的房间状态：%s 准备执行的动作:%s", rootTypeName[room.state], actionName[action])
	}

	switch room.state {
	case RSBaoStage:
		if action == ANone {
			room.SwitchRoomState(RSNotBankerBaiStage)
			room.CheckDoAction(card)

		}
	case RSNotBankerBaiStage:
		if action == ANone {
			room.SwitchRoomState(RSBankerJinPaiStage)

			openCard := room.OpenOneCard()
			if openCard == nil {
				logger.Error("在报牌阶段桌面上竟然会没有牌了。")
				break
			}

			banker := room.GetBanker()
			if banker == nil {
				logger.Error("竟然没有庄家！太不可思议了")
				return
			}

			//banker.SendActionACK(AMo, openCard, nil, ACSuccess)

			room.CheckDoAction(openCard)
		}
	case RSBankerJinPaiStage:
		if action == ANone {
			room.BankerDoMo(card)
		}
	case RSBankerBaiStage:
		if action == ANone {
			room.SwitchRoomState(RSBankerBaoStage)

			banker := room.GetBanker()
			if banker == nil {
				logger.Error("竟然没有庄家！太不可思议了")
				return
			}

			banker.SendActionNotifyACK(AChu, nil, nil, nil)
			//room.CheckDoAction(room.activeCard)
		}
	case RSBankerBaoStage:
		if action == ANone {
			room.SwitchRoomState(RSBankerChuPaiAfterStage)
			doAction := room.DoDelayAction(true)
			if !IsBaiAction(doAction) {
				room.CheckDoAction(room.activeCard)
			}
		}
	case RSBankerChuPaiAfterStage:
		if action == ANone {
			room.SwitchRoomState(RSLoopWorkStage)
			room.CheckAndAddShowCard(card)
			room.CheckDoAction(nil)
		}
	case RSLoopWorkStage:
		if action == ANone {
			room.CheckAndAddShowCard(card)

			room.ChangeActivePlayerToNext()

			openCard := room.OpenOneCard()
			if openCard == nil {
				room.ChaJiao()
				room.SwitchRoomState(RSSettlement)
				logger.Info("准备开始结算了|||||||||||||")
				room.CheckDoAction(nil)
			} else {
				ap := room.GetActivePlayer()
				if ap == nil {
					return
				}

				delayCallId := ap.id + strconv.Itoa(AJin)
				room.StartDelayCallback(delayCallId, room.OpenCardDelay, func(data interface{}) {
					ap.SendActionACK(AJin, openCard, nil, ACSuccess)
					room.CheckDoAction(openCard)
				}, nil)
			}
		}
	case RSSettlement:
		logger.Info("进入了结算阶段.............")
		room.rs.DoJieSuan()
		//room.DoJieSuan()
	}
}

//庄家摸一张牌
func (room *DaerRoom) BankerDoMo(card *DaerCard) {
	banker := room.GetBanker()
	if banker == nil {
		logger.Error("竟然米哟庄家！")
		return
	}

	if card == nil {
		logger.Error("没有牌，庄家不能摸")
		return
	}

	//缓存拢之前的拢的数量,摸完牌后检查是否有新的拢牌
	//longCount := len(banker.showPatterns)

	//有报且有新的拢牌时，则直接进行下一轮检查， 否则就通知玩家出牌
	//haveNewLong := len(banker.showPatterns) > longCount
	if banker.HaveBao() /*&& haveNewLong */ {
		if canZhao := banker.ObtainCard(card); !canZhao {
			banker.SendActionACK(AMo, card, nil, ACSuccess)

			doAction := room.DoDelayAction(true)
			if !IsBaiAction(doAction) {
				banker.SendActionNotifyACK(AChu, nil, nil, nil)
			}
		} else {
			room.SwitchRoomState(RSBankerChuPaiAfterStage)
			doAction := room.DoDelayAction(true)
			if !IsBaiAction(doAction) {
				room.CheckDoAction(nil)
			}
		}
	} else {
		banker.SendActionACK(AMo, card, nil, ACSuccess)
		banker.ObtainCard(card)
		room.SwitchRoomState(RSBankerBaiStage)
		room.CheckDoAction(card)
	}
}

//查叫
func (room *DaerRoom) ChaJiao() {
	for _, player := range room.players {
		if player == nil {
			continue
		}

		player.controller.ChaJiao()
	}
}

//是否有人被查叫
func (room *DaerRoom) HavePlayerChaJiao() bool {
	if len(room.ownCards) <= 0 {
		for _, player := range room.players {
			if player == nil {
				continue
			}

			if val, exist := player.multipleCount[MTChaJiao]; exist && val > 0 {
				return true
			}
		}
	}

	return false
}

//結算
func (room *DaerRoom) DoJieSuan() {

	logger.Info("清扫房间准备下一场")

	//统计最后的名堂
	room.StatisticsMinTangForAll()

	//发送结算数据
	jieSuanPlayer, isHuangZhuang := room.SendJieSuanACKAndJieSuanCoinForAll()
	if !isHuangZhuang && !room.HavePlayerChaJiao() {
		if jieSuanPlayer != nil {
			room.huPaiPlayerID = jieSuanPlayer.id
		} else {
			logger.Error("没有黄庄，怎么可能没有结算的玩家（赢家）")
		}
	}

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
func (room *DaerRoom) CheckCoinForAll() {

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

//发送进入房间错误ACK
func SendEnterRoomErrorACK(uid string, roomType, code int32, isNormalReqEnterRoom bool) {
	rmMsg := &rpc.EnterRoomACK{}
	rmMsg.SetRoomId(roomType)
	rmMsg.SetCode(code)
	rmMsg.SetIsNormalReqEnterRoom(isNormalReqEnterRoom)
	//logger.Info("进入房间错消息：", rmMsg)
	if err := conn.SendEnterRoom(uid, rmMsg); err != nil {
		logger.Error("发送进入房间出错：", err)
	}
}

//执行延迟执行的的动作
func (room *DaerRoom) DoDelayAction(excute bool) int32 {

	//logger.Error("执行了DoDelayAction函数：", excute)
	//如果有摆牌的，把延迟执行报牌的取消掉
	for _, p := range room.players {
		if p == nil {
			continue
		}

		if IsBaiAction(p.delayDoAction) {
			for _, ip := range room.players {
				if ip == nil {
					continue
				}

				if ip.delayDoAction == ABao {
					ip.SendActionACK(ip.delayDoAction, nil, nil, ACAbandon)
					ip.SetDelayDoAction(ANone)
				}
			}
			break
		}
	}

	//执行延迟执行的动作
	var doAction int32 = ANone
	for _, p := range room.players {
		if p == nil || p.delayDoAction == ANone {
			continue
		}

		doAction = p.delayDoAction

		p.DoDelayAction(excute)
	}

	return doAction
}

//统计玩家的名堂
func (room *DaerRoom) StatisticsMinTangForAll() {
	for _, p := range room.players {
		if p == nil {
			continue
		}

		//黑摆是不统计胡子和名堂的
		// if p.HaveHeiBai() {
		// 	continue
		// }

		//ogger.Info("是否有摆牌：", p.HaveBai())
		if hu, _ := p.IsHu(true); hu {
			finalPatternGroup := p.controller.GenerateFinalPatternGroup()
			p.controller.StatisticsRemainMinTangAndSave(finalPatternGroup)
		}
	}
}

//获取胡牌的玩家
func (room *DaerRoom) GetHuPaiPlayers() []*DaerPlayer {
	result := make([]*DaerPlayer, 0)
	for _, p := range room.players {
		if p == nil {
			continue
		}

		haveHu := p.cards == nil || len(p.cards) <= 0 &&
			(p.showPatterns != nil && len(p.showPatterns) > 0 || p.showPatterns != nil && len(p.fixedpatterns) > 0)
		if haveHu {
			result = append(result, p)
		}
	}

	return result
}

//能否检查下一个动作的执行
func (room *DaerRoom) CanCheckDoAction() bool {
	if room.state == RSReady {
		return false
	}

	canDoAction := true
	for _, player := range room.players {
		if player == nil || player.watingAction != ANone || player.readyDoAction != ANone {
			canDoAction = false
			break
		}
	}

	return canDoAction
}

//获取等待此动作的玩家
func (room *DaerRoom) GetWatingActionPlayer(watingAction []int32) (players []*DaerPlayer, have bool) {
	if len(watingAction) <= 0 {
		return
	}

	players = make([]*DaerPlayer, 0)

	for _, player := range room.players {
		if player == nil {
			continue
		}

		for _, act := range watingAction {
			if act == player.watingAction {
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
func (room *DaerRoom) GetDifen() int32 {
	return room.Difen
}

//是否带归
func (room *DaerRoom) GetIsDaiGui() bool {
	return room.IsDaigui
}

//倍数上限
func (room *DaerRoom) GetMaxMultiple() int32 {
	return room.MaxMultiple
}

//抽成比率
func (room *DaerRoom) GetRakeRate() int32 {
	return room.RakeRate
}

//获取替用的数量
func (room *DaerRoom) GetTiYongAmount() int32 {
	return 0
}

//获取起胡颗数
func (room *DaerRoom) GetQiHuKeAmount() int32 {
	return 0
}

//获取ID
func (room *DaerRoom) ID() int32 {
	return room.rtype
}

//获取房间的唯一标示
func (room *DaerRoom) UID() int32 {
	return room.uid
}

//获取房间名字
func (room *DaerRoom) Name() string {
	return ""
}

//设置房间选择器
func (room *DaerRoom) SetSelector(rs cmn.RoomSelector) {
	room.rs = rs
}

//线程接受chan
func (room *DaerRoom) GetRcvThreadHandle() *chan cmn.RoomMsgQueue {
	return &room.rcv
}

//线程退出chan
func (room *DaerRoom) GetExitThreadHandle() *chan bool {
	return &room.exit
}

//设置当前胡牌的玩家
func (room *DaerRoom) SetHuPaiPlayerID(uid string) {
	room.huPaiPlayerID = uid
}

//获取准备执行此动作的玩家
func (room *DaerRoom) GetPlayersForReadyDoAction(doAction int32) (players []*DaerPlayer, have bool) {
	if doAction == ANone {
		return
	}

	players = make([]*DaerPlayer, 0)

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
func (room *DaerRoom) ResetReadyDoAction() {
	for _, player := range room.players {
		if player == nil {
			continue
		}
		player.SwitchReadyDoAction(ANone)
	}
}

//重置玩家等待执行的动作
func (room *DaerRoom) ResetWaitingDoAction() {
	for _, player := range room.players {
		if player != nil && player.HaveWaitingDoAction() {
			player.SendActionACK(player.watingAction, nil, nil, ACAbandon)
			player.SwitchWatingAction(ANone)
		}
	}
}

//重置所有玩家的延迟执行动作
func (room *DaerRoom) ResetAllDelayAction() {
	for _, player := range room.players {
		if player != nil && player.delayDoAction != ANone {
			player.SendActionACK(player.delayDoAction, nil, nil, ACAbandon)
			player.SetDelayDoAction(ANone)
		}
	}
}

//重置玩家的等待和准备执行动作
func (room *DaerRoom) ResetAllAction() {
	room.ResetWaitingDoAction()
	room.ResetReadyDoAction()
}

//活动玩家->下家->下家去执行一个动作（针对：摆，吃，胡）
func (room *DaerRoom) DoReadyActionByOrder(checkAllPlayer bool) (bool, *DaerPlayer) {
	activePlayer := room.GetActivePlayer()

	checkEndPlayer := activePlayer.GetShangJia()
	if checkAllPlayer {
		checkEndPlayer = activePlayer
	}

	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {
		if player.watingAction == ANone {
			if player.readyDoAction != ANone {
				return true, player
			}
		} else {
			break
		}

		player = player.GetXiaJia()
		isInit = false
	}

	return false, nil
}

//活动玩家->上家->上家去执行一个动作（针对：胡）---废弃了
func (room *DaerRoom) DoReadyActionByROrder(checkAllPlayer bool) (bool, *DaerPlayer) {
	activePlayer := room.GetActivePlayer()

	checkEndPlayer := activePlayer.GetXiaJia()
	if checkAllPlayer {
		checkEndPlayer = activePlayer
	}

	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {
		if player.watingAction == ANone {
			if player.readyDoAction != ANone {
				return true, player
			}
		} else {
			break
		}

		player = player.GetShangJia()
		isInit = false
	}

	return false, nil
}

//活动玩家->下家->下家去执行一个动作（针对：摆）
func (room *DaerRoom) DoDelayActionByOrder(checkAllPlayer bool) (bool, *DaerPlayer) {
	activePlayer := room.GetActivePlayer()

	checkEndPlayer := activePlayer.GetShangJia()
	if checkAllPlayer {
		checkEndPlayer = activePlayer
	}

	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {
		if player.delayDoAction != ANone {
			return true, player
		}

		player = player.GetXiaJia()
		isInit = false
	}

	return false, nil
}

//获取倒计时的玩家
func (room *DaerRoom) GetCountdownPlayer(checkAllPlayer bool) *DaerPlayer {

	for _, p := range room.players {
		logger.Info("倒计时获取活动玩家  ---------:", actionName[p.watingAction])
	}
	activePlayer := room.GetActivePlayer()

	checkEndPlayer := activePlayer.GetShangJia()
	if checkAllPlayer {
		checkEndPlayer = activePlayer
	}

	isInit := true
	for player := activePlayer; isInit || player != checkEndPlayer; {
		if player.watingAction != ANone {
			return player
		}

		player = player.GetXiaJia()
		isInit = false
	}
	return nil
}

//获取最优先结算的player
func (room *DaerRoom) GetMaxJieSuanPlayer() (resultPlayer *DaerPlayer, isAllHu bool) {

	isAllHu = true
	resultPlayer = nil
	var maxHuScore int32 = 0
	for _, p := range room.players {
		if p == nil {
			isAllHu = false
			continue
		}

		if hu, _ := p.IsHu(false); !hu {
			isAllHu = false
		}

		if hu, huScore := p.IsHu(false); hu && huScore > maxHuScore {
			maxHuScore = huScore
			resultPlayer = p
		}
	}

	return
}

//获取胡数对应的分数
func GetScoreByHu(hu int32) int32 {
	if hu == 0 {
		return SpecificHuScore[MTLuanHu]
	} else if hu < 10 {
		return 0
	}

	isWhole := (0 == (hu % 10))

	reviseHu := hu
	if !isWhole {
		reviseHu = int32(hu/10)*10 + 5
	}

	return HuScore[reviseHu]
}

//启动计时
func (room *DaerRoom) StartTimer(second int64) {

	room.timeMgr.StartTimer(MainTimerName, second, room.OnMainTimer, nil)

	logger.Info("开启一个主计时器")
}

//获取剩余倒计时
func (room *DaerRoom) GetRemainTime() int32 {
	return room.timeMgr.GetRemainTime(MainTimerName)
}

//是否在计时
func (room *DaerRoom) IsTiming() bool {
	return room.timeMgr.IsTiming(MainTimerName)
}

//停止计时器
func (room *DaerRoom) StopTimer() {
	room.timeMgr.StopTimer(MainTimerName)
}

//计时器的update
func (room *DaerRoom) UpdateTimer(ft time.Duration) {
	//logger.Info("倒计时中..", room.GetPlayerAmount())

	room.handleMsg()        //处理网络消息
	room.timeMgr.Update(ft) //处理内部的延迟消息

	//room.timeMgr.Update(time.Second)
}

//到时间的回调
func (room *DaerRoom) OnMainTimer(data interface{}) {

	room.StopTimer()
	logger.Info("桌面时间到。==============")

	for _, p := range room.players {
		if p == nil || p.watingAction == ANone {
			continue
		}

		//超时就把它设置为自动
		p.PlayerDoAction(ATuoGuan, room.activeCard, nil, nil)
		//p.SendActionACK(ATuoGuan, nil, nil, ACSuccess)
		//p.PlayerDoAction(AGuo, room.activeCard, nil, nil)
	}
}

//开启一个定时器
func (room *DaerRoom) StartDelayCallback(name string, delay int64, call cmn.TimerCallback, data interface{}) {
	room.timeMgr.StartTimer(name, delay, call, data)
}

//停止一个定时器
func (room *DaerRoom) StopDelayCallback(name string) {
	room.timeMgr.StopTimer(name)
}

//网络消息相关
func (room *DaerRoom) GetAllPlayerIDs() []string {

	result := make([]string, 0)
	for _, p := range room.players {
		if p != nil {
			result = append(result, p.id)
		}
	}

	return result
}

func (room *DaerRoom) SendEnterRoomACK(p cmn.Player) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*DaerPlayer)

	logger.Info("发送进入房间：", player.id)

	//打印房间信息
	logger.Info("房间情况：", room.GetPlayerAmount())

	//给自己发送ACK
	msg := &rpc.EnterRoomACK{}
	msg.SetRoomId(room.rtype)
	msg.SetShangjiaType(3)
	msg.SetBReady(player.isReady)
	msg.SetPlayerInfo(player.client)
	if err := conn.SendEnterRoom(player.id, msg); err != nil {
		logger.Error("给自己发送进入房间时出错：", err, msg)
		return
	}
	logger.Info("sned EnterRoomACK to self(%s).", player.id)

	//给房间里的每个发送进入房间
	shangJia := player.GetShangJia()
	if shangJia != nil {
		msg := &rpc.EnterRoomACK{}
		msg.SetRoomId(room.rtype)
		msg.SetShangjiaType(2)
		msg.SetBReady(player.isReady)
		msg.SetPlayerInfo(player.client)
		if err := conn.SendEnterRoom(shangJia.id, msg); err != nil {
			logger.Error("给上家发送自己进入房间时出错：", err, msg)
			return
		}

		msg = &rpc.EnterRoomACK{}
		msg.SetRoomId(room.rtype)
		msg.SetShangjiaType(1)
		msg.SetBReady(shangJia.isReady)
		msg.SetPlayerInfo(shangJia.client)
		if err := conn.SendEnterRoom(player.id, msg); err != nil {
			logger.Error("给自己发送上家进入房间时出错：", err, msg)
			return
		}

		logger.Info("sned EnterRoomACK to shangJia(%s).", shangJia.id)
	}

	xiaJia := player.GetXiaJia()
	if xiaJia != nil {
		msg := &rpc.EnterRoomACK{}
		msg.SetRoomId(room.rtype)
		msg.SetShangjiaType(1)
		msg.SetBReady(player.isReady)
		msg.SetPlayerInfo(player.client)
		if err := conn.SendEnterRoom(xiaJia.id, msg); err != nil {
			logger.Error("给下家发送自己进入房间时出错：", err, msg)
			return
		}

		msg = &rpc.EnterRoomACK{}
		msg.SetRoomId(room.rtype)
		msg.SetShangjiaType(2)
		msg.SetBReady(xiaJia.isReady)
		msg.SetPlayerInfo(xiaJia.client)
		if err := conn.SendEnterRoom(player.id, msg); err != nil {
			logger.Error("给自己发送下家进入房间时出错：", err, msg)
			return
		}

		logger.Info("sned EnterRoomACK to xiaJia(%s).", xiaJia.id)
	}
}

//向客户端发送玩家离开房间的消息
func (room *DaerRoom) SendLeaveRoomACK(p cmn.Player, isChangeDesk bool) {
	if p == nil {
		logger.Error("p is nil.")
		return
	}

	player := p.(*DaerPlayer)

	//给房间里的每个发送离开房间
	for _, p := range room.players {
		if p != nil {
			msg := &rpc.LeaveRoomACK{}
			msg.SetPlayerID(player.id)
			msg.SetIsChangeDesk(isChangeDesk)
			if err := conn.SendLeaveRoom(p.id, msg); err != nil {
				logger.Error("发送离开房间出错：", err, msg)
				continue
			}
		}
	}
}

//向客户端发送玩家离开房间的消息
func (room *DaerRoom) SendCommonMsg2Others(msg *rpc.FightRoomChatNotify) {
	logger.Info("SendCommonMsg2Others has been called")

	//给房间里的每个发送离开房间
	uids := []string{}
	for _, p := range room.players {
		if p != nil {
			uids = append(uids, p.id)
		}
	}
	conn.SendCommonNotify2S(uids, msg, "FightRoomChatNotify")
}

//发送游戏开始
func (room *DaerRoom) sendGameStartForAll() {
	for _, p := range room.players {
		if p == nil {
			continue
		}
		p.SendGameStartACK(false)
	}

	logger.Info("——————————房间状态：", rootTypeName[room.state])
}

//发送桌面通知信息
func (room *DaerRoom) sendCountdownNotifyACK() {

	cp := room.GetCountdownPlayer(true)
	if cp != nil {
		for _, p := range room.players {
			if p == nil {
				continue
			}

			if p.watingAction != ANone {
				p.sendCountdownNotifyACK(p)
			} else {
				p.sendCountdownNotifyACK(cp)
			}
		}
	} else {
		logger.Error("没有活动玩家")
	}

}

//发送游戏结算
func (room *DaerRoom) SendJieSuanACKAndJieSuanCoinForAll() (jieSuanPlayer *DaerPlayer, isHuangZhuang bool) {

	var isAllHu bool
	jieSuanPlayer, isAllHu = room.GetMaxJieSuanPlayer()

	isHuangZhuang = jieSuanPlayer == nil || isAllHu

	logger.Info("向所有发送结算信息: 结算Player：%s, AllHu:%s", jieSuanPlayer, isAllHu, isHuangZhuang)

	var coins []*rpc.JieSuanCoin = nil
	//var msg *rpc.JieSuanNotifyACK = nil

	for _, p := range room.players {
		if p == nil {
			continue
		}

		//发送数据
		addiData := &rpc.JieSuanAdditionData{}
		addiData.SetSysType(cmn.PiPeiFang)

		if hu, _ := p.IsHu(true); hu {
			coins = room.SendJieSuanACK(p.id, p, isHuangZhuang, true, addiData)
		} else {
			coins = room.SendJieSuanACK(p.id, jieSuanPlayer, isHuangZhuang, true, addiData)
		}
	}

	//扣取金币
	if coins != nil {
		for _, p := range room.players {
			if p == nil {
				continue
			}

			p.JieSuanCoin(coins)
		}
	}

	return
}

//发送结算
func (room *DaerRoom) SendJieSuanACK(uid string, jieSuanPlayer cmn.Player, huangZhuang, isCorrection bool, addiData *rpc.JieSuanAdditionData) (result []*rpc.JieSuanCoin) {
	if !huangZhuang && jieSuanPlayer == nil {
		logger.Error("结算错误：没有黄庄时结算的玩家却是nil.")
		return
	}

	//转换到DaerPlayer
	p, _ := jieSuanPlayer.(*DaerPlayer)
	if !huangZhuang && p == nil {
		logger.Error("结算错误：没有黄庄时转换玩家信息到DaerPlayer出错.", jieSuanPlayer)
		return
	}

	//发送结算信息
	msg := &rpc.JieSuanNotifyACK{}

	msg.SetHuangZhuang(huangZhuang)
	msg.DiCards = append(msg.DiCards, convertCards(room.ownCards)...)
	msg.SetAddi(addiData)

	//没有黄庄的时候
	if !huangZhuang && p != nil && p.controller != nil {

		result = fillDaerPlayerJieSuanPatternForHu(msg, p, isCorrection)
		for _, py := range room.players {
			if py.ID() == p.ID() {
				continue
			}

			if hu, _ := py.IsHu(true); hu {
				fillDaerPlayerJieSuanPatternForHu(msg, py, isCorrection)
			} else {
				fillDaerPlayerJieSuanPatternForNotHu(msg, py, result)
			}
		}

	} else {
		result = []*rpc.JieSuanCoin{}
		for _, p := range room.players {
			temp := &rpc.JieSuanCoin{}
			temp.SetPlayerID(p.id)
			temp.SetCoin(0)
			temp.SetTag(JSNone)
			result = append(result, temp)
		}
	}

	//通知客服端结算
	if err := conn.SendJieSuan(uid, msg); err != nil {
		logger.Error("发送结算消息出错:", err, uid, msg)
		return
	}

	logger.Info("黄庄：%s   结算消息：", huangZhuang, msg)

	return
}

func fillDaerPlayerJieSuanPatternForHu(jieSuanACK *rpc.JieSuanNotifyACK, p *DaerPlayer, isCorrection bool) (result []*rpc.JieSuanCoin) {
	if p == nil || p.controller == nil {
		logger.Error("p is nil.")
		return
	}

	if jieSuanACK == nil {
		logger.Error("jieSuanACK is nil.")
		return
	}

	logger.Info("有玩家结算：", p.id)
	finalPatternGroup := p.controller.GenerateFinalPatternGroup()
	if finalPatternGroup == nil {
		logger.Error("不能产生最终的模式")
		return
	}

	huAmount, huScore := p.controller.StatisticsHuAndScore(finalPatternGroup, true)
	result = fillJieSuanCoin(p, huScore, isCorrection)

	jieSuanMsg := &rpc.DaerPlayerJieSuanPattern{}
	jieSuanMsg.SetPlayerID(p.ID())
	jieSuanMsg.SetHu(int32(huAmount))
	jieSuanMsg.SetScore(int32(GetScoreByHu(huAmount) * 10))
	jieSuanMsg.Patterns = append(jieSuanMsg.Patterns, convertPatterns(finalPatternGroup.patterns)...)
	jieSuanMsg.MingTang = append(jieSuanMsg.MingTang, p.fillMingTang()...)
	jieSuanInfo := GetJieSuanCoinByID(result, p.ID())
	if jieSuanInfo != nil {
		jieSuanMsg.SetCoin(jieSuanInfo.GetCoin())
		jieSuanMsg.SetTag(jieSuanInfo.GetTag())
	} else {
		logger.Error("jieSuanInfo is nil.")
	}

	jieSuanACK.DaerPlayerJieSuanPattern = append(jieSuanACK.DaerPlayerJieSuanPattern, jieSuanMsg)

	return
}

func fillDaerPlayerJieSuanPatternForNotHu(jieSuanACK *rpc.JieSuanNotifyACK, p *DaerPlayer, jieSuanCoin []*rpc.JieSuanCoin) {
	if p == nil || p.controller == nil {
		logger.Error("p is nil.")
		return
	}

	if jieSuanACK == nil {
		logger.Error("jieSuanACK is nil.")
		return
	}

	jieSuanMsg := &rpc.DaerPlayerJieSuanPattern{}
	jieSuanMsg.SetPlayerID(p.ID())

	finalPatterns := []*DaerPattern{}
	finalPatterns = append(finalPatterns, p.showPatterns...)
	finalPatterns = append(finalPatterns, p.fixedpatterns...)
	finalPatterns = append(finalPatterns, NewPattern(PTSingle, p.cards))
	jieSuanMsg.Patterns = append(jieSuanMsg.Patterns, convertPatterns(finalPatterns)...)
	jieSuanInfo := GetJieSuanCoinByID(jieSuanCoin, p.ID())
	if jieSuanInfo != nil {
		jieSuanMsg.SetCoin(jieSuanInfo.GetCoin())
		jieSuanMsg.SetTag(jieSuanInfo.GetTag())
	} else {
		logger.Error("jieSuanInfo is nil.")
	}

	jieSuanACK.DaerPlayerJieSuanPattern = append(jieSuanACK.DaerPlayerJieSuanPattern, jieSuanMsg)
}

func GetJieSuanCoinByID(jieSuanInfo []*rpc.JieSuanCoin, pid string) *rpc.JieSuanCoin {
	if jieSuanInfo == nil {
		logger.Error("jieSuanInfo is nil.")
		return nil
	}

	for _, jieSuan := range jieSuanInfo {
		if jieSuan.GetPlayerID() == pid {
			return jieSuan
		}
	}

	return nil
}

// func (room *DaerRoom) SendJieSuanACK(uid string, jieSuanPlayer cmn.Player, huangZhuang, isCorrection bool, addiData *rpc.JieSuanAdditionData) (result []*rpc.JieSuanCoin) {
// 	if !huangZhuang && jieSuanPlayer == nil {
// 		logger.Error("结算错误：没有黄庄时结算的玩家却是nil.")
// 		return
// 	}

// 	//转换到DaerPlayer
// 	p, _ := jieSuanPlayer.(*DaerPlayer)
// 	if !huangZhuang && p == nil {
// 		logger.Error("结算错误：没有黄庄时转换玩家信息到DaerPlayer出错.", jieSuanPlayer)
// 		return
// 	}

// 	//发送结算信息
// 	msg := &rpc.JieSuanNotifyACK{}

// 	msg.SetHuangZhuang(huangZhuang)
// 	msg.SetScore(0)
// 	msg.SetHu(0)
// 	msg.SetAddi(addiData)

// 	//没有黄庄的时候
// 	if !huangZhuang && p != nil && p.controller != nil {
// 		logger.Info("有玩家结算：", p.id)
// 		finalPatternGroup := p.controller.GenerateFinalPatternGroup()
// 		if finalPatternGroup == nil {
// 			logger.Error("不能产生最终的模式")
// 			return
// 		}

// 		huAmount, huScore := p.controller.StatisticsHuAndScore(finalPatternGroup, true)

// 		msg.SetHu(int32(huAmount))
// 		msg.SetScore(int32(GetScoreByHu(huAmount) * 10))

// 		msg.Patterns = append(msg.Patterns, convertPatterns(finalPatternGroup.patterns)...)
// 		msg.DiCards = append(msg.DiCards, convertCards(p.room.ownCards)...)
// 		msg.MingTang = append(msg.MingTang, p.fillMingTang()...)
// 		//logger.Error("HuScore:", huScore)
// 		msg.Coin = append(msg.Coin, fillJieSuanCoin(p, huScore, isCorrection)...)

// 		result = msg.Coin
// 	} else {
// 		result = []*rpc.JieSuanCoin{}
// 		for _, p := range room.players {
// 			temp := &rpc.JieSuanCoin{}
// 			temp.SetPlayerID(p.id)
// 			temp.SetCoin(0)
// 			temp.SetTag(JSNone)
// 			result = append(result, temp)
// 		}
// 	}

// 	//通知客服端结算
// 	if err := conn.SendJieSuan(uid, msg); err != nil {
// 		logger.Error("发送结算消息出错:", err, uid, msg)
// 		return
// 	}

// 	logger.Info("黄庄：%s   结算消息：", huangZhuang, msg)

// 	return

// }

//通过户数和最大的倍数，获取分数
// func GetScoreByHuAndMultiple(hu, mul int32) int32 {
// 	return hu * int32(math.Pow(2, float64(mul)))
// }

//接收客户端发来的消息
func (room *DaerRoom) OnPlayerDoAction(msg *rpc.ActionREQ) {
	if msg == nil {
		logger.Error("DaerRoom.OnPlayerDoAction:客户端发送来的数据为空！")
		return
	}

	// msg := m.(*rpc.ActionREQ)
	// if msg == nil {
	// 	logger.Error("转换消息失败")
	// 	return
	// }

	logger.Info("receive client msg:(Action:%s, playerID:%s)", msg.GetAction(), msg.GetPlayerID())

	player := room.GetPlayerByID(msg.GetPlayerID())
	if player == nil {
		logger.Error("DaerRoom.OnPlayerDoAction:此房间没有这个人%s", msg.GetPlayerID())
		return
	}

	//玩家吃牌
	action := int32(msg.GetAction())

	if action == AChi {
		if chiPattern := msg.GetChiCards(); chiPattern != nil {
			if room.activeCard == nil {
				logger.Error("没有活动牌")
				return
			}

			kaoCards := GetKaoCardsByRPC(chiPattern, room.activeCard)
			if kaoCards == nil || len(kaoCards) != 2 {
				logger.Error("DaerRoom.OnPlayerDoAction:吃牌有问题", msg.GetPlayerID(), chiPattern, room.activeCard)
				PrintCard(room.activeCard)
				return
			}

			biCards := msg.GetBiCards()
			if biCards == nil {
				logger.Info("客户端发送要吃的牌上来:", chiPattern)
				player.PlayerDoAction(action, room.activeCard, kaoCards, nil)
			} else {
				logger.Info("客户端发送要吃比的牌上来:", biCards)
				player.PlayerDoAction(action, room.activeCard, kaoCards, convertCardsToDaerCards(biCards.Cards))
			}
		} else {
			logger.Error("客户端没有发送吃的牌上来")
		}

	} else if action == AChu {
		if chuPai := msg.GetChuPai(); chuPai != nil {
			daerChuPai := convertCardToDaerCard(chuPai)
			logger.Info("客户端发送要出的牌上来:", daerChuPai)

			daerChuPai.owner = player
			player.PlayerDoAction(action, daerChuPai, nil, nil)
		} else {
			logger.Error("客户端没有发送要出的牌上来")
		}
	} else {
		logger.Info("客服端发送来的的动作：", actionName[action])
		player.PlayerDoAction(action, room.activeCard, nil, nil)
	}
}

//解散房间的接口
func (room *DaerRoom) OnJieSanRoom(uid string, msg *rpc.JieSanRoomREQ) {

}
