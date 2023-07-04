package otel_gin_plugin

import (
	"fmt"
	tchannel "github.com/uber/tchannel-go"
	"go.opentelemetry.io/otel"
	otelBridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
)

func beforeNewMethod(parameters []interface{}) {
	channelOpts := parameters[1].(**tchannel.ChannelOptions)
	if (*channelOpts) == nil {
		(*channelOpts) = &tchannel.ChannelOptions{}
	}

	otExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatal(fmt.Errorf("error creating trace exporter: %w", err))
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(otExporter))

	otelTracer := tp.Tracer("tracer_name")
	bridgeTracer, wrapperTracerProvider := otelBridge.NewTracerPair(otelTracer)
	otel.SetTracerProvider(wrapperTracerProvider)

	if (*channelOpts).Tracer == nil {
		(*channelOpts).Tracer = bridgeTracer
	} else {
		fmt.Println("channelOpts.Tracer is not nil")
	}
}

func afterNewMethod(ret []interface{}) {

}
