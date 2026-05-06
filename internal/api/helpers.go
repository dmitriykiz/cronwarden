package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Best-effort; headers already sent.
		_ = err
	}
}

// writeError writes a JSON error payload with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseID extracts a named path parameter as an int64.
// Returns the id and true on success, or writes a 400 response and returns false.
func parseID(w http.ResponseWriter, r *http.Request, param string) (int64, bool) {
	raw := r.PathValue(param)
	if raw == "" {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("missing path parameter: %s", param))
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid %s: %s", param, raw))
		return 0, false
	}
	return id, true
}

// allowMethods checks that the request method is one of the allowed methods.
// If not, it writes a 405 response and returns false.
func allowMethods(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}
	w.Header().Set("Allow", joinMethods(methods))
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	return false
}

func joinMethods(methods []string) string {
	out := ""
	for i, m := range methods {
		if i > 0 {
			out += ", "
		}
		out += m
	}
	return out
}
