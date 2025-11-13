package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"log/slog"

	"github.com/juzu400/avito-internship/internal/service"
)

type Handler struct {
	log      *slog.Logger
	services *service.Services
}

func NewRouter(log *slog.Logger, services *service.Services) http.Handler {
	h := &Handler{
		log:      log,
		services: services,
	}

	r := chi.NewRouter()

	r.Get("/health", h.Health)

	return r
}
