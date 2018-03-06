package matchdaerserver

import (
	// gp "code.google.com/p/goprotobuf/proto"
	// "code.google.com/p/snappy-go/snappy"
	"common"
	// "github.com/garyburd/redigo/redis"
	"logger"
	//"math/rand"
	"net"
	//"proto"
	//"rpc"
	"rpcplus"
	"runtime/debug"
	// "strconv"
	"sync"
	// "time"
	// conn "centerclient"
	// cmn "common"
)

type MatchDaerServer struct {
	l          sync.RWMutex
	pCachePool *common.CachePool
}

var pServer *MatchDaerServer

func CreateServices(cfg common.MatchDaerConfig, listener net.Listener) *MatchDaerServer {
	pServer = &MatchDaerServer{
		pCachePool: common.NewCachePool(cfg.Maincache),
	}
	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pServer)

	//读配置表
	common.LoadGlobalConfig()
	//common.LoadDaerGlobalConfig()
	//common.LoadDaerRoomConfig()
	//common.LoadCustomRoomConfig()

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
func (self *MatchDaerServer) init() {
	logger.Info("begin init begin...")
	defer logger.Info("begin init end...")
	// if customRoomMgr == nil {
	// 	customRoomMgr = &CustomRoomMgr{}
	// }
	// customRoomMgr.init()

}

// //进入房间
// func (self *RoomServer) EnterRoom(req *proto.ReqDaerRoom, rst *proto.OperRst) (err error) {
// 	logger.Info("EnterRoom begin")
// 	defer logger.Info("EnterRoom end")

// 	base := &rpc.PlayerBaseInfo{}
// 	err = common.DecodeMessage(req.Base, base)
// 	if err != nil {
// 		return err
// 	}

// 	logger.Info("*************base name:%s, sex:%d coin:%d ", base.GetName(), base.GetSex(), base.GetCoin())

// 	msg := &rpc.EnterCustomRoomREQ{}
// 	err = common.DecodeMessage(req.Client, msg)
// 	if err != nil {
// 		return err
// 	}
// 	customRoomMgr.OnEnterCustomRoom(msg.GetId(), msg.GetPwd(), base)
// 	return nil
// }

// //离开房间
// func (self *RoomServer) LeaveRoom(req *proto.ReqLeaveDaerRoom, rst *proto.OperRst) (err error) {
// 	logger.Info("LeavedRoom begin")
// 	defer logger.Info("LeavedRoom end")

// 	msg := &rpc.LeaveCustomRoomREQ{}
// 	err = common.DecodeMessage(req.Msg, msg)
// 	if err != nil {
// 		return err
// 	}
// 	customRoomMgr.OnLeaveGame(msg.GetPlayerID())
// 	return nil
// }

// //请求执行动作
// func (self *RoomServer) ActionREQ(req *proto.ReqActionREQ, rst *proto.OperRst) (err error) {
// 	logger.Info("ActionREQ begin")
// 	defer logger.Info("ActionREQ end")

// 	msg := &rpc.ActionREQ{}
// 	err = common.DecodeMessage(req.Msg, msg)
// 	if err != nil {
// 		return err
// 	}

// 	customRoomMgr.OnActionGame(msg)
// 	return nil
// }

// //创建房间
// func (self *RoomServer) CreateRoomREQ(req *proto.ReqDaerRoom, rst *proto.OperRst) (err error) {
// 	logger.Info("CreateRoomREQ begin")
// 	defer logger.Info("CreateRoomREQ end")

// 	base := &rpc.PlayerBaseInfo{}
// 	err = common.DecodeMessage(req.Base, base)
// 	if err != nil {
// 		return err
// 	}

// 	msg := &rpc.CreateRoomREQ{}
// 	err = common.DecodeMessage(req.Client, msg)
// 	if err != nil {
// 		return err
// 	}

// 	customRoomMgr.OnCreateRoom(base, msg)
// 	return nil
// }

// //请求房间列表
// func (self *RoomServer) RoomListREQ(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
// 	logger.Info("RoomListREQ begin")
// 	defer logger.Info("RoomListREQ end")

// 	// msg := &rpc.RoomListREQ{}
// 	// err = common.DecodeMessage(req.Msg, msg)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	customRoomMgr.OnObtainRoomList(req.Uid)
// 	return nil
// }

// //查找房间
// func (self *RoomServer) FindRoomREQ(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
// 	logger.Info("FindRoomREQ begin")
// 	defer logger.Info("FindRoomREQ end")

// 	msg := &rpc.FindRoomREQ{}
// 	err = common.DecodeMessage(req.Msg, msg)
// 	if err != nil {
// 		return err
// 	}

// 	customRoomMgr.OnFindRoom(req.Uid, msg)
// 	return nil
// }

// //踢出房间
// func (self *RoomServer) ForceLeaveRoom(req *proto.ReqCreateCustomRoom, rst *proto.OperRst) (err error) {
// 	logger.Info("ForceLeaveRoom begin")
// 	defer logger.Info("ForceLeaveRoom end")

// 	msg := &rpc.ForceLeaveRoomREQ{}
// 	err = common.DecodeMessage(req.Msg, msg)
// 	if err != nil {
// 		return err
// 	}

// 	customRoomMgr.OnForceLeaveRoom(req.Uid, msg)
// 	return nil
// }

// //是否在房间
// func (self *RoomServer) PlayerIsInRoom(req *proto.ReqIsInRoom, rst *proto.OperRst) (err error) {
// 	logger.Info("PlayerIsInRoom begin")
// 	defer logger.Info("PlayerIsInRoom end")

// 	if inRoom, _ := customRoomMgr.IsInRoom(req.Uid); inRoom {
// 		rst.Ok = "OK"
// 	}

// 	return nil
// }
