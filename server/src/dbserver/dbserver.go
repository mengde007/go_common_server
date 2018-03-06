package dbserver

import (
	"bytes"
	"common"
	"fmt"
	"logger"
	"net"
	"net/http"
	"proto"
	rpc "rpcplus"
	"runtime/debug"
)

type dbGroup map[uint32]*common.DbPool
type cacheGroup map[uint32]*common.CachePool

type DBServer struct {
	dbGroups    map[string]dbGroup
	dbNodes     map[string][]uint32
	cacheGroups map[string]cacheGroup
	cacheNodes  map[string][]uint32
	tables      map[string]*table
	exit        chan bool
}

func StartServices(self *DBServer, listener net.Listener) {
	rpcServer := rpc.NewServer()
	rpcServer.Register(self)

	rpcServer.HandleHTTP("/dbserver/rpc", "/debug/rpc")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("StartServices %s", err.Error())
			break
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info("DbServer Rpc Runtime Error: %s", r)
					debug.PrintStack()
				}
			}()
			rpcServer.ServeConn(conn)
			conn.Close()
		}()
	}
}

func WaitForExit(self *DBServer) {
	<-self.exit
	close(self.exit)
}

func NewDBServer(cfg common.DBConfig) (server *DBServer) {
	server = &DBServer{
		dbGroups:    map[string]dbGroup{},
		cacheGroups: map[string]cacheGroup{},
		tables:      map[string]*table{},
		exit:        make(chan bool),
	}

	http.Handle("/debug/state", debugHTTP{server})

	//初始化所有的db
	for key, pools := range cfg.DBProfiles {

		logger.Info("Init DB Profile %s", key)
		server.dbGroups = make(map[string]dbGroup)
		server.dbNodes = make(map[string][]uint32)

		temGroups := make(dbGroup)
		temDbInt := []uint32{}
		for _, poolCfg := range pools {
			logger.Info("Init DB %v", poolCfg)
			temGroups[poolCfg.NodeName] = common.NewDBPool(poolCfg)
			temDbInt = append(temDbInt, poolCfg.NodeName)
		}
		server.dbGroups[key] = temGroups
		common.BubbleSort(temDbInt) //排序节点
		server.dbNodes[key] = temDbInt
	}

	//初始化所有的cache
	for key, pools := range cfg.CacheProfiles {

		logger.Info("Init Cache Profile %s", key)
		server.cacheGroups = make(map[string]cacheGroup)
		server.cacheNodes = make(map[string][]uint32)

		temGroups := make(cacheGroup)
		temDbInt := []uint32{}
		for _, poolCfg := range pools {
			logger.Info("Init Cache %v", poolCfg)
			temGroups[poolCfg.NodeName] = common.NewCachePool(poolCfg)
			temDbInt = append(temDbInt, poolCfg.NodeName)

		}
		server.cacheGroups[key] = temGroups
		common.BubbleSort(temDbInt) //排序节点
		server.cacheNodes[key] = temDbInt

	}

	//初始化table
	for key, table := range cfg.Tables {

		logger.Info("Init Table: %s %v", key, table)
		server.tables[key] = NewTable(key, table, server)
	}

	return server
}

func (self *DBServer) Write(args *proto.DBWrite, reply *proto.DBWriteResult) error {
	if table, exist := self.tables[args.Table]; exist {
		err := table.write(args.Key, args.Value)
		if err != nil {
			return err
		}
		reply.Code = proto.Ok
	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *DBServer) Query(args *proto.DBQuery, reply *proto.DBQueryResult) error {

	if table, exist := self.tables[args.Table]; exist {
		rst, err := table.get(args.Key)
		if err != nil {
			return err
		}
		if rst != nil {
			reply.Value = rst
			reply.Code = proto.Ok
		} else {
			reply.Code = proto.NoExist
		}

	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *DBServer) Delete(args *proto.DBDel, reply *proto.DBDelResult) error {
	if table, exist := self.tables[args.Table]; exist {
		err := table.del(args.Key)
		if err != nil {
			return err
		}
		reply.Code = proto.Ok
	} else {
		reply.Code = proto.NoExist
	}

	return nil
}

func (self *DBServer) Quit(args *int32, reply *int32) error {
	self.exit <- true
	return nil
}

func (self *DBServer) statsJSON() string {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	fmt.Fprintf(buf, "{")
	for k, v := range self.tables {

		fmt.Fprintf(buf, "\n \"Table\": {")

		fmt.Fprintf(buf, "\n   \"Name\": \"%v\",", k)
		fmt.Fprintf(buf, "\n   \"States\": %v,", v.tableStats.String())
		fmt.Fprintf(buf, "\n   \"Rates\": %v,", v.qpsRates.String())

		fmt.Fprintf(buf, "\n }")
	}

	fmt.Fprintf(buf, "\n}")
	return buf.String()
}
