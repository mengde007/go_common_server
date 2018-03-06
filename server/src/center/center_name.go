package center

import (
	"common"
	"errors"
	"proto"
	"sync"
)

type CenterNameService struct {
	mapPlayerName map[string]string
	lockPlayer    sync.Mutex
}

var pCenterNameService *CenterNameService

func StartCenterNameService() {
	pCenterNameService = &CenterNameService{
		mapPlayerName: make(map[string]string),
	}

	pCenterNameService.lockPlayer.Lock()
	defer pCenterNameService.lockPlayer.Unlock()

	//数据库取内容
	if kvs, err := centerServer.hgetall(common.SystemTableName, common.SystemKeyName_Name2Id); err == nil {
		for i := 0; i < len(kvs)/2; i++ {
			pCenterNameService.mapPlayerName[kvs[2*i]] = kvs[2*i+1]
		}
	}
}

func (self *Center) CheckPlayerName(req *proto.QueryName, reply *proto.QueryNameResult) (err error) {
	pCenterNameService.lockPlayer.Lock()
	defer pCenterNameService.lockPlayer.Unlock()

	if req.Name == "" {
		return errors.New("wrong name")
	}

	//只查询
	if !req.BQuery && req.Id == "" {
		return errors.New("wrong id")
	}

	if id, ok := pCenterNameService.mapPlayerName[req.Name]; ok {
		if req.BQuery {
			reply.Success = true
			reply.Id = id
		} else {
			reply.Success = false
			reply.Id = ""
		}
	} else {
		if req.BQuery {
			reply.Success = false
			reply.Id = ""
		} else {
			//这里存数据库
			if err := centerServer.hset(common.SystemTableName, common.SystemKeyName_Name2Id, req.Name, req.Id); err == nil {
				pCenterNameService.mapPlayerName[req.Name] = req.Id

				reply.Success = true
				reply.Id = ""
			} else {
				return err
			}
		}
	}

	return nil
}
