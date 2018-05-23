package handler

import (
	"encoding/json"
	//	"fmt"
	"gomh/registry/raft"
	"sync"
)

type service struct {
	raftsrv raft.Server
	mutex   sync.RWMutex
}

type Service interface {
	Add(string, string) error
	Remove(string, string) error
}

func (s *service) Add(uripath string, host string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.raftsrv.State() != raft.Leader {
		return nil
	}
	cmd := &DefaultServiceRegCommand{
		UriPath: uripath,
		Host:    host,
	}
	cmdjson, _ := json.Marshal(cmd)
	s.raftsrv.OnAppendEntry(cmd, []byte(cmdjson))

	return nil
}

func (s *service) Remove(uripath string, host string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.raftsrv.State() != raft.Leader {
		return nil
	}
	cmd := &DefaultServiceRmCommand{
		UriPath: uripath,
		Host:    host,
	}
	cmdjson, _ := json.Marshal(cmd)
	s.raftsrv.OnAppendEntry(cmd, []byte(cmdjson))

	return nil
}

func (s *service) Start() error {
	return s.raftsrv.Start()
}

func NewService(confPath string) (*service, error) {
	raftsvr, err := raft.NewServer("raft", confPath)
	if err != nil {
		return nil, err
	}

	sv := &service{
		raftsrv: raftsvr,
	}
	// register handlers
	RegistryHandler(sv)
	// register commands
	sv.raftsrv.RegisterCommand(&DefaultServiceRegCommand{})
	sv.raftsrv.RegisterCommand(&DefaultServiceRmCommand{})
	return sv, nil
}
