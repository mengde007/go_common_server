package daerserver

import (
	cmn "common"
	"logger"
)

func InitGlobalConfig() {
	logger.Info("init........config")

	//初始化全局变量--胡数
	cfg := cmn.GetDaerGlobalConfig("6")
	if cfg != nil {
		BigWordHuValue[0][PTPeng] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("7")
	if cfg != nil {
		SmallWordHuValue[0][PTPeng] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("8")
	if cfg != nil {
		BigWordHuValue[0][PTKan] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("9")
	if cfg != nil {
		SmallWordHuValue[0][PTKan] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("10")
	if cfg != nil {
		BigWordHuValue[0][PTLong] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("11")
	if cfg != nil {
		SmallWordHuValue[0][PTLong] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("12")
	if cfg != nil {
		BigWordHuValue[0][PTZhao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("13")
	if cfg != nil {
		SmallWordHuValue[0][PTZhao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("14")
	if cfg != nil {
		BigWordHuValue[1][PTPeng] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("15")
	if cfg != nil {
		SmallWordHuValue[1][PTPeng] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("16")
	if cfg != nil {
		BigWordHuValue[1][PTKan] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("17")
	if cfg != nil {
		SmallWordHuValue[1][PTKan] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("18")
	if cfg != nil {
		BigWordHuValue[1][PTLong] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("19")
	if cfg != nil {
		SmallWordHuValue[1][PTLong] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("20")
	if cfg != nil {
		BigWordHuValue[1][PTZhao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("21")
	if cfg != nil {
		SmallWordHuValue[1][PTZhao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("22")
	if cfg != nil {
		BigWordHuValue[0][PTEQSColumn] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("23")
	if cfg != nil {
		SmallWordHuValue[0][PTEQSColumn] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("24")
	if cfg != nil {
		BigWordHuValue[1][PTOneTwoThree] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("25")
	if cfg != nil {
		SmallWordHuValue[1][PTOneTwoThree] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//初始化全局变量--名堂
	cfg = cmn.GetDaerGlobalConfig("26")
	if cfg != nil {
		MinTangFanShu[MTLuanHu] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("27")
	if cfg != nil {
		MinTangFanShu[MTDiHu] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("28")
	if cfg != nil {
		MinTangFanShu[MTBaoPai] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("29")
	if cfg != nil {
		MinTangFanShu[MTShuiShangPiao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("30")
	if cfg != nil {
		MinTangFanShu[MTHaiDiLao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("31")
	if cfg != nil {
		MinTangFanShu[MTKun] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("32")
	if cfg != nil {
		MinTangFanShu[MTHongPai] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("33")
	if cfg != nil {
		MinTangFanShu[MTHeiPai] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("34")
	if cfg != nil {
		MinTangFanShu[MTChaJiao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("35")
	if cfg != nil {
		MinTangFanShu[MTZhaTianBao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("36")
	if cfg != nil {
		MinTangFanShu[MTShaBao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("37")
	if cfg != nil {
		MinTangFanShu[MTGui] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("38")
	if cfg != nil {
		MinTangFanShu[MTZiMo] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("39")
	if cfg != nil {
		MinTangFanShu[MTDianPao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	logger.Info("init........config....end")
}

//常规数量定义
const (
	FirstCardsAmount    = 20 //初始化手牌的数量
	ThreeLong           = 3  //三拢
	FourKan             = 4  //四砍
	MaxCanHu            = 10 //最大能胡牌的胡数
	MaxHu               = 90 //最大胡数
	RoomMaxPlayerAmount = 3  //房间的最大容量
	CardTotalAmount     = 80 //卡牌的总数量
	ErLongTouYi         = 2  //二拢偷一
)

//var DiFen int32 = 100        //地分

//结算时的Tag信息 0:无，1:自摸，2：点炮, 3:破产
const (
	JSNone = iota
	JSZiMo
	JSDianPao
	JSPoChan
)

//名堂
const (
	MTSanLongBai    = iota //三拢摆牌
	MTSiKanBai             //四坎摆拍
	MTHeiBai               //黑摆
	MTLuanHu               //乱胡
	MTTianHu               //天胡
	MTDiHu                 //地胡
	MTBaoPai               //报牌
	MTShuiShangPiao        //谁上漂
	MTHaiDiLao             //海底捞
	MTKun                  //坤
	MTHongPai              //红牌
	MTHeiPai               //黑牌
	MTChaJiao              //查叫
	MTZhaTianBao           //炸天报
	MTShaBao               //杀报
	MTGui                  //归
	MTZiMo                 //自摸
	MTDianPao              //点炮
)

//动作(Action)
const (
	ANone          = iota
	AReady         //准备
	ACancelReady   //取消准备
	ATuoGuan       //托管
	ACancelTuoGuan //取消托管
	AGuo
	AChu
	AMo  //进到手上的牌，此牌可以替换手牌的
	AJin //桌面上翻的一张牌或其他玩家出的一张牌
	AChiBi
	AChi
	APeng
	AZhao
	AZhongZhao
	ALong
	ABaKuai
	AHu
	ABao
	ASanLongBai
	ASiKanBai
	AHeiBai
	//ABai
	//AWaitingOtherPlayer //等待其他玩家操作
)

const (
	ACSuccess            = iota //成功
	ACAbandon                   //放弃执行
	ACWaitingOtherPlayer        //等待其他玩家操作
	AOccursError                //发送生了错误
)

const (
	PTUknown = iota
	PTSingle
	PTPair
	PTKan
	PTZhao
	PTAABColumn
	PTEQSColumn
	PTSZColumn
	PTPeng
	PTLong
	PTOneTwoThree
)

const (
	REFull = iota
)

//游戏进行的阶段
const (
	RSReady = iota
	RSBaoStage
	RSNotBankerBaiStage
	RSBankerJinPaiStage
	RSBankerBaoStage
	RSBankerBaiStage
	RSBankerChuPaiAfterStage
	RSLoopWorkStage
	RSSettlement
)

//第一排是红色，第二排是黑色，特殊：1，2，3算黑色
var BigWordHuValue = []map[uint]int32{
	{PTPeng: 9, PTKan: 12, PTLong: 18, PTZhao: 15, PTEQSColumn: 9},
	{PTPeng: 3, PTKan: 9, PTLong: 15, PTZhao: 12, PTOneTwoThree: 6}}

var SmallWordHuValue = []map[uint]int32{
	{PTPeng: 6, PTKan: 9, PTLong: 15, PTZhao: 12, PTEQSColumn: 6},
	{PTPeng: 1, PTKan: 6, PTLong: 12, PTZhao: 9, PTOneTwoThree: 3}}

//特殊的分数
var SpecificHuScore = map[uint]int32{
	MTLuanHu: 4,
	MTHeiBai: 16}

//分数
var HuScore = map[int32]int32{
	15: 2, 25: 3, 35: 4, 45: 5, 55: 6, 65: 7, 75: 8, 85: 9,
	10: 3, 20: 4, 30: 6, 40: 8, 50: 10, 60: 12, 70: 14, 80: 16, 90: 18}

//名堂的翻数
var MinTangFanShu = map[uint]int32{
	MTTianHu:        1, //天胡
	MTDiHu:          1, //地胡
	MTBaoPai:        1, //报牌
	MTShuiShangPiao: 1, //谁上漂
	MTHaiDiLao:      1, //海底捞
	MTKun:           1, //坤
	MTHongPai:       1, //红牌
	MTHeiPai:        3, //黑牌
	MTChaJiao:       1, //查叫
	MTZhaTianBao:    3, //炸天报
	MTShaBao:        1, //杀报
	MTGui:           1, //归
	MTZiMo:          1, //自摸
	MTDianPao:       1, //点炮
}
