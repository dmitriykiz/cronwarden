package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/cronwarden/cronwarden/internal/db"
)

type noteRequest struct {
	Note string `json:"note"`
}

// newJobNotesHandler returns an http.Handler that manages notes for a specific run.
// Routes:
//
//	GET  /runs/{runID}/notes        — list notes
//	POST /runs/{runID}/notes        — add note
//	DELETE /runs/{runID}/notes/{id} — delete note
func newJobNotesHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// path: /runs/{runID}/notes[/{noteID}]
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		// parts[0]=runs, parts[1]=runID, parts[2]=notes, parts[3]=noteID (optional)
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		runID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			http.Error(w, "invalid run id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			notes, err := db.ListJobNotes(database, runID)
			if err != nil {
				http.Error(w, "db error", http.StatusInternalServerError)
				return
			}
			if notes == nil {
				notes = []db.JobNote{}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(notes)

		case http.MethodPost:
			var req noteRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Note == "" {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			noteID, err := db.AddJobNote(database, runID, req.Note)
			if err != nil {
				http.Error(w, "db error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]int64{"id": noteID})

		case http.MethodDelete:
			if len(parts) < 4 {
				http.Error(w, "missing note id", http.StatusBadRequest)
				return
			}
			noteID, err := strconv.ParseInt(parts[3], 10, 64)
			if err != nil {
				http.Error(w, "invalid note id", http.StatusBadRequest)
				return
			}
			found, err := db.DeleteJobNote(database, noteID)
			if err != nil {
				http.Error(w, "db error", http.StatusInternalServerError)
				return
			}
			if !found {
				http.NotFound(w, r)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
