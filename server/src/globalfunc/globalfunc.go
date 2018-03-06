package globalfunc

import (
	cmn "common"
	ds "daerserver"
	"logger"
	mj "majiangserver"
	"rpc"
)

func NewRoom(id, gameType int32, rtype int32) cmn.GameRoom {
	switch gameType {
	case cmn.DaerGame:
		return ds.NewDaerRoom(id, rtype)
	case cmn.MaJiang:
		return mj.NewMajiangRoom(id, rtype)
	case cmn.DeZhouPuker:
	default:
		logger.Error("不能识别的游戏类型！")
	}

	return nil
}

func NewPlayer(gameType int32, id string, playerInfo *rpc.PlayerBaseInfo) cmn.Player {
	switch gameType {
	case cmn.DaerGame:
		return ds.NewDaerPlayer(id, playerInfo)
	case cmn.MaJiang:
		return mj.NewMaJiangPlayer(id, playerInfo)
	case cmn.DeZhouPuker:
	default:
		logger.Error("不能识别的游戏类型！")
	}

	return nil
}
