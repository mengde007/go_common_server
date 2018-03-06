package common

import (
	"code.google.com/p/goprotobuf/proto"
	"code.google.com/p/snappy-go/snappy"
	"encoding/binary"
	"fmt"
	"logger"
	"net"
	"reflect"
	"rpc"
	// "runtime/debug"
	"time"
)

func WriteResult(conn rpc.RpcConn, value interface{}) bool {
	err := conn.WriteObj(value)
	if err != nil {
		logger.Info("WriteResult Error %s", err.Error())
		// debug.PrintStack()
		return false
	}
	return true
}

//简单发送数据
func SimpleWriteResult(conn net.Conn, value interface{}) error {
	var msg proto.Message

	switch m := value.(type) {
	case proto.Message:
		msg = m
	default:
		return fmt.Errorf("WriteObj value type error %v", value)
	}

	bufvalue, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	req := &rpc.Request{}
	t := reflect.Indirect(reflect.ValueOf(msg)).Type()
	req.SetMethod(t.PkgPath() + "." + t.Name())
	req.SerializedRequest = bufvalue

	buf, err := proto.Marshal(req)
	if err != nil {
		logger.Error("ProtoBufConn Marshal Error %s", err.Error())
		return err
	}

	dst, err := snappy.Encode(nil, buf)
	if err != nil {
		logger.Error("ProtoBufConn snappy.Encode Error %s", err.Error())
		return err
	}

	conn.SetWriteDeadline(time.Now().Add(rpc.ConnWriteTimeOut))
	err = binary.Write(conn, binary.BigEndian, int32(len(dst)))
	if err != nil {
		//logger.Error("ProtoBufConn Write Error %s", err.Error())
		return err
	}

	conn.SetWriteDeadline(time.Now().Add(rpc.ConnWriteTimeOut))
	_, err = conn.Write(dst)
	if err != nil {
		//logger.Error("ProtoBufConn Write Error %s", err.Error())
		return err
	}

	return nil
}

func SyncError(conn rpc.RpcConn, format string, args ...interface{}) {
	//tArgs := make([]interface{}, len(args))
	//for i, arg := range args {
	//	tArgs[i] = arg
	//}
	// logger.Error(format, args...)
	// msg := rpc.SyncError{}
	// msg.SetText(fmt.Sprintf(format, args...))

	// WriteResult(conn, &msg)
}

func SendMsg(conn rpc.RpcConn, code string) {
	msg := rpc.Msg{}
	msg.SetCode(code)

	WriteResult(conn, &msg)
}

func SendText(conn rpc.RpcConn, text string) {
	msg := rpc.Msg{}
	msg.SetText(text)

	WriteResult(conn, &msg)
}

//简单处理，只支持string与int32
func SendFormatedMsg(conn rpc.RpcConn, code string, args ...interface{}) {
	msg := rpc.FormatedMsg{}
	msg.SetCode(code)

	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			msg.Args = append(msg.Args, &rpc.MsgArg{S: &arg})
		case int32:
			msg.Args = append(msg.Args, &rpc.MsgArg{I: &arg})
		default:
			return
		}
	}
	WriteResult(conn, &msg)
}
