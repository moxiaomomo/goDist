package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"

	raft "github.com/moxiaomomo/goRaft"
)

func (s *service) AddHandler(w http.ResponseWriter, r *http.Request) {
	regResp := util.CommonResp{
		Code:    util.REG_WORKER_OK,
		Message: "ok",
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.raftsrv.State() != raft.Leader {
		url := fmt.Sprintf("http://%s%s", s.raftsrv.CurLeaderExHost(), r.RequestURI)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	var (
		host    = r.PostFormValue("host")
		uripath = r.PostFormValue("uripath")
	)

	if host == "" || uripath == "" {
		regResp.Code = util.REG_WORKER_FAILED
		regResp.Message = "invalid host or uripath"
	} else {
		hcurl := r.PostFormValue("hcurl")
		err := s.Add(uripath, host, hcurl)
		if err != nil {
			regResp.Code = util.REG_WORKER_FAILED
			regResp.Message = err.Error()
		}
	}

	respBody, err := json.Marshal(regResp)
	if err != nil {
		regResp.Code = util.REG_WORKER_FAILED
		regResp.Message = "internel server error"
	}
	w.Write(respBody)
}

func (s *service) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.raftsrv.State() != raft.Leader {
		url := fmt.Sprintf("http://%s%s", s.raftsrv.CurLeaderExHost(), r.RequestURI)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	regResp := util.CommonResp{
		Code:    util.REG_WORKER_OK,
		Message: "ok",
	}

	var (
		host    = r.PostFormValue("host")
		uripath = r.PostFormValue("uripath")
	)

	if host == "" || uripath == "" {
		regResp.Code = util.REG_WORKER_FAILED
		regResp.Message = "invalid host or uripath"
	} else {
		_ = s.Remove(uripath, host)
	}

	respBody, err := json.Marshal(regResp)
	if err != nil {
		w.Write([]byte("internel server error."))
		return
	}
	w.Write(respBody)
}

func (s *service) GetServerHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if up, ok := r.Form["uripath"]; !ok || len(up) <= 0 {
		w.Write([]byte("invalid request."))
		return
	}
	uripath := r.Form["uripath"][0]

	if worker, err := GetWorker(uripath); err != nil {
		ret := `{"error":-1,"msg":"no available server."}`
		w.Write([]byte(ret))
	} else {
		ret := `{"error":0,"data":{"host":"%s"}}`
		ret = fmt.Sprintf(ret, worker.HostToCall())
		w.Write([]byte(ret))
	}
}

func (s *service) ListServerHandler(w http.ResponseWriter, r *http.Request) {
	workers := ListWorkers()
	b, _ := json.Marshal(workers)
	w.Write([]byte(b))
}

// RegisterHandler Register handlers
func RegisterHandler(s *service) {
	logger.SetLogLevel(util.LOG_INFO)

	SetLBPolicy(util.LB_FASTRESP)

	s.raftsrv.RegisterHandler("/service/add", s.AddHandler)
	s.raftsrv.RegisterHandler("/service/remove", s.RemoveHandler)
	s.raftsrv.RegisterHandler("/service/get", s.GetServerHandler)
	s.raftsrv.RegisterHandler("/service/list", s.ListServerHandler)
}
