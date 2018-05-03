package raft

import (
	"encoding/json"
	"fmt"
	"net"
	//	"golang.org/x/net/context"
	pb "gomh/registry/raft/proto"
	"gomh/util"
	"google.golang.org/grpc"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"sync"
	"time"
)

type server struct {
	*eventDispatcher

	mutex   sync.RWMutex
	stopped chan bool

	name          string
	path          string
	state         string
	currentLeader string
	currentTerm   uint64
	confPath      string

	transporter    Transporter
	log            *Log
	conf           *Config
	peers          map[string]*Peer
	voteGrantedNum int
	votedForTerm   uint64 // vote one peer as a leader in curterm

	leaderAcceptTime  int64
	heartbeatInterval int64
}

type Server interface {
	Start() error
	IsRunning() bool
	State() string

	AddPeer(name string, connectionInfo string) error
	RemovePeer(name string) error
}

func NewServer(name, path, confPath string) (Server, error) {
	s := &server{
		name:              name,
		path:              path,
		confPath:          confPath,
		state:             Stopped,
		log:               newLog(),
		heartbeatInterval: 1000, // 1000ms
	}
	return s, nil
}

func (s *server) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := RunningStates[s.state]
	return ok
}

func (s *server) State() string {
	//	s.mutex.Lock()
	//	defer s.mutex.Unlock()

	return s.state
}

func (s *server) SetState(state string) {
	//	s.mutex.Lock()
	//	defer s.mutex.Unlock()

	s.state = state
}

func (s *server) Init() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%d", s.State())
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.State() == Initiated {
		s.SetState(Initiated)
		return nil
	}

	err := os.Mkdir(path.Join(s.path, "snapshot"), 0700)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("raft initiation error: %s", err)
	}

	err = s.loadConf()
	if err != nil {
		return fmt.Errorf("raft load config error: %s", err)
	}

	_, s.currentTerm = s.log.LastInfo()

	s.SetState(Initiated)
	return nil
}

func (s *server) Start() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%d", s.State())
	}

	if err := s.Init(); err != nil {
		return err
	}

	s.stopped = make(chan bool)

	s.SetState(Follower)

	loopch := make(chan int)
	go func() {
		defer func() { loopch <- 1 }()
		s.acceptVoteRequest()
	}()
	s.loop()
	return nil
}

func (s *server) acceptVoteRequest() {
	server := grpc.NewServer()
	pb.RegisterRequestVoteServer(server, &RequestVoteImp{server: s})
	pb.RegisterAppendEntriesServer(server, &AppendEntriesImp{server: s})
	pb.RegisterHeartbeatServer(server, &SendHeartbeatImp{server: s})

	fmt.Printf("To listen on %s\n", s.conf.Host)
	address, err := net.Listen("tcp", s.conf.Host)
	if err != nil {
		panic(err)
	}

	if err := server.Serve(address); err != nil {
		panic(err)
	}
}

func (s *server) loadConf() error {
	confpath := path.Join(s.path, "raft.cfg")
	if s.confPath != "" {
		confpath = path.Join(s.path, s.confPath)
	}

	cfg, err := ioutil.ReadFile(confpath)
	if err != nil {
		fmt.Errorf("open config file failed, err:%s", err)
		return nil
	}

	conf := &Config{}
	if err = json.Unmarshal(cfg, conf); err != nil {
		return err
	}
	s.conf = conf
	s.peers = make(map[string]*Peer)
	for _, c := range s.conf.PeerHosts {
		s.peers[c] = &Peer{
			Name:             c,
			Host:             c,
			VoteRequestState: NotYetVote,
		}
	}

	s.log.UpdateCommitIndex(conf.CommitIndex)
	return nil
}

func (s *server) loop() {
	for s.State() != Stopped {
		fmt.Printf("current state:%s, term:%d\n", s.State(), s.currentTerm)
		switch s.State() {
		case Follower:
			s.followerLoop()
		case Candidate:
			s.candidateLoop()
		case Leader:
			s.leaderLoop()
			//		case Snapshotting:
			//			s.snapshotLoop()
		}
	}
}

func (s *server) candidateLoop() {
	t := time.NewTimer(time.Duration(150+rand.Intn(150)) * time.Millisecond)
	for s.State() == Candidate {
		select {
		case <-t.C:
			if s.State() != Candidate {
				return
			}
			s.voteGrantedNum = 0
			s.currentTerm += 1
			for _, p := range s.peers {
				r := &RequestVoteRequest{
					peer:          p,
					Term:          s.currentTerm,
					LastLogIndex:  2,
					LastLogTerm:   2,
					CandidateName: s.conf.CandidateName,
				}
				RequestVoteMeCli(s, r)
			}

			t.Reset(time.Duration(150+rand.Intn(150)) * time.Millisecond)
		case isStop := <-s.stopped:
			if isStop {
				s.SetState(Stopped)
				break
			}
		}
	}
}

func (s *server) followerLoop() {
	t := time.NewTimer(time.Duration(s.heartbeatInterval) * time.Millisecond)
	for s.State() == Follower {
		select {
		case <-t.C:
			if s.State() != Follower {
				return
			}
			if util.GetTimestampInMilli()-s.leaderAcceptTime > s.heartbeatInterval*2 {
				s.SetState(Candidate)
			}
			t.Reset(time.Duration(s.heartbeatInterval) * time.Millisecond)
		case isStop := <-s.stopped:
			if isStop {
				s.SetState(Stopped)
				break
			}
		}
	}
}

func (s *server) leaderLoop() {
	for _, p := range s.peers {
		if s.conf.Host == p.Host {
			continue
		}
		r := &AppendEntriesRequest{
			peer:        p,
			leaderName:  s.conf.Host,
			leaderHost:  s.conf.Host,
			term:        s.currentTerm,
			commitIndex: 100,
		}
		RequestAppendEntriesCli(s, r)
	}

	t := time.NewTimer(time.Duration(s.heartbeatInterval) * time.Millisecond)
	for s.State() == Leader {
		select {
		case <-t.C:
			if s.State() != Leader {
				return
			}
			if util.GetTimestampInMilli()-s.leaderAcceptTime > s.heartbeatInterval {
				for _, p := range s.peers {
					if s.conf.Host == p.Host {
						continue
					}
					r := &HeartbeatRequest{
						peer: p,
						host: s.conf.Host,
						term: s.currentTerm,
					}
					SendHeartbeatCli(s, r)
				}
			}
			t.Reset(time.Duration(s.heartbeatInterval) * time.Millisecond)
		case isStop := <-s.stopped:
			if isStop {
				s.SetState(Stopped)
				break
			}
		}
	}
}

func (s *server) AddPeer(name string, connectionInfo string) error {
	if s.peers[name] != nil {
		return nil
	}

	if s.name != name {
		ti := time.Duration(s.heartbeatInterval) * time.Millisecond
		peer := NewPeer(s, name, connectionInfo, ti)

		s.peers[peer.Name] = peer
	}

	return nil
}

func (s *server) RemovePeer(name string) error {
	if name == s.name {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	peer := s.peers[name]
	if peer == nil {
		return nil
	}

	delete(s.peers, name)
	return nil
}
