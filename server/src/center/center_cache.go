/*
gc 缓存
*/

package center

import (
	"common"
	"logger"
	"proto"
)

const (
	CENTER_COST_MAIN = "center_cost_main"
)

func (self *Center) saveCost2Cache(req *proto.ReqCostRes) {
	logger.Info("saveCost2Cache has been called, req.PlayerList[0]:%s", req.PlayerList[0])

	saveBuf, err := common.GobEncode(req)
	if err != nil {
		logger.Error("saveCost2Cache GobEncode err", err)
		return
	}

	if err := common.Resis_setbuf(self.pCachePool, CENTER_COST_MAIN, req.PlayerList[0], saveBuf); err != nil {
		logger.Error("saveCost2Cache setbuf error", err)
		return
	}
}

func (self *Center) CheckCostFromCache(req *proto.GetCostCache, rst *proto.ReqCostRes) error {
	logger.Info("CheckCostFromCache has been called, req.PlayerList[0]:%s", req.Uid)

	buf, err := common.Resis_getbuf(self.pCachePool, CENTER_COST_MAIN, req.Uid)
	if err != nil {
		logger.Error("CheckCostFromCache common.Resis_getbuf err:%s", err)
		return nil
	}
	if buf == nil {
		return nil
	}

	if err := common.GobDecode(buf, rst); err != nil {
		logger.Error("CheckCostFromCache common.GobDecode err:%s", err)
		return nil
	}

	if err := common.Redis_del(self.pCachePool, CENTER_COST_MAIN, req.Uid); err != nil {
		logger.Error("CheckCostFromCache common.Redis_del err%s", err)
		return nil
	}
	return nil
}
