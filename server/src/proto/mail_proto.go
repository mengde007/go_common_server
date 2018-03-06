package proto

import (
// "rpc"
)

type MailSendAll struct {
	Title        string
	Content      string
	Attach       string
	Channel      uint32
	ContinueTime uint32
}

//所有用户邮件
type MailSendAllResult struct {
	Success bool
}

type GetMailAttach struct {
	PlayerId string
	MailId   string
}

type GetMailAttachResult struct {
	Attach string
}

type DelPlayerMail struct {
	PlayerId string
	MailId   string
}

type DelPlayerMailResult struct {
}

type SendPlayerMail struct {
	ToPlayerId string
	Title      string
	Content    string
	Attach     string
	ValidTime  uint32
	FromWhere  string
}

type SendPlayerMailResult struct {
}

type SendSystemMail struct {
	ToPlayerId string
	Title      string
	Content    string
	Attach     string
	ValidTime  uint32
	FromWhere  string
}

type SendSystemMailResult struct {
}

type ReadPlayerMail struct {
	PlayerId string
	MailId   string
}

type ReadPlayerMailResult struct {
}

type MailQueryAll struct {
	PlayerId string
	// Channel  rpc.GameLocation
}

type MailQueryAllResult struct {
	Values []byte
}

///////////////////// 系统公告 ////////////////////////
type ChangeSystemNotice struct {
	CanShow   bool
	Title     string
	Content   string
	BeginTime uint32
	EndTime   uint32
}
type ChangeSystemNoticeRst struct {
	Success bool
}

type QuerySystemNotice struct {
}
type QuerySystemNoticeRst struct {
	Values []byte
}

type SetCanShow struct {
	CanShow bool
}
type SetCanShowRst struct {
}

type AddSysNotice struct {
	Id        int64
	Type      uint32
	Title     string
	Content   string
	BeginTime uint32
	EndTime   uint32
	// Platform   rpc.Login_Platform
	Priority   uint32
	CreateTime uint32
}
type AddSysNoticeRst struct {
	Id      int64
	Success bool
}
type QuerySysNotice struct {
	// Platform  rpc.Login_Platform
	BeginTime uint32
	EndTime   uint32
}
type QuerySysNoticeRst struct {
	Success bool
	Values  []byte
}
type DelSysNotice struct {
	Id int64
	// Platform rpc.Login_Platform
}
type DelSysNoticeRst struct {
	Success bool
}

type DelPlayerMailByGM struct {
	PlayerId string
	MailId   string
}

type DelPlayerMailByGmResult struct {
}

/////////// end ///////////////////
