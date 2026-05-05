package runner_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/cronwarden/internal/db"
	"github.com/yourorg/cronwarden/internal/runner"
	"github.com/yourorg/cronwarden/internal/webhook"
)

func tempDB(t *testing.T) *db.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close(); os.Remove(path) })
	return d
}

func TestRunner_Success(t *testing.T) {
	d := tempDB(t)
	r := &runner.Runner{
		DB: d,
		Cfg: runner.Config{
			Name:    "echo-job",
			Command: []string{"echo", "hello"},
			AlertOn: "never",
		},
	}
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	runs, err := d.ListJobRuns("echo-job", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Status != "success" {
		t.Errorf("status: got %q, want success", runs[0].Status)
	}
}

func TestRunner_Failure_TriggersWebhook(t *testing.T) {
	var received webhook.Payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received) //nolint:errcheck
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	d := tempDB(t)
	r := &runner.Runner{
		DB: d,
		Cfg: runner.Config{
			Name:       "fail-job",
			Command:    []string{"false"},
			WebhookURL: ts.URL,
			AlertOn:    "failure",
		},
	}
	_ = r.Run(context.Background()) // error expected

	if received.JobName != "fail-job" {
		t.Errorf("webhook job_name: got %q, want fail-job", received.JobName)
	}
	if received.Status != "failure" {
		t.Errorf("webhook status: got %q, want failure", received.Status)
	}
}
