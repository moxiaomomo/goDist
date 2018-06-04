package handler

import (
	"net/http"
	"time"

	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"
)

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func InitHandlers() {
	http.HandleFunc("/healthcheck", util.HealthcheckHandler)
	http.HandleFunc("/api/hello", HelloHandler)
}

func RegisterWithHeartbeat() {
	// TODO: to register api server
	lbhost := "127.0.0.1:4000"
	svrHost := "127.0.0.1:6000"
	uripath := "/api/hello"
	hcurl := "http://127.0.0.1:6000/healthcheck"

	err := util.HeartbeatForRegistry(lbhost, svrHost, hcurl, []string{uripath})
	if err != nil {
		logger.LogErrorf("failed to register server, err:%s\n", err)
	}

	t := time.NewTicker(time.Second * util.HEARTBEAT_INTERVAL)
	for range t.C {
		err := util.HeartbeatForRegistry(lbhost, svrHost, hcurl, []string{uripath})
		if err != nil {
			logger.LogErrorf("failed to register server, err:%s\n", err)
		}
	}
}

func StartAPIServer(listenHost string) {
	InitHandlers()

	go RegisterWithHeartbeat()

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
