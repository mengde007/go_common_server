package connector

import (
	gp "code.google.com/p/goprotobuf/proto"
	"rpc"
	// "time"
)

const (
	CHAT_MSG_NUM = 100
)

type stChatMsg struct {
	sendTime uint32
	msgIndex uint32
	msg      gp.Message
}

func (self *CNServer) GetWorldChatMsg(conn rpc.RpcConn, msg rpc.Ping) error {
	// p, exist := self.getPlayerByConnId(conn.GetId())
	// if !exist {
	// 	return nil
	// }

	// self.chatLock.RLock()
	// defer self.chatLock.RUnlock()
	// sync := false
	// timeNow := uint32(time.Now().Unix())
	// messages := &rpc.S2CChatWorldMessages{}
	// for _, info := range self.chatMsgs {
	// 	if info.msgIndex > p.lastChatIndex || info.sendTime > p.lastGetChatTime {
	// 		sync = true
	// 	}
	// 	if sync {
	// 		// WriteResult(conn, info.msg)
	// 		if message, ok := info.msg.(*rpc.S2CChatWorld); ok {
	// 			message.SetMessageId(info.msgIndex)
	// 			messages.Msg = append(messages.Msg, message)
	// 			p.lastChatIndex = info.msgIndex
	// 		}
	// 	}
	// }
	// WriteResult(conn, messages)
	// p.lastGetChatTime = timeNow

	return nil
}
