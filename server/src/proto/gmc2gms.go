package proto

type GmPlayerSend struct {
	Msg []byte
}

type GmUpdateOpenId2Name struct {
	OpenId   string
	Platform uint32
	NameLast string
	Name     string
}

type GmUpdateOpenId2NameRst struct {
}

//指定用户邮件
//附件格式：type:num,type:num
//Users格式：userid,userid
type GmSendMail struct {
	Title   string
	Content string
	Attach  string
	Users   string
	SignId  uint64
	Channel uint32
}

//返回格式（0失败1成功）：userid:0,userid:1
type GmSendMailResult struct {
	SignId  uint64
	Success string
}

//所有用户邮件
type GmSendAllMail struct {
	Title        string
	Content      string
	Attach       string
	SignId       uint64
	Channel      uint32
	ContinueTime uint32
}

//所有用户邮件
type GmSendAllMailResult struct {
	SignId  uint64
	Success bool
}

//通知类型
const (
	GmNoticeType_Login = iota
	GmNoticeType_Marquee
)

// 通知
type GmSendNotice struct {
	SignId  uint64 //通知ID
	Content string //内容
	Type    int64  //通知类型
	Channel uint32 //渠道号
}

// 通知返回
type GmSendNoticeResult struct {
	SignId  uint64
	Success bool
}

//锁定玩家
type GmLockPlayer struct {
	Uid string
}

//锁定玩家返回
type GmLockPlayerResult struct {
	Success  bool
	OldValue uint64
}

//解锁玩家
type GmUnLockPlayer struct {
	Uid string
}

//解锁结果
type GmUnLockPlayerResult struct {
	Success bool
}

//玩家信息
type GmPlayerInfo struct {
	Uid         string //用户UID
	Name        string //用户名称
	Level       uint32 //用户等级
	Clan        string //所属联盟
	Diamonds    uint32 //宝石数
	Food        uint32 //食物粮草
	Gold        uint32 //银币？
	Wuhun       uint32 //武魂
	Trophy      uint32 //令牌
	DrillTimes  uint32 //演习次数
	CenterLevel uint32 //主营等级
	LastLogin   int64  //最后登录
	LoginStats  string //登录状态
	Updated     string //变更时间
	Channel     uint32 // 渠道号
}

//在线人数查询
type GmGetOnlineNum struct {
	Channel uint32
}

type GmGetOnlineNumResult struct {
	Value uint32
}

type GmCheckActivityConfig struct {
	ActivityId int
	Open       bool
	Test       bool
}

type GmCheckActivityConfigResult struct {
	Result    bool
	BeginTime uint32
	EndTime   uint32
	Plat      int
	Type      int
}

//活动配置表
type GetActivityConfig struct {
}

type GetActivityConfigRst struct {
	Value []byte
}
