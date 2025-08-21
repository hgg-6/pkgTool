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

func NewRedisSlideWindowKLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlideWindowKLimiter {
	return &RedisSlideWindowKLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (b RedisSlideWindowKLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return b.cmd.Eval(ctx, luaScript, []string{key},
		b.interval.Milliseconds(), b.rate, time.Now().UnixMilli()).Bool()
}
