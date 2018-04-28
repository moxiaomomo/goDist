package raft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	transporter Transporter
	log         *Log
}

type Server interface {
	Start() error
	IsRunning() bool
}

func NewServer(name, path string) (Server, error) {
	s := &server{
		name:  name,
		path:  path,
		state: Stopped,
		log:   newLog(),
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
	s.state = Follower

	loopch := make(chan int)
	go func() {
		defer func() { loopch <- 1 }()
		s.loop()
	}()
	<-loopch
	return nil
}

func (s *server) loadConf() error {
	confpath := path.Join(s.path, "raft.cfg")

	cfg, err := ioutil.ReadFile(confpath)
	if err != nil {
		fmt.Errorf("open config file failed, err:%s", err)
		return nil
	}

	conf := &Config{}
	if err = json.Unmarshal(cfg, conf); err != nil {
		return err
	}

	s.log.UpdateCommitIndex(conf.CommitIndex)
	return nil
}

func (s *server) loop() {
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		fmt.Println(i)
	}
}
