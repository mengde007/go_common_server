package majiangserver

import (
	"logger"
	//"strconv"
)

//输出模式数据
func PrintPattern(pattern *MaJiangPattern) {
	if pattern == nil || pattern.cards == nil || len(pattern.cards) > 4 {
		logger.Info("pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 4)
	for i := 0; i < 4; i++ {
		if i < len(pattern.cards) {
			content[i] += ConvertToWord(pattern.cards[i]) + ", "
		} else {
			content[i] += "   ， "
		}
	}

	for i := 0; i < 4; i++ {
		logger.Info(content[i])
	}
}

func PrintPatternS(msgPrev string, pattern *MaJiangPattern) {
	if pattern == nil || pattern.cards == nil || len(pattern.cards) > 4 {
		logger.Info(msgPrev + "pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 4)
	for i := 0; i < 4; i++ {
		if i < len(pattern.cards) {
			content[i] += ConvertToWord(pattern.cards[i]) + ", "
		} else {
			content[i] += "   ， "
		}
	}

	for i := 0; i < 4; i++ {
		logger.Info(msgPrev + content[i])
	}
}

//输出模式列表
func PrintPatterns(patterns []*MaJiangPattern) {
	if patterns == nil {
		logger.Info(" pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 4)
	//logger.Info(content)
	for _, pattern := range patterns {
		for i := 0; i < 4; i++ {
			if i < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i]) + ", "
			} else {
				content[i] += "   ， "
			}
		}
	}

	for i := 0; i < 4; i++ {
		logger.Info(content[i])
	}
}

func PrintPatternsS(msgPrev string, patterns []*MaJiangPattern) {
	if patterns == nil {
		logger.Info(msgPrev + " pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 4)
	//logger.Info(content)
	for _, pattern := range patterns {
		for i := 0; i < 4; i++ {
			if i < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i]) + ", "
			} else {
				content[i] += "   ， "
			}
		}
	}

	for i := 0; i < 4; i++ {
		logger.Info(msgPrev + content[i])
	}
}

//输出卡牌列表
func PrintCards(cards []*MaJiangCard) {

	strCards := ""
	for _, card := range cards {
		strCards += ConvertToWord(card)
	}

	logger.Info(strCards)
}

func PrintCardsS(msgPrev string, cards []*MaJiangCard) {

	strCards := ""
	for _, card := range cards {
		strCards += ConvertToWord(card) + ", "
	}

	logger.Info(msgPrev + strCards)
}

//输出卡牌
func PrintCard(card *MaJiangCard) {
	if card == nil {
		logger.Info("None")
		return
	}
	logger.Info(ConvertToWord(card))
}

func PrintCardS(msgPrev string, card *MaJiangCard) {
	if card == nil {
		logger.Info("None")
		return
	}
	logger.Info(msgPrev + ConvertToWord(card))
}

//输出模式组数据
func PrintPatternGroup(patternGroup *MaJiangPatternGroup, isHu bool) {
	if patternGroup == nil || patternGroup.patterns == nil {
		return
	}

	if isHu && (patternGroup.kaoCards == nil || len(patternGroup.kaoCards) <= 0 || patternGroup.huCards == nil || len(patternGroup.huCards) <= 0) {
		return
	}

	content := make([]string, 4)

	//模式
	for _, pattern := range patternGroup.patterns {
		for i := 0; i < 4; i++ {
			if i < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i]) + "， "
			} else {
				content[i] += "   ， "
			}
		}
	}

	//靠牌
	for i := 0; i < 4; i++ {
		if i < len(patternGroup.kaoCards) {
			content[i] += ConvertToWord(patternGroup.kaoCards[i]) + "， "
		} else {
			content[i] += "   ， "
		}
	}

	//胡牌
	for i := 0; i < 4; i++ {
		if i < len(patternGroup.huCards) {
			content[i] += ConvertToWord(patternGroup.huCards[i]) + "， "
		} else {
			content[i] += "   ， "
		}
	}

	for i := 0; i < 4; i++ {
		logger.Info(content[i])
	}
}

func PrintPatternGroupS(msgPrev string, patternGroup *MaJiangPatternGroup, isHu bool) {
	if patternGroup == nil || patternGroup.patterns == nil {
		return
	}

	if isHu && (patternGroup.kaoCards == nil || len(patternGroup.kaoCards) <= 0 || patternGroup.huCards == nil || len(patternGroup.huCards) <= 0) {
		return
	}

	content := make([]string, 4)

	//模式
	for _, pattern := range patternGroup.patterns {
		for i := 0; i < 4; i++ {
			if i < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i]) + "， "
			} else {
				content[i] += "   ， "
			}
		}
	}

	//靠牌
	for i := 0; i < 4; i++ {
		if i < len(patternGroup.kaoCards) {
			content[i] += ConvertToWord(patternGroup.kaoCards[i]) + "， "
		} else {
			content[i] += "   ， "
		}
	}

	//胡牌
	for i := 0; i < 4; i++ {
		if i < len(patternGroup.huCards) {
			content[i] += ConvertToWord(patternGroup.huCards[i]) + "， "
		} else {
			content[i] += "   ， "
		}
	}

	logger.Info(msgPrev)
	for i := 0; i < 4; i++ {
		logger.Info(content[i])
	}
}

//输出模式组列表
func PrintPatternGroups(patternGroups []*MaJiangPatternGroup, isHu bool) {
	logger.Info("Count:", len(patternGroups))
	for _, v := range patternGroups {
		PrintPatternGroup(v, isHu)
	}
}

func PrintPatternGroupsS(msgPrev string, patternGroups []*MaJiangPatternGroup, isHu bool) {
	logger.Info("Count:", len(patternGroups))
	for _, v := range patternGroups {
		PrintPatternGroupS(msgPrev, v, isHu)
	}
}

//打印玩家的牌
func PrintPlayer(player *MaJiangPlayer) {
	if player == nil {
		return
	}

	logger.Info("名字:", player.client.GetName())
	logger.Info("当前模式:", player.mode)
	logger.Info("等待执行的动作:", CnvtActsToStr(player.watingAction))
	logger.Info("准备执行的动作:", actionName[player.readyDoAction])
	logger.Info("手牌:")
	PrintCards(player.cards)
	logger.Info("显牌:")
	PrintPatterns(player.showPatterns)
}

//打印桌面上牌的情况
func PrintRoom(room *MaJiangRoom) {
	if room == nil {
		return
	}

	logger.Info("==================桌面上牌的情况==================")
	// logger.Info("房间的状态：")
	// logger.Info(roomTypeName[room.state])
	// logger.Info("活动玩家索引(%s),  名字：(%s)", room.activePlayerIndex, room.GetActivePlayer().client.GetName())
	// logger.Info("活动牌：")
	// PrintCard(room.activeCard)
	// logger.Info("桌面上牌：")
	// PrintCards(room.ownCards)
	// //logger.Info("桌面上已开的牌：")
	// //PrintCards(room.passCards)

	// logger.Info("玩家信息0：")
	// PrintPlayer(room.players[0])
	// logger.Info("玩家信息1：")
	// PrintPlayer(room.players[1])
	// logger.Info("玩家信息2：")
	// PrintPlayer(room.players[2])

	logger.Info("======================End======================")
}

//转换Card的名字
func ConvertToWord(card *MaJiangCard) (result string) {
	if card == nil {
		return ""
	}

	cType, value := card.CurValue()

	if card.cType == HongZhong {
		result = cValueWord[value] + cTypeWord[cType] + "|"
	} else {
		result = cValueWord[value] + cTypeWord[cType]
	}

	return
}

func CnvtActsToStr(actions []int32) (result string) {
	result = ""
	for _, a := range actions {
		result += actionName[a] + ", "
	}

	return
}

var cTypeWord = map[int32]string{
	UnknowCardType: "未知",
	Tiao:           "条",
	Tong:           "筒",
	Wan:            "万",
	HongZhong:      "中"}

var cValueWord = map[int32]string{
	0: "零",
	1: "一",
	2: "二",
	3: "三",
	4: "四",
	5: "五",
	6: "六",
	7: "七",
	8: "八",
	9: "九"}

var actionName = map[int32]string{
	ANone:          "无",
	AReady:         "准备",
	ACancelReady:   "取消准备",
	ATuoGuan:       "托管",
	ACancelTuoGuan: "取消托管",
	AAnGang:        "暗杠",
	AMingGang:      "明杠",
	ATieMingGang:   "贴鬼明杠",
	ABuGang:        "补杠",
	AGuo:           "过",
	AChu:           "出",
	AMo:            "摸",
	APeng:          "碰",
	ATiePeng:       "贴鬼碰",
	AHu:            "胡",
	ABao:           "报"}

var minTangName = map[int32]string{
	MTZiMo:          "自摸",
	MTGui:           "归",
	MTDaDuiZi:       "大对子",
	MTQingYiSe:      "清一色",
	MTNoneHongZhong: "无鬼",
	MTQiDui:         "七对",
	MTGangShangHua:  "杠上花",
	MTGangShangPao:  "杠上炮",
	MTQiangGang:     "抢杠",
	MTTianHu:        "天胡",
	MTBao:           "报牌",
	MTDingBao:       "顶报"}

var roomTypeName = map[int32]string{
	RSReady:                "准备阶段",
	RSBankerTianHuStage:    "庄家天胡阶段",
	RSNotBankerBaoPaiStage: "闲家报牌阶段",
	RSBankerChuPaiStage:    "庄家出牌阶段",
	RSBankerBaoPaiStage:    "庄家报牌阶段",
	RSLoopWorkStage:        "循环阶段",
	RSSettlement:           "结算阶段"}
