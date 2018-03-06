package connector

import (
	"logger"
	"matchclient"
	"rpc"
)

func (self *CNServer) MatchListREQ(conn rpc.RpcConn, msg rpc.MatchListREQ) error {
	logger.Info("client call MatchListREQ begin")

	matchclient.MatchListREQ()
	return nil
}

func (self *CNServer) EnrollREQ(conn rpc.RpcConn, msg rpc.EnrollREQ) error {
	logger.Info("client call EnrollREQ begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	matchclient.EnrollREQ(p.PlayerBaseInfo, &msg)
	return nil
}

func (self *CNServer) WithdrawREQ(conn rpc.RpcConn, msg rpc.WithdrawREQ) error {
	logger.Info("client call WithdrawREQ begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	matchclient.WithdrawREQ(p.PlayerBaseInfo.GetUid(), &msg)
	return nil
}
