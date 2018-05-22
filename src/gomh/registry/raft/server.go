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

	leaderAcceptTime  int64
	heartbeatInterval int64
	lastSnapshotTime  int64

	ch       chan interface{}
	syncpeer map[string]int
}

type Server interface {
	Start() error
	IsRunning() bool
	State() string
	//	CanCommitLog() bool

	AddPeer(name string, host string) error
	RemovePeer(name string, host string) error
}

func NewServer(path, confPath string) (Server, error) {
	s := &server{
		//		name:              name,
		path:              path,
		confPath:          confPath,
		state:             Stopped,
		log:               newLog(),
		heartbeatInterval: 300, // 300ms
		lastSnapshotTime:  util.GetTimestampInMilli(),
		syncpeer:          make(map[string]int),
	}
	return s, nil
}

func (s *server) SetTerm(term uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currentTerm = term
}

func (s *server) QuorumSize() int {
	return len(s.peers)/2 + 1
}

func (s *server) SyncPeerStatusOrReset() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sucCnt := 0
	failCnt := 0
	for _, v := range s.syncpeer {
		if v == 1 {
			sucCnt += 1
		} else if v == 0 {
			failCnt += 1
		}
	}

	qsize := s.QuorumSize()
	if sucCnt >= qsize {
		s.resetSyncPeer()
		return 1
	} else if failCnt >= qsize {
		s.resetSyncPeer()
		return 0
	}
	return -1
}

func (s *server) resetSyncPeer() {
	for k, _ := range s.syncpeer {
		s.syncpeer[k] = -1
	}
}

func (s *server) InitSyncPeer() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.resetSyncPeer()
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

func (s *server) Peers() map[string]*Peer {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.peers
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

func (s *server) VoteForSelf() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.voteGrantedNum = 1 // vote for itself
	s.peers[s.conf.Host].SetVoteRequestState(VoteGranted)
}

func (s *server) IsServerMember(host string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.peers[host]; ok {
		return true
	}
	return false
}

func (s *server) RegisterCommands() {
	RegisterCommand(&DefaultJoinCommand{})
	RegisterCommand(&DefaultLeaveCommand{})
	RegisterCommand(&NOPCommand{})
}

// Init steps:
// check if running or initiated before
// load configuration file
// load raft log
// recover server persistent status
// set state = Initiated
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
	fmt.Printf("config: %+v\n", s.conf)

	logpath := path.Join(s.path, "internlog")
	err = os.Mkdir(logpath, 0600)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("raft-log initiation error: %s", err)
	}
	if err = s.log.LogInit(fmt.Sprintf("%s/%s%s", logpath, s.conf.LogPrefix, s.conf.Name)); err != nil {
		return fmt.Errorf("raft-log initiation error: %s", err)
	}

	err = s.LoadState()
	if err != nil {
		return fmt.Errorf("raft load srvstate error: %s", err)
	}

	s.RegisterCommands()

	s.SetState(Initiated)
	return nil
}

// start steps:
// comlete initiation
// set state = Follower
// new goroutine for tcp listening
// enter loop with a propriate state
func (s *server) Start() error {
	if s.IsRunning() {
		return fmt.Errorf("server has been running with state:%d", s.State())
	}

	if err := s.Init(); err != nil {
		return err
	}

	s.stopped = make(chan bool)

	s.SetState(Follower)

	go s.ListenAndServe()
	go s.StartClientServe()
	go s.TickTask()

	s.loop()
	return nil
}

func (s *server) ListenAndServe() {
	server := grpc.NewServer()
	pb.RegisterRequestVoteServer(server, &RequestVoteImp{server: s})
	pb.RegisterAppendEntriesServer(server, &AppendEntriesImp{server: s})

	fmt.Printf("listen internal rpc address: %s\n", s.conf.Host)
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
			Name:   c,
			Host:   c,
			server: s,
		}
	}
	s.ch = make(chan interface{}, len(s.peers)*2)

	return nil
}

func (s *server) writeConf() error {
	confpath := path.Join(s.path, "raft.cfg")
	if s.confPath != "" {
		confpath = path.Join(s.path, s.confPath)
	}

	s.conf.PeerHosts = []string{}
	for _, p := range s.peers {
		s.conf.PeerHosts = append(s.conf.PeerHosts, p.Host)
	}

	f, err := os.OpenFile(confpath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := json.Marshal(s.conf)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(d))
	return err
}

func (s *server) TickTask() {
	t := time.NewTimer(300 * time.Millisecond)
	for {
		select {
		case <-t.C:
			s.FlushState()
			nowt := util.GetTimestampInMilli()
			if nowt-s.lastSnapshotTime > 5000 {
				s.onSnapShotting()
				s.lastSnapshotTime = nowt
			}
			t.Reset(300 * time.Millisecond)
		}
	}
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
		case Stopped:
			// TODO: do something before server stop
			break
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
			s.currentTerm += 1 // candidate term to increased by 1
			s.VoteForSelf()
			lindex, lterm := s.log.LastLogInfo()
			for idx, _ := range s.peers {
				if s.conf.Host == s.peers[idx].Host {
					continue
				}
				s.peers[idx].RequestVoteMe(lindex, lterm)
			}
			if s.VoteGrantedNum() >= s.QuorumSize() {
				s.SetState(Leader)
			} else {
				t.Reset(time.Duration(150+rand.Intn(150)) * time.Millisecond)
			}

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
			if util.GetTimestampInMilli()-s.leaderAcceptTime > s.heartbeatInterval*3 {
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
	// to request append entry as a new leader is elected
	s.syncpeer[s.conf.Host] = 1
	findex := s.log.FirstLogIndex()
	lindex, lterm := s.log.LastLogInfo()
	entry := &pb.LogEntry{
		Index:       lindex + 1,
		Term:        s.currentTerm,
		Commandname: "raft:nop",
		Command:     []byte(""),
	}
	s.log.AppendEntry(&LogEntry{Entry: entry})
	s.log.UpdateCommitIndex(lindex + 1)

	for idx, _ := range s.peers {
		if s.conf.Host == s.peers[idx].Host {
			continue
		}
		go s.peers[idx].RequestAppendEntries([]*pb.LogEntry{entry}, findex, lindex, lterm)
	}

	// send heartbeat as leader state
	s.leaderAcceptTime = util.GetTimestampInMilli()
	t := time.NewTimer(time.Duration(s.heartbeatInterval) * time.Millisecond)
	for s.State() == Leader {
		select {
		case c := <-s.ch:
			switch d := c.(type) {
			case *AppendLogRespChan:
				if d.Failed == false {
					if d.Resp != nil && d.Resp.Success && d.Resp.Term == s.currentTerm {
						s.syncpeer[d.PeerHost] = 1
					} else {
						s.syncpeer[d.PeerHost] = 0
					}
					respStatus := s.SyncPeerStatusOrReset()
					if respStatus != -1 {
						s.leaderAcceptTime = util.GetTimestampInMilli()
						if respStatus == 1 {
							index, _ := s.log.LastLogInfo()
							s.log.UpdateCommitIndex(index)
							//						fmt.Printf("to commit log, index:%d term:%d\n", index, s.currentTerm)
						}
					}
				}
				if util.GetTimestampInMilli()-s.leaderAcceptTime > s.heartbeatInterval*3 {
					s.SetState(Candidate)
				}
			}
		case <-t.C:
			if util.GetTimestampInMilli()-s.leaderAcceptTime > s.heartbeatInterval {
				findex := s.log.FirstLogIndex()
				lindex, lterm := s.log.LastLogInfo()
				s.syncpeer[s.conf.Host] = 1
				for idx, _ := range s.peers {
					if s.conf.Host == s.peers[idx].Host {
						continue
					}
					go s.peers[idx].RequestAppendEntries([]*pb.LogEntry{}, findex, lindex, lterm)
				}
			}
			if s.State() == Leader {
				t.Reset(time.Duration(s.heartbeatInterval) * time.Millisecond)
			}
		case isStop := <-s.stopped:
			if isStop {
				s.SetState(Stopped)
				break
			}
		}
	}
}

func (s *server) AddPeer(name string, host string) error {
	s.mutex.Lock()

	if s.peers[host] != nil {
		s.mutex.Unlock()
		return nil
	}

	if s.conf.Name != name {
		ti := time.Duration(s.heartbeatInterval) * time.Millisecond
		peer := NewPeer(s, name, host, ti)
		s.peers[host] = peer
	}

	// to flush configuration
	fmt.Println("To rewrite configuration to persistent storage.")
	_ = s.writeConf()

	s.mutex.Unlock()

	if s.State() == Leader {
		lindex, _ := s.log.LastLogInfo()
		cmdinfo := &DefaultJoinCommand{
			Name: name,
			Host: host,
		}
		cmdjson, _ := json.Marshal(cmdinfo)
		entry := &pb.LogEntry{
			Index:       lindex + 1,
			Term:        s.currentTerm,
			Commandname: "raft:join",
			Command:     []byte(cmdjson),
		}
		s.onMemberChanged(entry)
	}
	return nil
}

func (s *server) RemovePeer(name string, host string) error {
	s.mutex.Lock()
	if s.peers[host] == nil || s.conf.Host == host {
		s.mutex.Unlock()
		return nil
	}

	delete(s.peers, host)

	// to flush configuration
	fmt.Println("To rewrite configuration to persistent storage.")
	_ = s.writeConf()
	s.mutex.Unlock()

	if s.State() == Leader {
		lindex, _ := s.log.LastLogInfo()
		cmdinfo := &DefaultLeaveCommand{
			Name: name,
			Host: host,
		}
		cmdjson, _ := json.Marshal(cmdinfo)
		entry := &pb.LogEntry{
			Index:       lindex + 1,
			Term:        s.currentTerm,
			Commandname: "raft:leave",
			Command:     []byte(cmdjson),
		}

		s.onMemberChanged(entry)
	}
	return nil
}

func (s *server) onMemberChanged(entry *pb.LogEntry) {
	findex := s.log.FirstLogIndex()
	lindex, lterm := s.log.LastLogInfo()
	for idx, _ := range s.peers {
		if s.conf.Host == s.peers[idx].Host {
			s.log.AppendEntry(&LogEntry{Entry: entry})
			continue
		}
		go s.peers[idx].RequestAppendEntries([]*pb.LogEntry{entry}, findex, lindex, lterm)
	}
}

func (s *server) onSnapShotting() {
	if s.State() != Leader {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	cmiIndex, _ := s.log.LastCommitInfo()
	backindex := len(s.log.entries) - 1
	for i := backindex; i >= 0; i-- {
		if s.log.entries[i].Entry.GetIndex() < cmiIndex {
			backindex = i
			break
		}
	}
	if backindex < 0 {
		return
	}
	s.log.entries = s.log.entries[backindex:len(s.log.entries)]
	s.log.RefreshLog()

	findex := s.log.FirstLogIndex()
	lindex, lterm := s.log.LastLogInfo()

	pbentries := [](*pb.LogEntry){}
	for _, entry := range s.log.entries {
		pbentries = append(pbentries, entry.Entry)
	}
	for idx, _ := range s.peers {
		if s.conf.Host == s.peers[idx].Host {
			continue
		}
		go s.peers[idx].RequestAppendEntries(pbentries, findex, lindex, lterm)
	}
}
