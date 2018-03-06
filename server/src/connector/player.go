package connector

import (
	"bytes"
	"centerclient"
	"common"
	"daerclient"
	"encoding/json"
	"io/ioutil"
	"lockclient"
	"logger"
	"mailclient"
	"net/http"
	"proto"
	"rankclient"
	"roomclient"
	"rpc"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"timer"
)

const (
	RES_COIN = "1"
	RES_GEM  = "2"
)

func (p *player) OnInit(conn rpc.RpcConn) {
	p.conn = conn
	p.OnTick(false)

	p.t = timer.NewTimer(time.Second)
	p.t.Start(
		func() {
			defer func() {
				p.conn.Unlock()

				if r := recover(); r != nil {
					logger.Error("player tick runtime error begin:%s", r)
					debug.PrintStack()

					cns.onDisConn(conn)
					conn.Close()
					logger.Error("player tick runtime error end:%s", r)
				}
			}()
			p.conn.Lock()
			p.OnTick(true)
		},
	)

	p.AcrossMothDo()
	p.InitTask()
	p.InitBankrupt()
	p.check_has_recharge()
	p.across_day_do()
	// p.AddResource("1", 20000)
}

func (p *player) across_day_do() {
	now := uint32(time.Now().Unix())
	if common.IsTheSameDay(now, uint32(p.GetLastLoginTime()), 0) {
		return
	}

	//跨天

	p.SetProfits(0)

	//检测Vip过期
	if p.GetVipLeftDay() > int32(0) && (now-uint32(p.GetVipOpenTime()))/uint32(3600*24) > uint32(p.GetVipLeftDay()) {
		p.SetVipLeftDay(int32(0))
		p.SetVipOpenTime(int32(0))
	}
}

func (p *player) OnNewPlayer() {
	p.AddItem2Bag("7", 500)
}

//CR 没有改变的情况就不存
func (p *player) OnQuit(bSelfLogin bool) {
	p.LogInfo("OnQuit Begin")
	if p.t != nil {
		p.t.Stop()
	}

	// if bSelfLogin { //只记录自己登陆的情况
	p.Save()
	// }

	p.offlineNotice()

	p.LogInfo("quit: Unlock begin")
	p.Unlock() // lockserver解锁
	p.LogInfo("OnQuit p.Unlock() end")
}

func (p *player) Unlock() (err error) {
	p.LogInfo("player Unlock : begin")
	if _, err = lockclient.ForceUnLock(common.LockName_Player, p.GetUid()); err != nil {
		p.LogError("player Unlock Error: %v", err)
	}
	p.LogInfo("player Unlock : end")
	return
}

func (p *player) Save() (err error) {
	p.uSaveTickCount = 0

	p.LogInfo("player Save data begin gameType:%s", p.GetGameType())
	//看key有没有过期
	// if !lockclient.IsLockValid(common.LockName_Player, p.GetUid(), p.lid) {
	// 	p.LogError("player Save error: Save lock invalid")
	// 	return
	// }

	//没有变化过
	// if !p.bChanged {
	// 	p.LogInfo("player Save : No Changed, passed")
	// 	return
	// }

	//保存玩家基本数据
	if _, err = KVWriteBase(common.TB_t_base_playerbase, p.GetUid(), p.PlayerBaseInfo); err != nil {
		p.LogError("player Save error: Save playerbase Error %v", err)
	}
	//保存玩家额外数据
	if _, err = KVWriteExt(common.TB_t_ext_playerextra, p.GetUid(), p.PlayerExtraInfo); err != nil {
		p.LogError("player Save error: Save playerextra Error %v", err)
	}

	p.LogInfo("player Save data end")
	return
}

func (p *player) OnTick(canSync bool) {
	// 登录的时候不去tick
	if canSync {
	}

	p.uSaveTickCount++
	if p.uSaveTickCount > 300 { // 5分钟存储一次, 这个tick是1秒一次
		p.Save()
		// if p.v != nil {
		// 	p.v.Save()
		// }
	}

	// 记录玩家登录时间
	// p.SetOnLineTime(p.GetOnLineTime() + int32(1))
}

//锁定自己
func (self *player) lockMyself() {
	if self.conn != nil {
		self.conn.Lock()
	}
}

func (self *player) unlockMyself() {
	if self.conn != nil {
		self.conn.Unlock()
	}
}

func (p *player) UpdateScore(bWin bool, name string) {
	for _, v := range p.Scores {
		if v.GetName() == name {
			if bWin {
				v.SetWin(v.GetWin() + 1)
			} else {
				v.SetLoss(v.GetLoss() + 1)
			}
		}
	}
	msg := &rpc.ScoreNofify{}
	msg.Scores = p.Scores
	WriteResult(p.conn, msg)
}

func (p *player) CheckUplevel() {
	exp := common.GetDaerGlobalIntValue("52")
	if p.GetVipLeftDay() > int32(0) {
		exp *= int32(2)
	}
	p.SetExp(p.GetExp() + exp)
	p.SetExpTotal(p.GetExpTotal() + exp)

	cfg := GetUplevelCfg(uint32(p.GetLevel() + 1))
	if cfg == nil {
		logger.Error("CheckUplevel GetUplevelCfg return nil, lv:%d", p.GetLevel()+1)
		return
	}

	if p.GetExp() >= cfg.Exp {
		p.SetLevel(p.GetLevel() + 1)
		p.SetExp(0)

		p.AddResource("1", cfg.Rewards)
		p.AddResource("1", cfg.ExtraRewards)
	}
	msg := &rpc.ResourceNotify{}
	msg.SetCoin(p.GetCoin())
	msg.SetGem(p.GetGem())
	msg.SetInsurCoin(p.GetInsurCoin())
	msg.SetLevel(p.GetLevel())
	msg.SetExp(p.GetExp())
	WriteResult(p.conn, msg)

	rankclient.UpdateRankingInfo(p.GetUid(), RANK_COIN, p.GetCoin()+p.GetInsurCoin())
	rankclient.UpdateRankingInfo(p.GetUid(), RANK_EXP, p.GetExpTotal())
}

func (p *player) AddCostCommon(rst string, value int32) {
	if rst == RES_COIN || rst == RES_GEM || rst == "coin" || rst == "gem" {
		if value > 0 {
			p.AddResource(rst, value)
			p.ResourceChangeNotify()
		} else {
			p.CostResource(rst, -value)
		}
		return
	}
	if value > 0 {
		p.AddItem2Bag(rst, value)
	} else {
		p.CostItem2Bag(rst, -value)
	}

}

func (self *player) CostResource(rst string, value int32) {
	if value <= 0 {
		logger.Error("CostResource value < 0, :%d", value)
		return
	}

	if rst == RES_COIN || rst == "coin" {
		curValue := self.GetCoin()
		if curValue < value {
			logger.Error("CostResource error coin:%d < value:%d", curValue, value)
			self.SetCoin(int32(0))
			return
		}
		self.SetCoin(self.GetCoin() - value)
	} else if rst == RES_GEM || rst == "gem" {
		curValue := self.GetGem()
		if curValue < value {
			logger.Error("CostResource error gem:%d < value:%d", curValue, value)
			return
		}
		self.SetGem(self.GetGem() - value)
	}

	msg := &rpc.ResourceNotify{}
	msg.SetCoin(self.GetCoin())
	msg.SetGem(self.GetGem())
	WriteResult(self.conn, msg)

	rankclient.UpdateRankingInfo(self.GetUid(), RANK_COIN, self.GetCoin()+self.GetInsurCoin())
}

// 1 = 匹配房间大贰 2 = 匹配房间麻将 3 = 匹配房间德州
// 4 = 自建房间大贰 5 = 自建房间麻将 6 = 自建房间德州
// 7 = 比赛房间大贰 8 = 比赛房间麻将 9 = 比赛房间德州
func (self *player) ReconnectEnterRoom() {
	bFind := false
	k := self.GetGameType()
	logger.Info("*********ReconnectEnterRoom has been called k:%s", k)
	msg := &rpc.PlayerInRoomNotify{}
	if k == "1" {
		bInRoom, err := daerclient.PlayerInDaerGame(self.GetUid())
		if err != nil {
			logger.Error(" daerclient.PlayerInDaerGame return error, uid:%s err:%s", self.GetUid(), err)
			return
		}
		if bInRoom {
			msg.SetGameType(k)
			msg.SetRoomType(self.GetRoomType())
			logger.Info("ReconnectEnterRoom GameType:daer roomType:%d", self.GetRoomType())
			bFind = true
		}
	} else if k == "4" || k == "5" || k == "6" {
		bInRoom, err := roomclient.PlayerInRoom(self.GetUid())
		if err != nil {
			logger.Error("roomclient.PlayerInDaerGame return error, uid:%s err:%s", self.GetUid(), err)
			return
		}
		if bInRoom {
			msg.SetGameType(k)
			msg.SetRoomType(self.GetRoomType())
			logger.Info("ReconnectEnterRoom GameType:daer roomType:%d", self.GetRoomType())
			bFind = true
		}
	} else if k == "poker" {

	} else if k == "mj" {

	}

	if !bFind {
		self.SetGameType("")
		self.SetRoomType(int32(0))

		msg.SetGameType("none")
		msg.SetRoomType(int32(0))
		logger.Info("ReconnectEnterRoom GameType:daer roomType:%d", self.GetRoomType())
	}

	WriteResult(self.conn, msg)

	// if len(self.whichGame) > 1 {
	// 	logger.Error("ReconnectEnterRoom error,len (self.whichGame) must <=1, curValue:%d", len(self.whichGame))
	// 	return
	// }

	// for k, v := range self.whichGame {
	// 	logger.Info("ReconnectEnterRoom k:%s, v:%d", k, v)
	// 	if k == "daer" {
	// 		bInRoom := daerclient.PlayerInDaerGame(self.GetUid())
	// 		if bInRoom {
	// 			msg := &rpc.PlayerInRoomNotify{}
	// 			msg.SetGameType("daer")
	// 			msg.SetRoomType(v)
	// 			WriteResult(self.conn, msg)
	// 			logger.Info("ReconnectEnterRoom GameType:daer roomType:%d", v)
	// 		}
	// 	} else if k == "poker" {

	// 	} else if k == "mj" {

	// 	}

	// }
}

//超时连接
func createTransport() *http.Transport {
	return common.CreateTransport()
}

type StHeadNew1 struct {
	Serverid string `json:"serverid"` //服务器列表
	Commid   int    `json:"commid"`   //命令Id
}

//修改角色信息
type StReq_Modify_RoleInfo1 struct {
	Head *StHeadNew1                  `json:"head"`
	Body *StReq_Modify_RoleInfo_Body1 `json:"body"`
}
type StReq_Modify_RoleInfo_Body1 struct {
	Roleid string `json:"roleid"` //角色Id
	Type   string `json:"type"`   //修改信息类型
	Number int    `json:"number"` //正值表示增加，负值表示减少
}

//发送邮件
type StReq_Send_Mail1 struct {
	Head *StHeadNew1            `json:"head"`
	Body *StReq_Send_Mail_Body1 `json:"body"`
}
type StReq_Send_Mail_Body1 struct {
	Title     string `json:"title"`      //邮件标题
	Content   string `json:"content"`    //邮件内容
	Expire    int    `json:"expire"`     //邮件过期时间，单位天
	UserType  string `json:"user_type"`  //陌陌ID/角色ID/昵称
	Ids       string `json:"ids"`        //ids用“,”分割
	Attach    string `json:"attach"`     //附件
	EmailType int    `json:"email_type"` //0全服筛选/1指定ID
}

//创建轮播
type StReq_Notice1 struct {
	Head *StHeadNew1         `json:"head"`
	Body *StReq_Notice_Body1 `json:"body"`
}

type StReq_Notice_Body1 struct {
	StartTime int    `json:"start_time"` //轮播开始时间
	EndTime   int    `json:"end_time"`   //轮播结束时间
	Channel   string `json:"channel"`    //渠道以逗号隔开
	Interval  int    `json:"interval"`   //时间间隔,不小于5秒
	Content   string `json:"content"`    //轮播内容,小于30汉字
	Priority  int    `json:"priority"`   //优先级
}

//玩家下线
func (p *player) GmTester() {
	hd := &StHeadNew1{
		Commid:   5002,
		Serverid: "27",
	}

	bd := &StReq_Modify_RoleInfo_Body1{
		Type:   "gold",
		Roleid: p.GetUid(),
		Number: 1000,
	}
	heroInfo := &StReq_Modify_RoleInfo1{
		Head: hd,
		Body: bd,
	}

	//当前时间
	body, err := json.Marshal(heroInfo)
	if err != nil {
		logger.Error("GmTester Marshal  error: %v", err)
		return
	}

	logger.Info("************body:%v", string(body))

	buf := bytes.NewBuffer(body)

	//url
	fullurl := "http://127.0.0.1:9100"
	logger.Info("GmTester tx url: %v", fullurl)

	client := &http.Client{
		Transport: createTransport(),
	}

	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
	if err != nil {
		logger.Error("GmTester http.Post error: %v", err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("GmTester body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("GmTester ioutil.ReadAll error: %v", err)
		return
	}
}

//玩家下线
func (p *player) GmTester2() {
	hd := &StHeadNew1{
		Commid:   5005,
		Serverid: "27",
	}

	bd := &StReq_Send_Mail_Body1{
		Title:     "欢迎体验游戏",
		Content:   "新爱的玩家，欢迎参加本轮CB1测试，目前游戏还不够守完善，可能存在Bug。希望您提供宝贵意见，祝您游戏愉快！！",
		Expire:    2,
		UserType:  "",
		Ids:       "",
		Attach:    "1:1000,2:100",
		EmailType: 0,
	}
	heroInfo := &StReq_Send_Mail1{
		Head: hd,
		Body: bd,
	}

	//当前时间
	body, err := json.Marshal(heroInfo)
	if err != nil {
		logger.Error("GmTester2 Marshal  error: %v", err)
		return
	}

	logger.Info("************body:%v", string(body))

	buf := bytes.NewBuffer(body)

	//url
	fullurl := "http://127.0.0.1:9100"
	logger.Info("GmTester2 tx url: %v", fullurl)

	client := &http.Client{
		Transport: createTransport(),
	}

	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
	if err != nil {
		logger.Error("GmTester2 http.Post error: %v", err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("GmTester2 body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("GmTester2 ioutil.ReadAll error: %v", err)
		return
	}
}

//轮播
func (p *player) GmTester3() {
	hd := &StHeadNew1{
		Commid:   5008,
		Serverid: "27",
	}

	now := int(time.Now().Unix())
	bd := &StReq_Notice_Body1{
		Interval:  10,
		Content:   "跑巴灯走起哈",
		Priority:  0,
		StartTime: now,
		EndTime:   now + 1000,
	}
	heroInfo := &StReq_Notice1{
		Head: hd,
		Body: bd,
	}

	//当前时间
	body, err := json.Marshal(heroInfo)
	if err != nil {
		logger.Error("GmTester3 Marshal  error: %v", err)
		return
	}

	logger.Info("************body:%v", string(body))

	buf := bytes.NewBuffer(body)

	//url
	fullurl := "http://127.0.0.1:9100"
	logger.Info("GmTester3 tx url: %v", fullurl)

	client := &http.Client{
		Transport: createTransport(),
	}

	res, err := client.Post(fullurl, "application/x-www-form-urlencoded", buf)
	if err != nil {
		logger.Error("GmTester3 http.Post error: %v", err)
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	logger.Info("GmTester3 body info:%s", string(b))
	res.Body.Close()
	if err != nil {
		logger.Error("GmTester3 ioutil.ReadAll error: %v", err)
		return
	}
}

//把钱存在保险箱
func (p *player) SaveMoney(value int32) bool {
	if value <= int32(0) {
		logger.Error("SaveMoney err, value <= 0")
		return false
	}

	IntValue := common.GetDaerGlobalIntValue("41")
	if p.GetCoin()-IntValue < value {
		logger.Error("SaveMoney2InsurCoin not enough coin:p.GetCoin():%d, cfg.IntValue:%d, value:%d",
			p.GetCoin(), IntValue, value)
		return false
	}

	p.SetCoin(p.GetCoin() - value)
	p.SetInsurCoin(p.GetInsurCoin() + value)
	p.ResourceChangeNotify()

	// p.GmTester()
	// p.GmTester2()
	return true
}

//从保险箱取钱
func (p *player) Withdraw(value int32) bool {
	if value <= int32(0) {
		logger.Error("Withdraw err, value <= 0")
		return false
	}

	if p.GetInsurCoin() < value {
		logger.Error("Withdraw not enough money!")
		return false
	}

	p.SetInsurCoin(p.GetInsurCoin() - value)
	p.SetCoin(p.GetCoin() + value)
	p.ResourceChangeNotify()
	// p.GmTester3()
	return true
}

func (p *player) ResourceChangeNotify() {
	msg := &rpc.ResourceNotify{}
	msg.SetCoin(p.GetCoin())
	msg.SetGem(p.GetGem())
	msg.SetInsurCoin(p.GetInsurCoin())
	WriteResult(p.conn, msg)

	rankclient.UpdateRankingInfo(p.GetUid(), RANK_COIN, p.GetCoin()+p.GetInsurCoin())
}

//领取附件
func (p *player) GetAttach(mailId string) {
	attach, err := mailclient.GetMailAttach(p.GetUid(), mailId)
	if err != nil || attach == "" {
		logger.Error("GetAttach err:%s", err)
		return
	}

	attachs := strings.Split(attach, ",")
	if len(attach) == 0 {
		logger.Error("GetAttach err, attachs:%s", attachs)
		return
	}

	for _, id := range attachs {
		idNum := strings.Split(id, ":")
		if len(idNum) != 2 {
			logger.Error("GetAttach err, len(idNum) != 2 ,attach:%s", attachs)
			continue
		}
		num, _ := strconv.Atoi(idNum[1])
		p.AddResource(idNum[0], int32(num))
	}
	p.ResourceChangeNotify()
	return
}

func (p *player) AddResource(itemId string, num int32) {
	if num < 0 {
		logger.Error("AddResource num < 0, num:%d", num)
		return
	}

	if itemId == "1" || itemId == "coin" {
		p.SetCoin(p.GetCoin() + num)
	} else if itemId == "2" || itemId == "gem" {
		p.SetGem(p.GetGem() + num)
	}
}

func (p *player) AcrossMothDo() {
	sig := p.GetSign()
	if sig == nil {
		sig = &rpc.Signature{}
		p.SetSign(sig)
	}

	_, m, _ := time.Now().Date()
	if sig.GetMonth() != int32(m) {
		sig.SetMonth(int32(m))
		sig.Signs = []int32{}
		sig.SetLastSign(int32(0))
		sig.SetContiDay(int32(0))
		sig.SetContiRewardTms(int32(0))
	}
}

func (p *player) Signature() bool {
	_, _, d := time.Now().Date()
	signs := p.GetSign().Signs
	for _, v := range signs {
		if v == int32(d) {
			logger.Error("Signature error, this day:%d has signed", d)
			p.fixSignature()
			WriteResult(p.conn, p.GetSign())
			return false
		}
	}
	p.GetSign().Signs = append(p.GetSign().Signs, int32(d))

	rewards := common.GetDaerGlobalIntValue("45")
	if p.GetVipLeftDay() > 0 {
		rewards *= int32(2)
	}
	p.SetCoin(p.GetCoin() + rewards)

	p.checkGiveContiRewards()
	p.GetSign().SetLastSign(int32(d))
	p.ResourceChangeNotify()
	WriteResult(p.conn, p.GetSign())
	return true
}

func (p *player) fixSignature() {
	max := int32(0)
	for _, day := range p.GetSign().Signs {
		if max < day {
			max = day
		}
	}

	if p.GetSign().GetLastSign() < max {
		p.GetSign().SetLastSign(max)
	}
}

func (p *player) checkGiveContiRewards() {
	newContiDays := p.calContinueDays()
	if newContiDays > p.GetSign().GetContiDay() {
		p.GetSign().SetContiDay(newContiDays)
	}

	y, m, _ := time.Now().Date()
	contiDays := p.GetSign().GetContiDay()
	contiRewardTms := p.GetSign().GetContiRewardTms()
	if contiDays == int32(7) && contiRewardTms == 0 {
		rewards := common.GetDaerGlobalIntValue("46")
		if p.GetVipLeftDay() > 0 {
			rewards *= int32(2)
		}
		p.SetCoin(p.GetCoin() + rewards)
		p.GetSign().SetContiRewardTms(contiRewardTms + 1)
	} else if contiDays == int32(15) && contiRewardTms == 1 {
		rewards := common.GetDaerGlobalIntValue("47")
		if p.GetVipLeftDay() > 0 {
			rewards *= int32(2)
		}
		p.SetCoin(p.GetCoin() + rewards)
		p.GetSign().SetContiRewardTms(contiRewardTms + 1)
	} else if contiDays == int32(21) && contiRewardTms == 2 {
		rewards := common.GetDaerGlobalIntValue("48")
		if p.GetVipLeftDay() > 0 {
			rewards *= int32(2)
		}
		p.SetCoin(p.GetCoin() + rewards)
		p.GetSign().SetContiRewardTms(contiRewardTms + 1)
	} else if contiDays == common.DaysOfMonth(int(y), int(m)) && contiRewardTms == 3 {
		rewards := common.GetDaerGlobalIntValue("49")
		if p.GetVipLeftDay() > 0 {
			rewards *= int32(2)
		}
		p.SetCoin(p.GetCoin() + rewards)
		p.GetSign().SetContiRewardTms(contiRewardTms + 1)
	}
}

func (p *player) calContinueDays() int32 {
	common.BubbleSortExtra(p.GetSign().Signs)
	maxDays := int32(1)
	curDays := int32(1)
	for index, day := range p.GetSign().Signs {
		if len(p.GetSign().Signs) >= index+2 && day+1 == p.GetSign().Signs[index+1] {
			curDays += 1
		} else if curDays > maxDays {
			maxDays = curDays
			curDays = 1
		}
	}
	return maxDays
}

func (p *player) SignatureBefore(day int32) bool {
	_, _, d := time.Now().Date()
	if day >= int32(d) || day <= 0 {
		logger.Error("SignatureBefore 参数错误，day:%d d:%d", day, d)
		return false
	}

	if p.GetItemNum("3") <= int32(0) {
		logger.Error("SignatureBefore 没有补签道具")
		return false
	}

	bFind := false
	for _, v := range p.GetSign().Signs {
		if v == day {
			bFind = true
			break
		}
	}
	if bFind {
		logger.Error("SignatureBefore 这天已经签到过了, day:%d", day)
		return false
	}

	p.GetSign().Signs = append(p.GetSign().Signs, int32(day))
	rewards := common.GetDaerGlobalIntValue("45")
	if p.GetVipLeftDay() > 0 {
		rewards *= int32(2)
	}
	p.SetCoin(p.GetCoin() + rewards)

	p.checkGiveContiRewards()
	p.ResourceChangeNotify()
	p.CostItem2Bag("3", int32(1))

	logger.Error("SignatureBefore ok")
	WriteResult(p.conn, p.GetSign())
	return false
}

func (p *player) InitBankrupt() {
	rupt := p.GetBankrupt()
	if rupt == nil {
		rupt := &rpc.BankruptInfo{}
		p.SetBankrupt(rupt)
		return
	}

	now := time.Now().Unix()
	if !common.IsTheSameDay(uint32(rupt.GetTime()), uint32(now), 0) {
		rupt.SetRewardTimes(0)
		rupt.SetTime(0)
	}
}

func (p *player) BankruptRewards() {
	lowCoin := common.GetDaerGlobalIntValue("106")
	if p.GetCoin()+p.GetInsurCoin() > lowCoin {
		logger.Error("没破产，不能领取，当前金币:%d", p.GetCoin()+p.GetInsurCoin())
		return
	}

	rupt := p.GetBankrupt()
	strRewards := common.GetGlobalStringValue("105")
	if strRewards == "" {
		logger.Error("全局表的奖励没配，id:105")
		return
	}

	arrRewards := strings.Split(strRewards, "|")
	if int32(len(arrRewards)) < rupt.GetRewardTimes()+1 {
		logger.Error("今天已经领完了，客户端检测出错了, times:%d", rupt.GetRewardTimes())
		return
	}

	idTime := arrRewards[rupt.GetRewardTimes()]
	rewards := strings.Split(idTime, "_")
	if len(rewards) != 3 {
		logger.Error("奖励配错了，en(rewards) != 3，rewards:%s", idTime)
		return
	}

	now := time.Now().Unix()
	if rupt.GetRewardTimes() != 0 && p.GetVipLeftDay() <= int32(0) {
		interval, _ := strconv.Atoi(rewards[2])
		if int32(now)-rupt.GetTime() < int32(interval*60) {
			logger.Error("CD还没到，不能领")
			return
		}
	}

	coin, _ := strconv.Atoi(rewards[1])
	p.AddResource("1", int32(coin))
	rupt.SetTime(int32(now))
	rupt.SetRewardTimes(rupt.GetRewardTimes() + 1)
	p.ResourceChangeNotify()
}

func (p *player) Billing(req *proto.ReqCostRes) {
	p.AddCostCommon(req.ResName, req.ResValue)

	win := false
	if req.ResValue > 0 {
		win = true
	}
	gameType := SIG_PLAY_DAER
	if req.GameType == "mj" {
		gameType = SIG_PLAY_MJ
	}

	p.UpdateScore(win, req.GameType)
	p.TaskTrigger(gameType, win)
	p.CheckUplevel()

	//赚金数量
	if req.ResName == RES_COIN || req.ResName == "coin" {
		p.SetProfits(p.GetProfits() + req.ResValue)
		if p.GetProfits() > int32(0) {
			rankclient.UpdateRankingInfo(p.GetUid(), RANK_PROFIT, p.GetProfits())
		}

	}
}

func (p *player) checkBilling() {
	rst, err := centerclient.CheckCostFromCache(p.GetUid())
	if err != nil {
		logger.Error("checkBilling centerclient.CheckCostFromCache err:%s", err)
		return
	}
	if len(rst.PlayerList) == 0 {
		return
	}

	if rst.PlayerList[0] != p.GetUid() {
		logger.Error("checkBilling rst.PlayerList[0]:%s != p.GetUid():%s", rst.PlayerList[0], p.GetUid())
		return
	}

	p.Billing(rst)
}
