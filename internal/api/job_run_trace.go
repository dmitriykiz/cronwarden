package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/cronwarden/cronwarden/internal/db"
)

type traceRequest struct {
	TraceID string `json:"trace_id"`
	SpanID  string `json:"span_id"`
}

func newJobRunTraceHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID, ok := parseID(w, r, "runID")
		if !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetTrace(w, r, database, runID)
		case http.MethodPut:
			handlePutTrace(w, r, database, runID)
		case http.MethodDelete:
			handleDeleteTrace(w, r, database, runID)
		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	}
}

func handleGetTrace(w http.ResponseWriter, r *http.Request, database *sql.DB, runID int64) {
	trace, err := db.GetJobRunTrace(r.Context(), database, runID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if trace == nil {
		writeError(w, http.StatusNotFound, "trace not found")
		return
	}
	writeJSON(w, http.StatusOK, trace)
}

func handlePutTrace(w http.ResponseWriter, r *http.Request, database *sql.DB, runID int64) {
	var req traceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.TraceID == "" || req.SpanID == "" {
		writeError(w, http.StatusBadRequest, "trace_id and span_id are required")
		return
	}
	if err := db.UpsertJobRunTrace(r.Context(), database, runID, req.TraceID, req.SpanID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDeleteTrace(w http.ResponseWriter, r *http.Request, database *sql.DB, runID int64) {
	if err := db.DeleteJobRunTrace(r.Context(), database, runID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
