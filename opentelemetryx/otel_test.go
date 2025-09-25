package opentelemetryx

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/trace"
	"testing"
	"time"
)

func TestNewOtelStr(t *testing.T) {
	// 使用zipkin的实现trace.SpanExporter
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans") // zipkin exporter
	assert.NoError(t, err)
	// 初始化全局链路追踪
	ct, err := NewOtelStr(SvcInfo{ServiceName: "hgg", ServiceVersion: "v0.0.1"}, exporter)
	assert.NoError(t, err)

	// main方法里defer住
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ct(ctx)
	}()

	// 可以在其他业务中，结构体，依赖注入方式，注入 trace.Tracer ，链路追踪方法使用【提前初始化好-全局链路追踪】
	otr := NewOtelTracerStr()
	tracer := otr.NewTracer("gitee.com/hgg_test/jksj-study/opentelemetry")

	server(tracer)
}

// ===============================
// ===============================

func server(tracer trace.Tracer) {
	server := gin.Default()
	server.GET("/test", func(ginCtx *gin.Context) {
		// 名字唯一
		//tracer := otel.Tracer("gitee.com/hgg_test/jksj-study/opentelemetry")
		var ctx context.Context = ginCtx
		// 创建 span
		ctx, span := tracer.Start(ctx, "top_span")
		defer span.End()

		time.Sleep(time.Second)
		span.AddEvent("发生了什么事情") // 添加事件，强调在某个时间点/某个时间发生了什么

		ctx, subSpan := tracer.Start(ctx, "sub_span")
		defer subSpan.End()

		subSpan.SetAttributes(attribute.String("attr1", "value1")) // 添加属性, 强调在上下文里面有什么数据
		time.Sleep(time.Millisecond * 300)

		ginCtx.String(200, "hello world, 测试span")
	})
	server.Run(":8082")
}
