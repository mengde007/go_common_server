package daerserver

// import (
// 	cmn "common"
// 	"fmt"
// 	"strconv"
// )

// var players [6]*DaerPlayer

// func init() {
// 	daerRoomMgr = &DaerRoomMgr{}
// 	daerRoomMgr.init()
// 	fmt.Println("初始化包")

// 	//初始化6个玩家
// 	players = [6]*DaerPlayer{}
// 	for i, _ := range players {
// 		players[i] = NewPlayer(strconv.Itoa(i), nil)
// 	}

// 	daerRoomMgr.EnterGame(cmn.RTDaerHight, players[0])
// 	daerRoomMgr.EnterGame(cmn.RTDaerHight, players[1])
// 	daerRoomMgr.EnterGame(cmn.RTDaerHight, players[2])
// }

// //测试房间
// func TestBasic(iter func()) {

// 	//开始测试发牌
// 	player0 := players[0]
// 	player1 := players[1]
// 	player2 := players[2]

// 	player0.room.ResetRoom()

// 	player0.isReady = true
// 	player1.isReady = true
// 	player2.isReady = true

// 	iter()

// 	PrintPlayer(player0)

// 	fmt.Println("================发牌阶段结束=====================")
// }

// //测试控制器中的函数（CheckLong, Long...）
// func TestControllerFunc() {

// 	//测试报
// 	//TestBao()
// 	//测试摆
// 	//TestBai()
// 	//测试胡
// 	//测试招
// 	//TestZhao()
// 	//测试碰
// 	//TestPeng()
// 	//测试吃
// 	TestChi()
// 	//测试查叫
// 	//测试结算

// }

// //测试报
// func TestBao() {
// 	player0 := players[0]

// 	//case 1
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]
// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[44-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]
// 		cards[9] = room.cards[12-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]
// 		cards[12] = room.cards[42-1]
// 		cards[13] = room.cards[47-1]
// 		cards[14] = room.cards[50-1]
// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[46-1]
// 		cards[18] = room.cards[8-1]
// 		cards[19] = room.cards[18-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)

// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 2
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]
// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[44-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]
// 		cards[9] = room.cards[12-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]
// 		cards[12] = room.cards[42-1]
// 		cards[13] = room.cards[47-1]
// 		cards[14] = room.cards[50-1]
// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[46-1]
// 		cards[18] = room.cards[9-1]
// 		cards[19] = room.cards[20-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 3
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]
// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[44-1]
// 		cards[6] = room.cards[9-1]
// 		cards[7] = room.cards[19-1]
// 		cards[8] = room.cards[29-1]
// 		cards[9] = room.cards[12-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]
// 		cards[12] = room.cards[42-1]
// 		cards[13] = room.cards[47-1]
// 		cards[14] = room.cards[50-1]
// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[46-1]
// 		cards[18] = room.cards[5-1]
// 		cards[19] = room.cards[8-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 4
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]
// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[44-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]
// 		cards[9] = room.cards[12-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]
// 		cards[12] = room.cards[42-1]
// 		cards[13] = room.cards[47-1]
// 		cards[14] = room.cards[50-1]
// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[46-1]
// 		cards[18] = room.cards[6-1]
// 		cards[19] = room.cards[8-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 5
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]
// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[44-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]
// 		cards[19] = room.cards[35-1]
// 		cards[9] = room.cards[12-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]
// 		cards[12] = room.cards[42-1]
// 		cards[13] = room.cards[47-1]
// 		cards[14] = room.cards[50-1]
// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[46-1]
// 		cards[18] = room.cards[8-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 6
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]
// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[44-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]
// 		cards[19] = room.cards[35-1]
// 		cards[9] = room.cards[12-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]
// 		cards[12] = room.cards[42-1]
// 		cards[13] = room.cards[47-1]
// 		cards[14] = room.cards[50-1]
// 		cards[15] = room.cards[9-1]
// 		cards[16] = room.cards[19-1]
// 		cards[17] = room.cards[29-1]
// 		cards[18] = room.cards[39-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 7
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[3-1]
// 		cards[1] = room.cards[13-1]
// 		cards[2] = room.cards[43-1]
// 		cards[5] = room.cards[4-1]
// 		cards[3] = room.cards[14-1]
// 		cards[4] = room.cards[44-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[45-1]
// 		cards[9] = room.cards[7-1]
// 		cards[10] = room.cards[8-1]
// 		cards[11] = room.cards[9-1]
// 		cards[12] = room.cards[46-1]
// 		cards[13] = room.cards[47-1]
// 		cards[14] = room.cards[48-1]
// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[56-1]
// 		cards[18] = room.cards[18-1]
// 		cards[19] = room.cards[19-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 8
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[2-1]
// 		cards[1] = room.cards[3-1]
// 		cards[2] = room.cards[4-1]
// 		cards[5] = room.cards[12-1]
// 		cards[3] = room.cards[13-1]
// 		cards[4] = room.cards[14-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]
// 		cards[9] = room.cards[6-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[8-1]
// 		cards[12] = room.cards[17-1]
// 		cards[13] = room.cards[18-1]
// 		cards[14] = room.cards[19-1]
// 		cards[15] = room.cards[47-1]
// 		cards[16] = room.cards[48-1]
// 		cards[17] = room.cards[49-1]
// 		cards[18] = room.cards[44-1]
// 		cards[19] = room.cards[45-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 9
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[2-1]
// 		cards[1] = room.cards[3-1]
// 		cards[2] = room.cards[4-1]
// 		cards[5] = room.cards[12-1]
// 		cards[3] = room.cards[13-1]
// 		cards[4] = room.cards[14-1]
// 		cards[6] = room.cards[55-1]
// 		cards[7] = room.cards[56-1]
// 		cards[8] = room.cards[57-1]
// 		cards[9] = room.cards[6-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[8-1]
// 		cards[12] = room.cards[17-1]
// 		cards[13] = room.cards[18-1]
// 		cards[14] = room.cards[19-1]
// 		cards[15] = room.cards[47-1]
// 		cards[16] = room.cards[48-1]
// 		cards[17] = room.cards[49-1]
// 		cards[18] = room.cards[44-1]
// 		cards[19] = room.cards[45-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})

// 	//case 10
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[2-1]
// 		cards[1] = room.cards[3-1]
// 		cards[2] = room.cards[4-1]
// 		cards[5] = room.cards[12-1]
// 		cards[3] = room.cards[13-1]
// 		cards[4] = room.cards[14-1]
// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]
// 		cards[9] = room.cards[6-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[8-1]
// 		cards[12] = room.cards[17-1]
// 		cards[13] = room.cards[18-1]
// 		cards[14] = room.cards[19-1]
// 		cards[15] = room.cards[47-1]
// 		cards[16] = room.cards[48-1]
// 		cards[17] = room.cards[49-1]
// 		cards[18] = room.cards[42-1]
// 		cards[19] = room.cards[57-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		if ok, kaoCard := player0.controller.CheckBao(); ok {
// 			fmt.Print("能报:")
// 			PrintCards(kaoCard)
// 		} else {
// 			fmt.Println("不能报")
// 		}
// 	})
// }

// //测试摆
// func TestBai() {
// 	player0 := players[0]
// 	//case 1
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[11-1]
// 		cards[2] = room.cards[21-1]
// 		cards[3] = room.cards[31-1]

// 		cards[4] = room.cards[2-1]
// 		cards[5] = room.cards[12-1]
// 		cards[6] = room.cards[22-1]
// 		cards[7] = room.cards[32-1]

// 		cards[8] = room.cards[3-1]
// 		cards[9] = room.cards[13-1]
// 		cards[10] = room.cards[23-1]
// 		cards[11] = room.cards[33-1]

// 		cards[12] = room.cards[7-1]
// 		cards[13] = room.cards[17-1]
// 		cards[14] = room.cards[27-1]

// 		cards[15] = room.cards[4-1]
// 		cards[16] = room.cards[5-1]
// 		cards[17] = room.cards[6-1]

// 		cards[18] = room.cards[48-1]
// 		cards[19] = room.cards[49-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		CheckBai()
// 	})

// 	//case 2
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[11-1]
// 		cards[2] = room.cards[21-1]

// 		cards[4] = room.cards[2-1]
// 		cards[5] = room.cards[12-1]
// 		cards[6] = room.cards[22-1]

// 		cards[8] = room.cards[3-1]
// 		cards[9] = room.cards[13-1]
// 		cards[10] = room.cards[23-1]

// 		cards[12] = room.cards[7-1]
// 		cards[13] = room.cards[17-1]
// 		cards[14] = room.cards[27-1]

// 		cards[15] = room.cards[4-1]
// 		cards[16] = room.cards[5-1]
// 		cards[17] = room.cards[6-1]

// 		cards[3] = room.cards[14-1]
// 		cards[7] = room.cards[15-1]
// 		cards[11] = room.cards[16-1]

// 		cards[18] = room.cards[48-1]
// 		cards[19] = room.cards[49-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		CheckBai()
// 	})

// 	//case 3
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[11-1]
// 		cards[2] = room.cards[21-1]

// 		cards[4] = room.cards[3-1]
// 		cards[5] = room.cards[13-1]
// 		cards[6] = room.cards[23-1]

// 		cards[15] = room.cards[4-1]
// 		cards[16] = room.cards[5-1]
// 		cards[17] = room.cards[6-1]

// 		cards[3] = room.cards[14-1]
// 		cards[7] = room.cards[15-1]
// 		cards[11] = room.cards[16-1]

// 		cards[12] = room.cards[8-1]
// 		cards[13] = room.cards[9-1]
// 		cards[14] = room.cards[18-1]
// 		cards[18] = room.cards[19-1]

// 		cards[8] = room.cards[46-1]
// 		cards[9] = room.cards[56-1]
// 		cards[10] = room.cards[66-1]

// 		cards[19] = room.cards[41-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		CheckBai()
// 	})

// 	//case 4
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[11-1]
// 		cards[2] = room.cards[21-1]
// 		cards[3] = room.cards[31-1]

// 		cards[8] = room.cards[2-1]
// 		cards[9] = room.cards[12-1]
// 		cards[10] = room.cards[22-1]

// 		cards[4] = room.cards[3-1]
// 		cards[5] = room.cards[13-1]
// 		cards[6] = room.cards[23-1]
// 		cards[7] = room.cards[33-1]

// 		cards[15] = room.cards[4-1]
// 		cards[16] = room.cards[5-1]
// 		cards[17] = room.cards[6-1]

// 		cards[11] = room.cards[50-1]
// 		cards[12] = room.cards[60-1]
// 		cards[13] = room.cards[70-1]

// 		cards[14] = room.cards[47-1]
// 		cards[18] = room.cards[48-1]
// 		cards[19] = room.cards[49-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)
// 		CheckBai()
// 	})
// }

// func CheckBai() {
// 	player0 := players[0]
// 	canSanLongBai := player0.controller.CheckSanLongBai()
// 	if canSanLongBai {
// 		fmt.Println("能三拢")
// 		return
// 	}

// 	canHeiBai := player0.controller.CheckHeiBai()
// 	if canHeiBai {
// 		fmt.Println("能黑摆")
// 		return
// 	}

// 	canSiKanBai := player0.controller.CheckSiKanBai()
// 	if canSiKanBai {
// 		fmt.Println("能四坎")
// 		return
// 	}
// }

// //测试招
// func TestZhao() {
// 	player0 := players[0]
// 	//case 1
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[11-1]
// 		cards[2] = room.cards[41-1]

// 		cards[3] = room.cards[50-1]
// 		cards[4] = room.cards[60-1]
// 		cards[5] = room.cards[70-1]

// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]

// 		cards[9] = room.cards[2-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]

// 		cards[12] = room.cards[12-1]
// 		cards[13] = room.cards[17-1]
// 		cards[14] = room.cards[20-1]

// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[46-1]

// 		cards[18] = room.cards[8-1]
// 		cards[19] = room.cards[18-1]

// 		player0.room.Licensing(cards)

// 		//PrintPlayer(player0)

// 		openCard := player0.room.OpenOneCard()
// 		if ok := player0.controller.CheckZhao(openCard); ok {
// 			player0.controller.Zhao(openCard)
// 			fmt.Println("能招:")
// 		} else {
// 			fmt.Println("不能招")
// 		}
// 	})
// }

// //测试碰
// func TestPeng() {
// 	player0 := players[0]

// 	//case 1
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]

// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[6-1]

// 		cards[6] = room.cards[5-1]
// 		cards[7] = room.cards[15-1]
// 		cards[8] = room.cards[25-1]

// 		cards[9] = room.cards[2-1]
// 		cards[10] = room.cards[7-1]
// 		cards[11] = room.cards[10-1]

// 		cards[12] = room.cards[12-1]
// 		cards[13] = room.cards[17-1]
// 		cards[14] = room.cards[20-1]

// 		cards[15] = room.cards[6-1]
// 		cards[16] = room.cards[16-1]
// 		cards[17] = room.cards[46-1]

// 		cards[18] = room.cards[50-1]
// 		cards[19] = room.cards[60-1]

// 		player0.room.Licensing(cards)

// 		player0.controller.huController.UpdateData(player0.cards)

// 		//PrintPlayer(player0)

// 		openCard := player0.room.OpenOneCard()
// 		if ok := player0.controller.CheckPeng(openCard); ok {
// 			player0.controller.Peng(openCard)
// 			fmt.Println("能碰:")
// 		} else {
// 			fmt.Println("不能碰:", openCard.value, openCard.big)
// 		}
// 	})
// }

// //测试吃
// func TestChi() {
// 	player0 := players[0]

// 	//case 1
// 	TestBasic(func() {
// 		room := player0.room
// 		cards := make([]*DaerCard, CardTotalAmount)
// 		copy(cards, room.cards)
// 		//修改第一个玩家的牌
// 		cards[0] = room.cards[1-1]
// 		cards[1] = room.cards[2-1]
// 		cards[2] = room.cards[3-1]

// 		cards[3] = room.cards[4-1]
// 		cards[4] = room.cards[14-1]
// 		cards[5] = room.cards[5-1]

// 		cards[6] = room.cards[9-1]
// 		cards[7] = room.cards[19-1]

// 		cards[8] = room.cards[42-1]
// 		cards[9] = room.cards[47-1]
// 		cards[10] = room.cards[50-1]

// 		cards[11] = room.cards[46-1]
// 		cards[12] = room.cards[56-1]
// 		cards[13] = room.cards[66-1]

// 		cards[14] = room.cards[6-1]
// 		cards[15] = room.cards[7-1]
// 		cards[16] = room.cards[10-1]

// 		cards[17] = room.cards[8-1]
// 		cards[18] = room.cards[18-1]
// 		cards[19] = room.cards[28-1]

// 		player0.room.Licensing(cards)

// 		player0.controller.huController.UpdateData(player0.cards)

// 		//PrintPlayer(player0)

// 		openCard := NewCard(0, 3, false)
// 		if chiPattern, biPattern := player0.controller.CheckChi(openCard); len(chiPattern) > 0 {
// 			PrintPatterns(chiPattern)
// 			for _, v := range biPattern {
// 				PrintPatterns(v)
// 			}

// 			player0.controller.Chi(chiPattern[0].cards[:len(chiPattern[0].cards)-1], chiPattern[0].cards[len(chiPattern[0].cards)], biPattern[chiPattern[0].id][0].cards)
// 			fmt.Println("能吃:")
// 		} else {
// 			fmt.Println("不能吃:", openCard.value, openCard.big)
// 		}
// 	})
// }
