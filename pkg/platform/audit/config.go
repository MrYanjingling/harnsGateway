package audit

// Config configures the audit log middleware and writer.
type Config struct {
	Enabled      bool `json:"enabled"`
	MaxBodyBytes int  `json:"max-body-bytes"`
	QueueSize    int  `json:"queue-size"`
}

// DefaultConfig returns the default audit configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		MaxBodyBytes: 16384,
		QueueSize:    1024,
	}
}
