package connector

import (
	// "accountclient"
	// "centerclient"
	"common"
	// "fmt"
	"logger"
	// "mailclient"
	// "proto"
	"centerclient"
	"lockclient"
	"roleclient"
	"rpc"
	// "strconv"
	"time"
)

const (
	FRIEND_OK = iota
	FRIEND_REPEAT_ADD
	FRIEND_ID_ERROR
)

func (self *CNServer) GetFriendsList(conn rpc.RpcConn, p *player) error {
	rps := &rpc.FriendsList{}
	if p.friendscache == nil {
		for _, uidInfo := range p.FriendUids {
			if uidInfo == "" {
				continue
			}
			logger.Info("**********friendUid:%s", uidInfo)
			var friend rpc.PlayerBaseInfo
			exist, err := KVQueryBase(common.TB_t_base_playerbase, uidInfo, &friend)
			if err != nil || !exist {
				continue
			}

			var extra rpc.PlayerExtraInfo
			if exist, err := KVQueryExt(common.TB_t_ext_playerextra, uidInfo, &extra); err != nil || !exist {
				logger.Error("GetFriendsList can not load data, uid:%s, err:%s:", uidInfo, err)
				continue
			}

			rp := rpc.Player{}
			rp.SetName(friend.GetName())
			rp.SetSex(friend.GetSex())
			rp.SetLevel(friend.GetLevel())
			rp.SetHeader(friend.GetHeader())
			rp.SetRoleId(friend.GetRoleId())
			rp.SetHeaderUrl(friend.GetHeaderUrl())
			rp.SetUid(friend.GetUid())
			rp.SetBOnline(false)
			if lockclient.IsOnline(uidInfo) {
				rp.SetBOnline(true)
			}
			rp.SetExp(friend.GetExp())
			rp.SetCoin(friend.GetCoin())
			rp.Scores = extra.Scores

			rps.Friends = append(rps.Friends, &rp)
		}
		p.friendscache = rps
	} else {
		rps = p.friendscache
	}
	WriteResult(conn, rps)
	return nil

}

func (self *CNServer) AddFriend(conn rpc.RpcConn, msg rpc.ReqString) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}
	rps := &rpc.AddFriendNofify{}

	id := msg.GetId()
	if id == p.GetUid() {
		logger.Error("不能加自己为好友")
		return nil
	}

	bFind := false
	for _, v := range p.FriendUids {
		if v == id {
			bFind = true
			break
		}
	}
	if bFind {
		rps.SetRst(int32(25))
		WriteResult(conn, rps)
		return nil
	}

	var friend rpc.PlayerBaseInfo
	_, err := KVQueryBase(common.TB_t_base_playerbase, id, &friend)
	if err != nil {
		rps.SetRst(int32(26))
		WriteResult(conn, rps)
		return nil
	}

	ok, err := roleclient.RequestAddFriend(p.GetUid(), id, true, false)
	if err != nil || !ok {
		rps.SetRst(int32(26))
		WriteResult(conn, rps)
		return nil
	}

	rps.SetRst(FRIEND_OK)
	WriteResult(conn, rps)
	return nil
}

func (self *CNServer) DelFriend(conn rpc.RpcConn, msg rpc.ReqString) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	rps := &rpc.DelFriendNofity{}
	rps.SetId(msg.GetId())

	bHas := p.delFriend(msg.GetId())
	if bHas {
		rps.SetRst(FRIEND_OK)
		// 删除好友
		roleclient.RequestAddFriend(p.GetUid(), msg.GetId(), false, false)
	} else {
		rps.SetRst(FRIEND_ID_ERROR)
	}

	WriteResult(conn, rps)
	self.GetFriendsList(conn, p)
	return nil
}

func (self *CNServer) ResponseAddFriend(conn rpc.RpcConn, info rpc.ReqResponseAddFriend) error {
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	if p.friendRequestCache == nil {
		return nil
	}
	bFind := false
	tmpCache := &rpc.RequestFriendsList{}
	for i, v := range p.friendRequestCache.Friends {
		if v.GetUid() == info.GetUid() {
			bFind = true
			continue
		}
		tmpCache.Friends = append(tmpCache.Friends, p.friendRequestCache.Friends[i])
	}
	if !bFind {
		return nil
	}
	p.friendRequestCache = tmpCache
	if info.GetBAccept() {
		p.addFriend(info.GetUid())
	}
	roleclient.ResponseAddFriend(p.GetUid(), info.GetUid(), info.GetBAccept())
	self.GetFriendsList(conn, p)
	return nil
}

func (self *CNServer) SearchPlayer(conn rpc.RpcConn, msg rpc.ReqInt) error {
	logger.Info("client call SearchPlayer begin")
	_, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	uid, err := roleclient.GetUidByRoleId(msg.GetId())
	if err != nil {
		logger.Error("SearchPlayer err, %s", err)
		return nil
	}

	var other rpc.PlayerBaseInfo
	exist, err = KVQueryBase(common.TB_t_base_playerbase, uid, &other)
	if err != nil || !exist {
		logger.Error("SearchPlayer err,%s", err)
		return nil
	}
	rp := &rpc.Player{}
	rp.SetName(other.GetName())
	rp.SetSex(other.GetSex())
	rp.SetLevel(other.GetLevel())
	rp.SetHeader(other.GetHeader())
	rp.SetRoleId(other.GetRoleId())
	rp.SetHeaderUrl(other.GetHeaderUrl())
	rp.SetUid(other.GetUid())

	sendMsg := &rpc.SearchFriendNofify{}
	sendMsg.SetPlayer(rp)
	WriteResult(conn, sendMsg)
	return nil
}

func (self *CNServer) SendFriendChat(conn rpc.RpcConn, msg rpc.SendFriendChat) error {
	logger.Info("client call SendFriendChat begin")
	p, exist := self.getPlayerByConnId(conn.GetId())
	if !exist {
		return nil
	}

	rcvUid := msg.GetReceiverUid()
	if !p.IsFriend(rcvUid) {
		logger.Error("SendFriendChat player:%s not friend", rcvUid)
		return nil
	}

	msg.SetSenderUid(p.GetUid())
	msg.SetSendtime(int32(time.Now().Unix()))
	if lockclient.IsOnline(rcvUid) {
		centerclient.SendCommonNotify2S([]string{rcvUid}, &msg, "SendFriendChat")
	} else {
		roleclient.SaveOfflineChatMsg(&msg)
	}

	return nil
}
