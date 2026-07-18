package prometheusX

import "github.com/prometheus/client_golang/prometheus"

// NewCounter 创建并注册 Counter
func (p *PrometheusStr) NewCounter(name, help string) prometheus.Counter {
	c := prometheus.NewCounter(p.counterOpts(name, help))
	registered, err := p.registerOrReuse(c)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(prometheus.Counter); ok {
		return existing
	}
	return c
}

// NewCounterVec 创建并注册 CounterVec
func (p *PrometheusStr) NewCounterVec(name, help string, labels []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(p.counterOpts(name, help), labels)
	registered, err := p.registerOrReuse(c)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(*prometheus.CounterVec); ok {
		return existing
	}
	return c
}

// NewGauge 创建并注册 Gauge
func (p *PrometheusStr) NewGauge(name, help string) prometheus.Gauge {
	g := prometheus.NewGauge(p.gaugeOpts(name, help))
	registered, err := p.registerOrReuse(g)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(prometheus.Gauge); ok {
		return existing
	}
	return g
}

// NewGaugeVec 创建并注册 GaugeVec
func (p *PrometheusStr) NewGaugeVec(name, help string, labels []string) *prometheus.GaugeVec {
	g := prometheus.NewGaugeVec(p.gaugeOpts(name, help), labels)
	registered, err := p.registerOrReuse(g)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(*prometheus.GaugeVec); ok {
		return existing
	}
	return g
}

// NewHistogram 创建并注册 Histogram
func (p *PrometheusStr) NewHistogram(name, help string, buckets []float64) prometheus.Histogram {
	h := prometheus.NewHistogram(p.histogramOpts(name, help, buckets))
	registered, err := p.registerOrReuse(h)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(prometheus.Histogram); ok {
		return existing
	}
	return h
}

// NewHistogramVec 创建并注册 HistogramVec
func (p *PrometheusStr) NewHistogramVec(name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	h := prometheus.NewHistogramVec(p.histogramOpts(name, help, buckets), labels)
	registered, err := p.registerOrReuse(h)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(*prometheus.HistogramVec); ok {
		return existing
	}
	return h
}

// NewSummary 创建并注册 Summary
func (p *PrometheusStr) NewSummary(name, help string, objectives map[float64]float64) prometheus.Summary {
	s := prometheus.NewSummary(p.summaryOpts(name, help, objectives))
	registered, err := p.registerOrReuse(s)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(prometheus.Summary); ok {
		return existing
	}
	return s
}

// NewSummaryVec 创建并注册 SummaryVec
func (p *PrometheusStr) NewSummaryVec(name, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	s := prometheus.NewSummaryVec(p.summaryOpts(name, help, objectives), labels)
	registered, err := p.registerOrReuse(s)
	if err != nil {
		panic(err)
	}
	if existing, ok := registered.(*prometheus.SummaryVec); ok {
		return existing
	}
	return s
}
