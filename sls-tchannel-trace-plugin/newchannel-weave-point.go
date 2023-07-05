package otel_gin_plugin

import (
	"context"
	"fmt"
	tchannel "github.com/uber/tchannel-go"
	"go.opentelemetry.io/otel"
	otelBridge "go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"os"
	"time"
)

const (
	slsProjectHeader         = "x-sls-otel-project"
	slsInstanceIDHeader      = "x-sls-otel-instance-id"
	slsAccessKeyIDHeader     = "x-sls-otel-ak-id"
	slsAccessKeySecretHeader = "x-sls-otel-ak-secret"
	slsSecurityTokenHeader   = "x-sls-otel-token"
)

func beforeNewMethod(parameters []interface{}) {
	channelOpts := parameters[1].(**tchannel.ChannelOptions)
	if (*channelOpts) == nil {
		(*channelOpts) = &tchannel.ChannelOptions{}
	}

	traceProvider := initTraceProvider()

	otelTracer := traceProvider.Tracer("github.com/uber/tchannel-go")
	bridgeTracer, wrapperTracerProvider := otelBridge.NewTracerPair(otelTracer)
	otel.SetTracerProvider(wrapperTracerProvider)

	if (*channelOpts).Tracer == nil {
		(*channelOpts).Tracer = bridgeTracer
	} else {
		fmt.Println("channelOpts.Tracer is not nil")
	}
}

func initTraceProvider() *sdktrace.TracerProvider {
	ctx := context.Background()

	var otExporter *otlptrace.Exporter
	var err error

	if otExporter, err = initTraceExporter(ctx); err != nil {
		log.Printf("error creating trace exporter: %v", err)
		return nil
	}

	bsp := sdktrace.NewBatchSpanProcessor(otExporter)
	return sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(resource.Environment()),
	)
}

func initTraceExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	header := make(map[string]string)
	header[slsAccessKeyIDHeader] = os.Getenv("ALIYUN_ACCESS_KEY_ID")
	header[slsAccessKeySecretHeader] = os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	header[slsProjectHeader] = os.Getenv("ALIYUN_SLS_PROJECT")
	header[slsInstanceIDHeader] = os.Getenv("ALIYUN_SLS_TRACE_INSTANCE")

	return otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		otlptracegrpc.WithHeaders(header))
}

func afterNewMethod(ret []interface{}) {

}
