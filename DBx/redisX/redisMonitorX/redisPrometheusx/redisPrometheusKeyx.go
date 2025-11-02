package redisPrometheusx

/*
	监控缓存的命中率
*/

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type PrometheusHookKeyRate struct {
	vector *prometheus.SummaryVec
}

// NewPrometheusHookKeyRate 监控缓存get的命中率
func NewPrometheusHookKeyRate(opts prometheus.SummaryOpts) *PrometheusHookKeyRate {
	h := &PrometheusHookKeyRate{
		vector: prometheus.NewSummaryVec(opts, []string{"cmd", "key_exist"}),
	}
	prometheus.MustRegister(h.vector)
	return h
}

func (p *PrometheusHookKeyRate) DialHook(next redis.DialHook) redis.DialHook {
	return next // 透传，不监控
}

func (p *PrometheusHookKeyRate) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd) // 先执行命令，拿到 err

		// 再记录指标（不要 defer，避免 panic 影响 err 捕获）
		duration := time.Since(start).Seconds()
		cmdName := strings.ToLower(cmd.Name())
		keyExist := getLabelKeyExist(cmd)

		p.vector.WithLabelValues(cmdName, keyExist).Observe(duration)

		return err
	}
}

func (p *PrometheusHookKeyRate) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func getLabelKeyExist(cmd redis.Cmder) string {
	switch strings.ToLower(cmd.Name()) {
	case "get", "hget", "lindex", "zscore", "exists":
		if cmd.Err() == redis.Nil {
			return "false"
		} else if cmd.Err() == nil {
			return "true"
		}
	}
	return "n/a"
}
