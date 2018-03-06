package payserver

import (
	"bytes"
	"centerclient"
	"common"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"logger"
	"math/rand"
	"net"
	"net/http"
	"proto"
	"rpc"
	"rpcplus"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PayService struct {
	pCachePool *common.CachePool
	sl         *common.SimpleLockService
	payUrl     string
}

const (
	ORDER_TALBE = "order_table"
	ITEM_TABLE  = "item_table"
	PREODER     = "pre_order"
	STATISTIC   = "statistic"
)

var pPayService *PayService

//超时连接
func createTransport() *http.Transport {
	return common.CreateTransport()
}

func CreatePayServer() {
	var cfg common.PaySereverCfg
	if err := common.ReadPayConfig(&cfg); err != nil {
		logger.Error("ReadPayConfig failed", err)
		return
	}

	pPayService = &PayService{
		pCachePool: common.NewCachePool(cfg.Maincache),
		sl:         common.CreateSimpleLock(),
		payUrl:     cfg.Host,
	}

	//配置表
	//connector.LoadConfigFiles(common.GetDesignerDir())
	common.LoadGlobalConfig()

	wg := sync.WaitGroup{}
	wg.Add(2)

	//监听内网
	go pPayService.initTcp(&cfg, &wg)
	//监听网页
	go pPayService.initHttp(&cfg, &wg)

	wg.Wait()
}

func (self *PayService) initTcp(cfg *common.PaySereverCfg, wg *sync.WaitGroup) error {
	defer wg.Done()

	//监听
	listener, err := net.Listen("tcp", cfg.InnerHost)
	if err != nil {
		logger.Error("Listening to: %s %s", cfg.InnerHost, " failed !!")
		return err
	}
	defer listener.Close()

	rpcServer := rpcplus.NewServer()
	rpcServer.Register(pPayService)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("payServer StartServices %s", err.Error())
			break
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("payServer Rpc Runtime Error: %s", r)
					debug.PrintStack()
				}
			}()

			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}
	return nil
}

func (self *PayService) initHttp(cfg *common.PaySereverCfg, wg *sync.WaitGroup) error {
	defer wg.Done()

	http.HandleFunc("/", createHandleFunc(self.handle))

	//对外
	if err := http.ListenAndServe(cfg.Host, nil); err != nil {
		return err
	}

	return nil
}

//回调函数创建
type handleCALLBACK func(w http.ResponseWriter, r *http.Request)

func createHandleFunc(f handleCALLBACK) handleCALLBACK {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				errmsg := fmt.Sprintf("handle http failed :%s", r)
				writeString(w, "serious err occurred :", errmsg)
				return
			}
		}()

		f(w, r)
	}
}

const (
	REQHEAD    = "data_packet="
	SUCCESSMSG = "success"
	PAYTABLE   = "tb_all_pay"
)

func (self *PayService) handle(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("handle ReadAll body err", err)
		return
	}
	logger.Info("################收到支付回调,原始数据:%v", string(b))
	send := &stReturnWechat{}
	send.ReturnCode = "FAIL"

	rst := stWechatCallBack{}
	err = xml.Unmarshal(b, &rst)
	if err != nil {
		logger.Error("微信支付回调,解析xml出错:%v", string(b))
		writeXmlResult(w, send)
		return
	}

	//取出订单验证
	buf, err := common.Resis_getbuf(self.pCachePool, ORDER_TALBE, rst.OutTradeNo)
	if err != nil {
		logger.Error("微信支付回调, 取订单缓存失败, 订单号:%s, err:%s", rst.OutTradeNo, err)
		writeXmlResult(w, send)
		return
	}
	if buf == nil && err == nil {
		logger.Error("微信支付回调, 回调重复")
		send.ReturnMsg = "回调重复"
		writeXmlResult(w, send)
		return
	}
	orderInfo := &OrderInfo{}
	if err := common.GobDecode(buf, orderInfo); err != nil {
		logger.Error("微信支付回调, 解析订单缓存失败:%s", err)
		writeXmlResult(w, send)
		return
	}

	//缓存未完成订单信息
	complexId := rst.TransactionId + "_" + orderInfo.ItemId + "_" + orderInfo.Uid + "_" + orderInfo.OpenId
	err = common.Redis_setString(self.pCachePool, PREODER, rst.Openid, complexId)
	if err != nil {
		logger.Error("CreateOrder 缓存未完成订单失败  errorreq.Openid:%s", err, rst.Openid)
		writeXmlResult(w, send)
		return
	}

	//充值结果通知
	payRst := &rpc.PayResultNotify{}
	payRst.SetResult(false)
	payRst.SetPartnerId(orderInfo.ItemId)

	logger.Info("################收到支付回调:%v", rst)
	if rst.ReturnCode != "SUCCESS" {
		logger.Error("微信支付回调, 支持失败，原因:%s", rst.ReturnMsg)
		send.ReturnCode = "FAIL"
		payRst.SetErrorDesc(send.ReturnCode)
		writeXmlResult(w, send)
		centerclient.SendPayResult2Player([]string{orderInfo.Uid}, payRst)
		return
	}
	if rst.ResulCode != "SUCCESS" {
		logger.Error("微信支付回调, 支持失败，code:%s 原因:%s", rst.ErrCode, rst.ErrCodeDes)
		send.ReturnCode = "FAIL"
		payRst.SetErrorDesc(send.ReturnCode)
		writeXmlResult(w, send)
		centerclient.SendPayResult2Player([]string{orderInfo.Uid}, payRst)
		return
	}

	// 验证订单
	desiCfg := common.GetDesignerCfg()
	sign := self.gen_vertify_sign(&rst, desiCfg.CPKey)
	if sign != rst.Sign {
		send.ReturnMsg = "订单验证失败"
		logger.Error("订单验证失败，签名:%s, 签名wechat:%s", sign, rst.Sign)
		payRst.SetErrorDesc(send.ReturnMsg)
		writeXmlResult(w, send)
		centerclient.SendPayResult2Player([]string{orderInfo.Uid}, payRst)
		return
	}

	// //验证订单
	// if rst.MchId != orderInfo.PartnerId || rst.Openid != orderInfo.OpenId {
	// 	logger.Error("订单验证失败rst.Appid：%s, orderInfo.AppId:%s , rst.Openid:%s, orderInfo.OpendId:%s",
	// 		rst.MchId, orderInfo.PartnerId, rst.Openid, orderInfo.OpenId)
	// 	send.ReturnMsg = "订单验证失败"
	// 	payRst.SetErrorDesc(send.ReturnMsg)
	// 	writeXmlResult(w, send)
	// 	centerclient.SendPayResult2Player([]string{orderInfo.Uid}, payRst)
	// 	return
	// }

	//保存充值结果
	err = common.Redis_setString(self.pCachePool, ITEM_TABLE, rst.Openid, orderInfo.ItemId)
	if err != nil {
		logger.Error("微信支付回调, 保存支付结果出错:uid, itemId:%s", orderInfo.Uid, orderInfo.ItemId)
		payRst.SetErrorDesc("保存充值结果失败")
		writeXmlResult(w, send)
		centerclient.SendPayResult2Player([]string{orderInfo.Uid}, payRst)
		return
	}
	send.ReturnCode = "SUCCESS"
	writeXmlResult(w, send)

	//statistic recharge
	self.statistic(rst.TotalFee)

	//删除未完成订单信息
	err = common.Redis_del(self.pCachePool, PREODER, rst.Openid)
	if err != nil {
		logger.Error("删除未完成的订单失败，uid:%s, ", rst.Openid)

		return
	}
	//删除预付订单
	err = common.Redis_del(self.pCachePool, ORDER_TALBE, rst.OutTradeNo)
	if err != nil {
		logger.Error("微信支付回调, 删除订单失败, 订单号:%s", rst.OutTradeNo)
		return
	}
	payRst.SetResult(true)
	centerclient.SendPayResult2Player([]string{orderInfo.Uid}, payRst)
}

func (self *PayService) statistic(total_fee string) {
	logger.Info("statistic called total_fee:%s", total_fee)
	if total_fee == "" {
		return
	}
	fee, _ := strconv.Atoi(total_fee)
	value, err := common.Redis_getInt(self.pCachePool, ORDER_TALBE, STATISTIC)
	if err != nil {
		logger.Error("statistic  Redis_getInt err:%s", err)
		return
	}

	err = common.Redis_setInt(self.pCachePool, ORDER_TALBE, STATISTIC, fee+value)
	if err != nil {
		logger.Error("statistic  Redis_setInt err:%s", err)
		return
	}
}

func (self *PayService) GetRechargeStatistic(req *proto.RechargeStatisticReq, rst *proto.RechargeStatisticRst) error {
	logger.Info("GetRechargeStatistic called")
	value, err := common.Redis_getInt(self.pCachePool, ORDER_TALBE, STATISTIC)
	if err != nil {
		logger.Error("GetRechargeStatistic Redis_getInt err:%s", err)
		return nil
	}
	rst.Value = value
	return nil
}

//玩家创建订单
func (self *PayService) CreateOrder(req *proto.CreateOrder, rst *proto.CreateOrderRst) error {
	logger.Info("CreateOrder has been called...")
	self.sl.WaitLock(req.Uid)
	defer self.sl.WaitUnLock(req.Uid)

	client := &http.Client{
		Transport: createTransport(),
	}

	desiCfg := common.GetDesignerCfg()
	if desiCfg == nil {
		logger.Error("CreateOrder 获取desinger.json出错")
		return nil
	}

	//gen oder
	orderInfo := &OrderInfo{}
	orderNum := common.GenUUIDWith32(0)
	orderInfo.OrderNum = orderNum
	orderInfo.Uid = req.Uid
	orderInfo.ItemId = req.ItemId
	orderInfo.CreateTime = uint32(time.Now().Unix())
	orderInfo.AppId = desiCfg.Appid
	orderInfo.PartnerId = desiCfg.Mchid
	orderInfo.OpenId = req.OpenId

	prepayReq := &WechatPrepayReq{
		Appid:          orderInfo.AppId,
		Mchid:          orderInfo.PartnerId,
		Noncestr:       strconv.Itoa(rand.Intn(1000000)),
		Body:           "泸州棋牌-购买道具",
		OutTradeNo:     orderNum,
		TotalFee:       req.Money,
		SpbillCreateIp: req.Ip,
		NotifyUrl:      self.payUrl,
		TradeType:      "APP",
	}
	prepayReq.Sign = self.gen_prepay_sign(prepayReq, desiCfg.CPKey)
	logger.Info("CreateOrder 签名为:%s", prepayReq.Sign)
	logger.Info("CreateOrder OrderInfo", prepayReq)

	//生成xml
	body, err := xml.MarshalIndent(prepayReq, " ", " ")
	if err != nil {
		logger.Error("CreateOrder MarshalIndent error: %v", err)
		return err
	}
	buf := bytes.NewBuffer(body)

	//提交请求
	url := desiCfg.WeChatPayPreOrder
	logger.Info("post url:", url)
	res, err := client.Post(url, "application/x-www-form-urlencoded", buf)

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("CreateOrder body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("CreateOrder ioutil.ReadAll error: %v", err)
		return err
	}

	//解析xml
	prepayRst := WechatPrepayRst{}
	if err := xml.Unmarshal(b, &prepayRst); err != nil {
		logger.Error("CreateOrder ioutil.ReadAll error: %v", err)
		return nil
	}

	logger.Info("##########prepayRst:%v", prepayRst)
	if prepayRst.ReturnMsg != "OK" {
		logger.Error("CreateOrder FAILED, prepayRst.ReturnMsg:%s", prepayRst.ReturnMsg)
		return nil
	}
	if prepayRst.ResultCode != "SUCCESS" {
		logger.Error("CreateOrder FAILED, ErrorCode:%s, desc:%s", prepayRst.ErrCode, prepayRst.ErrCodeDes)
		return nil
	}
	orderInfo.PrepayId = prepayRst.PrepayId
	orderInfo.TimeStamp = strconv.Itoa(int(time.Now().Unix()))
	orderInfo.NonceSt = strconv.Itoa(rand.Intn(1000000))
	orderInfo.Sign = self.gen_prepay_sign2client(orderInfo, desiCfg.CPKey)

	logger.Info("#########订单信息:%v", orderInfo)
	newBuf, err := common.GobEncode(orderInfo)
	if err != nil {
		logger.Error("CreateOrder GobEncode  error, req.Uid:%s, itemId:%s", err, req.Uid, req.ItemId)
		return err
	}
	err = common.Resis_setbuf(self.pCachePool, ORDER_TALBE, orderNum, newBuf)
	if err != nil {
		logger.Error("CreateOrder setBuf  errorreq.Uid:%s, itemId:%s", err, req.Uid, req.ItemId)
		return err
	}

	err = common.Redis_setexpire(self.pCachePool, ORDER_TALBE, orderNum, "259200")
	if err != nil {
		logger.Error("CreateOrder setexpire  errorreq.Uid:%s, itemId:%s", err, req.Uid, req.ItemId)
		return err
	}

	rst.OrderNum = orderNum
	rst.Appid = orderInfo.AppId
	rst.Noncestr = orderInfo.NonceSt
	rst.Package = "Sign=WXPay"
	rst.Partnerid = orderInfo.PartnerId
	rst.Prepayid = orderInfo.PrepayId
	rst.Timestamp = orderInfo.TimeStamp
	rst.Sign = orderInfo.Sign
	return nil
}

func (self *PayService) gen_prepay_sign(info *WechatPrepayReq, key string) string {
	money := strconv.Itoa(int(info.TotalFee))
	stringA := "appid=" + info.Appid + "&body=" + info.Body +
		"&mch_id=" + info.Mchid + "&nonce_str=" + info.Noncestr +
		"&notify_url=" + info.NotifyUrl + "&out_trade_no=" + info.OutTradeNo +
		"&spbill_create_ip=" + info.SpbillCreateIp + "&total_fee=" + money +
		"&trade_type=" + info.TradeType

	stringSignTemp := stringA + "&key=" + key

	logger.Info("签名前:%s", stringSignTemp)

	h := md5.New()
	h.Write([]byte(stringSignTemp))
	cipherStr := hex.EncodeToString(h.Sum(nil))
	return strings.ToUpper(cipherStr)
}

func (self *PayService) gen_prepay_sign2client(info *OrderInfo, key string) string {
	stringA := "appid=" + info.AppId + "&noncestr=" + info.NonceSt +
		"&package=" + "Sign=WXPay" + "&partnerid=" + info.PartnerId +
		"&prepayid=" + info.PrepayId + "&timestamp=" + info.TimeStamp

	stringSignTemp := stringA + "&key=" + key

	h := md5.New()
	h.Write([]byte(stringSignTemp))
	cipherStr := hex.EncodeToString(h.Sum(nil))
	return strings.ToUpper(cipherStr)
}

//查询订单签名
func (self *PayService) gen_query_order_sign(info *stWechatPayQuery, key string) string {
	stringA := "appid=" + info.Appid + "&mch_id=" + info.MchId +
		"&nonce_str=" + info.NonceStr +
		"&transaction_id=" + info.RransactionId

	stringSignTemp := stringA + "&key=" + key

	logger.Info("签名前:%s", stringSignTemp)

	h := md5.New()
	h.Write([]byte(stringSignTemp))
	cipherStr := hex.EncodeToString(h.Sum(nil))
	return strings.ToUpper(cipherStr)
}

// appid
// bank_type
// cash_fee
// fee_type
// is_subscribe
// mch_id
// nonce_str
// openid
// out_trade_no
// result_code
// return_code
// time_end
// total_fee
// trade_type
// transaction_id

// <appid><![CDATA[wxac6228496497182c]]></appid>
// <bank_type><![CDATA[CCB_DEBIT]]></bank_type>
// <cash_fee><![CDATA[1]]></cash_fee>
// <fee_type><![CDATA[CNY]]></fee_type>
// <is_subscribe><![CDATA[N]]></is_subscribe>
// <mch_id><![CDATA[1422709602]]></mch_id>
// <nonce_str><![CDATA[498081]]></nonce_str>
// <openid><![CDATA[oplHewf0AO3U-Aq9wYgHOhK-OVXA]]></openid>
// <out_trade_no><![CDATA[00000100c81e34593cc96e59da8ec213]]></out_trade_no>
// <result_code><![CDATA[SUCCESS]]></result_code>
// <return_code><![CDATA[SUCCESS]]></return_code>
// <sign><![CDATA[BE5FFAD6EA6E388FC1491CBA5CC3DC5B]]></sign>
// <time_end><![CDATA[20170604225311]]></time_end>
// <total_fee>1</total_fee>
// <trade_type><![CDATA[APP]]></trade_type>
// <transaction_id><![CDATA[4005302001201706044312028606]]></transaction_id>

func (self *PayService) gen_vertify_sign(info *stWechatCallBack, key string) string {
	stringA := "appid=" + info.Appid + "&bank_type=" + info.BankType + "&cash_fee=" + info.CashFee +
		"&fee_type=" + info.FeeType + "&is_subscribe=" + info.IsSubscribe +
		"&mch_id=" + info.MchId + "&nonce_str=" + info.NonceStr +
		"&openid=" + info.Openid + "&out_trade_no=" + info.OutTradeNo +
		"&result_code=" + info.ResulCode + "&return_code=" + info.ReturnCode +
		"&time_end=" + info.TimeEnd + "&total_fee=" + info.TotalFee +
		"&trade_type=" + info.TradeType + "&transaction_id=" + info.TransactionId

	stringSignTemp := stringA + "&key=" + key

	h := md5.New()
	h.Write([]byte(stringSignTemp))
	cipherStr := hex.EncodeToString(h.Sum(nil))
	return strings.ToUpper(cipherStr)

}

func (self *PayService) QueryOrder(openId string) {
	client := &http.Client{
		Transport: createTransport(),
	}

	logger.Info("查询订单，QueryOrder：%s", openId)
	complexId, err := common.Redis_getString(self.pCachePool, PREODER, openId)
	if err != nil {
		logger.Error("QueryOrder 查询订单 common.Redis_getString error:", err, openId)
		return
	}
	if complexId == "" {
		logger.Info("此玩家没有未完成的订单：%s", openId)
		return
	}
	ids := strings.Split(complexId, "_")
	transactionId := ids[0]

	desiCfg := common.GetDesignerCfg()
	if desiCfg == nil {
		logger.Error("查询订单 获取desinger.json出错")
		return
	}

	//正确取出未完成的订单,到微信查询
	prepayReq := &stWechatPayQuery{
		Appid:         desiCfg.Appid,
		MchId:         desiCfg.Mchid,
		NonceStr:      strconv.Itoa(rand.Intn(1000000)),
		RransactionId: transactionId,
	}
	prepayReq.Sign = self.gen_query_order_sign(prepayReq, desiCfg.CPKey)
	logger.Info("QueryOrder 签名为:%s", prepayReq.Sign)

	//生成xml
	body, err := xml.MarshalIndent(prepayReq, " ", " ")
	if err != nil {
		logger.Error("CreateOrder MarshalIndent error: %v", err)
		return
	}
	bufBody := bytes.NewBuffer(body)

	//提交请求
	url := desiCfg.WeChatQueryUrl
	logger.Info("post url:", url)
	res, err := client.Post(url, "application/x-www-form-urlencoded", bufBody)

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("QueryOrder body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("QueryOrder ioutil.ReadAll error: %v", err)
		return
	}
	logger.Info("################查询订单,原始数据:%v", string(b))

	rst := stWechatCallBack{}
	err = xml.Unmarshal(b, &rst)
	if err != nil {
		logger.Error("查询订单,解析xml出错:%v", string(b))
		return
	}
	logger.Info("################收到支付回调:%v", rst)
	if rst.ReturnCode != "SUCCESS" {
		logger.Error("查询订单, 支持失败，原因:%s", rst.ReturnMsg)
		return
	}
	if rst.ResulCode != "SUCCESS" {
		logger.Error("查询订单失败，code:%s 原因:%s", rst.ErrCode, rst.ErrCodeDes)
		return
	}
	//取出订单验证
	// buf, err := common.Resis_getbuf(self.pCachePool, ORDER_TALBE, rst.OutTradeNo)
	// if err != nil {
	// 	logger.Error("查询订单, 取订单缓存失败, 订单号:%s, err:%s", rst.OutTradeNo, err)
	// 	return
	// }
	// if buf == nil && err == nil {
	// 	logger.Error("查询订单, 回调重复")
	// 	return
	// }
	// orderInfo := &OrderInfo{}
	// if err := common.GobDecode(buf, orderInfo); err != nil {
	// 	logger.Error("查询订单, 解析订单缓存失败:%s", err)
	// 	return
	// }

	// 验证订单
	sign := self.gen_vertify_sign(&rst, desiCfg.CPKey)
	if rst.Sign != sign {
		logger.Error("订单验证失败，签名:%s, 签名wechat:%s", sign, rst.Sign)
		// return
	}
	// if rst.Openid != ids[3] {
	// 	logger.Error("订单验证失败  rst.Openid:%s orderInfo.OpenId:%s", rst.Openid, ids[3])
	// 	return
	// }

	//保存充值结果
	err = common.Redis_setString(self.pCachePool, ITEM_TABLE, rst.Openid, ids[1])
	if err != nil {
		logger.Error("查询订单, 保存支付结果出错:openId:%s, itemId:%s", rst.Openid, ids[1])
		return
	}

	//删除预付订单缓存
	err = common.Redis_del(self.pCachePool, ORDER_TALBE, rst.OutTradeNo)
	if err != nil {
		logger.Error("查询订单, 删除订单失败, 订单号:%s", rst.OutTradeNo)
		return
	}

	//删除未完成订单信息
	err = common.Redis_del(self.pCachePool, PREODER, rst.Openid)
	if err != nil {
		logger.Error("删除未完成的订单失败，uid:%s, ", rst.Openid)
		return
	}

	//通知玩家
	payRst := &rpc.PayResultNotify{}
	payRst.SetResult(true)
	payRst.SetPartnerId(ids[1])
	centerclient.SendCommonNotify2S([]string{ids[2]}, payRst, "PayResultNotify")
}

//玩家查询接口
func (self *PayService) QueryPayInfo(req *proto.QueryPayInfo, rst *proto.QueryPayInfoRst) error {
	logger.Info("QueryPayInfo called..........")
	self.sl.WaitLock(req.OpenId)
	defer self.sl.WaitUnLock(req.OpenId)

	rst.ItemId = ""
	itemId, err := common.Redis_getString(self.pCachePool, ITEM_TABLE, req.OpenId)
	if err != nil {
		logger.Error("QueryPayInfo common.Redis_getString error:", err, req.OpenId)
		return err
	}

	//如果没有充值成功的订单，则检查有没有失败的订单
	if itemId == "" {
		self.QueryOrder(req.OpenId)
		return nil
	}

	err = common.Redis_del(self.pCachePool, ITEM_TABLE, req.OpenId)
	if err != nil {
		logger.Error("QueryPayInfo common.Redis_del error:", err, req.OpenId)
		return err
	}

	rst.ItemId = itemId
	return nil
}

func (self *PayService) DeletePayInfo(req *proto.QueryPayInfo, rst *proto.CommonRst) error {
	logger.Info("DeletePayInfo called..........", req.OpenId)
	self.sl.WaitLock(req.OpenId)
	defer self.sl.WaitUnLock(req.OpenId)

	err := common.Redis_del(self.pCachePool, ITEM_TABLE, req.OpenId)
	if err != nil {
		logger.Error("DeletePayInfo common.Redis_del error:", err, req.OpenId)
		return err
	}
	return nil
}
