package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cronwarden/cronwarden/internal/db"
)

func newJobOwnershipHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobName := r.URL.Query().Get("job")

		switch r.Method {
		case http.MethodGet:
			if jobName == "" {
				// List all
				list, err := db.ListJobOwnerships(database)
				if err != nil {
					writeError(w, http.StatusInternalServerError, err.Error())
					return
				}
				if list == nil {
					list = []db.JobOwnership{}
				}
				writeJSON(w, http.StatusOK, list)
				return
			}
			o, err := db.GetJobOwnership(database, jobName)
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "ownership not found")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, o)

		case http.MethodPut:
			if jobName == "" {
				writeError(w, http.StatusBadRequest, "job query param required")
				return
			}
			var o db.JobOwnership
			if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			o.JobName = jobName
			if err := db.UpsertJobOwnership(database, o); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, o)

		case http.MethodDelete:
			if jobName == "" {
				writeError(w, http.StatusBadRequest, "job query param required")
				return
			}
			if err := db.DeleteJobOwnership(database, jobName); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	}
}
