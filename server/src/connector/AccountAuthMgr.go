package connector

// import (
// 	"bytes"
// 	"common"
// 	"crypto/hmac"
// 	"crypto/md5"
// 	"crypto/sha1"
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"logger"
// 	"net/http"
// 	"net/url"
// 	"rpc"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// const (
// 	sLoginErrMsg       = "login failed!"
// 	sPayErrMsg         = "pay failed!"
// 	sQueryErrMsg       = "query failed!"
// 	sBalanceErrMsg     = "query balance failed!"
// 	sFriendsErrMsg     = "query friends failed"
// 	sCancelPayErrMsg   = "cancel pay failed"
// 	sShareErrMsg       = "share failed"
// 	sSendGiftErrMsg    = "send gift failed"
// 	sUpLoadScoreErrMsg = "upload score failed"
// )

// //是否qq渠道
// func isGamelocationQQ(gl rpc.GameLocation) bool {
// 	if gl == rpc.GameLocation_Tencent_Android_QQ || gl == rpc.GameLocation_Tencent_IOS_QQ {
// 		return true
// 	}

// 	return false
// }

// //是否wx渠道
// func isGamelocationWX(gl rpc.GameLocation) bool {
// 	if gl == rpc.GameLocation_Tencent_Android_Weixin || gl == rpc.GameLocation_Tencent_IOS_Weixin {
// 		return true
// 	}

// 	return false
// }

// //超时连接
// func createTransport() *http.Transport {
// 	return common.CreateTransport()
// }

// //encode value
// func encodeValue(v url.Values) string {
// 	if v == nil {
// 		return ""
// 	}

// 	var buf bytes.Buffer
// 	keys := make([]string, 0, len(v))
// 	for k := range v {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)
// 	for _, k := range keys {
// 		vs := v[k]
// 		prefix := k + "="
// 		for _, v := range vs {
// 			if buf.Len() > 0 {
// 				buf.WriteByte('&')
// 			}
// 			buf.WriteString(prefix)
// 			buf.WriteString(v)
// 		}
// 	}

// 	return buf.String()
// }

// //生成支付相关pf
// func genPayPf(p *player) string {
// 	pf := p.mobileqqinfo.Pf + "-"
// 	tencentAppId, _ := common.GetQQAppInfo(false, p.GetGamelocation())
// 	pf += strconv.Itoa(tencentAppId) + "*"
// 	pf += strconv.FormatUint(uint64(p.mobileqqinfo.PlatId), 10) + "*"
// 	pf += p.mobileqqinfo.Openid + "*"
// 	pf += strconv.FormatUint(uint64(p.GetLevel()), 10) + "*"
// 	pf += strconv.FormatUint(uint64(p.mobileqqinfo.LoginChannel), 10) + "*"
// 	pf += "0*0"

// 	return pf
// }

// //头像基础
// func getHeadurlBase(url string) string {
// 	if index := strings.LastIndex(url, "/"); index > 0 {
// 		return url[:index+1]
// 	}

// 	return ""
// }

// //取得支付相关session
// func addCookie(p *player, request *http.Request, urlPath string) {
// 	sessionId, sessionType := "", ""
// 	//微信
// 	if common.IsPlatformWX(p.GetGamelocation()) {
// 		sessionId, sessionType = "hy_gameid", "wc_actoken"
// 	} else if common.IsPlatformGuest(p.GetGamelocation()) {
// 		sessionId, sessionType = "hy_gameid", "st_dummy"
// 	} else {
// 		sessionId, sessionType = "openid", "kp_actoken"
// 	}

// 	request.AddCookie(&http.Cookie{
// 		Name:  "session_id",
// 		Value: url.QueryEscape(sessionId),
// 	})
// 	request.AddCookie(&http.Cookie{
// 		Name:  "session_type",
// 		Value: url.QueryEscape(sessionType),
// 	})
// 	request.AddCookie(&http.Cookie{
// 		Name:  "org_loc",
// 		Value: url.QueryEscape(urlPath),
// 	})
// }

// //手机qq登陆
// //登陆
// type stMobileQQLogin struct {
// 	Appid   int    `json:"appid"`
// 	Openid  string `json:"openid"`
// 	Openkey string `json:"openkey"`
// 	Userip  string `json:"userip"`
// }

// //登陆返回
// type stMobileQQLoginRet struct {
// 	Ret int    `json:"ret"`
// 	Msg string `json:"msg"`
// }

// func MobileQQAuth(login *rpc.Login, IP string) (success bool, errmsg string) {
// 	//微信
// 	if common.IsPlatformWX(login.GetChannelid()) {
// 		return WX_Auth(login, IP)
// 	}

// 	//游客
// 	if common.IsPlatformGuest(login.GetChannelid()) {
// 		return Guest_Auth(login, IP)
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(false, login.GetChannelid())
// 	sUrlBase := common.GetQQLoginUrl()

// 	//当前时间
// 	st := stMobileQQLogin{
// 		Appid:   tencentAppId,
// 		Openid:  login.GetOpenid(),
// 		Openkey: login.GetOpenkey(),
// 		Userip:  IP,
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("MobileQQAuth Marshal stMobileQQLogin error: %v", err)
// 		return false, sLoginErrMsg
// 	}
// 	//logger.Info("MobileQQAuth client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/auth/verify_login/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, login.GetOpenid())
// 	logger.Info("MobileQQAuth tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("MobileQQAuth http.Post error: %v", err)
// 		return false, sLoginErrMsg
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MobileQQAuth body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQAuth ioutil.ReadAll error: %v", err)
// 		return false, sLoginErrMsg
// 	}

// 	rst := stMobileQQLoginRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQAuth ioutil.ReadAll error: %v", err)
// 		return false, sLoginErrMsg
// 	}

// 	if rst.Ret != 0 {
// 		return false, rst.Msg
// 	}

// 	return true, ""
// }

// //查询名字等基础信息
// type stMobileQQQuery struct {
// 	Appid       int    `json:"appid"`
// 	AccessToken string `json:"accessToken"`
// 	Openid      string `json:"openid"`
// }

// type stMobileQQQueryResult struct {
// 	Ret        int    `json:"ret"`
// 	Msg        string `json:"msg"`
// 	NickName   string `json:"nickName"`
// 	Gender     string `json:"gender"`
// 	Picture40  string `json:"picture40"`
// 	Picture100 string `json:"picture100"`
// }

// func MobileQQQuery(p *player) (success bool, errmsg string, nickname string, gender string, picture string) {
// 	//微信
// 	if common.IsPlatformWX(p.GetGamelocation()) {
// 		return WX_Query(p)
// 	}

// 	//返回值初始化
// 	success = false
// 	errmsg = sQueryErrMsg
// 	nickname = ""
// 	gender = ""
// 	picture = ""

// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(false, p.GetGamelocation())
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey

// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//当前时间
// 	st := stMobileQQQuery{
// 		Appid:       tencentAppId,
// 		AccessToken: openkey,
// 		Openid:      openid,
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("MobileQQQuery Marshal stMobileQQLogin error: %v", err)
// 		return
// 	}
// 	//logger.Info("MobileQQQuery client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/relation/qqprofile/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, openid)

// 	logger.Info("MobileQQQuery tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("MobileQQQuery http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MobileQQQuery body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQQuery ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQQueryResult{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQQuery ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""
// 	nickname = rst.NickName
// 	gender = rst.Gender
// 	picture = getHeadurlBase(rst.Picture40)

// 	return
// }

// //好友列表
// //查询名字等基础信息
// type stMobileQQFriends struct {
// 	Appid       int    `json:"appid"`
// 	AccessToken string `json:"accessToken"`
// 	Openid      string `json:"openid"`
// 	Flag        int    `json:"flag"`
// }

// type stMobileQQFriendBase struct {
// 	OpenId   string `json:"openid"`
// 	NickName string `json:"nickName"`
// 	Gender   string `json:"gender"`
// 	Picture  string `json:"figureurl_qq"`
// }

// type stMobileQQFriendsResult struct {
// 	Ret  int                     `json:"ret"`
// 	Msg  string                  `json:"msg"`
// 	List []*stMobileQQFriendBase `json:"lists"`
// }

// func MobileQQFriends(p *player) (success bool, errmsg string, list []*stMobileQQFriendBase) {
// 	//微信
// 	if common.IsPlatformWX(p.GetGamelocation()) {
// 		return WX_Friends(p)
// 	}

// 	//游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		return false, "guest", nil
// 	}

// 	//返回值初始化
// 	success = false
// 	errmsg = sFriendsErrMsg

// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(false, p.GetGamelocation())
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey

// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//当前时间
// 	st := stMobileQQFriends{
// 		Appid:       tencentAppId,
// 		AccessToken: openkey,
// 		Openid:      openid,
// 		Flag:        1,
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("MobileQQFriends Marshal stMobileQQLogin error: %v", err)
// 		return
// 	}
// 	//logger.Info("MobileQQFriends client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/relation/qqfriends_detail/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, openid)
// 	logger.Info("MobileQQFriends tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("MobileQQFriends http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	//logger.Info("MobileQQFriends body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQFriends ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQFriendsResult{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQFriends ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		logger.Error("MobileQQFriends Ret failed", rst)
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""
// 	list = rst.List

// 	return
// }

// //分享
// type stMobileQQShareOneFriend struct {
// 	OpenId string `json:"openid"`
// 	Type   int    `json:"type"`
// }

// type stMobileQQShare struct {
// 	Act         int    `json:"act"`
// 	AppId       int    `json:"oauth_consumer_key"`
// 	Dst         int    `json:"dst"`
// 	Flag        int    `json:"flag"`
// 	ImageUrl    string `json:"image_url"`
// 	OpenId      string `json:"openid"`
// 	OpenKey     string `json:"access_token"`
// 	Src         int    `json:"src"`
// 	Summary     string `json:"summary"`
// 	TargetUrl   string `json:"target_url"`
// 	Title       string `json:"title"`
// 	FopenIds    string `json:"fopenids"`
// 	Appid       int    `json:"appid"`
// 	PreviewText string `json:"previewText"`
// 	GameTag     string `json:"game_tag"`
// }

// type stMobileQQShareResult struct {
// 	Ret int    `json:"ret"`
// 	Msg string `json:"msg"`
// }

// func MobileQQShare(p *player, fopenids []string, msg *rpc.QQShare) (success bool, errmsg string) {
// 	//微信
// 	if common.IsPlatformWX(p.GetGamelocation()) {
// 		return WX_Share(p, fopenids, msg)
// 	}

// 	success = false
// 	errmsg = sShareErrMsg

// 	if p == nil || p.mobileqqinfo == nil || msg == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(false, p.GetGamelocation())
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey

// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	if fopenids == nil || len(fopenids) == 0 {
// 		return
// 	}

// 	//好友列表
// 	list := make([]*stMobileQQShareOneFriend, 0)
// 	for _, id := range fopenids {
// 		one := &stMobileQQShareOneFriend{
// 			OpenId: id,
// 			Type:   0,
// 		}

// 		list = append(list, one)
// 	}
// 	friends, err := json.Marshal(list)
// 	if err != nil {
// 		logger.Error("MobileQQShare Marshal friend error: %v", err)
// 		return
// 	}
// 	logger.Error("MobileQQShare friend list", string(friends))

// 	//请求
// 	st := stMobileQQShare{
// 		Act:      1,
// 		AppId:    tencentAppId,
// 		Dst:      1001,
// 		Flag:     1,
// 		ImageUrl: msg.GetImageUrl(),
// 		OpenId:   openid,
// 		OpenKey:  openkey,
// 		Src:      0,
// 		Summary:  msg.GetSummary(),
// 		// Summary:     "share test",
// 		TargetUrl:   msg.GetTargetUrl(),
// 		Title:       msg.GetTitle(),
// 		FopenIds:    string(friends),
// 		Appid:       tencentAppId,
// 		PreviewText: msg.GetPreviewText(),
// 		GameTag:     msg.GetGameTag(),
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("MobileQQShare Marshal error: %v", err)
// 		return
// 	}
// 	logger.Info("MobileQQShare client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/share/qq/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, openid)
// 	logger.Info("MobileQQShare tx url: ", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("MobileQQShare http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MobileQQShare body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQShare ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQShareResult{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQShare ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""

// 	return
// }

// //////////////////////////////////////////////////////////
// //QQ支付相关
// //////////////////////////////////////////////////////////
// //支付返回
// type stMobileQQPayRet struct {
// 	Ret      int    `json:"ret"`
// 	Msg      string `json:"msg"`
// 	BillNo   string `json:"billno"`
// 	Balance  int    `json:"balance"`
// 	Giftused int    `json:"used_gen_amt"`
// }

// //cookies
// type Jar struct {
// 	cookies []*http.Cookie
// }

// func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
// 	jar.cookies = cookies
// }

// func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
// 	return jar.cookies
// }

// func MobileQQPay(p *player, number int) (success bool, errmsg string, billno string, balance, giftused int) {
// 	success, errmsg, billno, balance, giftused = _MobileQQPay(p, number, false)
// 	if !success {
// 		success, errmsg, billno, balance, giftused = _MobileQQPay(p, number, true)
// 	}

// 	return
// }

// func _MobileQQPay(p *player, number int, backup bool) (success bool, errmsg string, billno string, balance, giftused int) {
// 	if p == nil || p.mobileqqinfo == nil {
// 		return false, sPayErrMsg, "", 0, 0
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(true, p.GetGamelocation())
// 	sUrlBase, zoneId := common.GetPayUrlAndZoneId(backup)

// 	openid := p.mobileqqinfo.Openid
// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//路径
// 	_, urlPath, _, _ := common.GetQQPayPath()

// 	openkey := p.mobileqqinfo.Openkey
// 	//游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		openkey = "openkey"
// 	}

// 	pay_token := p.mobileqqinfo.Pay_token
// 	//pfkey := p.mobileqqinfo.Pfkey
// 	pfkey := "pfkey"
// 	//pf := p.mobileqqinfo.Pf
// 	pf := genPayPf(p)
// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	//ip := p.conn.GetRemoteIp()
// 	amt := strconv.FormatInt(int64(number), 10)

// 	v := make(url.Values)
// 	v.Add("openid", openid)
// 	v.Add("openkey", openkey)
// 	v.Add("pay_token", pay_token)
// 	v.Add("appid", strconv.FormatInt(int64(tencentAppId), 10))
// 	v.Add("ts", timecur)
// 	v.Add("pf", pf)
// 	v.Add("format", "json")
// 	//v.Add("userip", ip)
// 	v.Add("zoneid", strconv.Itoa(zoneId))
// 	v.Add("amt", amt)
// 	v.Add("pfkey", pfkey)

// 	sigurl := "GET&" + url.QueryEscape(urlPath) + "&" + url.QueryEscape(encodeValue(v))
// 	sigappkey := tencentAppKey + "&"
// 	h := sha1.New()
// 	io.WriteString(h, sigurl)
// 	mac := hmac.New(sha1.New, []byte(sigappkey))
// 	mac.Write([]byte(sigurl))
// 	dec := fmt.Sprintf("%s", mac.Sum(nil))
// 	sig := fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte(dec)))

// 	v.Add("sig", sig)

// 	//pf不编码
// 	v.Del("pf")
// 	fullurl := sUrlBase + urlPath + "?" + v.Encode() + "&pf=" + pf
// 	logger.Info("MobileQQPay tx url: ", fullurl)

// 	request, err := http.NewRequest("GET", fullurl, nil)
// 	if err != nil {
// 		logger.Error("MobileQQPay http.NewRequest error: %v", err)
// 		return false, sPayErrMsg, "", 0, 0
// 	}

// 	//添加cookie
// 	addCookie(p, request, urlPath)

// 	jar := &Jar{cookies: make([]*http.Cookie, 0)}
// 	client := &http.Client{
// 		Transport: createTransport(),
// 		Jar:       jar,
// 	}

// 	resp, err := client.Do(request)
// 	if err != nil {
// 		logger.Error("MobileQQPay client.Do error: %v", err)
// 		return false, sPayErrMsg, "", 0, 0
// 	}

// 	b, err := ioutil.ReadAll(resp.Body)
// 	logger.Info("MobileQQPay body info:%s", string(b))
// 	resp.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQPay ioutil.ReadAll error: %v", err)
// 		return false, sPayErrMsg, "", 0, 0
// 	}

// 	rst := stMobileQQPayRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQPay ioutil.ReadAll error: %v", err)
// 		return false, sPayErrMsg, "", 0, 0
// 	}

// 	/*0：成功；
// 	1004：余额不足。
// 	1018：登陆校验失败。
// 	其它：失败*/
// 	if rst.Ret != 0 {
// 		logger.Error("MobileQQPay rst.Ret != 0. failed! rst: ", rst.Ret, rst.Msg)
// 		return false, rst.Msg, "", 0, 0
// 	}

// 	return true, "", rst.BillNo, rst.Balance, rst.Giftused
// }

// //查询余额
// type stMobileQQBalanceRet struct {
// 	Ret int `json:"ret"`
// 	//总游戏币个数，包括赠送
// 	Balance int `json:"balance"`
// 	//赠送游戏币个数
// 	GenBalance int   `json:"gen_balance"`
// 	FirstSave  int   `json:"first_save"`
// 	SaveAmt    int64 `json:"save_amt"`
// }

// func MobileQQBalance(p *player) (success bool, errmsg string, number, numbergift int, totalGemNum int64) {
// 	success, errmsg, number, numbergift, totalGemNum = _MobileQQBalance(p, false)
// 	if !success {
// 		success, errmsg, number, numbergift, totalGemNum = _MobileQQBalance(p, true)
// 	}

// 	return
// }

// func _MobileQQBalance(p *player, backup bool) (success bool, errmsg string, number, numbergift int, totalGemNum int64) {
// 	success = false
// 	errmsg = sBalanceErrMsg
// 	number = 0
// 	numbergift = 0
// 	totalGemNum = 0

// 	//todo 暂时屏蔽
// 	//success = true
// 	//return

// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(true, p.GetGamelocation())
// 	sUrlBase, zoneId := common.GetPayUrlAndZoneId(backup)

// 	openid := p.mobileqqinfo.Openid
// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//路径
// 	urlPath, _, _, _ := common.GetQQPayPath()

// 	openkey := p.mobileqqinfo.Openkey
// 	//游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		openkey = "openkey"
// 	}

// 	pay_token := p.mobileqqinfo.Pay_token
// 	//pfkey := p.mobileqqinfo.Pfkey
// 	pfkey := "pfkey"
// 	pf := p.mobileqqinfo.Pf
// 	//pf := genPayPf(p)
// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	//ip := p.conn.GetRemoteIp()

// 	v := make(url.Values)
// 	v.Add("openid", openid)
// 	v.Add("openkey", openkey)
// 	v.Add("pay_token", pay_token)
// 	v.Add("appid", strconv.FormatInt(int64(tencentAppId), 10))
// 	v.Add("ts", timecur)
// 	v.Add("pf", pf)
// 	v.Add("format", "json")
// 	//v.Add("userip", ip)
// 	v.Add("zoneid", strconv.Itoa(zoneId))
// 	v.Add("pfkey", pfkey)

// 	sigurl := "GET&" + url.QueryEscape(urlPath) + "&" + url.QueryEscape(encodeValue(v))
// 	sigappkey := tencentAppKey + "&"
// 	h := sha1.New()
// 	io.WriteString(h, sigurl)
// 	mac := hmac.New(sha1.New, []byte(sigappkey))
// 	mac.Write([]byte(sigurl))
// 	dec := fmt.Sprintf("%s", mac.Sum(nil))
// 	sig := fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte(dec)))

// 	v.Add("sig", sig)

// 	fullurl := sUrlBase + urlPath + "?" + v.Encode()
// 	logger.Info("MobileQQBalance tx url: ", fullurl)

// 	request, err := http.NewRequest("GET", fullurl, nil)
// 	if err != nil {
// 		logger.Error("MobileQQBalance http.NewRequest error: %v", err)
// 		return
// 	}

// 	//添加cookie
// 	addCookie(p, request, urlPath)

// 	jar := &Jar{cookies: make([]*http.Cookie, 0)}
// 	client := &http.Client{
// 		Transport: createTransport(),
// 		Jar:       jar,
// 	}

// 	resp, err := client.Do(request)
// 	if err != nil {
// 		logger.Error("MobileQQBalance client.Do error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(resp.Body)
// 	logger.Info("MobileQQBalance body info:%s", string(b))
// 	resp.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQBalance ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQBalanceRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQBalance ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	/*0：成功；
// 	1001：参数错误
// 	1018：登陆校验失败。*/
// 	if rst.Ret != 0 {
// 		logger.Error("MobileQQBalance rst.Ret != 0: ", rst)
// 		return
// 	}

// 	success = true
// 	errmsg = ""
// 	number = rst.Balance - rst.GenBalance
// 	numbergift = rst.GenBalance
// 	totalGemNum = rst.SaveAmt

// 	return
// }

// //查询余额
// type stMobileQQCancelPayRet struct {
// 	Ret int `json:"ret"`
// 	Msg int `json:"msg"`
// }

// func MobileQQCancelPay(p *player, billno string, amt int) (success bool, errmsg string) {
// 	success, errmsg = _MobileQQCancelPay(p, billno, amt, false)
// 	if !success {
// 		success, errmsg = _MobileQQCancelPay(p, billno, amt, true)
// 	}

// 	return
// }

// func _MobileQQCancelPay(p *player, billno string, amt int, backup bool) (success bool, errmsg string) {
// 	success = false
// 	errmsg = sCancelPayErrMsg

// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(true, p.GetGamelocation())
// 	sUrlBase, zoneId := common.GetPayUrlAndZoneId(backup)

// 	openid := p.mobileqqinfo.Openid
// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//路径
// 	_, _, urlPath, _ := common.GetQQPayPath()

// 	openkey := p.mobileqqinfo.Openkey
// 	//游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		openkey = "openkey"
// 	}

// 	pay_token := p.mobileqqinfo.Pay_token
// 	//pfkey := p.mobileqqinfo.Pfkey
// 	pfkey := "pfkey"
// 	//pf := p.mobileqqinfo.Pf
// 	pf := genPayPf(p)
// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	//ip := p.conn.GetRemoteIp()

// 	v := make(url.Values)
// 	v.Add("openid", openid)
// 	v.Add("openkey", openkey)
// 	v.Add("pay_token", pay_token)
// 	v.Add("appid", strconv.FormatInt(int64(tencentAppId), 10))
// 	v.Add("ts", timecur)
// 	v.Add("pf", pf)
// 	v.Add("format", "json")
// 	//v.Add("userip", ip)
// 	v.Add("zoneid", strconv.Itoa(zoneId))
// 	v.Add("pfkey", pfkey)
// 	v.Add("amt", strconv.FormatInt(int64(amt), 10))
// 	v.Add("billno", billno)

// 	sigurl := "GET&" + url.QueryEscape(urlPath) + "&" + url.QueryEscape(encodeValue(v))
// 	sigappkey := tencentAppKey + "&"
// 	h := sha1.New()
// 	io.WriteString(h, sigurl)
// 	mac := hmac.New(sha1.New, []byte(sigappkey))
// 	mac.Write([]byte(sigurl))
// 	dec := fmt.Sprintf("%s", mac.Sum(nil))
// 	sig := fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte(dec)))

// 	v.Add("sig", sig)

// 	//pf不编码
// 	v.Del("pf")
// 	fullurl := sUrlBase + urlPath + "?" + v.Encode() + "&pf=" + pf
// 	logger.Info("MobileQQCancelPay tx url: ", fullurl)

// 	request, err := http.NewRequest("GET", fullurl, nil)
// 	if err != nil {
// 		logger.Error("MobileQQCancelPay http.NewRequest error: %v", err)
// 		return
// 	}

// 	//添加cookie
// 	addCookie(p, request, urlPath)

// 	jar := &Jar{cookies: make([]*http.Cookie, 0)}
// 	client := &http.Client{
// 		Transport: createTransport(),
// 		Jar:       jar,
// 	}

// 	resp, err := client.Do(request)
// 	if err != nil {
// 		logger.Error("MobileQQCancelPay client.Do error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(resp.Body)
// 	logger.Info("MobileQQCancelPay body info:%s", string(b))
// 	resp.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQCancelPay ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQCancelPayRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQCancelPay ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	/*0：成功；
// 	1001：参数错误
// 	1018：登陆校验失败。*/
// 	if rst.Ret != 0 {
// 		return
// 	}

// 	success = true
// 	errmsg = ""

// 	return
// }

// //宝石赠送
// type stMobileQQSendGiftRet struct {
// 	Ret     int `json:"ret"`
// 	Balance int `json:"balance"`
// }

// func MobileQQSendGift(p *player, amt int) (success bool, errmsg string, numbers int) {
// 	success, errmsg, numbers = _MobileQQSendGift(p, amt, false)
// 	if !success {
// 		success, errmsg, numbers = _MobileQQSendGift(p, amt, true)
// 	}

// 	return
// }

// func _MobileQQSendGift(p *player, amt int, backup bool) (success bool, errmsg string, numbers int) {
// 	success = false
// 	errmsg = sSendGiftErrMsg
// 	numbers = 0

// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(true, p.GetGamelocation())
// 	sUrlBase, zoneId := common.GetPayUrlAndZoneId(backup)
// 	giftAccId, giftGiftId := common.GetQQSendGiftId(p.GetGamelocation())

// 	openid := p.mobileqqinfo.Openid
// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//路径
// 	_, _, _, urlPath := common.GetQQPayPath()

// 	openkey := p.mobileqqinfo.Openkey
// 	//游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		openkey = "openkey"
// 	}

// 	pay_token := p.mobileqqinfo.Pay_token
// 	//pfkey := p.mobileqqinfo.Pfkey
// 	pfkey := "pfkey"
// 	//pf := p.mobileqqinfo.Pf
// 	pf := genPayPf(p)
// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	//ip := p.conn.GetRemoteIp()

// 	v := make(url.Values)
// 	v.Add("openid", openid)
// 	v.Add("openkey", openkey)
// 	v.Add("pay_token", pay_token)
// 	v.Add("appid", strconv.FormatInt(int64(tencentAppId), 10))
// 	v.Add("ts", timecur)
// 	v.Add("pf", pf)
// 	v.Add("format", "json")
// 	//v.Add("userip", ip)
// 	v.Add("zoneid", strconv.Itoa(zoneId))
// 	v.Add("pfkey", pfkey)
// 	v.Add("discountid", giftAccId)
// 	v.Add("giftid", giftGiftId)
// 	v.Add("presenttimes", strconv.FormatInt(int64(amt), 10))

// 	sigurl := "GET&" + url.QueryEscape(urlPath) + "&" + url.QueryEscape(encodeValue(v))
// 	sigappkey := tencentAppKey + "&"
// 	h := sha1.New()
// 	io.WriteString(h, sigurl)
// 	mac := hmac.New(sha1.New, []byte(sigappkey))
// 	mac.Write([]byte(sigurl))
// 	dec := fmt.Sprintf("%s", mac.Sum(nil))
// 	sig := fmt.Sprintf("%s", base64.StdEncoding.EncodeToString([]byte(dec)))

// 	v.Add("sig", sig)

// 	//pf不编码
// 	v.Del("pf")
// 	fullurl := sUrlBase + urlPath + "?" + v.Encode() + "&pf=" + pf
// 	logger.Info("MobileQQSendGift tx url: ", fullurl)

// 	request, err := http.NewRequest("GET", fullurl, nil)
// 	if err != nil {
// 		logger.Error("MobileQQSendGift http.NewRequest error: %v", err)
// 		return
// 	}

// 	//添加cookie
// 	addCookie(p, request, urlPath)

// 	jar := &Jar{cookies: make([]*http.Cookie, 0)}
// 	client := &http.Client{
// 		Transport: createTransport(),
// 		Jar:       jar,
// 	}

// 	resp, err := client.Do(request)
// 	if err != nil {
// 		logger.Error("MobileQQSendGift client.Do error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(resp.Body)
// 	logger.Info("MobileQQSendGift body info:%s", string(b))
// 	resp.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQSendGift ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQSendGiftRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQSendGift ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		return
// 	}

// 	success = true
// 	errmsg = ""
// 	numbers = rst.Balance

// 	return
// }

// //vip查询
// const (
// 	VIP_NORMAL = 1  //(会员)
// 	VIP_BLUE   = 4  //（蓝钻）
// 	VIP_RED    = 8  //（红钻）
// 	VIP_SUPER  = 16 //（超级会员）
// )

// type stReqMobileQQVip struct {
// 	Appid       string `json:"appid"`
// 	Login       int    `json:"login"`
// 	Uin         int    `json:"uin"`
// 	Openid      string `json:"openid"`
// 	Vip         int    `json:"vip"`
// 	AccessToken string `json:"accessToken"`
// }

// type stRspMobileQQVip_Info struct {
// 	Flag   int `json:"flag"`
// 	Year   int `json:"year"`
// 	Level  int `json:"level"`
// 	Luxury int `json:"luxury"`
// 	Isvip  int `json:"isvip"`
// }

// type stRspMobileQQVip struct {
// 	Ret   int                      `json:"ret"`
// 	Msg   string                   `json:"msg"`
// 	Lists []*stRspMobileQQVip_Info `json:"lists"`
// }

// func MobileQQVip(p *player, bVerifyServer bool) (success bool, errmsg string, vip bool, svip bool) {
// 	//审核服
// 	if bVerifyServer {
// 		return true, "", false, false
// 	}

// 	//微信
// 	if common.IsPlatformWX(p.GetGamelocation()) {
// 		return true, "", false, false
// 	}

// 	//游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		return true, "", false, false
// 	}

// 	//返回值初始化
// 	success = false
// 	errmsg = "query vip failed"
// 	vip = false
// 	svip = false
// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(false, p.GetGamelocation())
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey

// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//当前时间
// 	st := stReqMobileQQVip{
// 		Appid:       strconv.Itoa(tencentAppId),
// 		Login:       2,
// 		Uin:         0,
// 		Openid:      openid,
// 		Vip:         VIP_NORMAL + VIP_SUPER,
// 		AccessToken: openkey,
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("MobileQQVip Marshal stMobileQQLogin error: %v", err)
// 		return
// 	}
// 	//logger.Info("MobileQQVip client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/profile/query_vip/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, openid)
// 	logger.Info("MobileQQVip tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("MobileQQVip http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MobileQQVip body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQVip ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stRspMobileQQVip{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQVip ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		if rst.Ret == -103 && strings.Contains(sUrlBase, "test") {
// 			logger.Error("MobileQQVip rst.Ret == -103 && is test url", rst.Ret, errmsg)
// 			rst.Ret = 0
// 			success = true
// 			return
// 		}
// 		errmsg = rst.Msg
// 		logger.Error("MobileQQVip rst.Ret wrong", rst.Ret, errmsg)
// 		return
// 	}

// 	for _, info := range rst.Lists {

// 		p.LogInfo("*******MobileQQVip return result info from list", info)

// 		if info.Flag == VIP_NORMAL {
// 			success = true
// 			errmsg = ""
// 			vip = info.Isvip == 1
// 		} else if info.Flag == VIP_SUPER {
// 			success = true
// 			errmsg = ""
// 			svip = info.Isvip == 1
// 		}
// 	}

// 	return
// }

// // QQ平台上报分数 成就
// type stMobileQQScoreReq struct {
// 	AppId       string `json:"appid"`
// 	AccessToken string `json:"accessToken"`
// 	OpenId      string `json:"openid"`
// 	Data        string `json:"data"`
// 	Type        int    `json:"type"`
// 	Bcover      int    `json:"bcover"`
// 	Expires     string `json:"expires"`
// }

// type stMobileQQScoreRst struct {
// 	Ret  int    `json:"ret"`
// 	Msg  string `json:"msg"`
// 	Type int    `json:"type"`
// }

// func MobileQQUpLoadScore(p *player) (success bool, errmsg string) {
// 	// 微信
// 	if common.IsPlatformWX(p.GetGamelocation()) {
// 		return WXUpLoadScore(p)
// 	}
// 	// 游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		return true, ""
// 	}

// 	success = false
// 	errmsg = sUpLoadScoreErrMsg
// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}
// 	tencentAppId, tencentAppKey := common.GetQQAppInfo(false, p.GetGamelocation())
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey
// 	if len(openid) == 0 {
// 		return
// 	}
// 	data := strconv.FormatInt(int64(p.GetTrophy()), 10)
// 	endTime := strconv.FormatInt(time.Now().Unix()+14*3600, 10)
// 	appId := strconv.FormatInt(int64(tencentAppId), 10)
// 	st := stMobileQQScoreReq{
// 		AppId:       appId,
// 		AccessToken: openkey,
// 		OpenId:      openid,
// 		Data:        data,
// 		Type:        3,
// 		Bcover:      1,
// 		Expires:     endTime,
// 	}
// 	logger.Info("###################QQupload: json: ", st)
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("MobileQQUpLoadScore Marshal stMobileQQScoreReq error: %v", err)
// 		return
// 	}
// 	buf := bytes.NewBuffer(body)

// 	timeCur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timeCur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/profile/qqscore/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timeCur, tencentAppId, sig, openid)
// 	logger.Info("MobileQQUpLoadScore tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("MobileQQUpLoadScore http.Post error: %v", err)
// 		return
// 	}

// 	// /////////////////////////////
// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MobileQQ uploadscore readall body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MobileQQUpLoadScore ioutil.ReadAll error: %v", err)
// 		return
// 	}
// 	rst := stMobileQQScoreRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MobileQQUpLoadScore json.Unmarshal error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		logger.Error("MobileQQUpLoadScore rst.Ret wrong", rst.Ret, errmsg)
// 		return
// 	}

// 	// ////////////////////////////
// 	data2 := strconv.FormatInt(int64(p.GetLevel()), 10)
// 	st2 := stMobileQQScoreReq{
// 		AppId:       appId,
// 		AccessToken: openkey,
// 		OpenId:      openid,
// 		Data:        data2,
// 		Type:        1,
// 		Bcover:      1,
// 		Expires:     endTime,
// 	}

// 	body2, err2 := json.Marshal(st2)
// 	if err2 != nil {
// 		logger.Error("MobileQQUpLoadScore Marshal stMobileQQScoreReq error: %v", err2)
// 		return
// 	}
// 	buf2 := bytes.NewBuffer(body2)

// 	timeCur2 := strconv.FormatInt(time.Now().Unix(), 10)
// 	h2 := md5.New()
// 	io.WriteString(h2, tencentAppKey)
// 	io.WriteString(h2, timeCur2)
// 	sig2 := fmt.Sprintf("%x", h2.Sum(nil))
// 	fullurl2 := fmt.Sprintf("%s/profile/qqscore/?timestamp=%s&appid=%d&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timeCur2, tencentAppId, sig2, openid)
// 	logger.Info("MobileQQUpLoadScore tx url: %v", fullurl2)

// 	client2 := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res2, errlevel := client2.Post(fullurl2, "application/x-www-form-urlencoded", buf2)
// 	if errlevel != nil {
// 		logger.Error("MobileQQUpLoadScore http.Post error: %v", errlevel)
// 		return
// 	}
// 	res2.Body.Close()

// 	success = true
// 	errmsg = ""
// 	return
// }

// /*
// **********************************************
// 微信接口
// **********************************************
// */
// const (
// 	sWxLoginErrMsg       = "WX login failed!"
// 	sWxPayErrMsg         = "WX pay failed!"
// 	sWxQueryErrMsg       = "WX query failed!"
// 	sWxBalanceErrMsg     = "WX query balance failed!"
// 	sWxFriendsErrMsg     = "WX query friends failed"
// 	sWxCancelPayErrMsg   = "WX cancel pay failed"
// 	sWxShareErrMsg       = "WX share failed"
// 	sWxSendGiftErrMsg    = "WX send gift failed"
// 	sWxUpLoadScoreErrMsg = "WX upload score failed"
// )

// type stWxCommonReq struct {
// 	AccessToken string `json:"accessToken"`
// 	OpenId      string `json:"openid"`
// }

// type stWxLoginRst struct {
// 	Ret string `json:"ret"`
// 	Msg string `json:"msg"`
// }

// func WX_Auth(login *rpc.Login, IP string) (success bool, errmsg string) {
// 	success = false
// 	errmsg = sWxLoginErrMsg

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetWXAppInfo()
// 	sUrlBase := common.GetQQLoginUrl()

// 	//当前时间
// 	st := stWxCommonReq{
// 		AccessToken: login.GetOpenkey(),
// 		OpenId:      login.GetOpenid(),
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("WX_Auth Marshal stMobileQQLogin error: %v", err)
// 		return
// 	}

// 	logger.Info("WX_Auth client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/auth/check_token/?timestamp=%s&appid=%s&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, login.GetOpenid())
// 	logger.Info("WX_Auth tx url: ", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("WX_Auth http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("WX_Auth body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("WX_Auth ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQLoginRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("WX_Auth ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""

// 	return
// }

// //查询名字等基础信息
// type stWXQuery struct {
// 	AccessToken string `json:"accessToken"`
// 	OpenId      string `json:"openid"`
// 	AppId       string `json:"appid"`
// }

// type stWXQueryResult struct {
// 	Ret      int    `json:"ret"`
// 	Msg      string `json:"msg"`
// 	NickName string `json:"nickname"` //昵称
// 	Sex      string `json:"sex"`      //性别1男2女,用户未填则默认1男
// 	Picture  string `json:"picture"`  //用户头像URL,必须在URL后追加以下参数/0，/132，/96，/64，这样可以分别获得不同规格的图片：原始图片(/0)、132*132(/132)、96*96(/96)、64*64(/64)、46*46(/46)
// }

// func WX_Query(p *player) (success bool, errmsg string, nickname string, gender string, picture string) {
// 	//返回值初始化
// 	success = false
// 	errmsg = sWxQueryErrMsg
// 	nickname = ""
// 	gender = ""
// 	picture = ""

// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetWXAppInfo()
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey

// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//当前时间
// 	st := stWXQuery{
// 		AccessToken: openkey,
// 		OpenId:      openid,
// 		AppId:       tencentAppId,
// 	}

// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("WX_Query Marshal stMobileQQLogin error: %v", err)
// 		return
// 	}
// 	//logger.Info("WX_Query client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/relation/wxuserinfo/?timestamp=%s&appid=%s&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, openid)

// 	logger.Info("WX_Query tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("WX_Query http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("WX_Query body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("WX_Query ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stWXQueryResult{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("WX_Query ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""
// 	nickname = rst.NickName
// 	gender = rst.Sex
// 	picture = getHeadurlBase(rst.Picture + "/")
// 	// logger.Error("!!!!!!!!!!!!!!!!!!: ", picture, rst.Picture)
// 	return
// }

// // 微信分享
// type stWXShareUpLoad struct {
// 	Flag        int    `json:"flag"`
// 	AppId       string `json:"appid"`
// 	AppKey      string `json:"secret"`
// 	AccessToken string `json:"access_token"`
// 	Type        string `json:"type"`
// 	FileName    string `json:"filename"`
// 	FileLength  int    `json:"filelength"`
// 	ContentTYpe string `json:"content_type"`
// 	Binary      string `json:"binary"`
// }
// type stWXShareUpLoadRet struct {
// 	Ret         int    `json:"ret"`
// 	Msg         string `json:"msg"`
// 	Type        string `json:"type"`
// 	MediaID     string `json:"media_id"`
// 	CreatedAt   string `json:"created_at"`
// 	AccessToken string `json:"access_token"`
// 	Expire      string `json:"expire"`
// }

// type stWXShare struct {
// 	OpenId       string `json:"openid"`
// 	FopenId      string `json:"fopenid"`
// 	AccessToken  string `json:"access_token"`
// 	ExtInfo      string `json:"extinfo"`
// 	Title        string `json:"title"`
// 	Description  string `json:"description"`
// 	MediaTagName string `json:"media_tag_name"`
// 	ThumbMediaId string `json:"thumb_media_id"`
// }
// type stWXShareRet struct {
// 	Ret int    `json:"ret"`
// 	Msg string `json:"msg"`
// }

// func WX_Share(p *player, fopenids []string, msg *rpc.QQShare) (success bool, errmsg string) {
// 	success = false
// 	errmsg = sShareErrMsg

// 	if p == nil || p.mobileqqinfo == nil || msg == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetWXAppInfo()
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey

// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	if fopenids == nil || len(fopenids) == 0 {
// 		return
// 	}

// 	//请求
// 	st := stWXShare{
// 		OpenId:       openid,
// 		FopenId:      fopenids[0],
// 		AccessToken:  openkey,
// 		ExtInfo:      "",
// 		Title:        msg.GetTitle(),
// 		Description:  msg.GetSummary(),
// 		MediaTagName: msg.GetGameTag(),
// 		ThumbMediaId: "",
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("WXShare Marshal error: %v", err)
// 		return
// 	}
// 	logger.Info("WXShare client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/share/wx/?timestamp=%s&appid=%s&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, openid)
// 	logger.Info("WXShare tx url: ", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("WXShare http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("WXShare body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("WXShare ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQShareResult{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("WXShare ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""

// 	return

// }

// //微信好友
// type stWXFriendInfo struct {
// 	NickName string `json:"nickName"` //昵称
// 	Sex      int    `json:"sex"`      //性别1男2女,用户未填则默认1男
// 	Picture  string `json:"picture"`  //用户头像URL,必须在URL后追加以下参数/0，/132，/96，/64，这样可以分别获得不同规格的图片：原始图片(/0)、132*132(/132)、96*96(/96)、64*64(/64)、46*46(/46)
// 	OpenId   string `json:"openid"`   //用户标识
// }

// type stWxFriendsResult struct {
// 	Ret   int               `json:"ret"`
// 	Msg   string            `json:"msg"`
// 	Lists []*stWXFriendInfo `json:"lists"`
// }

// func WX_Friends(p *player) (success bool, errmsg string, list []*stMobileQQFriendBase) {
// 	//返回值初始化
// 	success = false
// 	errmsg = sWxFriendsErrMsg

// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}

// 	//配置表
// 	tencentAppId, tencentAppKey := common.GetWXAppInfo()
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	openkey := p.mobileqqinfo.Openkey

// 	//非qq渠道就不用走这个流程了
// 	if len(openid) == 0 {
// 		return
// 	}

// 	//当前时间
// 	st := stWXQuery{
// 		AccessToken: openkey,
// 		OpenId:      openid,
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("WX_Friends Marshal stMobileQQLogin error: %v", err)
// 		return
// 	}
// 	//logger.Info("WX_Friends client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/relation/wxfriends_profile/?timestamp=%s&appid=%s&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, openid)
// 	logger.Info("WX_Friends tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("WX_Friends http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	//logger.Info("WX_Friends body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("WX_Friends ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stWxFriendsResult{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("WX_Friends ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		logger.Error("WX_Friends Ret failed", rst)
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""
// 	list = make([]*stMobileQQFriendBase, 0)
// 	for _, info := range rst.Lists {
// 		list = append(list, &stMobileQQFriendBase{
// 			OpenId:   info.OpenId,
// 			NickName: info.NickName,
// 			Gender:   strconv.Itoa(info.Sex),
// 			Picture:  info.Picture + "/",
// 		})
// 	}

// 	return
// }

// // WX平台上报分数
// type stWXScoreReq struct {
// 	AppId     string `json:"appid"`
// 	GrantType string `json:"grantType"`
// 	OpenId    string `json:"openid"`
// 	Score     string `json:"score"`
// 	Expires   string `json:"expires"`
// }

// type stWXScoreRst struct {
// 	Ret int    `json:"ret"`
// 	Msg string `json:"msg"`
// }

// func WXUpLoadScore(p *player) (success bool, errmsg string) {
// 	// 游客
// 	if common.IsPlatformGuest(p.GetGamelocation()) {
// 		return true, ""
// 	}

// 	success = false
// 	errmsg = sWxUpLoadScoreErrMsg
// 	if p == nil || p.mobileqqinfo == nil {
// 		return
// 	}
// 	tencentAppId, tencentAppKey := common.GetWXAppInfo()
// 	sUrlBase := common.GetQQLoginUrl()

// 	openid := p.mobileqqinfo.Openid
// 	// openkey := p.mobileqqinfo.Openkey
// 	if len(openid) == 0 {
// 		return
// 	}
// 	data := strconv.FormatInt(int64(p.GetTrophy()), 10)
// 	endTime := strconv.FormatInt(time.Now().Unix()+14*3600, 10)
// 	st := stWXScoreReq{
// 		AppId:     tencentAppId,
// 		GrantType: "client_credential",
// 		OpenId:    openid,
// 		Score:     data,
// 		Expires:   endTime,
// 	}

// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("WXUpLoadScore Marshal stWXScoreReq error: %v", err)
// 		return
// 	}
// 	buf := bytes.NewBuffer(body)

// 	timeCur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timeCur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/profile/wxscore/?timestamp=%s&appid=%s&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timeCur, tencentAppId, sig, openid)
// 	logger.Info("WXUpLoadScore tx url: %v", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("WXUpLoadScore http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("WX_uploadscore body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("WXUpLoadScore ioutil.ReadAll error: %v", err)
// 		return
// 	}
// 	rst := stWXScoreRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("WXUpLoadScore json.Unmarshal error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		logger.Error("WXUpLoadScore rst.Ret wrong", rst.Ret, errmsg)
// 		return
// 	}
// 	success = true
// 	errmsg = ""

// 	return
// }

// /*******************************************
// 游客相关接口
// ********************************************/
// const (
// 	sGuestLoginErrMsg = "Guest login failed!"
// )

// type stGuestCommonReq struct {
// 	GuestID     string `json:"guestid"`
// 	AccessToken string `json:"accessToken"`
// }

// func Guest_Auth(login *rpc.Login, IP string) (success bool, errmsg string) {
// 	success = false
// 	errmsg = sGuestLoginErrMsg

// 	//配置表
// 	//tencentAppId, tencentAppKey := common.GetWXAppInfo()
// 	tencentAppIdi, tencentAppKey := common.GetQQAppInfo(false, login.GetChannelid())
// 	tencentAppId := strconv.FormatInt(int64(tencentAppIdi), 10)
// 	tencentAppId = "G_" + tencentAppId
// 	sUrlBase := common.GetQQLoginUrl()

// 	//当前时间
// 	st := stGuestCommonReq{
// 		GuestID:     login.GetOpenid(),
// 		AccessToken: login.GetOpenkey(),
// 	}
// 	body, err := json.Marshal(st)
// 	if err != nil {
// 		logger.Error("Guest_Auth Marshal stMobileQQLogin error: %v", err)
// 		return
// 	}

// 	logger.Info("Guest_Auth client info: ", string(body))
// 	buf := bytes.NewBuffer(body)

// 	timecur := strconv.FormatInt(time.Now().Unix(), 10)
// 	h := md5.New()
// 	io.WriteString(h, tencentAppKey)
// 	io.WriteString(h, timecur)
// 	sig := fmt.Sprintf("%x", h.Sum(nil))
// 	fullurl := fmt.Sprintf("%s/auth/guest_check_token/?timestamp=%s&appid=%s&sig=%s&openid=%s&encode=1",
// 		sUrlBase, timecur, tencentAppId, sig, login.GetOpenid())
// 	logger.Info("Guest_Auth tx url: ", fullurl)

// 	client := &http.Client{
// 		Transport: createTransport(),
// 	}
// 	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
// 	if err != nil {
// 		logger.Error("Guest_Auth http.Post error: %v", err)
// 		return
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("Guest_Auth body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("Guest_Auth ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	rst := stMobileQQLoginRet{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("Guest_Auth ioutil.ReadAll error: %v", err)
// 		return
// 	}

// 	if rst.Ret != 0 {
// 		errmsg = rst.Msg
// 		return
// 	}

// 	success = true
// 	errmsg = ""

// 	return
// }
