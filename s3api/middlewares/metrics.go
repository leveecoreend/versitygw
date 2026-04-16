package middlewares

import (
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector holds aggregated request metrics for the gateway.
type MetricsCollector struct {
	mu             sync.RWMutex
	TotalRequests  atomic.Int64
	TotalErrors    atomic.Int64
	StatusCounts   map[int]int64
	MethodCounts   map[string]int64
	TotalLatencyMs atomic.Int64
}

// NewMetricsCollector creates and initialises a MetricsCollector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		StatusCounts: make(map[int]int64),
		MethodCounts: make(map[string]int64),
	}
}

// record updates the collector with the result of a single request.
func (m *MetricsCollector) record(method string, status int, latency time.Duration) {
	m.TotalRequests.Add(1)
	if status >= 400 {
		m.TotalErrors.Add(1)
	}
	m.TotalLatencyMs.Add(latency.Milliseconds())

	m.mu.Lock()
	m.StatusCounts[status]++
	m.MethodCounts[method]++
	m.mu.Unlock()
}

// Snapshot returns a point-in-time copy of the current metrics.
func (m *MetricsCollector) Snapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]int64, len(m.StatusCounts))
	for code, count := range m.StatusCounts {
		statuses[strconv.Itoa(code)] = count
	}

	methods := make(map[string]int64, len(m.MethodCounts))
	for method, count := range m.MethodCounts {
		methods[method] = count
	}

	total := m.TotalRequests.Load()
	var avgLatencyMs int64
	if total > 0 {
		avgLatencyMs = m.TotalLatencyMs.Load() / total
	}

	return map[string]interface{}{
		"total_requests":   total,
		"total_errors":     m.TotalErrors.Load(),
		"avg_latency_ms":   avgLatencyMs,
		"status_counts":    statuses,
		"method_counts":    methods,
	}
}

// MetricsMiddleware wraps an HTTP handler to collect per-request metrics.
// The provided MetricsCollector is updated after every request completes.
func MetricsMiddleware(collector *MetricsCollector) func(http.Handler) http.Handler {
	if collector == nil {
		collector = NewMetricsCollector()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the ResponseWriter so we can capture the status code.
			mrw := &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(mrw, r)

			collector.record(r.Method, mrw.status, time.Since(start))
		})
	}
}

// metricsResponseWriter is a minimal ResponseWriter wrapper that captures the
// HTTP status code written by the downstream handler.
type metricsResponseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (m *metricsResponseWriter) WriteHeader(code int) {
	if !m.written {
		m.status = code
		m.written = true
	}
	m.ResponseWriter.WriteHeader(code)
}

func (m *metricsResponseWriter) Write(b []byte) (int, error) {
	if !m.written {
		m.written = true
	}
	return m.ResponseWriter.Write(b)
}
