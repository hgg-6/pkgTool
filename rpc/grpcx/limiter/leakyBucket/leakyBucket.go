package leakyBucket

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

// LeakyBucketLimiter 漏桶限流算法
type LeakyBucketLimiter struct {
	// 隔多久产生一个令牌
	interval  time.Duration
	closeCh   chan struct{}
	closeOnce sync.Once
}

func NewLeakyBucketLimiter(interval time.Duration) *LeakyBucketLimiter {
	closeC := make(chan struct{})
	return &LeakyBucketLimiter{interval: interval, closeCh: closeC}
}

// BuildServerInterceptor 漏桶限流算法
func (c *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	ticker := time.NewTicker(c.interval)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-ticker.C:
			return handler(ctx, req)
		case <-c.closeCh:
			// 限流器已经关了
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		//做法1
		default:
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
			// 做法2
			//case <-ctx.Done():
			//	return nil, ctx.Err()
		}
	}
}
func (c *LeakyBucketLimiter) Close() error {
	c.closeOnce.Do(func() {
		close(c.closeCh)
	})
	return nil
}
