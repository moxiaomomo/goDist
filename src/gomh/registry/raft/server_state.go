package raft

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type ServerState struct {
	CommitIndex uint64 `json:"commitIndex"`
	Term        uint64 `json:"term"`
	VoteFor     string `json:"voteFor"`
}

// save data into file
func (s *server) FlushState() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state := &ServerState{
		CommitIndex: s.log.CommitIndex(),
		Term:        s.currentTerm,
	}
	d, err := json.Marshal(state)
	if err != nil {
		return err
	}

	logpath := path.Join(s.path, "internlog")
	fname := fmt.Sprintf("%s/state-%s", logpath, s.name)
	file, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0600)

	w := bufio.NewWriter(file)
	w.Write([]byte(d))
	w.Flush()

	file.Close()
	return nil
}

// load data from file
func (s *server) LoadState() error {
	logpath := path.Join(s.path, "internlog")
	fname := fmt.Sprintf("%s/state-%s", logpath, s.name)

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil
	}
	//s.srvstate = ServerState{}
	srvstate := &ServerState{}
	if err = json.Unmarshal(b, srvstate); err != nil {
		return err
	}
	fmt.Printf("CommitIndex loaded: %d\n", srvstate.CommitIndex)
	s.log.UpdateCommitIndex(srvstate.CommitIndex)
	return nil
}
