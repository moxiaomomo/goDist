package raft

import (
	"io"
)

type JoinCommand interface {
	Command
	NodeName() string
}

type DefaultJoinCommand struct {
	Name           string `json:"name"`
	ConnectionInfo string `json:"connectioninfo"`
}

type LeaveCommand interface {
	Command
	NodeName() string
}

type DefaultLeaveCommand struct {
	Name string `json:"name"`
}

type NOPCommand struct {
}

func (c *DefaultJoinCommand) CommandName() string {
	return "raft:join"
}

func (c *DefaultJoinCommand) Apply(server Server) (interface{}, error) {
	err := server.AddPeer(c.Name, c.ConnectionInfo)
	return []byte("join"), err
}

func (c *DefaultJoinCommand) NodeName() string {
	return c.Name
}

func (c *DefaultLeaveCommand) CommandName() string {
	return "raft:leave"
}

func (c *DefaultLeaveCommand) Apply(server Server) (interface{}, error) {
	err := server.RemovePeer(c.Name)
	return []byte("leave"), err
}

func (c *DefaultLeaveCommand) NodeName() string {
	return c.Name
}

func (c NOPCommand) CommandName() string {
	return "raft:nop"
}

func (c NOPCommand) Applay(server Server) (interface{}, error) {
	return nil, nil
}

func (c NOPCommand) Encode(w *io.Writer) error {
	return nil
}

func (c NOPCommand) Decode(r io.Reader) error {
	return nil
}
