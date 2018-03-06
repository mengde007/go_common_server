package proto

import (
// "rpc"
)

type AddPlayer struct {
	PlayerId string
	AuthKey  string
	// ChannelId rpc.GameLocation
	ClanName string
}

type AddPlayerResult struct {
}

type DelPlayer struct {
	PlayerId string
}

type DelPlayerResult struct {
}

type PlayerChatToPlayer struct {
	FromPlayerId    string
	FromPlayerName  string
	FromPlayerLevel int32
	ToPlayerId      string
	Content         string
}

type PlayerChatToPlayerResult struct {
}

type PlayerWorldChat struct {
	FromPlayerId    string
	FromPlayerName  string
	FromPlayerLevel int32
	Content         string
	CName           string
	CSymbol         uint32
	LastLeagueRank  uint32
	Viplevel        uint32
	UseIM           bool
	VoiceTime       string
}

type PlayerWorldChatResult struct {
}

type ChatSendMsg2Player struct {
	MsgName    string
	PlayerList []string
	Buf        []byte
}

type ChatSendMsg2PlayerResult struct {
}

type ChatSendMsg2LPlayer struct {
	MsgName string
	Buf     []byte
	// Channel  rpc.Login_Platform
	LevelMin uint32
	LevelMax uint32
}

type ChatSendMsg2LPlayerResult struct {
}
