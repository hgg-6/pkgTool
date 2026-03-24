package prometheusX

import (
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InitPrometheus 启动 Prometheus metrics 服务
func InitPrometheus(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

// MustInitPrometheus 启动 Prometheus metrics 服务，失败时 panic
// Deprecated: 推荐使用 InitPrometheus 并处理返回的 error
func MustInitPrometheus(addr string) {
	if err := InitPrometheus(addr); err != nil {
		panic(err)
	}
}

// safeRegister 安全注册 collector，处理重复注册
func safeRegister(c prometheus.Collector) error {
	if err := prometheus.Register(c); err != nil {
		if _, ok := errors.AsType[prometheus.AlreadyRegisteredError](err); ok {
			return nil // 已注册，忽略
		}
		return err
	}
	return nil
}

// PrometheusCounter 计数器示例
func PrometheusCounter() {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_counter",
	})
	_ = safeRegister(counter)
	counter.Inc()
	counter.Add(10.2)
}

// PrometheusGauge 仪表示例
func PrometheusGauge() {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_gauge",
	})
	_ = safeRegister(gauge)
	gauge.Set(12)
	gauge.Add(10.2)
	gauge.Add(-3)
	gauge.Sub(3)
}

// PrometheusHistogram 直方图示例
func PrometheusHistogram() {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_histogram",
		Buckets:   []float64{10, 50, 100, 500, 1000, 10000},
	})
	_ = safeRegister(histogram)
	histogram.Observe(12.4)
}

// PrometheusSummary 概要示例
func PrometheusSummary() {
	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_summary",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.005,
			0.98:  0.002,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	_ = safeRegister(summary)
	summary.Observe(12.3)
}

// PrometheusVector 向量
//   - 实践中，我们采集的指标很有可能是根据一些业务特征来统计的，比如说分开统计HTTP响应码是2XX的以及非2XX的。
//   - 在这种情况下，可以考虑使用 Prometheus 中的Vector 用法。
func PrometheusVector() {
	labelNames := []string{"pattern", "method", "status"}
	opts := prometheus.SummaryOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_summaryVector",
		ConstLabels: map[string]string{
			"server":  "localhost:9091",
			"evn":     "test",
			"appName": "hgg_XiaoWeiShu",
		},
		Help: "The statics info for http request",
	}

	// labelNames设置观测哪些值, summaryVector.WithLabelValues()就是观测的值
	summaryVector := prometheus.NewSummaryVec(opts, labelNames)
	// 所以最后一个 Observe 方法，可以看成是当次请求 pattern = /profile, method = get 和 status = 200 的时候，响应时间是128单位是毫秒的都是被统计的+1。
	summaryVector.WithLabelValues("/profile", "GET", "200").Observe(128)
}
