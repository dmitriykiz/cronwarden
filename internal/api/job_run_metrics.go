package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/cronwarden/cronwarden/internal/db"
)

// newJobRunMetricsHandler returns an http.Handler for /runs/{id}/metrics and
// /runs/{id}/metrics/{metricID}.
func newJobRunMetricsHandler(database *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// GET /runs/{id}/metrics      — list metrics for a run
	// POST /runs/{id}/metrics     — add a metric
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		runID, ok := parseID(w, r, "run")
		if !ok {
			return
		}

		switch r.Method {
		case http.MethodGet:
			metrics, err := db.ListJobRunMetrics(database, runID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if metrics == nil {
				metrics = []db.JobRunMetric{}
			}
			writeJSON(w, http.StatusOK, metrics)

		case http.MethodPost:
			var body struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
				writeError(w, http.StatusBadRequest, "key and value are required")
				return
			}
			id, err := db.AddJobRunMetric(database, runID, body.Key, body.Value)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, map[string]int64{"id": id})

		default:
			allowMethods(w, http.MethodGet, http.MethodPost)
		}
	})

	// DELETE /runs/{id}/metrics/{metricID}
	mux.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if !allowMethods(w, http.MethodDelete) {
			return
		}
		metricID, ok := parseID(w, r, "metric")
		if !ok {
			return
		}
		deleted, err := db.DeleteJobRunMetric(database, metricID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !deleted {
			writeError(w, http.StatusNotFound, "metric not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
