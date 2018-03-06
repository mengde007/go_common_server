package pockerserver

import (
	"centerclient"
	"common"
	"logger"
	"math/rand"
	"net"
	"proto"
	"rpc"
	"rpcplus"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

type PockerServer struct {
	l          sync.RWMutex
	pCachePool *common.CachePool
}

var pServer *PockerServer

func CreateServices(cfg common.PockerConfig, listener net.Listener) *PockerServer {
	pServer = &PockerServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)

	//读配置表
	common.LoadDaerGlobalConfig()
	// common.LoadPockerConfig()
	common.LoadDaerRoomConfig()

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
func (self *PockerServer) init() {
	logger.Info("begin init begin...")
	defer logger.Info("begin init end...")
	if pockerRoomMgr == nil {
		pockerRoomMgr = &PockerRoomMgr{}
	}
	pockerRoomMgr.init()

	//自建房
	if customRoomMgr == nil {
		customRoomMgr = &CustomPockerRoomMgr{}
	}
	customRoomMgr.init()

	self.GlobalInit()
	rand.Seed(time.Now().UnixNano())
}

//全局变量初始化
func (self *PockerServer) GlobalInit() {
	cfg := common.GetDaerGlobalConfig(strconv.Itoa(573))
	if cfg == nil {
		logger.Fatal("GetDaerGlobalConfig(573) return nil")
	}
	// COUNTDOWN_MAX = cfg.IntValue
	COUNTDOWN_MAX = int32(20)

	cfg = common.GetDaerGlobalConfig(strconv.Itoa(574))
	if cfg == nil {
		logger.Fatal("GetDaerGlobalConfig(574) return nil")
	}
	ROOM_PLAYERS_LIMIT = cfg.IntValue

	cfg = common.GetDaerGlobalConfig(strconv.Itoa(575))
	if cfg == nil {
		logger.Fatal("GetDaerGlobalConfig(575) return nil")
	}
	// GAME_OVER_WAIT = cfg.IntValue
	GAME_OVER_WAIT = int32(6)
}

//玩家执行动作
func (self *PockerServer) PlayerAction(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("PlayerAction begin")
	defer logger.Info("PlayerAction end")

	msg := &rpc.C2SAction{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("PlayerAction err:%s", err)
		return err
	}

	if self.in_custom_room(msg.GetUid()) {
		customRoomMgr.PlayerAction(msg)
	} else {
		pockerRoomMgr.PlayerAction(msg)
	}
	return nil
}

//玩家执行动作
func (self *PockerServer) CreateCustomRoom(req *proto.ReqPockerCustomRoom, rst *proto.RstPockerCustomRoom) (err error) {
	logger.Info("CreateCustomRoom begin")
	defer logger.Info("CreateCustomRoom end")

	//进入房间
	msg := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("CreateCustomRoom common.DecodeMessage err:%s", err)
		return err
	}

	if ok, code := common.CheckCustomPockerCoin(req.BlindId, req.LimId, msg); !ok {
		send := &rpc.PockerRoomInfo{}
		send.SetCode(code)
		centerclient.SendCommonNotify2S([]string{msg.GetUid()}, send, "PockerRoomInfo")
		logger.Error("进入失败, code:%d", code)
		return nil
	}
	logger.Info("****************a1")
	//创建
	room := customRoomMgr.CreateRoom(req.BlindId, req.LimId)
	if room == nil {
		rst.RoomNo = int32(-1)
		return nil
	}
	// send := &rpc.CreatePockerRoomAck{}
	// send.SetRoomNo(room.roomNo)
	// centerclient.SendCommonNotify2S([]string{msg.GetUid()}, send, "CreatePockerRoomAck")
	rst.RoomNo = room.roomNo

	customRoomMgr.EnterRoom(room.roomNo, msg)
	return nil
}

//进入房间
func (self *PockerServer) PlayerEnterGame(req *proto.ReqEnterPockerRoom, rst *proto.OperRst) (err error) {
	logger.Info("PlayerEnterGame begin, gameType:%s", req.GameType)
	defer logger.Info("PlayerEnterGame end")

	msg := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("PlayerEnterGame common.DecodeMessage err:%s", err)
		return err
	}

	if req.GameType == "6" {
		customRoomMgr.EnterRoom(req.EType, msg)
	} else {
		pockerRoomMgr.EnterGame(req.EType, msg)
	}
	return nil
}

func (self *PockerServer) QuicklySeatdown(req *proto.ReqEnterPockerRoom, rst *proto.OperRst) (err error) {
	logger.Info("QuicklySeatdown begin")
	defer logger.Info("QuicklySeatdown end")

	msg := &rpc.PlayerBaseInfo{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		logger.Error("QuicklySeatdown common.DecodeMessage err:%s", err)
		return err
	}

	pockerRoomMgr.quickly_seatdown(msg)
	return nil
}

func (self *PockerServer) GetOnlineNums(req *proto.ReqOnlineNum, rst *proto.RstOnlineNum) (err error) {
	logger.Info("GetOnlineNums begin")
	defer logger.Info("GetOnlineNums end")

	msg := pockerRoomMgr.GetOnlineNum()

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	rst.RoomInfo = bufb
	return nil
}

func (self *PockerServer) PlayerIsInRoom(req *proto.ReqIsInRoom, rst *proto.OperRst) (err error) {
	logger.Info("PlayerIsInRoom begin")
	defer logger.Info("PlayerIsInRoom end")

	if self.in_custom_room(req.Uid) {
		if customRoomMgr.IsInRoom(req.Uid) {
			rst.Ok = "OK"
		}
	} else {
		if pockerRoomMgr.IsInRoom(req.Uid) {
			rst.Ok = "OK"
		}
	}

	return nil
}

func (self *PockerServer) SendDeskMsg(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
	logger.Info("SendDeskMsg begin")
	defer logger.Info("SendDeskMsg end")

	msg := &rpc.FightRoomChatNotify{}
	err = common.DecodeMessage(req.Msg, msg)
	if err != nil {
		return err
	}

	if self.in_custom_room(msg.GetPlayerID()) {
		customRoomMgr.SendDeskChatMsg(msg)
	} else {
		pockerRoomMgr.SendDeskChatMsg(msg)
	}
	return nil
}

func (self *PockerServer) in_custom_room(uid string) bool {
	if customRoomMgr.IsInRoom(uid) {
		return true
	}
	return false
}
