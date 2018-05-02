package raft

type Peer struct {
	server            *server
	Name              string
	Host              string
	VoteRequestState  int
	HeartBeatInterval int
}
