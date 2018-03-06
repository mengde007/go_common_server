package common

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"logger"
	"net"
	"pools"
	"time"
)

type CachePool struct {
	*pools.RoundRobin
	defaultInstance *cacheInstance //默认的空连接，必须保证并发安全
}

type CreateCacheFunc func() (redis.Conn, error)

func NewCachePoolTool(cfg CacheConfig) (pool *CachePool) {
	pool = &CachePool{pools.NewRoundRobin(int(cfg.PoolSize),
		time.Duration(cfg.IdleTimeOut*1e9)),
		&cacheInstance{conn: nil, pool: pool, isClosed: true, isdefaultInstance: true}}

	if cfg.PoolSize == 0 {
		logger.Fatal("pool size == 0?")
	}

	pool.Open(CacheCreator(cfg))
	return pool
}

func NewCachePool(cfg CacheConfig) (pool *CachePool) {
	pool = &CachePool{pools.NewRoundRobin(int(cfg.PoolSize),
		time.Duration(cfg.IdleTimeOut*1e9)),
		&cacheInstance{conn: nil, pool: pool, isClosed: true, isdefaultInstance: true}}

	if cfg.PoolSize == 0 {
		logger.Fatal("pool size == 0?")
	}

	pool.Open(CacheCreator(cfg))
	tmp := make([]*cacheInstance, cfg.PoolSize)
	for i := uint16(0); i < cfg.PoolSize; i++ {
		tmp[i] = pool.Get()
	}
	for i := uint16(0); i < cfg.PoolSize; i++ {
		tmp[i].Recycle()
	}
	return pool
}

func (self *CachePool) Open(cacheFactory CreateCacheFunc) {
	if cacheFactory == nil {
		return
	}
	f := func() (pools.Resource, error) {
		c, err := cacheFactory()
		if err != nil {
			return nil, err
		}
		return &cacheInstance{c, self, false, false}, nil
	}
	self.RoundRobin.Open(f)
}

func (self *CachePool) Get() *cacheInstance {
	r, err := self.RoundRobin.Get()
	if err != nil {
		//如果创建失败了，就返回默认得空连接,RoundRobin不会加计数
		return self.defaultInstance
	}

	return r.(*cacheInstance)
}

type cacheInstance struct {
	conn              redis.Conn
	pool              *CachePool
	isClosed          bool
	isdefaultInstance bool
}

func (self *cacheInstance) Do(cmd string, args ...interface{}) (reply interface{}, err error) {
	if self.isdefaultInstance {
		return nil, errors.New("redis lost, this is default redisconn")
	}
	return self.conn.Do(cmd, args...)
}

func (self *cacheInstance) Send(cmd string, args ...interface{}) (err error) {
	if self.isdefaultInstance {
		return errors.New("redis lost, this is default redisconn")
	}
	return self.conn.Send(cmd, args...)
}

func (self *cacheInstance) Close() {
	if self.isdefaultInstance {
		return
	}
	self.conn.Close()
	self.isClosed = true
}

func (self *cacheInstance) IsClosed() bool {
	if self.isdefaultInstance {
		return true
	}
	return self.isClosed
}

func (self *cacheInstance) Recycle() {
	if self.isdefaultInstance {
		//如果是默认得空连接，不能放回pool
		return
	}

	if self.conn.Err() != nil {
		self.Close()
	}
	self.pool.Put(self)
}

func CacheCreator(cfg CacheConfig) CreateCacheFunc {
	addrs, err := net.LookupHost(cfg.Host)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	if len(addrs) < 1 {
		logger.Fatal("no redis ip !!!!!!!!!")
	}

	for _, s := range addrs {
		fmt.Println("Domain Name :", cfg.Host, s)
	}

	dns := fmt.Sprintf("%s:%d", addrs[0], cfg.Port)

	return func() (c redis.Conn, err error) {
		var retry uint8 = 0

		for {
			c, err = redis.Dial("tcp", dns)
			if err == nil {
				if cfg.PassWord != "" {
					if _, err = c.Do("AUTH", cfg.PassWord); err != nil {
						c.Close()
						logger.Error("Error on auth redis: %s", err.Error())

						return
					}
				}
				break
			}

			logger.Error("Error on Create redis: %s; try: %d/%d", err.Error(), retry, cfg.MaxRetry)

			if retry >= cfg.MaxRetry {
				return
			}

			retry++
		}

		_, err = c.Do("SELECT", cfg.Index)

		return
	}
}
