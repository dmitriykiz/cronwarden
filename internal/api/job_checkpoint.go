package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/cronwarden/cronwarden/internal/db"
)

type checkpointRequest struct {
	Marker string `json:"marker"`
}

func newJobCheckpointHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobName := r.PathValue("name")
		if jobName == "" {
			writeError(w, http.StatusBadRequest, "missing job name")
			return
		}

		switch r.Method {
		case http.MethodGet:
			cp, err := db.GetJobCheckpoint(database, jobName)
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "no checkpoint found")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "db error")
				return
			}
			writeJSON(w, http.StatusOK, cp)

		case http.MethodPut:
			var req checkpointRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Marker == "" {
				writeError(w, http.StatusBadRequest, "invalid or missing marker")
				return
			}
			if err := db.UpsertJobCheckpoint(database, jobName, req.Marker); err != nil {
				writeError(w, http.StatusInternalServerError, "db error")
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if err := db.DeleteJobCheckpoint(database, jobName); err != nil {
				writeError(w, http.StatusInternalServerError, "db error")
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	}
}
