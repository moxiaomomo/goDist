package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/moxiaomomo/goDist/server/config"
	"github.com/moxiaomomo/goDist/server/handler"
	"github.com/moxiaomomo/goDist/util/logger"
	gtrace "github.com/moxiaomomo/grpc-jaeger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func main() {
	b, err := ioutil.ReadFile("./config/server.conf")
	if err != nil {
		logger.LogErrorf("failed to open configuration file\n")
		return
	}

	conf := &config.ServerConfig{}
	if err = json.Unmarshal(b, conf); err != nil {
		logger.LogErrorf("failed to load configuration\n")
		return
	}

	var servOpts []grpc.ServerOption
	tracer, _, err := gtrace.NewJaegerTracer(conf.ServiceName, "127.0.0.1:6831")
	if err != nil {
		grpclog.Errorf("new tracer err %v , continue", err)
	}
	if tracer != nil {
		servOpts = append(servOpts, gtrace.ServerOption(tracer))
	}

	handler.StartServer(conf, servOpts)
}
