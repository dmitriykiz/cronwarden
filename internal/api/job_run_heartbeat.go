package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

func newJobRunHeartbeatHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID, ok := parseID(w, r, "/api/runs/")
		if !ok {
			return
		}

		switch r.Method {
		case http.MethodGet:
			hb, err := db.GetJobRunHeartbeat(database, runID)
			if err != nil {
				if isNotFound(err) {
					writeError(w, http.StatusNotFound, "heartbeat not found")
					return
				}
				writeError(w, http.StatusInternalServerError, "failed to fetch heartbeat")
				return
			}
			writeJSON(w, http.StatusOK, hb)

		case http.MethodPut:
			var body struct {
				LastSeenAt time.Time `json:"last_seen_at"`
				TimeoutAt  time.Time `json:"timeout_at"`
			}
			if err := readJSON(r, &body); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if body.TimeoutAt.IsZero() {
				writeError(w, http.StatusBadRequest, "timeout_at is required")
				return
			}
			if body.LastSeenAt.IsZero() {
				body.LastSeenAt = time.Now().UTC()
			}
			if err := db.UpsertJobRunHeartbeat(database, runID, body.LastSeenAt, body.TimeoutAt); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to upsert heartbeat")
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if err := db.DeleteJobRunHeartbeat(database, runID); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to delete heartbeat")
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	}
}

func newListExpiredHeartbeatsHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowMethods(w, http.MethodGet) {
			return
		}
		alerts, err := db.ListExpiredHeartbeatAlerts(database, time.Now())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list expired heartbeats")
			return
		}
		if alerts == nil {
			alerts = []db.HeartbeatAlert{}
		}
		writeJSON(w, http.StatusOK, alerts)
	}
}
