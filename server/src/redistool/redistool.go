package redistool

import (
	"common"
	"errors"
	"github.com/garyburd/redigo/redis"
	"logger"
	"strconv"
	"strings"
	"sync"
	"timer"
)

const (
	SHIELD_UID_SEG = 10240
)

const (
	TN_SHIELD = "shield"
)

type stShieldRemain struct {
	tm        *timer.Timer
	timeStart uint32
	timeTotal uint32
	trophy    uint32
}

type stShieldSeg struct {
	m map[string]*stShieldRemain
	l sync.RWMutex
}

type MatchServer struct {
	pCatchPool  *common.CachePool
	pMatchcache *common.CachePool
	shields     []*stShieldSeg
}

func genShieldInfo(timeStart, timeTotal, trophy uint32) string {
	ts := strconv.FormatUint(uint64(timeStart), 10)
	tt := strconv.FormatUint(uint64(timeTotal), 10)
	tr := strconv.FormatUint(uint64(trophy), 10)

	return ts + "|" + tt + "|" + tr
}

func parseShieldInfo(info string) (timeStart, timeTotal, trophy uint32, err error) {
	ret := strings.Split(info, "|")
	if len(ret) != 3 {
		err = errors.New("wrong format")
		return
	}

	var t uint64
	if t, err = strconv.ParseUint(ret[0], 10, 0); err != nil {
		return
	} else {
		timeStart = uint32(t)
	}

	if t, err = strconv.ParseUint(ret[1], 10, 0); err != nil {
		return
	} else {
		timeTotal = uint32(t)
	}

	if t, err = strconv.ParseUint(ret[2], 10, 0); err != nil {
		return
	} else {
		trophy = uint32(t)
	}

	return
}

func MoveData(cfg []common.MatchServerConfig) error {
	pServer := make([]*MatchServer, 4, 4)

	logger.Info("len:%d", len(cfg))
	for i := len(cfg) - 1; i >= 0; i-- {
		logger.Info("len:%d,i:%d", len(cfg), i)
		pServer[i] = &MatchServer{
			pCatchPool: common.NewCachePool(cfg[i].Maincache),
			//			pMatchcache: common.NewCachePool(cfg[i].Matchcache),
			shields: make([]*stShieldSeg, SHIELD_UID_SEG),
		}
		logger.Info("1_len:%d;i:%d", len(cfg), i)
		//初始化
		for j := 0; j < SHIELD_UID_SEG; j++ {
			pServer[i].shields[j] = &stShieldSeg{
				m: make(map[string]*stShieldRemain),
			}
		}

	}

	logger.Info("2_len:%d", len(cfg))
	for i := len(cfg) - 1; i >= 0; i-- {
		cache := pServer[i].pCatchPool.Get()
		defer cache.Recycle()

		table := TN_SHIELD + common.DbTableKeySplit + ""
		iter := 0
		//		wg := &sync.WaitGroup{}

		for {
			var SaveValue [][]byte
			logger.Info("table:%s,i:%d", table, i)
			all, err := redis.Values(cache.Do("HSCAN", table, iter))
			if err != nil {
				logger.Fatal("HSCAN shield error: %s,pid:%d", err.Error(), i)
				return err
			}

			logger.Info("all_len:%v", all)
			if _, err = redis.Scan(all, &iter, &SaveValue); err != nil {
				logger.Fatal("redis.Scan shield error: %s", err.Error())
				return err
			}

			//			wg.Add(1)
			//			go func(retBuf *[][]byte) {
			retBuf := &SaveValue
			for j := 0; j < len(*retBuf); j += 2 {
				uid := string((*retBuf)[j])
				info := string((*retBuf)[j+1])
				//				timeStart, timeTotal, trophy, err := parseShieldInfo(info)
				//正确读出数据
				if err == nil {
					idxnew := common.MakeHash(uid) % uint32(len(pServer))
					logger.Info("uid:%s,idxnew:%d,i:%d", uid, idxnew, i)
					//需要移动i--->idxnew
					if idxnew != uint32(i) {
						//err = common.Redis_hset(pServer[idxnew].pCatchPool, TN_SHIELD, "", uid, genShieldInfo(timeStart, timeTotal, trophy))
						err = common.Redis_hset(pServer[idxnew].pCatchPool, TN_SHIELD, "", uid, info)
						if err != nil {
							logger.Fatal("hset %d cache error!!", idxnew)
						} else {
							err = common.Redis_hdel(pServer[i].pCatchPool, TN_SHIELD, "", uid)
							if err != nil {
								logger.Fatal("hdel %d cache error!!", idxnew)
							}
						}
					}
				} else {
					logger.Fatal("parseShieldInfo error: %s", err.Error())
				}
			}
			//				wg.Done()
			//			}(&SaveValue)

			if 0 == iter {
				break
			}
		}
	}
	return nil
}
