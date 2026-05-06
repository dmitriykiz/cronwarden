package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cronwarden/internal/db"
)

// handleNextRun returns the next scheduled run time for a named job.
func (s *Server) handleNextRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("job")
	if name == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}

	next, err := s.scheduler.NextRun(name)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job":      name,
		"next_run": next.UTC().Format("2006-01-02T15:04:05Z"),
	})
}

// handleJobStats returns aggregate statistics for a named job.
func (s *Server) handleJobStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("job")
	if name == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	runs, err := db.ListJobRunsByName(s.db, name, limit)
	if err != nil {
		http.Error(w, "failed to query runs", http.StatusInternalServerError)
		return
	}

	stats := computeStats(name, runs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
