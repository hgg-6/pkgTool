package opentelemetryX

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/trace"
)

// TestNewOtelStr_Integration 验证 OTel 初始化与 tracer span 的端到端链路。
// 依赖外部 collector（localhost:4317），默认跳过；用环境变量 OTEL_INTEGRATION=1 启用。
func TestNewOtelStr_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("OTEL_INTEGRATION") != "1" {
		t.Skip("skipping integration test; set OTEL_INTEGRATION=1 to run (requires collector at localhost:4317)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	assert.NoError(t, err)

	ct, err := NewOtelStr(SvcInfo{ServiceName: "hgg", ServiceVersion: "v0.0.1"}, exporter)
	assert.NoError(t, err)
	defer func() {
		sctx, scancel := context.WithTimeout(context.Background(), time.Second)
		defer scancel()
		ct(sctx)
	}()

	otr := NewOtelTracerStr()
	tracer := otr.NewTracer("github.com/hgg-6/pkgTool/v2/observationX/opentelemetryX")

	// 用 httptest 启动 server，请求一次后关闭，不再阻塞测试。
	runSpanOnce(t, tracer)
}

// runSpanOnce 启动一个 gin handler 创建 span，请求一次验证不 panic。
func runSpanOnce(t *testing.T, tracer trace.Tracer) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(ginCtx *gin.Context) {
		ctx, span := tracer.Start(ginCtx.Request.Context(), "top_span")
		defer span.End()

		time.Sleep(time.Millisecond * 10)
		span.AddEvent("event1")

		_, subSpan := tracer.Start(ctx, "sub_span")
		defer subSpan.End()
		subSpan.SetAttributes(attribute.String("attr1", "value1"))

		ginCtx.String(http.StatusOK, "ok")
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/test")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// TestNewOtelTracerStr 验证 tracer 包装可用（不依赖外部 collector）。
func TestNewOtelTracerStr(t *testing.T) {
	// 用 noop tracer provider 保证全局有 tracer，避免 nil。
	otel.SetTracerProvider(trace.NewNoopTracerProvider())

	otr := NewOtelTracerStr()
	tracer := otr.NewTracer("test")
	assert.NotNil(t, tracer)

	_, span := tracer.Start(context.Background(), "test_span")
	assert.NotNil(t, span)
	span.End()
}
