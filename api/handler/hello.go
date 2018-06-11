package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	pb "github.com/moxiaomomo/goDist/proto/greeter"
	"github.com/moxiaomomo/goDist/util"
	"github.com/moxiaomomo/goDist/util/logger"

	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// HelloHandler rpc-client
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	logger.LogDebug("To call SayHello.")

	requrl := fmt.Sprintf("http://%s/service/get?uripath=/srv/hello", apiConf.LBHost)
	url, _ := url.Parse(requrl)
	workRes, err := http.Get(url.String())
	if err != nil {
		w.Write([]byte("out of service."))
		return
	}
	result, _ := ioutil.ReadAll(workRes.Body)
	workRes.Body.Close()
	svrHost := gjson.Get(string(result), "data").Get("host").String()
	logger.LogInfof("togrpc:%+v %s\n", string(result), svrHost)

	conn, err := grpc.Dial(svrHost, dailOpts...)
	if err != nil {
		logger.LogError("grpc call failed.")
		w.Write([]byte("internel server error."))
		return
	}
	defer conn.Close()

	r.ParseForm()
	if _, ok := r.Form["name"]; !ok {
		util.WriteHTTPResponseAsJson(w, map[string]string{"error": "invalid name"})
		return
	}
	if _, ok := r.Form["message"]; !ok {
		util.WriteHTTPResponseAsJson(w, map[string]string{"error": "invalid message"})
		return
	}

	client := pb.NewGreeterClient(conn)
	reqbody := pb.HelloRequest{
		Name:    r.Form["name"][0],
		Message: r.Form["message"][0],
	}
	resp, err := client.SayHello(context.Background(), &reqbody)

	if err != nil {
		logger.LogError("call sayhello failed.")
		w.Write([]byte("internel server error."))
		return
	}

	//	worker.AsTaskFinished(timer.Duration())
	w.Write([]byte(resp.Message))
}
