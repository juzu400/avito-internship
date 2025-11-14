package repository

//go:generate mockgen -source=repository.go -destination=mocks/repository_mocks.go -package=mocks

import (
	"context"

	"github.com/juzu400/avito-internship/internal/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	SetIsActive(ctx context.Context, id domain.UserID, active bool) error
}

type TeamRepository interface {
	UpsertTeam(ctx context.Context, team *domain.Team) error
	GetByName(ctx context.Context, name string) (*domain.Team, error)
	GetByMemberID(ctx context.Context, userID domain.UserID) (*domain.Team, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr *domain.PullRequest) error
	Update(ctx context.Context, pr *domain.PullRequest) error
	GetByID(ctx context.Context, id domain.PullRequestID) (*domain.PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID domain.UserID) ([]*domain.PullRequest, error)
}

type Repositories struct {
	Users        UserRepository
	Teams        TeamRepository
	PullRequests PullRequestRepository
}

func NewRepositories(db *DB) *Repositories {
	return &Repositories{
		Users:        NewUserRepository(db),
		Teams:        NewTeamRepository(db),
		PullRequests: NewPullRequestRepository(db),
	}
}
