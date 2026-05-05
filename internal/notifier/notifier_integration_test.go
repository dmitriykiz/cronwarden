package notifier_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwarden/internal/notifier"
)

// TestNotify_BothFlags_AllEventsDispatched verifies that when both OnFailure
// and OnSuccess are enabled, every event triggers a webhook call.
func TestNotify_BothFlags_AllEventsDispatched(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notifier.New(notifier.Config{
		WebhookURL: ts.URL,
		OnFailure:  true,
		OnSuccess:  true,
	})

	events := []notifier.Event{
		{JobName: "job-a", ExitCode: 0, Duration: 50 * time.Millisecond, Timestamp: time.Now()},
		{JobName: "job-b", ExitCode: 1, Duration: 200 * time.Millisecond, Timestamp: time.Now(), Error: "exit status 1"},
		{JobName: "job-c", ExitCode: 0, Duration: 10 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, e := range events {
		n.Notify(e)
	}

	if got := calls.Load(); got != int32(len(events)) {
		t.Fatalf("expected %d webhook calls, got %d", len(events), got)
	}
}

// TestNotify_NoFlags_NeverDispatched verifies that when both flags are false,
// no webhook calls are made regardless of job outcome.
func TestNotify_NoFlags_NeverDispatched(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notifier.New(notifier.Config{
		WebhookURL: ts.URL,
		OnFailure:  false,
		OnSuccess:  false,
	})

	n.Notify(notifier.Event{JobName: "job-a", ExitCode: 0, Timestamp: time.Now()})
	n.Notify(notifier.Event{JobName: "job-b", ExitCode: 1, Timestamp: time.Now()})

	if got := calls.Load(); got != 0 {
		t.Fatalf("expected 0 webhook calls, got %d", got)
	}
}
