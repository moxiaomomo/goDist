package raft

import (
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
	fmt.Printf("receive leader heartbeat from %s term:%d\n", req.Host, req.GetTerm())
	if req.GetTerm() == e.server.currentTerm && req.GetHost() == e.server.currentLeader {
		e.server.leaderAcceptTime = util.GetTimestampInMilli()
		resp = 1
	} else if req.GetTerm() > e.server.currentTerm {
		e.server.SetState(Follower)
		e.server.currentTerm = req.GetTerm()
		e.server.currentLeader = req.GetHost()
		e.server.leaderAcceptTime = util.GetTimestampInMilli()
		resp = 1
	}

	pb := &proto.HeartbeatResponse{
		Respcode: uint32(resp),
	}
	return pb, nil
}

func SendHeartbeatCli(s *server, req *HeartbeatRequest) {
	conn, err := grpc.Dial(req.peer.Host, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("dail rpc failed, err: %s\n", err)
		return
	}

	client := proto.NewHeartbeatClient(conn)
	pb := &proto.HeartbeatRequest{
		Host:  s.conf.Host,
		Term:  req.term,
		Event: "",
	}
	res, err := client.SendHeartbeat(context.Background(), pb)

	if err != nil {
		fmt.Printf("leader SendHeartbeat failed, err:%s\n", err)
		return
	}
	fmt.Printf("[heartbeat]from:%s to:%s rpcRes:%+v\n", s.conf.Host, req.peer.Host, res)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	//TODO
}
