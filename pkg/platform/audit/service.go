package audit

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"harns-platform/pkg/logging"
)

// Service asynchronously persists audit logs.
type Service struct {
	store *Store
	cfg   Config
	ch    chan *Log
	done  chan struct{}
	once  sync.Once
	wg    sync.WaitGroup
}

func NewService(db *sql.DB, cfg Config) *Service {
	if cfg.MaxBodyBytes <= 0 {
		cfg.MaxBodyBytes = 16384
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}
	s := &Service{
		store: NewStore(db),
		cfg:   cfg,
		ch:    make(chan *Log, cfg.QueueSize),
		done:  make(chan struct{}),
	}
	s.wg.Add(1)
	go s.worker()
	return s
}

func (s *Service) worker() {
	defer s.wg.Done()
	for {
		select {
		case l := <-s.ch:
			s.write(l)
		case <-s.done:
			for {
				select {
				case l := <-s.ch:
					s.write(l)
				default:
					return
				}
			}
		}
	}
}

func (s *Service) write(l *Log) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.store.Insert(ctx, l); err != nil {
		logging.V(1).InfoS("Failed to write audit log", "err", err, "path", l.Path)
	}
}

func (s *Service) enqueue(l *Log) {
	select {
	case s.ch <- l:
	case <-s.done:
	default:
		logging.V(1).InfoS("Audit queue is full, dropping log", "path", l.Path)
	}
}

// Close stops the worker after draining pending logs.
func (s *Service) Close() {
	s.once.Do(func() { close(s.done) })
	s.wg.Wait()
}
