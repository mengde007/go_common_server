package proto

import (
// "rpc"
)

type ReqDaerRoom struct {
	Base   []byte
	Client []byte
}

type ReqDaerRoomWithItem struct {
	Base   []byte
	Client []byte
	Amount int32
}

type OperRst struct {
	Ok string
}

type ReqLeaveDaerRoom struct {
	Msg []byte
}

type ReqActionREQ struct {
	Msg []byte
}

type ReqEnterPockerRoom struct {
	GameType string
	EType    int32
	Msg      []byte
}

type ReqCreateCustomRoom struct {
	Uid string
	Msg []byte
}

type ReqWithItem struct {
	Uid    string
	Amount int32
	Msg    []byte
}

type DaerServer2Client struct {
	MsgName string
	Uid     string
	Msg     []byte
}

type ReqOnlineNum struct {
}

type RstOnlineNum struct {
	RoomInfo []byte
}

type ReqIsInRoom struct {
	Uid string
}

type ReqPockerCustomRoom struct {
	Msg     []byte
	BlindId int32
	LimId   int32
}

type RstPockerCustomRoom struct {
	RoomNo int32
}
