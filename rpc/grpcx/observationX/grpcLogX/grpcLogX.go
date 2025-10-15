package grpcLogX

import (
	"context"
	"fmt"
	"gitee.com/hgg_test/pkg_tool/v2/logx"
	"gitee.com/hgg_test/pkg_tool/v2/rpc/grpcx/observationX"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"time"
)

type InterceptorBuilder struct {
	l logx.Loggerx
	observationX.Builder
}

func NewInterceptorBuilder(l logx.Loggerx) *InterceptorBuilder {
	return &InterceptorBuilder{l: l}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		event := "normal"
		defer func() {
			// 最终输出日志
			cost := time.Since(start)

			// 发生了 panic
			if rec := recover(); rec != nil {
				switch re := rec.(type) {
				case error:
					err = re
				default:
					err = fmt.Errorf("%v", rec)
				}
				event = "recover"
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				err = status.New(codes.Internal, "panic, err "+err.Error()).Err()
			}
			
			fields := []logx.Field{
				// unary stream 是 grpc 的两种调用形态
				logx.String("type", "unary"),
				logx.Int64("cost", cost.Milliseconds()),
				logx.String("event", event),
				logx.String("method", info.FullMethod),
				// 客户端的信息
				logx.String("peer", b.PeerName(ctx)),
				logx.String("peer_ip", b.PeerIP(ctx)),
			}
			st, _ := status.FromError(err)
			if st != nil {
				// 错误码
				fields = append(fields, logx.String("code", st.Code().String()))
				fields = append(fields, logx.String("code_msg", st.Message()))
			}

			b.l.Info("RPC调用", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}
