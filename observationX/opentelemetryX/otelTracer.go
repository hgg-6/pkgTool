package opentelemetryX

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OtelTracerStr struct {
}

func NewOtelTracerStr() *OtelTracerStr {
	return &OtelTracerStr{}
}

func (o *OtelTracerStr) NewTracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(name, opts...)
}
