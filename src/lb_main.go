package main

import (
	"common"
	"encoding/json"
	"golb"
	"logger"
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

	regErr := golb.AddWorker(golb.Worker{Host: host, Heartbeat: time.Now().Unix()})
	if regErr == nil {
		logger.LogInfof("Suc to register worker: %s\n", host)
	}

	regResp := common.CommonResp{
		Code:    common.REG_WORKER_OK,
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

	golb.RemoveWorker(golb.Worker{Host: host})

	regResp := common.CommonResp{
		Code:    common.REG_WORKER_OK,
		Message: "ok",
	}
	respBody, err := json.Marshal(regResp)
	if err != nil {
		w.Write([]byte("internel server error."))
		return
	}
	w.Write(respBody)
}

func main() {
	logger.SetLogLevel(common.LOG_INFO)

	go golb.RemoveWorkerAsTimeout()
	golb.InitHandlers()
	golb.SetLBPolicy(common.LB_ROUNDROBIN)

	http.HandleFunc("/add", AddHandler)
	http.HandleFunc("/remove", RemoveHandler)

	logger.LogInfo("to start server on port: 8088")
	http.ListenAndServe(":8088", nil)
}
