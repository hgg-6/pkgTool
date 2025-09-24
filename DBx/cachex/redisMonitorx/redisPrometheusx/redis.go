// Package redisPrometheusx 基于prometheus监控redis
package redisPrometheusx

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"net"
	"strconv"
	"time"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opts prometheus.SummaryOpts) *PrometheusHook {
	h := &PrometheusHook{vector: prometheus.NewSummaryVec(opts, []string{"cmd", "key_exist"})}
	prometheus.MustRegister(h.vector) //  注册到prometheus
	return h
}

// DialHook 这个不用管没必要监控，是创建连接时候会执行这个回调
func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		var err error
		start := time.Now()
		defer func() {
			//biz := ctx.Value("biz")                                                                        // 获取业务信息
			duration := time.Since(start).Milliseconds()                                                   // 记录执行时间
			keyExists := err == redis.Nil                                                                  // 记录key是否存在
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExists)).Observe(float64(duration)) // 记录命令
		}()
		err = next(ctx, cmd) // 这里是执行命令
		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
