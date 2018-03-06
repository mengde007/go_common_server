package matchserver

//比赛类型
const (
	MTFree   = 1 //免费
	MTNaming = 2 //冠名
	MTCoin   = 3 //金币
	MTCredit = 4 //积分
)

//赛制类型
const (
	HitOut        = 1 //打立出
	ASSHitOut     = 2 //ASS打立出
	FixedCredit   = 3 //固定积分
	SwissTransfer = 4 //瑞士移位赛
)

//比赛开始模式
const (
	EverydayIntervalMode  = 1 //每日间隔赛 （每天定时开始，定时结束。每隔好久开一场）
	FixedTimeMode         = 2 //定时赛  （指定日期开赛，比如2017-1-10 22：00开始，且只有一场）
	EverydayFixedTimeMode = 3 //每日定时赛 （每天指定的时间都会开赛，比如20:00开始，且只有一场）
	FullStartMode         = 4 //满开赛 （满了指定人数就开）
)

//报名货币类型
const (
	EntranceCurrencyFree = 1 //免费
	EntranceCurrencyCoin = 2 //金币
	EntranceCurrencyGem  = 3 //砖石
)

//报名的错误码
const (
	EESuccess       = 0 //成功
	EENotExistID    = 1 //不存在的ID
	EENEnoughMoney  = 2 //没有足够的钱
	EEAlreadyEnroll = 3 //已经报名了
)

//退赛的错误码
const (
	WESuccess    = 0 //成功
	WENotExistID = 1 //不存在的ID
	WENotEnroll  = 2 //还没有报名
)
