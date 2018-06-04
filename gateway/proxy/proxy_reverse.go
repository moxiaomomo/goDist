package proxy

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/moxiaomomo/goDist/gateway/filter"
	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"

	"github.com/tidwall/gjson"
)

const (
	HTTPRespOK         = 0
	HTTPRespAuthFailed = -1
	HTTPRespProcFailed = -2
)

type HandleReverse struct {
	endpoint string
	proxy    *Proxy
}

type HTTPResponse struct {
	Code    int
	Message string
	Data    interface{}
}

func (h *HandleReverse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := filter.NewContext(r)

	fresp := h.DoFilteringAsBegin(ctx)
	if fresp.Code != filter.FilteredPassed {
		httpResp := &HTTPResponse{
			Code:    HTTPRespAuthFailed,
			Message: fresp.Message,
		}
		util.WriteHTTPResponseAsJson(w, httpResp)
		return
	}

	h.doServeHTTP(w, r)

	fresp = h.DoFilteringAsEnd(ctx)
	if fresp.Code != filter.FilteredPassed {
		httpResp := &HTTPResponse{
			Code:    HTTPRespProcFailed,
			Message: fresp.Message,
		}
		util.WriteHTTPResponseAsJson(w, httpResp)
		return
	}
}

func (h *HandleReverse) doServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := fmt.Sprintf("http://%s/service/get?uripath=/api%s", h.endpoint, r.URL.Path)
	cururl, _ := url.Parse(uri)

	// logger.LogInfof("lbget res: %s\n", cururl.String())
	workRes, err := http.Get(cururl.String())
	if err != nil {
		w.Write([]byte("out of service"))
		return
	}
	result, _ := ioutil.ReadAll(workRes.Body)
	workRes.Body.Close()
	svrHost := gjson.Get(string(result), "data").Get("host").String()
	// logger.LogInfof("togrpc:%+v %s\n", string(result), svrHost)
	if svrHost == "" {
		w.WriteHeader(502)
		w.Write([]byte("out of service\n"))
		return
	}

	apiURI, err := url.Parse(fmt.Sprintf("http://%s/api%s", svrHost, r.RequestURI))
	if err != nil {
		w.Write([]byte("internal server error"))
		return
	}

	logger.LogInfof("to request apisrv: %s\n", apiURI.String())

	// // use the non-default transport, use custom timeout
	transport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			deadline := time.Now().Add(5 * time.Second)
			c, err := net.DialTimeout(netw, addr, time.Second*5)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(deadline)
			return c, nil
		},
	}
	rproxy := util.NewMultipleHostsReverseProxy([]*url.URL{apiURI}, transport)
	rproxy.ServeHTTP(w, r)
}

// DoFilteringAsBegin return (resp, nil) if all filters passed, else (resp, err)
func (h *HandleReverse) DoFilteringAsBegin(ctx filter.Context) filter.Response {
	for _, f := range h.proxy.filters {
		resp := f.AsBegin(ctx)
		if resp.Code != filter.FilteredPassed {
			return resp
		}
	}
	return filter.Response{Code: filter.FilteredPassed}
}

// DoFilteringAsEnd return (resp, nil) if all filters passed, else (resp, err)
func (h *HandleReverse) DoFilteringAsEnd(ctx filter.Context) filter.Response {
	for _, f := range h.proxy.filters {
		resp := f.AsEnd(ctx)
		if resp.Code != filter.FilteredPassed {
			return resp
		}
	}
	return filter.Response{Code: filter.FilteredPassed}
}
