package connector

import (
	"common"
	"fmt"
	"logger"
	"reflect"
	"rpc"
	"time"
)

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	rst := fmt.Sprint("", reflect.TypeOf(value))
	if rst != "*rpc.ResourceNotify" {
		// logger.Info("+++=================================+++write :%v", value)
	}

	return common.WriteResult(conn, value)
}

func WriteLoginResult2(conn rpc.RpcConn, r string, login *rpc.Login) bool {
	rep := rpc.LoginResult{}
	rep.SetResult(r)
	rep.SetServerTime(time.Now().UnixNano() / 1e6)
	rep.SetOpenid(login.GetOpenid())
	rep.SetUid(login.GetUid())
	rep.SetRoleId(login.GetRoleId())
	logger.Info("send msg, uid:%s, openId:%s", login.GetUid(), login.GetOpenid())

	return WriteResult(conn, &rep)
}

func WriteLoginResult(conn rpc.RpcConn, r string) bool {
	rep := rpc.LoginResult{}
	rep.SetResult(r)
	rep.SetServerTime(time.Now().UnixNano() / 1e6)
	return WriteResult(conn, &rep)
}

func WriteLoginResultWithErrorMsg(conn rpc.RpcConn, r string, msg string) bool {
	rep := rpc.LoginResult{}
	rep.SetResult(r)
	rep.SetErrmsg(msg)
	rep.SetServerTime(time.Now().UnixNano() / 1e6)
	return WriteResult(conn, &rep)
}

// func WriteMatchResult(conn rpc.RpcConn, r rpc.MatchPlayerResult_Result) bool {
// 	rep := rpc.MatchPlayerResult{}
// 	rep.SetResult(r)
// 	return WriteResult(conn, &rep)
// }

func SyncError(conn rpc.RpcConn, format string, args ...interface{}) {
	common.SyncError(conn, format, args...)
}

func SendMsg(conn rpc.RpcConn, code string) {
	common.SendMsg(conn, code)
}

func SendText(conn rpc.RpcConn, text string) {
	common.SendText(conn, text)
}
