package roleserver

import (
	// gp "code.google.com/p/goprotobuf/proto"
	// "code.google.com/p/snappy-go/snappy"
	"common"
	// "github.com/garyburd/redigo/redis"
	"logger"
	//"math/rand"
	"net"
	"proto"
	// "rpc"
	"centerclient"
	"rpcplus"
	"runtime/debug"
	"strconv"
	"sync"
	// "time"
)

const (
	ROLE_MAIN_TALBE        = "role_main_talbe"
	ROLE_ID                = "role_id"
	GUEST_ID               = "guest_id"
	ROLE_INFO              = "role_info"
	ROLE_INFO_KEY          = "role_info_key"
	OFFLINE_CHAT_MSG_TABLE = "offline_chat_msg_table"
)

type RoleServer struct {
	l          sync.Mutex
	r          sync.RWMutex
	idInc      int32
	guInc      int32
	pCachePool *common.CachePool
	adl        *common.SimpleLockService
	cl         *common.SimpleLockService
	roleInfo   map[string]string
}

var pServer *RoleServer

func CreateServices(cfg common.RoleConfig, listener net.Listener) *RoleServer {
	pServer = &RoleServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
		adl:        common.CreateSimpleLock(),
		roleInfo:   make(map[string]string, 0),
		cl:         common.CreateSimpleLock(),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)

	//读配置表
	common.LoadGlobalConfig()

	//加载文明信息
	pServer.init()
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
					logger.Info("CreateServices Rpc Runtime Error: %s", r)
					debug.PrintStack()
				}
			}()
			rpcServer.ServeConn(conn)
		}()
	}
	return pServer
}

//初始化
func (self *RoleServer) init() {
	logger.Info("begin init begin...")
	defer logger.Info("begin init end...")
	self.l.Lock()
	rst, err := common.Redis_getInt(self.pCachePool, ROLE_MAIN_TALBE, ROLE_ID)
	if err != nil {
		logger.Error("common.Redis_getInt, mainTable:%s, Key:%s error:%s", ROLE_MAIN_TALBE, ROLE_ID, err)
		return
	}
	if rst == 0 {
		rst = 10000000
		common.Redis_setInt(self.pCachePool, ROLE_MAIN_TALBE, ROLE_ID, rst)
	}
	self.idInc = int32(rst)

	rst, err = common.Redis_getInt(self.pCachePool, ROLE_MAIN_TALBE, GUEST_ID)
	if err != nil {
		logger.Error("common.Redis_getInt, mainTable:%s, Key:%s error:%s", ROLE_MAIN_TALBE, GUEST_ID, err)
		return
	}
	if rst == 0 {
		rst = 10000
		common.Redis_setInt(self.pCachePool, ROLE_MAIN_TALBE, GUEST_ID, rst)
	}
	self.guInc = int32(rst)
	self.l.Unlock()

	//role info
	self.r.Lock()
	self.load()
	self.r.Unlock()
}

func (self *RoleServer) save() {
	saveBuf, err := common.GobEncode(self.roleInfo)
	if err != nil {
		logger.Error("save GobEncode err", err)
		return
	}

	if err := common.Resis_setbuf(self.pCachePool, ROLE_INFO, ROLE_INFO_KEY, saveBuf); err != nil {
		logger.Error("save setbuf error", err)
		return
	}
}

func (self *RoleServer) load() {
	buf, err := common.Resis_getbuf(self.pCachePool, ROLE_INFO, ROLE_INFO_KEY)
	if err != nil {
		logger.Error("load getbuf error,", err)
		return
	}

	if buf != nil {
		err = common.GobDecode(buf, &self.roleInfo)
		if err != nil {
			logger.Error("load GobDecode error", err)
			return
		}
	}
	logger.Info("load role form redis ok")
}

func (self *RoleServer) GetUidByRoleId(req *proto.ReqSearch, rst *proto.SearchRst) (err error) {
	logger.Info("GetUidByRoleId begin")
	defer logger.Info("GetUidByRoleId end roleId:%", self.idInc)

	strRoleId := strconv.Itoa(int(req.RoleId))

	self.r.RLock()
	defer self.r.RUnlock()
	uid, exist := self.roleInfo[strRoleId]
	if !exist {
		logger.Error("GetUidByRoleId can't find, roleId:%d", req.RoleId)
		return
	}
	rst.Uid = uid
	return nil
}

//生成角色Id
func (self *RoleServer) GenRoleId(req *proto.ReqGenRoleId, rst *proto.RstGenRoleId) (err error) {
	logger.Info("GenRoleId begin")
	defer logger.Info("GenRoleId end roleId:%", self.idInc)

	//gen roleId, nameId
	self.l.Lock()
	self.idInc += 1
	common.Redis_setInt(self.pCachePool, ROLE_MAIN_TALBE, ROLE_ID, int(self.idInc))

	self.guInc += 1
	common.Redis_setInt(self.pCachePool, ROLE_MAIN_TALBE, GUEST_ID, int(self.guInc))

	rst.RoleId = self.idInc
	rst.GuestId = self.guInc
	self.l.Unlock()

	//register roleId
	self.r.Lock()
	strRoleId := strconv.Itoa(int(rst.RoleId))
	self.roleInfo[strRoleId] = req.Uid
	self.save()
	self.r.Unlock()
	rst.Ok = true
	return nil
}

// 好友操作 添加删除
func (self *RoleServer) RequestAddFriend(req *proto.AddFriendRequest, rst *proto.AddFriendRequestRst) error {
	rst.Success = false

	self.adl.WaitLock(req.OtherUid)
	defer self.adl.WaitUnLock(req.OtherUid)

	buf, err := common.Resis_getbuf(self.pCachePool, common.PlayerFriendOperate, req.OtherUid)
	if err != nil {
		logger.Error("RequestAddFriend getbuf error,", req, err)
		return err
	}
	info := &proto.OperateList{}
	if buf != nil {
		err = common.GobDecode(buf, info)
		if err != nil {
			logger.Error("RequestAddFriend GobDecode error", req, err)
			return err
		}
	}
	newFriend := &proto.OperateFriendInfo{
		Uid:       req.MyUid,
		BeAdd:     req.BeAdd,
		BeConfirm: req.BeConfirm,
	}
	info.AddList = append(info.AddList, newFriend)
	saveBuf, err := common.GobEncode(info)
	if err != nil {
		logger.Error("RequestAddFriend GobEncode err", req, err)
		return err
	}

	if err := common.Resis_setbuf(self.pCachePool, common.PlayerFriendOperate, req.OtherUid, saveBuf); err != nil {
		logger.Error("RequestAddFriend setbuf error", req, err)
		return err
	}
	rst.Success = true

	//通知玩家去取此信息
	centerReq := &proto.FriendNoticeUpdate{
		Uid: req.OtherUid,
	}
	centerRst := &proto.FriendNoticeUpdateRst{}
	centerclient.Go("Center.NotifyAddDelFriend", centerReq, centerRst)

	return nil
}

// 查询 同时删除确认添加和删除的信息 这些信息查询回去之后就会被处理
func (self *RoleServer) QueryAddDelFriendInfo(req *proto.FriendQueryPFBI, rst *proto.FriendQueryPFBIRst) error {
	rst.Value = nil

	self.adl.WaitLock(req.Uid)
	defer self.adl.WaitUnLock(req.Uid)

	buf, err := common.Resis_getbuf(self.pCachePool, common.PlayerFriendOperate, req.Uid)
	if err != nil {
		logger.Error("QueryAddDelFriendInfo getbuf error,", req, err)
		return err
	}

	info := &proto.OperateList{}
	newInfo := &proto.OperateList{}
	if buf != nil {
		err = common.GobDecode(buf, info)
		if err != nil {
			logger.Error("QueryAddDelFriendInfo GobDecode error", req, err)
			return err
		}
	}
	for _, v := range info.AddList {
		if v.BeAdd && !v.BeConfirm {
			// 只留下请求的信息
			newInfo.AddList = append(newInfo.AddList, v)
		}
	}
	saveBuf, err := common.GobEncode(newInfo)
	if err != nil {
		logger.Error("QueryAddDelFriendInfo GobEncode err", req, err)
		return err
	}

	if err := common.Resis_setbuf(self.pCachePool, common.PlayerFriendOperate, req.Uid, saveBuf); err != nil {
		logger.Error("RequestAddDelFriendInfo setbuf error", req, err)
		return err
	}

	rst.Value = buf
	return nil
}

func (self *RoleServer) ResponseAddFriend(req *proto.ResponseAddFriend, rst *proto.ResponseAddFriendRst) error {
	if req.MyUid == "" {
		return nil
	}
	if !req.BeAll && req.OtherUid == "" {
		// 没有回应Uid
		return nil
	}

	self.adl.WaitLock(req.MyUid)

	buf, err := common.Resis_getbuf(self.pCachePool, common.PlayerFriendOperate, req.MyUid)
	if err != nil {
		logger.Error("ResponseAddDelFriendInfo getbuf error,", req, err)
		self.adl.WaitUnLock(req.MyUid)
		return err
	}
	if buf == nil {
		logger.Error("buf is nil")
		self.adl.WaitUnLock(req.MyUid)
		return nil
	}
	notifyUids := make(map[string]int)

	info := &proto.OperateList{}
	newInfo := &proto.OperateList{}
	err = common.GobDecode(buf, info)
	if err != nil {
		logger.Error("ResponseAddDelFriendInfo GobDecode error", req, err)
		self.adl.WaitUnLock(req.MyUid)
		return err
	}
	for _, v := range info.AddList {
		if v.BeAdd && !v.BeConfirm { // 是请求添加信息
			if req.OtherUid != v.Uid { // 留下其他请求信息
				newInfo.AddList = append(newInfo.AddList, v)
			} else { // 保存已经确认的id，用于通知对方
				notifyUids[req.OtherUid] = 1
			}
		} else {
			// 留下有非好友添加信息
			newInfo.AddList = append(newInfo.AddList, v)
		}
	}
	saveBuf, err := common.GobEncode(newInfo)
	if err != nil {
		logger.Error("ResponseAddDelFriendInfo GobEncode err", req, err)
		self.adl.WaitUnLock(req.MyUid)
		return err
	}

	if err := common.Resis_setbuf(self.pCachePool, common.PlayerFriendOperate, req.MyUid, saveBuf); err != nil {
		logger.Error("ResponseAddDelFriendInfo setbuf error", req, err)
		self.adl.WaitUnLock(req.MyUid)
		return err
	}

	self.adl.WaitUnLock(req.MyUid)

	if !req.BeAccept {
		// 如果是拒绝添加 直接返回不做其他处理
		return nil
	}

	// 确认添加 则给所有确认的玩家发信息
	unLockUid := ""
	for kuid, _ := range notifyUids {
		if unLockUid != "" {
			self.adl.WaitUnLock(unLockUid)
		}
		unLockUid = kuid
		self.adl.WaitLock(kuid)

		buf, err := common.Resis_getbuf(self.pCachePool, common.PlayerFriendOperate, kuid)
		if err != nil {
			logger.Error("ResponseAddFriend getbuf error,", kuid, err)
			continue
		}
		info := &proto.OperateList{}
		if buf != nil {
			err = common.GobDecode(buf, info)
			if err != nil {
				logger.Error("ResponseAddFriend GobDecode error", kuid, err)
				continue
			}
		}
		newFriend := &proto.OperateFriendInfo{
			Uid:       req.MyUid,
			BeAdd:     true,
			BeConfirm: true,
		}
		info.AddList = append(info.AddList, newFriend)
		saveBuf, err := common.GobEncode(info)
		if err != nil {
			logger.Error("ResponseAddFriend GobEncode err", kuid, err)
			continue
		}

		if err := common.Resis_setbuf(self.pCachePool, common.PlayerFriendOperate, kuid, saveBuf); err != nil {
			logger.Error("ResponseAddFriend setbuf error", kuid, err)
			continue
		}

		//通知玩家去取此信息
		centerReq := &proto.FriendNoticeUpdate{
			Uid: kuid,
		}
		centerRst := &proto.FriendNoticeUpdateRst{}
		centerclient.Go("Center.NotifyAddDelFriend", centerReq, centerRst)
	}
	// 最后一次循环的Lock
	if unLockUid != "" {
		self.adl.WaitUnLock(unLockUid)
	}
	return nil
}

func (self *RoleServer) SaveOfflineChatMsg(req *proto.OfflineChatMsg, rst *proto.CommonRst) error {
	logger.Info("SaveOfflineChatMsg has called")

	self.cl.WaitLock(req.Uid)
	defer self.cl.WaitUnLock(req.Uid)

	buf, err := common.Resis_getbuf(self.pCachePool, OFFLINE_CHAT_MSG_TABLE, req.Uid)
	if err != nil {
		logger.Error("QueryAddDelFriendInfo getbuf error,", req, err)
		return err
	}

	info := &proto.OfflineMsgList{}
	if buf != nil {
		err = common.GobDecode(buf, info)
		if err != nil {
			logger.Error("SaveOfflineChatMsg GobDecode error", req, err)
			return err
		}
	}

	info.MsgLst = append(info.MsgLst, req)
	saveBuf, err := common.GobEncode(info)
	if err != nil {
		logger.Error("QueryAddDelFriendInfo GobEncode err", req, err)
		return err
	}

	if err := common.Resis_setbuf(self.pCachePool, OFFLINE_CHAT_MSG_TABLE, req.Uid, saveBuf); err != nil {
		logger.Error("QueryAddDelFriendInfo setbuf error", req, err)
		return err
	}
	return nil
}

func (self *RoleServer) GetOfflineChatMsg(req *proto.ReqOfflineMsg, rst *proto.OfflineMsgList) error {
	logger.Info("GetOfflineChatMsg has called")

	self.cl.WaitLock(req.Uid)
	defer self.cl.WaitUnLock(req.Uid)

	buf, err := common.Resis_getbuf(self.pCachePool, OFFLINE_CHAT_MSG_TABLE, req.Uid)
	if err != nil {
		logger.Error("GetOfflineChatMsg getbuf error,", req, err)
		return err
	}

	if buf == nil {
		return nil
	}

	err = common.GobDecode(buf, rst)
	if err != nil {
		logger.Error("GetOfflineChatMsg GobDecode error", req, err)
		return err
	}

	if err := common.Redis_del(self.pCachePool, OFFLINE_CHAT_MSG_TABLE, req.Uid); err != nil {
		logger.Error("GetOfflineChatMsg common.Redis_del error", req, err)
		return err
	}

	return nil
}
