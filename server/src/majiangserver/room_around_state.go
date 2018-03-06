package majiangserver

type RoomAroundState struct {
	huPlayers []*MaJiangPlayer
}

//新建一个卡牌
func NewRoomAroundState() *RoomAroundState {
	o := &RoomAroundState{}
	o.huPlayers = make([]*MaJiangPlayer, 0)
	return o
}

func (self *RoomAroundState) ClearAll() {
	self.ClearHuPlayers()
}

func (self *RoomAroundState) ClearHuPlayers() {
	self.huPlayers = make([]*MaJiangPlayer, 0)
}

func (self *RoomAroundState) AddPlayerOfHu(player *MaJiangPlayer) {
	if player == nil {
		return
	}

	self.huPlayers = append(self.huPlayers, player)
}

func (self *RoomAroundState) GetPlayerOfHu() (amount int32, players []*MaJiangPlayer) {
	if self.huPlayers == nil {
		return 0, nil
	}

	return int32(len(self.huPlayers)), self.huPlayers
}

func (self *RoomAroundState) HaveHuPlayer() bool {
	huPlayerAmount, _ := self.GetPlayerOfHu()
	return huPlayerAmount > 0
}
