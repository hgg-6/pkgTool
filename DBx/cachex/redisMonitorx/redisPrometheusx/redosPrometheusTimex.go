package redisPrometheusx

/*
	监控redis命令耗时
*/

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type PrometheusHookTime struct {
	histogram *prometheus.HistogramVec
}

// NewPrometheusRedisHookTime 监控命令耗时
func NewPrometheusRedisHookTime(opts prometheus.HistogramOpts) *PrometheusHookTime {
	// 标签：命令名、是否成功、错误类型、业务标识
	h := &PrometheusHookTime{
		histogram: prometheus.NewHistogramVec(
			opts,
			[]string{"cmd", "success", "error_type", "biz"},
		),
	}
	prometheus.MustRegister(h.histogram)
	return h
}

func (p *PrometheusHookTime) DialHook(next redis.DialHook) redis.DialHook {
	return next // 连接阶段不监控（或可单独加监控）
}

func (p *PrometheusHookTime) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd) // 执行命令，获取真实错误

		duration := time.Since(start).Seconds() // Prometheus 标准单位：秒
		cmdName := strings.ToLower(cmd.Name())  // 统一小写，避免 cardinality 爆炸

		// 提取业务标识 biz
		biz, ok := ctx.Value("biz").(string)
		if biz == "" || !ok {
			biz = "unknown"
		}

		// 分析错误类型
		success := "true"
		errorType := "none"
		if err != nil {
			success = "false"
			if err == redis.Nil {
				errorType = "key_not_found"
			} else {
				errorType = "other" // 可扩展为 timeout, connection_error 等
			}
		}

		// 记录指标
		p.histogram.WithLabelValues(cmdName, success, errorType, biz).Observe(duration)

		return err
	}
}

func (p *PrometheusHookTime) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next // 如需监控 pipeline，可类似 ProcessHook 实现
}
