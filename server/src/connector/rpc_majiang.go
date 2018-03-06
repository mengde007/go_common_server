package connector

import (
	cmn "common"
	"logger"
	"majiangclient"
	"roomclient"
	"rpc"
)

// conn:这个参数不能变
// msg:客户端传上来的proto结构
func (self *CNServer) MJQuickEnterRoomREQ(conn rpc.RpcConn, msg rpc.MJQuickEnterRoomREQ) error {
	logger.Info("client call MJQuickEnterRoomREQ begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	majiangclient.QuickEnterMaJiangRoom(p.PlayerBaseInfo, &msg)
	return nil
}

func (self *CNServer) MJLeaveRoomREQ(conn rpc.RpcConn, msg rpc.MJLeaveRoomREQ) error {
	logger.Info("client call MJLeaveRoomREQ begin")
	_, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	majiangclient.LeaveMaJiangRoom(&msg)
	return nil
}

func (self *CNServer) MJActionREQ(conn rpc.RpcConn, msg rpc.ActionREQ) error {
	logger.Info("client called MJActionREQ begin")
	_, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	//根据不同类型调用相应的客服端函数
	switch msg.GetSysType() {
	case cmn.PiPeiFang:
		majiangclient.MaJiangActionREQ(&msg)
	case cmn.ZiJianFang:
		roomclient.ActionREQ(&msg)
	case cmn.BiSaiFang:
	default:
		logger.Error("未知的系统类型")
	}

	return nil
}

func (self *CNServer) MJTieGuiREQ(conn rpc.RpcConn, msg rpc.MJTieGuiREQ) error {
	logger.Info("client called MJTieGuiREQ begin")
	_, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	//根据不同类型调用相应的客服端函数
	switch msg.GetSysType() {
	case cmn.PiPeiFang:
		majiangclient.MaJiangTieGuiREQ(&msg)
	case cmn.ZiJianFang:
		roomclient.MaJiangTieGuiREQ(&msg)
	case cmn.BiSaiFang:
	default:
		logger.Error("未知的系统类型")
	}

	return nil
}
