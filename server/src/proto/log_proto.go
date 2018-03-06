package proto

const (
	ResType_Gold             = "gold"
	ResType_Food             = "food"
	ResType_Wuhun            = "wuhun"
	ResType_Gem              = "diamonds"
	ResType_Trophy           = "trophy"
	ResType_TiLi             = "tili"
	ResType_TTTScore         = "tttscore"
	ResType_Dbt              = "Dbt"
	ResType_Bullion          = "Bullion"
	ResType_HolyWater        = "holywater"
	ResType_CivilizationCoin = "civilizationcoin"
	ResType_ClanExp          = "clanexp"
	ResType_ClanContribute   = "clancontribute"

	Gain_Gather                  = 1


	// Gain条目到49后 从200开始 50-200留给Lose
	// 这里添加的新的条目要到 server/src/tlogserver/TlogV0.0.1.xml文件中添加对应的条目

	Lose_Plunder               = 50 //抢劫别人损失

)

type LogPlayerLoginLogout struct {
	ChannelId uint8
	Playerid  string
	Time      int64
	Logout    bool
	Ip        string
}

type LogPlayerLoginLogoutResult struct {
}

type LogResources struct {
	ChannelId uint8
	Uid       string
	Time      int64
	Gain      bool
	ResType   string
	ResNum    uint32
	ResWay    uint32
}

type LogResourcesResult struct {
}

type TaobaoPayLog struct {
	TradeEnd    bool
	TradeError  string
	TradeNumber string
	CharId      string
	ItemName    string
	TotoalPee   string
	TradeTime   int64
	ChannelId   uint32
}

type TaobaoPayLogResult struct {
}

type GameServerState struct {
	Gameappid        string
	Timekey          uint64
	Gsid             uint32
	Zoneid           uint32
	Onlinecntios     uint32
	Onlinecntandroid uint32
}

type GameServerStateResult struct {
}

type PlayerState struct {
	Gameappid      string
	Openid         string
	Regtime        uint64
	Level          uint32
	IFriends       uint32
	Diamondios     uint32
	Diamondandroid uint32
}

type PlayerStateResult struct {
}
