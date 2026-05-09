package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedEventAPIRun(t *testing.T, d *db.DB) int64 {
	t.Helper()
	id, err := d.InsertJobRun("event-api-job", true, 0, time.Now())
	require.NoError(t, err)
	return id
}

func TestJobRunEvents_ListEmpty(t *testing.T) {
	d := tempDB(t)
	runID := seedEventAPIRun(t, d)

	h := newJobRunEventsHandler(d)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = withPathParam(req, "run_id", itoa(int(runID)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var out []interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Empty(t, out)
}

func TestJobRunEvents_PostAndList(t *testing.T) {
	d := tempDB(t)
	runID := seedEventAPIRun(t, d)
	h := newJobRunEventsHandler(d)

	body := `{"event_type":"started","message":"job kicked off"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req = withPathParam(req, "run_id", itoa(int(runID)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2 = withPathParam(req2, "run_id", itoa(int(runID)))
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	var events []map[string]interface{}
	require.NoError(t, json.NewDecoder(rec2.Body).Decode(&events))
	require.Len(t, events, 1)
	assert.Equal(t, "started", events[0]["event_type"])
	assert.Equal(t, "job kicked off", events[0]["message"])
}

func TestJobRunEvents_DeleteAll(t *testing.T) {
	d := tempDB(t)
	runID := seedEventAPIRun(t, d)
	h := newJobRunEventsHandler(d)

	_ = d.AddJobRunEvent(runID, "info", "step 1")
	_ = d.AddJobRunEvent(runID, "info", "step 2")

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req = withPathParam(req, "run_id", itoa(int(runID)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	events, err := d.ListJobRunEvents(runID)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestJobRunEvents_MethodNotAllowed(t *testing.T) {
	d := tempDB(t)
	h := newJobRunEventsHandler(d)

	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	req = withPathParam(req, "run_id", "1")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
