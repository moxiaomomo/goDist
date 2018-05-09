package raft

import (
	"encoding/json"
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
	commitState int32
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
	if req.GetTerm() > e.server.currentTerm && req.CommitState != 1 {
		e.server.SetState(Follower)
		e.server.currentTerm = req.GetTerm()
		e.server.currentLeader = req.GetLeaderHost()
		e.server.leaderAcceptTime = util.GetTimestampInMilli()

		if err := json.Unmarshal(req.GetLogUnit(), &e.server.log.toc_entry); err != nil {
			fmt.Printf("decode logunit failed: %+v %s\n", req.GetLogUnit(), err)
		}

		resp = 1
		fmt.Printf("to be follower to %s\n", req.LeaderName)
	} else if req.CommitState == 1 && e.server.State() != Leader {
		fmt.Println("To commit log as leader log commited.")
		e.server.log.Commite(&e.server.log.toc_entry, e.server.log.file)
	}

	pb := &proto.AppendEntriesResponse{
		ResponseCode: int32(resp),
	}
	return pb, nil
}

func RequestAppendEntriesCli(s *server, req *AppendEntriesRequest, lastindexstart, lastindexend, lastterm uint64) {
	if s.State() != Leader {
		fmt.Println("only leader can request append entries.")
		return
	}

	conn, err := grpc.Dial(req.peer.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("dail rpc failed, err: %s\n", err)
		return
	}

	client := proto.NewAppendEntriesClient(conn)
	pb := &proto.AppendEntriesReuqest{
		LeaderName:  s.conf.CandidateName,
		LeaderHost:  s.conf.Host,
		Term:        req.term,
		CommitState: req.commitState,
	}

	if req.commitState != 1 {
		logunit := NewLogUnit(s.conf.CandidateName, req.term, lastterm, lastindexstart, lastindexend)
		logunitd, _ := json.Marshal(logunit)
		pb.LogUnit = []byte(logunitd)
	}

	res, err := client.AppendEntries(context.Background(), pb)

	if err != nil {
		fmt.Printf("leader reqeust AppendEntries failed, err:%s\n", err)
		return
	}
	fmt.Printf("[appendentry]from:%s to:%s rpcRes:%+v\n", s.conf.Host, req.peer.Host, res)

	if req.commitState != 1 && res.ResponseCode == 1 {
		s.IncrAppendEntryResp()
	}

	//TODO
}
