package counterLiniter

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CounterLimiter struct {
	cnt       atomic.Int32 // 当前正在处理的请求计数器
	threshold int32        // 阈值
}

// NewCounterLimiter 创建计数器限流算法
func NewCounterLimiter(threshold int32) *CounterLimiter {
	return &CounterLimiter{threshold: threshold}
}

// BuildServerInterceptor 计数器限流算法
func (c *CounterLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 请求进来，先增加计数器
		cnt := c.cnt.Add(1)

		// 检查是否超过阈值
		if cnt > c.threshold {
			// 超过阈值，回滚计数器增加（减少1）
			c.cnt.Add(-1)
			// 触发限流
			return nil, status.Errorf(codes.ResourceExhausted, "限流：当前并发数 %d，阈值 %d", cnt-1, c.threshold)
		}

		// 请求被接受处理，在处理完成后减少计数器
		defer func() {
			c.cnt.Add(-1)
		}()

		// 处理请求
		resp, err = handler(ctx, req)
		return
	}
}
