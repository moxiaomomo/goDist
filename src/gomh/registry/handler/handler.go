package handler

import (
	"encoding/json"
	//	"github.com/tidwall/gjson"
	"fmt"
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
	uripath := r.Form["uripath"][0]

	regErr := AddWorker(Worker{Host: host, UriPath: uripath, Heartbeat: time.Now().Unix()})
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
	uripath := r.Form["uripath"][0]

	RemoveWorker(Worker{Host: host, UriPath: uripath})

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

func StartRegistryServer(listenHost string) {
	logger.SetLogLevel(util.LOG_INFO)

	go RemoveWorkerAsTimeout()
	SetLBPolicy(util.LB_FASTRESP)

	http.HandleFunc("/add", AddHandler)
	http.HandleFunc("/remove", RemoveHandler)
	http.HandleFunc("/get", GetServerHandler)

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
