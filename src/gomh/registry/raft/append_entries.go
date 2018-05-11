package raft

import (
	"fmt"
	"golang.org/x/net/context"
	pb "gomh/registry/raft/proto"
	"gomh/util"
	"google.golang.org/grpc"
	"sync"
)

type AppendEntriesImp struct {
	server *server
	mutex  sync.Mutex
}

func (e *AppendEntriesImp) AppendEntries(ctx context.Context, req *pb.AppendEntriesReuqest) (*pb.AppendEntriesResponse, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	resp := false
	if req.GetTerm() > e.server.currentTerm {
		e.server.SetState(Follower)
		e.server.currentTerm = req.GetTerm()
		e.server.currentLeader = req.GetLeaderName()
		e.server.leaderAcceptTime = util.GetTimestampInMilli()

		e.server.log.UpdateCommitIndex(req.GetCommitIndex())

		entries := req.GetEntries()
		if len(entries) > 0 {
			for _, entry := range entries {
				e.server.log.AppendEntry(&LogEntry{Entry: entry})
			}
		}

		resp = true
		fmt.Printf("to be follower to %s\n", req.LeaderName)
	}

	pb := &pb.AppendEntriesResponse{
		Success: resp,
	}
	return pb, nil
}

func RequestAppendEntriesCli(s *server, peer *Peer, logEntries []*LogEntry) {
	if s.State() != Leader {
		fmt.Println("only leader can request append entries.")
		return
	}

	conn, err := grpc.Dial(peer.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("dail rpc failed, err: %s\n", err)
		return
	}

	plastindex, plastterm := s.log.PreLastLogInfo()
	client := pb.NewAppendEntriesClient(conn)

	entries := []*pb.LogEntry{}
	for _, l := range logEntries {
		entries = append(entries, l.Entry)
	}

	req := &pb.AppendEntriesReuqest{
		Term:        s.currentTerm,
		PreLogIndex: plastindex,
		PreLogTerm:  plastterm,
		CommitIndex: s.log.CommitIndex(),
		LeaderName:  s.conf.Host,
		Entries:     entries,
	}

	res, err := client.AppendEntries(context.Background(), req)

	if err != nil {
		fmt.Printf("leader reqeust AppendEntries failed, err:%s\n", err)
		return
	}
	fmt.Printf("[appendentry]from:%s to:%s rpcRes:%+v\n", s.conf.Host, peer.Host, res)

	if res.Success {
		s.IncrAppendEntryResp()
	}

	//TODO
}
