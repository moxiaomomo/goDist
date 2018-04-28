package raft

import (
	"gomh/registry/raft/proto"
)

type LogEntry struct {
	pb *proto.LogRequest
}

func (e *LogEntry) Index() uint64 {
	return e.pb.GetIndex()
}

func (e *LogEntry) Term() uint64 {
	return e.pb.GetTerm()
}

func (e *LogEntry) CommandName() string {
	return e.pb.GetCommandname()
}

func (e *LogEntry) Command() []byte {
	return e.pb.GetCommand()
}
