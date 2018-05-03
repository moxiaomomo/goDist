package raft

import (
	"sync"
	"time"
)

type Peer struct {
	server           *server
	Name             string
	Host             string
	VoteRequestState int

	lastActivity      time.Time
	heartbeatInterval time.Duration

	mutex sync.RWMutex
}

func NewPeer(server *server, name, host string, heartbeatInterval time.Duration) *Peer {
	return &Peer{
		server:            server,
		Name:              name,
		Host:              host,
		VoteRequestState:  NotYetVote,
		heartbeatInterval: heartbeatInterval,
	}
}
