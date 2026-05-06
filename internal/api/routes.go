package api

import "net/http"

// registerRoutes attaches all HTTP handlers to the server's mux.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/runs", s.handleListRuns)
	s.mux.HandleFunc("/runs/", s.handleListRunsByJob)
	s.mux.HandleFunc("/stats", s.handleStats)
}

// handleListRunsByJob handles GET /runs/{job_name} and returns runs filtered
// by job name using the last path segment.
func (s *Server) handleListRunsByJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract job name from path: /runs/<name>
	jobName := r.URL.Path[len("/runs/"):]
	if jobName == "" {
		http.Error(w, "job name required", http.StatusBadRequest)
		return
	}

	s.writeJobRuns(w, r, jobName)
}
