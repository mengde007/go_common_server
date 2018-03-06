package connector

import (
	"common"
	// "errors"
	// "friendclient"
	"lockclient"
	"logger"
	"proto"
	"rpc"
)

//顶号
func (self *CenterService) LoginKickPlayer(req *proto.LoginKickPlayer, rst *proto.LoginKickPlayerResult) error {
	logger.Info("LoginKickPlayer: Begin!!!", req.Id)
	defer logger.Info("LoginKickPlayer: End!!!", req.Id)

	p, ok := cns.getPlayerByUid(req.Id)
	if ok && p.conn != nil {
		// p.conn.Lock()
		msg := &rpc.NotifyMsg{}
		msg.SetTxtId("21")
		WriteResult(p.conn, msg)
		err := p.conn.Close()
		if err == nil {
			rst.Success = true
		}
		// p.conn.Unlock()
		p.Save()
		return err
	}

	logger.Error("getPlayerByUid error: no player", req.Id)
	//没有这个玩家就强制解锁
	result, err := lockclient.ForceUnLock(common.LockName_Player, req.Id)
	if err != nil {
		rst.Success = false
		return err
	}
	rst.Success = result
	return nil
}

//改变最大在线人数
func (s *CenterService) SetMaxOnlineNumbers(req *proto.SetMaxOnlinePlayers, rst *proto.SetMaxOnlinePlayersRst) error {
	cns.maxPlayerCount = int32(req.Numbers)
	rst.CurNumbers = cns.curPlayerCount

	return nil
}

// 添加删除好友通知
func (self *CenterService) NotifyAddDelFriend(req *proto.FriendNoticeUpdate, rst *proto.FriendNoticeUpdateRst) error {
	p, ok := cns.getPlayerByUid(req.Uid)
	if !ok || p == nil {
		return nil
	}

	if p.conn != nil {
		p.conn.Lock()
	} else {
		return nil
	}

	p.queryAddDelFriendInfo(true)

	if p.conn != nil {
		p.conn.Unlock()
	}

	return nil
}

// func (self *CenterService) SendMsg2Player(req *proto.ChatSendMsg2Player, rst *proto.ChatSendMsg2PlayerResult) error {
// 	for sid, uids := range self.getOnGasList(req.PlayerList) {
// 		if conn, ok := self.cnss[sid]; ok {
// 			reqGas := &proto.ChatSendMsg2Player{
// 				MsgName:    req.MsgName,
// 				PlayerList: uids,
// 				Buf:        req.Buf,
// 			}
// 			rstGas := &proto.ChatSendMsg2PlayerResult{}
// 			conn.Go("CenterService.SendMsg2Player", reqGas, rstGas, nil)
// 		}
// 	}

// 	return nil
// }

// 玩家支付成功通知
func (self *CenterService) NotifyPlayerGetPayInfo(req *proto.NotifyPlayerGetPayInfo, rst *proto.NotifyPlayerGetPayInfo) error {
	p, ok := cns.getPlayerByUid(req.Uid)
	if !ok || p == nil {
		return nil
	}
	if p.conn != nil {
		p.conn.Lock()
	} else {
		return nil
	}

	// p.queryPayInfo()

	if p.conn != nil {
		p.conn.Unlock()
	}
	return nil
}
