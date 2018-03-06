package majiangserver

// import (
// 	cmn "common"
// 	"fmt"
// )

// //测试胡牌控制器
// func TestHuController() {

// 	//初始化牌
// 	room := NewRoom(cmn.RTDaerHight)

// 	//玩家进入房间
// 	testPlayer := NewPlayer("0", nil)
// 	room.Enter(testPlayer)

// 	//洗牌
// 	outOfOrderCards := room.Shuffle()

// 	//发牌1
// 	testPlayer.cards = []*DaerCard{
// 		room.cards[4], room.cards[14], room.cards[5], room.cards[15], room.cards[6],
// 		room.cards[16], room.cards[7], room.cards[42], room.cards[50], room.cards[1],
// 		room.cards[2], room.cards[12], room.cards[3], room.cards[13], room.cards[17],
// 		room.cards[8], room.cards[18], room.cards[9], room.cards[10], room.cards[47]}

// 	//发牌2
// 	//	testPlayer.cards = []*DaerCard{
// 	//		room.cards[3], room.cards[4], room.cards[5], room.cards[6], room.cards[6],
// 	//		room.cards[7], room.cards[8]}

// 	fmt.Println("发牌")
// 	room.Licensing(outOfOrderCards)

// 	//计算胡牌
// 	testPlayer.controller.CheckHu()
// 	//fmt.Println(testPlayer.controller.CheckHu())

// 	//输出所有的胡牌组合
// 	PrintPatternGroups(testPlayer.controller.huController.patternGroups, true)
// }
