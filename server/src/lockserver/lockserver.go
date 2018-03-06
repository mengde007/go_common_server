package lockserver

import (
	"common"
	"logger"
	"time"
)

type LockValue struct {
	Value          uint64
	LockserverTime uint32
	ValidTime      uint32
}

type LockServer struct {
	*table

	tableName string
	lock      *common.SimpleHashLockService
}

func NewLockServer(name string, cfg common.TableConfig, db *LockServerServices) *LockServer {
	return &LockServer{table: NewTable(name, cfg, db), tableName: name, lock: common.CreateSimpleHashLock()}
}

func (self *LockServer) GetLock(name string, value uint64, validTime uint32) (ok bool, old_value uint64) {
	logger.Info("************ GetLock lockserver， name:%s, value:%d, validTime:%d", name, value, validTime)
	self.lock.WaitLock(name)
	rst, err := self.get(name)

	if err != nil {
		self.lock.WaitUnLock(name)
		logger.Info("=====================GetLock1")
		return false, 0
	}

	if rst != nil {
		var old LockValue
		if err := common.JsonDecode(rst, &old); err != nil {
			self.lock.WaitUnLock(name)
			logger.Info("=====================GetLock2")
			return false, 0
		}

		//未过期，锁失败，给45秒的缓冲时间
		if old.LockserverTime+old.ValidTime+45 > uint32(time.Now().Unix()) {
			logger.Error("lock time last valid %d, %d", old.LockserverTime, old.ValidTime)
			self.lock.WaitUnLock(name)
			logger.Info("=====================GetLock2.5")
			return false, old.Value
		}
	}

	saveValue := &LockValue{Value: value, LockserverTime: uint32(time.Now().Unix()), ValidTime: validTime}

	buf, err := common.JsonEncode(saveValue)
	if err != nil {
		self.lock.WaitUnLock(name)
		logger.Info("=====================GetLock3")
		return false, 0
	}

	if err := self.write(name, buf); err != nil {
		self.lock.WaitUnLock(name)
		logger.Info("=====================GetLock4")
		return false, 0
	}

	self.lock.WaitUnLock(name)
	logger.Info("=====================GetLock5")
	return true, value
}

func (self *LockServer) QueryPlayer(name string) uint64 {
	rst, err := self.get(name)
	if err != nil {
		return 0
	}

	if rst != nil {
		var old LockValue
		if err := common.JsonDecode(rst, &old); err != nil {
			return 0
		}

		//已经过期
		if old.LockserverTime+old.ValidTime < uint32(time.Now().Unix()) {
			return 0
		}

		return old.Value
	}

	return 0
}

func (self *LockServer) RenewLock(name string, value uint64) bool {
	self.lock.WaitLock(name)

	rst, err := self.get(name)

	if err != nil {
		self.lock.WaitUnLock(name)
		return false
	}

	//没数据则不能续期
	if rst == nil {
		self.lock.WaitUnLock(name)
		return false
	}

	var old LockValue
	if err := common.JsonDecode(rst, &old); err != nil {
		self.lock.WaitUnLock(name)
		return false
	}

	//value不一致
	if old.Value != value {
		self.lock.WaitUnLock(name)
		return false
	}

	saveValue := &LockValue{
		Value:          old.Value,
		LockserverTime: uint32(time.Now().Unix()),
		ValidTime:      old.ValidTime,
	}

	buf, err := common.JsonEncode(saveValue)
	if err != nil {
		self.lock.WaitUnLock(name)
		return false
	}

	if err := self.write(name, buf); err != nil {
		self.lock.WaitUnLock(name)
		return false
	}

	self.lock.WaitUnLock(name)

	return true
}

func (self *LockServer) UnLock(name string, value uint64) bool {
	self.lock.WaitLock(name)

	rst, err := self.get(name)

	if err != nil {
		self.lock.WaitUnLock(name)
		return false
	}

	if rst != nil {
		var old LockValue
		if err := common.JsonDecode(rst, &old); err != nil {
			self.lock.WaitUnLock(name)
			return false
		}

		if old.Value == value {
			if err := self.del(name); err != nil {
				self.lock.WaitUnLock(name)
				return false
			}
			self.lock.WaitUnLock(name)
			return true
		}
	}

	self.lock.WaitUnLock(name)
	return false
}

func (self *LockServer) ForceUnLock(name string) bool {
	self.lock.WaitLock(name)

	rst, err := self.get(name)
	if err != nil {
		self.lock.WaitUnLock(name)
		return false
	}

	if rst != nil {
		var old LockValue
		if err := common.JsonDecode(rst, &old); err != nil {
			self.lock.WaitUnLock(name)
			return false
		}

		if err := self.del(name); err != nil {
			self.lock.WaitUnLock(name)
			return false
		}
		self.lock.WaitUnLock(name)
		return true
	}

	self.lock.WaitUnLock(name)
	return false
}
