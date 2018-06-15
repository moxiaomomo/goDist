package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/moxiaomomo/goDist/api/config"
	"github.com/moxiaomomo/goDist/api/handler"
	"github.com/moxiaomomo/goDist/stat/plugins"
	"github.com/moxiaomomo/goDist/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func main() {
	b, err := ioutil.ReadFile("./config/api.conf")
	if err != nil {
		logger.LogErrorf("failed to open configuration file\n")
		return
	}

	conf := &config.APIConfig{}
	if err = json.Unmarshal(b, conf); err != nil {
		logger.LogErrorf("failed to load configuration\n")
		return
	}

	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	tracer, _, err := plugins.NewJaegerTracer(conf.ServiceName, "127.0.0.1:6831")
	if err != nil {
		grpclog.Errorf("new tracer err %v , continue", err)
	}
	if tracer != nil {
		dialOpts = append(dialOpts, plugins.DialOption(tracer))
	}

	handler.StartAPIServer(conf, dialOpts)
}
