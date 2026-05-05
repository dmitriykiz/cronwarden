package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cronwarden/internal/db"
)

// Handler holds dependencies for the HTTP API.
type Handler struct {
	db *db.DB
}

// New creates a new Handler.
func New(database *db.DB) *Handler {
	return &Handler{db: database}
}

// RegisterRoutes attaches API routes to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/runs", h.listRuns)
	mux.HandleFunc("/api/health", h.health)
}

// listRuns returns recent job run records as JSON.
// Accepts optional query params: job=<name>, limit=<n> (default 50).
func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	jobName := q.Get("job")

	limit := 50
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	runs, err := h.db.ListJobRuns(jobName, limit)
	if err != nil {
		http.Error(w, "failed to query runs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(runs); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// health returns a simple 200 OK response.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
