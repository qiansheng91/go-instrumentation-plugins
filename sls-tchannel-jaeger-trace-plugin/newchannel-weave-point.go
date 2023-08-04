package otel_gin_plugin

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
	tchannel "github.com/uber/tchannel-go"
	"os"
)

func beforeNewMethod(parameters []interface{}) {
	channelOpts := parameters[1].(**tchannel.ChannelOptions)
	if (*channelOpts) == nil {
		(*channelOpts) = &tchannel.ChannelOptions{}
	}

	jaegerTracer := initJaegerProvider()

	if (*channelOpts).Tracer == nil {
		(*channelOpts).Tracer = jaegerTracer
	} else {
		fmt.Println("channelOpts.Tracer is not nil")
	}
}

func initJaegerProvider() opentracing.Tracer {
	transport := transport.NewHTTPTransport(os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT"))
	jaegerReporter := jaeger.NewRemoteReporter(transport)

	jaegerTracer, _ := jaeger.NewTracer(os.Getenv("OTEL_SERVICE_NAME"),
		jaeger.NewConstSampler(true),
		jaegerReporter)
	return jaegerTracer
}

func afterNewMethod(ret []interface{}) {

}
