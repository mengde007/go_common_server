package gmserver

//平台类型
const (
	PlatId_Ios     = 0
	PlatId_Android = 1
	PlatId_Both    = 2
)

//qq wx 类型
const (
	ServerType_QQ uint32 = 1
	ServerType_WX uint32 = 2
)

//公告类型
const (
	Notice_Marquee = 0
	Notice_Login   = 1
	Notice_Offline = 2
)

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
	Commid   int    `json:"Commid"`   //命令Id
}

type StHeadParseNew struct {
	Head *StHeadNew `json:"head"`
}

type StHeadParse struct {
	Head *StHead `json:"head"`
}

//角色信息主结构
type StRoleInfo struct {
	Roleid        string         `json:"roleid"`          //⻆角⾊色ID
	Momoid        string         `json:"momoid"`          //陌陌ID
	Nickname      string         `json:"nickname"`        //昵称
	Level         int            `json:"level"`           //等级
	IsOnline      int            `json:"is_online"`       //是否在线 0离线 1在 线
	Channelid     string         `json:"channelid"`       //渠道ID
	Serverid      string         `json:"serverid"`        //服ID
	BanStatus     int            `json:"ban_status"`      //封禁状态
	CreateTime    int            `json:"create_time"`     //创⻆角时间戳
	LastLoginTime int            `json:"last_login_time"` //最后登录时间戳
	OnlineDays    int            `json:"online_days"`     //连续登录的天数
	LastPayTime   int            `json:"last_pay_time"`   //最后充值时间戳
	Gold          string         `json:"gold"`            //当前值/最大值
	Food          string         `json:"food"`            //当前值/最大值
	Oil           string         `json:"oil"`             //当前值/最大值
	CivGold       int            `json:"civ_gold"`        //⽂文明币
	CampLevel     int            `json:"camp_level"`      //⼤大本营等级
	Gem           int            `json:"gem"`             //宝⽯石
	Strength      int            `json:"strength"`        //体⼒
	HolyWater     int            `json:"holy_water"`      //圣⽔水
	Prestige      int            `json:"prestige"`        //声望
	HeroInfo      []*StHeroInfo  `json:"hero_info"`       //英雄信息
	AccountInfo   *StAccountInfo `json:"account_info"`    //同陌陌号其他⻆角⾊色
}

type StSimpleKV struct {
	Key   string
	Value string
}

type StHeroInfo struct {
	HeroId     string        `json:"heroid"`      //￼英雄ID
	HeroName   string        `json:"hero_name"`   //￼英雄名称
	Level      int           `json:"level"`       //￼等级
	StarLevel  int           `json:"star_level"`  //￼星级
	Commander  int           `json:"commander"`   //￼统帅
	SkillLevel []*StSimpleKV `json:"skill_level"` //￼技能等级
}

type StAccountInfo struct {
	Roleid        string `json:"roleid"`          //￼⻆角⾊色ID
	Nickname      string `json:"nickname"`        //￼昵称
	Fight         string `json:"fight"`           //战力
	LastLoginTime int    `json:"last_login_time"` //￼最后登录时间戳
	Serverid      string `json:"serverid"`
}

//查询角色信息
type StReq_RoleInfo struct {
	Head *StHeadNew           `json:"head"`
	Body *StReq_RoleInfo_Body `json:"body"`
}
type StReq_RoleInfo_Body struct {
	SearchType string `json:"user_type"` // 查询⽤用户类型: momoid/roleid/ nickname => 陌陌 ID/⻆角⾊色ID/⻆角⾊色昵称
	SearChId   string `json:"searchid"`
}

//-----------------new

//修改角色信息
type StReq_Modify_RoleInfo struct {
	Head *StHeadNew                  `json:"head"`
	Body *StReq_Modify_RoleInfo_Body `json:"body"`
}
type StReq_Modify_RoleInfo_Body struct {
	Roleid string `json:"roleid"` //角色Id
	Type   string `json:"type"`   //修改信息类型
	Number int    `json:"number"` //正值表示增加，负值表示减少
}

//统计信息
type StReq_Accounting_info struct {
	Head *StHeadNew             `json:"head"`
	Body *StReq_Accounting_Body `json:"body"`
}
type StReq_Accounting_Body struct {
	Type string `json:"type"` //修改信息类型
}

//统计信息
type StRst_Accounting struct {
	Ec   int                    `json:"ec"`
	Em   string                 `json:"em"`
	Data *StRst_Accounting_Body `json:"data"`
}
type StRst_Accounting_Body struct {
	Online int `json:"online"` //在线人数
}

//创建轮播
type StReq_Notice struct {
	Head *StHeadNew         `json:"head"`
	Body *StReq_Notice_Body `json:"body"`
}
type StReq_Notice_Body struct {
	StartTime int    `json:"start_time"` //轮播开始时间
	EndTime   int    `json:"end_time"`   //轮播结束时间
	Channel   string `json:"channel"`    //渠道以逗号隔开
	Interval  int    `json:"interval"`   //时间间隔,不小于5秒
	Content   string `json:"content"`    //轮播内容,小于30汉字
	Priority  int    `json:"priority"`   //优先级
}

//删除轮播
type StReq_Delete_Notice struct {
	Head *StHeadNew         `json:"head"`
	Body *StRst_Notice_Body `json:"body"`
}

//创建轮播的返回
type StRst_Notice_Body struct {
	NoticeId int `json:"noticeid"`
}

//---------------new end
//发送邮件
type StReq_Send_Mail struct {
	Head *StHeadNew            `json:"head"`
	Body *StReq_Send_Mail_Body `json:"body"`
}
type StReq_Send_Mail_Body struct {
	Title     string `json:"title"`      //邮件标题
	Content   string `json:"content"`    //邮件内容
	Expire    int    `json:"expire"`     //邮件过期时间，单位天
	UserType  string `json:"user_type"`  //陌陌ID/角色ID/昵称
	Ids       string `json:"ids"`        //ids用“,”分割
	Attach    string `json:"attach"`     //附件
	EmailType int    `json:"email_type"` //0全服筛选/1指定ID
}

//冻结帐号，禁言
type StReq_FreezeNew struct {
	Head *StHeadNew            `json:"head"`
	Body *StReq_Freeze_BodyNew `json:"body"`
}
type StReq_Freeze_BodyNew struct {
	Expire  int    `json:"expire"`   //封禁时长单位天
	BanedBy string `json:"baned_by"` //封禁方式(陌陌ID/角色ID/昵称
	BanType string `json:"ban_type"` //封禁项login/chat
	Ids     string `json:"ids"`      //封禁id
	Reason  string `json:"reason"`   //服务器
}

//解冻帐号，或解除禁言
type StReq_Unfreeze struct {
	Head *StHeadNew           `json:"head"`
	Body *StReq_Unfreeze_Body `json:"body"`
}
type StReq_Unfreeze_Body struct {
	UnbanedBy string `json:"unbaned_by"` //封禁方式(陌陌ID/角色ID/昵称
	BanType   string `json:"ban_type"`   //封禁项login/chat
	Ids       string `json:"ids"`        //封禁id
}

type StRst_Notice struct {
	Ec   int                `json:"ec"`
	Em   string             `json:"em"`
	Data *StRst_Notice_Body `json:"data"`
}

/////////////////////////////////////old
//玩家名字查询
type StReq_QueryName_Body struct {
	PlatId   int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	RoleName string /* 角色名 */
}
type StReq_QueryName struct {
	Head *StHead               `json:"head"`
	Body *StReq_QueryName_Body `json:"body"`
}

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

type StRst_RoleInfo struct {
	Ec   int           `json:"ec"`
	Em   string        `json:"em"`
	Data []*StRoleInfo `json:"data"`
}

/////////////////////////////////////
//发送邮件
type StReq_SendMail_Body struct {
	PlatId      int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId      string /* openid */
	Every       int    /*是否全服邮件：（0否，1是，为0时openid一定不能为空）*/
	MailTitle   string /* 邮件标题 */
	MailContent string /* 邮件内容 */
	ItemId      int    /* 道具ID（0宝石，1金钱,2粮草,3紫金，4武将碎片，5丹药，6VIP等级，7体力，8纯文本  ） */
	ItemNum     int    /* 道具数量 */
	Level       int    /*道具等级*/
	ExpireTime  int    /* 有效期限（秒） */
	Source      int    /* 渠道号，由前端生成，不需要填写 */
	Serial      string /* 流水号，由前端生成，不需要填写 */
}

type StReq_SendMail struct {
	Head *StHead              `json:"head"`
	Body *StReq_SendMail_Body `json:"body"`
}

//返回值
type StRsp_SendMail_Body struct {
	Result int    /* 结果（0：成功， 1：失败） */
	RetMsg string /* 返回消息 */
}

type StRsp_SendMail struct {
	Head *StHead              `json:"head"`
	Body *StRsp_SendMail_Body `json:"body"`
}

/////////////////////////////////////
//封号
type StReq_Freeze_Body struct {
	PlatId  int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId  string /* openid */
	BanTime int    /* 封停时长：*秒，0 永久 */
	Reason  string /* 操作原因 */
	Source  int    /* 渠道号，由前端生成，不需要填写 */
	Serial  string /* 流水号，由前端生成，不需要填写 */
}

type StReq_Freeze struct {
	Head *StHead            `json:"head"`
	Body *StReq_Freeze_Body `json:"body"`
}

//返回值
type StRsp_Freeze_Body struct {
	Result int    /* 结果（0：成功， 1：失败） */
	RetMsg string /* 返回消息 */
}

type StRsp_Freeze struct {
	Head *StHead            `json:"head"`
	Body *StRsp_Freeze_Body `json:"body"`
}

/////////////////////////////////////
//公告发送（跑马灯）
type StReq_Marquee_Body struct {
	PlatId        int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	Priority      int    /* 优先级 */
	NoticeContent string /* 公告内容 */
	BeginLevel    int    /* 开始等级区间（0-99） */
	EndLevel      int    /* 结束等级区间（0-99） */
	Interval      int    /* 滚动时间间隔：*秒 */
	Times         int    /* 滚动次数 */
	BeginTime     int64  /* 开始时间 */
	EndTime       int64  /* 结束时间 */
	Source        int    /* 渠道号，由前端生成，不需要填写 */
	Serial        string /* 流水号，由前端生成，不需要填写 */
}

type StReq_Marquee struct {
	Head *StHead             `json:"head"`
	Body *StReq_Marquee_Body `json:"body"`
}

//返回值
type StRsp_Marquee_Body struct {
	Result   int    /* 结果（0：成功， 1：失败） */
	RetMsg   string /* 返回消息 */
	NoticeId int64  /* 公告ID */
}

type StRsp_Marquee struct {
	Head *StHead             `json:"head"`
	Body *StRsp_Marquee_Body `json:"body"`
}

/////////////////////////////////////
//公告发送（系统弹窗）
type StReq_PopWindow_Body struct {
	PlatId                int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	Priority              int    /* 优先级 */
	Type                  int    /* 内容类型：文本（0） */
	ActivityNoticeTitle   string /* 活动公告标题 */
	ActivityNoticeContent string /* 活动公告内容 */
	BeginTime             int64  /* 开始时间 */
	EndTime               int64  /* 结束时间 */
	Source                int    /* 渠道号，由前端生成，不需要填写 */
	Serial                string /* 流水号，由前端生成，不需要填写 */
}

type StReq_PopWindow struct {
	Head *StHead               `json:"head"`
	Body *StReq_PopWindow_Body `json:"body"`
}

//返回值
type StRsp_PopWindow_Body struct {
	Result int    /* 结果（0：成功， 1：失败） */
	RetMsg string /* 返回消息 */
}

type StRsp_PopWindow struct {
	Head *StHead               `json:"head"`
	Body *StRsp_PopWindow_Body `json:"body"`
}

/////////////////////////////////////
//公告查询
type StReq_NoticeQuery_Body struct {
	PlatId    int   /* 平台：IOS（0），安卓（1） ,Both（2） */
	BeginTime int64 /* 开始时间 */
	EndTime   int64 /* 结束时间 */
}

type StReq_NoticeQuery struct {
	Head *StHead                 `json:"head"`
	Body *StReq_NoticeQuery_Body `json:"body"`
}

//返回值
type StRsp_NoticeQuery_Body_List struct {
	Type          int    /* 公告类型 */
	NoticeId      int64  /* 公告ID */
	NoticeTitle   string /* 公告标题 */
	NoticeContent string /* 公告内容 */
	SendTime      int64  /* 发送时间 */
	Partition     int    // 小区id
}

type StRsp_NoticeQuery_Body struct {
	NoticeList_count int                            /* 公告信息列表的最大数量 */
	NoticeList       []*StRsp_NoticeQuery_Body_List /* 公告信息列表 */
}

type StRsp_NoticeQuery struct {
	Head *StHead                 `json:"head"`
	Body *StRsp_NoticeQuery_Body `json:"body"`
}

/////////////////////////////////////
//公告删除
type StReq_NoticeDel_Body struct {
	PlatId   int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	Type     int    /* 公告类型：跑马灯（0），活动（1），维护（2） */
	NoticeId int64  /* 公告ID */
	Reason   string /* 操作原因 */
	Source   int    /* 渠道号，由前端生成，不需要填写 */
	Serial   string /* 流水号，由前端生成，不需要填写 */
}

type StReq_NoticeDel struct {
	Head *StHead               `json:"head"`
	Body *StReq_NoticeDel_Body `json:"body"`
}

//返回值
type StRsp_NoticeDel_Body struct {
	Result int    /* 结果：0 成功，其它失败 */
	RetMsg string /* 返回消息 */
}

type StRsp_NoticeDel struct {
	Head *StHead               `json:"head"`
	Body *StRsp_NoticeDel_Body `json:"body"`
}

/////////////////////////////////////
//资源修改
type StReq_ResModify_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Type   int    /* 资源类型：（1粮草，2银两，3宝石，4紫金，5体力） */
	Value  int    /* 数量：+加-减 */
	Reason string /* 操作原因 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_ResModify struct {
	Head *StHead               `json:"head"`
	Body *StReq_ResModify_Body `json:"body"`
}

//返回值
type StRsp_ResModify_Body struct {
	Result int    /* 结果（0）成功 */
	RetMsg string /* 返回消息 */
	Value  int64  /* 修改后数量 */
}

type StRsp_ResModify struct {
	Head *StHead               `json:"head"`
	Body *StRsp_ResModify_Body `json:"body"`
}

/////////////////////////////////////
//主营等级修改
type StReq_CenterModify_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Value  int    /* 数量：+加-减 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_CenterModify struct {
	Head *StHead                  `json:"head"`
	Body *StReq_CenterModify_Body `json:"body"`
}

//返回值
type StRsp_CenterModify_Body struct {
	Result int    /* 结果（0）成功 */
	RetMsg string /* 返回消息 */
	Value  int    /* 修改后大本营等级 */
}

type StRsp_CenterModify struct {
	Head *StHead                  `json:"head"`
	Body *StRsp_CenterModify_Body `json:"body"`
}

/////////////////////////////////////
//英雄等级
type StReq_HeroLevel_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Level  int    /* 指定英雄等级 */
	Index  int    /* 英雄索引 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_HeroLevel struct {
	Head *StHead               `json:"head"`
	Body *StReq_HeroLevel_Body `json:"body"`
}

//返回值
type StRsp_HeroLevel_Body struct {
	Result int    /* 结果（0）成功 */
	RetMsg string /* 返回消息 */
	Value  int    /* 修改后英雄等级 */
}

type StRsp_HeroLevel struct {
	Head *StHead               `json:"head"`
	Body *StRsp_HeroLevel_Body `json:"body"`
}

/////////////////////////////////////
//删除英雄
type StReq_HeroDel_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Index  int    /* 英雄索引 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_HeroDel struct {
	Head *StHead             `json:"head"`
	Body *StReq_HeroDel_Body `json:"body"`
}

//返回值
type StRsp_HeroDel_Body struct {
	Result int    /* 结果（0）成功 */
	RetMsg string /* 返回消息 */
}

type StRsp_HeroDel struct {
	Head *StHead             `json:"head"`
	Body *StRsp_HeroDel_Body `json:"body"`
}

/////////////////////////////////////
//杯数修改
type StReq_TrophyModify_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Value  int    /* 数量：+加-减 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_TrophyModify struct {
	Head *StHead                  `json:"head"`
	Body *StReq_TrophyModify_Body `json:"body"`
}

//返回值
type StRsp_TrophyModify_Body struct {
	Result int    /* 结果（0）成功 */
	RetMsg string /* 返回消息 */
	Value  int64  /* 设置后数量 */
}

type StRsp_TrophyModify struct {
	Head *StHead                  `json:"head"`
	Body *StRsp_TrophyModify_Body `json:"body"`
}

/////////////////////////////////////
//查询基础信息
type StReq_BaseQuery_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
}

type StReq_BaseQuery struct {
	Head *StHead               `json:"head"`
	Body *StReq_BaseQuery_Body `json:"body"`
}

//返回值
type StRsp_BaseQuery_Body struct {
	RoleName     string /* 角色名 */
	OpenId       string /* openid */
	RegisterTime int64  /* 注册时间 */
	Jewel        int64  /* 宝石数 */
	Level        int    /* 等级 */
	CampLevel    int    /* 大本营等级 */
	Flag         int64  /* 旗帜数量 */
	MatchStep    int    /* 所处至尊联赛的杯段 */
	Vip          int    /* VIP等级 */
	Money        int64  /* 银两数 */
	Food         int64  /* 粮草数 */
	VioletGold   int64  /* 紫金数 */
}

type StRsp_BaseQuery struct {
	Head *StHead               `json:"head"`
	Body *StRsp_BaseQuery_Body `json:"body"`
}

/////////////////////////////////////
//查询建筑信息
type StReq_VillageQuery_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
}

type StReq_VillageQuery struct {
	Head *StHead                  `json:"head"`
	Body *StReq_VillageQuery_Body `json:"body"`
}

//返回值
type StRsp_VillageQuery_Body struct {
	MainArmy          int /* 主营等级 */
	Wall              int /* 城墙等级 */
	GoldMine          int /* 金矿等级 */
	Army              int /* 兵营等级 */
	OrdnanceInstitute int /* 兵工研究所等级 */
	Farm              int /* 农田等级 */
	Barn              int /* 粮仓等级 */
	CampGround        int /* 营地等级 */
	WorkerHouse       int /* 工人房等级 */
	ArrowTower        int /* 箭塔等级 */
	Cannon            int /* 火炮等级 */
	OperTower         int /* 术士塔等级 */
	SkyArrow          int /* 破空火箭等级 */
	StoneCar          int /* 投石车等级 */
	SpiritBanner      int /* 招魂幡等级 */
	Crossbow          int /* 诸葛连弩等级 */
	LeagueHall        int /* 联盟议事厅等级 */
	MyaxHouse         int /* 丹药房等级 */
	Bomb              int /* 炸弹等级 */
	FirearmCar        int /* 火器车等级 */
	PoisonWell        int /* 毒井等级 */
	GeneralStep       int /* 点将台等级 */
	Decoration        int /* 装饰物等级 */
	Barry             int /* 障碍物等级 */
	Lantern           int /* 孔明燃灯等级 */
	Fire              int /* 玄天火烛等级 */
	OfficerHouse      int /* 紫金官坊等级 */
	Justice           int /* 紫金储司等级 */
	FriendHouse       int /* 益友工坊等级 */
	ArmyStation       int /* 兵站等级 */
}

type StRsp_VillageQuery struct {
	Head *StHead                  `json:"head"`
	Body *StRsp_VillageQuery_Body `json:"body"`
}

/////////////////////////////////////
//查询排行榜
type StReq_RankQuery_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
}

type StReq_RankQuery struct {
	Head *StHead               `json:"head"`
	Body *StReq_RankQuery_Body `json:"body"`
}

//返回值
type StRsp_RankQuery_Body struct {
	Time   int64 /* 时间 */
	Flag   int64 /* 旗帜排行 */
	League int64 /* 联盟排行 */
}

type StRsp_RankQuery struct {
	Head *StHead               `json:"head"`
	Body *StRsp_RankQuery_Body `json:"body"`
}

/////////////////////////////////////
//查询英雄
type StReq_HeroQuery_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	PageNo int    /* 页码 每页最多50个武将 */
}

type StReq_HeroQuery struct {
	Head *StHead               `json:"head"`
	Body *StReq_HeroQuery_Body `json:"body"`
}

//返回值
type StRsp_HeroQuery_Body_Hero_Skill struct {
	SkillName string /* 技能名称 */
}

type StRsp_HeroQuery_Body_Hero struct {
	HeroId              int                                /* 英雄ID */
	HeroName            string                             /* 英雄名称 */
	HeroLevel           int                                /* 英雄等级 */
	HeroSkillList_count int                                /* 英雄技能名称信息列表的最大数量 */
	HeroSkillList       []*StRsp_HeroQuery_Body_Hero_Skill /* 英雄技能名称信息列表 */
	Hurt                int                                /* 伤害 */
	Blood               int                                /* 血量 */
}

type StRsp_HeroQuery_Body struct {
	HeroList_count int                          /* 英雄信息列表的最大数量  */
	HeroList       []*StRsp_HeroQuery_Body_Hero /* 英雄信息列表 */
	TotalPageNo    int                          /* 总页码数 */
}

type StRsp_HeroQuery struct {
	Head *StHead               `json:"head"`
	Body *StRsp_HeroQuery_Body `json:"body"`
}

/////////////////////////////////////
//查询兵种
type StReq_CharQuery_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
}

type StReq_CharQuery struct {
	Head *StHead               `json:"head"`
	Body *StReq_CharQuery_Body `json:"body"`
}

//返回值
type StRsp_CharQuery_Body_Char struct {
	Num         int    /* 兵种总量 */
	ArmId       int    /* 兵种ID */
	ArmName     string /* 兵种名称 */
	Level       int    /* 兵种等级 */
	UnbanTime   int64  /* 兵种解封时间 */
	UpgradeTime int64  /* 升级时间 */
}

type StRsp_CharQuery_Body struct {
	ArmList_count int                          /* 兵种信息列表的最大数量 */
	ArmList       []*StRsp_CharQuery_Body_Char /* 兵种信息列表 */
}

type StRsp_CharQuery struct {
	Head *StHead               `json:"head"`
	Body *StRsp_CharQuery_Body `json:"body"`
}

/////////////////////////////////////
//查询联盟
type StReq_ClanQuery_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
}

type StReq_ClanQuery struct {
	Head *StHead               `json:"head"`
	Body *StReq_ClanQuery_Body `json:"body"`
}

//返回值
type StRsp_ClanQuery_Body_Member struct {
	LeagueName       string /* 联盟名称 */
	Flag             int64  /* 旗帜数 */
	LeaderName       string /* 首领名称 */
	LeaderOpenId     string /* 首领openid */
	MemberList_count int    /* 成员信息列表的最大数量  */
	TotalNum         int    /* 总人数 */
	RoleName         string /* 成员名称 */
	OpenId           string /* openid */
	Level            int    /* 等级 */
}

type StRsp_ClanQuery_Body struct {
	LeagueList_count int                            /* 联盟信息列表的最大数量 */
	LeagueList       []*StRsp_ClanQuery_Body_Member /* 联盟信息列表 */
}

type StRsp_ClanQuery struct {
	Head *StHead               `json:"head"`
	Body *StRsp_ClanQuery_Body `json:"body"`
}

/////////////////////////////////////
//查询活动
type StReq_ActivityQuery_Body struct {
	PlatId int /* 平台：IOS（0），安卓（1） ,Both（2） */
}

type StReq_ActivityQuery struct {
	Head *StHead                   `json:"head"`
	Body *StReq_ActivityQuery_Body `json:"body"`
}

//返回值
type StRsp_ActivityQuery_Body_List struct {
	BeginTime  int64 /* 开始时间 */
	EndTime    int64 /* 结束时间 */
	PlatId     int   /* 平台 */
	ActivityId int   /* 活动ID */
	Type       int   /* 活动类型 */
	Status     int   /* 活动状态（进行中/已结束） */
}

type StRsp_ActivityQuery_Body struct {
	ActivityList_count int                              /* 活动信息列表的最大数量  */
	ActivityList       []*StRsp_ActivityQuery_Body_List /* 活动信息列表 */
}

type StRsp_ActivityQuery struct {
	Head *StHead                   `json:"head"`
	Body *StRsp_ActivityQuery_Body `json:"body"`
}

/////////////////////////////////////
//关闭活动
type StReq_ActivityClose_Body struct {
	ActivityId int    /* 活动ID */
	Source     int    /* 渠道号，由前端生成，不需要填写 */
	Serial     string /* 流水号，由前端生成，不需要填写 */
}

type StReq_ActivityClose struct {
	Head *StHead                   `json:"head"`
	Body *StReq_ActivityClose_Body `json:"body"`
}

//返回值
type StRsp_ActivityClose_Body struct {
	BeginTime int64 /* 开始时间 */
	EndTime   int64 /* 结束时间 */
	PlatId    int   /* 平台 */
	Type      int   /* 活动类型 */
	Status    int   /* 活动状态（进行中/已结束），点击结束后返回状态（已关闭） */
}

type StRsp_ActivityClose struct {
	Head *StHead                   `json:"head"`
	Body *StRsp_ActivityClose_Body `json:"body"`
}

/////////////////////////////////////
//重新加载活动
type StReq_ActivityReload_Body struct {
	ActivityId int    /* 活动ID */
	Source     int    /* 渠道号，由前端生成，不需要填写 */
	Serial     string /* 流水号，由前端生成，不需要填写 */
}

type StReq_ActivityReload struct {
	Head *StHead                    `json:"head"`
	Body *StReq_ActivityReload_Body `json:"body"`
}

//返回值
type StRsp_ActivityReload_Body struct {
	BeginTime  int64 /* 开始时间 */
	EndTime    int64 /* 结束时间 */
	PlatId     int   /* 平台 */
	ActivityId int   /* 活动ID */
	Type       int   /* 活动类型 */
	Status     int   /* 活动状态（进行中/已结束），点击结束后返回状态（已关闭） */
}

type StRsp_ActivityReload struct {
	Head *StHead                    `json:"head"`
	Body *StRsp_ActivityReload_Body `json:"body"`
}

/////////////////////////////////////
//设置服务器最大人数
type StReq_OnlineNum_Body struct {
	Partition int    /* 小区 */
	MaxOnline int    /* 人数上限 */
	Source    int    /* 渠道号，由前端生成，不需要填写 */
	Serial    string /* 流水号，由前端生成，不需要填写 */
}

type StReq_OnlineNum struct {
	Head *StHead               `json:"head"`
	Body *StReq_OnlineNum_Body `json:"body"`
}

//返回值
type StRsp_OnlineNum_Body struct {
	Result    int    /* 结果 */
	RetMsg    string /* 返回消息 */
	CurOnline int    /* 当前人数 */
	MaxOnline int    /* 人数上限 */
}

type StRsp_OnlineNum struct {
	Head *StHead               `json:"head"`
	Body *StRsp_OnlineNum_Body `json:"body"`
}

/////////////////////////////////////////////////
// 禁止指定玩法
type StReq_BanPlay_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Type   int    /* 禁止类型（0征战天下、1玩家对战、2好友掠夺、3闯关，99 全选） */
	Time   int64  /* 禁止玩法时长（秒） */
	Reason string /* 提示内容 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_BanPlay struct {
	Head *StHead             `json:"head"`
	Body *StReq_BanPlay_Body `json:"body"`
}

// 返回值
type StRsp_BanPlay_Body struct {
	Result int
	RetMsg string
}

type StRsp_BanPlay struct {
	Head *StHead             `json:"head"`
	Body *StRsp_BanPlay_Body `json:"body"`
}

/////////////////////////////////////////////////
// 禁止所有玩法 (高权限APP接口)
type StReq_BanPlayAll_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Time   int64  /* 禁止玩法时长（秒） */
	Reason string /* 提示内容 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_BanPlayAll struct {
	Head *StHead                `json:"head"`
	Body *StReq_BanPlayAll_Body `json:"body"`
}

// 返回值
type StRsp_BanPlayAll_Body struct {
	Result int
	RetMsg string
}

type StRsp_BanPlayAll struct {
	Head *StHead                `json:"head"`
	Body *StRsp_BanPlayAll_Body `json:"body"`
}

/////////////////////////////////////////////////
// 禁止参与排行榜
type StReq_BanJoinRank_Body struct {
	PlatId     int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId     string /* openid */
	IsZeroRank int
	Type       int    /* 榜单类型（1 顶级玩家排行榜，2，顶级联盟排行榜，99 全选） */
	Time       int64  /* 禁止时长（秒） */
	Reason     string /* 提示内容 */
	Source     int    /* 渠道号，由前端生成，不需要填写 */
	Serial     string /* 流水号，由前端生成，不需要填写 */
}

type StReq_BanJoinRank struct {
	Head *StHead                 `json:"head"`
	Body *StReq_BanJoinRank_Body `json:"body"`
}

// 返回值
type StRsp_BanJoinRank_Body struct {
	Result int
	RetMsg string
}

type StRsp_BanJoinRank struct {
	Head *StHead                 `json:"head"`
	Body *StRsp_BanJoinRank_Body `json:"body"`
}

/////////////////////////////////////////////////
// 初始化帐号
type StReq_InitAccount_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	IsInit int    /* 是否初始化（0 否  1 是） */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_InitAccount struct {
	Head *StHead                 `json:"head"`
	Body *StReq_InitAccount_Body `json:"body"`
}

// 返回值
type StRsp_InitAccount_Body struct {
	Result int
	RetMsg string
}

type StRsp_InitAccount struct {
	Head *StHead                 `json:"head"`
	Body *StRsp_InitAccount_Body `json:"body"`
}

/////////////////////////////////////////////////
// 禁言
type StReq_MaskChat_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Time   int64  /* 禁言时长*/
	Reason string /* 禁言原因 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_MaskChat struct {
	Head *StHead              `json:"head"`
	Body *StReq_MaskChat_Body `json:"body"`
}

// 返回值
type StRsp_MaskChat_Body struct {
	Result int
	RetMsg string
}

type StRsp_MaskChat struct {
	Head *StHead              `json:"head"`
	Body *StRsp_MaskChat_Body `json:"body"`
}

/////////////////////////////////////////////////
// 零收益
type StReq_ZeroProfit_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Time   int64  /* 零收益时长*/
	Reason string /* 原因 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_ZeroProfit struct {
	Head *StHead                `json:"head"`
	Body *StReq_ZeroProfit_Body `json:"body"`
}

// 返回值
type StRsp_ZeroProfit_Body struct {
	Result int
	RetMsg string
}

type StRsp_ZeroProfit struct {
	Head *StHead                `json:"head"`
	Body *StRsp_ZeroProfit_Body `json:"body"`
}

/////////////////////////////////////////////////
// 解除处罚
type StReq_RelievePunish_Body struct {
	PlatId             int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId             string /* openid */
	RelieveZeroProfit  int    /* 解除零收益状态（0 否，1 是） */
	RelievePlay        int    /* 解除所有玩法限制（0 否，1 是） */
	RelieveBanJoinRank int    /* 解除禁止参排行榜限制（0 否，1 是） */
	RelieveBan         int    /* 解除封号（0 否，1 是） */
	RelieveMaskchat    int    /* 解除禁言（0 否，1 是） */
	Source             int    /* 渠道号，由前端生成，不需要填写 */
	Serial             string /* 流水号，由前端生成，不需要填写 */
}

type StReq_RelievePunish struct {
	Head *StHead                   `json:"head"`
	Body *StReq_RelievePunish_Body `json:"body"`
}

// 返回值
type StRsp_RelievePunish_Body struct {
	Result int
	RetMsg string
}

type StRsp_RelievePunish struct {
	Head *StHead                   `json:"head"`
	Body *StRsp_RelievePunish_Body `json:"body"`
}

/////////////////////////////////////
//设置pve关卡数
type StReq_PveStage_Body struct {
	AreaId    int    /* 服务器：（1）微信，（2）手Q，Both（3） */
	PlatId    int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	Partition int    /* 小区 */
	OpenId    string /* openid */
	Type      int    /* 游戏类型（1 征战天下，99为全选） */
	Value     int    /* 设定值 */
	Source    int    /* 渠道号，由前端生成，不需填写 */
	Serial    string /* 流水号，由前端生成，不需填写 */
}

type StReq_PveStage struct {
	Head *StHead              `json:"head"`
	Body *StReq_PveStage_Body `json:"body"`
}

//返回值
type StRsp_PveStage_Body struct {
	Result int    /* 结果：0 成功，其它失败 */
	RetMsg string /* 返回消息 */
}

type StRsp_PveStage struct {
	Head *StHead              `json:"head"`
	Body *StRsp_PveStage_Body `json:"body"`
}

///////////////////////////////////////////////
//查询基础信息2
type StReq_BaseQuery2_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
}

type StReq_BaseQuery2 struct {
	Head *StHead                `json:"head"`
	Body *StReq_BaseQuery2_Body `json:"body"`
}

//返回值
type StRsp_BaseQuery2_Body struct {
	RoleName   string /* 角色名称 */
	Diamond    int64  /* 钻石数量 */
	Money      int64  /* 银两数量 */
	Food       int64  /* 粮草数量 */
	VioletGold int64  /* 紫水晶数量 */
	Status     int    /* 征战天下进度 */
	CityLevel  int    /* 主城等级 */
	Level      int    /* 玩家等级 */
	Flag       int64  /* 令旗数量 */
}

type StRsp_BaseQuery2 struct {
	Head *StHead                `json:"head"`
	Body *StRsp_BaseQuery2_Body `json:"body"`
}

// /////////////////////////////////
// 修改游戏币数量
type StReq_ChangeResValue_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Type   int    /* 游戏币类型（0 银两，1 粮草，2紫水晶） */
	Value  int64  /* 变动值（正数增加，负数减少） */
	Source int    /* 渠道号，由前端生成，不需填写 */
	Serial string /* 流水号，由前端生成，不需填写 */
}

type StReq_ChangeResValue struct {
	Head *StHead                    `json:"head"`
	Body *StReq_ChangeResValue_Body `json:"body"`
}

//返回值
type StRsp_ChangeResValue_Body struct {
	Result int    /* 结果：0 成功，其它失败 */
	RetMsg string /* 返回消息 */
}

type StRsp_ChangeResValue struct {
	Head *StHead                    `json:"head"`
	Body *StRsp_ChangeResValue_Body `json:"body"`
}

// /////////////////////////////////
// 修改宝石数量
type StReq_ChangeGemValue_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Type   int    /*  */
	Value  int64  /* 变动值（正数增加，负数减少） */
	Source int    /* 渠道号，由前端生成，不需填写 */
	Serial string /* 流水号，由前端生成，不需填写 */
}

type StReq_ChangeGemValue struct {
	Head *StHead                    `json:"head"`
	Body *StReq_ChangeGemValue_Body `json:"body"`
}

//返回值
type StRsp_ChangeGemValue_Body struct {
	Result int    /* 结果：0 成功，其它失败 */
	RetMsg string /* 返回消息 */
}

type StRsp_ChangeGemValue struct {
	Head *StHead                    `json:"head"`
	Body *StRsp_ChangeGemValue_Body `json:"body"`
}

/////////////////////////////////////////////////
// 封号
type StReq_LimitLogin_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Time   int64  /* 时长*/
	Reason string /* 原因 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_LimitLogin struct {
	Head *StHead                `json:"head"`
	Body *StReq_LimitLogin_Body `json:"body"`
}

// 返回值
type StRsp_LimitLogin_Body struct {
	Result int
	RetMsg string
}

type StRsp_LimitLogin struct {
	Head *StHead                `json:"head"`
	Body *StRsp_LimitLogin_Body `json:"body"`
}

/////////////////////////////////////////////////
// 发送系统邮件
type StReq_NewSendMail_Body struct {
	PlatId      int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId      string /* openid */
	MailContent string /* 邮件内容 */
	Source      int    /* 渠道号，由前端生成，不需要填写 */
	Serial      string /* 流水号，由前端生成，不需要填写 */
}

type StReq_NewSendMail struct {
	Head *StHead                 `json:"head"`
	Body *StReq_NewSendMail_Body `json:"body"`
}

// 返回值
type StRsp_NewSendMail_Body struct {
	Result int
	RetMsg string
}

type StRsp_NewSendMail struct {
	Head *StHead                 `json:"head"`
	Body *StRsp_NewSendMail_Body `json:"body"`
}

/////////////////////////////////////
// 太守等级修改
type StReq_OfficeModify_Body struct {
	PlatId int    /* 平台：IOS（0），安卓（1） ,Both（2） */
	OpenId string /* openid */
	Value  int    /* 数量：+加-减 */
	Source int    /* 渠道号，由前端生成，不需要填写 */
	Serial string /* 流水号，由前端生成，不需要填写 */
}

type StReq_OfficeModify struct {
	Head *StHead                  `json:"head"`
	Body *StReq_OfficeModify_Body `json:"body"`
}

//返回值
type StRsp_OfficeModify_Body struct {
	Result   int    /* 结果（0）成功 */
	RetMsg   string /* 返回消息 */
	EndValue int    /* 修改后太守等级 */
}

type StRsp_OfficeModify struct {
	Head *StHead                  `json:"head"`
	Body *StRsp_OfficeModify_Body `json:"body"`
}
