package service

import "log/slog"

type UsersService struct {
	log *slog.Logger
}

type TeamsService struct {
	log *slog.Logger
}

type PullRequestService struct {
	log *slog.Logger
}

type Services struct {
	Users        *UsersService
	Teams        *TeamsService
	PullRequests *PullRequestService
}

func NewServices(log *slog.Logger) *Services {
	return &Services{
		Users:        &UsersService{log: log},
		Teams:        &TeamsService{log: log},
		PullRequests: &PullRequestService{log: log},
	}
}
