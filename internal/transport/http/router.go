package http

import (
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5"

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

	r.Post("/team/add", h.AddTeam)
	//r.Get("/team/get", h.GetTeam)
	//
	//r.Post("/users/setIsActive", h.SetUserActive)
	//r.Get("/users/getReview", h.GetUserReview)
	//
	//r.Post("/pullRequest/create", h.CreatePullRequest)
	//r.Post("/pullRequest/merge", h.MergePullRequest)
	//r.Post("/pullRequest/reassign", h.ReassignReviewer)

	return r
}
