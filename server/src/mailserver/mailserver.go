package mailserver

import (
	"chatclient"
	"common"
	"dbclient"
	"errors"
	"github.com/garyburd/redigo/redis"
	"logger"
	"net"
	"proto"
	"rpc"
	"rpcplus"
	"runtime/debug"
	"sync"
	"time"
)

type MailServer struct {
	pCachePool *common.CachePool
	mails      []*rpc.SysMail
	curversion int32
	l          sync.RWMutex
	playerl    *common.SimpleLockService
	snl        *common.SimpleLockService //系统公告锁
	sl         sync.RWMutex
	// pAllSysNotice *AllSysNotice
}

var pServer *MailServer

func CreateServices(cfg common.MailConfig, listener net.Listener) *MailServer {
	//加载配置表
	common.LoadGlobalConfig()

	pServer = &MailServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
		mails:      make([]*rpc.SysMail, 0),
		curversion: int32(0),
		playerl:    common.CreateSimpleLock(),
		snl:        common.CreateSimpleLock(),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)

	//系统邮件初始化
	if err := pServer.initSysMail(); err != nil {
		logger.Fatal("%s", err.Error())
		return nil
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			logger.Info("StartServices %s", err.Error())
			break
		}

		//开始对cns的RPC服务
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("MailServer Rpc Runtime Error: %s", r)
					debug.PrintStack()
				}
			}()
			rpcServer.ServeConn(conn)
		}()
	}

	return pServer
}

//初始化
func (self *MailServer) initSysMail() error {
	cache := self.pCachePool.Get()
	defer cache.Recycle()

	self.l.Lock()
	defer self.l.Unlock()

	mailsdata, err := cache.Do("SMEMBERS", common.GetSystemTableKey_Mail())
	if err != nil {
		logger.Error("init sys mail failed")
		return err
	}

	arrmails, err := redis.Values(mailsdata, err)
	if err != nil {
		logger.Error("array sys mail failed")
		return err
	}

	uMaxVersion := int32(0)
	for _, buf := range arrmails {
		bytes, err := redis.Bytes(buf, nil)
		if err != nil {
			logger.Error("Bytes sys mail failed")
			return err
		}

		var mail *rpc.SysMail = new(rpc.SysMail)
		if err = common.DecodeMessage(bytes, mail); err != nil {
			logger.Error("decode sys mail failed")
			return err
		}

		self.mails = append(self.mails, mail)

		if mail.GetVersion() > uMaxVersion {
			uMaxVersion = mail.GetVersion()
		}
	}
	self.curversion = uMaxVersion
	return nil
}

//发全服系统邮件
func (self *MailServer) addSysMail(mail *rpc.SysMail) error {
	// if err := common.CheckMailAttach(mail.GetAttach()); err != nil {
	// 	logger.Error("sys send mail attach error:%v", err, mail.GetAttach())
	// 	return err
	// }

	cache := self.pCachePool.Get()
	defer cache.Recycle()

	self.l.Lock()
	defer self.l.Unlock()

	//自动设置版本号
	self.curversion++
	mail.SetVersion(self.curversion)
	self.mails = append(self.mails, mail)

	vaule, err := common.EncodeMessage(mail)
	if err != nil {
		logger.Error("EncodeMessage sys mail failed, err:%s", err)
		return err
	}

	_, err = cache.Do("SADD", common.GetSystemTableKey_Mail(), vaule)
	if err != nil {
		logger.Error("add sys mail failed")
		return err
	}
	//暂时不向玩家推送，因为与玩家不是一个锁
	//self.pushMail2Online(mail, self.curversion)

	return nil
}

func (self *MailServer) SendAllMail(req *proto.MailSendAll, rst *proto.MailSendAllResult) error {
	logger.Info("******************SendAllMail has been called ")
	timeCur := int32(time.Now().Unix())
	pMail := &rpc.SysMail{}
	pMail.SetMailId(common.GenMailId())
	pMail.SetTitle(req.Title)
	pMail.SetContent(req.Content)
	pMail.SetSendtime(timeCur)
	pMail.SetAttach(req.Attach)
	// pMail.SetChannelid(rpc.GameLocation(req.Channel))
	pMail.SetOverduetime(timeCur + int32(req.ContinueTime))

	err := self.addSysMail(pMail)
	if err == nil {
		rst.Success = true
	} else {
		rst.Success = false
	}

	return err
}

//向玩家发送邮件
func (self *MailServer) SendMail2Player(info *proto.SendPlayerMail, result *proto.SendPlayerMailResult) (err error) {
	logger.Info("MailServer.SendMail2Player", info, result)

	self.playerl.WaitLock(info.ToPlayerId)
	defer self.playerl.WaitUnLock(info.ToPlayerId)

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := dbclient.KVQueryExt(common.TB_t_ext_playermail, info.ToPlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("send mail wrong player")
		}

		maxMail := uint32(100)
		numbers := len(mailinfo.Maillist)
		if numbers >= int(maxMail) {
			return errors.New("send mail number limit")
		}

		timeNow := int32(time.Now().Unix())
		mailSend := &rpc.AddMailNotify{}
		mail := &rpc.SysMail{}
		mail.SetMailId(common.GenMailId())
		mail.SetVersion(int32(0))
		mail.SetTitle(info.Title)
		mail.SetContent(info.Content)
		mail.SetAttach(info.Attach)
		mail.SetSendtime(timeNow)
		mail.SetOverduetime(timeNow + int32(info.ValidTime*3600))
		mail.SetBRead(false)
		mailinfo.Maillist = append(mailinfo.Maillist, mail)
		mailSend.Maillist = append(mailSend.Maillist, mail)

		//存储
		ok, err := dbclient.KVWriteExt(common.TB_t_ext_playermail, info.ToPlayerId, mailinfo)
		if err == nil && ok {
			uids := make([]string, 0)
			uids = append(uids, info.ToPlayerId)
			chatclient.SendMsg2Player(uids, mailSend, "AddMailNotify", true)

			//发送者消息
			// chatclient.SendCodeMsg(info.FromUid, "TID_MAIL_SEND_SUCCESS")
		}
	}

	return err
}

//玩家初始化邮件系统
func (self *MailServer) GetPlayerAllMail(req *proto.MailQueryAll, rst *proto.MailQueryAllResult) error {
	logger.Info("****************GetPlayerAllMail， playerid:%s", req.PlayerId)
	sPlayerId := req.PlayerId

	//锁玩家自己
	self.playerl.WaitLock(sPlayerId)
	defer self.playerl.WaitUnLock(sPlayerId)

	mailinfo := &rpc.PlayerMailInfo{}
	exist, err := dbclient.KVQueryExt(common.TB_t_ext_playermail, sPlayerId, mailinfo)
	if err != nil {
		return err
	}

	if !exist {
		logger.Info("***********can't find player")
		mailinfo = &rpc.PlayerMailInfo{}
	}

	//去掉过期的
	cutTime := int32(time.Now().Unix())
	index := 0
	for {
		if index >= len(mailinfo.Maillist) {
			break
		}

		ot := mailinfo.Maillist[index].GetOverduetime()
		if ot > 0 && ot < cutTime {
			mailinfo.Maillist = append(mailinfo.Maillist[:index], mailinfo.Maillist[index+1:]...)
		} else {
			index++
		}
	}

	//取系统邮件
	mails, version := self.playerPickupSysMail(mailinfo.GetSysmailVersion())
	mailinfo.Maillist = append(mailinfo.Maillist, mails[:]...)
	mailinfo.SetSysmailVersion(version)

	logger.Info("******************初始化邮件，剩余邮件:%d, version:%d", len(mailinfo.Maillist), mailinfo.GetSysmailVersion())

	//存储
	ok, err := dbclient.KVWriteExt(common.TB_t_ext_playermail, sPlayerId, mailinfo)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("save player mail failed !")
	}

	rst.Values, err = common.EncodeMessage(mailinfo)
	if err != nil {
		return err
	}

	return nil
}

//玩家取系统邮件
func (self *MailServer) playerPickupSysMail(uVersion int32) (mails []*rpc.SysMail, curversion int32) {
	logger.Info("***********playerPickupSysMail")
	//读系统邮件
	self.l.RLock()
	defer self.l.RUnlock()

	cutTime := int32(time.Now().Unix())
	for _, mail := range self.mails {
		logger.Info("************** mail.GetVersion():%d > uVersion:%d ", mail.GetVersion(), uVersion)
		if (mail.GetOverduetime() == 0 || cutTime < mail.GetOverduetime()) && mail.GetVersion() > uVersion {
			pm := &rpc.SysMail{}
			pm.SetMailId(mail.GetMailId())
			pm.SetVersion(mail.GetVersion())
			pm.SetTitle(mail.GetTitle())
			pm.SetContent(mail.GetContent())
			pm.SetSendtime(mail.GetSendtime())
			pm.SetAttach(mail.GetAttach())
			pm.SetOverduetime(mail.GetOverduetime())
			pm.SetBRead(false)
			mails = append(mails, pm)
		}
	}

	logger.Info("******************mails len:%d", len(mails))
	curversion = self.curversion

	return
}

//玩家读取邮件
func (self *MailServer) PlayerReadMail(info *proto.ReadPlayerMail, result *proto.ReadPlayerMailResult) (err error) {
	self.playerl.WaitLock(info.PlayerId)
	defer self.playerl.WaitUnLock(info.PlayerId)

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := dbclient.KVQueryExt(common.TB_t_ext_playermail, info.PlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("no mail")
		}

		for _, mail := range mailinfo.Maillist {
			if mail.GetMailId() == info.MailId {
				//删除
				mail.SetBRead(true)
			}
		}
		ok, err := dbclient.KVWriteExt(common.TB_t_ext_playermail, info.PlayerId, mailinfo)
		if err == nil && ok {
			return nil
		}
		logger.Info("******************阅读邮件，剩余邮件:%d, version:%d", len(mailinfo.Maillist), mailinfo.GetSysmailVersion())
	}

	return errors.New("del mail failed")
}

//玩家取附件
func (self *MailServer) PlayerGetAttach(info *proto.GetMailAttach, result *proto.GetMailAttachResult) (err error) {
	self.playerl.WaitLock(info.PlayerId)
	defer self.playerl.WaitUnLock(info.PlayerId)

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := dbclient.KVQueryExt(common.TB_t_ext_playermail, info.PlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("no mail 1")
		}

		for index, mail := range mailinfo.Maillist {
			if mail.GetMailId() == info.MailId {
				//过期了？
				cutTime := int32(time.Now().Unix())
				ot := mail.GetOverduetime()
				if ot == 0 || ot > cutTime {
					result.Attach = mail.GetAttach()
				} else {
					logger.Error("mail out of time:", ot, cutTime)
				}

				//存储
				//直接删除
				mailinfo.Maillist = append(mailinfo.Maillist[:index], mailinfo.Maillist[index+1:]...)
				logger.Info("******************取附件，剩余邮件:%d, version:%d", len(mailinfo.Maillist), mailinfo.GetSysmailVersion())
				ok, err := dbclient.KVWriteExt(common.TB_t_ext_playermail, info.PlayerId, mailinfo)
				if err == nil && ok {
					return nil
				}
				return errors.New("save mail error")
			}
		}
		return errors.New("no mail 2")
	}
	return
}

//************************************************************************old

//玩家删除邮件
func (self *MailServer) PlayerDeleteMail(info *proto.DelPlayerMail, result *proto.DelPlayerMailResult) (err error) {
	self.playerl.WaitLock(info.PlayerId)
	defer self.playerl.WaitUnLock(info.PlayerId)

	mailinfo := &rpc.PlayerMailInfo{}
	if exist, err := dbclient.KVQueryExt(common.TB_t_ext_playermail, info.PlayerId, mailinfo); err == nil {
		if !exist {
			return errors.New("no mail")
		}

		for index, mail := range mailinfo.Maillist {
			if mail.GetMailId() == info.MailId {
				//删除
				mailinfo.Maillist = append(mailinfo.Maillist[:index], mailinfo.Maillist[index+1:]...)
				ok, err := dbclient.KVWriteExt(common.TB_t_ext_playermail, info.PlayerId, mailinfo)
				if err == nil && ok {
					return nil
				}

				return errors.New("del mail failed")
			}
		}

		//到这里表示成功了，没找到算了
		return nil
	}

	return errors.New("del mail failed")
}
