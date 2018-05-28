package handler

import (
	"gomh/util/logger"
	"net/http"
)

func InitHandlers() {
	http.HandleFunc("/hello", HelloHandler)
}

func StartGatewayServer(listenHost string) {
	InitHandlers()

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
