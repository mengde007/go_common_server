package connector

import (
	"common"
	"logger"
	"pockerclient"
	"roomclient"
	"rpc"
	"strconv"
)

// conn:请求进入自建房间
func (self *CNServer) EnterCustomRoom(conn rpc.RpcConn, msg rpc.EnterCustomRoomREQ) error {
	logger.Info("client call EnterCustomRoomREQ begin, gameType:%s", msg.GetGameType())
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	p.SetGameType(msg.GetGameType())
	p.SetRoomType(msg.GetId())

	if msg.GetGameType() == "6" {
		pockerclient.EnterPockerRoom(p.PlayerBaseInfo, msg.GetGameType(), msg.GetId())
	} else {
		roomclient.EnterRoom(p.PlayerBaseInfo, &msg)
	}

	return nil
}

// conn:请求离开自建房间
func (self *CNServer) LeaveCustomRoom(conn rpc.RpcConn, msg rpc.LeaveCustomRoomREQ) error {
	logger.Info("client call LeaveCustomRoom begin")
	// p, exist := self.getPlayerByConnId(conn.GetId())
	// if !exist {
	// 	return nil
	// }

	roomclient.LeaveRoom(&msg)
	return nil
}

// conn:请求创建自建房间
func (self *CNServer) CreateCustomRoom(conn rpc.RpcConn, msg rpc.CreateRoomREQ) error {
	logger.Info("client call CreateCustomRoom begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	itemAmount := p.GetItemNum(strconv.Itoa(common.CustomRoomCardID))
	logger.Info("%s道具数量:%s", common.CustomRoomCardID, itemAmount)
	roomclient.CreateRoomREQ(p.PlayerBaseInfo, &msg, itemAmount)
	return nil
}

// conn:请求进入自建房间
func (self *CNServer) ObtainRoomList(conn rpc.RpcConn, msg rpc.RoomListREQ) error {
	logger.Info("client call ObtainRoomList begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	roomclient.RoomListREQ(p.PlayerBaseInfo.GetUid(), &msg)
	return nil
}

// conn:请求查找房间
func (self *CNServer) FindRoom(conn rpc.RpcConn, msg rpc.FindRoomREQ) error {
	logger.Info("client call FindRoom begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	roomclient.FindRoomREQ(p.PlayerBaseInfo.GetUid(), &msg)
	return nil
}

// conn:请求解散房间
func (self *CNServer) JieSanRoom(conn rpc.RpcConn, msg rpc.JieSanRoomREQ) error {
	logger.Info("client call JieSanRoom begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	roomclient.JieSanRoomREQ(p.PlayerBaseInfo.GetUid(), &msg)
	return nil
}
