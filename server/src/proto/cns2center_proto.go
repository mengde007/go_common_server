package proto

import (
// "rpc"
)

//大二战斗通知消息
type PushDaerMsg2Player struct {
	Func       string
	PlayerUids []string
	Value      []byte
}

//---------------------------------------old

//查询是否重名
type QueryName struct {
	Name   string
	Id     string
	BQuery bool //为true时表示根据名字查询返回值，false表示要插入名字与id的关联
}

type QueryNameResult struct {
	Success bool //返回值根据上面的BQuery不同
	Id      string
}

//cns请求center处理，center再分别发给每个cns
//登陆踢人
type LoginKickPlayer struct {
	Id string
}

type LoginKickPlayerResult struct {
	Success bool
}

//等级对应玩家id
type UpdaePlayerLevel2Id struct {
	Id    string
	Level uint32
}

type UpdaePlayerLevel2IdResult struct {
}

type RandomGetPlayerIdByLevel struct {
	Level uint32
}

type RandomGetPlayerIdByLevelResult struct {
	Id string
}

//玩家上周排名分享
type PlayerLastTrophy struct {
	Uids []string
}

type PlayerLastTrophyResult struct {
	M map[string]uint32
}

//在线人数
type GetOnlineNumber struct {
}

type GetOnlineNumberRst struct {
	Numbers uint32
}

//设置最大在线人数
type SetMaxOnlinePlayers struct {
	ServerId uint8
	Numbers  int32
}

type SetMaxOnlinePlayersRst struct {
	CurNumbers int32
}

//活动存储
type ActivityRankEnd struct {
	Id        uint32
	EndTime   uint32
	RankBegin int
	RankEnd   int
}

type ActivityRankEndRst struct {
}

//活动通知
type NotifyActivityRank struct {
	Buf []byte
}

type NotifyActivityRankRst struct {
}

//推送通知
type PushMsg struct {
	ID int
}

type PushMsgResult struct {
	ID int
}
