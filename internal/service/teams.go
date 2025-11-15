package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/juzu400/avito-internship/internal/domain"
)

// UpsertTeam validates the team and ensures each member belongs to at most one team,
// then creates or updates the team in the repository. If any member already belongs
// to a different team, ErrValidation is returned.
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
	memberIDs := make([]domain.UserID, 0, len(team.Members))
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
		memberIDs = append(memberIDs, m.ID)
	}

	if len(memberIDs) > 0 {
		existingTeams, err := s.teams.GetTeamsByMemberIDs(ctx, memberIDs)
		if err != nil {
			s.log.Error("UpsertTeam: GetTeamsByMemberIDs failed",
				slog.String("team_name", team.Name),
				slog.String("error_code", ErrorCode(err)),
				slog.Any("err", err),
			)
			return err
		}

		for _, id := range memberIDs {
			existingTeam, ok := existingTeams[id]
			if !ok {
				continue
			}
			if existingTeam.Name != team.Name {
				err := fmt.Errorf("%w: user %s already in team %s", domain.ErrValidation, id, existingTeam.Name)
				s.log.Warn("validate UpsertTeam failed",
					slog.String("error_code", ErrCodeValidation),
					slog.String("reason", "user already in another team"),
					slog.String("user_id", string(id)),
					slog.String("existing_team", existingTeam.Name),
				)
				return err
			}
		}
	}

	err := s.teams.UpsertTeam(ctx, team)
	if err != nil {
		s.log.Error("UpsertTeam failed",
			slog.String("team_name", team.Name),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
	}
	return err
}

// GetByName returns a team with all its members by team name.
// If the name is empty, ErrValidation is returned. If the team does not exist,
// domain.ErrNotFound is returned from the repository.
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
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}
	return team, nil
}

// GetByMemberID returns a team for the given user ID.
// If userID is empty, ErrValidation is returned. If the user does not belong
// to any team, domain.ErrNotFound is returned.
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
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}
	return team, nil
}
