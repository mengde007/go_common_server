package common

import (
// "rpc"
// "strings"
)

// redis key TTL: 3 days
const (
	RedisKeyTTL = 3 * 24 * 3600
)

const (
	SystemTableName = "system"
	UserTableName   = "user"
)

const DbTableKeySplit = ":"

const (
	SystemKeyName_Mail           = "mail"
	SystemKeyName_Name2Id        = "playername2id"
	SystemKeyName_TTTScore       = "tttscore"
	SystemKeyName_Level2PlayerId = "level2playerid"
	//上次至尊联赛开启时间
	SystemKeyName_SuperLeague = "superleaguetime"

	//排行活动
	SystemKeyName_AccRank = "activityrank"

	//add for add challenge to redis
	ChallengeTableName = "challenge"
	ChallengeKeyName   = "playerChallengeScore"

	//玩家每周炫耀redis的哈希map
	PlayerTrophyMapName_Old = "playerOldTrophy"
	PlayerTrophyMapName_New = "playerNewTrophy"

	//openid与名字，通知
	TableOpenId2Name  = "tb_openid2name"
	TableServerNotice = "tb_servernotice"

	// 好友操作(添加/删除)
	PlayerFriendOperate = "playerfriendAddDel"
)

//特殊uid
const (
	SystemUid_FriendSpeedUp = "su-friendsu-0"
)

//加锁特殊服务器id
const (
	Special_Server_Id = 255
)

//加锁服务名
const (
	LockName_Player = "lockplayer"
	LockName_Donate = "lockdonate"
)

//加锁时间（秒）
const (
	LockTime_Login = 3 * 60
	LockTime_GM    = 5*60 - 2
)

//表名
const (
	TB_t_base_playerbase = "t_base_playerbase"
	TB_t_ext_playerextra = "t_ext_playerextra"
	// TB_t_ext_attacklog              = "t_ext_attacklog"
	// TB_t_ext_battlelog              = "t_ext_battlelog"
	TB_t_ext_friendexecise          = "t_ext_friendexecise"
	TB_t_ext_friendexeciselog       = "t_ext_friendexeciselog"
	TB_t_ext_playermail             = "t_ext_playermail"
	TB_t_ext_pve                    = "t_ext_pve"
	TB_t_ext_replay                 = "t_ext_replay"
	TB_t_ext_village                = "t_ext_village"
	TB_t_ext_normalchallenge        = "t_ext_normalchallenge"
	TB_t_ext_moneychallenge         = "t_ext_moneychallenge"
	TB_t_ext_playerchallengeinfo    = "t_ext_playerchallengeinfo"
	TB_t_ext_playeractivity         = "t_ext_playeractivity"
	TB_t_account_tencentid2playerid = "t_account_tencentid2playerid"
	TB_t_ext_battlelogid            = "t_ext_battlelogid" //玩家记录attack defence battle id
	TB_t_ext_village_id             = "t_ext_village_id"  // 单独记录玩家 villageid
)

func GetSystemTableKey_Mail() string {
	return SystemTableName + DbTableKeySplit + SystemKeyName_Mail
}

const Activity = "Activity"

//每日刷新时间
const CommonDayResetTime = 4 * 3600

//qq wx 类型
const (
	ServerType_QQ uint32 = 1
	ServerType_WX uint32 = 2
)

type MapActivityRank map[uint32]*StActivityRank

//排行榜活动
type StActivityRank struct {
	Id      uint32
	EndTime uint32
	Uids    []string
}
