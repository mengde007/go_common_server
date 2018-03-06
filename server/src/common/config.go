package common

import (
	"csvcfg"
	"jscfg"
	"logger"
	"os"
	"path"
	"rpc"
	"strconv"
	"strings"
	"time"
	"timer"
)

//gateserver配置
type GateServerCfg struct {
	GsIpForClient string
	GsIpForServer string
	DebugHost     string
	GcTime        uint8
	CpuProfile    bool
	VersionOld    uint32
	VersionNew    uint32
	DownloadUrl   string
}

//version配置
type VersionCfg struct {
	VersionOld  int32
	VersionNew  int32
	VersionMid  int32
	DownloadUrl string
}

//gas rpc配置
type GasProfileCfg struct {
	RpcProfile uint32
}

//center配置
type CenterConfig struct {
	Host       string
	HostForGm  string
	DebugHost  string
	CpuProfile bool
	GcTime     uint8
	Maincache  CacheConfig
	//add for update rankresult
	UpdateTime uint32
}

//chatserver配置
type ChatServerCfg struct {
	ListenForClient string
	ListenForServer string
	ListenForGm     string
	DebugHost       string
	CpuProfile      bool
	GcTime          uint8
	Maincache       CacheConfig
}

//cns配置
type CnsConfig struct {
	CnsHost          string
	CnsHostForClient string
	CnsForCenter     string
	FsHost           []string
	DebugHost        string
	GcTime           uint8
	CpuProfile       bool
	ServerId         uint8
	MaxPlayerCount   int32
}

//邮件服务器配置
type MailConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	UpdateTime uint32
	Maincache  CacheConfig
}

//大二配置
type DaerConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//麻将配置
type MaJiangConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//扑克配置
type PockerConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//房间配置
type RoomConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//比赛服务配置
type MatchConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//比赛服务配置
type MatchDaerConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//角色服配置
type RoleConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//gmserver配置
type GmServerCfg struct {
	Host       string
	InnerHost  string
	DebugHost  string
	GcTime     uint32
	CpuProfile bool
	ServerType uint32
	Maincache  CacheConfig
}

//好友信息
type FriendServerConfig struct {
	Host       string
	DebugHost  string
	GcTime     uint8
	CpuProfile bool
	Maincache  CacheConfig
}

//designer路径
type DesignerDir struct {
	Designer          string
	OpenGm            int32
	ServerType        int32
	WeChatPayPreOrder string //微信预支付地址
	Appid             string //微信开放平台审核通过的应用APPID
	Mchid             string //微信支付分配的商户号
	CPKey             string //商户Key
	WeChatQueryUrl    string //查询订单地址
	VersionUrl        string //版本信息地址
}

//payserver 配置
type PaySereverCfg struct {
	Host       string
	InnerHost  string
	DebugHost  string
	GcTime     uint32
	CpuProfile bool
	PayProxy   string
	Maincache  CacheConfig
}

// openid白名单配置
type LimitLoginConfig struct {
	OpenCheck            bool
	CheckList            []string
	RegisterOnly         bool
	VerifyServer         bool
	ServerTypeForQQGroup int
	OpenRobotChannel     bool
}

//add for get daily money
type GlobalInfo struct {
	TID         string //分数
	Mark        string //爵位
	AwardType1  string //宝石
	AwardCount1 uint32 //宝石数量
	AwardType2  string //武魂
	AwardCount2 uint32 //武魂数量
}

var globalinfoCfg map[string]*[]GlobalInfo

//add for challenge
var challengeCfg map[string]*[]ChallengeCfg

type ChallengeCfg struct {
	Value int32
}

//add for send present
type PresentInfo struct {
	Tite   string //标志
	Pic    string //资源类型图标
	Number uint32 //数量
	Type   string //资源名称
}

var sendInfoCfg map[string]*[]PresentInfo

func GetPresentCfg(key string) *PresentInfo {
	cfg, exist := sendInfoCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

//add for challenge
func LoadChallengeConfigFiles() {
	challengeInfo := path.Join(GetDesignerDir(), "globals.csv")
	csvcfg.LoadCSVConfig(challengeInfo, &challengeCfg)
}

func GetChallengeInfoCfg(key string) int32 {
	cfg, exist := challengeCfg[strings.ToLower(key)]
	if !exist {
		return 0
	}

	return (*cfg)[0].Value
}

// openid 白名单
func ReadLimitLoginConfig(cfg *LimitLoginConfig) error {
	// cfgpath, _ := os.Getwd()

	// if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"limitlogin.json"), cfg); err != nil {
	// 	logger.Fatal("read LimitLogin config failed, %v", err)
	// 	return err
	// }

	return nil
}

//center
func ReadCenterConfig(cfg *CenterConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, "../cfg/centerserver.json"), cfg); err != nil {
		logger.Fatal("read center config failed, %v", err)
		return err
	}

	return nil
}

//gm
func ReadGmConfig(cfg *GmServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"gmserver.json"), cfg); err != nil {
		logger.Fatal("read chat config failed, %v", err)
		return err
	}

	return nil
}

//chat
func ReadChatConfig(cfg *ChatServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"chatserver.json"), cfg); err != nil {
		logger.Fatal("read chat config failed, %v", err)
		return err
	}

	return nil
}

//加锁服务器
func ReadLockServerConfig(file string, cfg *LockServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, file), cfg); err != nil {
		logger.Fatal("read lock config failed, %v", err)
		return err
	}

	return nil
}

//cns服务器配置
func ReadCnsServerConfig(file string, cfg *CnsConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, file), cfg); err != nil {
		logger.Fatal("read cns config failed, %v", err)
		return err
	}

	return nil
}

//邮件服务器配置表
func ReadMailServerConfig(cfg *MailConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"mailserver.json"), cfg); err != nil {
		logger.Fatal("read mail config failed, %v", err)
		return err
	}

	return nil
}

//gate服务器
func ReadGateServerConfig(cfg *GateServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"gateserver.json"), cfg); err != nil {
		logger.Fatal("read ttt config failed, %v", err)
		return err
	}

	return nil
}

//读取version配置
// func ReadVersionConfig(cfg *VersionCfg) error {
// 	cfgpath, _ := os.Getwd()

// 	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"version.json"), cfg); err != nil {
// 		logger.Fatal("read version config failed, %v", err)
// 		return err
// 	}

// 	return nil
// }

//读取version配置
func ReadProfileConfig(cfg *GasProfileCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"gasprofile.json"), cfg); err != nil {
		logger.Fatal("read gasprofile config failed, %v", err)
		return err
	}

	return nil
}

//pay
func ReadPayConfig(cfg *PaySereverCfg) error {
	cfgpath, _ := os.Getwd()
	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"payserver.json"), cfg); err != nil {
		logger.Fatal("read pay server config failed: ", err)
		return err
	}
	return nil
}

//大二服务器配置
func ReadDaerServerConfig(cfg *DaerConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"daerserver.json"), cfg); err != nil {
		logger.Fatal("read ReadDaerConfig config failed, %v", err)
		return err
	}

	return nil
}

//麻将服务器配置
func ReadMaJiangServerConfig(cfg *MaJiangConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"majiangserver.json"), cfg); err != nil {
		logger.Fatal("read ReadDaerConfig config failed, %v", err)
		return err
	}

	return nil
}

//德州扑克服务器配置
func ReadPockerServerConfig(cfg *PockerConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"pockerserver.json"), cfg); err != nil {
		logger.Fatal("read ReadPockerServerConfig config failed, %v", err)
		return err
	}

	return nil
}

//GeneralRankServer config
type GeneralRankServerCfg struct {
	GeneralRankHost string
	DebugHost       string
	GcTime          uint8
	CpuProfile      bool
	Maincache       CacheConfig
}

func ReadGeneralRankServerCfg(cfg *GeneralRankServerCfg) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"rankserver.json"), cfg); err != nil {
		logger.Fatal("read ReadGeneralRankServerCfg config failed, %v", err)
		return err
	}

	return nil
}

//房间服务器配置
func RoomServerConfig(cfg *RoomConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"roomserver.json"), cfg); err != nil {
		logger.Fatal("read RoomServerConfig config failed, %v", err)
		return err
	}

	return nil
}

//比赛服务器配置
func MatchServerConfig(cfg *MatchConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"matchserver.json"), cfg); err != nil {
		logger.Fatal("read RoomServerConfig config failed, %v", err)
		return err
	}

	return nil
}

//比赛服务器配置
func MatchDaerServerConfig(cfg *MatchDaerConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"matchdaerserver.json"), cfg); err != nil {
		logger.Fatal("read RoomServerConfig config failed, %v", err)
		return err
	}

	return nil
}

//角色服务器配置
func ReadRoleServerConfig(cfg *RoleConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"roleserver.json"), cfg); err != nil {
		logger.Fatal("read RoleConfig config failed, %v", err)
		return err
	}

	return nil
}

//好友服务器
func ReadFriendServerConfig(cfg *FriendServerConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"friendserver.json"), cfg); err != nil {
		logger.Fatal("read friend config failed, %v", err)
		return err
	}

	return nil
}

//account配置
func ReadAccountConfig(file string, cfg *DBConfig) error {
	return ReadDbConfig(file, cfg)
}

//designer配置
var pDesignerCfg *DesignerDir

func init() {
	cfgpath, _ := os.Getwd()
	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"designer.json"), &pDesignerCfg); err != nil {
		logger.Fatal("read designer config failed, %v", err)
		return
	}
}

func GetDesignerCfg() *DesignerDir {
	return pDesignerCfg
}

//designer
func GetDesignerDir() string {
	return GetDesignerCfg().Designer
}

//是否打开gm指令
func IsOpenGm() bool {
	return GetDesignerCfg().OpenGm == 1
}

//全局配置表
type GlobalCfg struct {
	Value uint32
}

var mapGlobalCfg map[string]*[]GlobalCfg

func LoadGlobalConfig() {
	filename := path.Join(GetDesignerDir(), "globals.csv")
	csvcfg.LoadCSVConfig(filename, &mapGlobalCfg)
}

func GetGlobalConfig(key string) uint32 {
	cfg, exist := mapGlobalCfg[strings.ToLower(key)]
	if !exist {
		return 0
	}

	return (*cfg)[0].Value
}

//全局配置表
type DaerGlobalCfg struct {
	IntValue    int32
	StringValue string
}

var mapDaerGlobalCfg map[string]*[]DaerGlobalCfg

func LoadDaerGlobalConfig() {
	filename := path.Join(GetDesignerDir(), "全局配置表.csv")
	csvcfg.LoadCSVConfig(filename, &mapDaerGlobalCfg)
}

func GetDaerGlobalConfig(key string) *DaerGlobalCfg {
	cfg, exist := mapDaerGlobalCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

func GetDaerGlobalIntValue(key string) int32 {
	cfg, exist := mapDaerGlobalCfg[strings.ToLower(key)]
	if !exist {
		return int32(0)
	}

	return (&(*cfg)[0]).IntValue
}

func GetGlobalStringValue(key string) string {
	cfg, exist := mapDaerGlobalCfg[strings.ToLower(key)]
	if !exist {
		return ""
	}

	return (&(*cfg)[0]).StringValue
}

//房间配置表
type DaerRoomCfg struct {
	Type            int32
	Difen           int32
	MinLimit        int32
	MaxLimit        int32
	GameType        int32
	MaxMultiple     int32
	RakeRate        int32
	AntiCheating    int32
	QiHuKeAmount    int32
	HongZhongAmount int32
	PockerExchange  int32
}

var mapDaerRoomCfg map[string]*[]DaerRoomCfg

func LoadDaerRoomConfig() {
	filename := path.Join(GetDesignerDir(), "房间.csv")
	csvcfg.LoadCSVConfig(filename, &mapDaerRoomCfg)
}

func GetDaerRoomConfig(key string) *DaerRoomCfg {
	cfg, exist := mapDaerRoomCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

//扑克配置表
type PockerCfg struct {
	Id             int32
	Difen          int32
	RakeRate       int32
	MinLimit       int32
	MaxLimit       int32
	PockerExchange int32
}

var mapPockerfg map[string]*[]PockerCfg

func LoadPockerConfig() {
	filename := path.Join(GetDesignerDir(), "扑克.csv")
	csvcfg.LoadCSVConfig(filename, &mapPockerfg)
}

func GetPockerConfig(key string) *PockerCfg {
	cfg, exist := mapPockerfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

//获取最佳房间类型
func GetBestRoomType(gameType int32, playerInfo *rpc.PlayerBaseInfo) int32 {
	roomType := -1

	var minOffset int32 = 1000000000

	for id, cfgs := range mapDaerRoomCfg {
		if cfgs == nil {
			continue
		}

		for _, cfg := range *cfgs {
			if gameType != cfg.GameType {
				continue
			}

			offset := playerInfo.GetCoin() - cfg.MinLimit
			if offset >= 0 && offset < minOffset {
				minOffset = offset
				roomType, _ = strconv.Atoi(id)
			}
		}
	}

	return int32(roomType)
}

// //全局配置表
// type TextsCfg struct {
// 	DescCN string
// }

// var mapTextsCfg map[string]*[]TextsCfg

// func LoadTextsCfg() {
// 	filename := path.Join(GetDesignerDir(), "全局.csv")
// 	csvcfg.LoadCSVConfig(filename, &mapTextsCfg)
// }

// func GetTextsCfg(key string) *TextsCfg {
// 	cfg, exist := mapTextsCfg[strings.ToLower(key)]
// 	if !exist {
// 		return nil
// 	}

// 	return &(*cfg)[0]
// }

//客户端连接
var mapServerClientCfg map[string][]string

func init() {
	if err := ReloadServerClientConfig(); err != nil {
		logger.Fatal("read serverclient config failed, %v", err)
	}
}

//客户端配置
func ReadServerClientConfig(server string) []string {
	if info, ok := mapServerClientCfg[server]; ok {
		return info
	}

	return nil
}

//重新读取
func ReloadServerClientConfig() error {
	mapTemp := make(map[string][]string)

	cfgpath, _ := os.Getwd()
	if err := jscfg.ReadJson(path.Join(cfgpath, logger.CfgBaseDir+"client.json"), &mapTemp); err != nil {
		logger.Error("ReloadServerClientConfig failed, %v", err)
		return err
	}

	mapServerClientCfg = mapTemp

	return nil
}

//注册读取配置表的tick
func RegisterReloadServerClientCfg(f func()) {
	tm := timer.NewTimer(time.Second * 30)
	tm.Start(func() {
		if err := ReloadServerClientConfig(); err != nil {
			return
		}

		f()
	})
}

//自建房间配置表
type CustomRoomCfg struct {
	GameType                  int32
	CoinDifenMulti            int32
	InitCredit                int32
	NameMinLength             int32
	NameMaxLength             int32
	PwdMinLength              int32
	PwdMaxLength              int32
	DifenMinLimit             int32
	DifenMaxLimit             int32
	EnterRoomMinLimit         int32
	TimesMinLimit             int32
	TimesMaxLimit             int32
	MaxPeople                 int32
	RechargeCoinTime          int32
	CreateRoomInterval        int32
	CreateRoomMinLimit        int32
	MinMultipleLimit          int32
	MaxMultipleLimit          int32
	RakeRate                  int32
	CoinWaitingReadyTime      int32
	CreditsWaitingReadyTime   int32
	CoinRoomDissolveTime      int32
	CreditsStartDissolveTime  int32
	CreditsMiddleDissolveTime int32
	CreateCreditsRoomCardCost string
}

var mapCustomRoomCfg map[string]*[]CustomRoomCfg

func LoadCustomRoomConfig() {
	filename := path.Join(GetDesignerDir(), "自建房间.csv")
	csvcfg.LoadCSVConfig(filename, &mapCustomRoomCfg)
}

func GetCustomRoomConfig(key string) *CustomRoomCfg {
	cfg, exist := mapCustomRoomCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

//比赛配置表
type MatchCfg struct {
	ID                    int32
	MatchType             int32
	GameType              int32
	StartMatchMode        int32
	InitCredit            int32
	StartTime             string
	EndTime               string
	StartMatchInterval    int32
	EntryThreshold        int32
	FullStartCount        int32
	EntryFeeCurrencyType  int32
	EntryFee              int32
	StartMatchPlayerLimit int32
	MatchSystem           int32
	MatchTime             int32
	Reward                string
	IsGiveBackCoin        int32
	IsShow                int32
}

var mapMatchCfg map[string]*[]MatchCfg

func LoadMatchConfig() {
	filename := path.Join(GetDesignerDir(), "比赛.csv")
	csvcfg.LoadCSVConfig(filename, &mapMatchCfg)
}

func GetMatchConfig(key string) *MatchCfg {
	cfg, exist := mapMatchCfg[strings.ToLower(key)]
	if !exist {
		return nil
	}

	return &(*cfg)[0]
}

func GetMatchConfigForAll() []*MatchCfg {
	result := make([]*MatchCfg, 0)

	for _, cfgs := range mapMatchCfg {
		result = append(result, &(*cfgs)[0])
	}

	return result
}
