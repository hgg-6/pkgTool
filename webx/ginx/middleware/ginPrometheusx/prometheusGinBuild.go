package ginPrometheusx

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type Builder struct {
	Namespace  string // 命名空间
	Subsystem  string // 子系统
	Name       string // 指标名称
	InstanceId string // 实例ID
	Help       string // 指标描述
}

type BuilderConfig Builder

func NewBuilder(builderConf BuilderConfig) PrometheusGinBuilder {
	return &Builder{
		Namespace:  builderConf.Namespace,
		Subsystem:  builderConf.Subsystem,
		Name:       builderConf.Name,
		InstanceId: builderConf.InstanceId,
		Help:       builderConf.Help,
	}
}

// BuildResponseTime 响应时间
func (b *Builder) BuildResponseTime() gin.HandlerFunc {
	// pattern 是指命中的路由
	labels := []string{"method", "pattern", "status"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{ // 指定指标
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		// 不能用 Namespace 和 Subsystem 和 Name都不能有除下划线 _ 以外的特殊字符
		Name: b.Name + "_resp_time",
		ConstLabels: map[string]string{
			"instance_id": b.InstanceId,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,   // 以响应时间为例：50%的观测值响应时间，在0.01的百分比内【误差在 %1】
			0.75:  0.01,   // 以响应时间为例：75%的观测值响应时间，在0.01的百分比内【误差在 %1】
			0.90:  0.01,   // 以响应时间为例：90%的观测值响应时间，在0.01的百分比内【误差在 %1】
			0.99:  0.001,  // 以响应时间为例：99%的观测值响应时间，在0.001的百分比内【误差在 %0.1】
			0.999: 0.0001, // 以响应时间为例：99.9%的观测值响应时间，在0.0001的百分比内【误差在 %0.01】
		},
		Help: b.Help,
	}, labels)
	// prometheus.MustRegister() 会自动注册指标，如果重复注册会 panic
	// 使用 Register 代替 MustRegister，并处理重复注册的情况
	if err := prometheus.Register(vector); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			// 如果已经注册，使用已有的 collector
			vector = are.ExistingCollector.(*prometheus.SummaryVec)
		} else {
			// 其他错误，panic
			panic(err)
		}
	}

	return func(ctx *gin.Context) {
		start := time.Now()

		defer func() {
			// 准备上报 Prometheus
			duration := time.Since(start).Milliseconds()                                             // 记录为毫秒
			method := ctx.Request.Method                                                             // 获取请求方法
			pattern := ctx.FullPath()                                                                // 获取完整路径
			status := ctx.Writer.Status()                                                            // 获取状态码
			vector.WithLabelValues(method, pattern, strconv.Itoa(status)).Observe(float64(duration)) // 指定标签
		}()

		// 执行完业务逻辑
		ctx.Next()
	}
}

// BuildActiveRequest 活跃请求
func (b *Builder) BuildActiveRequest() gin.HandlerFunc {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		// 不能用 Namespace 和 Subsystem 和 Name都不能有除下划线 _ 以外的特殊字符
		Name: b.Name + "_http_req",
		ConstLabels: map[string]string{
			"instance_id": b.InstanceId,
		},
		Help: b.Help,
	})
	// prometheus.MustRegister() 会自动注册指标，如果重复注册会 panic
	// 使用 Register 代替 MustRegister，并处理重复注册的情况
	if err := prometheus.Register(gauge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			// 如果已经注册，使用已有的 collector
			gauge = are.ExistingCollector.(prometheus.Gauge)
		} else {
			// 其他错误，panic
			panic(err)
		}
	}

	return func(ctx *gin.Context) {
		gauge.Inc()       // 每次请求加一
		defer gauge.Dec() // 请求结束后减一
		ctx.Next()
	}
}
