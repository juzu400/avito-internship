package service

import (
	"context"

	"github.com/juzu400/avito-internship/internal/domain"
)

func (s *TeamsService) UpsertTeam(ctx context.Context, team *domain.Team) error {
	return s.teams.UpsertTeam(ctx, team)
}
