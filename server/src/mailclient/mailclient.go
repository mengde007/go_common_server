package mailclient

import (
	"common"
	// "errors"
	"logger"
	"proto"
	"rpc"
	"rpcplusclientpool"
)

var pPoll *rpcplusclientpool.ClientPool

func init() {
	aServerHost := common.ReadServerClientConfig("mailserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "mailserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	return
}

//发系统邮件
func SendAllMail(title, content, attach string, channel, continuetime uint32) (error, bool) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err, false
	}

	req := &proto.MailSendAll{
		Title:        title,
		Content:      content,
		Attach:       attach,
		Channel:      channel,
		ContinueTime: continuetime,
	}
	rst := &proto.MailSendAllResult{}

	if err := conn.Call("MailServer.SendAllMail", req, rst); err != nil {
		return err, false
	}

	return nil, rst.Success
}

//发送系统个人邮件SendMonthCardMail2Player
func SendSysMail2Player(uid, title, content, attach string, validtime uint32, bgo bool) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	req := &proto.SendSystemMail{
		ToPlayerId: uid,
		Title:      title,
		Content:    content,
		Attach:     attach,
		ValidTime:  validtime,
	}
	rst := &proto.SendSystemMailResult{}

	if bgo {
		conn.Go("MailServer.SendMail2Player", req, rst, nil)
		return nil
	} else {
		return conn.Call("MailServer.SendMail2Player", req, rst)
	}

	return nil
}

//发玩家个人邮件
func SendPlayerMail2Player(req *proto.SendPlayerMail, bgo bool) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	rst := &proto.SendPlayerMailResult{}

	if bgo {
		conn.Go("MailServer.SendMail2Player", req, rst, nil)
		return nil
	} else {
		return conn.Call("MailServer.SendMail2Player", req, rst)
	}

	return nil
}

////////////////////////
//邮件操作
////////////////////////
//取全部邮件
func GetAllMail(uid string) (error, *rpc.PlayerMailInfo) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err, nil
	}

	req := &proto.MailQueryAll{
		PlayerId: uid,
		// Channel:  gl,
	}
	rst := &proto.MailQueryAllResult{}
	if err := conn.Call("MailServer.GetPlayerAllMail", req, rst); err != nil {
		return err, nil
	}

	info := &rpc.PlayerMailInfo{}
	if err := common.DecodeMessage(rst.Values, info); err != nil {
		return err, nil
	}

	return nil, info
}

//删除邮件
func DeleteMail(uid, mailid string) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	req := &proto.DelPlayerMail{
		PlayerId: uid,
		MailId:   mailid,
	}
	rst := &proto.DelPlayerMailResult{}

	return conn.Call("MailServer.PlayerDeleteMail", req, rst)
}

//标记读取邮件
func ReadMail(uid, mailid string) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	req := &proto.ReadPlayerMail{
		PlayerId: uid,
		MailId:   mailid,
	}
	rst := &proto.ReadPlayerMailResult{}

	conn.Go("MailServer.PlayerReadMail", req, rst, nil)

	return nil
}

//取附件
func GetMailAttach(uid, mailid string) (string, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return "", err
	}

	req := &proto.GetMailAttach{
		PlayerId: uid,
		MailId:   mailid,
	}
	rst := &proto.GetMailAttachResult{}

	if err := conn.Call("MailServer.PlayerGetAttach", req, rst); err != nil {
		return "", err
	}

	return rst.Attach, nil
}
