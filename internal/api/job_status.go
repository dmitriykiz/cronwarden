package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/example/cronwarden/internal/db"
)

// handleJobStatuses returns a JSON array of the latest run status for each
// known job. GET /api/jobs/status
func (s *Server) handleJobStatuses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statuses, err := db.ListJobStatuses(s.db)
	if err != nil {
		log.Printf("api: list job statuses: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if statuses == nil {
		statuses = []db.JobStatus{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statuses); err != nil {
		log.Printf("api: encode job statuses: %v", err)
	}
}
