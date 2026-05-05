package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwarden/internal/notifier"
)

func newTestServer(t *testing.T, statusCode int, calls *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(statusCode)
	}))
}

func TestNotify_NilNotifier_NoPanic(t *testing.T) {
	var n *notifier.Notifier
	// Should not panic
	n.Notify(notifier.Event{JobName: "test", ExitCode: 1})
}

func TestNotify_NoWebhookURL_ReturnsNil(t *testing.T) {
	n := notifier.New(notifier.Config{OnFailure: true})
	if n != nil {
		t.Fatal("expected nil notifier when no webhook URL provided")
	}
}

func TestNotify_OnFailure_SendsWebhook(t *testing.T) {
	var calls atomic.Int32
	ts := newTestServer(t, http.StatusOK, &calls)
	defer ts.Close()

	n := notifier.New(notifier.Config{
		WebhookURL: ts.URL,
		OnFailure:  true,
		OnSuccess:  false,
	})

	n.Notify(notifier.Event{
		JobName:   "backup",
		ExitCode:  1,
		Duration:  2 * time.Second,
		Timestamp: time.Now(),
		Error:     "exit status 1",
	})

	if calls.Load() != 1 {
		t.Fatalf("expected 1 webhook call, got %d", calls.Load())
	}
}

func TestNotify_OnSuccess_NotSentForFailure(t *testing.T) {
	var calls atomic.Int32
	ts := newTestServer(t, http.StatusOK, &calls)
	defer ts.Close()

	n := notifier.New(notifier.Config{
		WebhookURL: ts.URL,
		OnFailure:  false,
		OnSuccess:  true,
	})

	n.Notify(notifier.Event{
		JobName:   "backup",
		ExitCode:  2,
		Duration:  500 * time.Millisecond,
		Timestamp: time.Now(),
	})

	if calls.Load() != 0 {
		t.Fatalf("expected 0 webhook calls for failure when OnFailure=false, got %d", calls.Load())
	}
}

func TestNotify_OnSuccess_SendsWebhook(t *testing.T) {
	var calls atomic.Int32
	var lastBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		lastBody = payload["text"]
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notifier.New(notifier.Config{
		WebhookURL: ts.URL,
		OnFailure:  false,
		OnSuccess:  true,
	})

	n.Notify(notifier.Event{
		JobName:   "healthcheck",
		ExitCode:  0,
		Duration:  100 * time.Millisecond,
		Timestamp: time.Now(),
	})

	if calls.Load() != 1 {
		t.Fatalf("expected 1 webhook call, got %d", calls.Load())
	}
	_ = lastBody // payload content validated by webhook package tests
}
