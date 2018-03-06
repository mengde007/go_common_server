package daerserver

import (
	"logger"
	"strconv"
)

//输出模式数据
func PrintPattern(pattern *DaerPattern) {
	if pattern == nil || pattern.cards == nil || len(pattern.cards) > 4 {
		logger.Info("pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 5)
	//logger.Info(content)
	content[0] += strconv.Itoa(int(pattern.value())) + ", "
	for i := 1; i < 5; i++ {
		if i-1 < len(pattern.cards) {
			content[i] += ConvertToWord(pattern.cards[i-1]) + ", "
		} else {
			content[i] += "    , "
		}
	}

	for i := 0; i < 5; i++ {
		logger.Info(content[i])
	}
}

func PrintPatternS(msgPrev string, pattern *DaerPattern) {
	if pattern == nil || pattern.cards == nil || len(pattern.cards) > 4 {
		logger.Info(msgPrev + "pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 5)
	//logger.Info(content)
	content[0] += strconv.Itoa(int(pattern.value())) + ", "
	for i := 1; i < 5; i++ {
		if i-1 < len(pattern.cards) {
			content[i] += ConvertToWord(pattern.cards[i-1]) + ", "
		} else {
			content[i] += "    , "
		}
	}

	for i := 0; i < 5; i++ {
		logger.Info(msgPrev + content[i])
	}
}

//输出模式列表
func PrintPatterns(patterns []*DaerPattern) {
	if patterns == nil {
		logger.Info(" pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 5)
	//logger.Info(content)
	for _, pattern := range patterns {
		content[0] += strconv.Itoa(int(pattern.value())) + ", "
		for i := 1; i < 5; i++ {
			if i-1 < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i-1]) + ", "
			} else {
				content[i] += "  , "
			}
		}
	}

	for i := 0; i < 5; i++ {
		logger.Info(content[i])
	}
}

func PrintPatternsS(msgPrev string, patterns []*DaerPattern) {
	if patterns == nil {
		logger.Info(msgPrev + " pattern牌错误。。。。。。。")
		return
	}

	content := make([]string, 5)
	//logger.Info(content)
	for _, pattern := range patterns {
		content[0] += strconv.Itoa(int(pattern.value())) + ", "
		for i := 1; i < 5; i++ {
			if i-1 < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i-1]) + ", "
			} else {
				content[i] += "  , "
			}
		}
	}

	for i := 0; i < 5; i++ {
		logger.Info(msgPrev + content[i])
	}
}

//输出卡牌列表
func PrintCards(cards []*DaerCard) {

	strCards := ""
	for _, card := range cards {
		strCards += ConvertToWord(card)
	}

	logger.Info(strCards)
}

func PrintCardsS(msgPrev string, cards []*DaerCard) {

	strCards := ""
	for _, card := range cards {
		strCards += ConvertToWord(card)
	}

	logger.Info(msgPrev + strCards)
}

//输出卡牌
func PrintCard(card *DaerCard) {
	if card == nil {
		logger.Info("None")
		return
	}
	logger.Info(ConvertToWord(card))
}

func PrintCardS(msgPrev string, card *DaerCard) {
	if card == nil {
		logger.Info("None")
		return
	}
	logger.Info(msgPrev + ConvertToWord(card))
}

//输出模式组数据
func PrintPatternGroup(patternGroup *DaerPatternGroup, isHu bool) {
	if patternGroup == nil || patternGroup.patterns == nil || len(patternGroup.patterns) > 7 {
		return
	}

	if isHu && (patternGroup.kaoCards == nil || len(patternGroup.kaoCards) <= 0 || patternGroup.huCards == nil || len(patternGroup.huCards) <= 0) {
		return
	}

	content := make([]string, 5)

	//模式
	for _, pattern := range patternGroup.patterns {
		content[0] += strconv.Itoa(int(pattern.value())) + "， "
		for i := 1; i < 5; i++ {
			if i-1 < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i-1]) + "， "
			} else {
				content[i] += "  " + "， "
			}
		}
	}

	//靠牌
	content[0] += "     ， "
	for i := 1; i < 5; i++ {
		if i-1 < len(patternGroup.kaoCards) {
			content[i] += ConvertToWord(patternGroup.kaoCards[i-1]) + "， "
		} else {
			content[i] += "  " + "， "
		}
	}

	//胡牌
	content[0] += strconv.Itoa(int(patternGroup.Value())) + "， "
	for i := 1; i < 5; i++ {
		if i-1 < len(patternGroup.huCards) {
			content[i] += ConvertToWord(patternGroup.huCards[i-1]) + "， "
		} else {
			content[i] += "  " + "， "
		}
	}

	for i := 0; i < 5; i++ {
		logger.Info(content[i])
	}
}

func PrintPatternGroupS(msgPrev string, patternGroup *DaerPatternGroup, isHu bool) {
	if patternGroup == nil || patternGroup.patterns == nil || len(patternGroup.patterns) > 7 {
		return
	}

	if isHu && (patternGroup.kaoCards == nil || len(patternGroup.kaoCards) <= 0 || patternGroup.huCards == nil || len(patternGroup.huCards) <= 0) {
		return
	}

	content := make([]string, 5)

	//模式
	for _, pattern := range patternGroup.patterns {
		content[0] += strconv.Itoa(int(pattern.value())) + "， "
		for i := 1; i < 5; i++ {
			if i-1 < len(pattern.cards) {
				content[i] += ConvertToWord(pattern.cards[i-1]) + "， "
			} else {
				content[i] += "  " + "， "
			}
		}
	}

	//靠牌
	content[0] += "     ， "
	for i := 1; i < 5; i++ {
		if i-1 < len(patternGroup.kaoCards) {
			content[i] += ConvertToWord(patternGroup.kaoCards[i-1]) + "， "
		} else {
			content[i] += "  " + "， "
		}
	}

	//胡牌
	content[0] += strconv.Itoa(int(patternGroup.Value())) + "， "
	for i := 1; i < 5; i++ {
		if i-1 < len(patternGroup.huCards) {
			content[i] += ConvertToWord(patternGroup.huCards[i-1]) + "， "
		} else {
			content[i] += "  " + "， "
		}
	}

	logger.Info(msgPrev)
	for i := 0; i < 5; i++ {
		logger.Info(content[i])
	}
}

//输出模式组列表
func PrintPatternGroups(patternGroups []*DaerPatternGroup, isHu bool) {
	logger.Info("Count:", len(patternGroups))
	for _, v := range patternGroups {
		PrintPatternGroup(v, isHu)
		//logger.Info(i, v)
	}
}

func PrintPatternGroupsS(msgPrev string, patternGroups []*DaerPatternGroup, isHu bool) {
	logger.Info("Count:", len(patternGroups))
	for _, v := range patternGroups {
		PrintPatternGroupS(msgPrev, v, isHu)
		//logger.Info(i, v)
	}
}

//打印玩家的牌
func PrintPlayer(player *DaerPlayer) {
	if player == nil {
		return
	}

	logger.Info("名字:", player.client.GetName())
	logger.Info("当前模式:", player.mode)
	logger.Info("等待执行的动作:", actionName[player.watingAction])
	logger.Info("准备执行的动作:", actionName[player.readyDoAction])
	logger.Info("手牌:")
	PrintCards(player.cards)
	PrintPatterns(player.fixedpatterns)
	logger.Info("显牌:")
	PrintPatterns(player.showPatterns)
	logger.Info("过牌:")
	PrintCards(player.showCards)
}

//打印桌面上牌的情况
func PrintRoom(room *DaerRoom) {
	if room == nil {
		return
	}

	logger.Info("==================桌面上牌的情况==================")
	logger.Info("房间的状态：")
	logger.Info(rootTypeName[room.state])
	logger.Info("活动玩家索引(%s),  名字：(%s)", room.activePlayerIndex, room.GetActivePlayer().client.GetName())
	logger.Info("活动牌：")
	PrintCard(room.activeCard)
	logger.Info("桌面上牌：")
	PrintCards(room.ownCards)
	//logger.Info("桌面上已开的牌：")
	//PrintCards(room.passCards)

	logger.Info("活动玩家信息0：")
	PrintPlayer(room.players[0])
	logger.Info("活动玩家信息1：")
	PrintPlayer(room.players[1])
	logger.Info("活动玩家信息2：")
	PrintPlayer(room.players[2])

	logger.Info("======================End======================")
}

//转换Card的名字
func ConvertToWord(card *DaerCard) string {
	if card == nil {
		return ""
	}

	if card.big {
		return bigWord[card.value]
	}

	return smallWord[card.value]
}

var bigWord = map[int32]string{
	1:  "壹",
	2:  "贰",
	3:  "叁",
	4:  "肆",
	5:  "伍",
	6:  "陆",
	7:  "柒",
	8:  "捌",
	9:  "玖",
	10: "拾",
}

var smallWord = map[int32]string{
	1:  "一",
	2:  "二",
	3:  "三",
	4:  "四",
	5:  "五",
	6:  "六",
	7:  "七",
	8:  "八",
	9:  "九",
	10: "十",
}

var actionName = map[int32]string{
	ANone:          "无",
	ATuoGuan:       "托管",
	ACancelTuoGuan: "取消托管",
	//AJieSuan:       "结算",
	AGuo:        "过",
	AChu:        "出",
	AMo:         "摸",
	AJin:        "进",
	AChi:        "吃",
	APeng:       "碰",
	AZhao:       "招",
	ALong:       "拢",
	ABaKuai:     "八块",
	AHu:         "胡",
	ABao:        "报",
	ASanLongBai: "三摆牌",
	ASiKanBai:   "四摆牌",
	AHeiBai:     "黑摆牌"}

var minTangName = map[int32]string{
	MTSanLongBai:    "三拢摆",
	MTSiKanBai:      "四坎摆",
	MTHeiBai:        "黑摆",
	MTLuanHu:        "乱胡",
	MTTianHu:        "天胡",
	MTDiHu:          "地胡",
	MTBaoPai:        "报牌",
	MTShuiShangPiao: "水上漂",
	MTHaiDiLao:      "海底捞",
	MTKun:           "坤",
	MTHongPai:       "红牌",
	MTHeiPai:        "黑牌",
	MTChaJiao:       "查叫",
	MTZhaTianBao:    "炸天报",
	MTShaBao:        "杀报",
	MTGui:           "归",
	MTZiMo:          "自摸",
	MTDianPao:       "点炮"}

var rootTypeName = map[int32]string{
	RSReady:                  "准备阶段",
	RSNotBankerBaiStage:      "非庄家摆牌阶段",
	RSBaoStage:               "报牌阶段",
	RSBankerJinPaiStage:      "庄家进阶段",
	RSBankerBaoStage:         "庄家包牌阶段",
	RSBankerChuPaiAfterStage: "庄家出牌后的阶段",
	RSLoopWorkStage:          "循环阶段",
	RSSettlement:             "结算阶段"}
