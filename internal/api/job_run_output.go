package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/cronwarden/cronwarden/internal/db"
)

func newJobRunOutputHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID, ok := parseID(w, r, "runID")
		if !ok {
			return
		}

		switch r.Method {
		case http.MethodGet:
			out, err := db.GetJobRunOutput(database, runID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			if out == nil {
				writeError(w, http.StatusNotFound, "no output recorded for this run")
				return
			}
			writeJSON(w, http.StatusOK, out)

		case http.MethodPut:
			var body struct {
				Stdout string `json:"stdout"`
				Stderr string `json:"stderr"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON")
				return
			}
			if err := db.UpsertJobRunOutput(database, runID, body.Stdout, body.Stderr); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if err := db.DeleteJobRunOutput(database, runID); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	}
}
