package connector

import (
	// "centerclient"
	// "clanclient"
	// "common"
	"fmt"
	// "jfclient"
	"logger"
	// "proto"
	// "rpc"
	// "strings"
)

//同步宝石数量
// func (self *player) SyncPlayerGem() {
// 	update := &rpc.UpdatePlayerInfo{}
// 	update.SetDiamonds(self.GetPlayerTotalGem())

// 	WriteResult(self.conn, update)
// }

// log 相关函数
func getTitle(self *player) string {
	openId := self.txOpenId
	uid := self.GetUid()
	connId := uint64(0)
	if self.conn != nil {
		connId = self.conn.GetId()
	}
	return fmt.Sprintf("%s %s connId = %d : ", openId, uid, connId)
}

func (self *player) LogDebug(format string, args ...interface{}) {
	logger.Debug(getTitle(self)+format, args...)
}

func (self *player) LogInfo(format string, args ...interface{}) {
	logger.Info(getTitle(self)+format, args...)
}

func (self *player) LogWarning(format string, args ...interface{}) {
	logger.Warning(getTitle(self)+format, args...)
}

func (self *player) LogError(format string, args ...interface{}) {
	logger.Error(getTitle(self)+format, args...)
}

func (self *player) LogFatal(format string, args ...interface{}) {
	logger.Error(getTitle(self)+format, args...)
}
