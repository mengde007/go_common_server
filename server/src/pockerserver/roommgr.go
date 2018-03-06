//room_mgr应该包含到daerserver包里面，在大二服务器启动的时候 创建RoomMgr对象并初始化

package pockerserver

import (
	"centerclient"
	"common"
	// "fmt"
	"logger"
	"math/rand"
	"rpc"
	"strconv"
	"sync"
)

type PockerRoomMgr struct {
	rooms        map[int32][]*PockerRoom //保存所有的房间
	playerInRoom map[string]*PockerRoom  //玩家所在的房间
	onlinePlayer map[int32]uint32        //在线玩家数量
	rl           sync.RWMutex
	nrl          sync.Mutex //创建房间锁
	roomNo       int32      //房间号
}

//main thread <-> room thread msg
type RoomActQueue struct {
	Msg   interface{}
	Param string
	Func  string
}

var pockerRoomMgr *PockerRoomMgr

func (self *PockerRoomMgr) init() {
	self.rooms = make(map[int32][]*PockerRoom, 0)
	self.playerInRoom = make(map[string]*PockerRoom, 0)
	self.onlinePlayer = make(map[int32]uint32, 0)
}

func (p *PockerRoomMgr) CanChangeDesk(roomType int32, no int32) bool {
	if roomList, ok := p.rooms[roomType]; ok {
		for _, v := range roomList {
			if v.roomNo != no && !v.IsFull() {
				return true
			}
		}
	}
	return false
}

func (p *PockerRoomMgr) GetRoom(roomType int32, bRand bool, roomNo int32) *PockerRoom {
	if bRand {
		if roomList, ok := p.rooms[roomType]; ok {
			index := rand.Intn(len(roomList))
			for i := index; i < len(roomList); i++ {
				if roomList[i].roomNo != roomNo && roomList[i].AtomicInc() {
					return roomList[i]
				}
			}
			for i := 0; i < index; i++ {
				if roomList[i].roomNo != roomNo && roomList[i].AtomicInc() {
					return roomList[i]
				}
			}
		}
	} else {
		if roomList, ok := p.rooms[roomType]; ok {
			for _, room := range roomList {
				if room.AtomicInc() {
					return room
				}
			}
		}
	}

	p.nrl.Lock()
	p.roomNo += int32(1)
	room := NewPockerRoom(roomType)
	room.roomNo = p.roomNo

	if _, ok := p.rooms[roomType]; !ok {
		p.rooms[roomType] = make([]*PockerRoom, 0)
	}
	p.rooms[roomType] = append(p.rooms[roomType], room)
	room.AtomicInc()
	p.nrl.Unlock()
	return room

}

//进入游戏
func (self *PockerRoomMgr) EnterGame(roomType int32, msg *rpc.PlayerBaseInfo) {
	logger.Info("EnterGame begin...")
	defer logger.Info("EnterGame end...")

	if msg == nil {
		logger.Error("PockerRoomMgr.EnterGame: msg is nil.")
		return
	}
	if ok, code := common.CheckPockerCoin(roomType, msg); !ok {
		send := &rpc.PockerRoomInfo{}
		send.SetCode(code)
		centerclient.SendCommonNotify2S([]string{msg.GetUid()}, send, "PockerRoomInfo")
		logger.Error("进入失败, code:%d", code)
		return
	}

	//重连进入
	playerID := msg.GetUid()
	if room, ok := self.playerInRoom[playerID]; ok {
		if room.is_inroom(playerID) {
			logger.Info("PockerRoomMgr.EnterGame: palyer reEnter room:", roomType, msg.GetName())
			if room == nil {
				logger.Error("PockerRoomMgr:EnterGame, room is nil")
				return
			}
			pip := &RoomActQueue{
				Msg:   *msg,
				Param: "PlayerBaseInfo",
				Func:  "ReEnter",
			}
			room.rcv <- pip
			return
		} else {
			self.delete_from_room(playerID, roomType)
		}
	}

	//正常进入房间
	logger.Info("正常进入房间")
	room := self.GetRoom(roomType, false, int32(-1))
	logger.Info("GetRoom 成功")
	player := pockerman{
		baseinfo: msg,
		status:   STATUS_WATTING_JOIN,
	}
	pip := &RoomActQueue{
		Msg:   player,
		Param: "pockerman",
		Func:  "Enter",
	}
	room.rcv <- pip

	self.playerInRoom[player.baseinfo.GetUid()] = room
	self.onlinePlayer[roomType] += 1

	logger.Info("PockerRoomMgr.EnterGame: player:%s enter room: Online:%d", player.baseinfo.GetUid(), self.onlinePlayer[roomType])
	return
}

func (self *PockerRoomMgr) delete_from_room(uid string, eType int32) {
	delete(self.playerInRoom, uid)
	if self.onlinePlayer[eType] > 0 {
		self.onlinePlayer[eType] -= 1
	}
}

//leave room
func (self *PockerRoomMgr) LeaveGame(msg *rpc.C2SAction) bool {
	logger.Info("))))))))))))))))))))LeaveGame called")
	uid := msg.GetUid()
	room := self.playerInRoom[uid]
	if room == nil {
		logger.Error("PockerRoomMgr LeaveGame error, cant find room uid:%s", uid)
		return false
	}

	// p := room.GetPlayerByID(uid)
	// if p == nil {
	// 	p = room.get_stand_player(uid)
	// 	if p == nil {
	// 		logger.Error("在房间里没有查找到玩家：", uid)
	// 	}
	// 	return false
	// }

	pip := &RoomActQueue{
		Msg:   *msg,
		Param: "ActionREQ",
		Func:  "Action",
	}
	room.rcv <- pip

	self.delete_from_room(uid, room.eType)
	return true
}

func (p *PockerRoomMgr) PlayerAction(msg *rpc.C2SAction) error {
	logger.Info("PlayerAction begin..., act:%d", msg.GetAct())
	logger.Info("PlayerAction end...")
	room := p.playerInRoom[msg.GetUid()]
	if room == nil {
		logger.Error("PockerRoomMgr ActionGame error, cant find room uid:%s", msg.GetUid())
		return nil
	}

	if msg.GetAct() == int32(ACT_LEAVE) {
		p.LeaveGame(msg)
		return nil

	} else if msg.GetAct() == int32(ACT_SEATDOWN) {
		if !room.AtomicInc() {
			logger.Error("人数已满，不能坐下")
			return nil
		}
	} else if msg.GetAct() == int32(ACT_CHANGE_DESK) {
		if !p.CanChangeDesk(room.eType, room.roomNo) {
			return nil
		}

		if !p.LeaveGame(msg) {
			return nil
		}

		//正常进入房间
		room := p.GetRoom(room.eType, true, room.roomNo)
		player := pockerman{
			baseinfo: msg.GetBase(),
			status:   STATUS_WATTING_JOIN,
		}
		pip := &RoomActQueue{
			Msg:   player,
			Param: "pockerman",
			Func:  "Enter",
		}
		room.rcv <- pip

		p.playerInRoom[player.baseinfo.GetUid()] = room
		p.onlinePlayer[room.eType] += 1
		return nil
	}

	pip := &RoomActQueue{
		Msg:   *msg,
		Param: "ActionREQ",
		Func:  "Action",
	}
	room.rcv <- pip
	return nil
}

//自动坐下，按房间的限制进入金币，从高到低进入
func (self *PockerRoomMgr) quickly_seatdown(msg *rpc.PlayerBaseInfo) error {
	cfg := common.GetDaerRoomConfig(strconv.Itoa(24))
	if cfg == nil {
		logger.Error("QuicklySeatdown 出错 common.GetDaerRoomConfig(24) return nil ")
		return nil
	}
	if msg.GetCoin() >= cfg.MinLimit && msg.GetCoin() <= cfg.MaxLimit {
		self.EnterGame(int32(24), msg)
		return nil
	}

	///
	cfg = common.GetDaerRoomConfig(strconv.Itoa(23))
	if cfg == nil {
		logger.Error("QuicklySeatdown 出错 common.GetDaerRoomConfig(23) return nil ")
		return nil
	}
	if msg.GetCoin() >= cfg.MinLimit && msg.GetCoin() <= cfg.MaxLimit {
		self.EnterGame(int32(23), msg)
		return nil
	}

	///
	cfg = common.GetDaerRoomConfig(strconv.Itoa(22))
	if cfg == nil {
		logger.Error("QuicklySeatdown 出错 common.GetDaerRoomConfig(22) return nil ")
		return nil
	}
	if msg.GetCoin() >= cfg.MinLimit && msg.GetCoin() <= cfg.MaxLimit {
		self.EnterGame(int32(22), msg)
		return nil
	}

	///
	cfg = common.GetDaerRoomConfig(strconv.Itoa(21))
	if cfg == nil {
		logger.Error("QuicklySeatdown 出错 common.GetDaerRoomConfig(21) return nil ")
		return nil
	}
	if msg.GetCoin() >= cfg.MinLimit && msg.GetCoin() <= cfg.MaxLimit {
		self.EnterGame(int32(21), msg)
		return nil
	}
	logger.Error("QuicklySeatdown 出错了没打到相应房间, 玩家身上金币:%d", msg.GetCoin())
	return nil
}

//获取在线人数
func (self *PockerRoomMgr) GetOnlineNum() *rpc.OnlineInfo {
	sendMsg := &rpc.OnlineInfo{}
	for id, num := range self.onlinePlayer {
		msg := &rpc.OnlineBody{}
		msg.SetRoomId(int32(id))
		msg.SetNum(int32(num))
		sendMsg.Info = append(sendMsg.Info, msg)
	}
	return sendMsg
}

//检查玩家是否在游戏中（只要在房间都标示在游戏，哪怕他是在结算阶段）
func (self *PockerRoomMgr) IsInRoom(uid string) bool {
	if val, exist := self.playerInRoom[uid]; exist && val != nil {
		return true
	}

	return false
}

func (self *PockerRoomMgr) SendDeskChatMsg(msg *rpc.FightRoomChatNotify) bool {
	room := self.playerInRoom[msg.GetPlayerID()]
	if room == nil {
		logger.Error("PockerRoomMgr SendDeskChatMsg error, cant find room uid:%s", msg.GetPlayerID())
		return false
	}

	room.SendMsg2Others(msg, "FightRoomChatNotify")
	return false
}
