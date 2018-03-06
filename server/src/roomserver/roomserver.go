package roomserver

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
	"strconv"
	"sync"
	// "time"
	// conn "centerclient"
	cmn "common"
	ds "daerserver"
	"strings"
)

type RoomServer struct {
	l          sync.RWMutex
	pCachePool *common.CachePool
}

var pServer *RoomServer

func CreateServices(cfg common.RoomConfig, listener net.Listener) *RoomServer {
	pServer = &RoomServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)

	//读配置表
	common.LoadGlobalConfig()
	common.LoadDaerGlobalConfig()
	common.LoadDaerRoomConfig()
	common.LoadCustomRoomConfig()

	ds.InitGlobalConfig()

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
func (self *RoomServer) init() {
	logger.Info("begin init begin...")
	defer logger.Info("begin init end...")
	if customRoomMgr == nil {
		customRoomMgr = &CustomRoomMgr{}
	}
	customRoomMgr.init()

}

//进入房间
func (self *RoomServer) EnterRoom(req *proto.ReqDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("EnterRoom begin")
	defer logger.Info("EnterRoom end")

	base := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Base, base)
	if err != nil {
		return err
	}

	logger.Info("*************base name:%s, sex:%d coin:%d ", base.GetName(), base.GetSex(), base.GetCoin())

	msg := &rpc.EnterCustomRoomREQ{}
	err = common.DecodeMessage(req.Client, msg)
	if err != nil {
		return err
	}
	customRoomMgr.OnEnterCustomRoom(msg.GetId(), msg.GetPwd(), base)
	return nil
}

//离开房间
func (self *RoomServer) LeaveRoom(req *proto.ReqLeaveDaerRoom, rst *proto.OperRst) (err error) {
	logger.Info("LeavedRoom begin")
	defer logger.Info("LeavedRoom end")

	msg := &rpc.LeaveCustomRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}
	customRoomMgr.OnLeaveGame(msg.GetPlayerID())
	return nil
}

//请求执行动作
func (self *RoomServer) ActionREQ(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("ActionREQ begin")
	defer logger.Info("ActionREQ end")

	msg := &rpc.ActionREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}

	customRoomMgr.OnActionGame(msg)
	return nil
}

//请求贴鬼
func (self *RoomServer) TieGuiREQ(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("TieGuiREQ begin")
	defer logger.Info("TieGuiREQ end")

	msg := &rpc.MJTieGuiREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("解码消息出错：", err)
		return err
	}

	customRoomMgr.TieGuiREQ(msg)
	return nil
}

//请求解散房间
func (self *RoomServer) JieSanRoomREQ(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
	logger.Info("JieSanRoomREQ begin")
	defer logger.Info("JieSanRoomREQ end")

	msg := &rpc.JieSanRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}

	customRoomMgr.OnJieSanRoom(req.Uid, msg)
	return nil
}

//创建房间
func (self *RoomServer) CreateRoomREQ(req *proto.ReqDaerRoomWithItem, rst *proto.OperRst) (err error) {
	logger.Info("CreateRoomREQ begin")
	defer logger.Info("CreateRoomREQ end")

	base := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Base, base)
	if err != nil {
		return err
	}

	msg := &rpc.CreateRoomREQ{}
	err = common.DecodeMessage(req.Client, msg)
	if err != nil {
		return err
	}

	customRoomMgr.OnCreateRoom(base, msg, req.Amount)
	return nil
}

//请求房间列表
func (self *RoomServer) RoomListREQ(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
	logger.Info("RoomListREQ begin")
	defer logger.Info("RoomListREQ end")

	// msg := &rpc.RoomListREQ{}
	// err = common.DecodeMessage(req.Msg, msg)
	// if err != nil {
	// 	return err
	// }

	customRoomMgr.OnObtainRoomList(req.Uid)
	return nil
}

//查找房间
func (self *RoomServer) FindRoomREQ(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
	logger.Info("FindRoomREQ begin")
	defer logger.Info("FindRoomREQ end")

	msg := &rpc.FindRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}

	customRoomMgr.OnFindRoom(req.Uid, msg)
	return nil
}

//踢出房间
func (self *RoomServer) ForceLeaveRoom(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
	logger.Info("ForceLeaveRoom begin")
	defer logger.Info("ForceLeaveRoom end")

	msg := &rpc.ForceLeaveRoomREQ{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}

	customRoomMgr.OnForceLeaveRoom(req.Uid, msg)
	return nil
}

//是否在房间
func (self *RoomServer) PlayerIsInRoom(req *proto.ReqIsInRoom, rst *proto.OperRst) (err error) {
	logger.Info("PlayerIsInRoom begin")
	defer logger.Info("PlayerIsInRoom end")

	if inRoom, _ := customRoomMgr.IsInRoom(req.Uid); inRoom {
		rst.Ok = "OK"
	}

	return nil
}

func (self *RoomServer) SendDeskMsg(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("SendDeskMsg begin")
	defer logger.Info("SendDeskMsg end")

	msg := &rpc.FightRoomChatNotify{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}
	customRoomMgr.SendDeskChatMsg(msg)
	return nil
}

func GetCostRoomCardCount(gameType, times int32) int32 {
	cfg := cmn.GetCustomRoomConfig(strconv.Itoa(int(gameType)))
	if cfg == nil {
		logger.Error("Don't found any config")
		return -1
	}

	costs := strings.Split(cfg.CreateCreditsRoomCardCost, "#")

	for _, cost := range costs {

		costInfo := strings.Split(cost, "_")
		if len(costInfo) != 2 {
			logger.Error("配置的房卡消耗有问题，在自建房配置表中。")
			continue
		}

		var costTimes int
		var err error
		var costAmount int

		if costTimes, err = strconv.Atoi(costInfo[0]); err != nil {
			logger.Error("", err)
			continue
		}

		if costAmount, err = strconv.Atoi(costInfo[1]); err != nil {
			logger.Error("", err)
			continue
		}

		if times <= int32(costTimes) {
			return int32(costAmount)
		}
	}

	return 0
}
