package audit

import "time"

// Metrics receives audit recorder observations. Implementations must be safe
// for concurrent use. The default implementation is a no-op.
type Metrics interface {
	Dropped()
	BatchFlushed(result string, rows int, duration time.Duration)
	BatchFallback()
	BatchFailed(rows int)
}

type noopMetrics struct{}

func (noopMetrics) Dropped()                                {}
func (noopMetrics) BatchFlushed(string, int, time.Duration) {}
func (noopMetrics) BatchFallback()                          {}
func (noopMetrics) BatchFailed(int)                         {}
