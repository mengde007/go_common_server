package payclient

import (
	"common"
	"logger"
	"proto"
	"rpc"
	"rpcplusclientpool"
	"strconv"
)

var pPoll *rpcplusclientpool.ClientPool

func init() {
	aServerHost := common.ReadServerClientConfig("payserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "payserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	return
}

func QueryPayInfos(openId string) (string, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return "", err
	}

	req := &proto.QueryPayInfo{
		OpenId: openId,
	}
	rst := &proto.QueryPayInfoRst{}

	if err = conn.Call("PayService.QueryPayInfo", req, rst); err != nil {
		return "", err
	}
	return rst.ItemId, nil
}

func DeletePayInfo(openId string) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	req := &proto.QueryPayInfo{
		OpenId: openId,
	}
	rst := &proto.CommonRst{}
	if err = conn.Call("PayService.DeletePayInfo", req, rst); err != nil {
		return err
	}
	return nil
}

func CreatePayOrder(uid, openId, itemId, Ip string, money uint32) (*rpc.OrderInfoNofity, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return nil, err
	}

	req := &proto.CreateOrder{
		Uid:    uid,
		OpenId: openId,
		ItemId: itemId,
		Ip:     Ip,
		Money:  money,
	}
	rst := &proto.CreateOrderRst{}
	if err = conn.Call("PayService.CreateOrder", req, rst); err != nil {
		return nil, err
	}

	timestamp, _ := strconv.Atoi(rst.Timestamp)
	msg := &rpc.OrderInfoNofity{}
	msg.SetPartnerId(rst.Partnerid)
	msg.SetPrepayId(rst.Prepayid)
	msg.SetPackage(rst.Package)
	msg.SetNonceStr(rst.Noncestr)
	msg.SetTimeStamp(int32(timestamp))
	msg.SetSign(rst.Sign)
	msg.SetAppId(rst.Appid)
	return msg, nil
}

func GetRechargeStatistic() (int, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return 0, err
	}

	req := &proto.RechargeStatisticReq{}
	rst := &proto.RechargeStatisticRst{}
	if err = conn.Call("PayService.GetRechargeStatistic", req, rst); err != nil {
		return 0, err
	}
	return rst.Value, nil
}
