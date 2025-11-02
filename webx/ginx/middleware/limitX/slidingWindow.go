package limitX

import (
	"gitee.com/hgg_test/pkg_tool/v2/slicex/queueX"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	window    time.Duration                    // 统计窗口大小（如1分钟）
	threshold int                              // 窗口内允许的最大请求数（阈值）
	queue     *queueX.PriorityQueue[time.Time] // 存储请求时间戳的最小堆（队首是最早请求）
	lock      sync.Mutex                       // 限流器全局锁（保护队列操作）
}

// NewSlidingWindowBuilder 滑动窗口限流器, 构造函数
//   - window: 窗口时长（如time.Second）
//   - threshold: 窗口内最大请求数（如100）
//
// 每1秒最多有100个请求/秒，即1秒内最多100个请求
func NewSlidingWindowBuilder(window time.Duration, threshold int) *SlidingWindowLimiter {
	// 初始化最小堆：比较函数用time.Time.Before（早的时间戳更小，队首是最旧请求）
	return &SlidingWindowLimiter{
		window:    window,
		threshold: threshold,
		queue:     queueX.NewPriorityQueue[time.Time](func(a, b time.Time) bool { return a.Before(b) }, 0),
	}
}

// BuildServerInterceptor
//   - 构建gRPC服务端拦截器
//   - 构建gRPC UnaryServerInterceptor
func (c *SlidingWindowLimiter) BuildServerInterceptor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !c.Allow() {
			// 限流
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		// 允许请求：调用后续处理逻辑
		ctx.Next()
	}
}

// Allow
//   - 检查是否允许通过请求
//   - 判断是否允许当前请求（核心逻辑）
func (c *SlidingWindowLimiter) Allow() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now()
	windowStart := now.Add(-c.window) // 窗口左边界（当前时间往前推window）

	// 1. 清理过期请求：删除队列中所有早于windowStart的时间戳
	c.removeExpired(windowStart)
	//for {
	//	peekTime, ok := c.queue.Peek()
	//	if !ok || !peekTime.Before(windowStart) {
	//		break // 队首已处于窗口内，无需继续清理
	//	}
	//	c.queue.Pop() // 移除过期请求
	//}

	// 2. 检查是否超过阈值
	if c.queue.Size() >= c.threshold {
		return false // 超过阈值，不允许
	}

	// 3. 记录当前请求时间戳
	//c.queue.Push(now)
	c.queue.Enqueue(now)
	return true // 允许请求
}

// removeExpired 移除过期的请求时间戳
func (c *SlidingWindowLimiter) removeExpired(windowStart time.Time) {
	// 持续移除窗口开始时间之前的请求
	// 清理过期请求：删除队列中所有早于windowStart的时间戳
	for {
		peekTime, ok := c.queue.Peek()
		if !ok || !peekTime.Before(windowStart) {
			break // 队首已处于窗口内，无需继续清理
		}
		//c.queue.Pop() // 移除过期请求
		c.queue.Dequeue() // 移除过期请求
	}
}

// GetCurrentCount 获取当前队列长度（主要用于测试）
func (c *SlidingWindowLimiter) GetCurrentCount() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now()
	windowStart := now.Add(-c.window)
	c.removeExpired(windowStart)

	return c.queue.Len()
}
