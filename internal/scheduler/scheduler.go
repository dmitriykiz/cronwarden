package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"cronwarden/internal/config"
	"cronwarden/internal/runner"
)

// Scheduler wraps a cron scheduler and manages job lifecycles.
type Scheduler struct {
	c       *cron.Cron
	runner  *runner.Runner
	entries map[string]cron.EntryID
	mu      sync.Mutex
}

// New creates a new Scheduler using the provided Runner.
func New(r *runner.Runner) *Scheduler {
	return &Scheduler{
		c:       cron.New(cron.WithSeconds()),
		runner:  r,
		entries: make(map[string]cron.EntryID),
	}
}

// Register adds all jobs from the config to the cron scheduler.
func (s *Scheduler) Register(jobs []config.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, job := range jobs {
		j := job // capture loop variable
		id, err := s.c.AddFunc(j.Schedule, func() {
			if err := s.runner.Run(j); err != nil {
				log.Printf("[scheduler] job %q finished with error: %v", j.Name, err)
			} else {
				log.Printf("[scheduler] job %q completed successfully", j.Name)
			}
		})
		if err != nil {
			return fmt.Errorf("registering job %q: %w", j.Name, err)
		}
		s.entries[j.Name] = id
		log.Printf("[scheduler] registered job %q with schedule %q", j.Name, j.Schedule)
	}
	return nil
}

// Start begins the cron scheduler in the background.
func (s *Scheduler) Start() {
	s.c.Start()
	log.Println("[scheduler] started")
}

// Stop gracefully shuts down the scheduler, waiting for running jobs.
func (s *Scheduler) Stop() {
	ctx := s.c.Stop()
	<-ctx.Done()
	log.Println("[scheduler] stopped")
}

// NextRun returns the next scheduled time for the named job, or zero if not found.
func (s *Scheduler) NextRun(name string) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.entries[name]
	if !ok {
		return time.Time{}
	}
	return s.c.Entry(id).Next
}
