package raft

import (
	"fmt"
	"golang.org/x/net/context"
	"gomh/registry/raft/proto"
	"google.golang.org/grpc"
	"math"
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
	e.mutex.Lock()
	defer e.mutex.Unlock()

	voteGranted := false
	if e.server.State() == Candidate && req.Term > e.server.votedForTerm {
		voteGranted = true
	}
	// vote only once for one term
	e.server.votedForTerm = req.Term
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
		fmt.Errorf("client RequestVoteMe failed, err:%s\n", err)
		return
	}
	fmt.Printf("[requestvote]from:%s to:%s rpcRes:%+v\n", s.conf.Host, req.peer.Host, res)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if res.VoteGranted && s.State() == Candidate {
		s.voteGrantedNum += 1
		mostLen := int(math.Ceil(float64(len(s.conf.PeerHosts) / 2)))
		if s.voteGrantedNum > mostLen {
			s.SetState(Leader)
		}
		s.peers[req.peer.Host].VoteRequestState = VoteGranted
	} else {
		s.peers[req.peer.Host].VoteRequestState = VoteRejected
	}
	return
}
