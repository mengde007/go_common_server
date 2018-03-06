//room_mgr应该包含到daerserver包里面，在大二服务器启动的时候 创建RoomMgr对象并初始化

package majiangserver

//package main

import (
	//conn "centerclient"
	cmn "common"
	"logger"
	"rpc"
	//"runtime/debug"
	//"strconv"
	//"time"
	//"timer"
	"sync"
)

type MaJiangRoomMgr struct {
	rooms        map[int32][]*MaJiangRoom //保存所有的房间
	playerInRoom map[string]*MaJiangRoom  //玩家所在的房间
	onlinePlayer map[int32]uint32         //在线玩家数量
	idGenerator  *cmn.RoomIDS             //ID产生器

	rl sync.RWMutex
}

var maJiangRoomMgr *MaJiangRoomMgr

func (self *MaJiangRoomMgr) init() {
	//self.createTimer()
	self.rooms = make(map[int32][]*MaJiangRoom, 0)
	self.playerInRoom = make(map[string]*MaJiangRoom, 0)
	self.onlinePlayer = make(map[int32]uint32, 0)
	self.idGenerator = &cmn.RoomIDS{}
	self.idGenerator.Init(0, 1000)
}

func (self *MaJiangRoomMgr) getNotFullRoom(roomType int32) *MaJiangRoom {
	self.rl.RLock()
	defer self.rl.RUnlock()

	if roomList, ok := self.rooms[roomType]; ok {
		for _, room := range roomList {
			if !room.IsFull() {
				return room
			}
		}
	}

	return nil
}

//进入游戏
func (self *MaJiangRoomMgr) EnterGame(roomType int32, msg *rpc.PlayerBaseInfo, isChangeDesk bool) {
	//检测输入的参数
	if msg == nil {
		logger.Error("MaJiangRoomMgr.EnterGame: msg is nil.")
		return
	}

	self.rl.RLock()
	if ok, code := cmn.CheckCoin(roomType, msg); !ok {
		SendEnterRoomErrorACK(msg.GetUid(), roomType, code, !isChangeDesk)
		self.rl.RUnlock()
		return
	}
	self.rl.RUnlock()

	//重连进入
	playerID := msg.GetUid()
	self.rl.RLock()
	if room, ok := self.playerInRoom[playerID]; ok {
		self.rl.RUnlock()
		logger.Info("MaJiangRoomMgr.EnterGame: palyer reEnter room:", roomType, msg.GetName())

		if room == nil {
			logger.Error("MaJiangRoomMgr:EnterGame, room is nil")
			return
		}

		pip := cmn.RoomMsgQueue{
			Msg:  msg,
			Func: "ReEnter",
		}
		room.rcv <- pip

		return
	}
	self.rl.RUnlock()

	//首次进入
	logger.Info("MaJiangRoomMgr.EnterGame: palyer enter room:", roomType, msg.GetName())
	// self.rl.Lock()
	// defer self.rl.Unlock()

	room := self.getNotFullRoom(roomType)
	if room == nil {
		self.rl.Lock()
		id := self.idGenerator.UseID()
		self.rl.Unlock()
		room = NewMajiangRoom(id, roomType)
		logger.Info("创建一个新的房间： ", id)

		self.rl.Lock()
		if _, ok := self.rooms[roomType]; !ok {
			self.rooms[roomType] = make([]*MaJiangRoom, 0)
		}
		self.rooms[roomType] = append(self.rooms[roomType], room)
		self.rl.Unlock()
	}

	player := NewMaJiangPlayer(playerID, msg)
	logger.Info("MaJiangRoomMgr.EnterGame: create one palyer (%s), ready enter room:", player.id)

	pip := cmn.RoomMsgQueue{
		Msg:  player,
		Func: "Enter",
	}
	room.rcv <- pip

	self.rl.Lock()
	self.playerInRoom[player.id] = room
	self.onlinePlayer[roomType] += 1
	self.rl.Unlock()

	logger.Info("MaJiangRoomMgr.EnterGame: player:%s enter room: Online:%d", player.id, self.onlinePlayer[roomType])

	return
}

//leave room
func (self *MaJiangRoomMgr) LeaveGame(uid string, isChangeDesk bool) error {
	self.rl.RLock()
	room := self.playerInRoom[uid]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("MaJiangRoomMgr LeaveGame error, cant find room uid:%s", uid)
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  uid,
		Msg2: isChangeDesk,
		Func: "Leave",
	}
	room.rcv <- pip

	return nil
}

//修改离开的玩家数据
func (self *MaJiangRoomMgr) DeleteLeavePlayerInfo(rtype int32, uid string) {
	// self.rl.Lock()

	// defer self.rl.Unlock()

	logger.Error("开始删除一个玩家在房间管理的关系。")
	self.rl.RLock()
	room := self.playerInRoom[uid]
	self.rl.RUnlock()
	if room != nil {
		isDeleteRoom := room.IsEmpty()
		if isDeleteRoom {
			isDeleted := false
			self.rl.Lock()
			for k, rooms := range self.rooms {
				if rooms == nil {
					continue
				}

				for i, r := range rooms {
					logger.Error("遍历房间删除一个玩家在房间管理的关系。", r.UID(), room.UID())
					if r.uid == room.uid {
						self.rooms[k] = append(self.rooms[k][:i], self.rooms[k][i+1:]...)
						self.idGenerator.RecoverID(room.UID())
						isDeleted = true
						logger.Info("删除了一个房间:", room.UID())
						break
					}
				}
				if isDeleted {
					break
				}
			}
			self.rl.Unlock()
		}
	}

	self.rl.Lock()
	delete(self.playerInRoom, uid)
	if self.onlinePlayer[rtype] > 0 {
		self.onlinePlayer[rtype] -= 1
	}
	self.rl.Unlock()

}

//踢人
func (self *MaJiangRoomMgr) ForceLeaveRoom(uid string, kickUid string) error {
	self.rl.RLock()
	room := self.playerInRoom[kickUid]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("MaJiangRoomMgr LeaveGame error, cant find room uid:%s", kickUid)
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  kickUid,
		Msg2: false,
		Func: "Kick",
	}
	room.rcv <- pip

	// self.rl.Lock()
	// delete(self.playerInRoom, kickUid)
	// if self.onlinePlayer[room.rtype] > 0 {
	// 	self.onlinePlayer[room.rtype] -= 1
	// }
	// self.rl.Unlock()

	// //通知gameserver道具
	// if err := conn.SendCostResourceMsg(uid, strconv.Itoa(cmn.KickCardID), "majiang", -1); err != nil {
	// 	logger.Error("发送扣取踢人卡出错：", err)
	// }

	return nil
}

func (self *MaJiangRoomMgr) ActionGame(msg *rpc.ActionREQ) error {
	self.rl.RLock()
	room := self.playerInRoom[msg.GetPlayerID()]
	self.rl.RUnlock()

	if room == nil {
		logger.Error("MaJiangRoomMgr ActionGame error, cant find room uid:%s", msg.GetPlayerID())
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  msg,
		Func: "ActionREQ",
	}
	room.rcv <- pip

	return nil
}

func (self *MaJiangRoomMgr) TieGuiREQ(msg *rpc.MJTieGuiREQ) error {
	self.rl.RLock()
	room := self.playerInRoom[msg.GetPlayerID()]
	self.rl.RUnlock()

	if room == nil {
		logger.Error("MaJiangRoomMgr ActionGame error, cant find room uid:%s", msg.GetPlayerID())
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  msg,
		Func: "TieGuiREQ",
	}
	room.rcv <- pip

	return nil
}

//获取房间数量指定房间类型的
func (self *MaJiangRoomMgr) GetRoomAmount(roomType int32) int {
	self.rl.RLock()
	defer self.rl.RUnlock()

	if roomLst, ok := self.rooms[roomType]; ok {
		return len(roomLst)
	} else {
		return 0
	}
}

//获取在线人数
func (self *MaJiangRoomMgr) GetOnlineNum() *rpc.OnlineInfo {
	self.rl.RLock()
	defer self.rl.RUnlock()

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
func (self *MaJiangRoomMgr) IsInRoom(uid string) bool {
	self.rl.RLock()
	defer self.rl.RUnlock()

	if val, exist := self.playerInRoom[uid]; exist && val != nil {
		return true
	}

	return false
}

func (self *MaJiangRoomMgr) SendDeskChatMsg(msg *rpc.FightRoomChatNotify) bool {
	self.rl.RLock()
	defer self.rl.RUnlock()

	room := self.playerInRoom[msg.GetPlayerID()]
	if room == nil {
		logger.Error("MaJiangRoomMgr SendDeskChatMsg error, cant find room uid:%s", msg.GetPlayerID())
		return false
	}

	room.SendCommonMsg2Others(msg)
	return false
}
