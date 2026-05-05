package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwarden/internal/webhook"
)

func TestSend_Success(t *testing.T) {
	var received webhook.Payload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := webhook.New(ts.URL)
	p := webhook.Payload{
		JobName:   "backup",
		Status:    "failure",
		ExitCode:  1,
		Duration:  3.14,
		Output:    "disk full",
		Timestamp: time.Now().UTC(),
	}

	if err := n.Send(p); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if received.JobName != p.JobName {
		t.Errorf("job_name: got %q, want %q", received.JobName, p.JobName)
	}
	if received.ExitCode != p.ExitCode {
		t.Errorf("exit_code: got %d, want %d", received.ExitCode, p.ExitCode)
	}
}

func TestSend_Non2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := webhook.New(ts.URL)
	err := n.Send(webhook.Payload{JobName: "test", Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestSend_Unreachable(t *testing.T) {
	n := webhook.New("http://127.0.0.1:1") // nothing listening
	err := n.Send(webhook.Payload{JobName: "test", Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}
