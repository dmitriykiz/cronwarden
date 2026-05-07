package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cronwarden/cronwarden/internal/db"
)

// newJobEnvHandler handles GET/PUT/DELETE for a single env var and GET for all vars.
// Routes:
//
//	GET    /jobs/{name}/env          → list all env vars
//	PUT    /jobs/{name}/env/{key}    → set a single env var
//	DELETE /jobs/{name}/env/{key}    → delete a single env var
func newJobEnvHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Path: /jobs/{name}/env[/{key}]
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		// parts: ["jobs", name, "env"] or ["jobs", name, "env", key]
		if len(parts) < 3 {
			writeError(w, http.StatusBadRequest, "invalid path")
			return
		}
		jobName := parts[1]

		if len(parts) == 3 {
			// /jobs/{name}/env — list only
			if !allowMethods(w, r, http.MethodGet) {
				return
			}
			vars, err := db.ListJobEnvVars(database, jobName)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, vars)
			return
		}

		key := parts[3]

		switch r.Method {
		case http.MethodPut:
			var body struct {
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if err := db.UpsertJobEnvVar(database, jobName, key, body.Value); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if err := db.DeleteJobEnvVar(database, jobName, key); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			writeError(w, http.StatusMethodNotAllowed,
				"allowed: "+joinMethods(http.MethodPut, http.MethodDelete))
		}
	}
}
