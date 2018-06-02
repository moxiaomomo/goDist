package handler

import (
	"net/http"

	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"
)

// StartHealthCheckServer healthcheck
func StartHealthCheckServer(listenHost string) {
	http.HandleFunc("/healthcheck", util.HealthcheckHandler)
	logger.LogInfof("to start http server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
