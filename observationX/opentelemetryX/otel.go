package opentelemetryX

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"time"
)

type CtxFn func(ctx context.Context)

type SvcInfo struct {
	ServiceName    string
	ServiceVersion string
}

type OtelStr struct {
	serviceName    string
	serviceVersion string
	ResErr         error

	// 创建 span exporter
	spanExporter trace.SpanExporter

	resource *resource.Resource

	tracerProvider *trace.TracerProvider
}

func NewOtelStr(svc SvcInfo, spanExporter trace.SpanExporter) (CtxFn, error) {
	o := &OtelStr{
		serviceName:    svc.ServiceName,
		serviceVersion: svc.ServiceVersion,
		spanExporter:   spanExporter,
	}
	res, err := o.newResource()
	if err != nil {
		o.ResErr = err
		return nil, err
	}
	o.resource = res
	o.ResErr = nil

	otel.SetTextMapPropagator(o.newPropagator())

	tp, err := o.newTraceProvider()
	if err != nil {
		o.ResErr = err
		return nil, err
	}
	o.tracerProvider = tp
	otel.SetTracerProvider(tp)

	return o.initOtel(), nil
}

// initOtel 返回用于 defer 的 shutdown 函数：先 ForceFlush 落盘再 Shutdown。
func (o *OtelStr) initOtel() func(ctx context.Context) {
	return func(ctx context.Context) {
		_ = o.tracerProvider.ForceFlush(ctx)
		_ = o.tracerProvider.Shutdown(ctx)
	}
}

func (o *OtelStr) newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(o.serviceName),
			semconv.ServiceVersion(o.serviceVersion),
		))
}

func (o *OtelStr) newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func (o *OtelStr) newTraceProvider() (*trace.TracerProvider, error) {
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(o.spanExporter,
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(o.resource),
	)
	return traceProvider, nil
}
