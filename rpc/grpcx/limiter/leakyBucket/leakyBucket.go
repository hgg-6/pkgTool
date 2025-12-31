package leakyBucket

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LeakyBucketLimiter 漏桶限流算法
// 漏桶算法特点：
// 1. 固定容量的桶
// 2. 以恒定速率处理请求（漏水）
// 3. 当桶满时，新请求被拒绝（溢出）
type LeakyBucketLimiter struct {
	capacity  int           // 桶容量
	rate      time.Duration // 漏水速率（产生令牌的间隔）
	bucket    chan struct{} // 漏桶，缓冲channel模拟桶容量
	closeCh   chan struct{} // 关闭信号
	closeOnce sync.Once
	wg        sync.WaitGroup // 等待goroutine退出
}

// NewLeakyBucketLimiter 创建漏桶限流器
// capacity: 桶容量，最多可累积的请求数
// rate: 漏水速率，即处理请求的最小间隔时间
func NewLeakyBucketLimiter(capacity int, rate time.Duration) *LeakyBucketLimiter {
	if capacity <= 0 {
		capacity = 1
	}
	if rate <= 0 {
		rate = time.Millisecond * 100
	}

	limiter := &LeakyBucketLimiter{
		capacity: capacity,
		rate:     rate,
		bucket:   make(chan struct{}, capacity),
		closeCh:  make(chan struct{}),
	}

	// 启动漏水（令牌生成）goroutine
	limiter.wg.Add(1)
	go limiter.leakWater()

	return limiter
}

// leakWater 漏水过程：以固定速率向桶中添加令牌
func (l *LeakyBucketLimiter) leakWater() {
	defer l.wg.Done()

	ticker := time.NewTicker(l.rate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 尝试向桶中添加一个令牌（非阻塞）
			select {
			case l.bucket <- struct{}{}:
				// 成功添加令牌
			default:
				// 桶已满，丢弃令牌（正常情况）
			}
		case <-l.closeCh:
			// 收到关闭信号，退出goroutine
			return
		}
	}
}

// BuildServerInterceptor 构建gRPC服务端拦截器
func (l *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-l.bucket:
			// 成功从桶中取出一个令牌（漏水），允许处理请求
			return handler(ctx, req)
		case <-ctx.Done():
			// 请求上下文已取消
			return nil, ctx.Err()
		default:
			// 桶为空，触发限流
			return nil, status.Errorf(codes.ResourceExhausted,
				"漏桶限流：桶容量 %d，漏水速率 %v", l.capacity, l.rate)
		}
	}
}

// Close 关闭限流器，释放资源
func (l *LeakyBucketLimiter) Close() error {
	l.closeOnce.Do(func() {
		close(l.closeCh)
		l.wg.Wait() // 等待漏水goroutine退出
		close(l.bucket)
	})
	return nil
}

// Allow 检查是否允许请求通过（可用于非gRPC场景）
func (l *LeakyBucketLimiter) Allow(ctx context.Context) bool {
	select {
	case <-l.bucket:
		return true
	case <-ctx.Done():
		return false
	default:
		return false
	}
}

// AllowWithTimeout 带超时的检查
func (l *LeakyBucketLimiter) AllowWithTimeout(timeout time.Duration) bool {
	if timeout <= 0 {
		return l.Allow(context.Background())
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return l.Allow(ctx)
}
