package limitX

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sync/atomic"
)

type CounterLimiter struct {
	cnt       atomic.Int32 // 计数器
	threshold int32        // 阈值
}

// NewCounterBuilder 创建计数器限流算法
func NewCounterBuilder(threshold int32) *CounterLimiter {
	return &CounterLimiter{threshold: threshold}
}

// Build 计数器限流算法
func (c *CounterLimiter) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 请求进来，先占坑
		cnt := c.cnt.Add(1)
		if cnt > c.threshold {
			// 超过阈值，回滚计数并拒绝
			c.cnt.Add(-1)
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		// 在请求处理完成后减少计数
		defer c.cnt.Add(-1)
		ctx.Next()
	}
}
