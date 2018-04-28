package main

import (
	"fmt"
	"gomh/registry/raft"
	"os"
)

func main() {
	raftsvr, err := raft.NewServer("testsvr", "raft")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = raftsvr.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
