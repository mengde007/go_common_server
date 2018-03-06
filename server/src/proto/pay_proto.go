package proto

const (
	PayItem_MonthCard = "com" //月卡
)

type QueryPayInfo struct {
	OpenId string
}

type QueryPayInfoRst struct {
	ItemId string
}

type NotifyPlayerGetPayInfo struct {
	Uid string
}

type CreateOrder struct {
	Uid    string
	OpenId string
	ItemId string
	Ip     string
	Money  uint32
}

type CreateOrderRst struct {
	OrderNum  string
	Appid     string
	Noncestr  string
	Package   string
	Partnerid string
	Prepayid  string
	Timestamp string
	Sign      string
}

type RechargeStatisticReq struct {
}

type RechargeStatisticRst struct {
	Value int
}
