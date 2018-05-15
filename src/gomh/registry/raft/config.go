package raft

type Config struct {
	LogPrefix     string   `json:"logprefix"`
	CommitIndex   uint64   `json:"commitIndex"`
	PeerHosts     []string `json:"peerHosts"`
	Host          string   `json:"host"`
	Client        string   `json:"client"`
	CandidateName string   `json:"candidateName"`
}
