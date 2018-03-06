package connector

// import (
// 	"common"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"logger"
// 	"net/http"
// 	"rpc"
// )

// const (
// 	sMMLoginErrMsg     = "login failed!"
// 	sCreateClanErrMsg  = "create clan failed!"
// 	sJoinClanErrMsg    = "join clan failed!"
// 	sKickOutClanErrMsg = "kick out clan failed!"
// 	sDisBandClanErrMsg = "disband clan failed!"
// 	sQuitClanErrMsg    = "quit clan failed!"
// 	sChangeOwnerErrMsg = "change owner failed!"
// )

// const (
// 	sLoginUrl       = "https://game-api.immomo.com/game/2/server/app/check"
// 	sCreateClanUrl  = "https://game-api.immomo.com/game/2/server/group2/create"
// 	sJoinClanUrl    = "https://game-api.immomo.com/game/2/server/group2/join"
// 	sKickOutClanUrl = "https://game-api.immomo.com/game/2/server/group2/kickOut"
// 	sDisBandClanUrl = "https://game-api.immomo.com/game/2/server/group2/disband"
// 	sQuitClanUrl    = "https://game-api.immomo.com/game/2/server/group2/leave"
// 	sChangeOwnerUrl = "https://game-api.immomo.com/game/2/server/group2/changeOwner"
// )

// //超时连接
// func createHttpsTransport() *http.Transport {
// 	return common.CreateHttpsTransport()
// }

// type stMoMoLoginResult struct {
// 	Ec   int             `json:"ec"`
// 	Em   string          `json:"em"`
// 	Time uint32          `json:"timesec"`
// 	Data *stLoginRstData `json:"data"`
// }

// type stLoginRstData struct {
// 	Name   string `json:"name"`
// 	UserId string `json:"userid"`
// 	Vip    int    `json:"vip"`
// 	//Age     int64    `json:"age"`
// 	Sex     string   `json:"sex"`
// 	Photo   []string `json:"photo"`
// 	IsGuest int      `json:"is_guest"`
// }

// type stMoMoGuestLoginResult struct {
// 	Ec   int                  `json:"ec"`
// 	Em   string               `json:"em"`
// 	Time uint32               `json:"timesec"`
// 	Data *stGuestLoginRstData `json:"data"`
// }

// type stGuestLoginRstData struct {
// 	Name   string `json:"name"`
// 	UserId string `json:"userid"`
// 	Vip    int    `json:"vip"`
// 	//Age     string   `json:"age"`
// 	Sex     string   `json:"sex"`
// 	Photo   []string `json:"photo"`
// 	IsGuest int      `json:"is_guest"`
// }

// /// 登录 ///
// func MoMoLogin(login *rpc.Login) (bool, string, *stMoMoLoginResult) {
// 	appid, app_secret := common.GetMMAppInfo()
// 	vtoken := login.GetVToken()
// 	userid := login.GetUserId()

// 	fullurl := fmt.Sprintf("%s?appid=%s&app_secret=%s&vtoken=%s&userid=%s&encode=1",
// 		sLoginUrl, appid, app_secret, vtoken, userid)
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}
// 	logger.Info("fullurl:", fullurl)
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("MoMoLogin http.Get error: %v", err)
// 		return false, sMMLoginErrMsg, nil
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MoMoLogin body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MoMoLogin ioutil.ReadAll error: %v", err)
// 		return false, sMMLoginErrMsg, nil
// 	}

// 	rst := stMoMoLoginResult{}
// 	//if gl == rpc.GameLocation_Momo_Guest {
// 	//	rst = stMoMoGuestLoginResult{}
// 	//}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MoMoLogin ioutil.ReadAll error: %v", err)
// 		return false, sMMLoginErrMsg, nil
// 	}

// 	if rst.Ec != 0 {
// 		logger.Error("rst.Ec em", rst.Ec, rst.Em)
// 		return false, rst.Em, nil
// 	}
// 	logger.Info("MoMoLogin success!!!")
// 	return true, "", &rst
// }

// ///////////////////////// 创建群组 ////////////////////////
// //{
// //"ec": 0,
// //"em": "success",
// //"timesec": 1432625639,
// //"data": {
// //"gid": "33929269",
// //"name": "SDK测试游戏公会群",
// //"photos": [
// //"1A402A77-08DA-9367-51B1-AC75E366FFA2"
// //],
// //"sign": "欢迎加入陌陌争霸公会群",
// //"geoloc": {
// //"lat": 39.997772216797,
// //"lng": 116.4817276001
// //},
// //"create_time": 1432625639,
// //"owner": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
// //"games": [],
// //"game_union": {
// //"appid": "mm_sdk_test_3kKsqwvk",
// //"unionid": "union1"
// //},
// //"members": [{
// //"userid": " aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
// //"name": "秣马儿",
// //"avatar": "71ECDC41-723B-B44F-0336-84908D9C2E3E"
// //}],
// //"sid": "1f94533b82a3a640",
// //"sname": "保利万和电影院(卜蜂莲花望京宝星店)"
// //}
// //}
// type stCreateClanRst struct {
// 	Ec   int         `json:"ec"`
// 	Em   string      `json:"em"`
// 	Time uint32      `json:"timesec"`
// 	Data *stClanData `json:"data"`
// }

// type stClanData struct {
// 	Gid        string    `json:"gid"`
// 	Name       string    `json:"name"`
// 	Photos     []string  `json:"photos"`
// 	Sign       string    `json:"sign"`
// 	GeoLoc     *stGeoloc `json:"geoloc"`
// 	CreateTime uint32    `json:"create_time"`
// 	Owner      string    `json:"owner"`
// 	//Games []string `json:"games"`
// 	GameUnion *stGameUnion    `json:"game_union"`
// 	Members   []*stMemberInfo `json:"members"`
// 	Sid       string          `json:"sid"`
// 	Sname     string          `json:"sname"`
// }

// type stGeoloc struct {
// 	Lat float64 `json:"lat"`
// 	Lng float64 `json:"lng"`
// }

// type stGameUnion struct {
// 	AppId   string `json:"appid"`
// 	UnionId string `json:"unionid"`
// }

// type stMemberInfo struct {
// 	UserId string `json:"userid"`
// 	Name   string `json:"name"`
// 	Avatar string `json:"avatar"`
// }

// //创建
// //参数名	类型	必选	说明
// //appid	string	Y	应用id
// //app_secret	string	Y	应用密码
// //union_id	string	Y	公会标识(CP自定义公会唯一标识)
// //owner	string	Y	群主（正式陌陌用户userid，游客不可以）
// func MoMoCreateClan(clanUid, userUid string) (bool, string, *stCreateClanRst) {
// 	appid, app_secret := common.GetMMAppInfo()

// 	fullurl := fmt.Sprintf("%s?appid=%s&app_secret=%s&union_id=%s&owner=%s&encode=1",
// 		sCreateClanUrl, appid, app_secret, clanUid, userUid)
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}
// 	logger.Info("fullurl:", fullurl)
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("MoMoCreateClan http.Get error: %v", err)
// 		return false, sCreateClanErrMsg, nil
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MoMoCreateClan body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MoMoCreateClan ioutil.ReadAll error: %v", err)
// 		return false, sCreateClanErrMsg, nil
// 	}

// 	rst := stCreateClanRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MoMoCreateClan ioutil.ReadAll error: %v", err)
// 		return false, sCreateClanErrMsg, nil
// 	}

// 	if rst.Ec != 0 {
// 		logger.Error("rst.Ec em", rst.Ec, rst.Em)
// 		return false, rst.Em, nil
// 	}
// 	logger.Info("MoMoCreateClan success!!!")
// 	return true, "", &rst
// }

// /////// 群组相关操作 加入 踢出 等等 /////////
// //{
// //"ec": 0,
// //"em": "success",
// //"timesec": 1432625639,
// //"data": []
// //}
// type stOperateRst struct {
// 	Ec   int    `json:"ec"`
// 	Em   string `json:"em"`
// 	Time string `json:"timesec"`
// }

// //加入
// //参数名	类型	必选	说明
// //appid	string	Y	应用id
// //app_secret	string	Y	应用密码
// //union_id	string	Y	公会标识
// //userid	string	Y	用户标识
// func MoMoJoinClan(clanUid, userUid string) (bool, string, *stOperateRst) {
// 	appid, app_secret := common.GetMMAppInfo()

// 	fullurl := fmt.Sprintf("%s?appid=%s&app_secret=%s&union_id=%s&userid=%s&encode=1",
// 		sJoinClanUrl, appid, app_secret, clanUid, userUid)
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}
// 	logger.Info("fullurl:", fullurl)
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("MoMoJoinClan http.Get error: %v", err)
// 		return false, sJoinClanErrMsg, nil
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MoMoJoinClan body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MoMoJoinClan ioutil.ReadAll error: %v", err)
// 		return false, sJoinClanErrMsg, nil
// 	}

// 	rst := stOperateRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MoMoJoinClan ioutil.ReadAll error: %v", err)
// 		return false, sJoinClanErrMsg, nil
// 	}

// 	if rst.Ec != 0 {
// 		logger.Error("rst.Ec em", rst.Ec, rst.Em)
// 		return false, rst.Em, nil
// 	}
// 	logger.Info("MoMoJoinClan success!!!")
// 	return true, "", &rst
// }

// //踢出
// //参数名	类型	必选	说明
// //appid	string	Y	应用id
// //app_secret	string	Y	应用密码
// //union_id	string	Y	公会标识
// //operator	string	Y	操作人，用户标识
// //target	string	Y	被T的人，用户标识
// func MoMoKickOutClan(clanUid, userUid, targetId string) (bool, string, *stOperateRst) {
// 	appid, app_secret := common.GetMMAppInfo()

// 	fullurl := fmt.Sprintf("%s?appid=%s&app_secret=%s&union_id=%s&operator=%s&target=%s&encode=1",
// 		sKickOutClanUrl, appid, app_secret, clanUid, userUid, targetId)
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}
// 	logger.Info("fullurl:", fullurl)
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("MoMoKickOutClan http.Get error: %v", err)
// 		return false, sKickOutClanErrMsg, nil
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MoMoKickOutClan body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MoMoKickOutClan ioutil.ReadAll error: %v", err)
// 		return false, sKickOutClanErrMsg, nil
// 	}

// 	rst := stOperateRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MoMoKickOutClan ioutil.ReadAll error: %v", err)
// 		return false, sKickOutClanErrMsg, nil
// 	}

// 	if rst.Ec != 0 {
// 		logger.Error("rst.Ec em", rst.Ec, rst.Em)
// 		return false, rst.Em, nil
// 	}
// 	logger.Info("MoMoKickOutClan success!!!")
// 	return true, "", &rst
// }

// //解散
// //参数名	类型	必选	说明
// //appid	string	Y	应用id
// //app_secret	string	Y	应用密码
// //union_id	string	Y	公会标识
// //operator	string	Y	操作人，用户标识
// func MoMoDisbandClan(clanUid, userUid string) (bool, string, *stOperateRst) {
// 	appid, app_secret := common.GetMMAppInfo()

// 	fullurl := fmt.Sprintf("%s?appid=%s&app_secret=%s&union_id=%s&operator=%s&encode=1",
// 		sDisBandClanUrl, appid, app_secret, clanUid, userUid)
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}
// 	logger.Info("fullurl:", fullurl)
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("MoMoDisbandClan http.Get error: %v", err)
// 		return false, sDisBandClanErrMsg, nil
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MoMoDisbandClan body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MoMoDisbandClan ioutil.ReadAll error: %v", err)
// 		return false, sDisBandClanErrMsg, nil
// 	}

// 	rst := stOperateRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MoMoDisbandClan ioutil.ReadAll error: %v", err)
// 		return false, sDisBandClanErrMsg, nil
// 	}

// 	if rst.Ec != 0 {
// 		logger.Error("rst.Ec em", rst.Ec, rst.Em)
// 		return false, rst.Em, nil
// 	}
// 	logger.Info("MoMoDisbandClan success!!!")
// 	return true, "", &rst
// }

// //退出
// //参数名	类型	必选	说明
// //appid	string	Y	应用id
// //app_secret	string	Y	应用密码
// //union_id	string	Y	公会标识
// //userid	string	Y	用户标识
// func MoMoQuitClan(clanUid, userUid string) (bool, string, *stOperateRst) {
// 	appid, app_secret := common.GetMMAppInfo()

// 	fullurl := fmt.Sprintf("%s?appid=%s&app_secret=%s&union_id=%s&userid=%s&encode=1",
// 		sQuitClanUrl, appid, app_secret, clanUid, userUid)
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}
// 	logger.Info("fullurl:", fullurl)
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("MoMoQuitClan http.Get error: %v", err)
// 		return false, sQuitClanErrMsg, nil
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MoMoQuitClan body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MoMoQuitClan ioutil.ReadAll error: %v", err)
// 		return false, sQuitClanErrMsg, nil
// 	}

// 	rst := stOperateRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MoMoQuitClan ioutil.ReadAll error: %v", err)
// 		return false, sQuitClanErrMsg, nil
// 	}

// 	if rst.Ec != 0 {
// 		logger.Error("rst.Ec em", rst.Ec, rst.Em)
// 		return false, rst.Em, nil
// 	}
// 	logger.Info("MoMoQuitClan success!!!")
// 	return true, "", &rst
// }

// //群主换人
// //参数名	类型	必选	说明
// //appid	string	Y	应用id
// //app_secret	string	Y	应用密码
// //union_id	string	Y	公会标识
// //operator	string	Y	操作人，用户标识
// //new_owner	string	Y	新群主，用户标识
// func MoMoChangeOwner(clanUid, operator, newOwner string) (bool, string, *stOperateRst) {
// 	appid, app_secret := common.GetMMAppInfo()

// 	fullurl := fmt.Sprintf("%s?appid=%s&app_secret=%s&union_id=%s&operator=%s&new_owner=%s&encode=1",
// 		sChangeOwnerUrl, appid, app_secret, clanUid, operator, newOwner)
// 	client := &http.Client{
// 		Transport: createHttpsTransport(),
// 	}
// 	logger.Info("fullurl:", fullurl)
// 	res, err := client.Get(fullurl)
// 	if err != nil {
// 		logger.Error("MoMoChangeOwner http.Get error: %v", err)
// 		return false, sChangeOwnerErrMsg, nil
// 	}

// 	b, err := ioutil.ReadAll(res.Body)
// 	logger.Info("MoMoChangeOwner body info:%s", string(b))
// 	res.Body.Close()
// 	if err != nil {
// 		logger.Error("MoMoChangeOwner ioutil.ReadAll error: %v", err)
// 		return false, sChangeOwnerErrMsg, nil
// 	}

// 	rst := stOperateRst{}
// 	if err := json.Unmarshal(b, &rst); err != nil {
// 		logger.Error("MoMoChangeOwner ioutil.ReadAll error: %v", err)
// 		return false, sChangeOwnerErrMsg, nil
// 	}

// 	if rst.Ec != 0 {
// 		logger.Error("rst.Ec em", rst.Ec, rst.Em)
// 		return false, rst.Em, nil
// 	}
// 	logger.Info("MoMoChangeOwner success!!!")
// 	return true, "", &rst
// }
