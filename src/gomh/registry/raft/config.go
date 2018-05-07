package raft

type Config struct {
	LogPrefix     string   `json:"logprefix"`
	CommitIndex   uint64   `json:"commitIndex"`
	PeerHosts     []string `json:"peerHosts"`
	Host          string   `json:"host"`
	CandidateName string   `json:"candidateName"`
}
