package audit

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"harns-platform/pkg/logging"
)

// Maintenance periodically removes expired audit logs. Only the platform
// service should start it.
type Maintenance struct {
	store *store
	cfg   Config
	done  chan struct{}
	once  sync.Once
	wg    sync.WaitGroup
}

// NewMaintenance starts the cleanup loop when retention is enabled.
func NewMaintenance(db *sql.DB, cfg Config) *Maintenance {
	cfg = normalizeConfig(cfg)
	m := &Maintenance{
		store: newStore(db),
		cfg:   cfg,
		done:  make(chan struct{}),
	}
	if cfg.RetentionDays > 0 {
		m.wg.Add(1)
		go m.loop()
	}
	return m
}

func (m *Maintenance) loop() {
	defer m.wg.Done()
	m.cleanup()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.done:
			return
		}
	}
}

func (m *Maintenance) cleanup() {
	if m.cfg.RetentionDays <= 0 {
		return
	}
	before := time.Now().AddDate(0, 0, -m.cfg.RetentionDays)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	n, err := m.store.DeleteBefore(ctx, before, m.cfg.CleanupBatchSize)
	if err != nil {
		logging.V(1).InfoS("Failed to clean audit logs", "err", err)
		return
	}
	if n > 0 {
		logging.V(1).InfoS("Cleaned audit logs", "count", n, "before", before)
	}
}

// Close stops the cleanup loop.
func (m *Maintenance) Close(ctx context.Context) {
	m.once.Do(func() { close(m.done) })
	waited := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(waited)
	}()
	select {
	case <-waited:
	case <-ctx.Done():
		logging.Warnf("Audit maintenance shutdown timed out")
	}
}
