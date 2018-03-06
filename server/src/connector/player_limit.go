package connector

// import (
// 	"rpc"
// 	"time"
// )

// 初始化玩家的limitinfo
// func (self *player) initLimitInfo() {
// 	if self.GetLimitInfo() == nil {
// 		tempInfo := &rpc.LimitInfo{}
// 		self.SetLimitInfo(tempInfo)
// 	}
// }


// 是否可以玩指定玩法(是否IDIP禁止了所有玩法或者指定玩法)
// func (self *player) bePlayType(playType rpc.PlayType) (canPlay bool, errMsg string) {
// 	canPlay, errMsg = self.limitAll()
// 	if !canPlay {
// 		return
// 	}
// 	playTypeInfo := self.GetLimitInfo().GetLimitPlay()
// 	if playTypeInfo == nil {
// 		return
// 	}
// 	if playType == rpc.PlayType_PT_PVE {
// 		return self.limitPve()
// 	}
// 	if playType == rpc.PlayType_PT_PVP {
// 		return self.limitPvp()
// 	}
// 	if playType == rpc.PlayType_PT_FriendAttack {
// 		return self.limitFriendAttack()
// 	}
// 	if playType == rpc.PlayType_PT_TTT {
// 		return self.limitTTT()
// 	}
// 	return
// }

// 是否禁言 return true:可以发言
// func (self *player) beChat() (canChat bool, errMsg string) {
// 	canChat = true
// 	errMsg = ""
// 	if self.GetLimitInfo() == nil {
// 		self.initLimitInfo()
// 		return
// 	}
// 	chatInfo := self.GetLimitInfo().GetLimitChat()
// 	if chatInfo == nil {
// 		return
// 	}
// 	if !chatInfo.GetOpen() {
// 		return
// 	}
// 	curTime := uint32(time.Now().Unix())
// 	if curTime < chatInfo.GetEndTime() {
// 		errMsg = chatInfo.GetText()
// 		canChat = false
// 		return
// 	}
// 	return
// }


// 是否封号 return true:可以登录
// func (self *player) beFreeze() (beLogin bool, errMsg string) {
// 	beLogin = true
// 	errMsg = ""
// 	if self.GetLimitInfo() == nil {
// 		self.initLimitInfo()
// 		return
// 	}
// 	freezeInfo := self.GetLimitInfo().GetLimitLogin()
// 	if freezeInfo == nil {
// 		return
// 	}
// 	if !freezeInfo.GetOpen() {
// 		return
// 	}
// 	curTime := uint32(time.Now().Unix())
// 	if curTime < freezeInfo.GetEndTime() {
// 		errMsg = freezeInfo.GetText()
// 		beLogin = false
// 		return
// 	}
// 	return
// }
