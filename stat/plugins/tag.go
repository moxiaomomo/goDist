package plugins

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var (
	//TracingComponentTag tags
	TracingComponentTag = opentracing.Tag{Key: string(ext.Component), Value: "gRPC"}
)
