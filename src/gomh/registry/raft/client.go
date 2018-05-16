package raft

import (
	"fmt"
	"net/http"
)

func JoinHandler(w http.ResponseWriter, r *http.Request, s *server) {
	if s.State() == Leader {
		r.ParseForm()
		s.AddPeer(r.Form["name"][0], r.Form["host"][0])
	} else {
		// TODO: pass this request to the leader
	}
}

func LeaveHandler(w http.ResponseWriter, r *http.Request, s *server) {
	if s.State() == Leader {
		r.ParseForm()
		s.RemovePeer(r.Form["name"][0], r.Form["host"][0])
	} else {
		// TODO: pass this request to the leader
	}
}

func (s *server) StartClientServe() {
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) { JoinHandler(w, r, s) })
	http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) { LeaveHandler(w, r, s) })
	fmt.Printf("listen client address: %s\n", s.conf.Client)
	http.ListenAndServe(s.conf.Client, nil)
}
