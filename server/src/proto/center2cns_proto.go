package proto

import (
// "rpc"
)

//扣资源
type ReqCostRes struct {
	PlayerList []string
	ResName    string
	ResValue   int32
	GameType   string //daer, mj, poker
}

type GetCostCache struct {
	Uid string
}

type ReqRechargeNofity struct {
	PlayerList []string
	Buf        []byte
}

type CallCnserverMsg struct {
	Param1 string //PockerEnd 扑克完成
	Uids   []string
}
