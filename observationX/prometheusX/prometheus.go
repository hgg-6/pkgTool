package prometheusX

import (
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ErrAlreadyRegistered 重复注册时返回的 sentinel error
var ErrAlreadyRegistered = errors.New("prometheus: collector already registered")

// New 创建 PrometheusX 实例
func New(opts ...Option) *PrometheusStr {
	p := &PrometheusStr{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// WithNamespace 设置全局 namespace
func WithNamespace(ns string) Option {
	return func(p *PrometheusStr) { p.namespace = ns }
}

// WithSubsystem 设置全局 subsystem
func WithSubsystem(sub string) Option {
	return func(p *PrometheusStr) { p.subsystem = sub }
}

// WithConstLabels 设置全局常量标签
func WithConstLabels(labels map[string]string) Option {
	return func(p *PrometheusStr) { p.constLabels = labels }
}

// WithRegisterer 设置自定义 Registerer（默认 prometheus.DefaultRegisterer）
func WithRegisterer(reg prometheus.Registerer) Option {
	return func(p *PrometheusStr) { p.reg = reg }
}

// WithGatherer 设置自定义 Gatherer（默认 prometheus.DefaultGatherer），影响 Handler() 输出
func WithGatherer(g prometheus.Gatherer) Option {
	return func(p *PrometheusStr) { p.gatherer = g }
}

// Register 注册 collector；重复注册时返回 ErrAlreadyRegistered。
func (p *PrometheusStr) Register(c prometheus.Collector) error {
	_, err := p.registerOrGet(c)
	return err
}

// registerOrGet 注册 collector，重复注册时返回已存在的 collector 和 ErrAlreadyRegistered。
// 旧 Register 丢弃了 AlreadyRegisteredError.ExistingCollector，导致工厂方法返回
// 未注册的新实例，Inc 等操作不会出现在 /metrics（指标静默丢失）。
func (p *PrometheusStr) registerOrGet(c prometheus.Collector) (prometheus.Collector, error) {
	reg := p.reg
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	if err := reg.Register(c); err != nil {
		var alreadyErr prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyErr) {
			return alreadyErr.ExistingCollector, ErrAlreadyRegistered
		}
		return nil, err
	}
	return c, nil
}

// registerOrReuse 供工厂方法使用：重复注册时返回已存在的 collector 且不报错，
// 其它注册错误才返回 error（工厂方法对其它错误 panic）。
func (p *PrometheusStr) registerOrReuse(c prometheus.Collector) (prometheus.Collector, error) {
	registered, err := p.registerOrGet(c)
	if err != nil {
		if errors.Is(err, ErrAlreadyRegistered) {
			return registered, nil
		}
		return nil, err
	}
	return registered, nil
}

// MustRegister 注册 collector，重复注册被忽略，其他错误 panic
func (p *PrometheusStr) MustRegister(c prometheus.Collector) {
	if _, err := p.registerOrGet(c); err != nil {
		if errors.Is(err, ErrAlreadyRegistered) {
			return
		}
		panic(err)
	}
}

// Handler 返回 /metrics 端点对应的 http.Handler
// 若使用了自定义 Registerer/Gatherer，应使用此方法而非默认 promhttp.Handler
func (p *PrometheusStr) Handler(opts ...promhttp.HandlerOpts) http.Handler {
	g := p.gatherer
	if g == nil {
		g = prometheus.DefaultGatherer
	}
	if len(opts) > 0 {
		return promhttp.HandlerFor(g, opts[0])
	}
	return promhttp.HandlerFor(g, promhttp.HandlerOpts{})
}

// StartServer 在指定地址启动 /metrics HTTP 端点，使用独立 Mux，不影响全局
func (p *PrometheusStr) StartServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", p.Handler())
	return http.ListenAndServe(addr, mux)
}

// MustStartServer 同 StartServer，失败时 panic
func (p *PrometheusStr) MustStartServer(addr string) {
	if err := p.StartServer(addr); err != nil {
		panic(err)
	}
}

// Deprecated: 使用 PrometheusStr.StartServer 替代，此函数使用全局 DefaultGatherer，自定义 Registry 的指标不可见
func InitPrometheusServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

// Deprecated: 使用 PrometheusStr.MustStartServer 替代
func MustInitPrometheusServer(addr string) {
	if err := InitPrometheusServer(addr); err != nil {
		panic(err)
	}
}

// counterOpts 构建 CounterOpts，自动填充 namespace/subsystem/constLabels
func (p *PrometheusStr) counterOpts(name, help string) prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        help,
		ConstLabels: p.constLabels,
	}
}

// gaugeOpts 构建 GaugeOpts
func (p *PrometheusStr) gaugeOpts(name, help string) prometheus.GaugeOpts {
	return prometheus.GaugeOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        help,
		ConstLabels: p.constLabels,
	}
}

// histogramOpts 构建 HistogramOpts
func (p *PrometheusStr) histogramOpts(name, help string, buckets []float64) prometheus.HistogramOpts {
	return prometheus.HistogramOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        help,
		Buckets:     buckets,
		ConstLabels: p.constLabels,
	}
}

// summaryOpts 构建 SummaryOpts
func (p *PrometheusStr) summaryOpts(name, help string, objectives map[float64]float64) prometheus.SummaryOpts {
	return prometheus.SummaryOpts{
		Namespace:   p.namespace,
		Subsystem:   p.subsystem,
		Name:        name,
		Help:        help,
		Objectives:  objectives,
		ConstLabels: p.constLabels,
	}
}
