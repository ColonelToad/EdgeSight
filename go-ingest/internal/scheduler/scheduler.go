package scheduler

import (
	"time"
)

// Scheduler manages periodic data ingestion tasks
type Scheduler struct {
	interval time.Duration
	stopCh   chan struct{}
}

// NewScheduler creates a new scheduler with the specified interval
func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins the scheduler
func (s *Scheduler) Start() {
	// Scheduler logic will be implemented here
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopCh)
}
