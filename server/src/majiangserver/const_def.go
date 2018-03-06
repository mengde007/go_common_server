package majiangserver

import (
	cmn "common"
	"logger"
)

func InitGlobalConfig() {
	logger.Info("init........config")

	//初始化全局变量--翻对应的颗数
	cfg := cmn.GetDaerGlobalConfig("551")
	if cfg != nil {
		KeAmount[0] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}
	cfg = cmn.GetDaerGlobalConfig("552")
	if cfg != nil {
		KeAmount[1] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("553")
	if cfg != nil {
		KeAmount[2] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}
	cfg = cmn.GetDaerGlobalConfig("554")
	if cfg != nil {
		KeAmount[3] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}
	cfg = cmn.GetDaerGlobalConfig("555")
	if cfg != nil {
		KeAmount[4] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}
	cfg = cmn.GetDaerGlobalConfig("556")
	if cfg != nil {
		KeAmount[5] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	//初始化全局变量--名堂
	cfg = cmn.GetDaerGlobalConfig("561")
	if cfg != nil {
		MinTangFanShu[MTZiMo] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("562")
	if cfg != nil {
		MinTangFanShu[MTGui] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("563")
	if cfg != nil {
		MinTangFanShu[MTDaDuiZi] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("564")
	if cfg != nil {
		MinTangFanShu[MTQingYiSe] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("565")
	if cfg != nil {
		MinTangFanShu[MTNoneHongZhong] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("566")
	if cfg != nil {
		MinTangFanShu[MTQiDui] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("567")
	if cfg != nil {
		MinTangFanShu[MTGangShangHua] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("568")
	if cfg != nil {
		MinTangFanShu[MTGangShangPao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("569")
	if cfg != nil {
		MinTangFanShu[MTQiangGang] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("570")
	if cfg != nil {
		MinTangFanShu[MTTianHu] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("571")
	if cfg != nil {
		MinTangFanShu[MTBao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	cfg = cmn.GetDaerGlobalConfig("572")
	if cfg != nil {
		MinTangFanShu[MTDingBao] = cfg.IntValue
	} else {
		logger.Error("GetDaerGlobalConfig return nil")
	}

	logger.Info("init........config....end")
}

//常规数量定义
const (
	FirstCardsAmount    = 13 //初始化手牌的数量
	RoomMaxPlayerAmount = 4  //房间的最大容量
)

//结算时的Tag信息 0:无，1:自摸，2：点炮, 3:破产
const (
	JSNone = iota
	JSZiMo
	JSDianPao
	JSPoChan
)

//名堂
const (
	MTZiMo          = iota //自摸
	MTGui                  //归
	MTDaDuiZi              //大对子
	MTQingYiSe             //清一色
	MTNoneHongZhong        //无鬼
	MTQiDui                //七对
	MTGangShangHua         //杠上花
	MTGangShangPao         //杠上炮
	MTQiangGang            //抢杠
	MTTianHu               //天胡
	MTBao                  //报牌
	MTDingBao              //顶报
)

//动作(Action)
const (
	ANone          = iota
	AReady         //准备
	ACancelReady   //取消准备
	ATuoGuan       //托管
	ACancelTuoGuan //取消托管
	AGuo           //过
	AChu           //出
	AMo            //摸
	APeng          //碰
	ATiePeng       //贴鬼碰
	AAnGang        //暗杠
	AMingGang      //明杠
	ATieMingGang   //贴鬼明杠
	ABuGang        //补杠
	AHu            //胡
	ABao           //报
)

//动作的结果
const (
	ACSuccess            = iota //成功
	ACAbandon                   //放弃执行
	ACWaitingOtherPlayer        //等待其他玩家操作
	AOccursError                //发送生了错误
)

//模式类型
const (
	PTUknown = iota
	PTSingle
	PTPair
	PTKan
	PTGang
	PTAnGang
	PTSZ
)

const (
	REFull = iota
)

//游戏进行的阶段
const (
	RSReady = iota
	RSBankerTianHuStage
	RSNotBankerBaoPaiStage
	RSBankerChuPaiStage
	RSBankerBaoPaiStage
	RSLoopWorkStage
	RSSettlement
)

//番数对应的颗数
//名堂的翻数
var KeAmount = []int32{2, 5, 10, 20, 40, 80, 160, 320, 640, 1280, 2560, 5120, 10240, 20480, 40960, 81920, 163840, 327680, 655360}

//名堂的翻数
var MinTangFanShu = map[uint]int32{
	MTZiMo:          1, //自摸
	MTGui:           1, //归
	MTDaDuiZi:       2, //大对子
	MTQingYiSe:      2, //清一色
	MTNoneHongZhong: 3, //无鬼
	MTQiDui:         2, //七对
	MTGangShangHua:  1, //杠上花
	MTGangShangPao:  3, //杠上炮
	MTQiangGang:     3, //抢杠
	MTTianHu:        3, //天胡
	MTBao:           2, //报牌
	MTDingBao:       1, //顶报
}
