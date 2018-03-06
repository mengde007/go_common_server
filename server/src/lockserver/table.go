package lockserver

import (
	"common"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"hash/crc32"
	"io"
	"logger"
	"stats"
	//Record"time"
)

const (
	keylen = 64
)

type table struct {
	name         string
	caches       cacheGroup
	deleteExpiry uint64
	tableStats   *stats.Timings
	qpsRates     *stats.Rates
	cacheNode    []uint32
}

func NewTable(name string, cfg common.TableConfig, db *LockServerServices) (t *table) {

	var (
		caches    cacheGroup
		cacheNode []uint32
	)
	if cfg.CacheProfile != "" {
		var exist bool
		if caches, exist = db.cacheGroups[cfg.CacheProfile]; !exist {
			logger.Fatal("NewTable: table cache profile not found: %s", cfg.CacheProfile)
		}
		cacheNode, _ = db.cacheNodes[cfg.CacheProfile]
	}

	if caches == nil {
		logger.Fatal("NewTable: table %s need a save func", name)
	}

	queryStats := stats.NewTimings("")
	qpsRates := stats.NewRates("", queryStats, 20, 10e9)
	return &table{
		name, caches,
		cfg.DeleteExpiry,
		queryStats,
		qpsRates,
		cacheNode,
	}
}

func (self *table) write(key string, value []byte) (err error) {

	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	//defer self.tableStats.Record("write", time.Now())
	hid := makeHash(key)
	if self.caches != nil {
		cidx := self.getCacheNode(hid)
		cache := cidx.Get()
		defer cache.Recycle()

		_, err = cache.Do("SET", self.name+common.DbTableKeySplit+key, value)
		if err != nil {
			logger.Error("write error: %s (%s, %v)", err.Error(), key, value)
		}
	}

	return
}

func (self *table) get(key string) (ret []byte, err error) {
	if len(key) > keylen {
		return nil, fmt.Errorf("key (%s) len must <= 64", key)
	}

	//defer self.tableStats.Record("get", time.Now())
	hid := makeHash(key)
	if self.caches != nil {
		cidx := self.getCacheNode(hid)
		cache := cidx.Get()
		defer cache.Recycle()

		ret, err = redis.Bytes(cache.Do("GET", self.name+common.DbTableKeySplit+key))
		if err != nil {
			if err != redis.ErrNil {
				logger.Error("table get error: %s (%s, %v)", err.Error(), key, ret)
				return
			} else {
				err = nil
			}
		}
		if ret != nil {
			return
		}
	}
	return nil, nil
}

func (self *table) del(key string) (err error) {

	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	//defer self.tableStats.Record("del", time.Now())
	hid := makeHash(key)
	if self.caches != nil {
		cidx := self.getCacheNode(hid)
		cache := cidx.Get()
		defer cache.Recycle()

		_, err = cache.Do("DEL", self.name+common.DbTableKeySplit+key)
		if err != nil {
			if err != redis.ErrNil {
				logger.Error("del error: %s (%s)", err.Error(), key)
				return
			} else {
				err = nil
			}
		}
	}

	return
}

func (self *table) getCacheNode(key uint32) *common.CachePool {

	var index = 0
	for k, v := range self.cacheNode {
		if key < v {
			index = k
			break
		}
	}

	node, ok := self.caches[self.cacheNode[index]]
	if !ok {
		logger.Error("getCacheNode error: no find (%d)", key)
	}
	return node
}

func makeHash(key string) uint32 {

	ieee := crc32.NewIEEE()
	io.WriteString(ieee, key)
	return ieee.Sum32()
}
