package connector

import (
	//"crypto/rc4"
	//"encoding/binary"
	//"fmt"
	"common"
	"logger"
	// "rpc"
	"sync/atomic"
	"time"
)

func GenLockMessage(sid uint8, tid uint8, value uint8) uint64 {
	return common.GenLockMessage(sid, tid, value)
}

func ParseLockMessage(lid uint64) (sid uint8, tid uint8, value uint8, t uint32, tmpid uint8) {
	return common.ParseLockMessage(lid)
}

var pid uint32 = 0

func GenPlayerId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&pid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(sid)<<56
}

var vid uint32 = 0

func GenVillageId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&vid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(sid)<<56
}

var bid uint32 = 0

func GenBattleId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&bid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(sid)<<56
}

var rid uint32 = 0

func GenReplayId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&rid, 1))

	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(sid)<<56
}

var wsbid uint32 = 0

func GetWsBattleId(sid uint8) uint64 {
	tmpid := uint16(atomic.AddUint32(&wsbid, 1))
	return uint64(tmpid) | uint64(time.Now().Unix())<<16 | uint64(sid)<<56
}

// UUID() provides unique identifier strings.
func GenUUID(sid uint8) string {
	return common.GenUUID(sid)
}

func CheckUUID(uid string) bool {
	return common.CheckUUID(uid)
}

const DEBUG = true

func dbgf(format string, items ...interface{}) {
	if DEBUG {
		logger.Info(format, items...)
	}
}

const TRACE = true

func ts(name string, items ...interface{}) {
	if TRACE {
		logger.Info("+%s %v\n", name, items)
	}
}
func te(name string, items ...interface{}) {
	if TRACE {
		logger.Info("-%s %v\n", name, items)
	}
}

func pts(p *player, name string, items ...interface{}) {
	if TRACE && p != nil {
		p.LogInfo("+%s %v\n", name, items)
	}
}
func pte(p *player, name string, items ...interface{}) {
	if TRACE {
		p.LogInfo("-%s %v\n", name, items)
	}
}

func RandomNumber(start uint32, stop uint32) uint32 {
	return common.RandomNumber(start, stop)
}

func RandomWeightTable(table map[interface{}]uint32) interface{} {
	return common.RandomWeightTable(table)
}

