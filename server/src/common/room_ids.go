package common

import (
	"logger"
	"math/rand"
	"time"
)

// const (
// 	StartID = 1000
// 	EndID   = 9999
// )

type RoomIDS struct {
	notUsedIDS []int32
	usedIDS    []int32

	startID int32
	endID   int32
}

func (self *RoomIDS) Init(startID, endID int32) {

	self.startID = startID
	self.endID = endID

	generateIDCount := endID - startID + 1
	self.notUsedIDS = make([]int32, generateIDCount)
	self.usedIDS = make([]int32, 0)

	for i := int32(startID); i <= endID; i++ {
		self.notUsedIDS[i-startID] = i
	}

}

func (self *RoomIDS) UseID() int32 {
	remainIDAmount := len(self.notUsedIDS)
	if remainIDAmount <= 0 {
		logger.Error("ID已经用完了")
		return -1
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	index := r.Intn(remainIDAmount)

	resulst := self.notUsedIDS[index]
	self.usedIDS = append(self.usedIDS, self.notUsedIDS[index])
	self.notUsedIDS = append(self.notUsedIDS[0:index], self.notUsedIDS[index+1:]...)

	return resulst
}

func (self *RoomIDS) RemainIDCount() int {
	return len(self.notUsedIDS)
}

func (self *RoomIDS) UsedIDCount() int {
	return len(self.usedIDS)
}

func (self *RoomIDS) RecoverID(id int32) {

	for i, tempID := range self.usedIDS {
		if id == tempID {
			self.usedIDS = append(self.usedIDS[:i], self.notUsedIDS[i+1:]...)
			self.notUsedIDS = append(self.notUsedIDS, id)
			break
		}
	}

}
