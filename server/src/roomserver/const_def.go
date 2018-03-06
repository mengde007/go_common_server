package roomserver

//进入房间错误码定义
const (
	ECRNone              = iota
	ECRLessCoin          //金币太低
	ECRReachUpLimit      //金币超过上限了
	ECRPwdError          //密码错误
	ECRNotExistRoom      //不存在房间
	ECRConvertRoomFailed //转换房间失败
	ECRFull              //房间满了
)

//创建房间的错误码定义
const (
	ECCRNone                   = iota
	ECCRNameLength             //房间的名字长度错误
	ECCRPwdLength              //密码长度错误
	ECCRDifen                  //底注不在指定范围内
	ECCRMatchTimes             //比赛次数错误
	ECCRCreateFrequently       //创建房间太频繁了
	ECCRConfigError            //配置表错误
	ECCRConvertRoomFailed      //转换房间失败
	ECCRNoneID                 //没有可用的ID了
	ECCRCreateRoomMinCoinLimit //没有达到创建房间的最小金币限制
	ECCRUnknowError            //未知错误
	ECCRAlreadyInRoom          //已经在房间了，不能再创建房间
	ECCRGreaterSelfCoin        //房间的进入金币限制不应该大于自己的金币
	ECCRMultipleLimit          //不在倍数限制范围内

	ECCRNotEnoughRoomCard = 1000 //没有足够的房卡

)

//查找房间的错误码
const (
	EFRNone                  = iota
	EFRGenerateRoomInfoError //产生roomInfo信息错误
	EFRRequireParamError     //请求参数错误
	EFRNotFind               //没有找到指定的房间
)

//结算货币类型
const (
	CTCredits = 1 //游戏积分
	CTCoin    = 2 //游戏金币
)

//解散房间
const (
	JSClaimer       = 1 //"申请者"
	JSWatingDispose = 2 // "等待处理"
	JSAgree         = 3 // "同意"
	JSRefuse        = 4 // "拒绝"
)
