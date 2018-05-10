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

type HeartbeatRequest struct {
	peer *Peer
	host string
	term uint64
}

type HeartbeatResponse struct {
	peer         *Peer
	responseCode int32
}

type SendHeartbeatImp struct {
	server *server
	mutex  sync.Mutex
}

func (e *SendHeartbeatImp) SendHeartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	resp := 0
	curterm := e.server.currentTerm
	fmt.Printf("receive leader heartbeat from %s term:%d\n", req.Host, req.GetTerm())
	if req.GetTerm() == e.server.currentTerm && req.GetHost() == e.server.currentLeader {
		e.server.leaderAcceptTime = util.GetTimestampInMilli()
		resp = 1
	} else if req.GetTerm() > e.server.currentTerm {
		if len(req.Entries) > 0 {
			lentries := []*LogUnit{}
			err := json.Unmarshal(req.Entries, &lentries)
			if err != nil {
				fmt.Printf("failed to parse entries, err:%s\n", err)
				resp = 2
			} else {
				fmt.Printf("to sync log from leader, logcnt:%d\n", len(lentries))
				for _, l := range lentries {
					e.server.log.Commite(l, e.server.log.file)
				}
				e.server.SetState(Follower)
				e.server.currentTerm = req.GetTerm()
				e.server.currentLeader = req.GetHost()
				e.server.leaderAcceptTime = util.GetTimestampInMilli()
				resp = 1
			}
		}
	}

	pb := &proto.HeartbeatResponse{
		Peerterm: curterm,
		Respcode: uint32(resp),
	}
	return pb, nil
}

func SendHeartbeatCli(s *server, req *HeartbeatRequest, entries []byte) {
	conn, err := grpc.Dial(req.peer.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("dail rpc failed, err: %s\n", err)
		return
	}

	client := proto.NewHeartbeatClient(conn)
	pb := &proto.HeartbeatRequest{
		Host:    s.conf.Host,
		Term:    req.term,
		Event:   "",
		Entries: entries,
	}
	res, err := client.SendHeartbeat(context.Background(), pb)

	if err != nil {
		fmt.Printf("leader SendHeartbeat failed, err:%s\n", err)
		return
	}
	fmt.Printf("[heartbeat]from:%s to:%s rpcRes:%+v\n", s.conf.Host, req.peer.Host, res)

	//TODO
	if res.Peerterm < s.currentTerm && res.Respcode != 2 {
		lulist := s.log.LogEntriesToSync(res.Peerterm)
		ludata, _ := json.Marshal(lulist)
		SendHeartbeatCli(s, req, []byte(ludata))
	}
}
