package api

import "net/http"

// registerRoutes attaches all HTTP handlers to the server's mux.
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/api/runs", s.handleListRuns)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/jobs/status", s.handleJobStatuses)
}

// ServeHTTP implements http.Handler so *Server can be passed directly to
// httptest.NewServer and http.ListenAndServe.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
