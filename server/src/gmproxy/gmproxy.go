package gmproxy

import (
	"bytes"
	"common"
	"fmt"
	"io"
	"io/ioutil"
	"logger"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type GmProxy struct{}

var pGmProxy *GmProxy

//回调函数创建
type handleCALLBACK func(w http.ResponseWriter, r *http.Request)

func createHandleFunc(f handleCALLBACK) handleCALLBACK {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				errmsg := fmt.Sprintf("handle http failed :%s", r)
				writeString(w, "serious err occurred :", errmsg)
				return
			}
		}()

		f(w, r)
	}
}
func writeString(w http.ResponseWriter, msgs ...string) {
	for _, msg := range msgs {
		io.WriteString(w, msg)
	}
}

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

func ParseCmdBody(req string) (*StHead, error) {
	var st StHeadParse
	if err := common.JsonDecode([]byte(req), &st); err != nil {
		return nil, err
	}

	return st.Head, nil
}

const (
	REQHEAD = "data_packet="
)

func CreateGmProxy() {
	loadGmServerAddressConfig()

	registerLoadGmServerAddressCfg() // 读取gmserver地址和端口配置表

	wg := sync.WaitGroup{}
	wg.Add(1)

	// 监听http
	go pGmProxy.initHttp(&wg)

	wg.Wait()
}

func (g *GmProxy) initHttp(wg *sync.WaitGroup) error {
	defer wg.Done()

	http.HandleFunc("/", createHandleFunc(g.handle))
	host := mapGmServerAddressCfg["Host"]
	if err := http.ListenAndServe(host, nil); err != nil {
		logger.Error("initHttp ListenAndServe error:", err)
		return err
	}

	return nil
}

func (g *GmProxy) handle(w http.ResponseWriter, r *http.Request) {
	logger.Info("GmProxy handle has been called")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("handle ReadAll body err", err)
		return
	}

	data := string(b)
	// if len(data) < len(REQHEAD) || data[:len(REQHEAD)] != REQHEAD {
	// 	logger.Error("handle wrong request string", data)
	// 	return
	// }

	// data = data[len(REQHEAD):]
	// if data == "" {
	// 	logger.Error("no data_packet")
	// 	return
	// }
	logger.Info("handle: request data: ", data)

	head, err := ParseCmdHeadNew(data)
	if err != nil || head == nil {
		logger.Error("ParseCmdHead failed", err)
		return
	}

	logger.Info("GmProxy called, serverIds:%v", head.Serverid)

	strId := head.Serverid
	serverIds := strings.Split(strId, ",")
	for _, id := range serverIds {
		// 正常的有小区ID
		tmpId, _ := strconv.Atoi(id)
		address, ok := getGmServerAddressCfg(tmpId)
		if !ok {
			logger.Error("Serverid is wrong", id)
			writeError(w, err_code_wrong_param, "Serverid")
			continue
		}
		url := "http://" + address + "/"
		logger.Info("post url:", url)

		client := &http.Client{
			Transport: createTransport(),
		}
		buf := bytes.NewBuffer(b)
		res, err := client.Post(url, "application/x-www-form-urlencoded", buf)
		if err != nil {
			logger.Error("http post error: ", err)
			writeError(w, err_code_network, err.Error())
			return
		}
		defer res.Body.Close()

		resbody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error("resbody ioutil.Readall error:", err)
			writeError(w, err_code_api, err.Error())
			return
		}

		restr := string(resbody)
		logger.Info("retrun writeresult: ", restr)
		io.WriteString(w, restr)
	}
}

//超时连接
func createTransport() *http.Transport {
	return common.CreateTransport()
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

func writeError(w http.ResponseWriter, code int, ext ...string) {
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

	rst := &StRst_Error{}
	rst.Ec = 1
	rst.Em = msg

	logger.Error("writeError err is:", msg)

	writeResult(w, msg)
}
