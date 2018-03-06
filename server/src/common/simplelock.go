package common

import (
	"logger"
	"sync"
)

const (
	HASH_SEG = 4096
)

type SimpleSingleLock struct {
	l sync.Mutex
}

type SimpleLockService struct {
	l      sync.RWMutex
	m2Lock map[string]*SimpleSingleLock
}

type SimpleHashLockService struct {
	vecHash []*SimpleLockService
}

func getSeg(uid string) int {
	return int(MakeHash(uid) % uint32(HASH_SEG))
}

func (self *SimpleHashLockService) WaitLock(key string) {
	seg := getSeg(key)
	self.vecHash[seg].WaitLock(key)
}

func (self *SimpleHashLockService) WaitUnLock(key string) {
	seg := getSeg(key)
	self.vecHash[seg].WaitUnLock(key)
}

func (self *SimpleLockService) WaitLock(key string) {
	self.l.RLock()
	single, ok := self.m2Lock[key]
	self.l.RUnlock()

	if ok {
		single.l.Lock()

		return
	}

	self.l.Lock()
	single, ok = self.m2Lock[key]
	if ok {
		self.l.Unlock()

		single.l.Lock()

		return
	}

	pNew := &SimpleSingleLock{}
	pNew.l.Lock()
	self.m2Lock[key] = pNew

	self.l.Unlock()

	return
}

func (self *SimpleLockService) WaitUnLock(key string) {
	self.l.RLock()
	single, ok := self.m2Lock[key]
	self.l.RUnlock()

	if ok {
		single.l.Unlock()

		return
	} else {
		logger.Error("serious error unlock not find !")
	}
}

func CreateSimpleLock() *SimpleLockService {
	return &SimpleLockService{m2Lock: make(map[string]*SimpleSingleLock, 0)}
}

func CreateSimpleHashLock() *SimpleHashLockService {
	rt := &SimpleHashLockService{}
	rt.vecHash = make([]*SimpleLockService, HASH_SEG)

	for i, _ := range rt.vecHash {
		rt.vecHash[i] = &SimpleLockService{m2Lock: make(map[string]*SimpleSingleLock, 0)}
	}
	return rt
}
