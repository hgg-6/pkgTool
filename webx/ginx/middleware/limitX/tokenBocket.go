package limitX

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TokenBucketLimiter 令牌桶限流算法
type TokenBucketLimiter struct {
	// 隔多久产生一个令牌
	interval  time.Duration // 令牌产生的时间间隔
	buckets   chan struct{} // 令牌桶
	closeCh   chan struct{} // 关闭信号
	closeOnce sync.Once     // 关闭信号只关闭一次
	started   bool          // 标记是否已启动goroutine
	mu        sync.Mutex    // 保护started字段
}

// NewTokenBucketBuilder 创建令牌桶限流算法
//   - interval 令牌产生时间间隔
//   - capacity 令牌桶容量
func NewTokenBucketBuilder(interval time.Duration, capacity int) *TokenBucketLimiter {
	bucket := make(chan struct{}, capacity)
	closec := make(chan struct{})
	limiter := &TokenBucketLimiter{
		interval: interval,
		buckets:  bucket,
		closeCh:  closec,
		started:  false,
	}
	// 在构造函数中启动goroutine
	limiter.startTokenGenerator()
	return limiter
}

// startTokenGenerator 启动令牌生成goroutine（只启动一次）
func (c *TokenBucketLimiter) startTokenGenerator() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return // 已启动，避免重复
	}
	c.started = true

	// 发令牌
	ticker := time.NewTicker(c.interval)
	go func() {
		defer ticker.Stop() // 确保 ticker 被正确停止，防止资源泄漏
		for {
			select {
			case <-ticker.C:
				select {
				case c.buckets <- struct{}{}:
				default:
					// bucket 满了
				}
			case <-c.closeCh:
				return
			}
		}
	}()
}

// Build 令牌桶限流算法
func (c *TokenBucketLimiter) Build() gin.HandlerFunc {
	// 取令牌
	return func(ctx *gin.Context) {
		select {
		case <-c.buckets: // 从桶中取一个令牌
			ctx.Next() // 处理请求，可以执行业务代码
		// 限流
		default:
			// 做法1, 直接返回
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
			// 做法2，等待超时了再退出【根据context】
			//case <-ctx.Done():
			//	return nil, ctx.Err()
		}
	}
}

// Close 关闭限流器，释放资源
func (c *TokenBucketLimiter) Close() error {
	c.closeOnce.Do(func() {
		close(c.closeCh)
	})
	return nil
}
