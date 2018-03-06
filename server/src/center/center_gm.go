package center

import (
	"logger"
	"net"
	"proto"
	"rpcplus"
)

type CenterGmServices struct {
}

var pCenterGmServices *CenterGmServices

func CreateCenterServiceForGM(listener net.Listener) {
	pCenterGmServices = &CenterGmServices{}

	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pCenterGmServices)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("centerservergm StartServices %s", err.Error())
			break
		}
		go func() {
			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}
}

//取在线玩家数量，现在未区分地区
func (self *CenterGmServices) GmGetOnlineNum(req *proto.GmGetOnlineNum, rst *proto.GmGetOnlineNumResult) error {
	//rst.Value = uint32(len(centerServer.POnline))
	rst.Value = 0
	return nil
}

//取得排行数据
// func (self *CenterGmServices) GetRankMyself(req *proto.GetMyself, rst *proto.GetMyselfResult) error {
// 	return centerServer.GetRankMyselfGlobal(req, rst)
// }
