package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/moxiaomomo/goDist/api/config"
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

func RegisterWithHeartbeat(confPath string) {
	b, err := ioutil.ReadFile(confPath)
	if err != nil {
		logger.LogErrorf("failed to open configuration file\n")
		return
	}

	conf := &config.APIConfig{}
	if err = json.Unmarshal(b, conf); err != nil {
		logger.LogErrorf("failed to load configuration\n")
		return
	}

	err = util.HeartbeatForRegistry(conf.LBHost, conf.SvrAddr, conf.HCURL, conf.URIPath)
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
func StartAPIServer(listenHost string) {
	InitHandlers()

	go RegisterWithHeartbeat("./config/api.conf")

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
