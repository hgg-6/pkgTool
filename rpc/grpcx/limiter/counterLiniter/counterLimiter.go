package counterLiniter

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync/atomic"
)

type CounterLimiter struct {
	cnt       atomic.Int32 // 计数器
	threshold int32        // 阈值
}

// BuildServerInterceptor 计数器限流算法
func (c *CounterLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 请求进来，先占坑
		cnt := c.cnt.Add(1)
		defer func() {
			c.cnt.Add(-1)
		}()
		if cnt <= c.threshold {
			resp, err = handler(ctx, req) // 处理请求，可以执行业务代码
			// 返回了响应
			return
		}
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}
