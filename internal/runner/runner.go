package runner

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/yourorg/cronwarden/internal/db"
	"github.com/yourorg/cronwarden/internal/webhook"
)

// Config holds the configuration for a single monitored cron job.
type Config struct {
	Name       string
	Command    []string
	WebhookURL string // empty means no webhook
	AlertOn    string // "failure", "always", or "never"
}

// Runner executes a job, persists the result, and optionally fires a webhook.
type Runner struct {
	DB  *db.DB
	Cfg Config
}

// Run executes the configured command, records the outcome in SQLite, and
// triggers a webhook notification according to AlertOn policy.
func (r *Runner) Run(ctx context.Context) error {
	start := time.Now()

	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, r.Cfg.Command[0], r.Cfg.Command[1:]...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	runErr := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	status := "success"
	if runErr != nil {
		status = "failure"
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	run := db.JobRun{
		JobName:   r.Cfg.Name,
		StartedAt: start,
		Duration:  duration.Seconds(),
		ExitCode:  exitCode,
		Status:    status,
		Output:    buf.String(),
	}
	if err := r.DB.InsertJobRun(run); err != nil {
		return err
	}

	if r.shouldAlert(status) && r.Cfg.WebhookURL != "" {
		n := webhook.New(r.Cfg.WebhookURL)
		_ = n.Send(webhook.Payload{
			JobName:   r.Cfg.Name,
			Status:    status,
			ExitCode:  exitCode,
			Duration:  duration.Seconds(),
			Output:    buf.String(),
			Timestamp: start,
		})
	}

	return runErr
}

func (r *Runner) shouldAlert(status string) bool {
	switch r.Cfg.AlertOn {
	case "always":
		return true
	case "failure":
		return status == "failure"
	default:
		return false
	}
}
