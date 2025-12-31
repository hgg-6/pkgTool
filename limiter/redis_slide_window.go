package limiter

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

//go:embed slide_window.lua
var luaScript string

// RedisSlideWindowKLimiter Redis滑动窗口限流器
type RedisSlideWindowKLimiter struct {
	cmd redis.Cmdable

	prefix   string        // Redis key前缀，用于区分不同业务
	interval time.Duration // 窗口大小
	rate     int           // 阈值，窗口内允许的最大请求数

	// 用于生成唯一请求ID的计数器
	requestIDCounter atomic.Uint64
}

// NewRedisSlideWindowKLimiter 创建一个Redis滑动窗口限流器
//   - cmd: redis命令接口（如redis.Client）
//   - interval: 窗口大小（例如time.Second，表示每秒最多允许rate个请求）
//   - rate: 阈值，窗口内允许的最大请求数
func NewRedisSlideWindowKLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlideWindowKLimiter {
	return &RedisSlideWindowKLimiter{
		cmd:      cmd,
		prefix:   "limiter:slide_window:", // 默认前缀
		interval: interval,
		rate:     rate,
	}
}

// NewRedisSlideWindowKLimiterWithPrefix 创建带自定义前缀的Redis滑动窗口限流器
func NewRedisSlideWindowKLimiterWithPrefix(cmd redis.Cmdable, prefix string, interval time.Duration, rate int) *RedisSlideWindowKLimiter {
	if prefix == "" {
		prefix = "limiter:slide_window:"
	}
	return &RedisSlideWindowKLimiter{
		cmd:      cmd,
		prefix:   prefix,
		interval: interval,
		rate:     rate,
	}
}

// generateRequestID 生成唯一请求ID
func (b *RedisSlideWindowKLimiter) generateRequestID() string {
	// 使用UUID和时间戳的组合确保唯一性
	timestamp := time.Now().UnixNano()
	counter := b.requestIDCounter.Add(1)
	return fmt.Sprintf("%s:%d:%d", uuid.New().String(), timestamp, counter)
}

// Limit 检查是否触发限流（使用自动生成的请求ID）
//   - key: 限流的业务key
//   - 返回值：true表示触发限流，false表示通过
//   - 错误：Redis操作错误
func (b *RedisSlideWindowKLimiter) Limit(ctx context.Context, key string) (bool, error) {
	// 自动生成请求ID
	requestID := b.generateRequestID()
	return b.LimitWithRequestID(ctx, key, requestID)
}

// LimitWithRequestID 检查是否触发限流（使用指定的请求ID）
//   - key: 限流的业务key
//   - requestID: 唯一请求标识符，用于Lua脚本的member生成
//   - 返回值：true表示触发限流，false表示通过
//   - 错误：Redis操作错误
func (b *RedisSlideWindowKLimiter) LimitWithRequestID(ctx context.Context, key, requestID string) (bool, error) {
	if key == "" {
		return false, errors.New("限流key不能为空")
	}

	// 如果requestID为空，使用自动生成的
	if requestID == "" {
		requestID = b.generateRequestID()
	}

	// 构建完整的Redis key
	fullKey := b.prefix + key

	// 准备Lua脚本参数
	args := []interface{}{
		b.interval.Milliseconds(), // ARGV[1]: 窗口大小（毫秒）
		b.rate,                    // ARGV[2]: 阈值
		time.Now().UnixMilli(),    // ARGV[3]: 当前时间戳（毫秒）
		requestID,                 // ARGV[4]: 唯一请求ID
	}

	// 执行Lua脚本
	result, err := b.cmd.Eval(ctx, luaScript, []string{fullKey}, args...).Result()
	if err != nil {
		// 如果是Redis错误，包装返回
		if errors.Is(err, redis.Nil) {
			// Lua脚本返回nil的情况
			return false, fmt.Errorf("lua脚本返回nil: %w", err)
		}
		return false, fmt.Errorf("执行限流Lua脚本失败: %w", err)
	}

	// 解析结果
	return b.parseLuaResult(result)
}

// LimitWithCustomKey 使用自定义完整key进行检查（不添加前缀）
func (b *RedisSlideWindowKLimiter) LimitWithCustomKey(ctx context.Context, fullKey string) (bool, error) {
	return b.LimitWithCustomKeyAndRequestID(ctx, fullKey, "")
}

// LimitWithCustomKeyAndRequestID 使用自定义完整key和请求ID进行检查
func (b *RedisSlideWindowKLimiter) LimitWithCustomKeyAndRequestID(ctx context.Context, fullKey, requestID string) (bool, error) {
	if fullKey == "" {
		return false, errors.New("限流key不能为空")
	}

	// 如果requestID为空，使用自动生成的
	if requestID == "" {
		requestID = b.generateRequestID()
	}

	args := []interface{}{
		b.interval.Milliseconds(),
		b.rate,
		time.Now().UnixMilli(),
		requestID,
	}

	result, err := b.cmd.Eval(ctx, luaScript, []string{fullKey}, args...).Result()
	if err != nil {
		return false, fmt.Errorf("执行限流Lua脚本失败: %w", err)
	}

	return b.parseLuaResult(result)
}

// parseLuaResult 解析Lua脚本返回结果
func (b *RedisSlideWindowKLimiter) parseLuaResult(result interface{}) (bool, error) {
	// 尝试解析为int64
	if limitTriggered, ok := result.(int64); ok {
		// Lua脚本返回1表示限流，0表示通过
		return limitTriggered == 1, nil
	}

	// 尝试解析为string（有些redis客户端可能返回字符串）
	if strVal, ok := result.(string); ok {
		if strVal == "1" {
			return true, nil
		} else if strVal == "0" {
			return false, nil
		}
		// 尝试解析为整数
		if intVal, err := strconv.ParseInt(strVal, 10, 64); err == nil {
			return intVal == 1, nil
		}
	}

	// 尝试转换为bool类型（有些版本的redis客户端可能返回bool）
	if boolVal, ok := result.(bool); ok {
		return boolVal, nil
	}

	return false, fmt.Errorf("Lua脚本返回了意外的类型: %T, 值: %v", result, result)
}

// GetConfig 获取限流器配置
func (b *RedisSlideWindowKLimiter) GetConfig() (time.Duration, int) {
	return b.interval, b.rate
}

// GetPrefix 获取Redis key前缀
func (b *RedisSlideWindowKLimiter) GetPrefix() string {
	return b.prefix
}

// SetPrefix 设置Redis key前缀
func (b *RedisSlideWindowKLimiter) SetPrefix(prefix string) {
	if prefix == "" {
		prefix = "limiter:slide_window:"
	}
	b.prefix = prefix
}
