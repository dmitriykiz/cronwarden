package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

type statsResponse struct {
	TotalRuns   int        `json:"total_runs"`
	SuccessRuns int        `json:"success_runs"`
	FailureRuns int        `json:"failure_runs"`
	UniqueJobs  int        `json:"unique_jobs"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := db.ComputeGlobalStats(s.db)
	if err != nil {
		log.Printf("stats query error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := statsResponse{
		TotalRuns:   stats.TotalRuns,
		SuccessRuns: stats.SuccessRuns,
		FailureRuns: stats.FailureRuns,
		UniqueJobs:  stats.UniqueJobs,
		LastRunAt:   stats.LastRunAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("stats encode error: %v", err)
	}
}
