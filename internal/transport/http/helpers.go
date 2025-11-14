package http

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	resp := ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	}

	writeJSON(w, status, resp)
}
