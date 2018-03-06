package payserver

import (
	"common"
	// "encoding/json"
	//"fmt"
	// "bytes"
	"io"
	// "io/ioutil"
	"logger"
	"net/http"
	// "proto"
	"encoding/xml"
)

//超时连接
func createHttpsTransport() *http.Transport {
	return common.CreateHttpsTransport()
}

func writeXmlResult(w http.ResponseWriter, st interface{}) {
	//生成xml
	body, err := xml.MarshalIndent(st, " ", " ")
	if err != nil {
		logger.Error("CreateOrder MarshalIndent error: %v", err)
		return
	}
	// bufBody := bytes.NewBuffer(body)

	//str := "response=" + string(buf)
	str := string(body)
	logger.Info("writeResult:", str)
	io.WriteString(w, str)
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

func writeString(w http.ResponseWriter, msgs ...string) {
	for _, msg := range msgs {
		io.WriteString(w, msg)
	}
}

// func CheckOrder(orderNum, uid, url string) bool {
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}

// 	st := stPaySucessMsg{
// 		Uid:      uid,
// 		ItemId:   "",
// 		OrderNum: orderNum,
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("CheckOrder Marshal stPaySucessMsg error: %v", err)
// 		return false
// 	}

// 	buf2Payproxy := bytes.NewBuffer(body)
// 	res, err := client.Post(url, "application/x-www-form-urlencoded", buf2Payproxy)
// 	if err != nil {
// 		logger.Error("CheckOrder http post error: ", err)
// 		return false
// 	}
// 	defer res.Body.Close()
// 	//payproxy是否处理成功
// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("CheckOrder body info:%s", string(b))

// 	if err != nil {
// 		logger.Error("CheckOrder ioutil.ReadAll error: %v", err)
// 		return false
// 	}

// 	rst := stPaySucessMsgRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("CheckOrder ioutil.ReadAll error: %v", err)
// 		return false
// 	}

// 	if !rst.IsSuccess {
// 		logger.Error("CheckOrder !rst.IsSuccess")
// 		return false
// 	}

// 	return true
// }
