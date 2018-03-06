//room_mgr应该包含到daerserver包里面，在大二服务器启动的时候 创建RoomMgr对象并初始化

package daerserver

//package main

import (
	//conn "centerclient"
	cmn "common"
	//"fmt"
	"logger"
	"rpc"
	//"runtime/debug"
	//"strconv"
	//"time"
	"sync"
	"timer"
)

type DaerRoomMgr struct {
	rooms        map[int32][]*DaerRoom //保存所有的房间
	playerInRoom map[string]*DaerRoom  //玩家所在的房间
	t            *timer.Timer
	onlinePlayer map[int32]uint32 //在线玩家数量
	idGenerator  *cmn.RoomIDS     //ID产生器
	rl           sync.RWMutex     //读写锁
}

var daerRoomMgr *DaerRoomMgr

func (self *DaerRoomMgr) init() {
	self.rooms = make(map[int32][]*DaerRoom, 0)
	self.playerInRoom = make(map[string]*DaerRoom, 0)
	self.onlinePlayer = make(map[int32]uint32, 0)
	self.idGenerator = &cmn.RoomIDS{}
	self.idGenerator.Init(0, 1000)
}

func (self *DaerRoomMgr) getNotFullRoom(roomType int32) *DaerRoom {
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

// func (self *DaerRoomMgr) createTimer() {
// 	self.t = timer.NewTimer(time.Second)
// 	self.t.Start(
// 		func() {
// 			defer func() {
// 				if r := recover(); r != nil {
// 					fmt.Println("player tick runtime error :", r)
// 					debug.PrintStack()
// 				}
// 			}()

// 			self.OnTick()
// 		},
// 	)
// }

// func (self *DaerRoomMgr) OnTick() {
// 	//fmt.Println("OnTick time:%d",  time.Now().Unix())

// 	for _, rooms := range self.rooms {
// 		for _, room := range rooms {
// 			if !room.IsEmpty() {
// 				room.UpdateTimer()
// 			}
// 		}
// 	}

// }

//进入游戏
func (self *DaerRoomMgr) EnterGame(roomType int32, msg *rpc.PlayerBaseInfo, isChangeDesk bool) {
	//检测输入的参数
	if msg == nil {
		logger.Error("DaerRoomMgr.EnterGame: msg is nil.")
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

		logger.Info("DaerRoomMgr.EnterGame: palyer reEnter room:", roomType, msg.GetName())

		if room == nil {
			logger.Error("DaerRoomMgr:EnterGame, room is nil")
			return
		}

		//room.ReEnter(playerID, msg)
		pip := cmn.RoomMsgQueue{
			Msg:  msg,
			Func: "ReEnter",
		}
		room.rcv <- pip

		return
	}

	self.rl.RUnlock()

	//首次进入
	logger.Info("DaerRoomMgr.EnterGame: palyer enter room:", roomType, msg.GetName())
	// self.rl.Lock()
	// defer self.rl.Unlock()

	room := self.getNotFullRoom(roomType)
	if room == nil {
		self.rl.Lock()
		id := self.idGenerator.UseID()
		self.rl.Unlock()

		room = NewDaerRoom(id, roomType)

		self.rl.Lock()
		if _, ok := self.rooms[roomType]; !ok {
			self.rooms[roomType] = make([]*DaerRoom, 0)
		}
		self.rooms[roomType] = append(self.rooms[roomType], room)
		self.rl.Unlock()
	}

	player := NewDaerPlayer(playerID, msg)
	logger.Info("DaerRoomMgr.EnterGame: create one palyer (%s), ready enter room:", player.id)
	//room.Enter(player)

	pip := cmn.RoomMsgQueue{
		Msg:  player,
		Func: "Enter",
	}
	room.rcv <- pip

	self.rl.Lock()
	self.playerInRoom[player.id] = room
	self.onlinePlayer[roomType] += 1
	self.rl.Unlock()

	logger.Info("DaerRoomMgr.EnterGame: player:%s enter room: Online:%d", player.id, self.onlinePlayer[roomType])

	return
}

//leave room
func (self *DaerRoomMgr) LeaveGame(uid string, isChangeDesk bool) error {
	self.rl.RLock()
	room := self.playerInRoom[uid]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("DaerRoomMgr LeaveGame error, cant find room uid:%s", uid)
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  uid,
		Msg2: isChangeDesk,
		Func: "Leave",
	}
	room.rcv <- pip

	// leavePlayer := room.GetPlayerByID(uid)
	// if leavePlayer == nil {
	// 	logger.Error("在房间里没有查找到制定的玩家：", uid)
	// 	return nil
	// }

	// //离开房间
	// if room.Leave(uid, isChangeDesk) {
	// 	delete(self.playerInRoom, uid)

	// 	if self.onlinePlayer[room.rtype] > 0 {
	// 		self.onlinePlayer[room.rtype] -= 1
	// 	}
	// 	//如果是换桌
	// 	if isChangeDesk {
	// 		self.EnterGame(room.rtype, leavePlayer.client, isChangeDesk)
	// 	}
	// } else {
	// 	//调试阶段使用
	// 	//room.ResetRoom()
	// 	//room.SwitchRoomState(RSReady)
	// 	//room.ForceAllPlayerLe pave()
	// }

	return nil
}

//修改离开的玩家数据
func (self *DaerRoomMgr) DeleteLeavePlayerInfo(rtype int32, uid string) {
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
func (self *DaerRoomMgr) ForceLeaveRoom(uid string, kickUid string) error {
	self.rl.RLock()
	room := self.playerInRoom[kickUid]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("DaerRoomMgr LeaveGame error, cant find room uid:%s", kickUid)
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

	//离开房间
	// if room.Leave(kickUid, false) {
	// 	delete(self.playerInRoom, kickUid)

	// 	if self.onlinePlayer[room.rtype] > 0 {
	// 		self.onlinePlayer[room.rtype] -= 1
	// 	}

	// 	//通知gameserver道具
	// 	if err := conn.SendCostResourceMsg(uid, strconv.Itoa(cmn.KickCardID), "daer", -1); err != nil {
	// 		logger.Error("发送扣取踢人卡出错：", err)
	// 	}

	// } else {
	// 	//调试阶段使用
	// 	//room.ResetRoom()
	// 	//room.SwitchRoomState(RSReady)
	// 	//room.ForceAllPlayerLe pave()
	// }

	return nil
}

func (self *DaerRoomMgr) ActionGame(msg *rpc.ActionREQ) error {
	self.rl.RLock()
	room := self.playerInRoom[msg.GetPlayerID()]
	self.rl.RUnlock()
	if room == nil {
		logger.Error("DaerRoomMgr ActionGame error, cant find room uid:%s", msg.GetPlayerID())
		return nil
	}

	pip := cmn.RoomMsgQueue{
		Msg:  msg,
		Func: "ActionREQ",
	}
	room.rcv <- pip
	//room.OnPlayerDoAction(msg)
	return nil
}

//获取房间数量指定房间类型的
func (self *DaerRoomMgr) GetRoomAmount(roomType int32) int {
	self.rl.RLock()
	defer self.rl.RUnlock()

	if roomLst, ok := self.rooms[roomType]; ok {
		return len(roomLst)
	} else {
		return 0
	}
}

//获取在线人数
func (self *DaerRoomMgr) GetOnlineNum() *rpc.OnlineInfo {
	sendMsg := &rpc.OnlineInfo{}

	self.rl.RLock()
	defer self.rl.RUnlock()

	for id, num := range self.onlinePlayer {
		msg := &rpc.OnlineBody{}
		msg.SetRoomId(int32(id))
		msg.SetNum(int32(num))
		sendMsg.Info = append(sendMsg.Info, msg)
	}
	return sendMsg
}

//检查玩家是否在游戏中（只要在房间都标示在游戏，哪怕他是在结算阶段）
func (self *DaerRoomMgr) IsInRoom(uid string) bool {
	self.rl.RLock()
	defer self.rl.RUnlock()

	if val, exist := self.playerInRoom[uid]; exist && val != nil {
		return true
	}

	return false
}

func (self *DaerRoomMgr) SendDeskChatMsg(msg *rpc.FightRoomChatNotify) bool {
	self.rl.RLock()
	defer self.rl.RUnlock()

	room := self.playerInRoom[msg.GetPlayerID()]
	if room == nil {
		logger.Error("DaerRoomMgr SendDeskChatMsg error, cant find room uid:%s", msg.GetPlayerID())
		return false
	}

	room.SendCommonMsg2Others(msg)
	return false
}
