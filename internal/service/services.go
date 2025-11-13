package service

import (
	"log/slog"

	"github.com/juzu400/avito-internship/internal/repository"
)

type UsersService struct {
	log *slog.Logger
}

type TeamsService struct {
	log   *slog.Logger
	teams repository.TeamRepository
}

type PullRequestService struct {
	log *slog.Logger
}

type Services struct {
	Users        *UsersService
	Teams        *TeamsService
	PullRequests *PullRequestService
}

func NewServices(log *slog.Logger, repos *repository.Repositories) *Services {
	return &Services{
		Users: &UsersService{
			log: log,
		},
		Teams: &TeamsService{
			log:   log,
			teams: repos.Teams,
		},
		PullRequests: &PullRequestService{
			log: log,
		},
	}
}
