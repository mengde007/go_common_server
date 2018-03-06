package connector

import (
	"daerclient"
	"logger"
	"majiangclient"
	"pockerclient"
	"rpc"
)

//匹配房公共
func (self *CNServer) EnterRoomREQ(conn rpc.RpcConn, msg rpc.EnterRoomREQ) error {
	logger.Info("client call EnterRoomREQ begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	// p.whichGame["daer"] = msg.GetRoomType()
	p.SetGameType(msg.GetGameType())
	p.SetRoomType(msg.GetRoomType())

	switch msg.GetGameType() {
	case "1":
		daerclient.EnterDaerRoom(p.PlayerBaseInfo, &msg)
	case "2":
		majiangclient.EnterMaJiangRoom(p.PlayerBaseInfo, &msg)
	case "3":
		pockerclient.EnterPockerRoom(p.PlayerBaseInfo, msg.GetGameType(), msg.GetRoomType())
	default:
		logger.Error("未知的游戏类型")
	}

	return nil
}

//在线人数
func (self *CNServer) GetOnlineInfo(conn rpc.RpcConn, msg rpc.OnlinePlayerReq) error {
	logger.Info("GetOnlineInfo has been called begain")
	_, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	onlineInfo := &rpc.OnlinePlayerMsg{}
	for _, id := range msg.PartIds {
		if id == int32(1) { //大二
			daerInfo, err := daerclient.GetDaerServrOnlineNum()
			if daerInfo == nil {
				logger.Error("after login call daerclient.GetDaerServrOnlineNum() return nil, err:%s", err)
				return nil
			}
			onlineInfo.SetDaerInfo(daerInfo)
		} else if id == int32(2) { //麻将
			majiangInfo, err := majiangclient.GetMaJiangServrOnlineNum()
			if majiangInfo == nil {
				logger.Error("after login call majiangclient.GetMaJiangServrOnlineNum() return nil, err:%s", err)
				return nil
			}
			onlineInfo.SetDaerInfo(majiangInfo)
		} else if id == int32(3) { ///德州扑克

		} else {
			logger.Error("GetOnlineInfo client param err:%d", id)
		}
	}

	WriteResult(conn, onlineInfo)

	return nil
}
