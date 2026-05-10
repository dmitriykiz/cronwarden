package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

type acquireLockRequest struct {
	LockedBy string `json:"locked_by"`
	TTLSecs  int    `json:"ttl_seconds"`
}

func newJobRunLockHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobName := r.PathValue("job")
		if jobName == "" {
			writeError(w, http.StatusBadRequest, "missing job name")
			return
		}

		switch r.Method {
		case http.MethodGet:
			lock, err := db.GetJobRunLock(database, jobName)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if lock == nil {
				writeError(w, http.StatusNotFound, "no lock found")
				return
			}
			writeJSON(w, http.StatusOK, lock)

		case http.MethodPost:
			var req acquireLockRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if req.LockedBy == "" {
				writeError(w, http.StatusBadRequest, "locked_by is required")
				return
			}
			ttl := time.Duration(req.TTLSecs) * time.Second
			if ttl <= 0 {
				ttl = 60 * time.Second
			}
			ok, err := db.AcquireJobRunLock(database, jobName, req.LockedBy, ttl)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if !ok {
				writeError(w, http.StatusConflict, "lock already held")
				return
			}
			lock, _ := db.GetJobRunLock(database, jobName)
			writeJSON(w, http.StatusCreated, lock)

		case http.MethodDelete:
			owner := r.URL.Query().Get("locked_by")
			if owner == "" {
				writeError(w, http.StatusBadRequest, "locked_by query param required")
				return
			}
			if err := db.ReleaseJobRunLock(database, jobName, owner); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			allowMethods(w, http.MethodGet, http.MethodPost, http.MethodDelete)
		}
	}
}

func newListJobRunLocksHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowMethods(w, http.MethodGet) {
			return
		}
		locks, err := db.ListJobRunLocks(database)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if locks == nil {
			locks = []db.JobRunLock{}
		}
		writeJSON(w, http.StatusOK, locks)
	}
}
