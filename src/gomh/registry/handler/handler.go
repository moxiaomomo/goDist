package handler

import (
	"encoding/json"
	//	"github.com/tidwall/gjson"
	"fmt"
	"gomh/registry/raft"
	"gomh/util"
	//	"gomh/util/logger"
	"net/http"
	//	"time"
)

func (s *service) AddHandler(w http.ResponseWriter, r *http.Request) {
	if s.raftsrv.State() != raft.Leader {
		url := fmt.Sprintf("http://%s%s", s.raftsrv.CurLeaderExHost(), r.RequestURI)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	r.ParseForm()
	if host, ok := r.Form["host"]; !ok || len(host) <= 0 {
		w.Write([]byte("invalid request."))
		return
	}
	host := r.Form["host"][0]
	uripath := r.Form["uripath"][0]

	_ = s.Add(uripath, host)

	regResp := util.CommonResp{
		Code:    util.REG_WORKER_OK,
		Message: "ok",
	}
	respBody, err := json.Marshal(regResp)
	if err != nil {
		w.Write([]byte("internel server error."))
		return
	}
	w.Write(respBody)
}

func (s *service) RemoveHandler(w http.ResponseWriter, r *http.Request) {
	if s.raftsrv.State() != raft.Leader {
		url := fmt.Sprintf("http://%s%s", s.raftsrv.CurLeaderExHost(), r.RequestURI)
		http.Redirect(w, r, url, http.StatusFound)
		return
	}

	r.ParseForm()
	if host, ok := r.Form["host"]; !ok || len(host) <= 0 {
		w.Write([]byte("invalid request."))
		return
	}
	host := r.Form["host"][0]
	uripath := r.Form["uripath"][0]

	_ = s.Remove(uripath, host)

	regResp := util.CommonResp{
		Code:    util.REG_WORKER_OK,
		Message: "ok",
	}
	respBody, err := json.Marshal(regResp)
	if err != nil {
		w.Write([]byte("internel server error."))
		return
	}
	w.Write(respBody)
}

func GetServerHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
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

func RegistryHandler(s *service) {
	//	logger.SetLogLevel(util.LOG_INFO)

	//	go RemoveWorkerAsTimeout()
	SetLBPolicy(util.LB_FASTRESP)

	s.raftsrv.RegisterHandler("/service/add", s.AddHandler)
	s.raftsrv.RegisterHandler("/service/remove", s.RemoveHandler)
	//	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) { AddHandler(w, r, s) })
	//	http.HandleFunc("/remove", RemoveHandler)
	//	http.HandleFunc("/get", GetServerHandler)
}
