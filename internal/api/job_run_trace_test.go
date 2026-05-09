package api_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwarden/cronwarden/internal/api"
	"github.com/cronwarden/cronwarden/internal/db"
)

func seedTraceAPIRun(t *testing.T, database *db.DB) int64 {
	t.Helper()
	_ = db.InsertJobRun(context.Background(), database.DB, db.JobRun{
		JobName: "trace-api-job", Status: "success", Duration: 2.0,
	})
	runs, _ := db.ListJobRuns(context.Background(), database.DB, 1)
	return runs[0].ID
}

func TestJobRunTrace_PutAndGet(t *testing.T) {
	d := tempDB(t)
	runID := seedTraceAPIRun(t, d)
	srv := httptest.NewServer(api.New(d.DB, nil))
	defer srv.Close()

	url := srv.URL + "/api/runs/" + itoa(int(runID)) + "/trace"

	body := `{"trace_id":"trace-xyz","span_id":"span-999"}`
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(body))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("put status: got %d", resp.StatusCode)
	}

	resp2, err := http.Get(url)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("get status: got %d", resp2.StatusCode)
	}
}

func TestJobRunTrace_GetNotFound(t *testing.T) {
	d := tempDB(t)
	srv := httptest.NewServer(api.New(d.DB, nil))
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/runs/9999/trace")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestJobRunTrace_Delete(t *testing.T) {
	d := tempDB(t)
	runID := seedTraceAPIRun(t, d)
	srv := httptest.NewServer(api.New(d.DB, nil))
	defer srv.Close()

	url := srv.URL + "/api/runs/" + itoa(int(runID)) + "/trace"
	_ = db.UpsertJobRunTrace(context.Background(), d.DB, runID, "t", "s")

	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("delete status: got %d", resp.StatusCode)
	}
}

func TestJobRunTrace_MethodNotAllowed(t *testing.T) {
	d := tempDB(t)
	runID := seedTraceAPIRun(t, d)
	srv := httptest.NewServer(api.New(d.DB, nil))
	defer srv.Close()

	url := srv.URL + "/api/runs/" + itoa(int(runID)) + "/trace"
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}
