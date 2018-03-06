package proto

import (
// "rpc"
)

//查询
type FriendQueryPFBI struct {
	Uid string
	Del bool
}

type FriendQueryPFBIRst struct {
	Value []byte
}

//通知
type FriendAttackNotice struct {
	Uid   string
	BePVP bool
}
type FriendAttackNoticeRst struct {
}

// 添加/删除好友
type AddFriendRequest struct {
	MyUid     string
	OtherUid  string
	BeAdd     bool
	BeConfirm bool
}
type AddFriendRequestRst struct {
	Success bool
}

type ResponseAddFriend struct {
	MyUid    string
	OtherUid string
	BeAccept bool
	BeAll    bool
}
type ResponseAddFriendRst struct {
}

// 好友操作存储
type OperateFriendInfo struct {
	Uid       string
	BeAdd     bool
	BeConfirm bool
}
type OperateList struct {
	AddList []*OperateFriendInfo
}

//通知
type FriendNoticeUpdate struct {
	Uid string
}

type FriendNoticeUpdateRst struct {
}

type ReqSearchFriend struct {
	RoleId int32
}

type OfflineChatMsg struct {
	Uid   string
	Value []byte
}

type OfflineMsgList struct {
	MsgLst []*OfflineChatMsg
}

type ReqOfflineMsg struct {
	Uid string
}

type CommonRst struct {
	Ok bool
}
