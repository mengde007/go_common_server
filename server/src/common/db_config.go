package common

import (
	"jscfg"
	"logger"
	"os"
	"path"
)

type MySQLConfig struct {
	Host        string  `json:"host"`
	Port        uint16  `json:"port"`
	Uname       string  `json:"uname"`
	Pass        string  `json:"pass"`
	NodeName    uint32  `json:"nodename"`
	Dbname      string  `json:"dbname"`
	Charset     string  `json:"charset"`
	PoolSize    uint16  `json:"pool"`
	IdleTimeOut float64 `json:"idle"`
	MaxRetry    uint8   `json:"retry"`
}

type CacheConfig struct {
	Host        string  `json:"host"`
	Port        uint16  `json:"port"`
	Index       uint8   `json:"index"`
	NodeName    uint32  `json:"nodename"`
	PoolSize    uint16  `json:"pool"`
	IdleTimeOut float64 `json:"idle"`
	MaxRetry    uint8   `json:"retry"`
	PassWord    string  `json:"pass"`
}

type TableConfig struct {
	DBProfile    string `json:"db-profile"`
	CacheProfile string `json:"cache-profile"`
	DeleteExpiry uint64 `json:"expiry"`
}

type DBConfig struct {
	DBHost        string
	DebugHost     string
	CpuProfile    bool
	GcTime        uint8
	DBProfiles    map[string][]MySQLConfig `json:"database"`
	CacheProfiles map[string][]CacheConfig `json:"cache"`
	Tables        map[string]TableConfig   `json:"tables"`
	IsCheck       bool
}

type LockServerCfg struct {
	LockHost      string
	DebugHost     string
	CpuProfile    bool
	GcTime        uint8
	CacheProfiles map[string][]CacheConfig `json:"cache"`
	Tables        map[string]TableConfig   `json:"tables"`
}

//读取配置表
func ReadDbConfig(file string, cfg *DBConfig) error {
	cfgpath, _ := os.Getwd()

	if err := jscfg.ReadJson(path.Join(cfgpath, file), cfg); err != nil {
		logger.Fatal("read Db config failed, %v", err)
		return err
	}

	return nil
}
