package raft

import (
	//	"encoding/json"
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

// handle appendentries request
func (e *AppendEntriesImp) AppendEntries(ctx context.Context, req *pb.AppendEntriesReuqest) (*pb.AppendEntriesResponse, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	reqentries := req.GetEntries()
	lindex, _ := e.server.log.LastLogInfo()
	pb := &pb.AppendEntriesResponse{
		Success: false,
	}

	if e.server.IsServerMember(req.LeaderHost) {
		e.server.SetState(Follower)
		e.server.currentTerm = req.GetTerm()
		e.server.currentLeader = req.GetLeaderName()
		e.server.leaderAcceptTime = util.GetTimestampInMilli()
		//		if req.GetPreLogIndex() == lindex && (len(reqentries) > 0 || (len(reqentries) == 0 && !e.server.log.IsEmpty())) {
		//			fmt.Println("-=-=-=-=-=-=-=-********")
		//			fmt.Printf("%d %d %t\n", req.GetTerm(), e.server.currentTerm, e.server.IsServerMember(req.LeaderHost))
		//		}

		if req.GetPreLogIndex() > lindex {
			fmt.Println("1")
			pb.Success = false
			pb.Index = lindex
		} else if req.GetPreLogIndex() < lindex {
			fmt.Println("2")
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
				for _, entry := range reqentries {
					fmt.Printf("tosynclog: %+v\n", entry)
					e.server.log.AppendEntry(&LogEntry{Entry: entry})
				}

				pb.Success = true
			} else {
				pb.Index = e.server.log.entries[backindex].Entry.GetIndex()
				pb.Success = false
			}
		} else if req.GetPreLogIndex() == lindex && (len(reqentries) > 0 || (len(reqentries) == 0 && !e.server.log.IsEmpty())) {
			fmt.Printf("===tosynclog: %+v\n", reqentries)
			for _, entry := range reqentries {
				//				if entry.Commandname == "raft:join" {
				//					var cmdinfo = &DefaultJoinCommand{}
				//					_ = json.Unmarshal(entry.Command, cmdinfo)
				//					e.server.AddPeer(cmdinfo.Name, cmdinfo.ConnectionInfo)
				//				}
				e.server.log.AppendEntry(&LogEntry{Entry: entry})
			}
			pb.Success = true
		}
	}

	if pb.Success {
		// update commit index
		e.server.log.UpdateCommitIndex(req.GetCommitIndex())
		// apply the command
		for _, entry := range req.GetEntries() {
			cmd, _ := NewCommand(entry.Commandname, entry.Command)
			if cmdcopy, ok := cmd.(CommandApply); ok {
				cmdcopy.Apply(e.server)
			}
		}
	}

	fmt.Printf("idx:%d:%d h:%s m:%t pb:%+v en:%+v\n", req.GetPreLogIndex(), lindex, req.LeaderHost, e.server.IsServerMember(req.LeaderHost), pb, reqentries)

	lindex, lterm := e.server.log.LastLogInfo()
	pb.Index = lindex
	pb.Term = lterm

	return pb, nil
}
