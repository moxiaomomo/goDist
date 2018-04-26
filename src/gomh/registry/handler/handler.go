package handler

import (
	"encoding/json"
	"gomh/util"
	"gomh/util/logger"
	"net/http"
	"time"
)

func AddHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if host, ok := r.Form["host"]; !ok || len(host) <= 0 {
		w.Write([]byte("invalid request."))
		return
	}
	host := r.Form["host"][0]

	regErr := AddWorker(Worker{Host: host, Heartbeat: time.Now().Unix()})
	if regErr == nil {
		logger.LogInfof("Suc to register worker: %s\n", host)
	}

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

func RemoveHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if host, ok := r.Form["host"]; !ok || len(host) <= 0 {
		w.Write([]byte("invalid request."))
		return
	}
	host := r.Form["host"][0]

	RemoveWorker(Worker{Host: host})

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

func StartRegistryServer(listenHost string) {
	logger.SetLogLevel(util.LOG_INFO)

	go RemoveWorkerAsTimeout()
	SetLBPolicy(util.LB_FASTRESP)

	http.HandleFunc("/add", AddHandler)
	http.HandleFunc("/remove", RemoveHandler)

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
