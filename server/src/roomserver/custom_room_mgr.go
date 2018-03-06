package roomserver

import (
	conn "centerclient"
	cmn "common"
	//	"fmt"
	"logger"
	"rpc"
	//"runtime/debug"
	//	"strconv"
	"time"
	//  ds "daerserver"
	"errors"
	gf "globalfunc"
	"strconv"
	//"strings"
	"sync"
	"timer"
)

type CustomRoomMgr struct {
	t *timer.Timer

	rooms        map[int32]map[int32]cmn.GameRoom //保存所有的房间列表（类型：ID:Room）
	createTime   map[string]int64                 //创建房间的时间
	playerInRoom map[string]cmn.GameRoom          //保存玩家所在的游戏房间

	idGenerator *cmn.RoomIDS //ID产生器
	rl          sync.RWMutex //读写锁
}

var customRoomMgr *CustomRoomMgr

func (self *CustomRoomMgr) init() {
	//self.createTimer()
	self.idGenerator = &cmn.RoomIDS{}
	self.idGenerator.Init(1000, 7999)
	self.rooms = make(map[int32]map[int32]cmn.GameRoom, 0)
	self.createTime = make(map[string]int64, 0)
	self.playerInRoom = make(map[string]cmn.GameRoom)
}

// func (self *CustomRoomMgr) createTimer() {
// 	self.t = timer.NewTimer(time.Second)
// 	self.t.Start(
// 		func() {
// 			defer func() {
// 				if r := recover(); r != nil {
// 					logger.Error("player tick runtime error :", r)
// 					debug.PrintStack()
// 				}
// 			}()

// 			self.OnTick()
// 		},
// 	)
// }

// func (self *CustomRoomMgr) OnTick() {
// 	//fmt.Println("OnTick time:%d", time.Now().Unix())

// 	for _, rooms := range self.rooms {
// 		for _, room := range rooms {
// 			if !room.IsEmpty() {
// 				room.UpdateTimer(time.Microsecond * 50)
// 			}
// 		}
// 	}

// }

//添加一个房间到房间列表
func (self *CustomRoomMgr) AddRoom(gameType, roomID int32, room cmn.GameRoom) {
	self.rl.Lock()
	defer self.rl.Unlock()

	if _, exist := self.rooms[gameType]; !exist {
		self.rooms[gameType] = make(map[int32]cmn.GameRoom)
	}

	rooms := self.rooms[gameType]

	if _, exist := rooms[roomID]; !exist {
		rooms[roomID] = room
	} else {
		logger.Error("存在此房间：", roomID)
	}
}

//进入游戏
func (self *CustomRoomMgr) OnEnterCustomRoom(roomID int32, pwd string, msg *rpc.PlayerBaseInfo) {
	//检测输入的参数
	if msg == nil {
		logger.Error("CustomRoomMgr.EnterGame: msg is nil.")
		return
	}

	//检查能否进入房间
	self.rl.RLock()
	if ok, code := self.CanEnterRoom(roomID, pwd, msg); !ok {
		// gameType, room := self.GetRoomByID(roomID)
		// cr := ConvertToCustomRoom(gameType, room)

		self.rl.RUnlock()
		self.SendEnterCustomACK(msg.GetUid(), roomID, code)
		return
	}
	self.rl.RUnlock()

	//重连进入
	playerID := msg.GetUid()
	self.rl.RLock()
	if rm, ok := self.playerInRoom[playerID]; ok {
		self.rl.RUnlock()

		logger.Info("CustomRoomMgr.EnterGame: palyer reEnter room:", roomID, msg.GetName())
		if rm == nil {
			logger.Error("CustomRoomMgr:EnterGame, room is nil")
			return
		}

		//rm.ReEnter(playerID, msg)

		pip := cmn.RoomMsgQueue{
			Msg:  msg,
			Func: "ReEnter",
		}
		*rm.GetRcvThreadHandle() <- pip

		return
	}
	self.rl.RUnlock()

	//首次进入
	logger.Info("CustomRoomMgr.EnterGame: palyer enter room:", roomID, msg.GetName())

	// self.rl.Lock()
	// defer self.rl.Unlock()

	//获取房间
	gameType, room := self.GetRoomByID(roomID)

	player := gf.NewPlayer(gameType, playerID, msg)
	//room.Enter(player)

	pip := cmn.RoomMsgQueue{
		Msg:  player,
		Func: "Enter",
	}
	*room.GetRcvThreadHandle() <- pip

	self.rl.Lock()
	self.playerInRoom[playerID] = room
	self.rl.Unlock()

	return
}

//检查能进入房间吗
func (self *CustomRoomMgr) CanEnterRoom(roomID int32, pwd string, msg *rpc.PlayerBaseInfo) (ok bool, code int32) {
	self.rl.RLock()
	defer self.rl.RUnlock()
	//获取房间
	gameType, room := self.GetRoomByID(roomID)
	if room == nil {
		return false, ECRNotExistRoom
	}

	//房间是否已经满了
	_, isReEnter := self.playerInRoom[msg.GetUid()]
	if !isReEnter && room.IsFull() {
		return false, ECRFull
	}

	//转换房间到自建房间
	cr := ConvertToCustomRoom(gameType, room)
	if cr == nil {
		logger.Error("转换房间出错")
		return false, ECRConvertRoomFailed
	}

	//检查密码是否正确
	// uid := msg.GetUid()
	// isSelfRoom := uid == cr.owner
	// if !isSelfRoom && !room.IsInRoom(uid) {
	// 	if cr.pwd != "" && cr.pwd != pwd {
	// 		return false, ECRPwdError
	// 	}
	// }

	//根据结算类型确定是否需要对金币进行检查
	if cr.currencyType == CTCoin {
		//从登录不用检查金币
		if isReEnter {
			return true, ECRNone
		}
		//检测金币是否达到了房间的要求
		if success, c := self.CheckCustomRoomCoin(cr, msg); !success {
			return success, c
		}
	}

	return true, ECRNone

}

//发送进入房间ACK
func (self *CustomRoomMgr) SendEnterCustomACK(uid string, roomID, code int32) {

	//组织发送的参数
	rmMsg := &rpc.EnterCustomRoomACK{}
	rmMsg.SetRoomId(roomID)

	gameType, room := self.GetRoomByID(roomID)
	if room == nil {
		code = ECRNotExistRoom

	} else {
		rmMsg.SetDifen(room.GetDifen())

		cr := ConvertToCustomRoom(gameType, room)

		if cr == nil {
			code = ECRConvertRoomFailed
		} else {
			rmMsg.SetIsOwner(uid == cr.owner)
			rmMsg.SetGameType(cr.gameType)
			rmMsg.SetTimes(cr.times)
			rmMsg.SetCurTimes(cr.curTimes)
		}
	}

	rmMsg.SetCode(code)

	logger.Info("发送进入自建房间：", rmMsg)
	if err := conn.SendCommonNotify2S([]string{uid}, rmMsg, "EnterCustomRoomACK"); err != nil {
		logger.Error("发送进入自建房间出错：", err)
	}
}

//leave room
func (self *CustomRoomMgr) OnLeaveGame(uid string) error {
	self.rl.RLock()
	room := self.playerInRoom[uid]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("CustomRoomMgr LeaveGame error, cant find room uid:%s", uid)
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  uid,
		Msg2: false,
		Func: "Leave",
	}
	*room.GetRcvThreadHandle() <- pip

	// //离开房间
	// logger.Info("玩家：%s 离开自建房间。", uid)
	// if room.Leave(uid, false) {
	// 	delete(self.playerInRoom, uid)
	// 	//空了就把房间删除了
	// 	if room.IsEmpty() {
	// 		self.idGenerator.RecoverID(room.ID())
	// 		gameType := self.GetGameType(room.ID())
	// 		if _, exist := self.rooms[gameType]; exist {
	// 			delete(self.rooms[gameType], room.ID())
	// 		}

	// 		logger.Info("玩家：%s 离开自建房间。房间为空了删除房间：%s", uid, room.ID())
	// 	}
	// } else {
	// 	//调试阶段使用
	// 	//room.ResetRoom()
	// 	//room.SwitchRoomState(RSReady)
	// 	//room.ForceAllPlayerLeave()
	// }

	return nil
}

//修改离开的玩家数据
func (self *CustomRoomMgr) DeleteLeavePlayerInfo(uid string) {

	logger.Error("CustomRoomMgr===开始删除一个玩家在房间管理的关系。")
	self.rl.RLock()
	room := self.playerInRoom[uid]
	self.rl.RUnlock()
	if room != nil {
		isDeleteRoom := room.IsEmpty()
		if isDeleteRoom {
			self.rl.Lock()
			self.idGenerator.RecoverID(room.UID())
			self.rl.Unlock()

			gameType := self.GetGameType(room.UID())
			self.rl.Lock()
			if _, exist := self.rooms[gameType]; exist {
				delete(self.rooms[gameType], room.UID())
			}
			self.rl.Unlock()

			logger.Info("CustomRoomMgr===玩家：%s 离开自建房间。房间为空了删除房间：%s", uid, room.UID())
		}

		self.rl.Lock()
		delete(self.playerInRoom, uid)
		self.rl.Unlock()
	}

}

//请求执行的动作
func (self *CustomRoomMgr) OnActionGame(msg *rpc.ActionREQ) error {
	self.rl.RLock()
	room := self.playerInRoom[msg.GetPlayerID()]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("CustomRoomMgr ActionGame error, cant find room uid:%s, excute action:%d", msg.GetPlayerID(), msg.GetAction())
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  msg,
		Func: "ActionREQ",
	}
	*room.GetRcvThreadHandle() <- pip

	//room.OnPlayerDoAction(msg)
	return nil
}

//请求贴鬼
func (self *CustomRoomMgr) TieGuiREQ(msg *rpc.MJTieGuiREQ) error {
	self.rl.RLock()
	room := self.playerInRoom[msg.GetPlayerID()]
	self.rl.RUnlock()

	if room == nil {
		logger.Error("MaJiangRoomMgr TieGuiREQ error, cant find room uid:%s", msg.GetPlayerID())
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  msg,
		Func: "TieGuiREQ",
	}
	*room.GetRcvThreadHandle() <- pip

	return nil
}

//请求解散房间
func (self *CustomRoomMgr) OnJieSanRoom(uid string, msg *rpc.JieSanRoomREQ) error {
	self.rl.RLock()
	room := self.playerInRoom[uid]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("CustomRoomMgr OnJieSuanRoom error, cant find room uid:%s", uid)
		return nil
	}

	room.OnJieSanRoom(uid, msg)
	return nil
}

//创建房间
func (self *CustomRoomMgr) OnCreateRoom(playerBasic *rpc.PlayerBaseInfo, msg *rpc.CreateRoomREQ, ownRoomCardAmount int32) error {
	//检查输入参数
	if msg == nil {
		logger.Error("msg is nil.")
		return nil
	}

	if playerBasic == nil {
		logger.Error("playerBasic is nil.")
		return nil
	}

	logger.Info("创建房间：")

	//根据游戏类型创建房间
	gameType := msg.GetGameType()
	uid := playerBasic.GetUid()

	// self.rl.Lock()
	// defer self.rl.Unlock()

	switch gameType {
	case cmn.DaerGame:
		logger.Error("DaerGame===============0")
		self.rl.Lock()
		roomId := self.idGenerator.UseID()
		self.rl.Unlock()

		logger.Error("DaerGame===============1")
		if ok, code := self.CanCreateRoom(playerBasic, roomId, msg, ownRoomCardAmount); !ok {
			SendCreateRoomACK(uid, nil, code)
			self.idGenerator.RecoverID(roomId)
			return nil
		}
		logger.Error("DaerGame===============2")
		room := NewCustomDaerRoom(roomId, uid, msg)
		roomInfo := GenerateRoomInfo(gameType, room)
		if roomInfo != nil {
			logger.Error("DaerGame===============3")
			self.AddRoom(gameType, roomId, room)
			logger.Error("DaerGame===============4")
			self.rl.Lock()
			self.createTime[uid] = time.Now().Unix()
			self.rl.Unlock()
			logger.Error("DaerGame===============5")
			SendCreateRoomACK(uid, roomInfo, ECCRNone)
		} else {
			SendCreateRoomACK(uid, nil, ECCRConvertRoomFailed)
			self.idGenerator.RecoverID(roomId)
		}

	case cmn.MaJiang:
		logger.Error("MaJiang===============0")
		self.rl.Lock()
		roomId := self.idGenerator.UseID()
		self.rl.Unlock()

		logger.Error("MaJiang===============1")
		if ok, code := self.CanCreateRoom(playerBasic, roomId, msg, ownRoomCardAmount); !ok {
			SendCreateRoomACK(uid, nil, code)
			self.idGenerator.RecoverID(roomId)
			return nil
		}
		logger.Error("MaJiang===============2")
		room := NewCustomMaJiangRoom(roomId, uid, msg)
		roomInfo := GenerateRoomInfo(gameType, room)
		if roomInfo != nil {
			logger.Error("MaJiang===============3")
			self.AddRoom(gameType, roomId, room)
			logger.Error("MaJiang===============4")
			self.rl.Lock()
			self.createTime[uid] = time.Now().Unix()
			self.rl.Unlock()
			logger.Error("MaJiang===============5")
			SendCreateRoomACK(uid, roomInfo, ECCRNone)
		} else {
			SendCreateRoomACK(uid, nil, ECCRConvertRoomFailed)
			self.idGenerator.RecoverID(roomId)
		}
	case cmn.DeZhouPuker:
	default:
		logger.Error("不支持的游戏类型：", gameType)
	}
	return nil
}

func (self *CustomRoomMgr) CanCreateRoom(playerBasic *rpc.PlayerBaseInfo, id int32, msg *rpc.CreateRoomREQ, ownRoomCardAmount int32) (ok bool, code int32) {

	//检查输入参数
	if playerBasic == nil {
		return false, ECCRUnknowError
	}

	//检查是否已经在房间里了
	gameType := int(msg.GetGameType())
	uid := playerBasic.GetUid()

	if isInRoom, _ := self.IsInRoom(uid); isInRoom {
		return false, ECCRAlreadyInRoom
	}

	//检查ID
	if id <= 0 {
		return false, ECCRNoneID
	}

	//检查倍数范围
	maxMultiple := msg.GetMaxMultiple()
	cfg := cmn.GetCustomRoomConfig(strconv.Itoa(gameType))
	if cfg == nil {
		return false, ECCRConfigError
	}
	if maxMultiple < cfg.MinMultipleLimit || maxMultiple > cfg.MaxMultipleLimit {
		return false, ECCRMultipleLimit
	}

	//检查创建时间间隔
	cfg = cmn.GetCustomRoomConfig(strconv.Itoa(gameType))
	if cfg == nil {
		return false, ECCRConfigError
	}

	self.rl.RLock()
	if val, exist := self.createTime[uid]; exist {
		self.rl.RUnlock()
		if time.Now().Unix()-val < int64(cfg.CreateRoomInterval) {
			return false, ECCRCreateFrequently
		}
	} else {
		self.rl.RUnlock()
	}

	//根据结算货币类型，进行检查
	if msg.GetCurrencyType() == CTCoin {
		//检查自己的金币是否大于进入房间的金币，如果自己的金币都不足以进入房间，那么就不能创建
		if playerBasic.GetCoin() < msg.GetLimitCoin() {
			return false, ECCRGreaterSelfCoin
		}

		//检查创建房间的最低金币
		cfg := cmn.GetCustomRoomConfig(strconv.Itoa(gameType))
		if cfg == nil {
			return false, ECCRConfigError
		}
		if playerBasic.GetCoin() < cfg.CreateRoomMinLimit {
			return false, ECCRCreateRoomMinCoinLimit
		}

		//检查房间名字
		// nameLegth := int32(strings.Count(msg.GetName(), "") - 1)
		// cfg = cmn.GetCustomRoomConfig(strconv.Itoa(gameType))
		// if cfg == nil {
		// 	return false, ECCRConfigError
		// }

		// logger.Info("房间的名字长度：%s，房间的长度限制%s-%s", nameLegth, cfg.NameMinLength, cfg.NameMaxLength)
		// if nameLegth < cfg.NameMinLength || nameLegth > cfg.NameMaxLength {
		// 	return false, ECCRNameLength
		// }

		// //检查密码
		// pwdLegth := int32(strings.Count(msg.GetPwd(), "") - 1)
		// cfg = cmn.GetCustomRoomConfig(strconv.Itoa(gameType))
		// if cfg == nil {
		// 	return false, ECCRConfigError
		// }

		// if (pwdLegth > 0) && (pwdLegth < cfg.PwdMinLength || pwdLegth > cfg.PwdMaxLength) {
		// 	return false, ECCRPwdLength
		// }

		//检查底注
		diFen := msg.GetDifen()
		cfg = cmn.GetCustomRoomConfig(strconv.Itoa(gameType))
		if cfg == nil {
			return false, ECCRConfigError
		}

		if diFen < cfg.DifenMinLimit || diFen > cfg.DifenMaxLimit || diFen%cfg.CoinDifenMulti != 0 {
			return false, ECCRDifen
		}
	} else {
		//检查比赛次数
		matchTime := msg.GetTimes()
		cfg = cmn.GetCustomRoomConfig(strconv.Itoa(gameType))
		if cfg == nil {
			return false, ECCRConfigError
		}

		if matchTime < cfg.TimesMinLimit || matchTime > cfg.TimesMaxLimit {
			return false, ECCRMatchTimes
		}

		//积分房的房卡检查
		costRoomCardCount := GetCostRoomCardCount(int32(gameType), matchTime)
		if costRoomCardCount < 0 {
			return false, ECCRConfigError
		}

		if ownRoomCardAmount < costRoomCardCount {
			return false, ECCRNotEnoughRoomCard
		}
	}

	return true, ECCRNone
}

//发送创建房间消息
func SendCreateRoomACK(uid string, roomInfo *rpc.RoomInfo, code int32) {
	msg := &rpc.CreateRoomACK{}
	msg.SetCode(code)
	msg.SetRoom(roomInfo)

	logger.Info("向客服端发送创建房间ACK：", msg)
	if err := conn.SendCommonNotify2S([]string{uid}, msg, "CreateRoomACK"); err != nil {
		logger.Error("发送创建自建房间出错：", err)
	}
}

//请求房间列表
func (self *CustomRoomMgr) OnObtainRoomList(uid string) error {
	self.rl.RLock()
	defer self.rl.RUnlock()

	//组织网络数据
	msg := &rpc.RoomListACK{}
	for gameType, rooms := range self.rooms {
		for _, room := range rooms {
			roomInfo := GenerateRoomInfo(gameType, room)
			if roomInfo != nil {
				msg.RoomList = append(msg.RoomList, roomInfo)
			}
		}
	}

	//发送消息到客服端
	logger.Info("发送RoomListACK: ", msg)
	if err := conn.SendCommonNotify2S([]string{uid}, msg, "RoomListACK"); err != nil {
		logger.Error("发送进入自建房间出错：", err)
		return err
	}

	return nil
}

//查找房间
func (self *CustomRoomMgr) OnFindRoom(uid string, msg *rpc.FindRoomREQ) error {
	//检查输入参数
	if msg == nil {
		logger.Error("msg is nil.")
		return errors.New("msg is nil.")
	}

	//查找房间
	rMsg := &rpc.FindRoomACK{}
	rMsg.SetRoom(nil)
	rMsg.SetCode(EFRRequireParamError)

	name := msg.GetName()
	if name != "" {
		gameType, room := self.GetRoomByName(name)
		if room != nil {
			roomInfo := GenerateRoomInfo(gameType, room)
			rMsg.SetRoom(roomInfo)

			if roomInfo != nil {
				rMsg.SetCode(EFRNone)
			} else {
				logger.Error("产生RoomInfo时出错。")
				rMsg.SetCode(EFRGenerateRoomInfoError)
			}

		} else {
			rMsg.SetRoom(nil)
			rMsg.SetCode(EFRNotFind)
		}
	}

	id := msg.GetId()
	if id > 0 {
		gameType, room := self.GetRoomByID(id)
		if room != nil {
			roomInfo := GenerateRoomInfo(gameType, room)
			rMsg.SetRoom(roomInfo)

			if roomInfo != nil {
				rMsg.SetCode(EFRNone)
			} else {
				logger.Error("产生RoomInfo时出错。")
				rMsg.SetCode(EFRGenerateRoomInfoError)
			}
		} else {
			rMsg.SetRoom(nil)
			rMsg.SetCode(EFRNotFind)
		}
	}

	//发送消息
	if err := conn.SendCommonNotify2S([]string{uid}, rMsg, "FindRoomACK"); err != nil {
		logger.Error("发送查找自建房间出错：", err)
		return err
	}

	return nil
}

//踢出房间
func (self *CustomRoomMgr) OnForceLeaveRoom(uid string, msg *rpc.ForceLeaveRoomREQ) error {
	//检查输入参数
	if msg == nil {
		logger.Error("msg is nil.")
		return errors.New("msg is nil.")
	}

	//检查是否是房主，只有房主才能踢人
	leavePlayerID := msg.GetId()
	if exist, room := self.IsInRoom(leavePlayerID); exist && room != nil {
		//在游戏中是不能踢人的
		if room.IsGaming() {
			return nil
		}

		//检查踢人的是不是房主
		gameType := self.GetGameType(room.UID())
		cr := ConvertToCustomRoom(gameType, room)
		if cr != nil && cr.owner == uid {
			self.OnLeaveGame(leavePlayerID)
		}
	}

	//发送消息
	// if err := conn.SendCommonNotify2S([]string{uid}, rMsg, "FindRoomACK"); err != nil {
	// 	logger.Error("发送查找自建房间出错：", err)
	// 	return err
	// }

	return nil
}

func (self *CustomRoomMgr) IsExistRoom(roomID int32) (exist bool) {
	self.rl.RLock()
	defer self.rl.RUnlock()

	for _, rooms := range self.rooms {
		if rooms == nil {
			continue
		}

		if _, exist := rooms[roomID]; exist {
			return true
		}
	}

	return false
}

func (self *CustomRoomMgr) GetRoomByID(roomID int32) (gameType int32, room cmn.GameRoom) {
	self.rl.RLock()
	defer self.rl.RUnlock()

	for gt, rooms := range self.rooms {
		if rooms == nil {
			continue
		}

		if room, exist := rooms[roomID]; exist && room != nil {
			return gt, room
		}
	}

	return cmn.UnknownGame, nil
}

func (self *CustomRoomMgr) GetRoomByName(name string) (gameType int32, room cmn.GameRoom) {
	self.rl.RLock()
	defer self.rl.RUnlock()
	for gt, rooms := range self.rooms {
		if rooms == nil {
			continue
		}

		for _, room := range rooms {
			if room.Name() == name {
				return gt, room
			}
		}
	}

	return cmn.UnknownGame, nil
}

func (self *CustomRoomMgr) GetGameType(roomID int32) (gameType int32) {
	self.rl.RLock()
	defer self.rl.RUnlock()
	for gt, rooms := range self.rooms {
		if rooms == nil {
			continue
		}

		if _, exist := rooms[roomID]; exist {
			return gt
		}
	}

	return cmn.UnknownGame
}

func (self *CustomRoomMgr) IsInRoom(playerID string) (bool, cmn.GameRoom) {
	self.rl.RLock()
	defer self.rl.RUnlock()

	room, exist := self.playerInRoom[playerID]
	return exist, room
}

func (self *CustomRoomMgr) CheckCustomRoomCoin(cr *CustomRoom, pbi *rpc.PlayerBaseInfo) (ok bool, code int32) {
	if cr == nil {
		logger.Error("CustomRoomMgr.CheckCustomRoomCoin: cr is nil")
		return
	}
	if pbi == nil {
		logger.Error("CustomRoomMgr.CheckCustomRoomCoin: pbi is nil")
		return
	}

	if exist, _ := self.IsInRoom(pbi.GetUid()); exist {
		cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(cr.gameType)))
		if cfg == nil {
			logger.Error("GetCustomRoomConfig return nil")
			return
		}

		if pbi.GetCoin() < cfg.DifenMinLimit {
			return false, ECRLessCoin
		}
	} else {
		if pbi.GetCoin() < cr.limitCoin {
			return false, ECRLessCoin
		}
	}

	return true, ECRNone
}

func (self *CustomRoomMgr) SendDeskChatMsg(msg *rpc.FightRoomChatNotify) bool {
	self.rl.RLock()
	room := self.playerInRoom[msg.GetPlayerID()]
	self.rl.RUnlock()

	if room == nil {
		logger.Error("CustomRoomMgr SendDeskChatMsg error, cant find room uid:%s", msg.GetPlayerID())
		return false
	}

	room.SendCommonMsg2Others(msg)
	return false
}
