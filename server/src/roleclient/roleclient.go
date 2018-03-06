package roleclient

import (
	"common"
	"errors"
	"logger"
	"proto"
	"rpc"
	"rpcplusclientpool"
)

var pPoll *rpcplusclientpool.ClientPool

//初始化加锁客户端
func init() {
	aServerHost := common.ReadServerClientConfig("roleserver")
	if len(aServerHost) == 0 {
		logger.Fatal("load config failed, no server")
		return
	}

	pPoll = rpcplusclientpool.CreateClientPool(aServerHost, "roleserver")
	if pPoll == nil {
		logger.Fatal("create failed")
		return
	}

	return
}

//进入房间
func ReqGenRolesId(uid string) (int32, int32, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return 0, 0, err
	}

	req := &proto.ReqGenRoleId{
		Uid: uid,
	}
	rst := &proto.RstGenRoleId{}
	err = conn.Call("RoleServer.GenRoleId", req, rst)
	if err != nil {
		return 0, 0, err
	}
	return rst.RoleId, rst.GuestId, nil
}

// 添加删除好友请求 bConfirmAdd 是否是确认添加
func RequestAddFriend(uid, ouid string, bAdd, bConfirmAdd bool) (bool, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return false, err
	}
	req := &proto.AddFriendRequest{
		MyUid:     uid,
		OtherUid:  ouid,
		BeAdd:     bAdd,
		BeConfirm: bConfirmAdd,
	}
	rst := &proto.AddFriendRequestRst{}
	if err := conn.Call("RoleServer.RequestAddFriend", req, rst); err != nil {
		logger.Error("RequestAddFriend error", uid, ouid, err)
		return false, err
	}
	return rst.Success, nil
}

func QueryAddDelFriendInfo(uid string) (*proto.OperateList, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return nil, err
	}
	req := &proto.FriendQueryPFBI{
		Uid: uid,
	}
	rst := &proto.FriendQueryPFBIRst{}
	if err := conn.Call("RoleServer.QueryAddDelFriendInfo", req, rst); err != nil {
		logger.Error("QueryAddDelFriendInfo error", uid, err)
		return nil, err
	}
	if rst.Value == nil {
		return nil, nil
	}

	info := &proto.OperateList{}
	err = common.GobDecode(rst.Value, info)
	if err != nil {
		logger.Error("QueryAddDelFriendInfo error", uid, err)
		return nil, err
	}
	return info, nil
}

func ResponseAddFriend(uid, ouid string, bAccept bool) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}
	req := &proto.ResponseAddFriend{
		MyUid:    uid,
		OtherUid: ouid,
		BeAccept: bAccept,
		BeAll:    false,
	}
	rst := &proto.ResponseAddFriendRst{}
	if err := conn.Call("RoleServer.ResponseAddFriend", req, rst); err != nil {
		logger.Error("ResponseAddFriend error", uid, ouid, bAccept, err)
		return err
	}
	return nil
}

func GetUidByRoleId(roleId int32) (string, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return "", err
	}
	req := &proto.ReqSearch{
		RoleId: roleId,
	}
	rst := &proto.SearchRst{}
	if err := conn.Call("RoleServer.GetUidByRoleId", req, rst); err != nil {
		logger.Error("ResponseAddFriend error", roleId)
		return "", err
	}

	if rst.Uid == "" {
		return "", errors.New("GetUidByRoleId return nil")
	}

	return rst.Uid, nil
}

func SaveOfflineChatMsg(msg *rpc.SendFriendChat) error {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return err
	}

	bufb, err := common.EncodeMessage(msg)
	if err != nil {
		return err
	}

	req := &proto.OfflineChatMsg{
		Uid:   msg.GetReceiverUid(),
		Value: bufb,
	}
	rst := &proto.CommonRst{}
	if err := conn.Call("RoleServer.SaveOfflineChatMsg", req, rst); err != nil {
		logger.Error("SaveOfflineChatMsg error", msg.GetSenderUid())
		return err
	}
	return nil
}

func GetOfflineChatMsg(uid string) (*rpc.OfflineMsgNofity, error) {
	err, conn := pPoll.RandomGetConn()
	if err != nil {
		return nil, err
	}

	req := &proto.ReqOfflineMsg{
		Uid: uid,
	}
	rst := &proto.OfflineMsgList{}
	if err := conn.Call("RoleServer.GetOfflineChatMsg", req, rst); err != nil {
		logger.Error("GetOfflineChatMsg error", uid)
		return nil, err
	}

	if len(rst.MsgLst) == 0 {
		return nil, nil
	}

	msg := &rpc.OfflineMsgNofity{}
	for _, v := range rst.MsgLst {
		ctMsg := &rpc.SendFriendChat{}
		if err := common.DecodeMessage(v.Value, ctMsg); err != nil {
			logger.Error("GetOfflineChatMsg err:%s, uid:%s", err, uid)
			continue
		}
		msg.ChatInfo = append(msg.ChatInfo, ctMsg)
	}

	return msg, nil
}
