package fixedWindow

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

// FixedWindowLimiter 固定窗口限流算法
type FixedWindowLimiter struct {
	window          time.Duration // 窗口大小
	lastWindowStart time.Time     // 窗口开始时间
	cnt             int           // 窗口内请求数
	threshold       int           // 阈值
	lock            sync.Mutex    // 保护临界资源
}

func NewFixedWindowLimiter(window time.Duration, threshold int) *FixedWindowLimiter {
	return &FixedWindowLimiter{window: window, lastWindowStart: time.Now(), cnt: 0, threshold: threshold}
}

// BuildServerInterceptor 固定窗口限流算法
func (c *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		c.lock.Lock() // 加锁
		now := time.Now()
		// 判断是否到了新的窗口
		if now.After(c.lastWindowStart.Add(c.window)) {
			c.cnt = 0
			c.lastWindowStart = now
		}
		cnt := c.cnt + 1
		c.lock.Unlock() // 解锁
		if cnt <= c.threshold {
			resp, err = handler(ctx, req) // 处理请求，可以执行业务代码
			return
		}
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}
