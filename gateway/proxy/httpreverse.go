package proxy

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/moxiaomomo/goDist/util/logger"

	"github.com/tidwall/gjson"
)

type HandleReverse struct {
	endpoint string
}

func NewMultipleHostsReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		target := targets[rand.Int()%len(targets)]
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
	}
	return &httputil.ReverseProxy{Director: director}
}

func (h *HandleReverse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := fmt.Sprintf("http://%s/service/get?uripath=/api%s", h.endpoint, r.URL.Path)
	cururl, _ := url.Parse(uri)

	logger.LogInfof("lbget res: %s\n", cururl.String())
	workRes, err := http.Get(cururl.String())
	if err != nil {
		w.Write([]byte("out of service"))
		return
	}
	result, _ := ioutil.ReadAll(workRes.Body)
	workRes.Body.Close()
	svrHost := gjson.Get(string(result), "data").Get("host").String()
	logger.LogInfof("togrpc:%+v %s\n", string(result), svrHost)

	apiURI, err := url.Parse(fmt.Sprintf("http://%s/api%s", svrHost, r.RequestURI))
	if err != nil {
		w.Write([]byte("internal server error"))
		return
	}

	logger.LogInfof("to request apisrv: %s\n", apiURI.String())
	rproxy := NewMultipleHostsReverseProxy([]*url.URL{apiURI})
	rproxy.ServeHTTP(w, r)
}
