package http

import (
	"net/http"
)

// Health is a simple liveness endpoint that returns "ok" with HTTP 200 status.
// It can be used by load balancers or orchestration systems to check that
// the service process is running.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
