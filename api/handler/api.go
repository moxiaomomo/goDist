package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/moxiaomomo/goDist/util/logger"
)

func InitHandlers() {
	http.HandleFunc("/api/hello", HelloHandler)

	// TODO: to register api server
	_, err := http.Get(fmt.Sprintf("http://127.0.0.1:4000/service/add?host=%s&uripath=%s",
		"127.0.0.1:6000", "/api/hello"))
	if err != nil {
		logger.LogErrorf("register api server failed, err:%s", err)
		os.Exit(1)
	}
}

func StartAPIServer(listenHost string) {
	InitHandlers()

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
