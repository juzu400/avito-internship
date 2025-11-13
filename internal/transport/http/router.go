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
		log:      log.With(slog.String("layer", "http")),
		services: services,
	}

	r := chi.NewRouter()

	// health
	r.Get("/health", h.Health)

	// teams
	r.Post("/team/add", h.AddTeam)
	r.Get("/team/get", h.GetTeam)

	// users
	r.Post("/users/setIsActive", h.SetUserActive)
	r.Get("/users/getReview", h.GetUserReview)

	// pull requests
	r.Post("/pullRequest/create", h.CreatePullRequest)
	r.Post("/pullRequest/merge", h.MergePullRequest)
	r.Post("/pullRequest/reassign", h.ReassignReviewer)

	return r
}
