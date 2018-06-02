package util

import (
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
