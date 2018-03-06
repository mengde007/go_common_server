package connector

import (
	"common"
	"logger"
	"roleclient"
	"rpc"
	"strconv"
	"strings"
	// "time"
	"centerclient"
	"daerclient"
	"lockclient"
	"majiangclient"
	"pockerclient"
	"roomclient"
)

func (self *player) getFriendList() []string {
	if self.friendscache == nil {
		return nil
	}

	ret := make([]string, 0)
	for _, rp := range self.friendscache.Friends {
		if rp.GetUid() != self.GetUid() {
			ret = append(ret, rp.GetUid())
		}
	}

	return ret
}

//取得好友信息
func (self *player) getMyFriend(uid string) *rpc.Player {
	if self.GetUid() == uid {
		return nil
	}

	if self.friendscache == nil {
		return nil
	}

	for _, rp := range self.friendscache.Friends {
		if rp.GetUid() == uid {
			return rp
		}
	}
	return nil
}

//是否是我的好友
func (self *player) isMyFriend(uid string) bool {
	return self.getMyFriend(uid) != nil
}

// 添加好友
func (p *player) addFriend(newUid string) error {
	for _, v := range p.FriendUids {
		if v == newUid {
			logger.Error("addFriend, already has", newUid)
			return nil
		}
	}
	p.FriendUids = append(p.FriendUids, newUid)

	if p.friendscache == nil {
		return nil
	}

	var friend rpc.PlayerBaseInfo
	_, err := KVQueryBase(common.TB_t_base_playerbase, newUid, &friend)
	if err != nil {
		logger.Error("add friend can not load data, uid:%s, err:%s", newUid, err)
		return nil
	}

	var extra rpc.PlayerExtraInfo
	if exist, err := KVQueryExt(common.TB_t_ext_playerextra, newUid, &extra); err != nil || !exist {
		logger.Error("add friend can not load data, uid:%s, err:%s:", newUid, err)
		return nil
	}

	rp := rpc.Player{}
	rp.SetName(friend.GetName())
	rp.SetSex(friend.GetSex())
	rp.SetLevel(friend.GetLevel())
	rp.SetHeader(friend.GetHeader())
	rp.SetRoleId(friend.GetRoleId())
	rp.SetHeaderUrl(friend.GetHeaderUrl())
	rp.SetUid(friend.GetUid())
	rp.SetExp(friend.GetExp())
	rp.SetCoin(friend.GetCoin())
	rp.SetBOnline(false)
	if lockclient.IsOnline(friend.GetUid()) {
		rp.SetBOnline(true)
	}

	rp.Scores = extra.Scores

	p.friendscache.Friends = append(p.friendscache.Friends, &rp)

	return nil
}

// 添加好友请求缓存
func (p *player) addFriendRequest(uid string) error {
	if p.friendRequestCache == nil {
		p.friendRequestCache = &rpc.RequestFriendsList{}
	}
	if p.isMyFriend(uid) {
		return nil
	}
	for _, v := range p.friendRequestCache.Friends {
		if v.GetUid() == uid {
			logger.Error("addFriendRequest, already has", uid)
			return nil
		}
	}

	var friend rpc.PlayerBaseInfo
	_, err := KVQueryBase(common.TB_t_base_playerbase, uid, &friend)
	if err != nil {
		logger.Error("add friendrequest can not load data:", uid)
		return nil
	}

	var extra rpc.PlayerExtraInfo
	if exist, err := KVQueryExt(common.TB_t_ext_playerextra, uid, &extra); err != nil || !exist {
		logger.Error("add friendrequest can not load data, uid:%s, err:%s:", uid, err)
		return nil
	}

	rp := rpc.Player{}
	rp.SetName(friend.GetName())
	rp.SetSex(friend.GetSex())
	rp.SetLevel(friend.GetLevel())
	rp.SetHeader(friend.GetHeader())
	rp.SetRoleId(friend.GetRoleId())
	rp.SetHeaderUrl(friend.GetHeaderUrl())
	rp.SetUid(friend.GetUid())
	rp.SetExp(friend.GetExp())
	rp.SetCoin(friend.GetCoin())
	rp.Scores = extra.Scores

	p.friendRequestCache.Friends = append(p.friendRequestCache.Friends, &rp)
	return nil
}

// 删除好友
func (p *player) delFriend(delUid string) bool {
	bHas := false
	for i, v := range p.FriendUids {
		if v == delUid {
			bHas = true
			p.FriendUids = append(p.FriendUids[:i], p.FriendUids[i+1:]...)
			break
		}
	}
	if !bHas {
		return bHas
	}
	if p.friendscache == nil {
		return bHas
	}
	for i, v := range p.friendscache.Friends {
		if v.GetUid() == delUid {
			p.friendscache.Friends = append(p.friendscache.Friends[:i], p.friendscache.Friends[i+1:]...)
			break
		}
	}

	return bHas
}

// 查询好友添加删除信息
func (p *player) queryAddDelFriendInfo(sync bool) error {
	info, err := roleclient.QueryAddDelFriendInfo(p.GetUid())
	if err != nil {
		logger.Error("queryAddDelFriendInfo error", err, p.GetUid())
		return err
	}
	if info == nil {
		logger.Info("queryAddDelFriendInfo nil", p.GetUid())
		return nil
	}

	bNew := false     // 好友列表是否有更新
	bRequest := false // 是否有好友请求
	addUids := make([]string, 0)
	delUids := make([]string, 0)
	requestUids := make([]string, 0)
	logger.Info("===========info.AddList:", len(info.AddList))
	for _, v := range info.AddList {
		if v.BeAdd {
			if v.BeConfirm {
				// 确认添加
				addUids = append(addUids, v.Uid)
				bNew = true
				logger.Info("================v.BeConfirm")
			} else {
				// 好友请求
				requestUids = append(requestUids, v.Uid)
				bRequest = true
				logger.Info("================not BeConfirm")
			}
		} else {
			// 好友删除
			delUids = append(delUids, v.Uid)
			bNew = true

			logger.Info("================删除好友")
		}
	}

	for _, v := range addUids {
		p.addFriend(v)
	}
	for _, v := range delUids {
		p.delFriend(v)

		//nofity others
		msg := &rpc.DelFriendNofity{}
		msg.SetRst(int32(0))
		msg.SetId(p.GetUid())
		ids := []string{}
		ids = append(ids, p.GetUid())
		centerclient.SendCommonNotify2S(ids, msg, "DelFriendNofity")
	}

	for _, v := range requestUids {
		p.addFriendRequest(v)
	}

	if bRequest { // 有新的请求就同步
		logger.Info("==========有更新，同步 friendRequestCache")
		WriteResult(p.conn, p.friendRequestCache)
	}

	if sync && p.conn != nil {
		if bNew {
			WriteResult(p.conn, p.friendscache)
		}
	}
	return nil
}

func (p *player) IsFriend(uid string) bool {
	for _, v := range p.FriendUids {
		if v == uid {
			return true
		}
	}
	return false
}

func (p *player) OnlineNotice() {
	p.change_online_status(false)

	msg := &rpc.FriendStatusNofify{}
	msg.SetUid(p.GetUid())
	msg.SetBOnline(true)
	centerclient.SendCommonNotify2S(p.FriendUids, msg, "FriendStatusNofify")
}

func (p *player) change_online_status(offline bool) {
	msg := &rpc.FightRoomChatNotify{}
	msg.SetPlayerID(p.GetUid())
	msg.SetOffline(offline)

	switch p.GetGameType() {
	case "1":
		daerclient.ReqSendDeskChat(msg)
	case "2":
		majiangclient.ReqSendDeskChat(msg)
	case "3":
		pockerclient.ReqSendDeskChat(msg)
	case "4":
		roomclient.ReqSendDeskChat(msg)
	case "5":
		roomclient.ReqSendDeskChat(msg)
	case "6":
		pockerclient.ReqSendDeskChat(msg)
	default:
		logger.Error("未知的游戏类型")
	}
}

func (p *player) offlineNotice() {
	p.change_online_status(true)

	msg := &rpc.FriendStatusNofify{}
	msg.SetUid(p.GetUid())
	msg.SetBOnline(false)
	centerclient.SendCommonNotify2S(p.FriendUids, msg, "FriendStatusNofify")
}

func (p *player) playerGm(content string) bool {
	logger.Info("playerGm called:%s", content)
	if len(content) < 2 || content[:2] != "$$" {
		return true
	}

	content = strings.Trim(content[2:], " ")
	pos := strings.Index(content, " ")
	if pos == -1 {
		return false
	}

	cmd, _ := strings.ToLower(content[:pos]), content[pos+1:]
	// intarg, err := strconv.Atoi(args)

	switch cmd {
	//加钱
	case "item":
		{
			pa := strings.Split(content[pos+1:], " ")
			num, _ := strconv.Atoi(pa[1])

			if !CheckItemId(pa[0]) {
				logger.Error("道具Id不存在,id:%", pa[0])
				return false
			}
			p.AddCostCommon(pa[0], int32(num))
		}
		break
	case "add":
		{
			ids := GetAllItemIds()
			for _, id := range ids {
				p.AddCostCommon(id, int32(10))
			}
		}
		break
	}

	return false
}
