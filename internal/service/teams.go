package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/juzu400/avito-internship/internal/domain"
)

func (s *TeamsService) UpsertTeam(ctx context.Context, team *domain.Team) error {
	if team == nil {
		err := fmt.Errorf("%w: team is nil", domain.ErrValidation)
		s.log.Warn("validate UpsertTeam failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "nil team"),
		)
		return err
	}
	if team.Name == "" {
		err := fmt.Errorf("%w: team_name is empty", domain.ErrValidation)
		s.log.Warn("validate UpsertTeam failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty team_name"),
		)
		return err
	}

	s.log.Info("upserting team",
		slog.String("team_name", team.Name),
		slog.Int("members_count", len(team.Members)),
	)

	seen := make(map[domain.UserID]struct{}, len(team.Members))
	for _, m := range team.Members {
		if m.ID == "" {
			err := fmt.Errorf("%w: member user_id is empty", domain.ErrValidation)
			s.log.Warn("validate UpsertTeam failed",
				slog.String("error_code", ErrCodeValidation),
				slog.String("reason", "member with empty user_id"),
			)
			return err
		}
		if _, ok := seen[m.ID]; ok {
			err := fmt.Errorf("%w: duplicate member %s", domain.ErrValidation, m.ID)
			s.log.Warn("validate UpsertTeam failed",
				slog.String("error_code", ErrCodeValidation),
				slog.String("reason", "duplicate team member"),
				slog.String("user_id", string(m.ID)),
			)
			return err
		}
		seen[m.ID] = struct{}{}
	}

	err := s.teams.UpsertTeam(ctx, team)
	if err != nil {
		s.log.Error("UpsertTeam failed",
			slog.String("team_name", team.Name),
			slog.String("error_code", errorCode(err)),
			slog.Any("err", err),
		)
	}
	return err
}

func (s *TeamsService) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	if name == "" {
		err := fmt.Errorf("%w: team_name is empty", domain.ErrValidation)
		s.log.Warn("validate GetByName failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty team_name"),
		)
		return nil, err
	}

	s.log.Info("get team by name", slog.String("team_name", name))

	team, err := s.teams.GetByName(ctx, name)
	if err != nil {
		s.log.Error("GetByName failed",
			slog.String("team_name", name),
			slog.String("error_code", errorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}
	return team, nil
}

func (s *TeamsService) GetByMemberID(ctx context.Context, userID domain.UserID) (*domain.Team, error) {
	if userID == "" {
		err := fmt.Errorf("%w: user_id is empty", domain.ErrValidation)
		s.log.Warn("validate GetByMemberID failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty user_id"),
		)
		return nil, err
	}

	s.log.Info("get team by member", slog.String("user_id", string(userID)))

	team, err := s.teams.GetByMemberID(ctx, userID)
	if err != nil {
		s.log.Error("GetByMemberID failed",
			slog.String("user_id", string(userID)),
			slog.String("error_code", errorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}
	return team, nil
}
