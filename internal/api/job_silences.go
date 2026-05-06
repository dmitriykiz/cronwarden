package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/user/cronwarden/internal/db"
)

func newJobSilencesHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Path: /api/jobs/{name}/silences[/{id}]
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		// parts: ["api", "jobs", "{name}", "silences"] or [..., "{id}"]
		if len(parts) < 4 {
			writeError(w, http.StatusBadRequest, "invalid path")
			return
		}
		jobName := parts[2]

		if len(parts) == 5 {
			// /api/jobs/{name}/silences/{id}
			if !allowMethods(w, r, http.MethodDelete) {
				return
			}
			id, err := strconv.ParseInt(parts[4], 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid id")
				return
			}
			ok, err := db.DeleteJobSilence(database, id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if !ok {
				writeError(w, http.StatusNotFound, "silence not found")
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// /api/jobs/{name}/silences
		switch r.Method {
		case http.MethodGet:
			silences, err := db.ListJobSilences(database, jobName)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if silences == nil {
				silences = []db.JobSilence{}
			}
			writeJSON(w, http.StatusOK, silences)

		case http.MethodPost:
			var body struct {
				StartsAt time.Time `json:"starts_at"`
				EndsAt   time.Time `json:"ends_at"`
				Reason   string    `json:"reason"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			id, err := db.UpsertJobSilence(database, db.JobSilence{
				JobName:  jobName,
				StartsAt: body.StartsAt,
				EndsAt:   body.EndsAt,
				Reason:   body.Reason,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, map[string]int64{"id": id})

		default:
			writeError(w, http.StatusMethodNotAllowed, joinMethods(http.MethodGet, http.MethodPost))
		}
	}
}
