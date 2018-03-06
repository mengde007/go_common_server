//room_mgr应该包含到daerserver包里面，在大二服务器启动的时候 创建RoomMgr对象并初始化

package pockerserver

import (
	"centerclient"
	// "common"
	// "fmt"
	"logger"
	// "math/rand"
	"rpc"
	// "strconv"
	// "sync"
)

type CustomPockerRoomMgr struct {
	rooms        map[int32]*PockerRoom  //保存所有的房间
	playerInRoom map[string]*PockerRoom //玩家所在的房间
	onlinePlayer map[int32]uint32       //在线玩家数量
	roomNos      []int32                //房间号
}

var customRoomMgr *CustomPockerRoomMgr

func (self *CustomPockerRoomMgr) init() {
	self.rooms = make(map[int32]*PockerRoom, 0)
	self.playerInRoom = make(map[string]*PockerRoom, 0)
	self.onlinePlayer = make(map[int32]uint32, 0)

	//gen roomNo
	for i := int32(8000); i <= int32(9999); i++ {
		self.roomNos = append(self.roomNos, i)
	}
}

func (c *CustomPockerRoomMgr) get_room_number() int32 {
	logger.Info("**********get_room_number")
	if len(c.roomNos) <= 0 {
		logger.Error("get_room_number() roomNos is use out")
		return int32(-1)
	}

	num := c.roomNos[0]
	c.roomNos = append(c.roomNos[:0], c.roomNos[1:]...)
	return num
}

func (c *CustomPockerRoomMgr) back_room_number(num int32) {
	for _, v := range c.roomNos {
		if v == num {
			logger.Error("back_room_number 重复回收房间号:%d", num)
			return
		}
	}
	c.roomNos = append(c.roomNos, num)
}

func (c *CustomPockerRoomMgr) CreateRoom(blindId, limId int32) *PockerRoom {
	logger.Info("***********CreateRoom1")
	roomNo := c.get_room_number()
	if roomNo == int32(-1) {
		logger.Error("CreateRoom 房号已经用完 blindId:%d, limId:%d", blindId, limId)
		return nil
	}

	room := NewCustomPockerRoom(blindId, limId)
	if room == nil {
		logger.Error("**********CreateRoom 失败")
		return nil
	}
	room.roomNo = roomNo

	logger.Info("***********CreateRoom2")
	if _, ok := c.rooms[roomNo]; ok {
		logger.Error("CreateRoom 房间已经存在 blindId:%d, limId:%d， roomNo:%d", blindId, limId, roomNo)
		return nil
	}
	c.rooms[roomNo] = room
	return room
}

func (p *CustomPockerRoomMgr) GetRoom(roomNo int32) (*PockerRoom, int32) {
	room, ok := p.rooms[roomNo]
	if !ok {
		logger.Error("GetRoom 房间不存在 roomNo:%d", roomNo)
		return nil, int32(4)
	}

	if !room.AtomicInc() {
		logger.Error("房间人满了, roomNo:%d", roomNo)
		return nil, int32(6)
	}
	return room, int32(0)
}

func (c *CustomPockerRoomMgr) check_dismiss_room(roomNo int32) {
	room, ok := c.rooms[roomNo]
	if !ok {
		logger.Error("GetRoom 房间不存在 roomNo:%d", roomNo)
		return
	}

	if room.GetPlayerNum() != int32(0) {
		return
	}

	delete(c.rooms, roomNo)
	c.back_room_number(roomNo)
	room.exit <- false
}

//进入游戏
func (self *CustomPockerRoomMgr) EnterRoom(roomNo int32, msg *rpc.PlayerBaseInfo) {
	logger.Info("EnterRoom begin...")
	defer logger.Info("EnterRoom end...")

	if msg == nil {
		logger.Error("CustomPockerRoomMgr.EnterRoom: msg is nil.")
		return
	}

	//重连进入
	playerID := msg.GetUid()
	if room, ok := self.playerInRoom[playerID]; ok {
		if room.is_inroom(playerID) {
			logger.Info("CustomPockerRoomMgr.EnterGame: palyer reEnter room:", roomNo, msg.GetName())
			if room == nil {
				logger.Error("CustomPockerRoomMgr:EnterGame, room is nil")
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
			delete(self.playerInRoom, playerID)
		}
	}

	//正常进入房间
	logger.Info("正常进入自建房 roomNo:%d", roomNo)
	room, code := self.GetRoom(roomNo)
	if room == nil {
		send := &rpc.PockerRoomInfo{}
		send.SetCode(code)
		centerclient.SendCommonNotify2S([]string{msg.GetUid()}, send, "PockerRoomInfo")
		logger.Error("进入自建房失败, code:3")
		return
	}

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
	self.onlinePlayer[roomNo] += 1

	logger.Info("CustomPockerRoomMgr.EnterGame: player:%s enter room: Online:%d", player.baseinfo.GetUid(), self.onlinePlayer[roomNo])
	return
}

//leave room
func (self *CustomPockerRoomMgr) LeaveGame(msg *rpc.C2SAction) bool {
	logger.Info("LeaveGame called")
	uid := msg.GetUid()
	room := self.playerInRoom[uid]
	if room == nil {
		logger.Error("CustomPockerRoomMgr LeaveGame error, cant find room uid:%s", uid)
		return false
	}

	pip := &RoomActQueue{
		Msg:   *msg,
		Param: "ActionREQ",
		Func:  "Action",
	}
	room.rcv <- pip

	delete(self.playerInRoom, uid)
	self.onlinePlayer[room.roomNo] -= 1

	self.check_dismiss_room(room.roomNo)
	return true
}

func (c *CustomPockerRoomMgr) PlayerAction(msg *rpc.C2SAction) error {
	logger.Info("PlayerAction begin..., act:%d", msg.GetAct())
	logger.Info("PlayerAction end...")
	room := c.playerInRoom[msg.GetUid()]
	if room == nil {
		logger.Error("CustomPockerRoomMgr ActionGame error, cant find room uid:%s", msg.GetUid())
		return nil
	}

	if msg.GetAct() == int32(ACT_LEAVE) {
		c.LeaveGame(msg)
		return nil

	} else if msg.GetAct() == int32(ACT_SEATDOWN) {
		if !room.AtomicInc() {
			logger.Error("人数已满，不能坐下")
			return nil
		}
	}

	pip := &RoomActQueue{
		Msg:   *msg,
		Param: "ActionREQ",
		Func:  "Action",
	}
	room.rcv <- pip
	return nil
}

//检查玩家是否在游戏中（只要在房间都标示在游戏，哪怕他是在结算阶段）
func (self *CustomPockerRoomMgr) IsInRoom(uid string) bool {
	if val, exist := self.playerInRoom[uid]; exist && val != nil {
		return true
	}
	return false
}

func (self *CustomPockerRoomMgr) SendDeskChatMsg(msg *rpc.FightRoomChatNotify) bool {
	room := self.playerInRoom[msg.GetPlayerID()]
	if room == nil {
		logger.Error("CustomPockerRoomMgr SendDeskChatMsg error, cant find room uid:%s", msg.GetPlayerID())
		return false
	}

	room.SendMsg2Others(msg, "FightRoomChatNotify")
	return false
}
