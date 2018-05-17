package raft

import (
	"fmt"
	"golang.org/x/net/context"
	pb "gomh/registry/raft/proto"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Peer struct {
	server           *server
	Name             string
	Host             string
	Client           string
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

func (p *Peer) RequestVoteMe(lastLogIndex, lastTerm uint64) {
	conn, err := grpc.Dial(p.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Errorf("dail rpc failed, err: %s\n", err)
		return
	}
	defer conn.Close()

	client := pb.NewRequestVoteClient(conn)
	pb := &pb.VoteRequest{
		Term:          p.server.currentTerm,
		LastLogIndex:  lastLogIndex,
		LastLogTerm:   lastTerm,
		CandidateName: p.server.conf.CandidateName,
		Host:          p.server.conf.Host,
	}
	res, err := client.RequestVoteMe(context.Background(), pb)

	if err != nil {
		//		fmt.Printf("client RequestVoteMe failed, err:%s\n", err)
		return
	}
	//	fmt.Printf("[requestvote]from:%s to:%s rpcRes:%+v\n", p.server.conf.Host, p.Host, res)

	if res.VoteGranted && p.server.State() == Candidate {
		p.server.IncrVoteGrantedNum()
		p.SetVoteRequestState(VoteGranted)
	} else {
		p.SetVoteRequestState(VoteRejected)
	}
	return
}

func (p *Peer) RequestAppendEntries(entries []*pb.LogEntry, sindex, lindex, lterm uint64) {
	if p.server.State() != Leader {
		fmt.Println("only leader can request append entries.")
		return
	}

	conn, err := grpc.Dial(p.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("dail rpc failed, err: %s\n", err)
		return
	}
	defer conn.Close()

	client := pb.NewAppendEntriesClient(conn)

	req := &pb.AppendEntriesReuqest{
		Term:          p.server.currentTerm,
		FirstLogIndex: sindex,
		PreLogIndex:   lindex,
		PreLogTerm:    lterm,
		CommitIndex:   p.server.log.CommitIndex(),
		LeaderName:    p.server.conf.Host,
		LeaderHost:    p.server.conf.Host,
		Entries:       entries,
	}

	res, err := client.AppendEntries(context.Background(), req)
	// fmt.Printf("response from %s\n", p.Host)

	if err != nil {
		fmt.Printf("leader reqeust AppendEntries failed, err:%s\n", err)
		return
	}

	if res.Success {
		p.server.IncrAppendEntryResp()
	} else {
		el := []*pb.LogEntry{}
		for _, e := range p.server.log.entries {
			if e.Entry.GetIndex() <= res.Index {
				continue
			}
			el = append(el, e.Entry)
		}
		req := &pb.AppendEntriesReuqest{
			Term:          p.server.currentTerm,
			FirstLogIndex: sindex,
			PreLogIndex:   res.Index,
			PreLogTerm:    res.Term,
			CommitIndex:   p.server.log.CommitIndex(),
			LeaderName:    p.server.conf.Host,
			LeaderHost:    p.server.conf.Host,
			Entries:       el,
		}

		res, err = client.AppendEntries(context.Background(), req)

		if err != nil {
			fmt.Printf("leader reqeust AppendEntries failed, err:%s\n", err)
			return
		} else {
			fmt.Printf("synclog res: %s %+v\n", p.Host, res)
		}
		if res.Success {
			p.server.IncrAppendEntryResp()
		}
	}

	//TODO
}
