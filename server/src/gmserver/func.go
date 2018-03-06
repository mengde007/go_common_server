package gmserver

import (
	"common"
	"io"
	"logger"
	"net/http"
	// "rpc"
	// "strconv"
	"time"
)

//转换平台id

func ParseCmdHead(req string) (*StHead, error) {
	var st StHeadParse
	if err := common.JsonDecode([]byte(req), &st); err != nil {
		return nil, err
	}

	return st.Head, nil
}

func ParseCmdHeadNew(req string) (*StHeadNew, error) {
	var st StHeadParseNew
	if err := common.JsonDecode([]byte(req), &st); err != nil {
		return nil, err
	}

	return st.Head, nil
}

func writeString(w http.ResponseWriter, msgs ...string) {
	for _, msg := range msgs {
		io.WriteString(w, msg)
	}
}

//已过期接口
func writeCommonError(w http.ResponseWriter, code int, msgs ...string) {
	writeString(w, msgs...)
}

func writeResult(w http.ResponseWriter, st interface{}) {
	buf, err := common.JsonEncode(st)
	if err != nil {
		logger.Error("common.JsonEncode failed", err)
		return
	}

	//str := "response=" + string(buf)
	str := string(buf)
	logger.Info("writeResult:", str)
	io.WriteString(w, str)
}

func writeErrorNew(w http.ResponseWriter, cmdHead *StHeadNew, code int, ext ...string) {
	msg, ok := mapErrStrings[code]
	if !ok {
		msg = "unknown error code"
	}

	//额外参数
	if len(ext) > 0 {
		msg += ":"
		for _, s := range ext {
			msg += " " + s
		}
	}
	rst := &StRst_RoleInfo{}
	rst.Ec = 1
	rst.Em = msg
	rst.Data = []*StRoleInfo{}

	logger.Error("writeError err is:", msg)

	writeResult(w, rst)
}

func writeError(w http.ResponseWriter, cmdHead *StHead, code int, ext ...string) {
	msg, ok := mapErrStrings[code]
	if !ok {
		msg = "unknown error code"
	}

	//额外参数
	if len(ext) > 0 {
		msg += ":"
		for _, s := range ext {
			msg += " " + s
		}
	}
	rst := &StRst_RoleInfo{}
	rst.Ec = 1
	rst.Em = msg
	rst.Data = []*StRoleInfo{}

	logger.Error("writeError err is:", msg)

	writeResult(w, rst)
}

func getTlogTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
