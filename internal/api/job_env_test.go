package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwarden/cronwarden/internal/api"
	"github.com/cronwarden/cronwarden/internal/db"
)

func TestJobEnv_ListEmpty(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	req := httptest.NewRequest(http.MethodGet, "/jobs/myjob/env", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestJobEnv_PutAndList(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	body := bytes.NewBufferString(`{"value":"us-west-2"}`)
	req := httptest.NewRequest(http.MethodPut, "/jobs/myjob/env/AWS_REGION", body)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	vars, err := db.ListJobEnvVars(conn, "myjob")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if vars["AWS_REGION"] != "us-west-2" {
		t.Errorf("expected us-west-2, got %s", vars["AWS_REGION"])
	}
}

func TestJobEnv_Delete(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	db.UpsertJobEnvVar(conn, "myjob", "TOKEN", "secret")

	req := httptest.NewRequest(http.MethodDelete, "/jobs/myjob/env/TOKEN", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	vars, _ := db.ListJobEnvVars(conn, "myjob")
	if _, ok := vars["TOKEN"]; ok {
		t.Error("expected TOKEN to be deleted")
	}
}

func TestJobEnv_MethodNotAllowed(t *testing.T) {
	conn := tempDB(t)
	srv := api.New(conn, nil)

	req := httptest.NewRequest(http.MethodPost, "/jobs/myjob/env/KEY", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
