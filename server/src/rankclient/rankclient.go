package rankclient

import (
	"common"
	"logger"
	"proto"
	// "rpc"
	"rpcplusclientpool"
)

var pPoll *rpcplusclientpool.ClientPool

func init() {
	aServerHost := common.ReadServerClientConfig("rankserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no rankserver server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "rankserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}
	return
}

func UpdateRankingInfo(uid string, eType int, value int32) (error, bool) {
	logger.Info("UpdateRankingInfo player")
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err, false
	}
	req := &proto.SetRankInfo{
		Uid:   uid,
		EType: eType,
		Value: value,
	}
	rst := &proto.SetRankInfoRst{}
	if err := conn.Call("GeneralRankServer.UpdateRankingInfo", req, rst); err != nil {
		return err, false
	}
	return nil, true
}

func GetRankingInfo(rankMax int) (*proto.GetRankInfoRst, error) {
	logger.Info("GetRankingInfo   rankMax:", rankMax)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return nil, err
	}
	req := &proto.GetRankInfo{
		Number: rankMax,
	}
	rst := &proto.GetRankInfoRst{}
	if err := conn.Call("GeneralRankServer.GetRankingInfo", req, rst); err != nil {
		return nil, err
	}
	return rst, nil
}

func GetMyRankingInfo(eType int, uid string) (int32, error) {
	logger.Info("GetMyRankingInfo eType:%d, uid:%s", eType, uid)
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return int32(0), err
	}
	req := &proto.GetMyRankInfo{
		Uid:   uid,
		EType: eType,
	}
	rst := &proto.GetMyRankInfoRst{}
	if err := conn.Call("GeneralRankServer.GetMyRankingInfo", req, rst); err != nil {
		return int32(0), err
	}
	return rst.Ranking, nil

}
