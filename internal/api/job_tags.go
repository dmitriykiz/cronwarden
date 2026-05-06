package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/example/cronwarden/internal/db"
)

type tagRequest struct {
	Tags []string `json:"tags"`
}

type tagResponse struct {
	JobName string   `json:"job_name"`
	Tags    []string `json:"tags"`
}

func newJobTagsHandler(sqlDB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// URL: /jobs/{name}/tags
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			writeError(w, http.StatusBadRequest, "invalid path")
			return
		}
		jobName := parts[1]

		switch r.Method {
		case http.MethodGet:
			tags, err := db.ListJobTags(sqlDB, jobName)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to list tags")
				return
			}
			if tags == nil {
				tags = []string{}
			}
			writeJSON(w, http.StatusOK, tagResponse{JobName: jobName, Tags: tags})

		case http.MethodPut:
			var req tagRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if err := db.SetJobTags(sqlDB, jobName, req.Tags); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to set tags")
				return
			}
			writeJSON(w, http.StatusOK, tagResponse{JobName: jobName, Tags: req.Tags})

		case http.MethodDelete:
			if err := db.DeleteJobTags(sqlDB, jobName); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to delete tags")
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	}
}
