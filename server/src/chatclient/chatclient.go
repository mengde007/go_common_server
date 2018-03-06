package chatclient

import (
	// "clanclient"
	gp "code.google.com/p/goprotobuf/proto"
	"common"
	"logger"
	"proto"
	"rpc"
	"rpcplusclientpool"
	// "time"
)

var pPoll *rpcplusclientpool.ClientPool

func init() {
	aServerHost := common.ReadServerClientConfig("centerserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "centerserver -> chat")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	return
}

//世界聊天
func P2WChat(fuid, fname string, flevel uint32, content, cname string, csymb uint32, lastRank uint32, vipLevel uint32, useIM bool, voiceTime string) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	req := &proto.PlayerWorldChat{
	// FromPlayerId:    fuid,
	// FromPlayerName:  fname,
	// FromPlayerLevel: flevel,
	// Content:         content,
	// CName:           cname,
	// CSymbol:         csymb,
	// LastLeagueRank:  lastRank,
	// Viplevel:        vipLevel,
	// UseIM:           useIM,
	// VoiceTime:       voiceTime,
	}
	rst := &proto.PlayerWorldChatResult{}

	conn.Go("Center.PlayerWorldChat", req, rst, nil)

	return nil
}

//玩家对玩家的聊天
func P2PChat(fromuid, fromname string, fromlevel uint32, touid, content string) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	req := &proto.PlayerChatToPlayer{
		FromPlayerId:   fromuid,
		FromPlayerName: fromname,
		// FromPlayerLevel: fromlevel,
		ToPlayerId: touid,
		Content:    content,
	}
	rst := &proto.PlayerChatToPlayerResult{}
	conn.Go("Center.PlayerChatToPlayer", req, rst, nil)

	return nil
}

//发送消息
func SendMsg2Player(desp []string, msg gp.Message, msgdes string, bgo bool) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ChatSendMsg2Player{
		MsgName:    msgdes,
		PlayerList: desp,
		Buf:        buf,
	}
	rst := &proto.ChatSendMsg2PlayerResult{}

	if bgo {
		conn.Go("Center.SendMsg2Player", req, rst, nil)
		return nil
	} else {
		return conn.Call("Center.SendMsg2Player", req, rst)
	}

	return nil
}

//发送轮播
func SendMsgBroadcastPlayer(msg gp.Message, Func string, bgo bool) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	buf, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.ChatSendMsg2LPlayer{
		MsgName: Func,
		Buf:     buf,
	}
	rst := &proto.ChatSendMsg2LPlayerResult{}

	if bgo {
		conn.Go("Center.SendMsg2LocationPlayer", req, rst, nil)
		return nil
	} else {
		return conn.Call("Center.SendMsg2LocationPlayer", req, rst)
	}

	return nil
}

func SendCodeMsg(uid, code string) error {
	uids := make([]string, 0)
	uids = append(uids, uid)

	msg := &rpc.Msg{}
	msg.SetCode(code)

	SendMsg2Player(uids, msg, "Msg", true)

	return nil
}
