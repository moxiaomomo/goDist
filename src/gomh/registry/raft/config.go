package raft

type Config struct {
	CommitIndex uint64  `json:"commitIndex"`
	Peers       []*Peer `json:"peers"`
}
