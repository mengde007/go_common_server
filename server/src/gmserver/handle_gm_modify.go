package gmserver

import (
	// "accountclient"
	"common"
	// "connector"
	"dbclient"
	"errors"
	// "logger"
	// "net/http"
	"rpc"
	// "strconv"
)

func modify_gem(uid string, absolute bool, gem_num int) (bool, int32, error) {
	var base rpc.PlayerBaseInfo
	var exist bool
	var err error
	if exist, err = dbclient.KVQueryBase(common.TB_t_base_playerbase, uid, &base); err != nil || !exist {
		return false, 0, errors.New("no player base info")
	}

	base.SetGem(base.GetGem() + int32(gem_num))
	if base.GetGem() < 0 {
		base.SetGem(int32(0))
	}

	_, err = dbclient.KVWriteBase(common.TB_t_base_playerbase, uid, &base)
	if err != nil {
		return false, 0, errors.New("save player extern info failed")
	}

	return true, base.GetGem(), nil
}

func modify_gold(uid string, absolute bool, gold_num int) (bool, int32, error) {
	var base rpc.PlayerBaseInfo
	var exist bool
	var err error
	if exist, err = dbclient.KVQueryBase(common.TB_t_base_playerbase, uid, &base); err != nil || !exist {
		return false, 0, errors.New("no player base info")
	}
	base.SetCoin(base.GetCoin() + int32(gold_num))
	if base.GetCoin() < 0 {
		base.SetCoin(int32(0))
	}

	_, err = dbclient.KVWriteBase(common.TB_t_base_playerbase, uid, &base)
	if err != nil {
		return false, 0, errors.New("save player base info failed")
	}

	return true, base.GetCoin(), nil
}
