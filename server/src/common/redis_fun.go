package common

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"logger"
	"strconv"
)

const (
	keylen = 64
)

func Resis_setbuf(pool *CachePool, table string, key string, value []byte) error {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := pool.Get()
	defer cache.Recycle()

	if _, err := cache.Do("SET", table+":"+key, value); err != nil {
		logger.Error("Resis_setbuf error", table, key, err)

		return err
	}

	return nil
}

func Resis_getbuf(pool *CachePool, table string, key string) ([]byte, error) {
	if len(key) > keylen {
		return nil, fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := pool.Get()
	defer cache.Recycle()

	ret, err := redis.Bytes(cache.Do("GET", table+":"+key))
	if err != nil && err != redis.ErrNil {
		logger.Error("Resis_getbuf error", table, key, err)
	}
	if err == redis.ErrNil {
		logger.Info("redis_getbuf: ErrNil")
		ret = nil
		err = nil
	}

	return ret, err
}

func Redis_del(pool *CachePool, table string, key string) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("DEL", table+":"+key)
	if err != nil {
		logger.Error("del error: %s (%s, %s, %d)", err.Error(), table, key)
	}

	return
}

func Redis_setInt(pool *CachePool, table string, key string, value int) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("SET", table+":"+key, value)
	if err != nil {
		logger.Error("setInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
	}

	return
}

func Redis_getInt(pool *CachePool, table string, key string) (value int, err error) {
	if len(key) > keylen {
		return 0, fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := pool.Get()
	defer cache.Recycle()

	value, err = redis.Int(cache.Do("GET", table+":"+key))
	if err != nil {
		if err != redis.ErrNil {
			logger.Error("getInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
		} else {
			err = nil
		}
	}

	return
}

func Redis_setString(pool *CachePool, table string, key string, value string) (err error) {
	if len(key) > keylen {
		return fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("SET", table+":"+key, value)
	if err != nil {
		logger.Error("setInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
	}

	return
}

func Redis_getString(pool *CachePool, table string, key string) (value string, err error) {
	if len(key) > keylen {
		return "", fmt.Errorf("key (%s) len must <= 64", key)
	}

	cache := pool.Get()
	defer cache.Recycle()

	value, err = redis.String(cache.Do("GET", table+":"+key))
	if err != nil {
		if err != redis.ErrNil {
			logger.Error("getInt error: %s (%s, %s, %d)", err.Error(), table, key, value)
		} else {
			err = nil
		}
	}

	return
}

func Redis_setexpire(pool *CachePool, table string, key string, time string) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("EXPIRE", table+":"+key, time)
	if err != nil {
		logger.Error("PEXPIRE error: %s (%s, %s, %s)", err.Error(), table, key, time)
	}

	return
}

func Redis_sadd(pool *CachePool, table string, key string, value string) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("SADD", table+":"+key, value)
	if err != nil {
		logger.Error("sadd error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return
}

func Redis_srem(pool *CachePool, table string, key string, value string) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("SREM", table+":"+key, value)
	if err != nil {
		logger.Error("srem error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return
}

func Redis_exists(pool *CachePool, table string, key string) bool {
	cache := pool.Get()
	defer cache.Recycle()

	exist, err := redis.Int(cache.Do("EXISTS", table+":"+key))
	if err != nil {
		logger.Error("exists error: %s (%s, %s, %d)", err.Error(), table, key, exist)
	}

	return exist == 1
}

func Redis_scard(pool *CachePool, table string, key string) (num int) {
	cache := pool.Get()
	defer cache.Recycle()

	num, err := redis.Int(cache.Do("SCARD", table+":"+key))
	if err != nil && err != redis.ErrNil {
		logger.Error("scard error: %s (%s, %s, %d)", err.Error(), table, key, num)
	}
	if err == redis.ErrNil {
		logger.Error("redis_scard: ErrNil")
		num = 0
		err = nil
	}

	return
}

func Redis_srandmember(pool *CachePool, table string, key string) (value string, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	value, err = redis.String(cache.Do("SRANDMEMBER", table+":"+key))
	if err != nil && err != redis.ErrNil {
		logger.Error("srandmember error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}
	if err == redis.ErrNil {
		logger.Error("redis_srandmember: ErrNil")
		value = ""
		err = nil
	}

	return
}

func Redis_zadd(pool *CachePool, table string, key string, value string, score uint32) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("ZADD", table+":"+key, strconv.FormatInt(int64(score), 10), value)
	if err != nil {
		logger.Error("zadd error: %s (%s, %s, %s, %d)", err.Error(), table, key, value, score)
	}

	return
}

func Redis_zrem(pool *CachePool, table string, key string, value string) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("ZREM", table+":"+key, value)
	if err != nil {
		logger.Error("zrem error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}

	return
}

func Redis_zcard(pool *CachePool, table string, key string) (uint32, error) {
	cache := pool.Get()
	defer cache.Recycle()

	length, err := redis.Int(cache.Do("zcard", table+":"+key))
	if err != nil && err != redis.ErrNil {
		logger.Error("zcard error: %s (%s, %s, %s, %d)", err.Error(), table, key, length)
	}
	if err == redis.ErrNil {
		logger.Error("redis_zcard: ErrNil")
		length = 0
		err = nil
	}

	return uint32(length), err
}

func Redis_zscore(pool *CachePool, table string, key string, value string) (uint32, error) {
	cache := pool.Get()
	defer cache.Recycle()

	score, err := redis.Int(cache.Do("ZSCORE", table+":"+key, value))
	if err != nil && err != redis.ErrNil {
		logger.Error("zscore error: %s (%s, %s, %s, %d)", err.Error(), table, key, value, score)
	}
	if err == redis.ErrNil {
		logger.Error("redis_zscore: ErrNil")
		score = 0
		err = nil
	}

	return uint32(score), err
}

func Redis_zrevrange(pool *CachePool, table string, key string, start int, stop int) (rets []string, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Do("ZREVRANGE", table+":"+key, start, stop))
	if err != nil && err != redis.ErrNil {
		logger.Error("zrevrange error: %s (%s, %s, %d, %d)", err.Error(), table, key, start, stop)
	}
	if err == redis.ErrNil {
		logger.Error("redis_zrevrange: ErrNil")
		rets = nil
		err = nil
	}

	return
}

func Redis_zrevrank(pool *CachePool, table string, key string, value string) (uint32, error) {
	cache := pool.Get()
	defer cache.Recycle()

	rank, err := redis.Int(cache.Do("ZREVRANK", table+":"+key, value))
	if err != nil && err != redis.ErrNil {
		logger.Error("zrevrank error: %s (%s, %s, %s)", err.Error(), table, key, value)
	}
	if err == redis.ErrNil {
		logger.Error("redis_zrevrank: ErrNil")
		rank = 0
		err = nil
	}

	return uint32(rank), err
}

/*func Redis_keys(pool *CachePool, table string, key string) (rets []string, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Conn.Do("keys", table+DbTableKeySplit+key))
	if err != nil {
		logger.Error("keys error: %s (%s, %s)", err.Error(), table, key)
	}

	return
}*/

func Redis_hset(pool *CachePool, table string, key string, field string, value string) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("hset", table+DbTableKeySplit+key, field, value)
	if err != nil {
		logger.Error("hset error: %s (%s, %s, %s, %s)", err.Error(), table, key, field, value)
	}

	return
}

func Redis_hgetall(pool *CachePool, table string, key string) (rets []string, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Do("hgetall", table+DbTableKeySplit+key))
	if err != nil && err != redis.ErrNil {
		logger.Error("keys error: %s (%s, %s)", err.Error(), table, key)
	}
	if err == redis.ErrNil {
		logger.Error("redis_hgetall: ErrNil")
		rets = nil
		err = nil
	}

	return
}

func Redis_hdel(pool *CachePool, table string, key string, field string) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("hdel", table+DbTableKeySplit+key, field)
	if err != nil {
		logger.Error("hdel error: %s (%s, %s, %s)", err.Error(), table, key, field)
	}

	return
}

func Redis_zrank(pool *CachePool, table, key, value string) (rank int, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	rank, err = redis.Int(cache.Do("ZRANK", table+":"+key, value))
	if err != nil && err != redis.ErrNil {
		logger.Error("zrank error: %s (%s, %s)", err.Error(), table, value)
	}
	if err == redis.ErrNil {
		logger.Error("redis_zrank: ErrNil")
		rank = 0
		err = nil
	}

	return
}

func Redis_hRename(pool *CachePool, tableName, oldKeyName, newKeyName string) error {
	cache := pool.Get()
	defer cache.Recycle()

	_, err := cache.Do("RENAME", tableName+DbTableKeySplit+oldKeyName, tableName+DbTableKeySplit+newKeyName)
	if err != nil {
		logger.Error("RENAME error: %s (%s, %s)", err.Error(), oldKeyName, newKeyName)
	}

	return err
}

func Redis_hGet(pool *CachePool, tableName, keyName, fieldName string) (number int, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	number, err = redis.Int(cache.Do("Hget", tableName+DbTableKeySplit+keyName, fieldName))
	if err != nil && err != redis.ErrNil {
		logger.Error("Get error: %s (get %s from %s)", err.Error(), tableName+DbTableKeySplit+keyName, keyName)
	}
	if err == redis.ErrNil {
		logger.Error("redis_hGet: ErrNil")
		number = 0
		err = nil
	}

	return number, err
}

func Redis_zrangebyscore(pool *CachePool, table, key string, min int, max int) (rets []string, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Do("ZRANGEBYSCORE", table+":"+key, min, max, "LIMIT", 0, 50))
	if err != nil && err != redis.ErrNil {
		logger.Error("Redis_zrangebyscore error: %s (%s, %s, %d ,%d)", err.Error(), table+":"+key, min, max)
	}
	if err == redis.ErrNil {
		logger.Error("redis_zrangebyscore ErrNil")
		rets = nil
		err = nil
	}

	return
}

func Redis_zrange(pool *CachePool, table, key string, begin int, end int) (rets []string, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	rets, err = redis.Strings(cache.Do("ZRANGE", table+":"+key, begin, end))
	if err != nil && err != redis.ErrNil {
		logger.Error("Redis_zrange error:", err.Error(), table+":"+key, begin, end)
	}
	if err == redis.ErrNil {
		logger.Error("redis_zrange: ErrNil")
		rets = nil
		err = nil
	}

	return
}

func Redis_hashSet(pool *CachePool, table string, key string, field string, value uint32) (err error) {
	cache := pool.Get()
	defer cache.Recycle()

	_, err = cache.Do("hset", table+DbTableKeySplit+key, field, value)
	if err != nil {
		logger.Error("hset error: %s (%s, %s, %s, %s)", err.Error(), table, key, field, value)
	}

	return
}

func Redis_hashGet(pool *CachePool, tableName, keyName, fieldName string) (number int64, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	number, err = redis.Int64(cache.Do("Hget", tableName+DbTableKeySplit+keyName, fieldName))
	if err != nil && err != redis.ErrNil {
		logger.Error("Get error: %s (get %s from %s)", err.Error(), tableName+DbTableKeySplit+keyName, keyName)
	}
	if err == redis.ErrNil {
		number = 0
		err = nil
	}

	return number, err
}

func Redis_hashLen(pool *CachePool, tableName, keyName string) (number int64, err error) {
	cache := pool.Get()
	defer cache.Recycle()

	number, err = redis.Int64(cache.Do("HLEN", tableName+DbTableKeySplit+keyName))
	if err != nil && err != redis.ErrNil {
		logger.Error("HLEN error: ", err.Error(), tableName, keyName)
	}
	if err == redis.ErrNil {
		number = 0
		err = nil
	}
	return number, err
}
