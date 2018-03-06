package gmserver

import (
	// "accountclient"
	// "centerclient"
	"common"
	// "connector"
	// "dbclient"
	"errors"
	// "fmt"
	// "language"
	"logger"
	"mailclient"
	"net/http"
	"net/url"
	// "rpc"
	"roleclient"
	"strconv"
	"strings"
	// "time"
	"centerclient"
	"payclient"
)

func (self *GmService) handle_modify_role_info(w http.ResponseWriter, cmdHead *StHeadNew, data string) {
	logger.Info("handle_modify_role_info called !")

	var st StReq_Modify_RoleInfo
	if err := common.JsonDecode([]byte(data), &st); err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	if st.Body == nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "body is nil")
		return
	}

	logger.Info("handle_modify_role_info called, roleIds:%s", st.Body.Roleid)
	if len(st.Body.Roleid) == 0 {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "Roleid is empty")
		return
	}
	roleid, _ := strconv.Atoi(st.Body.Roleid)
	uid, err := roleclient.GetUidByRoleId(int32(roleid))
	if err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "can't find roleid")
		return
	}

	absolute := false
	resType := strings.ToLower(st.Body.Type)

	lid, err := self.lockPlayer(uid)
	if err != nil {
		logger.Error("handle_modify_role_info lock player failed, uid:%s", uid)
		return
	}
	defer self.unlockPlayer(uid, lid)

	switch resType {
	case "gem":
		{
			_, _, err = modify_gem(uid, absolute, st.Body.Number)
		}
	case "gold":
		{
			_, _, err = modify_gold(uid, absolute, st.Body.Number)
		}
	default:
		err = errors.New("wrong resource type")
	}

	if err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	rst := &StRst_RoleInfo{}
	rst.Ec = 0
	rst.Em = "success"
	rst.Data = []*StRoleInfo{}
	writeResult(w, rst)
}

//发送邮件
func (self *GmService) handle_send_mail(w http.ResponseWriter, cmdHead *StHeadNew, data string) {
	logger.Info("handle_send_mail called !")

	var st StReq_Send_Mail
	if err := common.JsonDecode([]byte(data), &st); err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	// plat := 0
	title := st.Body.Title
	content := st.Body.Content
	attachs := st.Body.Attach
	ids := strings.Split(st.Body.Ids, ",")
	sendType := st.Body.EmailType
	validtime := st.Body.Expire

	if len(title) == 0 && len(content) == 0 {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "title or content")
		return
	}

	//urldecode
	var err error
	if title, err = url.QueryUnescape(title); err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "title or content")
		return
	}

	if content, err = url.QueryUnescape(content); err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "title or content")
		return
	}

	logger.Info("attachs:", attachs)
	// if err := common.CheckMailAttach(attachs); err != nil {
	// 	writeErrorNew(w, cmdHead, err_code_wrong_param, "attach")
	// 	return
	// }

	//全服邮件
	if sendType == 0 && st.Body.Ids == "" {
		err, ok := mailclient.SendAllMail(title, content, attachs,
			uint32(0), uint32(validtime*3600))
		if err != nil {
			writeErrorNew(w, cmdHead, err_code_inner_error, err.Error())
			return
		}
		if !ok {
			writeErrorNew(w, cmdHead, err_code_inner_error)
			return
		}
	} else {
		//部分玩家邮件
		for _, id := range ids {
			roleid, _ := strconv.Atoi(id)
			uid, err := roleclient.GetUidByRoleId(int32(roleid))
			if err != nil {
				logger.Error("handle_send_mail roleclient.GetUidByRoleId err:%s, roleid%s", err, roleid)
				writeErrorNew(w, cmdHead, err_code_wrong_param, "roleid is empty string")
				continue
			}

			err = mailclient.SendSysMail2Player(uid, title, content, attachs, uint32(validtime*3600), false)
			if err != nil {
				logger.Error("handle_send_mail send mail err:", err)
				writeErrorNew(w, cmdHead, err_code_inner_error, err.Error())
				return
			}
		}
	}

	rst := &StRst_RoleInfo{}
	rst.Ec = 0
	rst.Em = "success"
	rst.Data = []*StRoleInfo{}
	writeResult(w, rst)
}

func (self *GmService) handle_accounting_info(w http.ResponseWriter, cmdHead *StHeadNew, data string) {
	logger.Info("handle_accounting_info called !")

	var st StReq_Accounting_info
	err := common.JsonDecode([]byte(data), &st)
	if err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	if st.Body == nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, "body is nil")
		return
	}

	resType := strings.ToLower(st.Body.Type)
	act := &StRst_Accounting_Body{}
	switch resType {
	case "online":
		{
			num, err := centerclient.GetOnlineNumbers()
			if err != nil {
				logger.Error("andle_accounting_info centerclient.GetOnlineNumbers() err:%s", err)
				writeErrorNew(w, cmdHead, err_code_wrong_param, "handle_accounting_info centerclient.GetOnlineNumbers() err")
			}
			act.Online = num
		}
	case "+@xx&july":
		{
			logger.Info("handle_accounting_info statistic called")
			value, err := payclient.GetRechargeStatistic()
			if err != nil {
				logger.Error("payclient.GetRechargeStatistic error")
				act.Online = 0
			} else {
				act.Online = value
			}
		}

	default:
		err = errors.New("wrong resource type")
	}

	if err != nil {
		writeErrorNew(w, cmdHead, err_code_wrong_param, err.Error())
		return
	}

	rst := &StRst_Accounting{}
	rst.Ec = 0
	rst.Em = "success"
	rst.Data = act
	writeResult(w, rst)
}
