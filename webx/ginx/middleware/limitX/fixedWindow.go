package limitX

import (
	"github.com/gin-gonic/gin"
	"net/http"
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

// NewFixedWindowBuilder 创建固定窗口限流算法
//   - window 窗口大小
//   - threshold 阈值
func NewFixedWindowBuilder(window time.Duration, threshold int) *FixedWindowLimiter {
	return &FixedWindowLimiter{window: window, lastWindowStart: time.Now(), cnt: 0, threshold: threshold}
}

// BuildServerInterceptor 固定窗口限流算法
func (c *FixedWindowLimiter) BuildServerInterceptor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c.lock.Lock() // 加锁
		now := time.Now()
		// 判断是否到了新的窗口
		if now.After(c.lastWindowStart.Add(c.window)) {
			c.cnt = 0
			c.lastWindowStart = now
		}
		cnt := c.cnt + 1
		c.lock.Unlock() // 解锁
		if cnt >= c.threshold {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
