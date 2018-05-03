package raft

import (
	"os"
	"sync"
)

type Log struct {
	ApplyFunc func(*LogEntry, Command) (interface{}, error)

	mutex       sync.RWMutex
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

func (l *Log) CurrentIndex() uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.currentIndex()
}

func (l *Log) currentIndex() uint64 {
	if len(l.entries) == 0 {
		return l.startIndex
	}
	return l.entries[len(l.entries)-1].Index()
}

func (l *Log) nextIndex() uint64 {
	return l.currentIndex() + 1
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

func (l *Log) currentTerm() uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if len(l.entries) == 0 {
		return l.startTerm
	}
	return l.entries[len(l.entries)-1].Term()
}

func (l *Log) UpdateCommitIndex(index uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if index > l.commitIndex {
		l.commitIndex = index
	}
}

func (l *Log) LogInit() {

}
