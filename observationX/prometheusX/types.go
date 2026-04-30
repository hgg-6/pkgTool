package prometheusX

import "github.com/prometheus/client_golang/prometheus"

// PrometheusStr  统一封装 Prometheus 指标
type PrometheusStr struct {
	namespace   string
	subsystem   string
	constLabels prometheus.Labels
	reg         prometheus.Registerer
	gatherer    prometheus.Gatherer
}

// Option 函数式选项模式
type Option func(*PrometheusStr)
