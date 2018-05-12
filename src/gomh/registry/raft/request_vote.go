package raft

import (
	"fmt"
	"golang.org/x/net/context"
	"gomh/registry/raft/proto"
	"google.golang.org/grpc"
	//	"math"
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

func RequestVoteMeCli(s *server, req *RequestVoteRequest) {
	conn, err := grpc.Dial(req.peer.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Errorf("dail rpc failed, err: %s\n", err)
		return
	}

	client := proto.NewRequestVoteClient(conn)
	pb := &proto.VoteRequest{
		Term:          req.Term,
		LastLogIndex:  req.LastLogIndex,
		LastLogTerm:   req.LastLogTerm,
		CandidateName: req.CandidateName,
	}
	res, err := client.RequestVoteMe(context.Background(), pb)

	if err != nil {
		fmt.Printf("client RequestVoteMe failed, err:%s\n", err)
		return
	}
	fmt.Printf("[requestvote]from:%s to:%s rpcRes:%+v\n", s.conf.Host, req.peer.Host, res)

	if res.VoteGranted && s.State() == Candidate {
		s.IncrVoteGrantedNum()
		s.peers[req.peer.Host].SetVoteRequestState(VoteGranted)
	} else {
		s.peers[req.peer.Host].SetVoteRequestState(VoteRejected)
	}
	return
}
