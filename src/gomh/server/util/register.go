package util

import (
	"encoding/json"
	"fmt"
	"gomh/util"
	"gomh/util/logger"
	"net/http"
	"net/url"
	"time"
)

var regClient = &http.Client{Timeout: 10 * time.Second}

func Register(uripath string, lbhost string, svrHost string) error {
	go func() {
		for {
			time.Sleep(time.Second * util.HEARTBEAT_INTERVAL)
			err := reportHeartbeat(uripath, lbhost, svrHost)
			if err != nil {
				logger.LogErrorf("Send heartbeat failed: %s\n", err.Error())
			}

		}
	}()
	return reportHeartbeat(uripath, lbhost, svrHost)
}

func reportHeartbeat(uripath string, lbhost string, svrHost string) error {
	data := make(url.Values)
	data["host"] = []string{svrHost}
	data["uripath"] = []string{uripath}

	url := fmt.Sprintf("http://%s/service/add?host=%s&uripath=%s", lbhost, svrHost, uripath)
	resp, err := regClient.PostForm(url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var regResp util.CommonResp
	err = json.NewDecoder(resp.Body).Decode(&regResp)

	if err != nil {
		return err
	}
	if regResp.Code != util.REG_WORKER_OK {
		return fmt.Errorf("Error: %s", regResp.Message)
	}
	return nil
}
