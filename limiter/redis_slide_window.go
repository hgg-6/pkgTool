package limiter

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var luaScript string

type RedisSlideWindowKLimiter struct {
	cmd redis.Cmdable

	prefix   string
	interval time.Duration
	// 阈值
	rate int
}

// NewRedisSlideWindowKLimiter 创建一个 Redis 滑动窗口限流器
//   - cmd: redis.Client
//   - interval: 窗口大小【eg: time.second，每秒有rate个请求，超过则触发限流】
//   - rate: 阈值
func NewRedisSlideWindowKLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlideWindowKLimiter {
	return &RedisSlideWindowKLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

// Limit
//   - key: 限流key
//   - 返回值：true则限流，false则通过
func (b *RedisSlideWindowKLimiter) Limit(ctx context.Context, key string) (bool, error) {
	// 获取锁
	return b.cmd.Eval(ctx, luaScript, []string{key}, b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}
