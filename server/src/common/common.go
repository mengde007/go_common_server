package common

import (
	"rpc"
	"time"
)

//控制模式
const (
	Auto = iota
	Manual
)

//房间类型
const (
	RTDaerLow   = 1 //大贰低倍场
	RTDaerHight = 2 //大贰高倍场
)

const (
	CUnknown            = 1 << iota //未知
	CPositive                       //正面
	CBack                           //背面
	CLock                           //锁定出牌
	CChi                            //吃牌
	CHu                             //胡牌
	CLockHongZhongValue             //锁定鬼牌的替换值
	CTianHu                         //天胡
	CZiMoHu                         //自摸胡
	CDianPaoHu                      //点炮胡
	CGangShangHu                    //杠上花
	CGangShangPaoHu                 //杠上炮
)

const (
	PTBanker = iota
	PTNormal
)

const (
	UnknownGame = 0 //未知游戏
	DaerGame    = 1 //大贰游戏
	MaJiang     = 2 //麻将
	DeZhouPuker = 3 //德州扑克
)

const (
	PiPeiFang  = 1 //匹配房
	ZiJianFang = 2 //自建房
	BiSaiFang  = 3 //比赛房
)

//进入房间错误码定义
const (
	ERNone = iota
	ERLessCoin
	ERReachUpLimit
)

//战斗状态定义
const (
	FSReady = iota
	FSFighting
	FSSettlement
)

//接口
type GameRoom interface {
	ID() int32    //房间的ID(房间在配置表中的类型ID)
	UID() int32   //房间唯一标示
	Name() string //房间名字
	//Enter(player Player)                                     //进入房间
	// ReEnter(playerID string, playerInfo *rpc.PlayerBaseInfo) //从新进入房间
	// Leave(uid string, isChangeDesk bool) bool                //离开房间
	//Shuffle() []Card                                         //洗牌
	//Licensing(cards []Card)                                  //发牌
	DecideBanker()                      //定庄
	ChangeActivePlayerTo(player Player) //改变活动玩家
	GetPlayerAmount() int32             //获取房间人数
	IsInRoom(uid string) bool           //是否在房间
	IsGaming() bool                     //是否在游戏中
	IsFull() bool                       //房间是否满员
	IsEmpty() bool                      //房间是否为空
	GetDifen() int32                    //获取底注
	GetIsDaiGui() bool                  //是否带归
	GetMaxMultiple() int32              //倍数上限
	GetRakeRate() int32                 //抽成比率
	GetTiYongAmount() int32             //替用数量
	GetQiHuKeAmount() int32             //起胡颗数
	// OnPlayerDoAction(msg *rpc.ActionREQ)               //执行客服端发上来的消息
	OnJieSanRoom(uid string, msg *rpc.JieSanRoomREQ)   //解散房间的接口
	StartGame()                                        //开始游戏
	SendCommonMsg2Others(msg *rpc.FightRoomChatNotify) //向战斗中的玩家发消息
	GetGameType() int32                                //获取房间里的游戏类型
	GetRcvThreadHandle() *chan RoomMsgQueue            //获取线程接受数据chan
	GetExitThreadHandle() *chan bool                   //获取线程退出chan

	RoomTimerUpdate
	RoomMsg
	RoomSelector
}

type RoomTimerUpdate interface {
	UpdateTimer(ft time.Duration)
}

type RoomMsg interface {
	SendEnterRoomACK(p Player)
	SendLeaveRoomACK(p Player, isChangeDesk bool)
	//SendJieSuanACK(uid string, jieSuanPlayer Player, sysType int32, huangZhuang stageEnd, success bool) (result []*rpc.JieSuanCoin)
}

type RoomSelector interface {
	OnStartGameAfter()                                       //开始游戏后
	Enter(player Player)                                     //进入房间
	ReEnter(playerID string, playerInfo *rpc.PlayerBaseInfo) //从新进入房间
	Leave(uid string, isChangeDesk bool) bool                //离开房间
	OnPlayerDoAction(msg *rpc.ActionREQ)                     //执行客服端发上来的消息
	DoJieSuan()
	ForceAllPlayerLeave()
}

type Player interface {
	ID() string //获取ID
	GetPlayerBasicInfo() *rpc.PlayerBaseInfo
	//Compose(cards []*Card)       //组牌
	SwitchControllMode(mode int) //切换手动或自动模式
	//ObtainCard(card *Card)       //获得牌（进牌）

}

type Card interface {
	IsEqual(c *Card) bool
	IsIncomeCard() bool
	IsLock() bool
}

//type Pattern interface {
//}

//type Controller interface {
//}

const (
	CustomRoomCardID = 7
	KickCardID       = 11
)

//游戏类型对应的字符串
var GameTypeName = map[int32]string{
	DaerGame:    "daer",
	MaJiang:     "majiang",
	DeZhouPuker: "dezhoupuker",
}

//main thread -> room thread msg
type RoomMsgQueue struct {
	Msg  interface{}
	Msg2 interface{}
	Func string
}
