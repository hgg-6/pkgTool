package prometheusX

import (
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRegistry 创建隔离的自定义 Registry
func testRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

// newTestPrometheusStr 创建带自定义 Registry 的 PrometheusStr
func newTestPrometheusStr(t *testing.T, opts ...Option) *PrometheusStr {
	t.Helper()
	reg := testRegistry()
	defaultOpts := []Option{
		WithNamespace("hgg"),
		WithSubsystem("test"),
		WithRegisterer(reg),
		WithGatherer(reg),
	}
	allOpts := append(defaultOpts, opts...)
	return New(allOpts...)
}

// gatherMetricNames 从 Gatherer 中收集所有已注册的 metric 名称
func gatherMetricNames(t *testing.T, g prometheus.Gatherer) []string {
	t.Helper()
	mfs, err := g.Gather()
	require.NoError(t, err)
	names := make([]string, 0, len(mfs))
	for _, mf := range mfs {
		names = append(names, mf.GetName())
	}
	return names
}

// ============================== New / Options ==============================

func TestNew_Defaults(t *testing.T) {
	p := New()
	assert.NotNil(t, p)
	assert.Equal(t, "", p.namespace)
	assert.Equal(t, "", p.subsystem)
	assert.Nil(t, p.constLabels)
	assert.Nil(t, p.reg)
	assert.Nil(t, p.gatherer)
}

func TestNew_WithNamespace(t *testing.T) {
	p := New(WithNamespace("myapp"))
	assert.Equal(t, "myapp", p.namespace)
}

func TestNew_WithSubsystem(t *testing.T) {
	p := New(WithSubsystem("mysub"))
	assert.Equal(t, "mysub", p.subsystem)
}

func TestNew_WithConstLabels(t *testing.T) {
	labels := map[string]string{"env": "test", "region": "cn"}
	p := New(WithConstLabels(labels))
	assert.Len(t, p.constLabels, 2)
	assert.Equal(t, "test", p.constLabels["env"])
	assert.Equal(t, "cn", p.constLabels["region"])
}

func TestNew_WithRegisterer(t *testing.T) {
	reg := prometheus.NewRegistry()
	p := New(WithRegisterer(reg))
	assert.Equal(t, reg, p.reg)
}

func TestNew_WithGatherer(t *testing.T) {
	reg := prometheus.NewRegistry()
	p := New(WithGatherer(reg))
	assert.Equal(t, reg, p.gatherer)
}

func TestNew_AllOptions(t *testing.T) {
	reg := prometheus.NewRegistry()
	p := New(
		WithNamespace("app"),
		WithSubsystem("core"),
		WithConstLabels(map[string]string{"env": "prod"}),
		WithRegisterer(reg),
		WithGatherer(reg),
	)
	assert.Equal(t, "app", p.namespace)
	assert.Equal(t, "core", p.subsystem)
	assert.Len(t, p.constLabels, 1)
	assert.Equal(t, "prod", p.constLabels["env"])
	assert.Equal(t, reg, p.reg)
	assert.Equal(t, reg, p.gatherer)
}

// ============================== Register ==============================

func TestRegister_Success(t *testing.T) {
	p := newTestPrometheusStr(t)
	c := prometheus.NewCounter(prometheus.CounterOpts{Name: "test_counter"})
	err := p.Register(c)
	assert.NoError(t, err)

	names := gatherMetricNames(t, p.gatherer)
	assert.Contains(t, names, "test_counter")
}

func TestRegister_DuplicateIgnored(t *testing.T) {
	p := newTestPrometheusStr(t)

	c1 := prometheus.NewCounter(prometheus.CounterOpts{Name: "dup_counter"})
	err := p.Register(c1)
	require.NoError(t, err)

	c2 := prometheus.NewCounter(prometheus.CounterOpts{Name: "dup_counter"})
	err = p.Register(c2)
	assert.ErrorIs(t, err, ErrAlreadyRegistered)

	names := gatherMetricNames(t, p.gatherer)
	assert.Contains(t, names, "dup_counter")
}

func TestRegister_ReturnsAlreadyRegisteredError(t *testing.T) {
	p := newTestPrometheusStr(t)

	c1 := prometheus.NewCounter(prometheus.CounterOpts{Name: "already_err_test"})
	err := p.Register(c1)
	require.NoError(t, err)

	c2 := prometheus.NewCounter(prometheus.CounterOpts{Name: "already_err_test"})
	err = p.Register(c2)
	assert.True(t, errors.Is(err, ErrAlreadyRegistered))
}

func TestMustRegister_Success(t *testing.T) {
	p := newTestPrometheusStr(t)
	c := prometheus.NewCounter(prometheus.CounterOpts{Name: "must_counter"})

	assert.NotPanics(t, func() {
		p.MustRegister(c)
	})

	names := gatherMetricNames(t, p.gatherer)
	assert.Contains(t, names, "must_counter")
}

func TestMustRegister_DuplicateIgnored(t *testing.T) {
	p := newTestPrometheusStr(t)
	c := prometheus.NewCounter(prometheus.CounterOpts{Name: "must_dup_counter"})
	p.MustRegister(c)

	c2 := prometheus.NewCounter(prometheus.CounterOpts{Name: "must_dup_counter"})
	assert.NotPanics(t, func() {
		p.MustRegister(c2)
	})
}

func TestErrAlreadyRegistered_SentinelType(t *testing.T) {
	p := newTestPrometheusStr(t)
	c := prometheus.NewCounter(prometheus.CounterOpts{Name: "sentinel_test"})
	require.NoError(t, p.Register(c))

	c2 := prometheus.NewCounter(prometheus.CounterOpts{Name: "sentinel_test"})
	err := p.Register(c2)
	assert.True(t, errors.Is(err, ErrAlreadyRegistered))
	assert.Equal(t, "prometheus: collector already registered", err.Error())
}

// ============================== Handler ==============================

func TestHandler_DefaultGatherer(t *testing.T) {
	p := New()
	h := p.Handler()
	assert.NotNil(t, h)
	assert.Implements(t, (*http.Handler)(nil), h)
}

func TestHandler_WithOpts(t *testing.T) {
	reg := testRegistry()
	p := New(WithGatherer(reg))
	h := p.Handler(promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError})
	assert.NotNil(t, h)
}

func TestHandler_CustomGatherer(t *testing.T) {
	reg := testRegistry()
	p := New(WithGatherer(reg))
	h := p.Handler()
	assert.NotNil(t, h)

	c := prometheus.NewCounter(prometheus.CounterOpts{Name: "handler_counter"})
	require.NoError(t, reg.Register(c))

	names := gatherMetricNames(t, reg)
	assert.Contains(t, names, "handler_counter")
}

func TestStartServer_UsesCustomGatherer(t *testing.T) {
	reg := testRegistry()
	p := New(WithRegisterer(reg), WithGatherer(reg), WithNamespace("startsrv"))
	p.NewCounter("test_counter", "test").Inc()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	ln.Close()

	go p.StartServer(addr)

	time.Sleep(50 * time.Millisecond)
	resp, err := http.Get("http://" + addr + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "startsrv_test_counter")
}

// ============================== NewCounter ==============================

func TestNewCounter_CreationAndRegistration(t *testing.T) {
	p := newTestPrometheusStr(t)
	c := p.NewCounter("requests_total", "Total requests")
	assert.NotNil(t, c)

	names := gatherMetricNames(t, p.gatherer)
	assert.Contains(t, names, "hgg_test_requests_total")
}

func TestNewCounter_Observation(t *testing.T) {
	p := newTestPrometheusStr(t)
	c := p.NewCounter("ops_total", "Total ops")
	c.Inc()
	c.Add(10)

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	var found bool
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_ops_total" {
			assert.Equal(t, 11.0, mf.GetMetric()[0].GetCounter().GetValue())
			found = true
		}
	}
	assert.True(t, found, "metric hgg_test_ops_total not found")
}

// ============================== NewCounterVec ==============================

func TestNewCounterVec_Creation(t *testing.T) {
	p := newTestPrometheusStr(t)
	cv := p.NewCounterVec("http_requests_total", "HTTP requests",
		[]string{"method", "status"})
	assert.NotNil(t, cv)
}

func TestNewCounterVec_WithLabelValues(t *testing.T) {
	p := newTestPrometheusStr(t)
	cv := p.NewCounterVec("vec_counter_total", "Vec counter",
		[]string{"label1"})
	cv.WithLabelValues("v1").Inc()
	cv.WithLabelValues("v2").Add(5)

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_vec_counter_total" {
			assert.Len(t, mf.GetMetric(), 2)
			return
		}
	}
	t.Fatal("metric hgg_test_vec_counter_total not found")
}

// ============================== NewGauge ==============================

func TestNewGauge_CreationAndRegistration(t *testing.T) {
	p := newTestPrometheusStr(t)
	g := p.NewGauge("memory_bytes", "Memory usage")
	assert.NotNil(t, g)

	names := gatherMetricNames(t, p.gatherer)
	assert.Contains(t, names, "hgg_test_memory_bytes")
}

func TestNewGauge_SetAndAdd(t *testing.T) {
	p := newTestPrometheusStr(t)
	g := p.NewGauge("queue_size", "Queue size")
	g.Set(100)
	g.Add(20)

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_queue_size" {
			assert.Equal(t, 120.0, mf.GetMetric()[0].GetGauge().GetValue())
		}
	}
}

// ============================== NewGaugeVec ==============================

func TestNewGaugeVec_Creation(t *testing.T) {
	p := newTestPrometheusStr(t)
	gv := p.NewGaugeVec("cpu_usage", "CPU usage", []string{"core"})
	assert.NotNil(t, gv)
}

func TestNewGaugeVec_WithLabelValues(t *testing.T) {
	p := newTestPrometheusStr(t)
	gv := p.NewGaugeVec("temp_celsius", "Temperature", []string{"location"})
	gv.WithLabelValues("room_a").Set(25.5)
	gv.WithLabelValues("room_b").Set(18.0)

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_temp_celsius" {
			assert.Len(t, mf.GetMetric(), 2)
		}
	}
}

// ============================== NewHistogram ==============================

func TestNewHistogram_CreationAndRegistration(t *testing.T) {
	p := newTestPrometheusStr(t)
	h := p.NewHistogram("latency_ms", "Request latency",
		[]float64{10, 50, 100, 500})
	assert.NotNil(t, h)

	names := gatherMetricNames(t, p.gatherer)
	assert.Contains(t, names, "hgg_test_latency_ms")
}

func TestNewHistogram_Observation(t *testing.T) {
	p := newTestPrometheusStr(t)
	h := p.NewHistogram("rpc_latency_ms", "RPC latency",
		[]float64{10, 50, 100})
	h.Observe(25)
	h.Observe(75)
	h.Observe(25)

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_rpc_latency_ms" {
			hist := mf.GetMetric()[0].GetHistogram()
			assert.Equal(t, uint64(3), hist.GetSampleCount())
			assert.Equal(t, 125.0, hist.GetSampleSum())
		}
	}
}

// ============================== NewHistogramVec ==============================

func TestNewHistogramVec_Creation(t *testing.T) {
	p := newTestPrometheusStr(t)
	hv := p.NewHistogramVec("endpoint_latency", "Endpoint latency",
		[]string{"endpoint"}, []float64{10, 50, 100})
	assert.NotNil(t, hv)
}

func TestNewHistogramVec_WithLabelValues(t *testing.T) {
	p := newTestPrometheusStr(t)
	hv := p.NewHistogramVec("api_latency", "API latency",
		[]string{"method"}, []float64{10, 100})
	hv.WithLabelValues("GET").Observe(30)
	hv.WithLabelValues("POST").Observe(80)

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_api_latency" {
			assert.Len(t, mf.GetMetric(), 2)
			return
		}
	}
	t.Fatal("metric hgg_test_api_latency not found")
}

// ============================== NewSummary ==============================

func TestNewSummary_CreationAndRegistration(t *testing.T) {
	p := newTestPrometheusStr(t)
	s := p.NewSummary("response_size", "Response size",
		map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001})
	assert.NotNil(t, s)

	names := gatherMetricNames(t, p.gatherer)
	assert.Contains(t, names, "hgg_test_response_size")
}

func TestNewSummary_Observation(t *testing.T) {
	p := newTestPrometheusStr(t)
	s := p.NewSummary("task_latency", "Task latency",
		map[float64]float64{0.5: 0.05, 0.9: 0.01})

	for i := 0; i < 100; i++ {
		s.Observe(float64(i))
	}

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_task_latency" {
			assert.Equal(t, uint64(100), mf.GetMetric()[0].GetSummary().GetSampleCount())
		}
	}
}

// ============================== NewSummaryVec ==============================

func TestNewSummaryVec_Creation(t *testing.T) {
	p := newTestPrometheusStr(t)
	sv := p.NewSummaryVec("rpc_size", "RPC size",
		[]string{"service"},
		map[float64]float64{0.5: 0.05, 0.9: 0.01})
	assert.NotNil(t, sv)
}

func TestNewSummaryVec_WithLabelValues(t *testing.T) {
	p := newTestPrometheusStr(t)
	sv := p.NewSummaryVec("job_latency", "Job latency",
		[]string{"job_type"},
		map[float64]float64{0.5: 0.05, 0.9: 0.01})
	sv.WithLabelValues("ingest").Observe(42)
	sv.WithLabelValues("process").Observe(88)

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_job_latency" {
			assert.Len(t, mf.GetMetric(), 2)
		}
	}
}

// ============================== ConstLabels ==============================

func TestConstLabels_AppliedToMetrics(t *testing.T) {
	p := newTestPrometheusStr(t, WithConstLabels(map[string]string{"env": "test"}))
	c := p.NewCounter("labeled_counter", "Counter with labels")
	c.Inc()

	mfs, err := p.gatherer.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == "hgg_test_labeled_counter" {
			labels := mf.GetMetric()[0].GetLabel()
			hasEnv := false
			for _, l := range labels {
				if l.GetName() == "env" && l.GetValue() == "test" {
					hasEnv = true
				}
			}
			assert.True(t, hasEnv, "const label 'env=test' not found")
		}
	}
}

// ============================== Custom Registry Isolation ==============================

func TestCustomRegistry_Isolation(t *testing.T) {
	reg1 := testRegistry()
	reg2 := testRegistry()

	p1 := New(WithRegisterer(reg1), WithGatherer(reg1), WithNamespace("app1"))
	p2 := New(WithRegisterer(reg2), WithGatherer(reg2), WithNamespace("app2"))

	p1.NewCounter("counter", "Counter 1").Inc()
	p2.NewCounter("counter", "Counter 2").Inc()

	names1 := gatherMetricNames(t, reg1)
	assert.Contains(t, names1, "app1_counter")
	assert.NotContains(t, names1, "app2_counter")

	names2 := gatherMetricNames(t, reg2)
	assert.Contains(t, names2, "app2_counter")
	assert.NotContains(t, names2, "app1_counter")
}

// ============================== MustInitPrometheusServer ==============================

func TestMustStartServer_PanicOnInvalidAddr(t *testing.T) {
	p := newTestPrometheusStr(t)
	assert.Panics(t, func() {
		p.MustStartServer(":invalid_port_format")
	})
}
