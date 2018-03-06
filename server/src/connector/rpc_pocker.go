package connector

import (
	"logger"
	"pockerclient"
	"rpc"
	// "strconv"
)

//进入德州自寻房 或 匹配房间
func (self *CNServer) EnterPockerRoomREQ(conn rpc.RpcConn, msg rpc.EnterRoomREQ) error {
	logger.Info("client call EnterPockerRoomREQ begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	// p.whichGame["daer"] = msg.GetRoomType()
	p.SetGameType(msg.GetGameType())
	p.SetRoomType(msg.GetRoomType())
	pockerclient.EnterPockerRoom(p.PlayerBaseInfo, msg.GetGameType(), msg.GetRoomType())
	return nil
}

func (self *CNServer) PockerAction(conn rpc.RpcConn, msg rpc.C2SAction) error {
	logger.Info("client call PockerAction begin, act:%d", msg.GetAct())
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	if msg.GetAct() == int32(3) { //ACT_CHANGE_DESK
		msg.SetBase(p.PlayerBaseInfo)
	}
	pockerclient.ReqAction(&msg)
	return nil
}

func (self *CNServer) QuickEnterPockerRoomREQ(conn rpc.RpcConn, msg rpc.QuickEnterRoomREQ) error {
	logger.Info("client call QuickEnterPockerRoomREQ begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	pockerclient.QuicklySeatdown(p.PlayerBaseInfo)
	return nil
}

//创建自寻房
func (self *CNServer) CreatePockerCustomRoom(conn rpc.RpcConn, msg rpc.CreatePockerRoomReq) error {
	logger.Info("client call CreateCustomRoom begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	_, roomNo := pockerclient.CreateCustomRoom(msg.GetBlindId(), msg.GetLimId(), p.PlayerBaseInfo)
	if roomNo != -1 {
		p.SetGameType("6")
		p.SetRoomType(roomNo)
	}

	return nil
}
