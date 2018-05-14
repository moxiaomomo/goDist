package raft

import (
	"net/http"
)

func JoinHandler(w http.ResponseWriter, r *http.Request, s *server) {
}

func LeaveHandler(w http.ResponseWriter, r *http.Request, s *server) {

}

func (s *server) StartClientServe() {
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) { JoinHandler(w, r, s) })
	http.HandleFunc("leave", func(w http.ResponseWriter, r *http.Request) { LeaveHandler(w, r, s) })
	http.ListenAndServe(":8080", nil)
}
