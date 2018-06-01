package handler

import (
	"io/ioutil"
	"net/http"
	"net/url"

	pb "github.com/moxiaomomo/goDist/proto/greeter"
	"github.com/moxiaomomo/goDist/util/logger"

	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	logger.LogDebug("To call SayHello.")

	// TODO: read config from configfile
	url, _ := url.Parse("http://127.0.0.1:4000/service/get?uripath=/srv/hello")
	workRes, err := http.Get(url.String())
	if err != nil {
		w.Write([]byte("out of service."))
		return
	}
	result, _ := ioutil.ReadAll(workRes.Body)
	workRes.Body.Close()
	svrHost := gjson.Get(string(result), "data").Get("host").String()
	logger.LogInfof("togrpc:%+v %s\n", string(result), svrHost)

	//	timer := ProcTimer{}
	//	timer.OnStart()
	conn, err := grpc.Dial(svrHost, grpc.WithInsecure())
	if err != nil {
		logger.LogError("grpc call failed.")
		w.Write([]byte("internel server error."))
		return
	}
	defer conn.Close()

	client := pb.NewGreeterClient(conn)
	reqbody := pb.HelloRequest{
		Name:    "xiaomo",
		Message: "just4fun",
	}
	resp, err := client.SayHello(context.Background(), &reqbody)
	//	timer.OnEnd()
	if err != nil {
		logger.LogError("call sayhello failed.")
		w.Write([]byte("internel server error."))
		return
	}

	//	worker.AsTaskFinished(timer.Duration())
	w.Write([]byte(resp.Message))
}
