package raft

type Peer struct {
	server *server
	Name   string
	Host   string
}
