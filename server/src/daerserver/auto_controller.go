//暂时未使用

package daerserver

import (
	"logger"
)

type AutoController struct {
	DaerController
}

//新建一个自动控制器
func NewAutoController(player *DaerPlayer) *AutoController {
	controller := new(AutoController)
	controller.player = player
	return controller
}

//获取一个出牌
func (controller *AutoController) GetChuPai() *DaerCard {
	player := controller.player
	if player == nil {
		logger.Error("controller.player is nil.")
		return nil
	}

	if player.cards == nil || len(player.cards) <= 0 {
		return nil
	}

	return player.cards[len(player.cards)-1]
}
