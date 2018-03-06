package gmserver

import (
	"chatclient"
	"common"
	"github.com/garyburd/redigo/redis"
	"logger"
	// "mailclient"
	// "encoding/json"
	// "fmt"
	// "io/ioutil"
	"net/http"
	"net/url"
	"proto"
	"rpc"
	"sync"
	"time"
	"timer"
)

type Notice struct {
	Id         int64
	CreateTime uint32
	// Platform     rpc.Login_Platform
	Priority     int
	Content      string
	LevelMin     uint32
	LevelMax     uint32
	TimeInterval uint32
	TotalTimes   uint32
	TimeBegin    uint32
	TimeEnd      uint32
	Msg          *rpc.BroadCastNotify
	//下面是统计
	SendLastTime uint32
	SendTimes    uint32
}

type AllNotice struct {
	M map[int64]*Notice
}

var giNoticeId int64 = 0

var pAllNotice *AllNotice
var lockNotice sync.RWMutex

func (self *GmService) startNoticeService() {
	//从数据库里面读取
	buf, err := common.Resis_getbuf(self.pCachePool,
		common.TableServerNotice,
		"")

	if err != nil && err != redis.ErrNil {
		logger.Fatal("startNoticeService failed", err)
		return
	}

	pAllNotice = &AllNotice{
		M: make(map[int64]*Notice),
	}

	if buf != nil {
		if err := common.GobDecode(buf, pAllNotice); err != nil {
			logger.Fatal("startNoticeService failed", err)
			return
		}
	}

	for id, _ := range pAllNotice.M {
		if id > giNoticeId {
			giNoticeId = id
		}
	}

	//开始tick
	tm := timer.NewTimer(time.Second)
	tm.Start(func() {
		timeNow := uint32(time.Now().Unix())

		sendid := int64(0)
		maxPriority := 0
		lockNotice.Lock()
		for id, notice := range pAllNotice.M {
			if timeNow >= notice.TimeBegin && timeNow <= notice.TimeEnd &&
				notice.SendTimes < notice.TotalTimes &&
				timeNow >= notice.SendLastTime+notice.TimeInterval {
				if notice.Priority >= maxPriority {
					sendid = id
					maxPriority = notice.Priority
				}
			}
		}

		if sendid != 0 {
			notice, _ := pAllNotice.M[sendid]
			err := chatclient.SendMsgBroadcastPlayer(
				notice.Msg,
				"BroadCastNotify",
				false)

			//成功了才改变
			if err == nil {
				notice.SendLastTime = timeNow
				notice.SendTimes++

				go self.saveNotice()
			}
		}

		lockNotice.Unlock()
	})
}

//添加轮播
func (self *GmService) handle_notice_add(w http.ResponseWriter, cmdHead *StHeadNew, data string) {
	logger.Info("handle_gm_notice_add called !")

	var st StReq_Notice
	if err := common.JsonDecode([]byte(data), &st); err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	if st.Body == nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "body")
		return
	}

	// plat := 0
	// if !checkPlatform(plat, true) {
	// 	writeError(w, cmdHead, err_code_wrong_param, "PlatId")
	// 	return
	// }

	priority := st.Body.Priority
	if priority < 0 {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "Priority")
		return
	}
	s_content := st.Body.Content
	if len(s_content) == 0 {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "Content empty")
		return
	}
	levelmin := 0
	// if levelmin < 0 {
	// 	writeError(w, cmdHead, err_code_wrong_param, "BeginLevel")
	// 	return
	// }

	levelmax := 999
	// if levelmax < levelmin {
	// 	writeError(w, cmdHead, err_code_wrong_param, "EndLevel")
	// 	return
	// }

	timeinterval := st.Body.Interval
	if timeinterval < 5 {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "Interval >= 5(s)")
		return
	}

	totaltimes := 999
	// if totaltimes <= 0 {
	// 	writeError(w, cmdHead, err_code_wrong_param, "Times")
	// 	return
	// }

	timebegin := st.Body.StartTime
	timeend := st.Body.EndTime
	if timebegin > timeend {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "StartTime >= EndTime ?")
		return
	}

	id, err := self.addNotice(priority, s_content, uint32(levelmin), uint32(levelmax),
		uint32(timeinterval), uint32(totaltimes),
		uint32(timebegin), uint32(timeend), nil)
	if err != nil {
		writeErrorNew(w, cmdHead, err_code_inner_error, err.Error())
		return
	}

	notice := &StRst_Notice_Body{}
	notice.NoticeId = int(id)

	rst := &StRst_Notice{}
	rst.Ec = 0
	rst.Em = "success"
	rst.Data = notice

	writeResult(w, rst)
}

func (self *GmService) AddPlayerNotice(req *proto.GmPlayerSend, rst *proto.GmUpdateOpenId2NameRst) (err error) {
	levelmin := 0
	levelmax := 999
	priority := 0

	timeinterval := common.GetDaerGlobalIntValue("42")
	totaltimes := common.GetDaerGlobalIntValue("43")
	timebegin := time.Now().Unix()
	timeend := time.Now().Unix() + 30*60

	msg := &rpc.ReqBroadCast{}
	if err := common.DecodeMessage(req.Msg, msg); err != nil {
		return err
	}

	id, err := self.addNotice(priority, "", uint32(levelmin), uint32(levelmax),
		uint32(timeinterval), uint32(totaltimes),
		uint32(timebegin), uint32(timeend), msg)
	if err != nil {
		logger.Error("player send notify failed playerName:%s, noticeId:%d", msg.GetPlayerName(), id)
		return err
	}
	return nil
}

func (self *GmService) handle_delete_notice(w http.ResponseWriter, cmdHead *StHeadNew, data string) {
	logger.Info("handle_delete_notice called !")

	var st StReq_Delete_Notice
	if err := common.JsonDecode([]byte(data), &st); err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	if st.Body == nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "body")
		return
	}

	// platform := 0
	// if !checkPlatform(platform, true) {
	// 	writeError(w, cmdHead, err_code_wrong_param, "platform")
	// 	return
	// }

	id := int64(st.Body.NoticeId)

	ok := self.delNotice(id)
	if !ok {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "NoticeId")
		return
	}

	rst := &StRst_RoleInfo{}
	rst.Ec = 0
	rst.Em = "success"
	rst.Data = []*StRoleInfo{}

	writeResult(w, rst)
}

func (self *GmService) addNotice(priority int, content string,
	levelmin, levelmax, timeinterval, totaltimes, timebegin, timeend uint32, msg *rpc.ReqBroadCast) (int64, error) {

	lockNotice.Lock()
	giNoticeId++
	id := giNoticeId

	broadMsg := &rpc.BroadCastNotify{}
	broadMsg.SetBroadCastID(int32(id))
	if msg == nil {
		broadMsg.SetSysBroad(true)
		broadMsg.SetContent(content)
	} else {
		broadMsg.SetSysBroad(false)
		broadMsg.SetContent(msg.GetContent())
		broadMsg.SetPlayerName(msg.GetPlayerName())
		broadMsg.SetPlayerID(msg.GetPlayerID())
		broadMsg.SetVip(msg.GetBVip())
	}

	pAllNotice.M[id] = &Notice{
		Id:           id,
		CreateTime:   uint32(time.Now().Unix()),
		Priority:     priority,
		Content:      content,
		LevelMin:     levelmin,
		LevelMax:     levelmax,
		TimeInterval: timeinterval,
		TotalTimes:   totaltimes,
		TimeBegin:    timebegin,
		TimeEnd:      timeend,
		//下面是统计
		SendLastTime: uint32(0),
		SendTimes:    uint32(0),
		Msg:          broadMsg,
	}
	lockNotice.Unlock()

	go self.saveNotice()

	return id, nil
}

func (self *GmService) queryNotice(timebegin, timeend uint32) ([]*Notice, error) {
	sRet := make([]*Notice, 0)

	lockNotice.RLock()
	for _, notice := range pAllNotice.M {
		if notice.TimeBegin >= timebegin &&
			(notice.TimeEnd <= timeend || timeend == uint32(0)) {
			sRet = append(sRet, notice)
		}
	}
	lockNotice.RUnlock()

	return sRet, nil
}

func (self *GmService) delNotice(id int64) bool {
	lockNotice.Lock()

	_, ok := pAllNotice.M[id]
	if ok {
		delete(pAllNotice.M, id)

		go self.saveNotice()
	}

	lockNotice.Unlock()

	return ok
}

func (self *GmService) saveNotice() {
	lockNotice.RLock()
	if buf, err := common.GobEncode(pAllNotice); err == nil {
		common.Resis_setbuf(self.pCachePool,
			common.TableServerNotice,
			"", buf)
	}
	lockNotice.RUnlock()
}

func (self *GmService) handle_gm_notice_add(w http.ResponseWriter, cmdHead *StHead, data string) {
	logger.Info("handle_gm_notice_add called !")

	var st StReq_Marquee
	if err := common.JsonDecode([]byte(data), &st); err != nil {
		writeError(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	if st.Body == nil {
		writeError(w, cmdHead, err_code_wrong_param, "body")
		return
	}

	// plat := ChangePlatId(st.Body.PlatId)
	// if !checkPlatform(plat, true) {
	// 	writeError(w, cmdHead, err_code_wrong_param, "PlatId")
	// 	return
	// }

	priority := st.Body.Priority
	if priority < 0 {
		writeError(w, cmdHead, err_code_wrong_param, "Priority")
		return
	}
	s_content := st.Body.NoticeContent
	if len(s_content) == 0 {
		writeError(w, cmdHead, err_code_wrong_param, "NoticeContent")
		return
	}
	levelmin := st.Body.BeginLevel
	if levelmin < 0 {
		writeError(w, cmdHead, err_code_wrong_param, "BeginLevel")
		return
	}

	levelmax := st.Body.EndLevel
	if levelmax < levelmin {
		writeError(w, cmdHead, err_code_wrong_param, "EndLevel")
		return
	}

	timeinterval := st.Body.Interval
	if timeinterval < 10 {
		writeError(w, cmdHead, err_code_wrong_param, "Interval >= 10(s)")
		return
	}

	totaltimes := st.Body.Times
	if totaltimes <= 0 {
		writeError(w, cmdHead, err_code_wrong_param, "Times")
		return
	}

	timebegin := st.Body.BeginTime
	timeend := st.Body.EndTime
	if timebegin > timeend {
		writeError(w, cmdHead, err_code_wrong_param, "BeginTime >= EndTime ?")
		return
	}

	id, err := self.addNotice(priority, s_content, uint32(levelmin), uint32(levelmax),
		uint32(timeinterval), uint32(totaltimes),
		uint32(timebegin), uint32(timeend), nil)
	if err != nil {
		writeError(w, cmdHead, err_code_inner_error, err.Error())
		return
	}

	stRet := &StRsp_Marquee{
		Head: cmdHead,
		Body: &StRsp_Marquee_Body{
			Result:   0,
			RetMsg:   "",
			NoticeId: id,
		},
	}
	stRet.Head.Cmdid++
	writeResult(w, stRet)
}

func (self *GmService) handle_gm_notice_del(w http.ResponseWriter, cmdHead *StHead, data string) {
	logger.Info("handle_gm_notice_del called !")

	var st StReq_NoticeDel
	if err := common.JsonDecode([]byte(data), &st); err != nil {
		writeError(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	if st.Body == nil {
		writeError(w, cmdHead, err_code_wrong_param, "body")
		return
	}

	// platform := ChangePlatId(st.Body.PlatId)
	// if !checkPlatform(platform, true) {
	// 	writeError(w, cmdHead, err_code_wrong_param, "platform")
	// 	return
	// }

	id := st.Body.NoticeId

	switch st.Body.Type {
	case Notice_Marquee:
		ok := self.delNotice(id)
		if !ok {
			writeError(w, cmdHead, err_code_wrong_param, "NoticeId")
			return
		}
	// case Notice_Login:
	// 	if err := mailclient.DelSysNotice(id); err != nil {
	// 		writeError(w, cmdHead, err_code_inner_error, err.Error())
	// 		return
	// 	}
	default:
		writeError(w, cmdHead, err_code_wrong_param, "Type")
		return
	}

	stRet := &StRsp_NoticeDel{
		Head: cmdHead,
		Body: &StRsp_NoticeDel_Body{
			Result: 0,
			RetMsg: "",
		},
	}

	stRet.Head.Cmdid++
	writeResult(w, stRet)
}

func (self *GmService) handle_gm_notice_query(w http.ResponseWriter, cmdHead *StHead, data string) {
	logger.Info("handle_gm_notice_query called !")

	var st StReq_NoticeQuery
	if err := common.JsonDecode([]byte(data), &st); err != nil {
		writeError(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	if st.Body == nil {
		writeError(w, cmdHead, err_code_wrong_param, "body")
		return
	}

	// platform := ChangePlatId(st.Body.PlatId)
	// if !checkPlatform(platform, true) {
	// 	writeError(w, cmdHead, err_code_wrong_param, "PlatId")
	// 	return
	// }

	timebegin := st.Body.BeginTime
	timeend := st.Body.EndTime
	if timebegin > timeend && timeend != int64(0) {
		writeError(w, cmdHead, err_code_wrong_param, "BeginTime > EndTime ?")
		return
	}

	notices, err := self.queryNotice(uint32(timebegin), uint32(timeend))
	if err != nil {
		writeError(w, cmdHead, err_code_inner_error, err.Error())
		return
	}

	stRet := &StRsp_NoticeQuery{
		Head: cmdHead,
		Body: &StRsp_NoticeQuery_Body{
			NoticeList: make([]*StRsp_NoticeQuery_Body_List, 0),
		},
	}

	// zoneId 即是小区ID
	// _, zoneId := common.GetPayUrlAndZoneId(false)
	//跑马灯
	for _, notice := range notices {
		stRet.Body.NoticeList = append(stRet.Body.NoticeList, &StRsp_NoticeQuery_Body_List{
			Type:          Notice_Marquee,
			NoticeId:      notice.Id,
			NoticeTitle:   "",
			NoticeContent: url.QueryEscape(notice.Content),
			SendTime:      int64(notice.CreateTime),
			// Partition:     zoneId,
		})
	}

	//弹窗公告
	// if err, snis := mailclient.QuerySysNotice(uint32(timebegin), uint32(timeend)); err == nil {
	// 	for _, sni := range snis {
	// 		stRet.Body.NoticeList = append(stRet.Body.NoticeList, &StRsp_NoticeQuery_Body_List{
	// 			Type:          Notice_Login,
	// 			NoticeId:      sni.Id,
	// 			NoticeTitle:   url.QueryEscape(sni.Title),
	// 			NoticeContent: url.QueryEscape(sni.Content),
	// 			SendTime:      int64(sni.CreateTime),
	// 		})
	// 	}
	// }

	stRet.Body.NoticeList_count = len(stRet.Body.NoticeList)

	stRet.Head.Cmdid++
	writeResult(w, stRet)
}
