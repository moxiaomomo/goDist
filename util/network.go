package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func GetLocalIP() string {
	addrSlice, err := net.InterfaceAddrs()
	if nil != err {
		return ""
	}

	for _, addr := range addrSlice {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if nil != ipnet.IP.To4() {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// NewMultipleHostsReverseProxy creates reverse proxy instance,
// the parameter `transport` can be nil, which is used to replace the default one
func NewMultipleHostsReverseProxy(targets []*url.URL, transport *http.Transport) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		target := targets[rand.Int()%len(targets)]
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
	}

	// replace default transport by another one
	if transport != nil {
		return &httputil.ReverseProxy{
			Director:  director,
			Transport: transport,
		}
	}

	return &httputil.ReverseProxy{Director: director}
}

// HealthcheckHandler for healthcheck
func HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

//HeartbeatForRegistry HeartbeatForRegistry
func HeartbeatForRegistry(lbHost, svrHost, hcURL string, uriPath []string) error {
	data := make(url.Values)
	data["host"] = []string{svrHost}
	data["uripath"] = uriPath
	data["hcurl"] = []string{hcURL}

	url := fmt.Sprintf("http://%s/service/add", lbHost)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var regResp CommonResp
	err = json.Unmarshal(body, &regResp)

	if err != nil {

		return err
	}
	if regResp.Code != REG_WORKER_OK {
		return fmt.Errorf("Error: %s", regResp.Message)
	}
	return nil
}
