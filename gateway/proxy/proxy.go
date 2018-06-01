package proxy

import (
	"net/http"
	"sync"

	"github.com/moxiaomomo/goDist/gateway/config"
	"github.com/moxiaomomo/goDist/gateway/filter"
	"github.com/moxiaomomo/goDist/util/logger"
)

// Proxy proxy struct
type Proxy struct {
	mutex sync.RWMutex

	cfg *config.ProxyConfig

	filters   []filter.Filter
	ctx       *context
	isrunning bool
	stopped   bool
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
		ctx:     &context{},
		stopped: false,
	}
	return proxy
}

// Start start the proxy
func (p *Proxy) Start() {
	if p.isrunning {
		return
	}

	// TODO: initialize filters using configuration
	rl := &filter.RateLimitFilter{}
	rl.Init("")

	p.RegisterFilters([]filter.Filter{rl})

	h := &HandleReverse{
		endpoint: p.cfg.LBHost,
		proxy:    p,
	}
	p.isrunning = true

	// listen on specified address
	if err := http.ListenAndServe(p.cfg.SvrAddr, h); err != nil {
		logger.LogErrorf("proxy start failed, err:%s", err)
	}
}

// RegisterFilters register filters into proxy
func (p *Proxy) RegisterFilters(filters []filter.Filter) {
	for _, f := range filters {
		p.filters = append(p.filters, f)
	}
}

// DoFilteringAsBegin return (resp, nil) if all filters passed, else (resp, err)
func (p *Proxy) DoFilteringAsBegin() filter.Response {
	for _, f := range p.filters {
		resp := f.AsBegin(p.ctx)
		if resp.Code != filter.FilteredPassed {
			return resp
		}
	}
	return filter.Response{Code: filter.FilteredPassed}
}

// DoFilteringAsEnd return (resp, nil) if all filters passed, else (resp, err)
func (p *Proxy) DoFilteringAsEnd() filter.Response {
	for _, f := range p.filters {
		resp := f.AsEnd(p.ctx)
		if resp.Code != filter.FilteredPassed {
			return resp
		}
	}
	return filter.Response{Code: filter.FilteredPassed}
}
