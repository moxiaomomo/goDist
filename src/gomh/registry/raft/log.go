package raft

import (
	"os"
	"sync"
)

type Log struct {
	ApplyFunc func(*LogEntry, Command) (interface{}, error)

	mutex       sync.Mutex
	file        *os.File
	path        string
	entries     []*LogEntry
	commitIndex uint64
	startIndex  uint64
	startTerm   uint64
	initialized bool
}

func newLog() *Log {
	log := &Log{
		entries: make([]*LogEntry, 0),
	}
	return log
}

func (l *Log) LastInfo() (index uint64, term uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.entries) == 0 {
		return l.startIndex, l.startTerm
	}

	last := l.entries[len(l.entries)-1]
	return last.Index(), last.Term()
}

func (l *Log) UpdateCommitIndex(index uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if index > l.commitIndex {
		l.commitIndex = index
	}
}
