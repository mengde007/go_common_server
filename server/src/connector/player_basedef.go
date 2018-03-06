// player_basedef
package connector

import (
	// "clanclient"
	"common"
	// "language"
	"logger"
	// "proto"
	"rpc"
	"sync"
	// "time"
	// "strconv"
	"timer"
)

const MaxLogNumber = 30
const MaxRepNumber = 4

type MobileQQInfo struct {
	Openid       string
	Openkey      string
	Pay_token    string
	Pf           string
	Pfkey        string
	Balance      uint32
	LoginChannel uint32
	PlatId       uint32
}

type player struct {
	*rpc.PlayerBaseInfo
	*rpc.PlayerExtraInfo
	lid                   uint64
	t                     *timer.Timer
	friendscache          *rpc.FriendsList        //好友缓存
	friendRequestCache    *rpc.RequestFriendsList // 好友请求缓存
	mobileqqinfo          *MobileQQInfo
	conn                  rpc.RpcConn
	bChanged              bool
	txOpenId              string
	Ip                    string
	bPVPShadow            bool //pvp攻击影子玩家，即有护盾玩家
	lastmatch             string
	lastWorldChatTime     uint32
	paylock               sync.Mutex
	uSaveTickCount        uint32
	iClientTimeDiff       int64
	uCurClientVersionCode int32
	whichGame             map[string]int32 //<K:1.大二，2.德州，3.鬼麻将,V:房间类型>
}

func LoadPlayer(uid, nickName string, lid uint64, roleId int32) (*player, bool) {
	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	var exist bool
	var err error
	if exist, err = KVQueryBase(common.TB_t_base_playerbase, uid, &base); err != nil {
		logger.Error("query PlayerBase failed!", err)
		return nil, false
	}

	//查询不到基础信息，说明是个全新玩家
	if !exist {
		return NewPlayer(uid, nickName, lid, roleId), true
	}

	//下面才是老玩家流程
	if exist, err = KVQueryExt(common.TB_t_ext_playerextra, uid, &extra); err != nil || !exist {
		logger.Error("query PlayerExtra failed!", err)
		return nil, false
	}

	// 完成基本成员变量组装
	ret := &player{
		lid:             lid,
		PlayerBaseInfo:  &base,
		PlayerExtraInfo: &extra,
		whichGame:       make(map[string]int32, 0),
	}

	return ret, false
}

func LoadOtherPlayer(uid string) *player {
	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo
	// var brlogid rpc.PlayerBatRepId

	if exist, err := KVQueryBase(common.TB_t_base_playerbase, uid, &base); err != nil || !exist {
		return nil
	}

	if exist, err := KVQueryExt(common.TB_t_ext_playerextra, uid, &extra); err != nil || !exist {
		return nil
	}

	ret := &player{
		PlayerBaseInfo:  &base,
		PlayerExtraInfo: &extra,
	}
	return ret
}

func NewPlayer(uid, nickName string, lid uint64, roleId int32) *player {
	logger.Info("NewPlayer begin", uid, lid)

	var base rpc.PlayerBaseInfo
	var extra rpc.PlayerExtraInfo

	base.SetUid(uid)
	// base.SetName(language.GetLanguage("TID_TOUR") + strconv.Itoa(int(roleId)))
	base.SetName(nickName)
	base.SetSex(int32(0))
	base.SetLevel(int32(1))
	base.SetExp(int32(0))
	base.SetVipLeftDay(int32(0))
	base.SetGem(int32(1000))
	base.SetLevel(1)
	base.SetRoleId(roleId)
	base.SetBModifyName(false)
	base.SetBModifySex(false)

	coin := common.GetDaerGlobalIntValue("50")
	base.SetCoin(int32(coin))

	logger.Info("============NewPlayer init coin:%d", coin)

	//这里新建玩家的活动数据
	ret := &player{
		lid:             lid,
		PlayerBaseInfo:  &base,
		PlayerExtraInfo: &extra,
	}

	//对于新玩家，修正数据标记为1，表示不需要再修正
	if result, err := KVWriteBase(common.TB_t_base_playerbase, uid, &base); err != nil || result == false {
		ret.LogError("NewPlayer base", result, err)
		return nil
	}

	if result, err := KVWriteExt(common.TB_t_ext_playerextra, uid, &extra); err != nil || result == false {
		ret.LogError("NewPlayer ext", result, err)
		return nil
	}

	return ret
}
