package daerserver

import (
	// gp "code.google.com/p/goprotobuf/proto"
	// "code.google.com/p/snappy-go/snappy"
	"common"
	// "github.com/garyburd/redigo/redis"
	"logger"
	//"math/rand"
	"net"
	"proto"
	"rpc"
	"rpcplus"
	"runtime/debug"
	// "strconv"
	"sync"
	// "time"
	conn "centerclient"
	cmn "common"
)

type DaerServer struct {
	l          sync.RWMutex
	pCachePool *common.CachePool
}

var pServer *DaerServer

func CreateServices(cfg common.DaerConfig, listener net.Listener) *DaerServer {
	pServer = &DaerServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)

	//读配置表
	common.LoadGlobalConfig()
	common.LoadDaerGlobalConfig()
	common.LoadDaerRoomConfig()

	InitGlobalConfig()

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
func (self *DaerServer) init() {
	logger.Info("begin init begin...")
	defer logger.Info("begin init end...")
	if daerRoomMgr == nil {
		daerRoomMgr = &DaerRoomMgr{}
	}
	daerRoomMgr.init()
	//self.GlobalTst()
}

//全局配置表测试
func (self *DaerServer) GlobalTst() {
	cfg := common.GetDaerGlobalConfig("21")
	if cfg == nil {
		logger.Error("GetDaerGlobalConfig return nil")
	} else {
		// logger.Error("********cfg.Rst:%s", cfg.Q_int_value)
		// logger.Error("********cfg.Rst:%s", cfg.Q_string_value)
	}

}

//快速进入房间
func (self *DaerServer) QuickEnteredRoom(req *proto.ReqDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("QuickEnteredDaerRoom begin")
	defer logger.Info("QuickEnteredDaerRoom end")

	base := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Base, base)
	if err != nil {
		return err
	}

	logger.Info("*************base name:%s, sex:%d coin:%d ", base.GetName(), base.GetSex(), base.GetCoin())

	msg := &rpc.QuickEnterRoomREQ{}
	err = common.DecodeMessage(req.Client, msg)
	if err != nil {
		return err
	}

	roomType := cmn.GetBestRoomType(cmn.DaerGame, base)
	if roomType <= 0 {
		msg := &rpc.EnterRoomACK{}
		msg.SetRoomId(roomType)
		msg.SetCode(cmn.ERLessCoin)
		msg.SetIsNormalReqEnterRoom(true)
		if err := conn.SendEnterRoom(base.GetUid(), msg); err != nil {
			logger.Error("发送结算消息出错：", err, msg)
		}
		return
	}

	daerRoomMgr.EnterGame(roomType, base, false)
	return nil
}

//进入房间
func (self *DaerServer) EnteredDaerRoom(req *proto.ReqDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("EnteredDaerRoom begin")
	defer logger.Info("EnteredDaerRoom end")

	base := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Base, base)
	if err != nil {
		return err
	}

	logger.Info("*************base name:%s, sex:%d coin:%d ", base.GetName(), base.GetSex(), base.GetCoin())

	msg := &rpc.EnterRoomREQ{}
	err = common.DecodeMessage(req.Client, msg)
	if err != nil {
		return err
	}
	daerRoomMgr.EnterGame(int32(msg.GetRoomType()), base, false)
	return nil
}

//离开房间
func (self *DaerServer) LeavedDaerRoom(req *proto.ReqLeaveDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("LeavedDaerRoom begin")
	defer logger.Info("LeavedDaerRoom end")

	msg := &rpc.LeaveRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}
	daerRoomMgr.LeaveGame(msg.GetPlayerID(), msg.GetIsChangeDesk())
	return nil
}

//踢人
func (self *DaerServer) ForceLeaveRoom(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
	logger.Info("ForceLeaveRoom begin")
	defer logger.Info("ForceLeaveRoom end")

	msg := &rpc.ForceLeaveRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}

	daerRoomMgr.ForceLeaveRoom(req.Uid, msg.GetId())
	return nil
}

//
func (self *DaerServer) DaerActionsREQ(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("DaerActionsREQ begin")
	defer logger.Info("DaerActionsREQ end")

	msg := &rpc.ActionREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("解码消息出错：", err)
		return err
	}

	daerRoomMgr.ActionGame(msg)
	return nil
}

func (self *DaerServer) GetOnlineNums(req *proto.ReqOnlineNum, rst *proto.RstOnlineNum) (err error) {
	logger.Info("GetOnlineNums begin")
	defer logger.Info("GetOnlineNums end")

	msg := daerRoomMgr.GetOnlineNum()

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	rst.RoomInfo = bufb
	return nil
}

func (self *DaerServer) PlayerIsInRoom(req *proto.ReqIsInRoom, rst *proto.OperRst) (err error) {
	logger.Info("PlayerIsInRoom begin")
	defer logger.Info("PlayerIsInRoom end")

	if daerRoomMgr.IsInRoom(req.Uid) {
		rst.Ok = "OK"
	}

	return nil
}

func (self *DaerServer) SendDeskMsg(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("SendDeskMsg begin")
	defer logger.Info("SendDeskMsg end")

	msg := &rpc.FightRoomChatNotify{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}
	daerRoomMgr.SendDeskChatMsg(msg)
	return nil
}
