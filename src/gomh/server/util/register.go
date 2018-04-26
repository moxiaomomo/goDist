package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"gomh/util"
	"gomh/util/logger"
	"net/http"
	"net/url"
	"time"
)

var regClient = &http.Client{Timeout: 10 * time.Second}

func Register(host string, port int, uripath string, svrHost string) error {
	go func() {
		for {
			time.Sleep(time.Second * util.HEARTBEAT_INTERVAL)
			err := reportHeartbeat(host, port, uripath, svrHost)
			if err != nil {
				logger.LogErrorf("Send heartbeat failed: %s\n", err.Error())
			}

		}
	}()
	return reportHeartbeat(host, port, uripath, svrHost)
}

func reportHeartbeat(host string, port int, uripath string, svrHost string) error {
	data := make(url.Values)
	data["host"] = []string{fmt.Sprintf("%s:%d", host, port)}
	data["uripath"] = []string{uripath}

	url := fmt.Sprintf("http://%s/add", svrHost)
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
		return errors.New(fmt.Sprintf("Error: %s", regResp.Message))
	}
	return nil
}
