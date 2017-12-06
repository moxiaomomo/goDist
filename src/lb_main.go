package main

import (
	"common"
	"encoding/json"
	"fmt"
	"golb"
	"net/http"
)

func AddHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if host, ok := r.Form["host"]; !ok || len(host) <= 0 {
		println("111")
		w.Write([]byte("invalid request."))
		return
	}
	host := r.Form["host"][0]

	golb.AddWorker(golb.Worker{Host: host})

	regResp := common.CommonResp{
		Code:    common.REG_WORKER_OK,
		Message: "ok",
	}
	respBody, err := json.Marshal(regResp)
	if err != nil {
		println("222")
		w.Write([]byte("internel server error."))
		return
	}
	println(respBody)
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
	http.HandleFunc("/add", AddHandler)
	http.HandleFunc("/remove", RemoveHandler)
	fmt.Println("to start server on port: 8088")
	http.ListenAndServe(":8088", nil)
}
