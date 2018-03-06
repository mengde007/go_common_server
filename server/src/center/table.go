package center

import (
	"common"
)

func (self *Center) del(table string, key string) (err error) {
	return common.Redis_del(self.maincache, table, key)
}

func (self *Center) setInt(table string, key string, value int) (err error) {
	return common.Redis_setInt(self.maincache, table, key, value)
}

func (self *Center) getInt(table string, key string) (value int, err error) {
	return common.Redis_getInt(self.maincache, table, key)
}

func (self *Center) setString(table string, key string, value string) (err error) {
	return common.Redis_setString(self.maincache, table, key, value)
}

func (self *Center) getString(table string, key string) (value string, err error) {
	return common.Redis_getString(self.maincache, table, key)
}

func (self *Center) setexpire(table string, key string, time string) (err error) {
	return common.Redis_setexpire(self.maincache, table, key, time)
}

func (self *Center) sadd(table string, key string, value string) (err error) {
	return common.Redis_sadd(self.maincache, table, key, value)
}

func (self *Center) srem(table string, key string, value string) (err error) {
	return common.Redis_srem(self.maincache, table, key, value)
}

func (self *Center) exists(table string, key string) bool {
	return common.Redis_exists(self.maincache, table, key)
}

func (self *Center) scard(table string, key string) (num int) {
	return common.Redis_scard(self.maincache, table, key)
}

func (self *Center) srandmember(table string, key string) (value string, err error) {
	return common.Redis_srandmember(self.maincache, table, key)
}

func (self *Center) zadd(table string, key string, value string, score uint32) (err error) {
	return common.Redis_zadd(self.maincache, table, key, value, score)
}

func (self *Center) zrem(table string, key string, value string) (err error) {
	return common.Redis_zrem(self.maincache, table, key, value)
}

func (self *Center) zcard(table string, key string) (uint32, error) {
	return common.Redis_zcard(self.maincache, table, key)
}

func (self *Center) zscore(table string, key string, value string) (uint32, error) {
	return common.Redis_zscore(self.maincache, table, key, value)
}

func (self *Center) zrevrange(table string, key string, start int, stop int) (rets []string, err error) {
	return common.Redis_zrevrange(self.maincache, table, key, start, stop)
}

func (self *Center) zrevrank(table string, key string, value string) (uint32, error) {
	return common.Redis_zrevrank(self.maincache, table, key, value)
}

/*func (self *Center) keys(table string, key string) (rets []string, err error) {
	return common.Redis_keys(self.maincache, table, key)
}*/

func (self *Center) hset(table string, key string, field string, value string) (err error) {
	return common.Redis_hset(self.maincache, table, key, field, value)
}

func (self *Center) hgetall(table string, key string) (rets []string, err error) {
	return common.Redis_hgetall(self.maincache, table, key)
}

//add for seach myself
func (self *Center) zrank(table, key, value string) (rank int, err error) {
	return common.Redis_zrank(self.maincache, table, key, value)
}

func (self *Center) delData(tableName, keyName string) error {
	return common.Redis_del(self.maincache, tableName, keyName)
}

func (self *Center) hgetValue(tableName string, keyName, fieldName string) (number int, err error) {
	return common.Redis_hGet(self.maincache, tableName, keyName, fieldName)
}

func (self *Center) hsetValue(table string, key string, field string, value string) error {
	return common.Redis_hset(self.maincache, table, key, field, value)
}

func (self *Center) hreName(tableName, oldKeyName, newKeyName string) error {
	return common.Redis_hRename(self.maincache, tableName, oldKeyName, newKeyName)
}
