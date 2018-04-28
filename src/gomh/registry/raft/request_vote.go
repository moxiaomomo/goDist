package raft

import (
	"fmt"
	"golang.org/x/net/context"
	"gomh/registry/raft/proto"
	"google.golang.org/grpc"
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

type RequestVoteImp struct{}

func (e *RequestVoteImp) RequestVoteMe(ctx context.Context, req *proto.VoteRequest) (*proto.VoteResponse, error) {
	pb := &proto.VoteResponse{
		Term:        req.Term,
		VoteGranted: true,
	}
	return pb, nil
}

func RequestVoteMeCli(req *RequestVoteRequest) {
	fmt.Println("to request vote...")
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
	fmt.Printf("rpcRes:%+v\n", res)
	return
}
