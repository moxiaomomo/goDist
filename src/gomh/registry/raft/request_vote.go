package raft

import (
	"golang.org/x/net/context"
	"gomh/registry/raft/proto"
	"sync"
)

type RequestVoteRequest struct {
	peer          *Peer
	Term          uint64
	LastLogIndex  uint64
	LastLogTerm   uint64
	CandidateName string
}

type RequestVoteResponse struct {
	peer        *Peer
	Term        uint64
	VoteGranted bool
}

type RequestVoteImp struct {
	mutex  sync.RWMutex
	server *server
}

func (e *RequestVoteImp) RequestVoteMe(ctx context.Context, req *proto.VoteRequest) (*proto.VoteResponse, error) {
	voteGranted := false
	lastindex, _ := e.server.log.LastLogInfo()
	if e.server.State() == Candidate && req.Term > e.server.VotedForTerm() && req.LastLogIndex >= lastindex {
		voteGranted = true
	}
	// vote only once for one term
	e.server.SetVotedForTerm(req.Term)
	pb := &proto.VoteResponse{
		Term:        req.Term,
		VoteGranted: voteGranted,
	}
	return pb, nil
}
