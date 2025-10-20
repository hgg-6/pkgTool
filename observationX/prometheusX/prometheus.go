package prometheusX

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func InitPrometheus(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

// PrometheusCounter 计数器
func PrometheusCounter() {
	// 创建计数器
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "hgg",            // 命名空间
		Subsystem: "hgg_XiaoWeiShu", // 子系统
		Name:      "hgg_counter",    // 名称
	})
	//  注册
	prometheus.MustRegister(counter)
	// +1，默认为0
	counter.Inc()
	// 必须是正数，不能小于0
	counter.Add(10.2)
}

// PrometheusGauge 仪表
func PrometheusGauge() {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_gauge",
	})
	prometheus.MustRegister(gauge)
	// 设置 gauge 值
	gauge.Set(12)
	// gauge Add
	gauge.Add(10.2)
	gauge.Add(-3)
	gauge.Sub(3)
}

// PrometheusHistogram 直方图
func PrometheusHistogram() {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_histogram",
		// 按照这个分桶
		Buckets: []float64{10, 50, 100, 500, 1000, 10000},
	})
	prometheus.MustRegister(histogram)
	// 观测,12.4是观测的值
	histogram.Observe(12.4)
}

// PrometheusSummary 概要、总结
func PrometheusSummary() {
	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "hgg",
		Subsystem: "hgg_XiaoWeiShu",
		Name:      "hgg_summary",
		Objectives: map[float64]float64{
			0.5:   0.01,   // 以响应时间为例：50%的观测值响应时间，在0.01的百分比内【误差在 %1】
			0.75:  0.01,   // 以响应时间为例：75%的观测值响应时间，在0.01的百分比内【误差在 %1】
			0.90:  0.005,  // 以响应时间为例：90%的观测值响应时间，在0.005的百分比内【误差在 %0.5】
			0.98:  0.002,  // 以响应时间为例：98%的观测值响应时间，在0.002的百分比内【误差在 %0.2】
			0.99:  0.001,  // 以响应时间为例：99%的观测值响应时间，在0.001的百分比内【误差在 %0.1】
			0.999: 0.0001, // 以响应时间为例：99.9%的观测值响应时间，在0.0001的百分比内【误差在 %0.01】
		},
	})
	prometheus.MustRegister(summary)
	// 观测
	// Observe 12.3是观测的值，就是响应时间有多少在哪一个区间内
	// eg: 百分之99的请求，在12.3毫秒内完成
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
