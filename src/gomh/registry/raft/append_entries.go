package raft

import (
	"fmt"
	"golang.org/x/net/context"
	pb "gomh/registry/raft/proto"
	"gomh/util"
	"sync"
)

type AppendEntriesImp struct {
	server *server
	mutex  sync.Mutex
}

func (e *AppendEntriesImp) AppendEntries(ctx context.Context, req *pb.AppendEntriesReuqest) (*pb.AppendEntriesResponse, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	pb := &pb.AppendEntriesResponse{
		Success: false,
	}
	if req.GetTerm() >= e.server.currentTerm {
		e.server.SetState(Follower)
		e.server.currentTerm = req.GetTerm()
		e.server.currentLeader = req.GetLeaderName()
		e.server.leaderAcceptTime = util.GetTimestampInMilli()

		lindex, _ := e.server.log.LastLogInfo()
		entries := req.GetEntries()
		if req.GetPreLogIndex() > lindex {
			pb.Success = false
			pb.Index = lindex
		} else if req.GetPreLogIndex() < lindex {
			backindex := len(e.server.log.entries) - 1
			for i := backindex; i >= 0; i-- {
				if e.server.log.entries[i].Entry.GetIndex() <= req.GetPreLogIndex() {
					backindex = i
					break
				}
			}
			e.server.log.entries = e.server.log.entries[0 : backindex+1]
			e.server.log.RefreshLog()
			if e.server.log.entries[backindex].Entry.GetIndex() == req.GetPreLogIndex() {
				for _, entry := range entries {
					e.server.log.AppendEntry(&LogEntry{Entry: entry})
				}
				pb.Success = true
			} else {
				pb.Index = e.server.log.entries[backindex].Entry.GetIndex()
				pb.Success = false
			}
		} else if req.GetPreLogIndex() == lindex {
			for _, entry := range entries {
				e.server.log.AppendEntry(&LogEntry{Entry: entry})
			}
			pb.Success = true
		}
	}

	if pb.Success {
		e.server.log.UpdateCommitIndex(req.GetCommitIndex())
	}
	lindex, lterm := e.server.log.LastLogInfo()
	pb.Index = lindex
	pb.Term = lterm

	fmt.Printf("to be follower to %s\n", req.LeaderName)
	return pb, nil
}
