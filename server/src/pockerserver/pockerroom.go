package pockerserver

import (
	"centerclient"
	gp "code.google.com/p/goprotobuf/proto"
	"common"
	"logger"
	//"math"
	"connector"
	"lockclient"
	"math/rand"
	"rpc"
	// "runtime/debug"
	"proto"
	"strconv"
	"strings"
	"sync"
	"time"
	"timer"
)

const (
	ACT_LEAVE       = iota + 1 //离开
	ACT_STANDUP                //起身
	ACT_CHANGE_DESK            //换桌
	ACT_FOLD                   //弃牌
	ACT_CHECK                  //看牌
	ACT_CALL                   //跟注
	ACT_RAISE                  //加注
	ACT_ALLIN                  //ALLIN
	ACT_SEATDOWN               //坐下
	ACT_TIP                    //打赏
)
const (
	ACT_DEALER         = 30 //定庄
	ACT_GIVE_POCKER    = 31 //发牌
	ACT_COUNTDOWN      = 32 //开始倒计时
	ACT_SHOW_POCKER    = 33 //翻牌
	ACT_GAMEOVER       = 34 //比牌
	ACT_PLAYER_JOIN    = 35 //新玩家加入
	ACT_EXCHANGE_CHIPS = 36 //兑换筹码
	ACT_ROUND_OVER     = 37 //一轮结束收取筹码、彩池
)

const (
	ROUND_CNT  = 4  //总回合数
	POCKER_CNT = 52 //扑克总数
)

var ROOM_PLAYERS_LIMIT int32 //房间人数上限
var GAME_OVER_WAIT int32     //游戏结束等待时间

type PockerRoom struct {
	players      []*pockerman //桌上玩家
	standPlayers []*pockerman //站起的玩家
	pockers      []pocker     //公共牌
	leftpockers  []pocker     //没使用过的牌
	eType        int32        //房间类型
	// Antes        int32           //底注
	RakeRate    int32           //抽成比率
	t           *timer.Timer    //定时器
	msgQueue    []*RoomActQueue //消息队列
	rcv         chan *RoomActQueue
	exit        chan bool
	ql          sync.Mutex     //队列锁
	pl          sync.Mutex     //players 锁
	playerNum   uint16         //桌上人数，包括预定
	dIndex      int32          //庄家索引
	pot         int32          //总筹码
	spkIndex    int32          //说话玩家索引
	raiseIdx    int32          //最近1次加注人的索引
	rounds      int32          //已经过去的回合数
	attends     int32          //此次总人数
	allined     bool           //本轮是否allin
	foldsNum    int32          //弃牌人数
	allinsNum   int32          //allin人数
	pots        []int32        //彩池
	potLeft     int32          //本轮最后一个彩池
	candi       [][]*pockerman //参与分彩池的玩家
	overtime    int            //游戏结束时 时间
	lastAct     int            //最近1次动作
	lastValue   int32          //最近1次动作值(eg.跟50)
	bigblind    int32          //大盲注
	smallblind  int32          //小盲注
	ft          [6]*time.Timer //弃牌延时结算timer
	cdtOver     bool           //是否无延迟结束
	playWin     bool           //是否播放胜利动画
	roomNo      int32          //房间号
	bigBlindUid string         //大盲注uid
	sbValue     int32          //小盲注金额
	bigValue    int32          //大盲注金额

	customInfo *CustomInfo //自建房信息
}

type CustomInfo struct {
	limMin       int32 //进入金币上限
	limMax       int32 //进入金币下限
	exchngeValue int32 //每次兑换金额
}

func NewCustomPockerRoom(blindId, limId int32) *PockerRoom {
	logger.Info("*************NewCustomPockerRoom:blindId:%d, limId:%d", blindId, limId)
	r := &PockerRoom{
		// Antes:    10,
		RakeRate: 10,
		eType:    int32(0),
		overtime: int(time.Now().Unix()),
		rcv:      make(chan *RoomActQueue, 10),
		players:  make([]*pockerman, ROOM_PLAYERS_LIMIT),
		dIndex:   int32(0),
	}

	//大小盲注
	cfg := common.GetDaerGlobalConfig(strconv.Itoa(int(blindId)))
	if cfg == nil {
		logger.Error("GetDaerGlobalConfig(:%d) return nil", blindId)
		return nil
	}
	arrs := strings.Split(cfg.StringValue, "_")
	if len(arrs) != 3 {
		logger.Error("NewCustomPockerRoom cfg.StringValue:%s , blindId:%d err", cfg.StringValue, blindId)
		return nil
	}
	info := &CustomInfo{}

	v1, _ := strconv.Atoi(arrs[0])
	v2, _ := strconv.Atoi(arrs[1])
	r.sbValue = int32(v1)
	r.bigValue = int32(v2)

	v1, _ = strconv.Atoi(arrs[2])
	info.exchngeValue = int32(v1)

	logger.Info("*************NewCustomPockerRoom 1")
	//进入上下限
	cfg = common.GetDaerGlobalConfig(strconv.Itoa(int(limId)))
	if cfg == nil {
		logger.Error("GetDaerGlobalConfig(:%d) return nil", limId)
		return nil
	}
	arrs = strings.Split(cfg.StringValue, "_")
	if len(arrs) != 2 {
		logger.Error("NewCustomPockerRoom cfg.StringValue:%s , limId:%d err", cfg.StringValue, limId)
		return nil
	}
	v1, _ = strconv.Atoi(arrs[0])
	v2, _ = strconv.Atoi(arrs[1])
	info.limMin = int32(v1)
	info.limMax = int32(v2)

	r.customInfo = info

	r.msgQueue = []*RoomActQueue{}
	r.t = timer.NewTimer(time.Second)
	r.t.Start(
		func() {
			now := int32(time.Now().Unix())
			r.update(now)
		},
	)
	go r.process()

	return r
}

func NewPockerRoom(rtype int32) *PockerRoom {
	r := &PockerRoom{
		// Antes:    10,
		RakeRate: 10,
		eType:    rtype,
		overtime: int(time.Now().Unix()),
		rcv:      make(chan *RoomActQueue, 10),
		players:  make([]*pockerman, ROOM_PLAYERS_LIMIT),
		dIndex:   int32(0),
	}

	//获取底注
	cfg := common.GetDaerRoomConfig(strconv.Itoa(int(rtype)))
	if cfg == nil {
		logger.Error("common.GetDaerRoomConfig(rtype)", rtype)
	} else {
		// r.Antes = cfg.Difen
		r.sbValue = cfg.Difen / 2
		r.bigValue = cfg.Difen

		r.RakeRate = cfg.RakeRate
	}

	r.msgQueue = []*RoomActQueue{}
	r.t = timer.NewTimer(time.Second)
	r.t.Start(
		func() {
			now := int32(time.Now().Unix())
			r.update(now)
		},
	)
	go r.process()

	return r
}

func (r *PockerRoom) AtomicInc() bool {
	r.pl.Lock()
	defer r.pl.Unlock()
	if r.IsFull() {
		logger.Info("当前人数:%d", r.playerNum)
		return false
	}

	r.playerNum++
	logger.Info("当前人数:%d", r.playerNum)
	return true
}

func (r *PockerRoom) GetPlayerNum() int32 {
	r.pl.Lock()
	defer r.pl.Unlock()

	logger.Info("当前人数:%d", r.playerNum)
	return int32(r.playerNum)
}

func (r *PockerRoom) IsFull() bool {
	if int32(r.playerNum) >= ROOM_PLAYERS_LIMIT {
		return true
	}
	return false
}

func (r *PockerRoom) AtomicDesc() {
	r.pl.Lock()
	defer r.pl.Unlock()
	if r.playerNum == uint16(0) {
		logger.Error("AtomicDesc r.playerNum is 0")
		return
	}
	r.playerNum--
	logger.Info("当前人数:%d", r.playerNum)
}

func (r *PockerRoom) update(now int32) {
	r.handleMsg()
	for _, v := range r.players {
		if v == nil {
			continue
		}
		v.Update(now)
	}
	r.check_start_game(now)
}

func (r *PockerRoom) check_start_game(now int32) {
	overwait := GAME_OVER_WAIT*int32(len(r.pots)) + int32(2)
	if !r.playWin {
		overwait = int32(3)
	}
	if r.overtime <= 0 || int32(int(now)-r.overtime) < overwait {
		if r.overtime > 0 {
			logger.Info("*************下局开始倒计时:%d", int32(int(time.Now().Unix())-r.overtime))
		}
		return
	}

	logger.Info("*************interval:%d", int32(int(time.Now().Unix())-r.overtime))
	logger.Info("*************overwait:%d GAME_OVER_WAIT:%d", overwait, GAME_OVER_WAIT)
	logger.Info("*************len(r.pots):%d", len(r.pots))

	cnt := 0
	for _, v := range r.players {
		if v == nil {
			continue
		}

		if v.status != STATUS_STAND {
			cnt++
		}
	}
	if cnt < 2 {
		return
	}

	logger.Info("游戏开始...")
	if !r.shuffle() {
		return
	}
	r.reset_room()
	r.start_init()
	r.givePocker()
	r.nextplayer()
}
func (r *PockerRoom) reset_room() {
	logger.Info("********** 重置房间reset_room called")
	r.pockers = []pocker{}
	// r.dIndex = int32(0)
	r.pot = int32(0)
	r.spkIndex = int32(0)
	r.raiseIdx = int32(0)
	r.rounds = int32(0)
	r.allined = false
	r.pots = []int32{}
	r.potLeft = int32(0)
	r.candi = [][]*pockerman{}
	r.overtime = 0
	r.lastAct = 0
	r.bigblind = int32(0)
	r.smallblind = int32(0)
	r.foldsNum = int32(0)
	r.allinsNum = int32(0)
	r.cdtOver = false
	// r.forcewinner = ""

	for _, v := range r.players {
		if v == nil {
			continue
		}
		if v.autofold {
			if !lockclient.IsOnline(v.baseinfo.GetUid()) {
				// r.standup(v.baseinfo.GetUid())
				r.leave(v.baseinfo.GetUid())
			}
		}
		v.rest_data()
	}
	r.attends = int32(len(r.get_attend_playeruids()))
}

func (r *PockerRoom) get_attend_playeruids() []string {
	mans := []string{}
	for _, v := range r.players {
		if v == nil {
			continue
		}
		if v.status == STATUS_STAND || v.status == STATUS_FOLD {
			continue
		}
		mans = append(mans, v.baseinfo.GetUid())
	}
	return mans
}

func (r *PockerRoom) get_noraml_player_cnts() int32 {
	cnt := int32(0)
	for _, v := range r.players {
		if v == nil {
			continue
		}

		if v.status == STATUS_RAISE || v.status == STATUS_CALL || v.status == STATUS_CHEDK || v.status == STATUS_ALLIN {
			cnt++
		}
	}
	return cnt
}

func (r *PockerRoom) start_init() bool {
	logger.Info("初始化 start_init...")
	index := r.nextindex(r.dIndex)
	if index == -1 {
		logger.Error("r.nextindex() return -1")
		return false
	}
	r.dIndex = index

	//小盲
	rIndex := r.nextindex(index)
	if index == -1 {
		logger.Error("r.nextindex() return -1")
		return false
	}
	if r.players[rIndex] == nil {
		logger.Error("小盲 is nil, rIndex:%d", rIndex)
		return false
	}
	r.players[rIndex].drops += r.sbValue
	r.players[rIndex].chips -= r.sbValue
	// r.players[rIndex].status = STATUS_CALL
	r.smallblind = rIndex
	logger.Info("小盲，下注:%d 索引：%d", r.players[rIndex].drops, rIndex)

	//大盲
	bIndex := r.nextindex(rIndex)
	if index == -1 {
		logger.Error("r.nextindex() return -1")
		return false
	}
	if r.players[bIndex] == nil {
		logger.Error("大盲 is nil, rIndex:%d", bIndex)
		return false
	}
	r.players[bIndex].drops += r.bigValue
	r.players[bIndex].chips -= r.bigValue
	// r.players[bIndex].status = STATUS_CALL
	r.bigblind = bIndex
	// r.spkIndex = r.bigblind
	r.spkIndex = int32(-1)
	r.raiseIdx = r.bigblind
	logger.Info("大盲，下注:%d 索引：%d", r.players[bIndex].drops, bIndex)
	r.lastAct = ACT_RAISE
	r.lastValue = r.bigValue

	logger.Info("小盲：%d, 大盲:%d, spkIndex：%d, raiseIdx:%d", r.smallblind, r.bigblind, r.spkIndex, r.raiseIdx)

	//同步消息
	msg := &rpc.S2CAction{}
	msg.SetOperater(r.players[r.dIndex].baseinfo.GetUid())
	msg.SetAct(int32(ACT_DEALER))
	info := &rpc.PockerBegin{}
	info.SetDealerUid(r.players[r.dIndex].baseinfo.GetUid())
	info.SetSmBlindUid(r.players[r.smallblind].baseinfo.GetUid())
	info.SetBigBlindUid(r.players[r.bigblind].baseinfo.GetUid())
	allUids := r.get_attend_playeruids()
	info.AttendUids = append(info.AttendUids, allUids...)
	msg.SetBeginInfo(info)

	//大小盲注信息
	sp := r.players[r.smallblind]
	mbase := &rpc.PockerManBase{}
	mbase.SetUid(sp.baseinfo.GetUid())
	mbase.SetStatus(int32(sp.status))
	mbase.SetCoin(sp.chips)
	mbase.SetDrops(sp.drops)
	msg.Infos = append(msg.Infos, mbase)

	bp := r.players[r.bigblind]
	bbase := &rpc.PockerManBase{}
	bbase.SetUid(bp.baseinfo.GetUid())
	bbase.SetStatus(int32(bp.status))
	bbase.SetCoin(bp.chips)
	bbase.SetDrops(bp.drops)
	msg.Infos = append(msg.Infos, bbase)
	r.SendMsg2Others(msg, "S2CAction")

	return true
}

func (r *PockerRoom) handleMsg() {
	r.ql.Lock()
	defer r.ql.Unlock()
	bLoop := true
	for bLoop {
		bLoop = false
		for index, e := range r.msgQueue {
			if e.Func == "Action" {
				msg := e.Msg
				s, ok := msg.(rpc.C2SAction)
				if !ok {
					logger.Error("handleMsg  msg.(rpc.C2SAction) error")
					return
				}
				logger.Info("分发消息uid:%s, act:%d，len:%d", s.GetUid(), s.GetAct(), len(r.msgQueue))
				r.Action(&s)
				r.msgQueue = append(r.msgQueue[:index], r.msgQueue[index+1:]...)
				bLoop = true
			} else if e.Func == "ReEnter" {
				msg := e.Msg
				s, ok := msg.(rpc.PlayerBaseInfo)
				if !ok {
					logger.Error("handleMsg  msg.(rpc.PlayerBaseInfo) error")
					return
				}
				r.ReEnter(&s)
				r.msgQueue = append(r.msgQueue[:index], r.msgQueue[index+1:]...)
				bLoop = true
			} else if e.Func == "Enter" {
				msg := e.Msg
				s, ok := msg.(pockerman)
				if !ok {
					logger.Error("handleMsg  msg.(pockerman) error")
					return
				}
				r.Enter(&s)
				r.msgQueue = append(r.msgQueue[:index], r.msgQueue[index+1:]...)
				bLoop = true
			} else {
				logger.Error("handleMsg Func cant't find:%s", e.Func)
				bLoop = true
			}
			break
		}
	}
}

func (room *PockerRoom) process() {
	for {
		select {
		case r := <-room.rcv:
			logger.Info("process 收到消息")
			room.ql.Lock()
			room.msgQueue = append(room.msgQueue, r)
			room.ql.Unlock()
		case <-room.exit:
			logger.Info("process 房间收到退出消息")
			room.t.Stop()
			return
		}
	}
}

//向其它玩家发送消息
func (room *PockerRoom) SendMsg2Others(msg gp.Message, Func string) {
	logger.Info("-----------------SendMsg2Others发送消息给其它玩家:%s", Func)

	s, ok := msg.(*rpc.S2CAction)
	if ok && s.GetAct() == ACT_DEALER {
		logger.Error("**************S2C定庄消息")
	}

	uids := []string{}
	for _, p := range room.players {
		if p == nil {
			continue
		}
		uids = append(uids, p.baseinfo.GetUid())
	}

	for _, p := range room.standPlayers {
		if p == nil {
			continue
		}
		uids = append(uids, p.baseinfo.GetUid())
	}
	centerclient.SendCommonNotify2S(uids, msg, Func)
}

//进入房间
func (room *PockerRoom) Enter(player *pockerman) {
	//检查输入参数
	if player == nil {
		logger.Error("player is nil.")
		return
	}

	//检查能否进入房间
	// if room.IsFull() {
	// 	logger.Error("PockerRoom.Enter: room is full")
	// 	return
	// }

	//修改玩家的房间相关的信息
	player.room = room
	for i := 0; i < int(ROOM_PLAYERS_LIMIT); i++ {
		if room.players[i] == nil {
			room.players[i] = player
			// room.attends += int32(1)
			player.posIndex = int32(i)

			room.exchange(player)
			room.SyncRoomInfo(player.baseinfo.GetUid())
			room.TellOthersIJoin(player)
			if len(room.get_attend_playeruids()) == 2 {
				room.overtime = int(time.Now().Unix())
			}
			logger.Info("PockerRoom.Enter: player:%s enter room(%s):", player.baseinfo.GetName(), i)
			break
		}
	}
}

//下发牌桌上的所有信息
func (r *PockerRoom) SyncRoomInfo(uid string) {
	pri := &rpc.PockerRoomInfo{}
	for _, v := range r.players {
		if v == nil {
			continue
		}
		pmb := &rpc.PockerManBase{}
		pmb.SetUid(v.baseinfo.GetUid())
		pmb.SetHeaderUrl(v.baseinfo.GetHeaderUrl())
		pmb.SetCoin(v.chips)
		pmb.SetDrops(v.drops)
		pmb.SetStatus(int32(v.status))
		pmb.SetDeskIdx(int32(v.posIndex) + 1)
		pmb.SetNickName(v.baseinfo.GetName())
		pmb.SetSex(v.baseinfo.GetSex())
		for n, e := range v.pockers {
			if n >= 2 {
				break
			}
			pk := &rpc.Pocker{}
			pk.SetEType(int32(e.eType))
			pk.SetNum(e.num)
			pmb.Pockers = append(pmb.Pockers, pk)
		}
		if v.waitFrom > 0 {
			pmb.SetEndTime(v.waitFrom + COUNTDOWN_MAX)
		}
		pri.Players = append(pri.Players, pmb)
	}

	//fill roombase
	rbase := &rpc.PockerRoomBase{}
	for _, e := range r.pockers {
		// if n >= 2 {
		// 	break
		// }
		pk := &rpc.Pocker{}
		pk.SetEType(int32(e.eType))
		pk.SetNum(e.num)
		rbase.Pockers = append(rbase.Pockers, pk)
	}
	rbase.Pots = append(rbase.Pots, r.pots...)
	if len(r.get_attend_playeruids()) > 1 {
		if r.players[r.dIndex] == nil {
			logger.Error("SyncRoomInfo 庄家不能为nil, dIndex:%d", r.dIndex)
		} else {
			rbase.SetDealerUid(r.players[r.dIndex].baseinfo.GetUid())
		}
	}

	// rbase.SetPot(r.pot)
	// rbase.SetDIndex(r.dIndex)
	rbase.SetSmallBlind(r.sbValue)
	rbase.SetBigBlind(r.bigValue)
	rbase.SetRoomNo(r.roomNo)
	pri.SetRoombase(rbase)

	uids := []string{}
	uids = append(uids, uid)
	centerclient.SendCommonNotify2S(uids, pri, "PockerRoomInfo")
}

//通知其它玩家，有新人加入
func (r *PockerRoom) TellOthersIJoin(player *pockerman) {
	logger.Info("TellOthersIJoin begin")
	defer logger.Info("TellOthersIJoin end")

	info := player.baseinfo
	base := &rpc.PockerManBase{}
	base.SetUid(info.GetUid())
	base.SetHeaderUrl(info.GetHeaderUrl())
	base.SetCoin(player.chips)
	base.SetStatus(int32(player.status))
	base.SetDeskIdx(player.posIndex + int32(1))
	base.SetNickName(info.GetName())
	base.SetSex(info.GetSex())

	msg := &rpc.S2CAction{}
	msg.SetOperater(info.GetUid())
	msg.SetAct(ACT_PLAYER_JOIN)
	msg.Infos = append(msg.Infos, base)

	r.SendMsg2Others(msg, "S2CAction")
}

func (r *PockerRoom) is_inroom(id string) bool {
	if r.GetPlayerByID(id) != nil || r.get_stand_player(id) != nil {
		return true
	}
	return false
}

func (r *PockerRoom) GetPlayerByID(id string) *pockerman {
	for _, player := range r.players {
		if player != nil && player.baseinfo.GetUid() == id {
			return player
		}
	}
	return nil
}

func (r *PockerRoom) get_stand_player(id string) *pockerman {
	for _, player := range r.standPlayers {
		if player != nil && player.baseinfo.GetUid() == id {
			return player
		}
	}
	return nil
}

func (r *PockerRoom) remove_stand_player(id string) bool {
	for index, player := range r.standPlayers {
		if player != nil && player.baseinfo.GetUid() == id {
			r.standPlayers = append(r.standPlayers[:index], r.standPlayers[index+1:]...)
			return true
		}
	}
	return false
}

//重新进入房间
func (r *PockerRoom) ReEnter(playerInfo *rpc.PlayerBaseInfo) {
	logger.Info("ReEnter begin")
	if playerInfo == nil {
		logger.Error("ReEnter param playerInfo is nil")
		return
	}

	r.shows_all_players()
	p := r.GetPlayerByID(playerInfo.GetUid())
	if p != nil {
		p.baseinfo = playerInfo
		r.SyncRoomInfo(playerInfo.GetUid())
		return
	}

	p = r.get_stand_player(playerInfo.GetUid())
	if p == nil {
		logger.Error("PockerRoom:player not in the room")
		return
	}
	p.baseinfo = playerInfo
	r.seatdown(playerInfo.GetUid())
}

//离开房间
func (r *PockerRoom) leave(uid string) bool {
	logger.Info("离开房间：", uid)
	if uid == "" {
		logger.Error("离开房间，参数uid为空")
		return false
	}

	bStand := false
	p := r.GetPlayerByID(uid)
	if p == nil {
		p = r.get_stand_player(uid)
		if p == nil {
			logger.Error("Leave r.GetPlayerByID return nil, uid:%s", uid)
			return false
		}
		bStand = true
	}

	if bStand {
		logger.Info("玩家离开房间，处于站起状态")
	} else {
		logger.Info("玩家离开房间，处于坐下状态")
	}

	if p.islittle_blind() {
		r.smallblind = r.nextindex(p.posIndex)
	}

	r.sync_s2c_status(p, ACT_LEAVE)
	if p.status == STATUS_THINKING {
		r.spkIndex = p.posIndex
	}

	msg := &rpc.LeavePockerRoom{}
	msg.SetUid(uid)
	centerclient.SendCommonNotify2S([]string{uid}, msg, "LeavePockerRoom")

	if !bStand {
		centerclient.SendCostResourceMsg(p.baseinfo.GetUid(), connector.RES_COIN, "pocker", p.chips)
		r.potLeft += p.drops
		r.players[p.posIndex] = nil
		r.AtomicDesc()
		r.foldsNum += int32(1)
		r.players[p.posIndex] = nil

		//check gameover
		if r.forceover() {
			r.force_over()
			return false
		}
		if r.nextindex(p.posIndex) == -1 {
			r.over_undelay(r.rounds)
			return false
		}
	}

	r.remove_stand_player(uid)
	if r.spkIndex == p.posIndex {
		r.nextplayer()
	}
	return true
}

//离开房间
func (r *PockerRoom) change_desk(uid string) bool {
	logger.Info("离开房间：", uid)
	if uid == "" {
		logger.Error("离开房间,参数uid为空")
		return false
	}

	p := r.GetPlayerByID(uid)
	if p != nil {
		logger.Error("离开房间出错，玩家不是站起状态：%s", p.baseinfo.GetUid())
		return false
	}

	p = r.get_stand_player(uid)
	if p == nil {
		logger.Error("Leave r.GetPlayerByID return nil, uid:%s", uid)
		return false
	}

	if p.islittle_blind() {
		r.smallblind = r.nextindex(p.posIndex)
	}

	if p.status == STATUS_THINKING {
		r.spkIndex = p.posIndex
	}

	r.sync_s2c_status(p, ACT_CHANGE_DESK)
	r.remove_stand_player(uid)

	if r.spkIndex == p.posIndex {
		r.nextplayer()
	}
	return true
}

func (r *PockerRoom) nextindex(index int32) int32 {
	for i := 0; i < len(r.players); i++ {
		index++
		if index >= int32(len(r.players)) {
			index = 0
		}

		if r.players[index] == nil {
			continue
		}

		status := r.players[index].status
		if status != STATUS_STAND && status != STATUS_FOLD && status != STATUS_WATTING_JOIN && status != STATUS_ALLIN {
			return index
		}
	}
	return -1
}

func (r *PockerRoom) shows_all_players() {
	logger.Info("################正常房间列表")
	for _, v := range r.players {
		if v == nil {
			continue
		}
		logger.Info("玩家名字:%s, uid:%s", v.baseinfo.GetName(), v.baseinfo.GetUid())
	}

	logger.Info("################站起列表")
	for _, v := range r.standPlayers {
		if v == nil {
			continue
		}
		logger.Info("玩家名字:%s, uid:%s", v.baseinfo.GetName(), v.baseinfo.GetUid())
	}
}

//站起
func (r *PockerRoom) standup(uid string) bool {
	logger.Info("站起:", uid)
	if uid == "" {
		logger.Error("站起，参数uid为空")
		return false
	}

	p := r.GetPlayerByID(uid)
	if p == nil {
		logger.Error("Standup r.GetPlayerByID return nil, uid:%s", uid)
		return false
	}
	if p.islittle_blind() {
		r.smallblind = r.nextindex(p.posIndex)
	}

	if p.status == STATUS_THINKING {
		r.spkIndex = p.posIndex
	}
	p.status = STATUS_STAND
	// r.attends -= int32(1)
	r.foldsNum += int32(1)
	r.AtomicDesc()
	r.sync_s2c_status(p, ACT_STANDUP)

	r.potLeft += p.drops
	p.drops = int32(0)
	centerclient.SendCostResourceMsg(p.baseinfo.GetUid(), connector.RES_COIN, "pocker", p.chips)

	r.standPlayers = append(r.standPlayers, p)
	r.players[p.posIndex] = nil

	r.shows_all_players()

	if r.forceover() {
		r.force_over()
		return false
	}
	if r.nextindex(p.posIndex) == -1 {
		r.over_undelay(r.rounds)
		return false
	}
	// if r.spkIndex != p.posIndex {
	// 	//check gameover
	// 	if r.forceover() {
	// 		// r.forcewinner = p.baseinfo.GetUid()
	// 		r.force_over()
	// 		return false
	// 	}
	// 	if r.nextindex(p.posIndex) == -1 {
	// 		r.over_undelay(r.rounds)
	// 		return false
	// 	}
	// 	return false
	// }
	return true
}

func (r *PockerRoom) seat_empty_place(p *pockerman) bool {
	for i := 0; i < int(ROOM_PLAYERS_LIMIT); i++ {
		if r.players[i] == nil {
			r.players[i] = p
			p.posIndex = int32(i)
			p.rest_data()
			p.status = STATUS_WATTING_JOIN
			return true
		}
	}
	return false
}

//坐下
func (r *PockerRoom) seatdown(uid string) bool {
	logger.Info("坐下：", uid)
	if uid == "" {
		logger.Error("坐下，参数uid为空")
		return false
	}
	r.shows_all_players()
	p := r.get_stand_player(uid)
	if p == nil {
		logger.Error("seatdown r.get_stand_player return nil, uid:%s", uid)
		return false
	}

	if !r.seat_empty_place(p) {
		logger.Error("坐下出错，房间已满")
		return false
	}
	r.remove_stand_player(uid)
	r.shows_all_players()

	r.exchange(p)
	r.sync_s2c_status(p, ACT_SEATDOWN)

	// r.TellOthersIJoin(p)
	if len(r.get_attend_playeruids()) == 2 {
		r.overtime = int(time.Now().Unix())
	}
	return true
}

func (r *PockerRoom) Action(msg *rpc.C2SAction) {
	logger.Info("###########Action be called uid:%s, act:%d", msg.GetUid(), msg.GetAct())
	if msg.GetAct() == ACT_SEATDOWN {
		r.seatdown(msg.GetUid())
		return
	}
	if msg.GetAct() == ACT_LEAVE {
		r.leave(msg.GetUid())
		return
	} else if msg.GetAct() == ACT_CHANGE_DESK {
		r.change_desk(msg.GetUid())
		return
	}

	p := r.GetPlayerByID(msg.GetUid())
	if p == nil {
		p = r.get_stand_player(msg.GetUid())
		if p == nil {
			logger.Error("Action r.GetPlayerByID return nil:%s", msg.GetUid())
			return
		}
	}
	logger.Info("客户端:%s Action msg:%d 名字：%s", msg.GetUid(), msg.GetAct(), p.baseinfo.GetName())

	if msg.GetAct() != ACT_LEAVE && msg.GetAct() != ACT_CHANGE_DESK &&
		msg.GetAct() != ACT_SEATDOWN && msg.GetAct() != ACT_STANDUP {
		if p.status != STATUS_THINKING {
			logger.Error("Action 当前状态不为thinking，不能执行动作,当前状态：%d", p.status)
			return
		}
	}

	// if p.waitFrom == 0 {
	// 	logger.Error("Action 出错了，不该此玩家说话，：%s", p.baseinfo.GetName())
	// 	return
	// }

	switch int(msg.GetAct()) {
	// case ACT_LEAVE, ACT_CHANGE_DESK:
	// 	if !r.leave(msg.GetUid()) {
	// 		return
	// 	}
	case ACT_STANDUP:
		if !r.standup(msg.GetUid()) {
			return
		}
	// case ACT_SEATDOWN:
	// 	r.seatdown(msg.GetUid())
	// 	return
	// if !r.seatdown(msg.GetUid()) {
	// 	return
	// }
	case ACT_FOLD: //弃牌
		if !r.fold(msg) {
			return
		}

	case ACT_CHECK: //看牌
		if !r.check(msg) {
			return
		}

	case ACT_CALL: //跟注
		if !r.call(msg) {
			return
		}
	case ACT_RAISE: //加注
		if msg.GetRaise() <= 0 {
			logger.Error("加注金额不能<=0, value:%d", msg.GetRaise())
			return
		}
		if !r.raise(msg) {
			return
		}
	case ACT_ALLIN: //allin
		if !r.allin(msg) {
			return
		}

	default:
		logger.Error("Action(msg *rpc.C2SAction) act:%d error", msg.GetAct())
		return
	}

	act := msg.GetAct()
	if act == ACT_CHECK || act == ACT_CALL || act == ACT_RAISE || act == ACT_ALLIN {
		r.lastAct = int(msg.GetAct())
		if msg.GetRaise() > r.lastValue {
			r.lastValue = msg.GetRaise()
		}
		p.autofold = false
	}
	p.waitFrom = 0

	if r.nextindex(p.posIndex) == p.posIndex {
		r.over_undelay(r.rounds)
		return
	}

	// cnt := r.get_noraml_player_cnts()
	// nm_cnt := r.get_player_cnt_no_allin()

	// index := r.nextindex(p.posIndex)
	// logger.Info("**********Action, nm_cnt:%d, cnt:%d", nm_cnt, cnt)
	// if nm_cnt <= int32(1) && cnt > int32(1) && index == -1 {
	// 	r.cdtOver = 1
	// 	logger.Info("设置cdtOver true")
	// 	// r.over()
	// 	r.over_undelay()
	// 	return
	// }
	r.nextplayer()
}

func (r *PockerRoom) get_player_cnt_no_allin() int32 {
	cnt := int32(0)
	for _, v := range r.players {
		if v == nil {
			continue
		}
		if v.status == STATUS_RAISE || v.status == STATUS_CALL || v.status == STATUS_CHEDK {
			cnt++
		}
	}
	return cnt
}

// func (r *PockerRoom) all_speaked() bool {
// 	if r.rounds != int32(0) {
// 		return true
// 	}

// 	bp := r.players[r.bigblind]
// 	if bp == nil || bp.status == STATUS_CALL || bp.status == STATUS_CHEDK || bp.status == STATUS_FOLD ||
// 		bp.status == STATUS_ALLIN || bp.status == STATUS_RAISE || bp.status == STATUS_STAND {
// 		return true
// 	}
// 	return false
// }

func (r *PockerRoom) delay_call_over() {
	r.ft[2] = time.AfterFunc(time.Duration(1)*time.Second, func() {
		if r.ft[2] != nil {
			r.ft[2].Stop()
			r.ft[2] = nil
		}
		r.over()
	})
}

func (r *PockerRoom) nextplayer() {
	logger.Info("下一个玩家说话")
	if r.spkIndex >= int32(len(r.players)) {
		logger.Error("nextplayer r.spkIndex:%d >= len(r.players):%d ", r.spkIndex, len(r.players))
		return
	}

	index := r.spkIndex
	if index == int32(-1) {
		index = r.bigblind
	}
	if r.players[index] != nil {
		logger.Info("old spkindex:%d, uid:%s, len:%d, 名字:%s", index,
			r.players[index].baseinfo.GetUid(), len(r.players), r.players[index].baseinfo.GetName())
	}

	for i := 0; i < len(r.players); i++ {
		index++
		if index >= int32(len(r.players)) {
			index = 0
		}

		logger.Info("###################测试大盲, r.rounds:%d, r.bigblind:%d, r.raiseIdx:%d, r.spkIndex:%d, index:%d",
			r.rounds, r.bigblind, r.raiseIdx, r.spkIndex, index)
		if r.rounds == int32(1) && r.bigblind == r.raiseIdx { //首轮，要大盲注说过话，才能结束
			if r.spkIndex == r.bigblind || (index == r.bigblind && r.players[r.bigblind] == nil) {
				logger.Info("###########首轮调用")
				r.delay_call_over()
				break
			}
		} else if index == r.raiseIdx { //一轮完成
			logger.Info("###########其它轮调用")
			r.delay_call_over()
			break
		}
		if r.players[index] == nil {
			continue
		}

		status := r.players[index].status
		if status == STATUS_FOLD || status == STATUS_ALLIN || status == STATUS_STAND || status == STATUS_WATTING_JOIN {
			continue
		}
		r.spkIndex = int32(index)
		r.players[r.spkIndex].waitFrom = int32(time.Now().Unix())
		r.players[r.spkIndex].status = STATUS_THINKING
		logger.Info("new spkIndex:%d, nextplayer :%s,索引：%s, 名字：%s", index, r.players[r.spkIndex].baseinfo.GetUid(), r.players[r.spkIndex].baseinfo.GetName())

		r.sync_s2c_status(r.players[r.spkIndex], ACT_COUNTDOWN)
		break
	}
}

func (r *PockerRoom) isover() bool {
	if r.rounds >= int32(ROUND_CNT) {
		return true
	}

	return false
	// return r.forceover()
}

func (r *PockerRoom) forceover() bool {
	if r.attends-r.foldsNum <= 1 {
		return true
	}
	return false
}

// func (r *PockerRoom) allpocker_cdt() bool {
// 	if r.cdtOver == 1 && r.rounds <= int32(ROUND_CNT) {
// 		return true
// 	}

// 	logger.Info("allpocker_cdt r.rounds:%d,ROUND_CNT:%d ", r.rounds, ROUND_CNT)
// 	logger.Info("allpocker_cdt cdtOver:%d", r.cdtOver)

// 	// num := 0
// 	// for _, v := range r.players {
// 	// 	if v.status == STATUS_RAISE || v.status == STATUS_CALL || v.status == STATUS_CHEDK {
// 	// 		num++
// 	// 	}
// 	// }
// 	// if num == 1 && r.allinsNum == int32(1) && r.rounds < int32(ROUND_CNT) {
// 	// 	r.cdtOver = true
// 	// 	return true
// 	// }

// 	// if num == 0 && r.allinsNum > int32(1) && r.rounds < int32(ROUND_CNT) {
// 	// 	r.cdtOver = true
// 	// 	return true
// 	// }
// 	return false
// }

// func (r *PockerRoom) over_undelay() {
// 	if r.attends-r.foldsNum <= 0 {
// 		return
// 	}

// 	//若本轮allin过，则需要分分彩池
// 	if r.allined {
// 		r.calcpools()
// 		r.allined = false
// 	}

// 	// if !r.allpocker_cdt() {
// 	if r.cdtOver != 1 { //直接结束，不用收筹码
// 		r.calcPots()
// 	}

// 	//game over
// 	if r.isover() {
// 		r.gameover()
// 		r.overtime = int(time.Now().Unix())
// 		return
// 	}
// 	r.givePocker()
// 	r.over_undelay()
// }

//just happen in fold
//1.direct battleend if r.attends - r.fold <= 1 (after fold call)
func (r *PockerRoom) force_over() {
	logger.Info("**************force_over called")
	// winner := ""
	// sum := r.potLeft
	// for _, v := range r.players {
	// 	if v == nil {
	// 		continue
	// 	}
	// 	if v.status == STATUS_RAISE || v.status == STATUS_CALL || v.status == STATUS_CHEDK || v.status == STATUS_ALLIN {
	// 		winner = v.baseinfo.GetUid()
	// 	}

	// 	sum += v.drops
	// 	v.drops = int32(0)
	// }
	// for _, v := range r.pots {
	// 	sum += v
	// }
	winner := ""
	for _, v := range r.players {
		if v != nil {
			v.waitFrom = 0
			// if r.allined {
			r.potLeft += v.drops
			v.drops = int32(0)
			// }
			// if v.status == STATUS_RAISE || v.status == STATUS_CALL ||
			// 	v.status == STATUS_CHEDK || v.status == STATUS_ALLIN || v.status == STATUS_READY {
			// 	winner = v.baseinfo.GetUid()
			// }
			if v.status != STATUS_FOLD {
				winner = v.baseinfo.GetUid()
			}
			if v.status == STATUS_THINKING {
				v.status = STATUS_WATTING_JOIN
			}
		}
	}

	r.ft[0] = time.AfterFunc(time.Duration(1)*time.Second, func() {
		if r.ft[0] != nil {
			r.ft[0].Stop()
			r.ft[0] = nil
		}

		msg := &rpc.S2CAction{}
		msg.SetAct(int32(ACT_ROUND_OVER))
		msg.Pots = append(msg.Pots, r.potLeft)
		r.SendMsg2Others(msg, "S2CAction")

		// r.calcPots()

		// if len(r.candi) != 1 || len(r.candi[0]) != 1 || len(r.pots) != 1 {
		// 	logger.Error("force_over error r.candi:%v, r.pots:%v", r.candi, r.pots)
		// 	r.overtime = int(time.Now().Unix())
		// 	return
		// }
		// winner := r.candi[0][0].baseinfo.GetUid()

		p := r.GetPlayerByID(winner)
		if p == nil {
			logger.Error("force_over r.GetPlayerByID(:%s) return nil", winner)
			r.overtime = int(time.Now().Unix())
			return
		}

		p.chips += r.potLeft
		r.pots = append(r.pots, r.potLeft)
		// p.chips += r.pots[0]
		// if r.pots[0] == 0 {
		// 	p.chips += r.potLeft
		// }

		logger.Info("强制结束：p.chips:%d, r.potLeft:%d, r.pots[0]:%v, winner:%v", p.chips, r.potLeft, r.pots, winner)

		r.ft[0] = time.AfterFunc(time.Duration(2)*time.Second, func() {
			if r.ft[0] != nil {
				r.ft[0].Stop()
				r.ft[0] = nil
			}

			msg := &rpc.S2CAction{}
			mbase := &rpc.PockerManBase{}
			mbase.SetUid(p.baseinfo.GetUid())
			mbase.SetStatus(int32(p.status))
			mbase.SetCoin(p.chips)
			msg.Infos = append(msg.Infos, mbase)

			msg.SetAct(int32(ACT_GAMEOVER))
			msg.Pots = append(msg.Pots, r.pots[0])
			msg.Winners = append(msg.Winners, winner)

			r.SendMsg2Others(msg, "S2CAction")
			r.auto_exchange_chips()
			r.reset_room()
			r.overtime = int(time.Now().Unix())
		})
	})
}

//
//2.give all left, then battleend 		 if nextindex(posIndex) == -1 || nextindex(posIndex) = posIndex (after acton call, include fold(positive, negative))
func (r *PockerRoom) over_undelay(round int32) {
	r.ft[0] = time.AfterFunc(time.Duration(1)*time.Second, func() {
		if r.ft[0] != nil {
			r.ft[0].Stop()
			r.ft[0] = nil
		}

		r.cdtOver = true
		if r.allined {
			r.calcpools()
			r.allined = false
		}
		r.ft[0] = time.AfterFunc(time.Duration(2)*time.Second, func() {
			if r.ft[0] != nil {
				r.ft[0].Stop()
				r.ft[0] = nil
			}

			for i := r.rounds; i < ROUND_CNT; i++ {
				r.givePocker()
			}

			r.ft[0] = time.AfterFunc(time.Duration(1)*time.Second, func() {
				if r.ft[0] != nil {
					r.ft[0].Stop()
					r.ft[0] = nil
				}
				r.gameover()
				r.overtime = int(time.Now().Unix())
				return
			})
		})
	})
	return
}

func (r *PockerRoom) over() {
	logger.Info("over called")
	r.ft[0] = time.AfterFunc(time.Duration(1)*time.Second, func() {
		if r.ft[0] != nil {
			r.ft[0].Stop()
			r.ft[0] = nil
		}
		if r.attends-r.foldsNum <= 0 {
			return
		}
		if r.allined {
			r.calcpools()
		} else {
			r.calcPots()
		}

		//game over
		if r.rounds >= ROUND_CNT { //round == 4
			r.ft[0] = time.AfterFunc(time.Duration(2)*time.Second, func() {
				if r.ft[0] != nil {
					r.ft[0].Stop()
					r.ft[0] = nil
				}
				r.gameover()
				r.overtime = int(time.Now().Unix())
			})
			return
		}

		r.ft[1] = time.AfterFunc(time.Duration(1)*time.Second, func() {
			if r.ft[1] != nil {
				r.ft[1].Stop()
				r.ft[1] = nil
			}
			r.givePocker()

			r.ft[3] = time.AfterFunc(time.Duration(1)*time.Second, func() {
				if r.ft[3] != nil {
					r.ft[3].Stop()
					r.ft[3] = nil
				}
				r.nextround()
			})
			r.allined = false
		})
	})
}

// func (r *PockerRoom) over() {
// 	logger.Info("over called")
// 	if r.attends-r.foldsNum <= 0 {
// 		return
// 	}

// 	//若本轮allin过，则需要分分彩池
// 	if r.allined {
// 		r.calcpools()
// 	}

// 	// if !r.allpocker_cdt() {
// 	if r.cdtOver != 1 {
// 		r.calcPots()
// 	}

// 	//game over
// 	if r.isover() {
// 		r.ft[0] = time.AfterFunc(time.Duration(2)*time.Second, func() {
// 			if r.ft[0] != nil {
// 				r.ft[0].Stop()
// 				r.ft[0] = nil
// 			}
// 			r.gameover()
// 			r.overtime = int(time.Now().Unix())
// 		})
// 		return
// 	}

// 	r.ft[1] = time.AfterFunc(time.Duration(1)*time.Second, func() {
// 		if r.ft[1] != nil {
// 			r.ft[1].Stop()
// 			r.ft[1] = nil
// 		}
// 		r.givePocker()

// 		r.ft[3] = time.AfterFunc(time.Duration(1)*time.Second, func() {
// 			if r.ft[3] != nil {
// 				r.ft[3].Stop()
// 				r.ft[3] = nil
// 			}
// 			r.nextround()
// 		})
// 		r.allined = false
// 	})
// }

// func (r *PockerRoom) sync_round_over_info() {
// 	msg := &rpc.S2CAction{}
// 	msg.SetAct(int32(ACT_ROUND_OVER))
// 	msg.Pots = r.pots
// 	r.SendMsg2Others(msg, "S2CAction")
// 	logger.Info("sync_round_over_info pots:%v", msg.Pots)
// }

//分彩池
func (r *PockerRoom) calcpools() {
	logger.Info("开始分彩池 calcpools...")
	//按照投入筹码数量升序排
	tmpPlayers := []*pockerman{}
	for _, v := range r.players {
		if v == nil || v.drops <= 0 {
			if v != nil && v.drops <= 0 {
				logger.Error("*********v.drops:%d, name:%s", v.drops, v.baseinfo.GetName())
			}
			continue
		}
		tmpPlayers = append(tmpPlayers, v)
	}
	for i := 0; i < len(tmpPlayers); i++ {
		for j := i + 1; j < len(tmpPlayers); j++ {
			if tmpPlayers[i] == nil || tmpPlayers[j] == nil {
				logger.Error("calcpools tmpPlayers[i] == nil || tmpPlayers[j] == nil, i:%d, j:%d", i, j)
				return
			}
			if tmpPlayers[i].drops > tmpPlayers[j].drops {
				tmp := tmpPlayers[i]
				tmpPlayers[i] = tmpPlayers[j]
				tmpPlayers[j] = tmp
			}
		}
	}

	logger.Info("根据drpos排序后的结果")
	for _, v := range tmpPlayers {
		if v == nil {
			continue
		}
		logger.Info("玩家:%s 下注金额:%d,剩余筹码:%d", v.baseinfo.GetName(), v.drops, v.chips)
	}

	folds := int32(0)
	bLoop := true
	for bLoop {
		bLoop = false
		for k, v := range tmpPlayers {
			if v.status == STATUS_STAND || v.status == STATUS_FOLD {
				folds += v.drops
				v.drops = int32(0)
				tmpPlayers = append(tmpPlayers[:k], tmpPlayers[k+1:]...)
				bLoop = true
				break
			}
		}
	}

	logger.Info("去掉 stand fold")
	for _, v := range tmpPlayers {
		if v == nil {
			continue
		}
		logger.Info("玩家:%s 下注金额:%d,剩余筹码:%d", v.baseinfo.GetName(), v.drops, v.chips)
	}

	pools := []int32{}
	candis := [][]*pockerman{}
	for i := 0; i < len(tmpPlayers); i++ {
		if tmpPlayers[i].drops <= 0 {
			continue
		}

		if i == 0 {
			pool := folds + r.potLeft + tmpPlayers[0].drops*int32(len(tmpPlayers))
			logger.Info("第0个彩池:%d folds:%d + r.potLeft:%d + tmpPlayers[0].drops:%d * len:%d",
				pool, folds, r.potLeft, tmpPlayers[0].drops, int32(len(tmpPlayers)))
			r.potLeft = int32(0)
			pools = append(pools, pool)
			mans := []*pockerman{}

			tmpValue := tmpPlayers[0].drops
			for _, v := range tmpPlayers {
				v.drops -= tmpValue
				mans = append(mans, v)
			}
			candis = append(candis, mans)
			logger.Info("候选人")
			for _, v := range mans {
				if v == nil {
					continue
				}
				logger.Info("名字:%d, drops:%d", v.baseinfo.GetName(), v.drops)
			}
			continue
		}

		pool := int32(0)
		mans := []*pockerman{}
		logger.Info("第:%d个彩池候选人, 当前drops:%d", i, tmpPlayers[i].drops)
		tmpValue := tmpPlayers[i].drops
		for j := i; j < len(tmpPlayers); j++ {
			pool += tmpPlayers[i].drops
			tmpPlayers[j].drops -= tmpValue
			logger.Info("名字:%d drpos:%d", tmpPlayers[j].baseinfo.GetName(), tmpPlayers[j].drops)
			mans = append(mans, tmpPlayers[j])
		}
		logger.Info("第:%d个彩池候选人, 彩池:%d", i, pool)
		if pool > int32(0) {
			pools = append(pools, pool)
			candis = append(candis, mans)
		}
	}

	if r.isover() || r.cdtOver {
		r.pots = append(r.pots, pools...)
		r.candi = append(r.candi, candis...)

		msg := &rpc.S2CAction{}
		msg.SetAct(int32(ACT_ROUND_OVER))
		msg.Pots = r.pots
		r.SendMsg2Others(msg, "S2CAction")

		logger.Info("最终pots:%s, potleft:%d", r.pots, r.potLeft)
		for i, v := range r.candi {
			logger.Info("********index:%d, 候选人", i)
			for _, m := range v {
				logger.Info("*******:%s", m.baseinfo.GetName())
			}
		}
		return
	}
	for _, v := range r.players {
		if v == nil {
			continue
		}
		v.drops = int32(0)
	}
	logger.Info("========分彩池， pots:%v, candi:%d", pools, len(candis))

	for k, v := range candis {
		allin := false
		for _, m := range v {
			if m.status == STATUS_ALLIN {
				allin = true
				break
			}
		}
		if allin {
			r.potLeft += pools[k]
			r.pots = append(r.pots, pools[k])
			r.candi = append(r.candi, candis[k])
			bLoop = true
			break
		}
	}
	msg := &rpc.S2CAction{}
	msg.SetAct(int32(ACT_ROUND_OVER))
	msg.Pots = r.pots
	r.SendMsg2Others(msg, "S2CAction")
	// if len(pools) >= 2 {
	// 	r.pots = append(r.pots, pools[:len(pools)-1]...)
	// 	r.candi = append(r.candi, candis[:len(pools)-1]...)
	// 	r.potLeft = pools[len(pools)-1]
	// 	msg.Pots = r.pots
	// } else {
	// 	r.potLeft = int32(0)
	// 	r.pots = pools
	// 	r.candi = candis
	// 	msg.Pots = r.pots
	// }

	logger.Info("分彩池pots:%s, potleft:%d, pools:%s", r.pots, r.potLeft, pools)
	for i, v := range r.candi {
		logger.Info("********index:%d, 候选人", i)
		for _, m := range v {
			logger.Info("*******:%s", m.baseinfo.GetName())
		}
	}

	r.SendMsg2Others(msg, "S2CAction")
	// logger.Info("最终pots:%s, candi:%s, potleft:%d", r.pots, r.candi, r.potLeft)
}

func (r *PockerRoom) show_pocker() {

	for _, v := range r.players {
		if v == nil {
			continue
		}
		logger.Info("====================玩家身上的牌==================:%s", v.baseinfo.GetName())
		for _, m := range v.pockers {
			logger.Info("花色:%d, 数字:%d", m.eType, m.num)
		}
	}
	logger.Info("\n====================公共牌==================")
	for _, m := range r.pockers {
		logger.Info("花色:%d, 数字:%d", m.eType, m.num)
	}

}

//发牌
func (r *PockerRoom) givePocker() {
	logger.Info("开始发牌咯 r.rounds:%d", r.rounds)
	defer logger.Info("发牌结束")
	if r.rounds == ROUND_CNT {
		logger.Error("givePocker 已经结束了，还发啥子牌呢")
		return
	}

	if r.rounds == 0 { //给每个玩家发2张牌
		logger.Info("“发2张牌")
		cnt := 0
		for _, v := range r.players {
			if v == nil {
				continue
			}
			if v.status == STATUS_STAND || v.status == STATUS_FOLD || v.status == STATUS_WATTING_JOIN {
				continue
			}

			msg := &rpc.S2CAction{}
			msg.SetOperater(v.baseinfo.GetUid())
			msg.SetAct(int32(ACT_GIVE_POCKER))

			for i := cnt * 2; i < cnt*2+2; i++ {
				pk := &rpc.Pocker{}
				pk.SetEType(int32(r.leftpockers[i].eType))
				pk.SetNum(r.leftpockers[i].num)
				msg.Pockers = append(msg.Pockers, pk)
				logger.Info("花：%d, 数字:%d", r.leftpockers[i].eType, r.leftpockers[i].num)
			}
			uids := []string{}
			uids = append(uids, v.baseinfo.GetUid())
			centerclient.SendCommonNotify2S(uids, msg, "S2CAction")
			v.pockers = append(v.pockers, r.leftpockers[cnt*2:cnt*2+2]...)
			cnt++
		}
		r.give_stand_pockers()
		r.leftpockers = append(r.leftpockers[:0], r.leftpockers[cnt*2+2:]...)
	} else if r.rounds == 1 { //发3张
		logger.Info("发3张公共牌")
		for _, v := range r.players {
			if v == nil {
				continue
			}
			if v.status == STATUS_STAND || v.status == STATUS_FOLD || v.status == STATUS_WATTING_JOIN {
				continue
			}

			msg := &rpc.S2CAction{}
			msg.SetOperater(v.baseinfo.GetUid())
			msg.SetAct(int32(ACT_SHOW_POCKER))
			// msg.Pockers = append(msg.Pockers, r.leftpockers[:3]...)
			for n, e := range r.leftpockers {
				if n >= 3 {
					break
				}
				pk := &rpc.Pocker{}
				pk.SetEType(int32(e.eType))
				pk.SetNum(e.num)
				msg.Pockers = append(msg.Pockers, pk)
			}
			v.pockers = append(v.pockers, r.leftpockers[:3]...)

			v.rest_combine()
			v.CalcCombine()
			// v.Show()
			v.Sort()
			_, v2 := v.GetMaxValue()
			// v.combineNum = int32(v2)
			msg.SetCombineNum(int32(v2))

			uids := []string{}
			uids = append(uids, v.baseinfo.GetUid())
			centerclient.SendCommonNotify2S(uids, msg, "S2CAction")
		}
		r.give_stand_pockers()
		r.pockers = append(r.pockers, r.leftpockers[:3]...)
		r.leftpockers = append(r.leftpockers[:0], r.leftpockers[3:]...)
	} else {
		logger.Info("发第4、5张公共牌")
		for _, v := range r.players {
			if v == nil {
				continue
			}
			if v.status == STATUS_STAND || v.status == STATUS_FOLD || v.status == STATUS_WATTING_JOIN {
				continue
			}
			msg := &rpc.S2CAction{}
			msg.SetOperater(v.baseinfo.GetUid())
			msg.SetAct(int32(ACT_SHOW_POCKER))
			// msg.Pockers = append(msg.Pockers, r.leftpockers[:1]...)
			for n, e := range r.leftpockers {
				if n >= 1 {
					break
				}
				pk := &rpc.Pocker{}
				pk.SetEType(int32(e.eType))
				pk.SetNum(e.num)
				msg.Pockers = append(msg.Pockers, pk)
			}
			v.pockers = append(v.pockers, r.leftpockers[:1]...)

			v.rest_combine()
			v.CalcCombine()
			// v.Show()
			v.Sort()
			_, v2 := v.GetMaxValue()
			// v.combineNum = int32(v2)
			msg.SetCombineNum(int32(v2))

			uids := []string{}
			uids = append(uids, v.baseinfo.GetUid())
			centerclient.SendCommonNotify2S(uids, msg, "S2CAction")
		}
		r.give_stand_pockers()
		r.pockers = append(r.pockers, r.leftpockers[:1]...)
		r.leftpockers = append(r.leftpockers[:0], r.leftpockers[1:]...)
	}
	r.rounds += 1
	// r.show_pocker()
}

func (r *PockerRoom) give_stand_pockers() {
	uids := []string{}
	for _, v := range r.players {
		if v == nil {
			continue
		}
		if v.status == STATUS_FOLD || v.status == STATUS_WATTING_JOIN {
			uids = append(uids, v.baseinfo.GetUid())
		}
	}
	for _, v := range r.standPlayers {
		if v == nil {
			continue
		}
		uids = append(uids, v.baseinfo.GetUid())
	}
	if len(uids) == 0 {
		return
	}

	msg := &rpc.S2CAction{}
	// msg.SetOperater(v.baseinfo.GetUid())
	msg.SetAct(int32(ACT_SHOW_POCKER))
	if r.rounds == 0 {
		msg.SetAct(int32(ACT_GIVE_POCKER))
	}

	if r.rounds == 1 {
		for n, e := range r.leftpockers {
			if n >= 3 {
				break
			}
			pk := &rpc.Pocker{}
			pk.SetEType(int32(e.eType))
			pk.SetNum(e.num)
			msg.Pockers = append(msg.Pockers, pk)
		}
	} else if r.rounds > 1 {
		for n, e := range r.leftpockers {
			if n >= 1 {
				break
			}
			pk := &rpc.Pocker{}
			pk.SetEType(int32(e.eType))
			pk.SetNum(e.num)
			msg.Pockers = append(msg.Pockers, pk)
		}
	}
	// uids = append(uids, v.baseinfo.GetUid())
	centerclient.SendCommonNotify2S(uids, msg, "S2CAction")

}

func (r *PockerRoom) calcPots() {
	logger.Info("calcPots 无人allin 收桌上筹码")
	//没人allin，单独算彩池
	if !r.allined {
		roundpots := int32(0)
		mans := []*pockerman{}
		for _, v := range r.players {
			if v == nil {
				if v != nil {
					logger.Info("calcPots v.drops :%d, uid:%s", v.drops, v.baseinfo.GetUid())
				}
				continue
			}
			logger.Info("calcPots uid:%s, status:%d", v.baseinfo.GetUid(), v.status)
			if (r.isover() || r.forceover()) && v.status != STATUS_STAND && v.status != STATUS_FOLD && v.status != STATUS_WATTING_JOIN {
				mans = append(mans, v)
			}
			roundpots += v.drops
			v.drops = int32(0)
		}

		r.potLeft += roundpots
		if (r.isover() || r.forceover()) && len(mans) == 0 {
			logger.Info("*******len(man):%d", len(mans))
			return
		}

		if r.isover() || r.forceover() {
			r.pots = append(r.pots, r.potLeft)
			r.candi = append(r.candi, mans) //参与比牌的人
		}

		if roundpots <= int32(0) {
			logger.Info("calcPots roundpots:%d", roundpots)
			return
		}
	}

	if len(r.pots) != len(r.candi) {
		logger.Error("calcPots len(r.pots) != len(r.candi) , %d, %v", len(r.pots), len(r.candi))
		return
	}
	logger.Info("*********pots:%v, r.potLeft:%d", r.pots, r.potLeft)
	logger.Info("*********candi:%v", r.candi)

	msg := &rpc.S2CAction{}
	msg.SetAct(int32(ACT_ROUND_OVER))
	msg.Pots = r.pots
	if r.potLeft >= 0 {
		msg.Pots = append(msg.Pots, r.potLeft)
	}
	r.SendMsg2Others(msg, "S2CAction")
	logger.Info("sync_round_over_info pots:%v", msg.Pots)
}

//游戏结束
func (r *PockerRoom) gameover() {
	logger.Info("gameover called")
	for _, v := range r.players {
		if v != nil {
			v.waitFrom = 0
		}
	}

	//计算赢得彩池的玩家
	wins := [][]*pockerman{}
	for k, _ := range r.pots {
		maxPocker := int(0)
		for i := 0; i < len(r.candi[k]); i++ {
			if r.candi[k][i].combineNum >= maxPocker {
				maxPocker = r.candi[k][i].combineNum
			}
		}

		logger.Info("******gameover,  maxPocker:%d", maxPocker)
		win := []*pockerman{}
		for _, v := range r.candi[k] {
			if v.combineNum == maxPocker {
				logger.Info("********win:%s, num:%d", v.baseinfo.GetUid(), v.combineNum)
				win = append(win, v)
			}
		}
		wins = append(wins, win)
	}

	logger.Info("\n******####################r.pots:%v", r.pots)
	logger.Info("******####################r.wins:%v", wins)

	//给获取玩家奖励
	uids := []string{}
	for k, v := range r.pots {
		logger.Info("********发奖励, k:%d, v:%d", k, v)
		uid := ""
		num := v / int32(len(wins[k]))
		for _, e := range wins[k] {
			uid += e.baseinfo.GetUid()
			uid += "_"
			e.chips += int32(num)
		}
		uid = strings.Trim(uid, "_")
		uids = append(uids, uid)
	}

	msg := &rpc.S2CAction{}
	msg.SetAct(int32(ACT_GAMEOVER))
	msg.Pots = r.pots
	msg.Winners = uids
	for _, v := range wins {
		for _, w := range v {
			mbase := &rpc.PockerManBase{}
			mbase.SetUid(w.baseinfo.GetUid())
			mbase.SetStatus(int32(w.status))
			mbase.SetCoin(w.chips)
			msg.Infos = append(msg.Infos, mbase)
		}
	}

	if !r.forceover() {
		for _, v := range r.players {
			if v == nil {
				continue
			}

			if v.status == STATUS_STAND || v.status == STATUS_FOLD || v.status == STATUS_WATTING_JOIN {
				continue
			}

			attMsg := &rpc.ComparePokerPlayer{}
			attMsg.SetUid(v.baseinfo.GetUid())
			for _, m := range v.pockers {
				pk := &rpc.Pocker{}
				pk.SetEType(int32(m.eType))
				pk.SetNum(m.num)
				attMsg.Pockers = append(attMsg.Pockers, pk)
			}

			logger.Info("v.bestIndex:%d, len(v.combines):%d", v.bestIndex, len(v.combines))
			// for _, v := range v.combines {
			// 	logger.Info("*********v:%v", v)
			// }

			if v.bestIndex >= len(v.combines) {
				logger.Error("v.combines[v.bestIndex] == nil, bestIndex:%d, v.bestPocker:%d", v.bestIndex, v.bestPocker)
				continue
			}

			for _, m := range v.combines[v.bestIndex] {
				pk := &rpc.Pocker{}
				pk.SetEType(int32(m.eType))
				pk.SetNum(m.num)
				attMsg.CombinePockers = append(attMsg.CombinePockers, pk)
			}
			attMsg.SetCombineNum(int32(v.bestPocker))
			msg.ComparePlayers = append(msg.ComparePlayers, attMsg)
		}
	}

	r.show_pocker()
	r.SendMsg2Others(msg, "S2CAction")
	r.auto_exchange_chips()

	r.playWin = true
	if r.get_noraml_player_cnts() < 2 {
		r.playWin = false
	}
	r.after_gameover()
}

func (r *PockerRoom) after_gameover() {
	//task
	task := &proto.CallCnserverMsg{}
	task.Param1 = "PockerEnd"
	task.Uids = r.get_attend_playeruids()
	centerclient.CallCnserverFunc(task)
}

//回收扑克
func (r *PockerRoom) recycling_pocker() {
	for _, v := range r.players {
		if v == nil {
			continue
		}
		if v.status == STATUS_FOLD || v.status == STATUS_STAND {
			continue
		}
		v.status = STATUS_READY
		r.leftpockers = append(r.leftpockers, v.pockers[:2]...)
	}
	r.leftpockers = append(r.leftpockers, r.pockers...)
}

func (r *PockerRoom) auto_exchange_chips() {
	for _, v := range r.players {
		if v == nil {
			continue
		}

		if v.chips > r.bigValue {
			continue
		}
		r.exchange(v)
	}
}

func (r *PockerRoom) exchange(p *pockerman) {
	if p == nil {
		logger.Error("exchange param p is nil")
		return
	}

	//通知客户端退出
	cgValue := int32(0)
	if r.customInfo == nil {
		cfg := common.GetDaerRoomConfig(strconv.Itoa(int(r.eType)))
		if cfg == nil {
			logger.Error("auto_exchange_chips common.GetDaerRoomConfig return nil,roomType:%d", r.eType)
			return
		}
		cgValue = cfg.PockerExchange
	} else {
		cgValue = r.customInfo.exchngeValue
	}

	if p.baseinfo.GetCoin() < cgValue {
		msg := &rpc.S2CAction{}
		msg.SetOperater(p.baseinfo.GetUid())
		msg.SetAct(ACT_LEAVE)
		r.SendMsg2Others(msg, "S2CAction")

		logger.Info("兑换筹码 余额不够，强退:%d", p.chips)
		return
	}
	p.baseinfo.SetCoin(p.baseinfo.GetCoin() - cgValue)
	p.chips += cgValue
	centerclient.SendCostResourceMsg(p.baseinfo.GetUid(), connector.RES_COIN, "pocker", -cgValue)

	r.sync_s2c_status(p, ACT_EXCHANGE_CHIPS)
	logger.Info("兑换筹码:%d, etype:%d", p.chips, r.eType)
}

//下一轮
func (r *PockerRoom) nextround() {
	logger.Info("nextround called")

	// if r.allpocker_cdt() || r.forceover() {
	// if r.forceover() {
	// 	r.over()
	// 	return
	// } else {
	// 	logger.Error("allpocker_cdt return false, rounds:%d", r.rounds)
	// }

	if r.spkIndex >= int32(len(r.players)) {
		logger.Error("nextround r.spkIndex:%d >= len(r.players):%d ", r.spkIndex, len(r.players))
		return
	}
	r.lastValue = 0
	r.lastAct = 0

	index := r.dIndex
	for i := 0; i < len(r.players); i++ {
		index++
		if index >= int32(len(r.players)) {
			index = 0
		}

		if r.players[index] == nil {
			continue
		}
		status := r.players[index].status

		if status == STATUS_FOLD || status == STATUS_ALLIN || status == STATUS_STAND || status == STATUS_WATTING_JOIN {
			continue
		}
		r.spkIndex = int32(index)
		r.players[r.spkIndex].waitFrom = int32(time.Now().Unix())
		r.players[r.spkIndex].status = STATUS_THINKING

		logger.Info("nextround 说话玩家:%s, 索引：%d 名字：%s", r.players[r.spkIndex].baseinfo.GetUid(),
			r.players[r.spkIndex].posIndex, r.players[r.spkIndex].baseinfo.GetName())

		r.sync_s2c_status(r.players[r.spkIndex], ACT_COUNTDOWN)
		return
	}
	// if r.isover() {
	// 	r.over()
	// }
}

func (r *PockerRoom) fold(msg *rpc.C2SAction) bool {
	p := r.GetPlayerByID(msg.GetUid())
	if p.status == STATUS_FOLD || p.status == STATUS_STAND || p.status == STATUS_WATTING_JOIN {
		logger.Error("Action status can't fold:%d", p.status)
		return false
	}
	// r.leftpockers = append(r.leftpockers, p.pockers[:2]...)
	if p.islittle_blind() {
		r.smallblind = r.nextindex(p.posIndex)
	}

	p.status = STATUS_FOLD
	// r.attends -= int32(1)
	r.foldsNum += int32(1)
	p.autofold = false

	r.sync_status(msg)
	r.spkIndex = p.posIndex

	//check gameover
	if r.forceover() {
		// r.forcewinner = p.baseinfo.GetUid()
		r.force_over()
		return false
	}
	if r.nextindex(p.posIndex) == -1 {
		r.over_undelay(r.rounds)
		return false
	}
	return true
}

func (r *PockerRoom) check(msg *rpc.C2SAction) bool {
	p := r.GetPlayerByID(msg.GetUid())
	if p.status == STATUS_FOLD || p.status == STATUS_STAND || p.status == STATUS_WATTING_JOIN {
		logger.Error("Action status can't check:%d", p.status)
		return false
	}
	if !p.islittle_blind() && r.lastAct != ACT_CHECK {
		logger.Error("看牌失败，非小盲注，或 r.lastAct != act， last:%d", r.lastAct)
		return false
	}
	p.status = STATUS_CHEDK
	if p.islittle_blind() {
		r.raiseIdx = p.posIndex
	}

	r.sync_status(msg)
	r.spkIndex = p.posIndex
	return true
}

func (r *PockerRoom) call(msg *rpc.C2SAction) bool {
	p := r.GetPlayerByID(msg.GetUid())
	if p.status == STATUS_FOLD || p.status == STATUS_STAND || p.status == STATUS_WATTING_JOIN {
		logger.Error("Action status can't call:%d", p.status)
		return false
	}
	// if p.islittle_blind() {
	// 	logger.Error("小盲不能跟注")
	// 	return false
	// }

	if r.lastAct != ACT_CALL && r.lastAct != ACT_RAISE && r.lastAct != ACT_ALLIN {
		logger.Error("最后1次执行动作为：%d，不能跟注", r.lastAct)
		return false
	}
	if !p.call() {
		return false
	}
	if p.status == STATUS_ALLIN {
		msg.SetAct(ACT_ALLIN)
	}
	msg.SetRaise(r.lastValue)

	r.sync_status(msg)
	r.spkIndex = p.posIndex
	return true
}

func (r *PockerRoom) raise(msg *rpc.C2SAction) bool {
	p := r.GetPlayerByID(msg.GetUid())
	if p.status == STATUS_FOLD || p.status == STATUS_STAND || p.status == STATUS_WATTING_JOIN {
		logger.Error("Action status can't raise:%d", p.status)
		return false
	}

	// if msg.GetRaise() -

	if !p.raise(msg.GetRaise()) {
		return false
	}

	if p.status == STATUS_ALLIN {
		msg.SetAct(ACT_ALLIN)
	}

	r.sync_status(msg)
	r.spkIndex = p.posIndex
	r.raiseIdx = p.posIndex
	return true
}

func (r *PockerRoom) allin(msg *rpc.C2SAction) bool {
	p := r.GetPlayerByID(msg.GetUid())
	if p.status == STATUS_FOLD || p.status == STATUS_STAND || p.status == STATUS_WATTING_JOIN {
		logger.Error("Action status can't allin:%d", p.status)
		return false
	}

	msg.SetRaise(p.chips)
	if !p.allin() {
		return false
	}
	// r.attends -= int32(1)
	r.allinsNum += int32(1)
	r.allined = true

	r.sync_status(msg)
	r.spkIndex = p.posIndex

	if p.islittle_blind() {
		r.smallblind = r.nextindex(p.posIndex)
	}

	if r.nextindex(p.posIndex) == -1 {
		r.over_undelay(r.rounds)
		return false
	}

	if r.raiseIdx != p.posIndex {
		// total := r.get_noraml_player_cnts()
		noallin := r.get_player_cnt_no_allin()
		if noallin == int32(1) {
			r.over_undelay(r.rounds)
			return false
		}
	}

	return true
}

func (r *PockerRoom) sync_s2c_status(p *pockerman, act int) {
	if p == nil {
		logger.Error("sync_s2c_status param p is nil")
		return
	}

	msg := &rpc.S2CAction{}
	msg.SetOperater(p.baseinfo.GetUid())
	msg.SetAct(int32(act))

	mbase := &rpc.PockerManBase{}
	mbase.SetUid(p.baseinfo.GetUid())
	mbase.SetStatus(int32(p.status))
	mbase.SetCoin(p.chips)
	if act == ACT_SEATDOWN {
		mbase.SetHeaderUrl(p.baseinfo.GetHeaderUrl())
		mbase.SetDeskIdx(p.posIndex + int32(1))
		mbase.SetNickName(p.baseinfo.GetName())
		mbase.SetSex(p.baseinfo.GetSex())
	}
	msg.Infos = append(msg.Infos, mbase)

	if p.waitFrom > 0 {
		msg.SetCountdownEnd(int32(p.waitFrom + COUNTDOWN_MAX))
	}
	r.SendMsg2Others(msg, "S2CAction")
}

func (r *PockerRoom) sync_status(c2s *rpc.C2SAction) {
	p := r.GetPlayerByID(c2s.GetUid())
	if p == nil {
		logger.Error("sync_status r.GetPlayerByID return nil, uid:", c2s.GetUid())
		return
	}

	msg := &rpc.S2CAction{}
	msg.SetOperater(c2s.GetUid())
	msg.SetAct(c2s.GetAct())
	if c2s.GetAct() == int32(ACT_CALL) {
		msg.SetRaise(r.lastValue)
	} else if c2s.GetAct() == int32(ACT_ALLIN) || c2s.GetAct() == int32(ACT_RAISE) {
		msg.SetRaise(c2s.GetRaise())
	}
	mbase := &rpc.PockerManBase{}
	mbase.SetUid(p.baseinfo.GetUid())
	mbase.SetStatus(int32(p.status))
	mbase.SetCoin(p.chips)
	mbase.SetDrops(p.drops)
	msg.Infos = append(msg.Infos, mbase)
	r.SendMsg2Others(msg, "S2CAction")
}

func (r *PockerRoom) gen_pocker() {
	r.leftpockers = []pocker{}
	for i := 0; i <= 3; i++ {
		for j := int32(2); j <= 14; j++ {
			pk := pocker{
				eType: i,
				num:   j,
			}
			r.leftpockers = append(r.leftpockers, pk)
		}
	}
}

//洗牌
func (r *PockerRoom) shuffle() bool {
	logger.Info("洗牌...")
	r.gen_pocker()

	if len(r.leftpockers) != POCKER_CNT {
		logger.Error("Shuffle, 扑克总数不对：%d, 需要数量：%d", len(r.leftpockers), POCKER_CNT)
		return false
	}

	end := len(r.leftpockers) - 2
	for i := end; i > 0; i-- {
		index := rand.Intn(i)
		tmp := r.leftpockers[index]
		r.leftpockers[index] = r.leftpockers[i]
		r.leftpockers[i] = tmp
	}
	return true
}
