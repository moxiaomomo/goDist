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
	"math"
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

	name  string
	path  string
	state string

	currentLeader string
	currentTerm   uint64
	confPath      string

	transporter    Transporter
	log            *Log
	conf           *Config
	peers          map[string]*Peer
	voteGrantedNum int
	votedForTerm   uint64 // vote one peer as a leader in curterm

	leaderAcceptTime   int64
	heartbeatInterval  int64
	appendEntryRespCnt int
}

type Server interface {
	Start() error
	IsRunning() bool
	State() string
	CanCommitLog() bool

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

func (s *server) InitAppendEntry() {
	s.appendEntryRespCnt = 0
}

func (s *server) IncrAppendEntryResp() {
	s.appendEntryRespCnt += 1
}

func (s *server) CanCommitLog() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.appendEntryRespCnt >= int(math.Ceil(float64(len(s.peers)/2)))
}

func (s *server) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := RunningStates[s.state]
	return ok
}

func (s *server) State() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.state
}

func (s *server) VotedForTerm() uint64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.votedForTerm
}

func (s *server) SetVotedForTerm(term uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.votedForTerm = term
}

func (s *server) VoteGrantedNum() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.voteGrantedNum
}

func (s *server) IncrVoteGrantedNum() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.voteGrantedNum += 1
}

func (s *server) IncrTermForvote() {
	s.currentTerm += 1
}

func (s *server) SetState(state string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state = state
}

func (s *server) Init() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%d", s.State())
	}

	if s.State() == Initiated {
		s.SetState(Initiated)
		return nil
	}

	err := os.Mkdir(path.Join(s.path, "snapshot"), 0600)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("raft initiation error: %s", err)
	}

	err = s.loadConf()
	if err != nil {
		return fmt.Errorf("raft load config error: %s", err)
	}

	logpath := path.Join(s.path, "internlog")
	err = os.Mkdir(logpath, 0600)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("raft-log initiation error: %s", err)
	}
	if err = s.log.LogInit(fmt.Sprintf("%s/%s%s", logpath, s.conf.LogPrefix, s.conf.CandidateName)); err != nil {
		return fmt.Errorf("raft-log initiation error: %s", err)
	}
	fmt.Printf("%+v\n", s.log.entries)

	err = s.LoadState()
	if err != nil {
		return fmt.Errorf("raft load srvstate error: %s", err)
	}
	_, s.currentTerm = s.log.LastCommitInfo()
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
	//pb.RegisterHeartbeatServer(server, &SendHeartbeatImp{server: s})

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
			Name: c,
			Host: c,
		}
	}

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
			s.currentTerm += 1   // candidate term to increased by 1
			s.voteGrantedNum = 1 // vote for itself
			s.peers[s.conf.Host].SetVoteRequestState(VoteGranted)
			for _, p := range s.peers {
				if s.conf.Host == p.Host {
					continue
				}
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
				s.IncrTermForvote()
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
	s.InitAppendEntry()
	lentries := s.log.GetLogEntries(len(s.log.entries)-1, len(s.log.entries))
	for _, p := range s.peers {
		if s.conf.Host == p.Host {
			continue
		}
		RequestAppendEntriesCli(s, p, lentries)
	}
	if s.CanCommitLog() {
		index, _ := s.log.LastLogInfo()
		s.log.UpdateCommitIndex(index)
		s.FlushState()
		fmt.Printf("to commit log, index:%d term:%d\n", index, s.currentTerm)
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
					RequestAppendEntriesCli(s, p, []*LogEntry{})
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
