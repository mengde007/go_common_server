package proto

import (
// "rpc"
)

type ReqGenRoleId struct {
	Uid string
}

type RstGenRoleId struct {
	RoleId  int32
	GuestId int32
	Ok      bool
}

type ReqSearch struct {
	RoleId int32
}

type SearchRst struct {
	Uid string
}

type ReqOnlinneNumber struct {
}

type RstOnlineNumber struct {
	Number uint32
}
