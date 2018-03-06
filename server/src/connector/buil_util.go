package connector

import (
	"common"
	// "logger"
)

const PLAYERLEVEL_MAX = 200

// 		//return (*cfg)[0].Deco
// 		return reflect.ValueOf(&(*cfg)[0]).Elem().FieldByName(fmt.Sprintf("Deco%d", level)).Interface().(uint32)
// 	case rpc.BuildingId_Bomb:

func GetGlobalCfg(key string) uint32 {
	return common.GetGlobalConfig(key)
}

func GetTaskCfg(id string) *TaskCfg {
	cfg, exist := gTaskCfg[id]
	if !exist {
		return nil
	}
	return &(*cfg)[0]
}

func GetUplevelCfg(id uint32) *UplevelCfg {
	cfg, exist := gUplevelCfg[id]
	if !exist {
		return nil
	}
	return &(*cfg)[0]
}

func CheckItemId(id string) bool {
	if _, ok := gItemCfg[id]; ok {
		return true
	}
	return false
}

func GetItemCfg(id string) *ItemCfg {
	cfg, exist := gItemCfg[id]
	if !exist {
		return nil
	}
	return &(*cfg)[0]
}

func GetAllItemIds() []string {
	ids := []string{}
	for id, _ := range gItemCfg {
		ids = append(ids, id)
	}
	return ids
}
