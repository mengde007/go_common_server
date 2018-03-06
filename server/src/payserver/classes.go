package payserver

//订单信息
type OrderInfo struct {
	OrderNum   string
	Uid        string
	CreateTime uint32
	ItemId     string
	PrepayId   string //预支付交易会话标识
	AppId      string
	PartnerId  string
	NonceSt    string
	TimeStamp  string
	Sign       string
	OpenId     string
}

//预付请求
type WechatPrepayReq struct {
	Appid          string `xml:"appid"`            //应用ID
	Mchid          string `xml:"mch_id"`           //商户号
	Noncestr       string `xml:"nonce_str"`        //随机字符串
	Sign           string `xml:"sign"`             //签名
	Body           string `xml:"body"`             //商品描述
	OutTradeNo     string `xml:"out_trade_no"`     //商户订单号
	TotalFee       uint32 `xml:"total_fee"`        //总金额
	SpbillCreateIp string `xml:"spbill_create_ip"` //终端IP
	NotifyUrl      string `xml:"notify_url"`       //通知地址
	TradeType      string `xml:"trade_type"`       //交易类型
}

//预付返回
type WechatPrepayRst struct {
	ReturnCode string `xml:"return_code"` //返回状态码
	ReturnMsg  string `xml:"return_msg"`  //返回信息
	Appid      string `xml:"appid"`
	MchId      string `xml:"mch_id"`
	NonceStr   string `xml:"nonce_str"`    //微信返回的随机字符串
	Sign       string `xml:"sign"`         //微信返回的签名
	ResultCode string `xml:"result_code"`  //业务结果SUCCESS/FAIL
	ErrCode    string `xml:"err_code"`     //错误代码
	TradeType  string `xml:"trade_type"`   //交易类型
	PrepayId   string `xml:"prepay_id"`    //预支付交易会话标识
	ErrCodeDes string `xml:"err_code_des"` //错误描述

}

//微信支付回调
type stWechatCallBack struct {
	ReturnCode     string `xml:"return_code"`      //返回状态码
	ReturnMsg      string `xml:"return_msg"`       //返回信息
	Appid          string `xml:"appid"`            //应用ID
	MchId          string `xml:"mch_id"`           //商户号
	NonceStr       string `xml:"nonce_str"`        //随机字符串
	Sign           string `xml:"sign"`             //签名
	ResulCode      string `xml:"result_code"`      //业务结果
	ErrCode        string `xml:"err_code"`         //错误代码
	ErrCodeDes     string `xml:"err_code_des"`     //错误代码描述
	Openid         string `xml:"openid"`           //用户标识
	TradeType      string `xml:"trade_type"`       //交易类型
	BankType       string `xml:"bank_type"`        //付款银行
	TotalFee       string `xml:"total_fee"`        //总金额
	CashFee        string `xml:"cash_fee"`         //现金支付金额
	TransactionId  string `xml:"transaction_id"`   //微信支付订单号
	OutTradeNo     string `xml:"out_trade_no"`     //商户订单号
	TimeEnd        string `xml:"time_end"`         //支付完成时间
	TradeStateDesc string `xml:"trade_state_desc"` //支付完成时间
	PrepayId       string `xml:"prepay_id"`        //预支付Is
	FeeType        string `xml:"fee_type"`
	IsSubscribe    string `xml:"is_subscribe"`
}

//返回给微信的结果
type stReturnWechat struct {
	ReturnCode string `xml:"return_code"` //返回状态码
	ReturnMsg  string `xml:"return_msg"`  //返回信息
}

//查询支付请求
type stWechatPayQuery struct {
	Appid         string `xml:"appid"`          //应用ID
	MchId         string `xml:"mch_id"`         //商户号
	NonceStr      string `xml:"nonce_str"`      //随机字符串
	RransactionId string `xml:"transaction_id"` //商户订单号
	Sign          string `xml:"sign"`           //签名
}

//old====================================================================

//从gameserver来的请求生成订单的请求
type stCreateOrderMsg struct {
	Uid       string `json:"uid"`
	ItemId    string `json:"itemid"`
	MoneyYuan uint32 `json:"moneyyuan"`
}

//生成订单的结果
type stCreateOrderMsgRst struct {
	OrderNum  string `json:"ordernum"`
	IsSuccess bool   `json:"issuccess"`
	ChannelId string `json:"channelid"`
}

//向gameserver回调，成功充值后的结果,gameserver拿到之后再向路由反查
type stPaySucessMsg struct {
	Uid      string `json:"uid"`
	ItemId   string `json:"itemid"`
	OrderNum string `json:"ordernum"`
}

//gameserver回调后的结果，是否操作成功等
type stPaySucessMsgRst struct {
	IsSuccess bool `json:"issuccess"`
}
