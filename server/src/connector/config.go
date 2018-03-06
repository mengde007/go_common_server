package connector

//任务配置表
type TaskCfg struct {
	Value   int32
	Rewards string
}

//经验配置表
type UplevelCfg struct {
	Exp          int32
	Rewards      int32
	ExtraRewards int32
}

type ItemCfg struct {
	SellPrice int32
	BuyPrice  string
	BuyAddID  string
	VipCard   int32
}

//cns配置
type CnsConfig struct {
	CnsHost          string
	CnsHostForClient string
	CnsForCenter     string
	FsHost           []string
	DebugHost        string
	GcTime           uint8
	CpuProfile       bool
	ServerId         uint8
	MaxPlayerCount   int32
}
