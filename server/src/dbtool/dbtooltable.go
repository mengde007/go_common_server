package dbtool

import (
	gp "code.google.com/p/goprotobuf/proto"
	"code.google.com/p/snappy-go/snappy"
	"common"
	"connector"
	"database/sql"
	"fmt"
	"hash/crc32"
	"io"
	"logger"
	"os"
	"proto"
	"rpc"
	"stats"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	keylen = 64
)

type table struct {
	name         string
	caches       cacheGroup
	dbs          dbGroup
	dbn          dbName
	deleteExpiry uint64
	tableStats   *stats.Timings
	qpsRates     *stats.Rates
	cacheNode    []uint32
	dbNode       []uint32
	isCheck      bool
}

func makeHash(key string) uint32 {
	ieee := crc32.NewIEEE()
	io.WriteString(ieee, key)
	return ieee.Sum32()
}

func NewTable(name string, cfg common.TableConfig, db *DBServer, IsCheck bool) (t *table) {
	var (
		caches    cacheGroup
		cacheNode []uint32
		dbs       dbGroup
		dbNode    []uint32
		dbn       dbName
	)

	if cfg.DBProfile != "" {
		var exist bool
		if dbs, exist = db.dbGroups[cfg.DBProfile]; !exist {
			logger.Fatal("NewTable: table db profile not found: %s", cfg.DBProfile)
		}
		dbn, _ = db.dbNames[cfg.DBProfile]

		dbNode, _ = db.dbNodes[cfg.DBProfile]
		for _, dbpool := range dbs {
			db := dbpool.Get()

			query := fmt.Sprintf(`
			    CREATE TABLE IF NOT EXISTS %s (
				    id BINARY(64) NOT NULL PRIMARY KEY,
					hash_id BINARY(32) NOT NULL,
					auto_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
				    body MEDIUMBLOB,
				    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				    KEY (updated),
				    key (auto_id)
				    ) ENGINE=InnoDB;
			    `, name)

			logger.Info("CreateQuery :%s", query)
			rst, err := db.Exec(
				query,
			)

			if err != nil {
				logger.Fatal("NewTable: db %v create table %s faild! %s", dbpool, name, err.Error())
			}
			logger.Info("NewTable: db %v init %s: %v", dbpool, name, rst)
			db.Recycle()
		}
	}

	if caches == nil && dbs == nil {
		logger.Fatal("NewTable: table %s need a save func", name)
	}

	queryStats := stats.NewTimings("")
	qpsRates := stats.NewRates("", queryStats, 20, 10e9)
	return &table{
		name, caches, dbs, dbn,
		cfg.DeleteExpiry,
		queryStats,
		qpsRates,
		cacheNode,
		dbNode,
		IsCheck,
	}
}

func (dbServer *DBServer) GetVillageId(query *proto.DBQuery, key string) {
	if table, exist := dbServer.tables[query.Table]; exist {
		var index = 0
		hid := makeHash(key)
		for k, v := range table.dbNode {
			if hid < v {
				index = k
				break
			}
		}

		didx, ok := table.dbs[table.dbNode[index]]
		if !ok {
			logger.Fatal("getDbNode error: no find  (%d)", key)
		}
		db := didx.Get()
		defer db.Recycle()

		var rows *sql.Rows
		var ret []byte
		var err error
		rows, err = db.Query("SELECT body from "+table.name+" where id = CAST(? as BINARY(64)) LIMIT 1;", key)
		if err != nil {
			logger.Error("get error: %s (%s, %v)", err.Error(), key, rows)
			return
		}
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&ret)
			if err != nil {
				logger.Error("get scan error %s (%s)", err.Error(), key)
			}
			var dst []byte
			dst, err = snappy.Decode(nil, ret)

			if err != nil {
				logger.Error("KVQuery Unmarshal Error On snappy.Decode %s : %s (%s)", table, key, err.Error())
				return
			}
			var base rpc.PlayerBaseInfo
			unmarshalData(dst, &base)
			println(base.GetVillageId())
		}

	} else {
		logger.Info("has no table:%s\n", query.Table)
	}
}

func unmarshalData(dst []byte, value gp.Message) {
	if err := gp.Unmarshal(dst, value); err != nil {
		logger.Error("has no table:%s\n", err)
	}
}

func (self *table) AddHash(dbsize int) error {
	defer self.tableStats.Record("get", time.Now())
	for i := 0; i < dbsize; i++ {
		node, ok := self.dbs[self.dbNode[i]]
		if !ok {
			logger.Fatal("get node error!!")
		}

		db := node.Get()
		defer db.Recycle()
		query := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN hash_id BINARY(32) NOT NULL AFTER id;`, self.name)
		rst, err := db.Exec(
			query,
		)
		if err != nil {
			logger.Fatal("table:%s add new column hashid faild!sql:%s[rst:%v]\n", self.name, query, rst)
		}

		var count uint = 0
		rows, err := db.Query("SELECT count(*) from " + self.name)
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				logger.Error("get scan error %s ", err.Error())
			}
		}
		var countNum uint
		for countNum = 0; countNum < count; countNum = countNum + 10 {
			logger.Info("count: %d;num: (%d)", count, countNum)
			rows, err := db.Query("SELECT id,auto_id from "+self.name+" LIMIT ?,10;", countNum)
			if err != nil {
				logger.Error("get error: %s (%s, %v)", err.Error(), err, rows)
				return err
			}
			defer rows.Close()
			var key string
			var auto_id int
			for rows.Next() {
				err = rows.Scan(&key, &auto_id)
				hashid := makeHash(key)
				logger.Info("hashid:%d,key:%s,count:%d,countNum:%d,auto_id:%d", hashid, key, count, countNum, auto_id)
				//rst, err := db.Exec("UPDATE "+ self.name + " SET hash_id=? WHERE id=CAST(? as BINARY(64));",hashid,key)
				rst, err := db.Exec("UPDATE "+self.name+" SET hash_id=? WHERE auto_id=?;", hashid, auto_id)
				logger.Error("err error:  (%s, %v)", key, rst)
				if err != nil {
					logger.Error("get error: %s (%s, %v)", err.Error(), err, rst)
					return err
				}
			}
		}
	}
	return nil
}

func (self *table) deletedata(newtable *table) error {
	var olddbsize = len(self.dbNode)

	for i := 0; i < olddbsize; i++ {
		oldnode, ok := self.dbs[self.dbNode[i]]
		if !ok {
			logger.Info("get old node err!!")
		}
		var rows *sql.Rows
		db := oldnode.Get()

		//get count(*)
		var count uint
		rows, err := db.Query("SELECT count(*) from " + self.name)
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				logger.Error("get scan error %s ", err.Error())
			}
		}
		var countNum uint
		for countNum = 0; countNum < count; countNum = countNum + 10 {
			logger.Info("count: %d;num: (%d)", count, countNum)
			rows, err := db.Query("SELECT id,hash_id,body from "+self.name+" LIMIT ?,10;", countNum)
			if err != nil {
				logger.Error("get error: %s (%s, %v)", err.Error(), err, rows)
				return err
			}

			defer rows.Close()
			for rows.Next() {
				var body []byte
				var key string
				var hash_id string
				err = rows.Scan(&key, &hash_id, &body)
				//logger.Info("get key:%s;hashid:%s", key, hash_id)
				if err != nil {
					logger.Error("get scan error %s ", err.Error())
				}
				tmp_s_hid := strings.Trim(hash_id, "\x00")
				tmp_hid, e := strconv.ParseInt(tmp_s_hid, 10, 64)
				if e != nil {
					logger.Error("parse unit error:%v,tmp_hid:%d,%s", e, tmp_hid, tmp_s_hid)
				}
				hid := uint32(tmp_hid)
				var newidx = 0
				for k, v := range newtable.dbNode {
					if hid < v {
						newidx = k
						break
					}
				}

				//处在相同的块，不需要移动
				if newidx == i {
					continue
				}

				node_new, ok2 := newtable.dbs[newtable.dbNode[newidx]]
				if !ok2 {
					logger.Fatal("getDbNode node_new error: no find  ()")
				}

				//logger.Info("data %d", 1)
				chgdb := node_new.Get()
				//logger.Info("data %d" , 2)
				//db_new.Recycle()

				_, reterr := chgdb.Exec("INSERT ignore INTO "+newtable.name+" (id, hash_id, body) values(?, ?, ?);", key, hid, body)
				if reterr != nil {
					logger.Error("write error: %s (%s, %v)", reterr.Error(), key, body)
				} else {
					_, err = db.Exec("DELETE from "+self.name+" where id = CAST(? as BINARY(64));", key)
					if err != nil {
						logger.Error("delete error: %s (%s, %v)", reterr.Error(), key, body)
					} else {
						logger.Info("move data:%d-->%d,hashid:%d", i, newidx, hid)
					}
				}
				chgdb.Close()
			}
		}
		db.Close()
	}
	return nil
}

func (self *table) copydata(newtable *table) error {
	olddbsize := len(self.dbNode)

	for i := 0; i < olddbsize; i++ {
		logger.Info("node:%d", i)
		oldnode, ok := self.dbs[self.dbNode[i]]
		if !ok {
			logger.Info("get old node err!!")
		}
		var rows *sql.Rows
		db := oldnode.Get()

		//get count(*)
		var count uint
		rows, err := db.Query("SELECT count(*) from " + self.name)
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				logger.Error("get scan error %s ", err.Error())
			}
		}
		var countNum uint
		var countStay uint = 0
		for countNum = 0; countNum < count; countNum = countNum + 10 {
			logger.Info("count: %d;num: (%d)", count, countNum)
			rows, err := db.Query("SELECT id,hash_id,body from "+self.name+" LIMIT ?,10;", countNum)
			if err != nil {
				logger.Error("get error: %s (%s, %v)", err.Error(), err, rows)
				rows.Close()
				db.Recycle()
				return err
			}

			for rows.Next() {
				var body []byte
				var key string
				var hash_id string
				err = rows.Scan(&key, &hash_id, &body)
				if err != nil {
					logger.Error("get scan error %s ", err.Error())
				}
				tmp_s_hid := strings.Trim(hash_id, "\x00")
				tmp_hid, e := strconv.ParseInt(tmp_s_hid, 10, 64)
				if e != nil {
					logger.Error("parse unit error:%v,tmp_hid:%d,%s", e, tmp_hid, tmp_s_hid)
				}
				hid := uint32(tmp_hid)
				var newidx = 0
				for k, v := range newtable.dbNode {
					if hid < v {
						newidx = k
						break
					}
				}
				//logger.Info("hashid:%s;newid:%d;oldid:%d", hash_id,newidx,i)

				//处在相同的块，不需要移动
				if newtable.dbn[newtable.dbNode[newidx]] == self.dbn[self.dbNode[i]] {
					countStay = countStay + 1
					continue
				}

				node_new, ok2 := newtable.dbs[newtable.dbNode[newidx]]
				if !ok2 {
					logger.Fatal("getDbNode node_new error: no find  ()")
				}

				chgdb := node_new.Get()

				_, reterr := chgdb.Exec("INSERT ignore INTO "+newtable.name+" (id, hash_id, body) values(?, ?, ?);", key, hid, body)
				if reterr != nil {
					logger.Error("write error: %s (%s, %v)", reterr.Error(), key, body)
				} else {
					_, err = db.Exec("DELETE from "+self.name+" where id = CAST(? as BINARY(64));", key)
					if err != nil {
						logger.Error("delete error: %s (%s, %v)", reterr.Error(), key, body)
					} else {
						countNum = countNum - 1
						logger.Info("move data:%s-->%s,hashid:%d\n", self.dbn[self.dbNode[i]], newtable.dbn[newtable.dbNode[newidx]], hid)
						fmt.Fprintf(os.Stdin, "move data:%s-->%s,hashid:%d\n", self.dbn[self.dbNode[i]], newtable.dbn[newtable.dbNode[newidx]], hid)
					}
				}
				chgdb.Recycle()
			}
			rows.Close()
		}
		if newtable.isCheck {
			var countStayReal uint
			rows, err := db.Query("SELECT count(*) from " + self.name)
			if err != nil {
				logger.Error("get error: %s (%s, %v)", err.Error(), err, rows)
				db.Recycle()
				return err
			}
			for rows.Next() {
				err = rows.Scan(&countStayReal)
				if err != nil {
					logger.Error("get scan error %s ", err.Error())
				}
			}
			if countStayReal != countStay {
				logger.Error("move db:%s error, countStayReal:%d,countStay:%d", self.dbn[self.dbNode[i]], countStayReal, countStay)
				fmt.Fprintf(os.Stdin, "move db:%s error, countStayReal:%d,countStay:%d\n", self.dbn[self.dbNode[i]], countStayReal, countStay)
				db.Recycle()
				return nil
			} else {
				logger.Info("move db:%s sucess, countStayReal:%d,countStay:%d", self.dbn[self.dbNode[i]], countStayReal, countStay)
				fmt.Fprintf(os.Stdin, "move db:%s sucess, countStayReal:%d,countStay:%d\n", self.dbn[self.dbNode[i]], countStayReal, countStay)
			}
		}
		db.Recycle()
	}
	return nil
}

func getplayerbase(dst []byte, value gp.Message) {
	if err := gp.Unmarshal(dst, value); err != nil {
		logger.Error("has no table:%s\n", err)
	}
}

func getplayerextern(dst []byte, value gp.Message) {
	if err := gp.Unmarshal(dst, value); err != nil {
		logger.Error("has no table:%s\n", err)
	}
}

func playerbase(dst []byte, destTable *table) error {
	/*destnode, ok2 := destTable.dbs[destTable.dbNode[0]] // dest库只有一个节点
	if !ok2 {
		logger.Fatal("getDbNode node_new error: no find  ()")
		return nil
	}

	destdb := destnode.Get()
	_, reterr := destdb.Exec("INSERT ignore INTO "+destTable.name+" (id, body) values(?, ?);", key, body)
	if reterr != nil {
		logger.Error("write error: %s (%s, %v)", reterr.Error(), key, body)
		return reterr
	}

	destdb.Recycle()	*/

	return nil
}

func GetPVEStopIDByDIFF(pve *rpc.PveStages, curDifLevel uint32) uint32 {
	CurID := uint32(0)
	if curDifLevel >= connector.SD_Betray {
		return CurID
	}
	StageDifLevel := ToGetCurSD(curDifLevel)

	if pve != nil {
		for _, stage := range pve.Stages {
			curMostDS := len(stage.StageDatas) - 1
			if curMostDS >= int(StageDifLevel) {
				CurID++
			}
		}
	}

	return CurID
}

func ToGetCurSD(curDifLevel uint32) uint32 {
	CurSD := uint32(0)
	switch curDifLevel {
	case 0:
		CurSD = connector.SD_Normal
		break
	case 1:
		CurSD = connector.SD_Hard
		break
	case 2:
		CurSD = connector.SD_Nightmare
		break
	case 3:
		CurSD = connector.SD_Betray
		break
	default:
		CurSD = connector.SD_Error
	}

	return CurSD
}

func (self *table) decodePlayerData(dbDest *DBServer) error {

	//配置表
	connector.LoadConfigFiles(common.GetDesignerDir())
	common.LoadGlobalConfig()

	desttable := "t_player_data"

	destTable, exist := dbDest.tables[desttable]
	if !exist {
		logger.Error("has no destTable:%s\n", desttable)
		return nil
	}

	destnode := destTable.dbs[destTable.dbNode[0]] // dest 只有一个node节点

	wg := &sync.WaitGroup{}
	srcdbsize := len(self.dbNode)

	for i := 0; i < srcdbsize; i++ {
		logger.Info("node:%d", i)
		srcnode, ok := self.dbs[self.dbNode[i]]
		if !ok {
			logger.Error("get src node err!!")
		}
		var rows *sql.Rows
		db := srcnode.Get()

		//get count(*)
		var count uint
		rows, err := db.Query("SELECT count(*) from " + self.name)
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				logger.Error("get scan error %s ", err.Error())
			}
		}
		var countNum uint
		for countNum = 0; countNum < count; countNum = countNum + 100 {
			logger.Info("count: %d;num: (%d)", count, countNum)
			rows, err := db.Query("SELECT  id, relateid from "+self.name+" LIMIT ?,100;", countNum)

			// 遍历绑定的账号
			if err != nil {
				logger.Error("get error: %s (%s, %v)", err.Error(), err, rows)
				rows.Close()
				db.Recycle()
				return err
			}

			for rows.Next() {
				var openid string
				var uid string
				err = rows.Scan(&openid, &uid)
				if err != nil {
					logger.Error("get scan error %s ", err.Error())
				}

				wg.Add(1) //监听client要算一个
				go func(openid string, uid string) {
					defer wg.Done()
					//去掉前后的空格
					uid = strings.TrimRight(uid, string(byte(0)))

					/*openid = strings.TrimRight(openid, string(byte(0)))
					retopenid, pf, errnew := parsePartnerId(openid)
					if errnew != nil {
						logger.Error("parsePartnerId error %s  %s", errnew.Error(), openid)
						continue
					}

					uid, err = QueryPlayerIdByPartnerId(
						common.TB_t_account_tencentid2playerid,
						retopenid, pf)*/

					var p rpc.PlayerBaseInfo
					exists, err := KVQueryBase(common.TB_t_base_playerbase, uid, &p)
					if err != nil {
						logger.Error("KVQueryBase TB_t_base_playerbase error %s ", err.Error())
						return
					}

					if !exists {
						logger.Error("KVQueryBase TB_t_base_playerbase !exists %s", uid)
						return
					}

					var extra rpc.PlayerExtraInfo
					exists, err = KVQueryExt(common.TB_t_ext_playerextra, uid, &extra)
					if err != nil {
						logger.Error("KVQueryExt TB_t_ext_playerextra error %s ", err.Error())
						return
					}

					if !exists {
						logger.Error("KVQueryExt TB_t_ext_playerextra !exists %s", uid)
						return
					}
					var taskInfo string
					if extra.Tasks != nil {
						for index, task := range extra.Tasks {
							if index > 0 {
								taskInfo += ":"
							}
							taskInfo += fmt.Sprintf("%s,%d", task.GetName(), task.GetFinishedTime())
						}
					}

					var v rpc.VillageInfo
					vid := strconv.FormatUint(p.GetVillageId(), 16)
					exists, err = KVQueryExt(common.TB_t_ext_village, vid, &v)
					if err != nil {
						logger.Error("KVQueryExt TB_t_ext_village error %s ", err.Error())
						return
					}

					if !exists {
						logger.Error("KVQueryExt TB_t_ext_village !exists %s", vid)
						return
					}

					vi := &connector.Village{Vid: p.GetVillageId(), VillageInfo: &v}
					vi.Buildings_Init()

					//logger.Info("village info %v", v)

					curGold, _ := vi.GetGoldStorage()
					curFood, _ := vi.GetFoodStorage()
					curWuhun, _ := vi.GetWuhunStorage()

					var Normal uint32
					var Hard uint32
					var Nightmare uint32

					var pve rpc.PveStages
					if exists, err = KVQueryExt(common.TB_t_ext_pve, uid, &pve); err != nil {
						logger.Error("query pve failed! %s", err.Error())
					}

					if exists {
						Normal = GetPVEStopIDByDIFF(&pve, connector.SD_Normal)
						Hard = GetPVEStopIDByDIFF(&pve, connector.SD_Hard)
						Nightmare = GetPVEStopIDByDIFF(&pve, connector.SD_Nightmare)
					}

					var heroret string
					heroContainer := v.Center.GetHeroContainer()
					if nil != heroContainer {
						for n, hero := range heroContainer.Heroes {
							id := hero.GetCharacter().GetType()
							level := hero.GetCharacter().GetLevel()

							if n > 0 {
								heroret += ":"
							}
							heroret += fmt.Sprintf("%d,%d", id, level)
						}
					}

					destdb := destnode.Get()
					_, err = destdb.Exec(`insert into t_player_data (OpenId, Uid, CharName, OfficialLevel, OfficialExp, Trophy, CenterLevel, PveNormalStage,PveHardStage,
					PveNightmareStage,Gold,Food,ZiJin,Clan,Heros,TaskInfo,VillageId) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE  Uid=?, CharName=?, OfficialLevel=?, 
					OfficialExp=?, Trophy=?, CenterLevel=?, PveNormalStage=?,PveHardStage=?,
					PveNightmareStage=?,Gold=?,Food=?,ZiJin=?,Clan=?,Heros=?,TaskInfo=?,VillageId=?;`,
						openid,
						uid,
						p.GetName(),
						p.GetOfficialTitleLevel(),
						p.GetOfficialTitleExp(),
						p.GetTrophy(),
						v.GetCenter().GetLevel(),
						Normal,
						Hard,
						Nightmare,
						curGold,
						curFood,
						curWuhun,
						p.GetClan(),
						heroret,
						taskInfo,
						vid,
						uid,
						p.GetName(),
						p.GetOfficialTitleLevel(),
						p.GetOfficialTitleExp(),
						p.GetTrophy(),
						v.GetCenter().GetLevel(),
						Normal,
						Hard,
						Nightmare,
						curGold,
						curFood,
						curWuhun,
						p.GetClan(),
						heroret,
						taskInfo,
						vid)

					if err != nil {
						logger.Error("insert into t_player_data error: %s (%s, %v)", err.Error())
					}

					destdb.Recycle()

					/*logger.Info("player data: %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v",
					openid,
					uid,
					p.GetName(),
					p.GetOfficialTitleLevel(),
					p.GetOfficialTitleExp(),
					p.GetTrophy(),
					v.GetCenter().GetLevel(),
					Normal,
					Hard,
					Nightmare,
					curGold,
					curFood,
					curWuhun,
					p.GetClan(),
					heroret)*/
				}(openid, uid)
			}
			rows.Close()
		}
		db.Recycle()
	}
	wg.Wait()
	return nil
}
