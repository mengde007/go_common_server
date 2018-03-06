package connector

import (
	cmn "common"
	"daerclient"
	"logger"
	"roomclient"
	"rpc"
)

// conn:这个参数不能变
// msg:客户端传上来的proto结构
func (self *CNServer) QuickEnterRoomREQ(conn rpc.RpcConn, msg rpc.QuickEnterRoomREQ) error {
	logger.Info("client call EnterRoomREQ begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	daerclient.QuickEnterRoom(p.PlayerBaseInfo, &msg)
	return nil
}

func (self *CNServer) LeaveRoomREQ(conn rpc.RpcConn, msg rpc.LeaveRoomREQ) error {
	logger.Info("client call LeaveRoomREQ begin")
	_, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	daerclient.LeaveDaerRoom(&msg)
	return nil
}

func (self *CNServer) ActionREQ(conn rpc.RpcConn, msg rpc.ActionREQ) error {
	logger.Info("client called ActionREQ begin")
	_, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	//根据不同类型调用相应的客服端函数
	switch msg.GetSysType() {
	case cmn.PiPeiFang:
		daerclient.DaerActionREQ(&msg)
	case cmn.ZiJianFang:
		roomclient.ActionREQ(&msg)
	case cmn.BiSaiFang:
	default:
		logger.Error("未知的系统类型")
	}

	return nil
}
