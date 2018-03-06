package dbtool

import (
	"common"
	"logger"
	"proto"
	//	"net/http"
)

type dbGroup map[uint32]*common.DbPool
type cacheGroup map[uint32]*common.CachePool
type dbName map[uint32]string

type DBServer struct {
	dbGroups    map[string]dbGroup
	dbNames     map[string]dbName
	dbNodes     map[string][]uint32
	cacheGroups map[string]cacheGroup
	cacheNodes  map[string][]uint32
	tables      map[string]*table
	exit        chan bool
}

func NewDBServerTool(cfg common.DBConfig) (server *DBServer) {
	server = &DBServer{
		dbGroups:    map[string]dbGroup{},
		cacheGroups: map[string]cacheGroup{},
		tables:      map[string]*table{},
		dbNames:     map[string]dbName{},
		exit:        make(chan bool),
	}
	//		http.Handle("/debug/state", debugHTTP{server})

	//初始化所有的db
	for key, pools := range cfg.DBProfiles {

		logger.Info("Init DB Profile %s", key)
		server.dbGroups = make(map[string]dbGroup)
		server.dbNodes = make(map[string][]uint32)
		server.dbNames = make(map[string]dbName)

		temGroups := make(dbGroup)
		tmpName := make(dbName)
		temDbInt := []uint32{}

		for _, poolCfg := range pools {
			logger.Info("Init DB %v", poolCfg)
			temGroups[poolCfg.NodeName] = common.NewDBPool(poolCfg)
			tmpName[poolCfg.NodeName] = poolCfg.Dbname
			temDbInt = append(temDbInt, poolCfg.NodeName)
		}
		server.dbGroups[key] = temGroups
		common.BubbleSort(temDbInt) //排序节点
		server.dbNodes[key] = temDbInt
		server.dbNames[key] = tmpName
		logger.Info("dbnodes:%s", temDbInt)
		logger.Info("dbgroups:%s", temGroups)
	}
	//初始化table
	for key, table := range cfg.Tables {
		logger.Info("Init Table: %s %v", key, table)
		server.tables[key] = NewTable(key, table, server, cfg.IsCheck)
	}

	return server
}

func (server *DBServer) AddHashId(query *proto.DBQuery, dbsize int) error {
	if table, exist := server.tables[query.Table]; exist {
		err := table.AddHash(dbsize)
		if err != nil {
			return err
		}
	}
	return nil
}

func MoveDataTrue(oldServer *DBServer, newServer *DBServer, query *proto.DBQuery) error {
	if table, exist := oldServer.tables[query.Table]; exist {
		if tablenew, existNew := newServer.tables[query.Table]; existNew {
			err := table.copydata(tablenew)
			if err != nil {
				return err
			}
		}
	} else {
		logger.Info("has no table:%s\n", query.Table)
	}
	return nil
}

func (oldServer *DBServer) MoveData(newServer *DBServer, query *proto.DBQuery) error {
	if table, exist := oldServer.tables[query.Table]; exist {
		if tablenew, existNew := newServer.tables[query.Table]; existNew {
			err := table.copydata(tablenew)
			if err != nil {
				return err
			}
		}
	} else {
		logger.Info("has no table:%s\n", query.Table)
	}
	return nil
}

func (oldServer *DBServer) DelOldData(newServer *DBServer, query *proto.DBQuery) error {
	if table, exist := oldServer.tables[query.Table]; exist {
		if tablenew, existNew := newServer.tables[query.Table]; existNew {
			err := table.deletedata(tablenew)
			if err != nil {
				return err
			}
		}
	} else {
		logger.Info("has no table:%s\n", query.Table)
	}
	return nil
}

func (oldServer *DBServer) ProccessData(dbDest *DBServer) error {
	tablename := "t_account_tencentid2playerid"
	if table, exist := oldServer.tables[tablename]; exist {
		err := table.decodePlayerData(dbDest)
		if err != nil {
			return err
		}
	} else {
		logger.Info("has no table:%s\n", tablename)
	}
	return nil
}
