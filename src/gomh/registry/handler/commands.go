package handler

import (
	"gomh/registry/raft"
	"time"
)

type ServiceCommand interface {
	raft.Command
	Data() interface{}
}

type DefaultServiceRegCommand struct {
	UriPath string
	Host    string
}

type DefaultServiceRmCommand struct {
	UriPath string
	Host    string
}

func (c *DefaultServiceRegCommand) CommandName() string {
	return "raft:srvreg"
}

func (c *DefaultServiceRegCommand) Apply(server raft.Server) (interface{}, error) {
	AddWorker(Worker{Host: c.Host, UriPath: c.UriPath, Heartbeat: time.Now().Unix()})
	return []byte("raft:srvreg"), nil
}

func (c *DefaultServiceRmCommand) CommandName() string {
	return "raft:srvrm"
}

func (c *DefaultServiceRmCommand) Apply(server raft.Server) (interface{}, error) {
	RemoveWorker(Worker{Host: c.Host, UriPath: c.UriPath, Heartbeat: time.Now().Unix()})
	return []byte("raft:srvrm"), nil
}
