package notifier

import (
	"fmt"
	"log"
	"time"

	"github.com/cronwarden/internal/webhook"
)

// Event represents a job execution event to be notified.
type Event struct {
	JobName   string
	Command   string
	ExitCode  int
	Duration  time.Duration
	Error     string
	Timestamp time.Time
}

// Notifier dispatches webhook alerts for job events.
type Notifier struct {
	client    *webhook.Client
	onFailure bool
	onSuccess bool
}

// Config holds notifier configuration.
type Config struct {
	WebhookURL string
	OnFailure  bool
	OnSuccess  bool
}

// New creates a new Notifier. Returns nil if no webhook URL is configured.
func New(cfg Config) *Notifier {
	if cfg.WebhookURL == "" {
		return nil
	}
	return &Notifier{
		client:    webhook.New(cfg.WebhookURL),
		onFailure: cfg.OnFailure,
		onSuccess: cfg.OnSuccess,
	}
}

// Notify sends a webhook alert for the given event if conditions are met.
func (n *Notifier) Notify(event Event) {
	if n == nil {
		return
	}

	isFailure := event.ExitCode != 0 || event.Error != ""

	if isFailure && !n.onFailure {
		return
	}
	if !isFailure && !n.onSuccess {
		return
	}

	status := "success"
	if isFailure {
		status = "failure"
	}

	msg := fmt.Sprintf(
		"[cronwarden] job=%q status=%s exit_code=%d duration=%s time=%s",
		event.JobName,
		status,
		event.ExitCode,
		event.Duration.Round(time.Millisecond),
		event.Timestamp.Format(time.RFC3339),
	)
	if event.Error != "" {
		msg += fmt.Sprintf(" error=%q", event.Error)
	}

	if err := n.client.Send(msg); err != nil {
		log.Printf("notifier: failed to send webhook for job %q: %v", event.JobName, err)
	}
}
