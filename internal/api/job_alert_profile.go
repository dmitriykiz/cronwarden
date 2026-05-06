package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cronwarden/cronwarden/internal/db"
)

type alertProfileRequest struct {
	MaxDurationSecs int     `json:"max_duration_secs"`
	MinSuccessRate  float64 `json:"min_success_rate"`
}

func newJobAlertProfileHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect path: /jobs/{name}/alert-profile
		seg := strings.TrimPrefix(r.URL.Path, "/jobs/")
		seg = strings.TrimSuffix(seg, "/alert-profile")
		jobName := strings.TrimSpace(seg)
		if jobName == "" {
			writeError(w, http.StatusBadRequest, "missing job name")
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetAlertProfile(w, database, jobName)
		case http.MethodPut:
			handlePutAlertProfile(w, r, database, jobName)
		case http.MethodDelete:
			handleDeleteAlertProfile(w, database, jobName)
		default:
			allowMethods(w, http.MethodGet, http.MethodPut, http.MethodDelete)
		}
	}
}

func handleGetAlertProfile(w http.ResponseWriter, database *sql.DB, jobName string) {
	p, err := db.GetAlertProfile(database, jobName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "alert profile not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch alert profile")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func handlePutAlertProfile(w http.ResponseWriter, r *http.Request, database *sql.DB, jobName string) {
	var req alertProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	now := time.Now().UTC()
	p := db.AlertProfile{
		JobName:         jobName,
		MaxDurationSecs: req.MaxDurationSecs,
		MinSuccessRate:  req.MinSuccessRate,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.UpsertAlertProfile(database, p); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save alert profile")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func handleDeleteAlertProfile(w http.ResponseWriter, database *sql.DB, jobName string) {
	if err := db.DeleteAlertProfile(database, jobName); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete alert profile")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
