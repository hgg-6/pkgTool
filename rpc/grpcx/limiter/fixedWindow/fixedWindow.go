package fixedWindow

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FixedWindowLimiter 固定窗口限流算法
type FixedWindowLimiter struct {
	window          time.Duration // 窗口大小
	lastWindowStart time.Time     // 窗口开始时间
	cnt             int           // 窗口内请求数
	threshold       int           // 阈值
	lock            sync.Mutex    // 保护临界资源
}

// NewFixedWindowLimiter 创建固定窗口限流器
func NewFixedWindowLimiter(window time.Duration, threshold int) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		window:          window,
		lastWindowStart: time.Now(),
		cnt:             0,
		threshold:       threshold,
	}
}

// BuildServerInterceptor 构建gRPC服务端拦截器
func (c *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 加锁保护共享状态
		c.lock.Lock()

		now := time.Now()
		// 判断是否进入新的窗口
		if now.Sub(c.lastWindowStart) > c.window {
			// 重置窗口
			c.cnt = 0
			c.lastWindowStart = now
		}

		// 检查是否超过阈值
		if c.cnt >= c.threshold {
			c.lock.Unlock()
			return nil, status.Errorf(codes.ResourceExhausted,
				"固定窗口限流：当前窗口请求数 %d，阈值 %d", c.cnt, c.threshold)
		}

		// 增加计数器并允许请求
		c.cnt++
		c.lock.Unlock()

		// 处理请求
		return handler(ctx, req)
	}
}
