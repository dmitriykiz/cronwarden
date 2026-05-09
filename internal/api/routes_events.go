package api

import "net/http"

// registerJobRunEventRoutes wires the job run events handler into the given mux.
// Pattern: /runs/{run_id}/events
func (s *Server) registerJobRunEventRoutes(mux *http.ServeMux) {
	h := newJobRunEventsHandler(s.db)
	mux.Handle("/runs/", http.StripPrefix("/runs/", &runEventRouter{handler: h}))
}

// runEventRouter dispatches /runs/{run_id}/events requests by extracting run_id
// from the path and delegating to the events handler.
type runEventRouter struct {
	handler *jobRunEventsHandler
}

func (r *runEventRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Expect path: {run_id}/events[?id=N]
	var runIDStr string
	path := req.URL.Path
	for i, c := range path {
		if c == '/' {
			runIDStr = path[:i]
			path = path[i+1:]
			break
		}
	}
	if runIDStr == "" {
		runIDStr = path
		path = ""
	}
	if path != "events" && path != "events/" && path != "" {
		http.NotFound(w, req)
		return
	}
	req = withPathParam(req, "run_id", runIDStr)
	r.handler.ServeHTTP(w, req)
}
