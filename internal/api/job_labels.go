package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cronwarden/cronwarden/internal/db"
)

// newJobLabelsHandler returns an http.Handler for GET/PUT/DELETE on
// /jobs/{name}/labels.
//
// GET    /jobs/{name}/labels  - returns all labels for the job as a JSON object
// PUT    /jobs/{name}/labels  - replaces all labels with the provided JSON object
// DELETE /jobs/{name}/labels  - removes all labels for the job
func newJobLabelsHandler(sqlDB *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract job name from path: /jobs/<name>/labels
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			writeError(w, http.StatusBadRequest, "missing job name")
			return
		}
		jobName := parts[1]
		if jobName == "" {
			writeError(w, http.StatusBadRequest, "job name must not be empty")
			return
		}

		switch r.Method {
		case http.MethodGet:
			labels, err := db.ListJobLabels(sqlDB, jobName)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, labels)

		case http.MethodPut:
			var labels map[string]string
			if err := json.NewDecoder(r.Body).Decode(&labels); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if err := db.SetJobLabels(sqlDB, jobName, labels); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if err := db.DeleteJobLabels(sqlDB, jobName); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	})
}
