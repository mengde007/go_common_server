package proto

type SendCnsInfo struct {
	PlayerCount uint16
	ServerIp    string
}

type SendCnsInfoResult struct {
	SendResult uint8
}
