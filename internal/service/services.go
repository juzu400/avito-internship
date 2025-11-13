package service

import (
	"log/slog"

	"github.com/juzu400/avito-internship/internal/repository"
)

type UsersService struct {
	log   *slog.Logger
	users repository.UserRepository
	prs   repository.PullRequestRepository
}

type TeamsService struct {
	log   *slog.Logger
	teams repository.TeamRepository
}

type PullRequestService struct {
	log   *slog.Logger
	users repository.UserRepository
	teams repository.TeamRepository
	prs   repository.PullRequestRepository
}

type Services struct {
	Users        *UsersService
	Teams        *TeamsService
	PullRequests *PullRequestService
}

func NewServices(log *slog.Logger, repos *repository.Repositories) *Services {
	return &Services{
		Users: &UsersService{
			log:   log.With(slog.String("service", "users")),
			users: repos.Users,
			prs:   repos.PullRequests,
		},
		Teams: &TeamsService{
			log:   log.With(slog.String("service", "teams")),
			teams: repos.Teams,
		},
		PullRequests: &PullRequestService{
			log:   log.With(slog.String("service", "pull_requests")),
			users: repos.Users,
			teams: repos.Teams,
			prs:   repos.PullRequests,
		},
	}
}
