package retention

import (
	"fmt"
	"log"
	"time"

	"github.com/cronwarden/internal/db"
)

// Policy defines how long job run records are retained.
type Policy struct {
	MaxAgeDays int
	MaxRows    int
}

// Cleaner removes stale job run records according to a Policy.
type Cleaner struct {
	db     *db.DB
	policy Policy
	log    *log.Logger
}

// New creates a new Cleaner with the given database and policy.
func New(database *db.DB, policy Policy, logger *log.Logger) *Cleaner {
	if logger == nil {
		logger = log.Default()
	}
	return &Cleaner{db: database, policy: policy, log: logger}
}

// Run executes a single cleanup pass, returning the number of rows deleted.
func (c *Cleaner) Run() (int64, error) {
	var total int64

	if c.policy.MaxAgeDays > 0 {
		cutoff := time.Now().UTC().AddDate(0, 0, -c.policy.MaxAgeDays)
		n, err := c.db.DeleteRunsBefore(cutoff)
		if err != nil {
			return total, fmt.Errorf("retention: age purge: %w", err)
		}
		total += n
		c.log.Printf("retention: removed %d rows older than %d days", n, c.policy.MaxAgeDays)
	}

	if c.policy.MaxRows > 0 {
		n, err := c.db.TrimRunsToLimit(c.policy.MaxRows)
		if err != nil {
			return total, fmt.Errorf("retention: row-limit trim: %w", err)
		}
		total += n
		c.log.Printf("retention: trimmed %d rows to enforce max-rows=%d", n, c.policy.MaxRows)
	}

	return total, nil
}

// Start runs the cleaner on the given interval until the done channel is closed.
func (c *Cleaner) Start(interval time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if _, err := c.Run(); err != nil {
				c.log.Printf("retention: error during cleanup: %v", err)
			}
		case <-done:
			return
		}
	}
}
