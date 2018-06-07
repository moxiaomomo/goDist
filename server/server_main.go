package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/moxiaomomo/goDist/server/config"
	"github.com/moxiaomomo/goDist/server/handler"
	"github.com/moxiaomomo/goDist/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/tidwall/gjson"

	"github.com/moxiaomomo/goDist/stat/plugins"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
)

// LoadConfig LoadConfig
func LoadConfig(confPath string) (map[string]interface{}, string, error) {
	cfg, err := ioutil.ReadFile(confPath)
	if err != nil {
		fmt.Printf("LoadConfig failed, err:%s\n", err.Error())
		return nil, "", err
	}
	m, ok := gjson.Parse(string(cfg)).Value().(map[string]interface{})
	if !ok {
		return nil, "", errors.New("Parse config failed.")
	}
	logger.LogInfof("config:%+v\n", m)
	return m, string(cfg), nil
}

func NewJaegerTracer(serviceName string) (tracer opentracing.Tracer, closer io.Closer, err error) {
	cfg := jaegerCfg.Configuration{
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  "127.0.0.1:6831",
		},
	}
	tracer, closer, err = cfg.New(
		serviceName,
		jaegerCfg.Logger(jaeger.StdLogger),
	)
	//defer closer.Close()

	if err != nil {
		return
	}
	opentracing.SetGlobalTracer(tracer)
	return
}

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

	//open tracing
	tracer, _, err := NewJaegerTracer(conf.ServiceName)
	if err != nil {
		grpclog.Errorf("new tracer err %v , continue", err)
	}
	if tracer != nil {
		servOpts = append(servOpts, grpc.UnaryInterceptor(plugins.OpentracingServerInterceptor(tracer)))
	}

	handler.StartServer(conf, servOpts)
}
