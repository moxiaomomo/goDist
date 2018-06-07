package handler

import (
	"net/http"
	"time"

	"github.com/moxiaomomo/goDist/api/config"
	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"
	"google.golang.org/grpc"
)

var dailOpts []grpc.DialOption

func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func InitHandlers() {
	http.HandleFunc("/healthcheck", util.HealthcheckHandler)
	http.HandleFunc("/api/hello", HelloHandler)
}

func RegisterWithHeartbeat(conf *config.APIConfig) {
	err := util.HeartbeatForRegistry(conf.LBHost, conf.SvrAddr, conf.HCURL, conf.URIPath)
	if err != nil {
		logger.LogErrorf("failed to register server, err:%s\n", err)
	}

	t := time.NewTicker(time.Second * util.HEARTBEAT_INTERVAL)
	for range t.C {
		err := util.HeartbeatForRegistry(conf.LBHost, conf.SvrAddr, conf.HCURL, conf.URIPath)
		if err != nil {
			logger.LogErrorf("failed to register server, err:%s\n", err)
		}
	}
}

// StartAPIServer StartAPIServer
func StartAPIServer(conf *config.APIConfig, dialOpt []grpc.DialOption) {
	dailOpts = dialOpt

	InitHandlers()

	go RegisterWithHeartbeat(conf)

	logger.LogInfof("to start server on %s\n", conf.SvrAddr)
	http.ListenAndServe(conf.SvrAddr, nil)
}
