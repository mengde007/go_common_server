package pushmsg

// import (
// 	"common"
// 	"crypto/md5"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"logger"
// 	"net/http"
// 	"net/url"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// func init() {
// }

// //有某个账号推送消息
// type stPushAccount struct {
// 	AccessId    int    `json:"access_id"`
// 	Timestamp   uint32 `json:"timestamp"`
// 	Sign        string `json:"sign"`
// 	Account     string `json:"account"`
// 	MessageType int    `json:"message_type"` //1：通知 2：透传消息
// 	Message     string `json:"message"`
// 	ExpireTime  int    `json:"expire_time"` //消息离线存储多久，单位为秒，最长存储时间3天。设为0，则不存储
// 	Environment int    `json:"environment"` //向iOS设备推送时必填，1表示推送生产环境；2表示推送开发环境。本字段对Android平台无效
// }

// //登陆返回
// type stPushAccountRet struct {
// 	RetCode int    `json:"ret_code"`
// 	ErrMsg  string `json:"err_msg"`
// 	Result  string `json:"result"`
// }

// type stPushAllRet struct {
// 	RetCode int    `json:"ret_code"`
// 	ErrMsg  string `json:"err_msg"`
// 	Result  struct {
// 		PushId string `json:"push_id"`
// 	} `json:"result"`
// }

// type stBrowserAndroid struct { // url：打开的url，confirm是否需要用户确认
// 	Url     string `json:"url"`
// 	Confirm int    `json:"confirm"`
// }

// type stActAndroid struct {
// 	ActionType int              `json:"action_type"` // 动作类型，1打开activity或app本身，2打开浏览器，3打开Intent
// 	Browser    stBrowserAndroid `json:"browser"`
// 	Activity   string           `json:"activity"`
// 	Intent     string           `json:"intent"`
// }

// type stMsgAndroid struct {
// 	Title   string       `json:"title"`
// 	Content string       `json:"content"`
// 	Ring    int          `json:"ring"`    // 是否响铃，0否，1是，下同
// 	Vibrate int          `json:"vibrate"` // 是否振动，选填，默认0
// 	Action  stActAndroid `json:"action"`
// }

// type stMsgIos struct {
// 	APS struct {
// 		Alert struct {
// 			Body         string `json:"body"`
// 			ActionLocKey string `json:"action-loc-key"`
// 		} `json:"alert"`

// 		Badge int    `json:"badge"`
// 		Sound string `json:"sound"`
// 	} `json:"aps"`
// }

// //android消息
// func genAndroidMsg(strTitle string, strContent string) (string, error) {
// 	stMsgAndroid := stMsgAndroid{
// 		Title:   strTitle,
// 		Content: strContent,
// 		Ring:    1,
// 		Vibrate: 1,
// 		Action: stActAndroid{
// 			ActionType: 1,
// 			Browser: stBrowserAndroid{
// 				Url:     "",
// 				Confirm: 1,
// 			},
// 			Activity: "",
// 			Intent:   "",
// 		},
// 	}

// 	buf, err := json.Marshal(stMsgAndroid)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(buf), nil
// }

// //ios消息
// func genIosMsg(strTitle string, strContent string) (string, error) {
// 	st := stMsgIos{}
// 	st.APS.Alert.Body = strContent
// 	st.APS.Alert.ActionLocKey = "PLAY"
// 	st.APS.Badge = 1
// 	st.APS.Sound = "default"

// 	buf, err := json.Marshal(st)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(buf), nil
// }

// //签名
// func genSig(method, url string, m map[string]string, seckey string) string {
// 	keys := make([]string, 0, len(m))
// 	for k, _ := range m {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)

// 	s := method + url
// 	for _, k := range keys {
// 		s += k + "=" + m[k]
// 	}
// 	s += seckey

// 	h := md5.New()
// 	io.WriteString(h, s)
// 	return fmt.Sprintf("%x", h.Sum(nil))
// }

// func pushMsg(strUser, msg string, accid uint64, baseUrl string, env int, seckey string) error {
// 	timeNow := strconv.FormatInt(time.Now().Unix(), 10)

// 	m := make(map[string]string)
// 	m["access_id"] = strconv.FormatUint(accid, 10)
// 	m["timestamp"] = timeNow
// 	m["account"] = strUser
// 	m["message_type"] = "1" //通知
// 	m["message"] = msg
// 	m["expire_time"] = "600" //延迟10分钟
// 	if env >= 0 {            //android不能传这个，否则会失败
// 		m["environment"] = strconv.Itoa(env)
// 	}
// 	sign := genSig("GET", baseUrl+"/v2/push/single_account", m, seckey)
// 	//logger.Info("pushMsg genSig %v %v %v ", sign, strUser, msg)

// 	fullurl := ""
// 	if env >= 0 {
// 		fullurl = fmt.Sprintf("http://%s/v2/push/single_account?access_id=%s&timestamp=%s&account=%s&message_type=1&message=%s&expire_time=600&environment=%d&sign=%s",
// 			baseUrl,
// 			strconv.FormatUint(accid, 10),
// 			timeNow,
// 			url.QueryEscape(strUser),
// 			url.QueryEscape(msg),
// 			env, sign)
// 	} else {
// 		fullurl = fmt.Sprintf("http://%s/v2/push/single_account?access_id=%s&timestamp=%s&account=%s&message_type=1&message=%s&expire_time=600&sign=%s",
// 			baseUrl,
// 			strconv.FormatUint(accid, 10),
// 			timeNow,
// 			url.QueryEscape(strUser),
// 			url.QueryEscape(msg),
// 			sign)
// 	}
// 	//logger.Info("pushMsg fullurl %v", fullurl)

// 	client := &http.Client{
// 		Transport: common.CreateTransport(),
// 	}
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("pushMsg http.Get error: %v", err)
// 		return err
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("ioutil.ReadAll error:%v", err)
// 		return err
// 	}
// 	//logger.Info("pushMsg http.Get body info:%s", string(b))

// 	rst := stPushAccountRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("pushMsg json.Unmarshal error:%v", err)
// 		return err
// 	}

// 	if rst.RetCode != 0 {
// 		logger.Error("pushMsg error:%v %v", rst.RetCode, rst.ErrMsg)
// 	}

// 	return nil
// }

// func run(strUser string, strTitle string, strContent string) error {
// 	accidA, accidI, baseUrl, env, secA, secI := common.GetXGUrl()
// 	msgAndroid, err := genAndroidMsg(strTitle, strContent)
// 	if err != nil {
// 		logger.Error("genAndroidMsg failed", err)
// 		return err
// 	}
// 	msgIos, err := genIosMsg(strTitle, strContent)
// 	if err != nil {
// 		logger.Error("genIosMsg failed", err)
// 		return err
// 	}
// 	//logger.Info("msgAndroid : %s\nmsgIos : %s", msgAndroid, msgIos)

// 	pushMsg(strUser, msgAndroid, accidA, baseUrl, -1, secA)
// 	pushMsg(strUser, msgIos, accidI, baseUrl, env, secI)

// 	return nil
// }

// //推送消息给指定玩家
// func PushMsg(strUser string, strTitle string, strContent string) {
// 	if strings.Contains(common.GetQQLoginUrl(), "test") {
// 		accidA, _, _, _, _, _ := common.GetXGUrl()
// 		if accidA != 2100064462 { //写死，只要是test一定不能是此id，外网出现过配置错误推送到玩家
// 			logger.Fatal("wrong cfg: test url with official xgid !")
// 			return
// 		}
// 	}

// 	go run(strUser, strTitle, strContent)
// }

// func pushMsg2all(msg string, accid uint64, baseUrl string, env int, seckey string) error {
// 	timeNow := strconv.FormatInt(time.Now().Unix(), 10)

// 	m := make(map[string]string)
// 	m["access_id"] = strconv.FormatUint(accid, 10)
// 	m["timestamp"] = timeNow
// 	m["message_type"] = "1" //通知
// 	m["message"] = msg
// 	m["expire_time"] = "300" //延迟5分钟
// 	if env >= 0 {            //android不能传这个，否则会失败
// 		m["environment"] = strconv.Itoa(env)
// 	}
// 	sign := genSig("GET", baseUrl+"/v2/push/all_device", m, seckey)
// 	//logger.Info("pushMsg genSig %v %v %v ", sign, strUser, msg)

// 	fullurl := ""
// 	if env >= 0 {
// 		fullurl = fmt.Sprintf("http://%s/v2/push/all_device?access_id=%s&timestamp=%s&message_type=1&message=%s&expire_time=300&environment=%d&sign=%s",
// 			baseUrl,
// 			strconv.FormatUint(accid, 10),
// 			timeNow,
// 			url.QueryEscape(msg),
// 			env, sign)
// 	} else {
// 		fullurl = fmt.Sprintf("http://%s/v2/push/all_device?access_id=%s&timestamp=%s&message_type=1&message=%s&expire_time=300&sign=%s",
// 			baseUrl,
// 			strconv.FormatUint(accid, 10),
// 			timeNow,
// 			url.QueryEscape(msg),
// 			sign)
// 	}
// 	//logger.Info("pushMsg fullurl %v", fullurl)

// 	client := &http.Client{
// 		Transport: common.CreateTransport(),
// 	}
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("pushMsg2all http.Get error: %v", err)
// 		return err
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("ioutil.ReadAll error:%v", err)
// 		return err
// 	}
// 	//logger.Info("pushMsg http.Get body info:%s", string(b))

// 	rst := stPushAllRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("pushMsg2all json.Unmarshal error:%v", err)
// 		return err
// 	}

// 	logger.Info("pushMsg stPushAllRet msg ErrMsg is %v, rst.RetCode is %v, rst.Result %v", rst.ErrMsg, rst.RetCode, rst.Result)

// 	if rst.RetCode != 0 {
// 		logger.Error("pushMsg2all error:%v %v", rst.RetCode, rst.ErrMsg)
// 	}

// 	return nil
// }

// func run2all(strTitle string, strContent string) error {
// 	accidA, accidI, baseUrl, env, secA, secI := common.GetXGUrl()
// 	msgAndroid, err := genAndroidMsg(strTitle, strContent)
// 	if err != nil {
// 		logger.Error("genAndroidMsg failed", err)
// 		return err
// 	}
// 	msgIos, err := genIosMsg(strTitle, strContent)
// 	if err != nil {
// 		logger.Error("genIosMsg failed", err)
// 		return err
// 	}
// 	//logger.Info("msgAndroid : %s\nmsgIos : %s", msgAndroid, msgIos)

// 	pushMsg2all(msgAndroid, accidA, baseUrl, -1, secA)
// 	pushMsg2all(msgIos, accidI, baseUrl, env, secI)

// 	return nil
// }

// //向全服玩家发送消息
// func PushMsg2All(strTitle string, strContent string) {
// 	if strings.Contains(common.GetQQLoginUrl(), "test") {
// 		accidA, _, _, _, _, _ := common.GetXGUrl()
// 		if accidA != 2100064462 { //写死，只要是test一定不能是此id，外网出现过配置错误推送到玩家
// 			logger.Fatal("wrong cfg: test url with official xgid !")
// 			return
// 		}
// 	}

// 	if common.GetDesignerCfg().ServerType != common.ServerType_QQ {
// 		// 只有QQ服务器像全服推送信息 (解决重复推送的问题)
// 		return
// 	}

// 	go run2all(strTitle, strContent)
// }
