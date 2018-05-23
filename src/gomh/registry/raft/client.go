package raft

import (
	"fmt"
	"net/http"
	"time"
)

func redirect(w http.ResponseWriter, r *http.Request, s *server) {

	url := fmt.Sprintf("http://%s%s", s.currentLeaderExHost, r.RequestURI)
	fmt.Println(url)
	req, err := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: time.Duration(time.Second),
	}
	_, err = client.Do(req)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Member changed OK."))
	}
}

func JoinHandler(w http.ResponseWriter, r *http.Request, s *server) {
	if s.State() == Leader {
		r.ParseForm()
		s.AddPeer(r.Form["name"][0], r.Form["host"][0])
	} else {
		// TODO: pass this request to the leader
		redirect(w, r, s)
	}
}

func LeaveHandler(w http.ResponseWriter, r *http.Request, s *server) {
	if s.State() == Leader {
		r.ParseForm()
		s.RemovePeer(r.Form["name"][0], r.Form["host"][0])
	} else {
		// TODO: pass this request to the leader
		redirect(w, r, s)
	}
}

func (s *server) StartClientServe() {
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) { JoinHandler(w, r, s) })
	http.HandleFunc("/leave", func(w http.ResponseWriter, r *http.Request) { LeaveHandler(w, r, s) })
	fmt.Printf("listen client address: %s\n", s.conf.Client)
	http.ListenAndServe(s.conf.Client, nil)
}
