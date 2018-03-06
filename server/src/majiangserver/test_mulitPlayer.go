package majiangserver

// import (
// 	cmn "common"
// 	"fmt"
// 	"strconv"
// )

// var mulPlayers [6]*DaerPlayer

// func init() {
// 	daerRoomMgr = &DaerRoomMgr{}
// 	daerRoomMgr.init()
// 	fmt.Println("初始化包")

// 	//初始化6个玩家
// 	mulPlayers = [6]*DaerPlayer{}
// 	for i, _ := range mulPlayers {
// 		mulPlayers[i] = NewPlayer(strconv.Itoa(i), nil)
// 	}
// }

// //测试房间的进入和离开
// func TestPlayerEnterLeaveRoom() {

// 	daerRoomMgr.EnterGame(cmn.RTDaerHight, mulPlayers[0])
// 	daerRoomMgr.EnterGame(cmn.RTDaerHight, mulPlayers[1])

// 	fmt.Println("房间数量(1)：", daerRoomMgr.GetRoomAmount(cmn.RTDaerHight))

// 	daerRoomMgr.EnterGame(cmn.RTDaerHight, mulPlayers[2])

// 	fmt.Println("房间数量(1)：", daerRoomMgr.GetRoomAmount(cmn.RTDaerHight))

// 	daerRoomMgr.EnterGame(cmn.RTDaerHight, mulPlayers[3])

// 	fmt.Println("房间数量(2)：", daerRoomMgr.GetRoomAmount(cmn.RTDaerHight))

// 	//	mulPlayers[0].room.Leave(mulPlayers[0])
// 	//	curAmount := daerRoomMgr.rooms[cmn.RTDaerHight][0].GetPlayerAmount()
// 	//	fmt.Println("第一房间的玩家数量(2)：", curAmount)

// 	//开始测试发牌
// 	player0 := mulPlayers[0]
// 	player1 := mulPlayers[1]
// 	player2 := mulPlayers[2]

// 	player0.isReady = true
// 	player1.isReady = true
// 	player2.isReady = true

// 	cards := player0.room.Shuffle()
// 	if cards == nil {
// 		fmt.Println("TestPlayerEnterLeaveRoom:cards is nil.")
// 		return
// 	}

// 	player0.room.Licensing(cards)

// 	fmt.Println("第一个玩家的手牌状态：==================")
// 	PrintPatterns(player0.showPatterns)
// 	PrintPatterns(player0.fixedpatterns)
// 	PrintCards(player0.cards)

// 	fmt.Println("第二个玩家的手牌状态：==================")
// 	PrintPatterns(player1.showPatterns)
// 	PrintPatterns(player1.fixedpatterns)
// 	PrintCards(player1.cards)

// 	fmt.Println("第三个玩家的手牌状态：==================")
// 	PrintPatterns(player2.showPatterns)
// 	PrintPatterns(player2.fixedpatterns)
// 	PrintCards(player2.cards)

// 	fmt.Println("桌面上牌的情况：==================")
// 	PrintCards(player0.room.ownCards)
// 	PrintCards(player0.room.showCards)

// 	fmt.Println("================发牌阶段结束=====================")
// }
