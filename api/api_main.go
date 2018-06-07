package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"time"

	"github.com/moxiaomomo/goDist/api/config"
	"github.com/moxiaomomo/goDist/api/handler"
	"github.com/moxiaomomo/goDist/stat/plugins"
	"github.com/moxiaomomo/goDist/util/logger"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// func Test_apiServer(t *testing.T) {
// 	handler.StartAPIServer("127.0.0.1:6000")
// }

//NewJaegerTracer New Jaeger for opentracing
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
	//open tracing
	tracer, _, err := NewJaegerTracer(conf.ServiceName)
	if err != nil {
		grpclog.Errorf("new tracer err %v , continue", err)
	}
	if tracer != nil {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(plugins.OpenTracingClientInterceptor(tracer)))
	}

	handler.StartAPIServer(conf, dialOpts)
}
