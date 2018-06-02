package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
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

	// TODO: to register api server
	lbhost := "127.0.0.1:4000"
	svrHost := "127.0.0.1:6000"
	uripath := "/api/hello"
	hcurl := "http://127.0.0.1:6000/healthcheck"

	data := make(url.Values)
	data["host"] = []string{svrHost}
	data["uripath"] = []string{uripath}
	data["hcurl"] = []string{hcurl}

	url := fmt.Sprintf("http://%s/service/add", lbhost)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(url, data)
	if err != nil {
		logger.LogErrorf("register api server failed, err:%s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
}

func StartAPIServer(listenHost string) {
	InitHandlers()

	logger.LogInfof("to start server on port: %s\n", listenHost)
	http.ListenAndServe(listenHost, nil)
}
