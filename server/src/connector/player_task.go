package connector

import (
	// "bytes"
	"common"
	// "daerclient"
	// "encoding/json"
	// "io/ioutil"
	// "lockclient"
	// "logger"
	// "mailclient"
	// "net/http"
	// "rpc"
	// "runtime/debug"
	"strconv"
	// "strings"
	"logger"
	"rpc"
	"strings"
	"time"
	// "timer"
)

const (
	SIGNATURE       = iota
	SIG_PLAY_DAER   //玩大二
	SIG_PLAY_MJ     //玩麻将
	SIG_PLAY_POCKER //德州扑克
)

func (p *player) InitTask() {
	tasks := p.GetTasks()
	if tasks == nil {
		tasks = &rpc.DailyTask{}
		p.SetTasks(tasks)
	}

	now := uint32(time.Now().Unix())
	if !common.IsTheSameDay(now, uint32(tasks.GetResetTime()), 0) {
		tasks.SetResetTime(int32(now))
		tasks.DoneIds = []int32{}
		tasks.GetIds = []int32{}
		tasks.SetShares(int32(0))
		tasks.SetShareFris(int32(0))
		tasks.SetDaerTms(int32(0))
		tasks.SetMjTms(int32(0))
		tasks.SetPokerTms(int32(0))
		tasks.SetWinDaerTms(int32(0))
		tasks.SetWinMjTms(int32(0))
		tasks.SetWinPokerTms(int32(0))
	}
}

func (p *player) TaskTrigger(task int, win bool) {
	switch task {
	case SIGNATURE:
		p.TaskSignature()
		break
	case SIG_PLAY_DAER:
		p.TaskPlayDaer(win)
		break
	case SIG_PLAY_MJ:
		p.task_play_mj(win)
		break
	case SIG_PLAY_POCKER:
		p.task_play_pocker(win)
		break
	default:
		break
	}

}

func (p *player) TaskSignature() {

}

func (p *player) TaskShares(bShare bool) {
	tasks := p.GetTasks()
	if tasks == nil {
		logger.Error("TaskShares p.GetTasks() return nil")
		return
	}

	if bShare {
		tasks.SetShareFris(tasks.GetShareFris() + 1)
		cfg := GetTaskCfg("2")
		if cfg == nil {
			logger.Error("TaskShares GetTaskCfg(2) return nil")
			return
		}

		if tasks.GetShareFris() == cfg.Value {
			tasks.DoneIds = append(tasks.DoneIds, 2)
			// msg := &rpc.TaskFinishNofity{}
			// msg.SetTaskId("2")
			// WriteResult(p.conn, msg)
		}

	} else {
		tasks.SetShares(tasks.GetShares() + 1)
		cfg := GetTaskCfg("1")
		if cfg == nil {
			logger.Error("TaskShares GetTaskCfg(1) return nil")
			return
		}

		if tasks.GetShares() == cfg.Value {
			tasks.DoneIds = append(tasks.DoneIds, 1)
			// msg := &rpc.TaskFinishNofity{}
			// msg.SetTaskId("1")
			// WriteResult(p.conn, msg)
		}
	}
	WriteResult(p.conn, p.GetTasks())
}

func (p *player) TaskPlayDaer(win bool) {
	tasks := p.GetTasks()
	if tasks == nil {
		logger.Error("TaskPlayDaer p.GetTasks() return nil")
		return
	}

	if win {
		tasks.SetWinDaerTms(tasks.GetWinDaerTms() + 1)
		cfg := GetTaskCfg("6")
		if cfg == nil {
			logger.Error("TaskPlayDaer GetTaskCfg(6) return nil")
			return
		}

		if tasks.GetWinDaerTms() == cfg.Value {
			tasks.DoneIds = append(tasks.DoneIds, 6)
		}
	}

	//just for play
	tasks.SetDaerTms(tasks.GetDaerTms() + 1)
	cfg := GetTaskCfg("3")
	if cfg == nil {
		logger.Error("TaskPlayDaer GetTaskCfg(3) return nil")
		return
	}

	if tasks.GetDaerTms() == cfg.Value {
		tasks.DoneIds = append(tasks.DoneIds, 3)
	}
	WriteResult(p.conn, p.GetTasks())
}

func (p *player) task_play_mj(win bool) {
	tasks := p.GetTasks()
	if tasks == nil {
		logger.Error("task_play_mj p.GetTasks() return nil")
		return
	}

	if win {
		tasks.SetWinMjTms(tasks.GetWinMjTms() + 1)
		cfg := GetTaskCfg("7")
		if cfg == nil {
			logger.Error("task_play_mj GetTaskCfg(7) return nil")
			return
		}

		if tasks.GetWinMjTms() == cfg.Value {
			tasks.DoneIds = append(tasks.DoneIds, 7)
		}
	}

	//just for play
	tasks.SetMjTms(tasks.GetMjTms() + 1)
	cfg := GetTaskCfg("4")
	if cfg == nil {
		logger.Error("task_play_mj GetTaskCfg(4) return nil")
		return
	}

	if tasks.GetMjTms() == cfg.Value {
		tasks.DoneIds = append(tasks.DoneIds, 4)
	}
	WriteResult(p.conn, p.GetTasks())
}

func (p *player) task_play_pocker(win bool) {
	tasks := p.GetTasks()
	if tasks == nil {
		logger.Error("task_play_pocker p.GetTasks() return nil")
		return
	}

	//pocker have'
	// if win {
	// 	tasks.SetWinPokerTms(tasks.GetWinPokerTms() + 1)
	// 	cfg := GetTaskCfg("8")
	// 	if cfg == nil {
	// 		logger.Error("task_play_pocker GetTaskCfg(8) return nil")
	// 		return
	// 	}

	// 	if tasks.GetWinPokerTms() == cfg.Value {
	// 		tasks.DoneIds = append(tasks.DoneIds, 8)
	// 	}
	// }

	//just for play
	tasks.SetPokerTms(tasks.GetPokerTms() + 1)
	cfg := GetTaskCfg("5")
	if cfg == nil {
		logger.Error("task_play_pocker GetTaskCfg(5) return nil")
		return
	}

	if tasks.GetPokerTms() == cfg.Value {
		tasks.DoneIds = append(tasks.DoneIds, 5)
	}
	WriteResult(p.conn, p.GetTasks())
}

func (p *player) GetTaskRewards(taskId int32) {
	tasks := p.GetTasks()
	if tasks == nil {
		logger.Error("GetTaskRewards p.GetTasks() return nil")
		return
	}

	bDone := false
	for _, v := range tasks.DoneIds {
		if v == taskId {
			bDone = true
			break
		}
	}
	if !bDone {
		logger.Error("GetTaskRewards task not finished, taskId:%d", taskId)
		return
	}

	bGet := false
	for _, v := range tasks.GetIds {
		if v == taskId {
			bGet = true
			break
		}
	}
	if bGet {
		logger.Error("GetTaskRewards already get rewards, taskId:%d", taskId)
		return
	}

	//给奖励了
	cfg := GetTaskCfg(strconv.Itoa(int(taskId)))
	if cfg == nil {
		logger.Error("GetTaskRewards GetTaskCfg(%s)", taskId)
		return
	}

	rewards := strings.Split(cfg.Rewards, ",")
	if len(rewards) == 0 {
		return
	}

	for _, v := range rewards {
		idNum := strings.Split(v, ":")
		if len(idNum) != 2 {
			logger.Error("GetTaskRewards give rewards error, cfg.Rewards:%s", cfg.Rewards)
			return
		}
		num, _ := strconv.Atoi(idNum[1])
		p.AddResource(idNum[0], int32(num))

		logger.Info("===========num:%d, totall:%d", num, p.GetCoin())
	}
	p.GetTasks().GetIds = append(p.GetTasks().GetIds, taskId)
	p.ResourceChangeNotify()
	WriteResult(p.conn, p.GetTasks())
}
