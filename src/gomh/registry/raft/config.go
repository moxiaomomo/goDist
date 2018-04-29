package raft

type Config struct {
	CommitIndex   uint64   `json:"commitIndex"`
	PeerHosts     []string `json:"peerHosts"`
	Host          string   `json:"host"`
	CandidateName string   `json:"candidateName"`
}
