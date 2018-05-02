package raft

import (
	"fmt"
	"golang.org/x/net/context"
	"gomh/registry/raft/proto"
	"gomh/util"
	"google.golang.org/grpc"
	"sync"
)

type AppendEntriesRequest struct {
	peer        *Peer
	leaderName  string
	leaderHost  string
	term        uint64
	commitIndex uint64
}

type AppendEntriesResponse struct {
	peer         *Peer
	responseCode int32
}

type AppendEntriesImp struct {
	server *server
	mutex  sync.Mutex
}

func (e *AppendEntriesImp) AppendEntries(ctx context.Context, req *proto.AppendEntriesReuqest) (*proto.AppendEntriesResponse, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	resp := 0
	if req.GetTerm() > e.server.currentTerm {
		e.server.SetState(Follower)
		e.server.currentTerm = req.GetTerm()
		e.server.currentLeader = req.GetLeaderHost()
		e.server.leaderAcceptTime = util.GetTimestampInMilli()

		resp = 1
		fmt.Printf("to be follower to %s\n", req.LeaderName)
	}

	pb := &proto.AppendEntriesResponse{
		ResponseCode: int32(resp),
	}
	return pb, nil
}

func RequestAppendEntriesCli(s *server, req *AppendEntriesRequest) {
	conn, err := grpc.Dial(req.peer.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("dail rpc failed, err: %s\n", err)
		return
	}

	client := proto.NewAppendEntriesClient(conn)
	pb := &proto.AppendEntriesReuqest{
		LeaderName:  s.conf.Host,
		LeaderHost:  s.conf.Host,
		Term:        req.term,
		CommitIndex: 100,
	}
	res, err := client.AppendEntries(context.Background(), pb)

	if err != nil {
		fmt.Printf("leader reqeust AppendEntries failed, err:%s\n", err)
		return
	}
	fmt.Printf("[appendentry]from:%s to:%s rpcRes:%+v\n", s.conf.Host, req.peer.Host, res)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	//TODO
}
