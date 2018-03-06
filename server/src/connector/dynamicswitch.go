package connector

// import (
// 	"common"
// 	"csvcfg"
// 	"logger"
// 	"path"
// 	"rpc"
// 	"time"
// 	"timer"
// )

// //动态开关
// type StDynamicSwitch struct {
// 	Types   uint32
// 	Version uint32
// 	Value   uint32
// 	Sync    uint32
// }

// var gDynamicSwitch map[uint32]*[]StDynamicSwitch
// var gSyncMsg *rpc.DynamicSwitchs

// func init() {
// 	filename := path.Join(common.GetDesignerDir(), "DynamicSwitch.csv")

// 	checkDynamicSwitch(filename)

// 	tm := timer.NewTimer(time.Second * 30)
// 	tm.Start(func() {
// 		checkDynamicSwitch(filename)
// 	})
// }

// func checkDynamicSwitch(filename string) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			logger.Error("checkDynamicSwitch failed", r)
// 		}
// 	}()

// 	m := make(map[uint32]*[]StDynamicSwitch)
// 	csvcfg.LoadCSVConfig(filename, &m)

// 	//第一次直接赋值就可以了，否则要判断版本号
// 	if gDynamicSwitch != nil {
// 		for types, arrOld := range gDynamicSwitch {
// 			stOld := (*arrOld)[0]

// 			if arrNew, ok := m[types]; ok {
// 				stNew := (*arrNew)[0]
// 				if stNew.Version != stOld.Version {
// 					go OnDynamicSwitchChanged(types, stNew.Version, stNew.Value)
// 				}
// 			}
// 		}
// 	}

// 	gDynamicSwitch = m

// 	//生成下发消息
// 	genDynamicSwitchsMsg()
// }

// func genDynamicSwitchsMsg() {
// 	msg := &rpc.DynamicSwitchs{}

// 	for types, arr := range gDynamicSwitch {
// 		st := (*arr)[0]

// 		info := &rpc.DynamicSwitch{}

// 		//对于大于等于太守系统的开关，采用数字型type传输，以兼容客户端老版本
// 		if rpc.DynamicSwitch_Type(types) >= rpc.DynamicSwitch_Offical {
// 			info.SetType(rpc.DynamicSwitch_Hero)
// 			info.SetUtype(types)
// 		} else {
// 			info.SetType(rpc.DynamicSwitch_Type(types))
// 		}

// 		info.SetValue(st.Value)
// 		msg.Switchs = append(msg.Switchs, info)
// 	}

// 	gSyncMsg = msg
// }

// //状态改变
// func OnDynamicSwitchChanged(types, version, value uint32) {
// 	logger.Info("OnSwitchChanged:", types, version, value)

// 	defer func() {
// 		if r := recover(); r != nil {
// 			logger.Error("OnSwitchChanged failed", r)
// 		}
// 	}()

// 	switch rpc.DynamicSwitch_Type(types) {
// 	case rpc.DynamicSwitch_Hero:
// 		LoadHeroCfg()
// 	case rpc.DynamicSwitch_Pve:
// 		ReLoadPVEGenneralCfg()
// 	}
// }

// func syncDynamicSwitchsMsg(conn rpc.RpcConn) {
// 	if gSyncMsg != nil {
// 		WriteResult(conn, gSyncMsg)
// 	}
// }

// func getDynamicSwitchs(types rpc.DynamicSwitch_Type) uint32 {
// 	if gDynamicSwitch == nil {
// 		return 0
// 	}

// 	t := uint32(types)
// 	if arr, ok := gDynamicSwitch[t]; ok {
// 		return (*arr)[0].Value
// 	}

// 	return 0
// }
