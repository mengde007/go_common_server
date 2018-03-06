package majiangserver

import (
	// gp "code.google.com/p/goprotobuf/proto"
	// "code.google.com/p/snappy-go/snappy"
	// "github.com/garyburd/redigo/redis"
	"logger"
	//"math/rand"
	"net"
	"proto"
	"rpc"
	"rpcplus"
	"runtime/debug"
	// "strconv"
	//"sync"
	// "time"
	//conn "centerclient"
	"common"
)

type MaJiangServer struct {
	//l          sync.RWMutex
	pCachePool *common.CachePool
}

var pServer *MaJiangServer

func CreateServices(cfg common.MaJiangConfig, listener net.Listener) *MaJiangServer {
	pServer = &MaJiangServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)

	//读配置表
	common.LoadGlobalConfig()
	common.LoadDaerGlobalConfig()
	common.LoadDaerRoomConfig()

	InitGlobalConfig()

	//启动服务器监听
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
func (self *MaJiangServer) init() {
	logger.Info("begin init begin...")
	defer logger.Info("begin init end...")
	if maJiangRoomMgr == nil {
		maJiangRoomMgr = &MaJiangRoomMgr{}
	}
	maJiangRoomMgr.init()
}

//快速进入房间
func (self *MaJiangServer) QuickEnterRoom(req *proto.ReqDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("QuickEnterRoom begin --麻将")
	defer logger.Info("QuickEnterRoom end -- 麻将")

	base := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Base, base)
	if err != nil {
		return err
	}

	logger.Info("*************base name:%s, sex:%d coin:%d ", base.GetName(), base.GetSex(), base.GetCoin())

	msg := &rpc.MJQuickEnterRoomREQ{}
	err = common.DecodeMessage(req.Client, msg)
	if err != nil {
		return err
	}

	roomType := common.GetBestRoomType(common.MaJiang, base)
	if roomType <= 0 {
		SendEnterRoomErrorACK(base.GetUid(), roomType, common.ERLessCoin, true)
		return
	}

	maJiangRoomMgr.EnterGame(roomType, base, false)
	return nil
}

//进入房间
func (self *MaJiangServer) EnterRoom(req *proto.ReqDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("EnterRoom begin")
	defer logger.Info("EnterRoom end")

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
	maJiangRoomMgr.EnterGame(int32(msg.GetRoomType()), base, false)
	return nil
}

//离开房间
func (self *MaJiangServer) LeaveRoom(req *proto.ReqLeaveDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("LeaveRoom begin")
	defer logger.Info("LeaveRoom end")

	msg := &rpc.MJLeaveRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}
	maJiangRoomMgr.LeaveGame(msg.GetPlayerID(), msg.GetIsChangeDesk())
	return nil
}

// //踢人
func (self *MaJiangServer) ForceLeaveRoom(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
	logger.Info("ForceLeaveRoom begin")
	defer logger.Info("ForceLeaveRoom end")

	msg := &rpc.ForceLeaveRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}

	maJiangRoomMgr.ForceLeaveRoom(req.Uid, msg.GetId())
	return nil
}

//
func (self *MaJiangServer) ActionsREQ(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("ActionsREQ begin")
	defer logger.Info("ActionsREQ end")

	msg := &rpc.ActionREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("解码消息出错：", err)
		return err
	}

	maJiangRoomMgr.ActionGame(msg)
	return nil
}

func (self *MaJiangServer) TieGuiREQ(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("TieGuiREQ begin")
	defer logger.Info("TieGuiREQ end")

	msg := &rpc.MJTieGuiREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("解码消息出错：", err)
		return err
	}

	maJiangRoomMgr.TieGuiREQ(msg)
	return nil
}

func (self *MaJiangServer) GetOnlineNums(req *proto.ReqOnlineNum, rst *proto.RstOnlineNum) (err error) {
	logger.Info("GetOnlineNums begin")
	defer logger.Info("GetOnlineNums end")

	msg := maJiangRoomMgr.GetOnlineNum()

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	rst.RoomInfo = bufb
	return nil
}

func (self *MaJiangServer) PlayerIsInRoom(req *proto.ReqIsInRoom, rst *proto.OperRst) (err error) {
	logger.Info("PlayerIsInRoom begin")
	defer logger.Info("PlayerIsInRoom end")

	if maJiangRoomMgr.IsInRoom(req.Uid) {
		rst.Ok = "OK"
	}

	return nil
}

func (self *MaJiangServer) SendDeskMsg(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("SendDeskMsg begin")
	defer logger.Info("SendDeskMsg end")

	msg := &rpc.FightRoomChatNotify{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}
	maJiangRoomMgr.SendDeskChatMsg(msg)
	return nil
}
