package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"
)

var regClient = &http.Client{Timeout: 10 * time.Second}

func RegisterWithHealthCheck(uripath string, lbhost string, svrHost string, hcurl string) error {
	return doRegister(uripath, lbhost, svrHost, hcurl)
}

func Register(uripath string, lbhost string, svrHost string) error {
	go func() {
		for {
			time.Sleep(time.Second * util.HEARTBEAT_INTERVAL)
			err := doRegister(uripath, lbhost, svrHost, "")
			if err != nil {
				logger.LogErrorf("Send heartbeat failed: %s\n", err.Error())
			}

		}
	}()
	return doRegister(uripath, lbhost, svrHost, "")
}

func doRegister(uripath string, lbhost string, svrHost string, hcurl string) error {
	data := make(url.Values)
	data["host"] = []string{svrHost}
	data["uripath"] = []string{uripath}
	if hcurl != "" {
		data["hcurl"] = []string{hcurl}
	}

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
