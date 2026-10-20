package audit

const (
	// DefaultBatchSize is the number of logs written per INSERT batch.
	DefaultBatchSize = 100
	// MaxBatchSize caps the batch size to stay well below MySQL limits.
	MaxBatchSize = 500
	// DefaultFlushIntervalMs is the maximum delay before a partial batch is written.
	DefaultFlushIntervalMs = 1000
)

// Config configures the audit recorder and maintenance.
type Config struct {
	Enabled          bool `json:"enabled"`
	MaxBodyBytes     int  `json:"max-body-bytes"`
	QueueSize        int  `json:"queue-size"`
	RetentionDays    int  `json:"retention-days"`
	CleanupBatchSize int  `json:"cleanup-batch-size"`
	BatchSize        int  `json:"batch-size"`
	FlushIntervalMs  int  `json:"flush-interval-ms"`
}

// DefaultConfig returns the default audit configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:          true,
		MaxBodyBytes:     16384,
		QueueSize:        1024,
		RetentionDays:    180,
		CleanupBatchSize: 1000,
		BatchSize:        DefaultBatchSize,
		FlushIntervalMs:  DefaultFlushIntervalMs,
	}
}
