package grpcLogX

import (
	"context"
	"fmt"
	"github.com/hgg-6/pkgTool/v2/logx"
	"github.com/hgg-6/pkgTool/v2/rpc/grpcx/observationX"
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
			cost := time.Since(start)

			fields := []logx.Field{
				logx.String("type", "unary"),
				logx.Int64("cost", cost.Milliseconds()),
				logx.String("event", event),
				logx.String("method", info.FullMethod),
				logx.String("peer", b.PeerName(ctx)),
				logx.String("peer_ip", b.PeerIP(ctx)),
			}

			// P0-25: recover 时收集 panic 堆栈并记入日志（旧实现收集了 stack 却从未使用）。
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
				fields = append(fields, logx.String("stack", string(stack)))
				err = status.New(codes.Internal, "panic, err "+err.Error()).Err()
			}

			// P0-25: 旧实现用 status.FromError(err) 判断，但 FromError(nil) 返回
			// (status{Code:OK}, true)，st != nil 恒真，导致成功请求也走 Error 日志。
			// 改为按 err 是否为 nil 判断：err != nil 才是真正的错误调用。
			if err != nil {
				if st, _ := status.FromError(err); st != nil {
					fields = append(fields, logx.String("code", st.Code().String()))
					fields = append(fields, logx.String("code_msg", st.Message()))
				}
				b.l.Error("RPC调用", fields...)
			} else {
				b.l.Info("RPC调用", fields...)
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}
