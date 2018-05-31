package proxy

import (
	"fmt"
	"gomh/gateway/config"
	"gomh/gateway/filter"
	"gomh/util/logger"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/tidwall/gjson"
)

// Proxy proxy struct
type Proxy struct {
	mutex sync.RWMutex

	cfg *config.ProxyConfig

	filters []filter.Filter
	stopped bool
}

// NewProxy creates a new proxy instance
func NewProxy(confpath string) *Proxy {
	// TODO: load config from configuration file
	cfg := &config.ProxyConfig{
		LBHost:     "127.0.0.1:4000",
		SvrAddr:    "127.0.0.1:5000",
		PathPrefix: "/",
	}

	proxy := &Proxy{
		cfg:     cfg,
		filters: []filter.Filter{},
		stopped: false,
	}
	return proxy
}

// Start start the proxy
func (p *Proxy) Start() {
	http.HandleFunc(p.cfg.PathPrefix, p.HTTPReverseHandler)
	if err := http.ListenAndServe(p.cfg.SvrAddr, nil); err != nil {
		logger.LogErrorf("proxy start failed, err:%s", err)
	}
}

// HTTPReverseHandler pass request to the api-server node
func (p *Proxy) HTTPReverseHandler(w http.ResponseWriter, r *http.Request) {
	uri := fmt.Sprintf("http://%s/service/get?uripath=/api/%s", p.cfg.LBHost, r.RequestURI)
	url, _ := url.Parse(uri)

	fmt.Println(url.String())
	workRes, err := http.Get(url.String())
	if err != nil {
		w.Write([]byte("out of service"))
		return
	}
	result, _ := ioutil.ReadAll(workRes.Body)
	workRes.Body.Close()
	svrHost := gjson.Get(string(result), "data").Get("host").String()
	logger.LogInfof("togrpc:%+v %s\v", string(result), svrHost)

	apiURI, err := url.Parse(fmt.Sprintf("http://%s", svrHost))
	if err != nil {
		w.Write([]byte("internal server error"))
		return
	}
	rproxy := httputil.NewSingleHostReverseProxy(apiURI)
	rproxy.ServeHTTP(w, r)
}
