package proto

const (
	Ok      = 0
	NoExist = 404
)

type DBQuery struct {
	Table string
	Key   string
}

type DBQueryResult struct {
	Code  uint32
	Value []byte
}

type DBDel struct {
	Table string
	Key   string
}

type DBDelResult struct {
	Code uint32
}

type DBWrite struct {
	Table string
	Key   string
	Value []byte
}

type DBWriteResult struct {
	Code uint32
}

//账号服务器转用
type AccountDbWrite struct {
	Table string
	Key   string
	Value string
}

type AccountDbWriteResult struct {
	Code uint32
}

type AccountDbDel struct {
	Table string
	Key   string
}

type AccountDbDelResult struct {
	Code uint32
}

type AccountDbQuery struct {
	Table string
	Key   string
}

type AccountDbQueryResult struct {
	Code  uint32
	Value string
}
