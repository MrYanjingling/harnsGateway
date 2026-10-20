package auditapi

import (
	"time"

	"harns-platform/pkg/metrics"
)

// PlatformMetrics adapts the audit recorder to the platform Prometheus metrics.
type PlatformMetrics struct{}

func (PlatformMetrics) Dropped() {
	metrics.AuditDroppedInc()
}

func (PlatformMetrics) BatchFlushed(result string, rows int, duration time.Duration) {
	metrics.AuditBatchFlushed(result)
	metrics.AuditBatchObserved(rows, duration)
}

func (PlatformMetrics) BatchFallback() {
	metrics.AuditBatchFallback()
}

func (PlatformMetrics) BatchFailed(rows int) {
	metrics.AuditBatchFailedRows(rows)
}
