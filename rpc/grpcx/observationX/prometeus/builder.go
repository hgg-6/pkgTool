package prometeus

import (
	"context"
	"gitee.com/hgg_test/pkg_tool/v2/rpc/grpcx/observationX"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

type InterceptorBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	InstanceId string
	Help       string
	observationX.Builder
}

// NewInterceptorBuilder 创建 prometheus 拦截器
func NewInterceptorBuilder(namespace string, subsystem string, name string, instanceId string, help string) *InterceptorBuilder {
	return &InterceptorBuilder{Namespace: namespace, Subsystem: subsystem, Name: name, InstanceId: instanceId, Help: help}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	// pattern 是指命中的路由
	labels := []string{"type", "service", "method", "peer", "code"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Help:      b.Help,
		// Namespace 和 Subsystem 和 Name 都不能有 _ 以外的其它符号
		Name: b.Name + "_resp_time",
		ConstLabels: map[string]string{
			"instance_id": b.InstanceId,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(vector)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			sn, method := b.splitMethodName(info.FullMethod)
			code := "OK"
			if err != nil {
				st, _ := status.FromError(err)
				code = st.Code().String()
			}
			cost := float64(time.Since(start).Milliseconds())
			vector.WithLabelValues("unary", sn, method, b.PeerName(ctx), code).Observe(cost)
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func (b *InterceptorBuilder) splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}
