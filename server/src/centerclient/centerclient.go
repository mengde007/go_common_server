package centerclient

import (
	"common"
	"logger"
	"proto"
	"rpcplusclientpool"
)

var pPoll *rpcplusclientpool.ClientPool

func init() {
	aServerHost := common.ReadServerClientConfig("centerserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "centerserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	return
}

//回调
func SetConnectedCallback(f rpcplusclientpool.CALLBACK) {
	pPoll.SetConnectedCallback(f)
}

func Call(method string, req interface{}, rst interface{}) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	return conn.Call(method, req, rst)
}

func Go(method string, req interface{}, rst interface{}) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	conn.Go(method, req, rst, nil)

	return nil
}

//设置最大玩家数量
func SetMaxOnlineNumbers(serverid uint8, numbers int32) (int, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return 0, err
	}

	req := &proto.SetMaxOnlinePlayers{
		ServerId: serverid,
		Numbers:  numbers,
	}
	rst := &proto.SetMaxOnlinePlayersRst{}

	if err := conn.Call("Center.SetMaxOnlineNumbers", req, rst); err != nil {
		return 0, err
	}

	return int(rst.CurNumbers), nil
}

//获得在线玩家数
func GetOnlineNumbers() (int, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return 0, err
	}

	req := &proto.ReqOnlinneNumber{}
	rst := &proto.RstOnlineNumber{}
	if err := conn.Call("Center.GetOnlineNumbers", req, rst); err != nil {
		return 0, err
	}

	return int(rst.Number), nil
}

func CheckCostFromCache(uid string) (*proto.ReqCostRes, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return nil, err
	}

	req := &proto.GetCostCache{
		Uid: uid,
	}
	rst := &proto.ReqCostRes{}
	if err := conn.Call("Center.CheckCostFromCache", req, rst); err != nil {
		return nil, err
	}
	return rst, nil
}

func CallCnserverFunc(req *proto.CallCnserverMsg) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	rst := &proto.CommonRst{}
	if err := conn.Call("Center.CallCnserverFunc", req, rst); err != nil {
		return err
	}
	return nil
}
