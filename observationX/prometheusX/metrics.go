package prometheusX

import "github.com/prometheus/client_golang/prometheus"

// NewCounter 创建并注册 Counter
func (p *PrometheusStr) NewCounter(name, help string) prometheus.Counter {
	c := prometheus.NewCounter(p.counterOpts(name, help))
	p.MustRegister(c)
	return c
}

// NewCounterVec 创建并注册 CounterVec
func (p *PrometheusStr) NewCounterVec(name, help string, labels []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(p.counterOpts(name, help), labels)
	p.MustRegister(c)
	return c
}

// NewGauge 创建并注册 Gauge
func (p *PrometheusStr) NewGauge(name, help string) prometheus.Gauge {
	g := prometheus.NewGauge(p.gaugeOpts(name, help))
	p.MustRegister(g)
	return g
}

// NewGaugeVec 创建并注册 GaugeVec
func (p *PrometheusStr) NewGaugeVec(name, help string, labels []string) *prometheus.GaugeVec {
	g := prometheus.NewGaugeVec(p.gaugeOpts(name, help), labels)
	p.MustRegister(g)
	return g
}

// NewHistogram 创建并注册 Histogram
func (p *PrometheusStr) NewHistogram(name, help string, buckets []float64) prometheus.Histogram {
	h := prometheus.NewHistogram(p.histogramOpts(name, help, buckets))
	p.MustRegister(h)
	return h
}

// NewHistogramVec 创建并注册 HistogramVec
func (p *PrometheusStr) NewHistogramVec(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	h := prometheus.NewHistogramVec(p.histogramOpts(name, help, buckets), labels)
	p.MustRegister(h)
	return h
}

// NewSummary 创建并注册 Summary
func (p *PrometheusStr) NewSummary(name, help string, objectives map[float64]float64) prometheus.Summary {
	s := prometheus.NewSummary(p.summaryOpts(name, help, objectives))
	p.MustRegister(s)
	return s
}

// NewSummaryVec 创建并注册 SummaryVec
func (p *PrometheusStr) NewSummaryVec(name, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	s := prometheus.NewSummaryVec(p.summaryOpts(name, help, objectives), labels)
	p.MustRegister(s)
	return s
}
