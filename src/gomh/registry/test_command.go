package main

import (
	"fmt"
	"gomh/registry/raft"
	"reflect"
)

func main() {
	raft.RegisterCommand(&raft.DefaultJoinCommand{})
	raft.RegisterCommand(&raft.DefaultLeaveCommand{})
	raft.RegisterCommand(&raft.NOPCommand{})

	ncmd, err := raft.NewCommand("raft:join", []byte(`{"name":"ttt","connectioninfo":"kkk"}`))
	fmt.Printf("%+v type:%s %v\n", ncmd, reflect.TypeOf(ncmd), err)

	ncmd2, err := raft.NewCommand("raft:leave", []byte(`{"name":"vvv"}`))
	fmt.Printf("%+v type:%s %v\n", ncmd2, reflect.TypeOf(ncmd2), err)
}
