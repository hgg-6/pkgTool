package opentelemetryx

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
	//res, err := newResource("demo", "v0.0.1")
	res, err := o.newResource()
	if err == nil {
		o.ResErr = nil
		o.resource = res
	}
	o.ResErr = err

	prop := o.newPropagator()
	// 设置 propagator, 在客户端和服务端之间传递 tracing 的相关信息
	otel.SetTextMapPropagator(prop)

	// 初始化 trace provider
	// 这个 provider 就是用来在打点的时候构建 trace 的
	tp, err := o.newTraceProvider()
	if err != nil {
		o.ResErr = err
	}
	o.tracerProvider = tp

	// 设置 trace provider
	otel.SetTracerProvider(tp)

	return o.initOtel(), err
}

// InitOtel main方法里，defer住 tp.Shutdown(ctx)，InitOtel
func (o *OtelStr) initOtel() func(ctx context.Context) {
	return func(ctx context.Context) {
		_ = o.tracerProvider.Shutdown(ctx)
	}
}

// newResource
func (o *OtelStr) newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(o.serviceName),
			semconv.ServiceVersion(o.serviceVersion),
		))
}

// newPropagator 用于在客户端和服务端之间传递 tracing 的相关信息
func (o *OtelStr) newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

// newTraceProvider 用于初始化 trace provider
func (o *OtelStr) newTraceProvider() (*trace.TracerProvider, error) {
	//exporter, err := zipkin.New("http://localhost:9411/api/v2/spans") // zipkin exporter
	//if err != nil {
	//	return nil, err
	//}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(o.spanExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(o.resource),
	)
	return traceProvider, nil
}
