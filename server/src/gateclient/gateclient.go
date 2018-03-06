package gateclient

import (
	"common"
	"logger"
	"proto"
	"rpcplusclientpool"
)

type FailedCallback func()

type GateClient struct {
	serviceName string
	key         string
	callback    FailedCallback
}

var pPoll *rpcplusclientpool.ClientPool

//初始化加锁客户端
func init() {
	var gscfg common.GateServerCfg
	err := common.ReadGateServerConfig(&gscfg)
	if err != nil {
		logger.Fatal("load config failed, error is: %v", err)
		return
	}
	aServerHost := make([]string, 0)
	aServerHost = append(aServerHost, gscfg.GsIpForServer)

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "gateserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	return
}

func SendPlayerCountToGateServer(nCount uint32, sListenIp string) {

	req := proto.SendCnsInfo{uint16(nCount), sListenIp}
	var rst proto.SendCnsInfoResult

	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return
	}

	err = conn.Call("GateServices.UpdateCnsPlayerCount", req, &rst)
	if err != nil {
		return
	}
}
