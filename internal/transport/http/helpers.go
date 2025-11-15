package http

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes the given value as a JSON response with the provided HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError builds a standard JSON error response with the given status, code and message
func writeError(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	}

	writeJSON(w, status, resp)
}
