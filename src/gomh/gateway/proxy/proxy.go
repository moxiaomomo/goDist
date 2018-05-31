package proxy

import (
	"gomh/gateway/config"
	"gomh/gateway/filter"
	"gomh/util/logger"
	"net/http"
	"sync"
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
	h := &HandleReverse{endpoint: p.cfg.LBHost}
	if err := http.ListenAndServe(p.cfg.SvrAddr, h); err != nil {
		logger.LogErrorf("proxy start failed, err:%s", err)
	}
}
