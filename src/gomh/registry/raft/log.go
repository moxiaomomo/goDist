package raft

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Log struct {
	ApplyFunc func(*LogEntry, Command) (interface{}, error)

	mutex sync.RWMutex
	file  *os.File
	path  string
	//	entries     []*LogEntry
	//	units       []*LogUnit
	entries     []*LogUnit
	toc_entry   LogUnit
	logIndexEnd uint64
	initialized bool
}

func newLog() *Log {
	log := &Log{
		//		entries: make([]*LogEntry, 0),
		entries:     make([]*LogUnit, 0),
		toc_entry:   LogUnit{},
		logIndexEnd: 0,
	}
	return log
}

func (l *Log) CurrentLogIndexEnd() uint64 {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.logIndexEnd
}

func (l *Log) LastCommitedInfo() (index uint64, term uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.entries) == 0 {
		return 0, 0
	}

	last := l.entries[len(l.entries)-1]
	return last.CurIndexStart, last.CurTerm
}

func (l *Log) LogInit(path string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var err error
	l.file, err = os.OpenFile(path, os.O_RDWR, 0600)
	l.path = path

	if err != nil {
		if os.IsNotExist(err) {
			l.file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
			if err == nil {
				l.initialized = true
			}
			return err
		}
		return err
	}

	startIndex := int64(0)
	for {
		lunit := &LogUnit{}
		startIndex, err = lunit.load(l.file, startIndex)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return fmt.Errorf("Failed to recover raft.log: %v", err)
			}
		}
		l.entries = append(l.entries, lunit)
		l.logIndexEnd = uint64(startIndex)
	}
	l.initialized = true
	return nil
}

func (l *Log) Commite(lu *LogUnit, file *os.File) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	endidx, err := lu.dump(file)
	if err != nil {
		return err
	}
	l.entries = append(l.entries, lu)
	l.logIndexEnd = uint64(endidx)
	return nil
}
