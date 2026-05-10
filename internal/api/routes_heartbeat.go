package api

import (
	"database/sql"
	"net/http"
)

// registerHeartbeatRoutes wires up the heartbeat-related API endpoints onto mux.
//
//	PUT    /api/runs/{id}/heartbeat  — upsert a heartbeat for a run
//	GET    /api/runs/{id}/heartbeat  — fetch the current heartbeat for a run
//	DELETE /api/runs/{id}/heartbeat  — remove the heartbeat record
//	GET    /api/heartbeats/expired   — list all runs whose heartbeat has timed out
func registerHeartbeatRoutes(mux *http.ServeMux, database *sql.DB) {
	heartbeatHandler := newJobRunHeartbeatHandler(database)

	// The per-run heartbeat endpoint is matched by prefix so that parseID can
	// extract the run ID from the path segment between /api/runs/ and /heartbeat.
	mux.Handle("/api/runs/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle paths that end with /heartbeat; delegate everything else.
		if len(r.URL.Path) > len("/api/runs/") && hasTrailingSuffix(r.URL.Path, "/heartbeat") {
			heartbeatHandler.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}))

	mux.HandleFunc("/api/heartbeats/expired", newListExpiredHeartbeatsHandler(database))
}

// hasTrailingSuffix reports whether s ends with suffix.
func hasTrailingSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
