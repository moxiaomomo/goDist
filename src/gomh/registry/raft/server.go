package raft

import (
	"encoding/json"
	"fmt"
	"net"
	//	"golang.org/x/net/context"
	pb "gomh/registry/raft/proto"
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

	name        string
	path        string
	state       string
	currentTerm uint64
	confPath    string

	transporter    Transporter
	log            *Log
	conf           *Config
	peers          map[string]*Peer
	voteGrantedNum int
	hasVoteLeader  bool // vote one peer as a leader in curterm
}

type Server interface {
	Start() error
	IsRunning() bool
}

func NewServer(name, path, confPath string) (Server, error) {
	s := &server{
		name:     name,
		path:     path,
		confPath: confPath,
		state:    Stopped,
		log:      newLog(),
	}
	return s, nil
}

func (s *server) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := RunningStates[s.state]
	return ok
}

func (s *server) Init() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%d", s.state)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.state == Initiated {
		s.state = Initiated
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

	s.state = Initiated
	return nil
}

func (s *server) Start() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%d", s.state)
	}

	if err := s.Init(); err != nil {
		return err
	}

	s.stopped = make(chan bool)
	s.state = Candidate

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
	t := time.NewTimer(time.Duration(150+rand.Intn(150)) * time.Millisecond)
	for {
		select {
		case <-t.C:
			s.hasVoteLeader = false
			if s.state == Candidate {
				s.tryRequestVote()
			} else {
				fmt.Printf("current state:%s\n", s.state)
			}
			t.Reset(time.Duration(150+rand.Intn(150)) * time.Millisecond)
		case isStop := <-s.stopped:
			if isStop {
				s.state = Stopped
				break
			}
		}
	}
}

func (s *server) tryRequestVote() {
	s.voteGrantedNum = 0
	for _, p := range s.peers {
		//		if p.VoteRequestState != NotYetVote {
		//			continue
		//		}
		r := &RequestVoteRequest{
			peer:          p,
			Term:          3,
			LastLogIndex:  2,
			LastLogTerm:   2,
			CandidateName: s.conf.CandidateName,
		}
		RequestVoteMeCli(s, r)
	}
	if s.state == Leader {
		for _, p := range s.peers {
			if s.conf.Host == p.Host {
				continue
			}
			r := &AppendEntriesRequest{
				peer:        p,
				leaderName:  s.conf.Host,
				leaderHost:  s.conf.Host,
				term:        s.currentTerm + 100,
				commitIndex: 100,
			}
			RequestAppendEntriesCli(s, r)
		}
	}
}
