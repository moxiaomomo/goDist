package util

import (
	"net/http"
	"time"

	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"
)

var regClient = &http.Client{Timeout: 10 * time.Second}

func RegisterWithHealthCheck(uripath, lbHost, svrHost, hcURL string) {
	err := util.HeartbeatForRegistry(lbHost, svrHost, hcURL, []string{uripath})
	if err != nil {
		logger.LogErrorf("failed to register server, err:%s\n", err)
	}

	t := time.NewTicker(time.Second * util.HEARTBEAT_INTERVAL)
	for range t.C {
		err := util.HeartbeatForRegistry(lbHost, svrHost, hcURL, []string{uripath})
		if err != nil {
			logger.LogErrorf("failed to register server, err:%s\n", err)
		}
	}
}

func RegisterWithHeartbeat(uripath, lbHost, svrHost string) {
	err := util.HeartbeatForRegistry(lbHost, svrHost, "", []string{uripath})
	if err != nil {
		logger.LogErrorf("failed to register server, err:%s\n", err)
	}

	t := time.NewTicker(time.Second * util.HEARTBEAT_INTERVAL)
	for range t.C {
		err := util.HeartbeatForRegistry(lbHost, svrHost, "", []string{uripath})
		if err != nil {
			logger.LogErrorf("failed to register server, err:%s\n", err)
		}
	}
}
