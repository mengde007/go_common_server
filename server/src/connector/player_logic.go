package connector

import (
	"logger"
	"payclient"
	"rpc"
	"strconv"
	"strings"
	"time"
)

func (p *player) AddItem2Bag(itemId string, num int32) {
	logger.Info("*******AddItem2Bag资源通知 itemId：%s, num:%d", itemId, num)
	bRes := false
	if itemId == "1" || itemId == "coin" {
		p.SetCoin(p.GetCoin() + num)
		bRes = true
	} else if itemId == "2" || itemId == "gem" {
		p.SetGem(p.GetGem() + num)
		bRes = true
	}
	if bRes {
		msg := &rpc.ResourceNotify{}
		msg.SetCoin(p.GetCoin())
		msg.SetGem(p.GetGem())
		WriteResult(p.conn, msg)
		logger.Info("*******资源通知")
		return
	}

	bFind := false
	for _, v := range p.Items {
		if itemId == v.GetId() {
			v.SetNum(v.GetNum() + num)
			bFind = true
			break
		}
	}

	if !bFind {
		itm := &rpc.BagItem{}
		itm.SetId(itemId)
		itm.SetNum(num)
		p.Items = append(p.Items, itm)
	}
	msg := &rpc.BagItemNofity{}
	msg.Items = p.Items
	WriteResult(p.conn, msg)
}

func (p *player) GetItemNum(itemId string) int32 {
	if itemId == "1" || itemId == "coin" {
		return p.GetCoin()
	} else if itemId == "2" || itemId == "gem" {
		return p.GetGem()
	}

	for _, v := range p.Items {
		if itemId == v.GetId() {
			return v.GetNum()
		}
	}
	return 0
}

func (p *player) CostItem2Bag(itemId string, num int32) bool {
	bFind := false
	for index, v := range p.Items {
		if itemId == v.GetId() {
			if v.GetNum() > num {
				v.SetNum(v.GetNum() - num)
			} else if v.GetNum() == num {
				p.Items = append(p.Items[:index], p.Items[index+1:]...)
			} else {
				logger.Error("CostItem2Bag not enough, id:%s, cur:%d, need:%d", itemId, v.GetNum(), num)
				return false
			}
			bFind = true
			break
		}
	}
	msg := &rpc.BagItemNofity{}
	msg.Items = p.Items
	WriteResult(p.conn, msg)
	return bFind
}

func (p *player) check_has_recharge() {
	openId := p.mobileqqinfo.Openid
	logger.Info("check_has_recharge openId:%s", openId)
	itemid, _ := payclient.QueryPayInfos(openId)
	if itemid == "" {
		return
	}

	msg := &rpc.PayResultNotify{}
	msg.SetResult(true)
	msg.SetPartnerId(itemid)
	p.OnRecharged(msg)
}

//充值成功
func (p *player) OnRecharged(msg *rpc.PayResultNotify) {
	logger.Info("OnRecharged 充值成功,itemid:%s", msg.GetPartnerId())
	if msg.GetResult() {
		itemid := msg.GetPartnerId()
		cfg := GetItemCfg(itemid)
		if cfg == nil {
			logger.Error("OnRecharged common.GetItemCfg(:%d) return nil", itemid)
			return
		}

		if cfg.VipCard > int32(0) {
			p.SetVipLeftDay(p.GetVipLeftDay() + int32(cfg.VipCard))
			p.SetVipOpenTime(int32(time.Now().Unix()))
			msg.SetVipDay(p.GetVipLeftDay())
		} else {
			itemIdNum := strings.Split(cfg.BuyAddID, "_")
			if len(itemIdNum) != 2 {
				logger.Error("OnRecharged cfg.BuyAddID err:%s", cfg.BuyAddID)
				return
			}

			cnt, _ := strconv.Atoi(itemIdNum[1])
			p.AddItem2Bag(itemIdNum[0], int32(cnt))
		}

		openId := p.mobileqqinfo.Openid
		payclient.DeletePayInfo(openId)
	}
	WriteResult(p.conn, msg)
}

func (p *player) Shopping(idnum []string) {
	logger.Info("Shopping called....idnum:%v", idnum)

	cfg := GetItemCfg(idnum[0])
	if cfg == nil {
		logger.Error("ReqPrePay common.GetItemCfg(:%d) return nil", idnum[0])
		return
	}

	arrs := strings.Split(cfg.BuyPrice, "_")
	if len(arrs) != 2 {
		logger.Error("Shopping err cfg.BuyPrice:%s", cfg.BuyPrice)
		return
	}
	if arrs[0] == "3" {
		logger.Error("Shopping 此道具为付费道具")
		return
	}

	price, _ := strconv.Atoi(arrs[1])
	num, _ := strconv.Atoi(idnum[1])
	price = price * num

	if p.GetItemNum(arrs[0]) < int32(price) {
		logger.Error("Shopping not enought money to buy")
		return
	}
	if cfg.BuyAddID == "" {
		logger.Error("Shopping cfg.BuyAddID is empty string")
		return
	}

	p.CostResource(arrs[0], int32(price))
	itemIdNum := strings.Split(cfg.BuyAddID, "_")
	if len(itemIdNum) != 2 {
		logger.Error("shopping cfg.BuyAddID err:%s", cfg.BuyAddID)
		return
	}

	cnt, _ := strconv.Atoi(itemIdNum[1])
	cnt *= num
	p.AddItem2Bag(itemIdNum[0], int32(cnt))
}
