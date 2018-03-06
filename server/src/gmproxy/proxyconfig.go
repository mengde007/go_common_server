package gmproxy

import (
	"jscfg"
	"logger"
	"os"
	"path"
	"strconv"
	"time"
	"timer"
)

var mapGmServerAddressCfg map[string]string
var reloadTick *timer.Timer

func loadGmServerAddressConfig() error {
	mapTemp := make(map[string]string)
	cfgpath, _ := os.Getwd()
	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"gmproxy.json"), &mapTemp); err != nil {
		logger.Error("loadGmServerAddressCfg failed: ", err)
		return err
	}
	mapGmServerAddressCfg = mapTemp
	return nil
}

func registerLoadGmServerAddressCfg() {
	reloadTick := timer.NewTimer(time.Minute)
	reloadTick.Start(func() {
		if err := loadGmServerAddressConfig(); err != nil {
			return
		}
	})
}

func getGmServerAddressCfg(id int) (string, bool) {
	idStr := strconv.Itoa(id)
	if address, ok := mapGmServerAddressCfg[idStr]; ok {
		return address, true
	}
	return "", false
}

// class struct 定义
type StHeadParse struct {
	Head *StHead `json:"head"`
}

type StHeadParseNew struct {
	Head *StHeadNew `json:"head"`
}

type StBodyParse struct {
	Body *StReq_Body `json:"body"`
}

type StHead struct {
	PacketLen    int    /* 包长 */
	Cmdid        int    /* 命令ID */
	Seqid        int64  /* 流水号 */
	ServiceName  string /* 服务名 */
	SendTime     int    /* 发送时间YYYYMMDD对应的整数 */
	Version      int    /* 版本号 */
	Authenticate string /* 加密串 */
	Result       int    /* 错误码,返回码类型：0：处理成功，需要解开包体获得详细信息,1：处理成功，但包体返回为空，不需要处理包体（eg：查询用户角色，用户角色不存在等），-1: 网络通信异常,-2：超时,-3：数据库操作异常,-4：API返回异常,-5：服务器忙,-6：其他错误,小于-100 ：用户自定义错误，需要填写szRetErrMsg */
	RetErrMsg    string /* 错误信息 */
}

type StHeadNew struct {
	Serverid string `json:"serverid"` //服务器列表
	Commid   int    `json:"commid"`   //命令Id
}

type StReqNew struct {
	Head *StHeadNew  `json:"head"`
	Body *StReq_Body `json:"body"`
}

type StReq struct {
	Head *StHead     `json:"head"`
	Body *StReq_Body `json:"body"`
}

type StReq_Body struct {
	Serverid string /* 小区 */
}

type StRst_Error struct {
	Ec   int           `json:"ec"`
	Em   string        `json:"em"`
	Data []*StReq_Body `json:"data"`
}

//错误
const (
	err_code_ok         = 0  //0：处理成功，需要解开包体获得详细信息
	err_code_ok_nothing = 1  //1：处理成功，但包体返回为空，不需要处理包体（eg：查询用户角色，用户角色不存在等）
	err_code_network    = -1 //-1: 网络通信异常
	err_code_timeout    = -2 //-2：超时
	err_code_db         = -3 //-3：数据库操作异常
	err_code_api        = -4 //-4：API返回异常
	err_code_busy       = -5 //-5：服务器忙
	err_code_other      = -6 //-6：其他错误
	//小于-100 ：用户自定义错误，需要填写szRetErrMsg
	err_code_wrong_param = -101 //参数错误
	err_code_inner_error = -102 //内部错误
	err_code_panic       = -103 //服务器崩溃
)

var mapErrStrings = map[int]string{
	err_code_wrong_param: "param error",
	err_code_inner_error: "inner error",
	err_code_panic:       "serious error",
}

/////////////////////////////////////
//玩家名字查询

//返回值
type StRsp_QueryName_Body_Role struct {
	RoleName  string /* 角色名 */
	OpenId    string /* openid */
	Partition int    // 小区id
}

type StRsp_QueryName_Body struct {
	UsrList_count int                          /* 角色信息列表的最大数量 */
	UsrList       []*StRsp_QueryName_Body_Role /* 角色信息列表 */
}

type StRsp_QueryName struct {
	Head *StHead               `json:"head"`
	Body *StRsp_QueryName_Body `json:"body"`
}

/////////////////////////////////////
