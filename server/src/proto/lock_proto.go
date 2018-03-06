package proto

const (
	MethodGetDonateInfo = iota
	MethodAddDonateInfo
)

type CenterConnCns struct {
	ServerId uint8
	Addr     string
}

type CenterConnCnsResult struct {
	Ret bool
}

type CenterDisConnCns struct {
}

type CenterDisConnCnsResult struct {
}

type TryGetLock struct {
	Service   string
	Name      string
	Value     uint64
	ValidTime uint32
}

type GetLockResult struct {
	Result   bool
	OldValue uint64
}

type FreeLock struct {
	Service string
	Name    string
	Value   uint64
}

type FreeLockResult struct {
	Result bool
}

type ForceUnLock struct {
	Service string
	Name    string
}

type ForceUnLockResult struct {
	Result bool
}

type RenewLock struct {
	Service string
	Name    string
	Value   uint64
}

type RenewLockResult struct {
	Result bool
}

type QueryPlayer struct {
	Service string
	Name    string
}

type QueryPlayerResult struct {
	Value uint64
}
