package majiangserver

import (
	"fmt"
	"logger"
	//"testing"
)

var HandCardInitAmount = 13

//测试控制器中的函数（CheckLong, Long...）
func TestControllerFunc() {

	//测试报
	//TestBao()

	//测试胡

	//测试招
	//TestZhao()
	//测试碰
	//TestPeng()
	//测试查叫
	//测试结算

}

//测试房间
func Template(iter func(p *MaJiangPlayer)) {

	//创建房间
	room := NewMajiangRoom(1, 1)
	room.QiHuKeAmount = 2

	player := NewMaJiangPlayer("test01", nil)
	player.SetRoom(room)

	//room.ResetRoom()

	player.isReady = true

	iter(player)

	PrintPlayer(player)

	fmt.Println("================发牌阶段结束=====================")
}

//测试报
func TestBao() {

	// //case 1
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 2)
	// 	cards[2] = NewCard(2, Tiao, 3)
	// 	cards[3] = NewCard(3, Tiao, 4)
	// 	cards[4] = NewCard(4, Tiao, 5)
	// 	cards[5] = NewCard(5, Tiao, 6)
	// 	cards[6] = NewCard(6, Tiao, 8)
	// 	cards[7] = NewCard(7, Tiao, 8)
	// 	cards[8] = NewCard(8, Tiao, 8)
	// 	cards[9] = NewCard(9, Tiao, 8)
	// 	cards[10] = NewCard(10, Wan, 1)
	// 	cards[11] = NewCard(11, Wan, 2)
	// 	cards[12] = NewCard(12, Wan, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 2
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 2)
	// 	cards[2] = NewCard(2, Tiao, 3)
	// 	cards[3] = NewCard(3, Tiao, 4)
	// 	cards[4] = NewCard(4, Tiao, 5)
	// 	cards[5] = NewCard(5, Tiao, 6)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tong, 2)
	// 	cards[11] = NewCard(11, Tong, 2)
	// 	cards[12] = NewCard(12, Tong, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 3
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 1)
	// 	cards[2] = NewCard(2, Tiao, 2)
	// 	cards[3] = NewCard(3, Tiao, 2)
	// 	cards[4] = NewCard(4, Tiao, 3)
	// 	cards[5] = NewCard(5, Tiao, 3)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tong, 2)
	// 	cards[11] = NewCard(11, Tong, 2)
	// 	cards[12] = NewCard(12, Tong, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 4
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 1)
	// 	cards[2] = NewCard(2, Tiao, 1)
	// 	cards[3] = NewCard(3, Tiao, 1)
	// 	cards[4] = NewCard(4, Tiao, 3)
	// 	cards[5] = NewCard(5, Tiao, 3)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tong, 2)
	// 	cards[11] = NewCard(11, Tong, 2)
	// 	cards[12] = NewCard(12, Tong, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 5
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 1)
	// 	cards[2] = NewCard(2, Tiao, 1)
	// 	cards[3] = NewCard(3, Tiao, 1)
	// 	cards[4] = NewCard(4, Tiao, 2)
	// 	cards[5] = NewCard(5, Tiao, 3)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tiao, 4)
	// 	cards[11] = NewCard(11, Tiao, 5)
	// 	cards[12] = NewCard(12, Tiao, 6)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 6
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 1)
	// 	cards[2] = NewCard(2, Tiao, 1)
	// 	cards[3] = NewCard(3, Tiao, 1)
	// 	cards[4] = NewCard(4, Tiao, 2)
	// 	cards[5] = NewCard(5, Tiao, 2)
	// 	cards[6] = NewCard(6, Tiao, 2)
	// 	cards[7] = NewCard(7, Tiao, 2)
	// 	cards[8] = NewCard(8, Tiao, 3)
	// 	cards[9] = NewCard(9, Tiao, 3)
	// 	cards[10] = NewCard(10, Tiao, 3)
	// 	cards[11] = NewCard(11, Tiao, 3)
	// 	cards[12] = NewCard(12, Tiao, 6)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	//case 7
	Template(func(player *MaJiangPlayer) {
		//初始化手牌
		cards := make([]*MaJiangCard, HandCardInitAmount)

		cards[0] = NewCard(0, Tiao, 1)
		cards[1] = NewCard(1, Tiao, 1)
		cards[2] = NewCard(2, Tiao, 1)
		cards[3] = NewCard(3, Tiao, 2)
		cards[4] = NewCard(4, Tiao, 3)
		cards[5] = NewCard(5, Tiao, 4)
		cards[6] = NewCard(6, Tiao, 5)
		cards[7] = NewCard(7, Tiao, 6)
		cards[8] = NewCard(8, Tiao, 7)
		cards[9] = NewCard(9, Tiao, 8)
		cards[10] = NewCard(10, Tiao, 9)
		cards[11] = NewCard(11, Tiao, 9)
		cards[12] = NewCard(12, Tiao, 9)

		player.Compose(cards)

		if ok, kaoCard := player.controller.CheckBao(); ok {
			logger.Info("能报:")
			PrintCards(kaoCard)
		} else {
			logger.Info("不能报")
		}
	})

	//case 1
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, HongZhong, 0)
	// 	cards[1] = NewCard(1, Tiao, 2)
	// 	cards[2] = NewCard(2, Tiao, 3)
	// 	cards[3] = NewCard(3, Tiao, 4)
	// 	cards[4] = NewCard(4, Tiao, 5)
	// 	cards[5] = NewCard(5, Tiao, 6)
	// 	cards[6] = NewCard(6, Tiao, 8)
	// 	cards[7] = NewCard(7, Tiao, 8)
	// 	cards[8] = NewCard(8, Tiao, 8)
	// 	cards[9] = NewCard(9, Tiao, 8)
	// 	cards[10] = NewCard(10, Tiao, 1)
	// 	cards[11] = NewCard(11, Tiao, 2)
	// 	cards[12] = NewCard(12, Tiao, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 2
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 2)
	// 	cards[2] = NewCard(2, Tiao, 3)
	// 	cards[3] = NewCard(3, Tiao, 4)
	// 	cards[4] = NewCard(4, Tiao, 5)
	// 	cards[5] = NewCard(5, Tiao, 6)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tong, 2)
	// 	cards[11] = NewCard(11, Tong, 2)
	// 	cards[12] = NewCard(12, Tong, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 3
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 1)
	// 	cards[2] = NewCard(2, Tiao, 2)
	// 	cards[3] = NewCard(3, Tiao, 2)
	// 	cards[4] = NewCard(4, Tiao, 3)
	// 	cards[5] = NewCard(5, Tiao, 3)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tong, 2)
	// 	cards[11] = NewCard(11, Tong, 2)
	// 	cards[12] = NewCard(12, Tong, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 4
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 1)
	// 	cards[2] = NewCard(2, Tiao, 1)
	// 	cards[3] = NewCard(3, Tiao, 1)
	// 	cards[4] = NewCard(4, Tiao, 3)
	// 	cards[5] = NewCard(5, Tiao, 3)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tong, 2)
	// 	cards[11] = NewCard(11, Tong, 2)
	// 	cards[12] = NewCard(12, Tong, 3)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	// //case 5
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, Tiao, 1)
	// 	cards[1] = NewCard(1, Tiao, 1)
	// 	cards[2] = NewCard(2, Tiao, 1)
	// 	cards[3] = NewCard(3, Tiao, 1)
	// 	cards[4] = NewCard(4, Tiao, 2)
	// 	cards[5] = NewCard(5, Tiao, 3)
	// 	cards[6] = NewCard(6, Tong, 8)
	// 	cards[7] = NewCard(7, Tong, 8)
	// 	cards[8] = NewCard(8, Tong, 8)
	// 	cards[9] = NewCard(9, Tong, 8)
	// 	cards[10] = NewCard(10, Tiao, 4)
	// 	cards[11] = NewCard(11, Tiao, 5)
	// 	cards[12] = NewCard(12, Tiao, 6)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

	//case 6
	// Template(func(player *MaJiangPlayer) {
	// 	//初始化手牌
	// 	cards := make([]*MaJiangCard, HandCardInitAmount)

	// 	cards[0] = NewCard(0, HongZhong, 0)
	// 	cards[1] = NewCard(1, HongZhong, 0)
	// 	cards[2] = NewCard(2, HongZhong, 1)
	// 	cards[3] = NewCard(3, HongZhong, 1)
	// 	cards[4] = NewCard(4, Tiao, 2)
	// 	cards[5] = NewCard(5, Tiao, 2)
	// 	cards[6] = NewCard(6, Tiao, 2)
	// 	cards[7] = NewCard(7, Tiao, 2)
	// 	cards[8] = NewCard(8, Tiao, 3)
	// 	cards[9] = NewCard(9, Tiao, 3)
	// 	cards[10] = NewCard(10, Tiao, 3)
	// 	cards[11] = NewCard(11, Tiao, 3)
	// 	cards[12] = NewCard(12, Tong, 6)

	// 	player.Compose(cards)

	// 	if ok, kaoCard := player.controller.CheckBao(); ok {
	// 		logger.Info("能报:")
	// 		PrintCards(kaoCard)
	// 	} else {
	// 		logger.Info("不能报")
	// 	}
	// })

}

//测试招
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
