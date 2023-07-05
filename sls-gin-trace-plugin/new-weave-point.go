package otel_gin_plugin

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"os"
	"time"

	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	slsProjectHeader         = "x-sls-otel-project"
	slsInstanceIDHeader      = "x-sls-otel-instance-id"
	slsAccessKeyIDHeader     = "x-sls-otel-ak-id"
	slsAccessKeySecretHeader = "x-sls-otel-ak-secret"
	slsSecurityTokenHeader   = "x-sls-otel-token"
)

func beforeNewMethod(parameters []interface{}) {
	otel.SetTracerProvider(initTraceProvider())
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
}

func afterNewMethod(ret []interface{}) {
	ginEngine := ret[0].(**gin.Engine)
	(*ginEngine).Use(middleWare)
}

func middleWare(c *gin.Context) {
	savedCtx := c.Request.Context()
	defer func() {
		c.Request = c.Request.WithContext(savedCtx)
	}()
	ctx := otel.GetTextMapPropagator().Extract(savedCtx, propagation.HeaderCarrier(c.Request.Header))
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}
	tracer := otel.GetTracerProvider().Tracer(
		"github.com/gin-gonic/gin",
		oteltrace.WithInstrumentationVersion("1.0.0"),
	)

	var spanName = c.FullPath()
	ctx, span := tracer.Start(ctx, spanName, opts...)
	defer span.End()

	c.Request = c.Request.WithContext(ctx)
	c.Next()

	status := c.Writer.Status()
	if status > 0 {
		span.SetAttributes(semconv.HTTPStatusCode(status))
	}
	if len(c.Errors) > 0 {
		span.SetAttributes(attribute.String("gin.errors", c.Errors.String()))
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
