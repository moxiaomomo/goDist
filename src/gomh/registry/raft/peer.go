package raft

import (
	"sync"
	"time"
)

type Peer struct {
	server           *server
	Name             string
	Host             string
	voteRequestState int

	lastActivity      time.Time
	heartbeatInterval time.Duration

	mutex sync.RWMutex
}

func NewPeer(server *server, name, host string, heartbeatInterval time.Duration) *Peer {
	return &Peer{
		server:            server,
		Name:              name,
		Host:              host,
		voteRequestState:  NotYetVote,
		heartbeatInterval: heartbeatInterval,
	}
}

func (p *Peer) SetVoteRequestState(state int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.voteRequestState = state
}

func (p *Peer) VoteRequestState() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.voteRequestState
}
