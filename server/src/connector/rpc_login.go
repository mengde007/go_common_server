package connector

import (
	"accountclient"
	"centerclient"
	"common"
	"encoding/json"
	"io/ioutil"
	"language"
	"lockclient"
	"logger"
	"mailclient"
	"net/http"
	"proto"
	"rankclient"
	"roleclient"
	"rpc"
	"runtime/debug"
	"strconv"
	"time"
)

const (
	wechat_register = iota
	wechat_login
	guest_register
	guest_login
)

//登陆错误处理函数
func (self *CNServer) LoginError(p *player) {
	//注销tick
	if p.t != nil {
		p.t.Stop()
	}

	// lockclient.TryUnlock(common.LockName_Player, p.GetUid(), p.lid)
	self.DisConnOnLogin(p.conn, p.GetUid(), p.lid, 0, uint32(0), true, true)
}

type VersionInfo struct {
	Version string `json:"cur_version"`
}

func (self *CNServer) vertify_version(login rpc.Login) bool {
	desiCfg := common.GetDesignerCfg()
	if desiCfg == nil {
		logger.Error("vertify_version 获取desinger.json出错")
		return false
	}
	url := desiCfg.VersionUrl
	logger.Info("vertify_version 请求地址 url:%s", url)

	client := &http.Client{
		Transport: createTransport(),
	}
	res, err := client.Get(url)
	if err != nil {
		logger.Error("vertify_version http.Post error: %v", err)
		return false
	}

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("vertify_version body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("vertify_version ioutil.ReadAll error: %v", err)
		return false
	}

	rst := VersionInfo{}
	if err := json.Unmarshal(b, &rst); err != nil {
		logger.Error("vertify_version ioutil.ReadAll error: %v", err)
		return false
	}

	if rst.Version != login.GetClientVersion() {
		logger.Error("vertify_version rst.Version:%d != login.GetClientVersion:%d", rst.Version, login.GetClientVersion())
		return false
	}
	return true
}

func (self *CNServer) Login(conn rpc.RpcConn, login rpc.Login) error {
	if !self.vertify_version(login) {
		WriteLoginResultWithErrorMsg(conn, "failed", "version_error")
		return nil
	}

	logger.Info("***********Login req openId:%s, uid:%s, bindId:%s", login.GetOpenid(), login.GetUid(), login.GetRoleId())
	uid := ""
	openId := login.GetOpenid()
	loginType := wechat_login
	if login.GetOpenid() != "" {
		binduid, err := accountclient.QueryPlayerIdByPartnerId(
			common.TB_t_account_tencentid2playerid,
			openId)
		if err != nil {
			logger.Info("connId = %d login QueryPlayerIdByPartnerId :%v", conn.GetId(), err)
			return nil
		}

		//wechat login
		if len(binduid) > 0 {
			self.AfterAccountServer(&login, binduid, conn, loginType)
			return nil
		}

		//wechat register
		uid = GenUUID(self.GetServerId())
		roleId, _, err := roleclient.ReqGenRolesId(uid)
		if err != nil {
			logger.Error("Login roleclient.ReqGenRolesId() err:%s", err)
		}
		login.SetRoleId(roleId)
		loginType = wechat_register
	} else if login.GetRoleId() == int32(0) { //guest  register
		uid = GenUUID(self.GetServerId())
		roleId, nameId, err := roleclient.ReqGenRolesId(uid)
		if err != nil {
			logger.Error("Login roleclient.ReqGenRolesId(%s), err:%s", uid, err)
			return err
		}

		login.SetRoleId(roleId)
		login.SetNickName(language.GetLanguage("TID_TOUR") + strconv.Itoa(int(nameId)))
		openId = strconv.Itoa(int(roleId))
		loginType = guest_register
	} else { //guest  login
		openId = strconv.Itoa(int(login.GetRoleId()))
		binduid, err := accountclient.QueryPlayerIdByPartnerId(
			common.TB_t_account_tencentid2playerid,
			openId)
		if err != nil {
			logger.Info("connId = %d login QueryPlayerIdByPartnerId :%v", conn.GetId(), err)
			return nil
		}

		if len(binduid) <= 0 {
			msg := &rpc.ErrorCodeNofify{}
			msg.SetCode(22)
			WriteResult(conn, msg)
			return nil
		}
		loginType = guest_login
		self.AfterAccountServer(&login, binduid, conn, loginType)
	}

	connId := conn.GetId()
	_, exist := self.getPlayerByConnId(connId)
	if exist {
		te("CNServer:Login error: has exist connId = ", connId)
		return nil
	}

	binduid, err := accountclient.QueryPlayerIdByPartnerId(
		common.TB_t_account_tencentid2playerid,
		openId)
	if err != nil {
		logger.Info("connId = %d login QueryPlayerIdByPartnerId :%v", connId, err)
		return nil
	}

	logger.Info("***********guestLogin openId:%s, roleId:%d, bindId:%s", login.GetOpenid(), login.GetRoleId(), binduid)

	if len(binduid) > 0 {
		if loginType == guest_login && login.GetUid() != binduid {
			logger.Error("loginType == guest_login && login.GetUid() != binduid")
			return nil
		}

		self.AfterAccountServer(&login, binduid, conn, loginType)
	} else {
		err = accountclient.SetPartnerIdToPlayerId(
			common.TB_t_account_tencentid2playerid,
			openId,
			10,
			uid)

		if err == nil {
			self.AfterAccountServer(&login, uid, conn, loginType)
		} else {
			logger.Error("connId = %d GuestLogin SetPartnerIdToPlayerId :%v", connId, err)
			return nil
		}
	}
	return nil
}

func (self *CNServer) AfterAccountServer(login *rpc.Login, uid string, conn rpc.RpcConn, loginType int) error {
	connId := conn.GetId()
	logger.Info("connId = %d uid := %s : AfterAccountServer begin", connId, uid)

	lid := GenLockMessage(self.GetServerId(), proto.MethodPlayerLogin, 0)

	//顶号次数，做限制，最多5次，因为其它逻辑会出现一直顶不下的情况
	iKickTimes := 0
	for {
		successed, old_value, err := lockclient.TryLock(common.LockName_Player, uid, lid, common.LockTime_Login, nil)
		if err != nil {
			logger.Error("lockclient.TryLock error", err)
			WriteLoginResult(conn, "LoginResult_SERVERERROR")
			return nil
		}

		logger.Info("connId = %d uid := %s : TryLock end", connId, uid, successed)

		if successed {
			break
		}

		_, tid, _, _, _ := ParseLockMessage(old_value)

		switch tid {
		case proto.MethodPlayerLogin:
			{
				iKickTimes++
				bFailed := true
				if iKickTimes <= 5 {
					//顶号
					logger.Info("connId = %d uid := %s : begin kick self", connId, uid)
					req := &proto.LoginKickPlayer{
						Id: uid,
					}
					rst := &proto.LoginKickPlayerResult{}
					if err := centerclient.Call("Center.KickCnsPlayer", req, rst); err == nil && rst.Success {
						logger.Info("connId = %d uid := %s : kick self success", connId, uid)
						//不处理，等下面等待1秒后重试
						bFailed = false
					} else {
						logger.Info("connId = %d uid := %s : kick self error:%v", connId, uid, err)
					}
				}

				if bFailed {
					logger.Info("connId = %d uid := %s : login kick self failed", connId, uid)
					WriteLoginResult(conn, "LoginResult_kick_player_failed")
					return nil
				}
			}
		default:
			logger.Info("connId = %d uid := %s default, tid(%d)", connId, uid, tid)
			return nil
		}

		time.Sleep(time.Second)
	}

	logger.Info("connId = %d uid := %s : begin LoadPlayer", connId, uid)

	var p *player = nil
	gl := int32(10)
	//防止下面功能出错
	defer func() {
		if r := recover(); r != nil {
			logger.Error("connId = %d uid := %s : login serious error !!", connId, uid)
			debug.PrintStack()

			if p != nil {
				self.LoginError(p)
			} else {
				self.DisConnOnLogin(conn, uid, lid, gl, uint32(0), false, true)
			}
		}
	}()

	var newPlayer bool = false
	p, newPlayer = LoadPlayer(uid, login.GetNickName(), lid, login.GetRoleId())
	if p == nil {
		self.DisConnOnLogin(conn, uid, lid, gl, uint32(0), false, true)
		return nil
	}
	if loginType == wechat_register || loginType == wechat_login {
		if !p.GetBModifyName() {
			p.SetName(login.GetNickName())
		}
		if !p.GetBModifySex() {
			p.SetSex(login.GetSex())
		}
		p.SetHeaderUrl(login.GetHeaderUrl())
		p.SetAccountType(1)
	} else if loginType == guest_register {
		p.SetAccountType(0)
	}
	// p.SetVipLeftDay(int32(7))

	//这里就要赋值，防止有用到连接的地方
	p.conn = conn
	logger.Info("connId = %d uid := %s : begin OnInit", connId, uid)

	openId := login.GetOpenid()
	if openId == "" {
		openId = strconv.Itoa(int(login.GetRoleId()))
	}

	p.mobileqqinfo = &MobileQQInfo{
		Openid: openId,
	}
	//应该放在判断非空的后面
	p.OnInit(conn)
	if newPlayer {
		p.OnNewPlayer()
		rankclient.UpdateRankingInfo(p.GetUid(), RANK_COIN, p.GetCoin()+p.GetInsurCoin())
		rankclient.UpdateRankingInfo(p.GetUid(), RANK_EXP, p.GetExp())
		rankclient.UpdateRankingInfo(p.GetUid(), RANK_PROFIT, int32(0))
	}

	p.Ip = conn.GetRemoteIp()
	login.SetUid(uid)

	//发送登录成功消息
	ret := WriteLoginResult2(conn, "ok", login)
	if !ret {
		self.LoginError(p)
		return nil
	}

	//下发玩家数据,这里不把信息下发给玩家看
	playerinfo := &rpc.PlayerInfo{Base: p.PlayerBaseInfo, Extra: p.PlayerExtraInfo}
	ret = WriteResult(conn, playerinfo)
	if !ret {
		self.LoginError(p)
		return nil
	}

	// 邮件初始化
	err, mailinfo := mailclient.GetAllMail(p.GetUid())
	if err == nil && mailinfo != nil {
		WriteResult(conn, mailinfo)
	}

	//friends
	p.queryAddDelFriendInfo(false)

	myerr := self.GetFriendsList(conn, p)
	if myerr != nil {
		logger.Error("connId = %d uid := %s : Send Friend Msg error : %v", connId, uid, myerr)
	}

	//offline msg
	offlineMsg, err := roleclient.GetOfflineChatMsg(p.GetUid())
	if err != nil {
		logger.Error("AfterAccountServer roleclient.GetOfflineChatMsg err:%s, uid:%s", err, p.GetUid())
	} else if offlineMsg != nil {
		WriteResult(conn, offlineMsg)
	}

	// p.ReconnectEnterRoom()
	logger.Info("player name :%s", p.GetName())

	//进入服务器全局表
	self.addPlayer(conn.GetId(), p)

	//after login
	p.OnlineNotice()
	p.checkBilling()
	return nil
}

func (self *CNServer) HeartBeatCall(conn rpc.RpcConn, beat rpc.HeartBeat) error {
	if conn == nil {
		return nil
	}

	msg := &rpc.HeartBeatRst{}
	msg.SetTime(int64(time.Now().UnixNano() / 1e6))
	WriteResult(conn, msg)
	return nil
}
