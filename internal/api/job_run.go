package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/user/cronwarden/internal/db"
)

type jobRunHandler struct {
	db *sql.DB
}

func newJobRunHandler(database *sql.DB) *jobRunHandler {
	return &jobRunHandler{db: database}
}

// ServeHTTP handles GET /api/runs/{id} and DELETE /api/runs/{id}.
func (h *jobRunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/runs/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid run id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getRun(w, r, id)
	case http.MethodDelete:
		h.deleteRun(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *jobRunHandler) getRun(w http.ResponseWriter, _ *http.Request, id int64) {
	run, err := db.GetJobRun(h.db, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if run == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(run)
}

func (h *jobRunHandler) deleteRun(w http.ResponseWriter, _ *http.Request, id int64) {
	deleted, err := db.DeleteJobRun(h.db, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
