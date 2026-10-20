package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "harns",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "harns",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	auditDroppedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "harns",
		Subsystem: "audit",
		Name:      "dropped_total",
		Help:      "Total number of audit logs dropped because the queue was full.",
	})

	auditBatchFlushTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "harns",
		Subsystem: "audit",
		Name:      "batch_flush_total",
		Help:      "Total number of audit batch flush attempts by result.",
	}, []string{"result"})

	auditBatchRows = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "harns",
		Subsystem: "audit",
		Name:      "batch_rows",
		Help:      "Number of audit logs per batch flush.",
		Buckets:   []float64{1, 10, 25, 50, 100, 250, 500},
	})

	auditBatchDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "harns",
		Subsystem: "audit",
		Name:      "batch_duration_seconds",
		Help:      "Time spent flushing an audit batch.",
		Buckets:   prometheus.DefBuckets,
	})

	auditBatchFallbackTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "harns",
		Subsystem: "audit",
		Name:      "batch_fallback_total",
		Help:      "Total number of batches that fell back to per-row inserts.",
	})

	auditBatchFailedRowsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "harns",
		Subsystem: "audit",
		Name:      "batch_failed_rows_total",
		Help:      "Total number of audit rows that could not be persisted.",
	})
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		auditDroppedTotal,
		auditBatchFlushTotal,
		auditBatchRows,
		auditBatchDuration,
		auditBatchFallbackTotal,
		auditBatchFailedRowsTotal,
	)
}

// Middleware records request count and latency. It uses the matched route
// template as the path label to keep cardinality bounded.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
	}
}

// Handler exposes the Prometheus metrics endpoint.
func Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// AuditDroppedInc increments the audit log drop counter.
func AuditDroppedInc() {
	auditDroppedTotal.Inc()
}

// AuditBatchFlushed records a batch flush result.
func AuditBatchFlushed(result string) {
	auditBatchFlushTotal.WithLabelValues(result).Inc()
}

// AuditBatchObserved records the batch size and duration.
func AuditBatchObserved(rows int, d time.Duration) {
	auditBatchRows.Observe(float64(rows))
	auditBatchDuration.Observe(d.Seconds())
}

// AuditBatchFallback increments the per-row fallback counter.
func AuditBatchFallback() {
	auditBatchFallbackTotal.Inc()
}

// AuditBatchFailedRows adds the number of rows that could not be persisted.
func AuditBatchFailedRows(n int) {
	if n > 0 {
		auditBatchFailedRowsTotal.Add(float64(n))
	}
}
