package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cronwarden/internal/db"
)

type jobRunEventsHandler struct {
	db *db.DB
}

func newJobRunEventsHandler(database *db.DB) *jobRunEventsHandler {
	return &jobRunEventsHandler{db: database}
}

func (h *jobRunEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runID, ok := parseID(w, r, "run_id")
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.list(w, r, runID)
	case http.MethodPost:
		h.add(w, r, runID)
	case http.MethodDelete:
		h.deleteAll(w, r, runID)
	default:
		allowMethods(w, http.MethodGet, http.MethodPost, http.MethodDelete)
	}
}

func (h *jobRunEventsHandler) list(w http.ResponseWriter, _ *http.Request, runID int64) {
	events, err := h.db.ListJobRunEvents(runID)
	if err != nil {
		writeError(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, events)
}

func (h *jobRunEventsHandler) add(w http.ResponseWriter, r *http.Request, runID int64) {
	var body struct {
		EventType string `json:"event_type"`
		Message   string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.EventType == "" {
		writeError(w, "event_type is required", http.StatusBadRequest)
		return
	}
	if err := h.db.AddJobRunEvent(runID, body.EventType, body.Message); err != nil {
		writeError(w, "failed to add event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *jobRunEventsHandler) deleteAll(w http.ResponseWriter, r *http.Request, runID int64) {
	// Support DELETE /runs/{run_id}/events/{event_id} via query param
	if idStr := r.URL.Query().Get("id"); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			writeError(w, "invalid event id", http.StatusBadRequest)
			return
		}
		_ = runID
		if err := h.db.DeleteJobRunEvent(id); err != nil {
			writeError(w, "failed to delete event", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := h.db.DeleteAllJobRunEvents(runID); err != nil {
		writeError(w, "failed to delete events", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
