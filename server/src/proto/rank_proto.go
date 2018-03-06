package proto

type SetRankInfo struct {
	Uid   string
	EType int
	Value int32
}

type SetRankInfoRst struct {
}

type GetRankInfo struct {
	EType  int
	Number int
}

type GetRankInfoRst struct {
	Exps    []string
	Coins   []string
	Profits []string
}

type SaveRankPlayers struct {
	Exps    []byte
	Coins   []byte
	Profits []byte
}
type SaveRankPlayersRst struct {
	OK bool
}

type GetMyRankInfo struct {
	Uid   string
	EType int
}

type GetMyRankInfoRst struct {
	Ranking int32
}
