package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/juzu400/avito-internship/internal/domain"
)

func (s *UsersService) SetIsActive(ctx context.Context, id domain.UserID, active bool) error {
	if id == "" {
		err := fmt.Errorf("%w: user_id is empty", domain.ErrValidation)
		s.log.Warn("validate SetIsActive failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty user_id"),
		)
		return err
	}

	s.log.Info("setting user active flag",
		slog.String("user_id", string(id)),
		slog.Bool("is_active", active),
	)

	err := s.users.SetIsActive(ctx, id, active)
	if err != nil {
		s.log.Error("SetIsActive failed",
			slog.String("user_id", string(id)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
	}
	return err
}

func (s *UsersService) GetReviews(ctx context.Context, id domain.UserID) ([]*domain.PullRequest, error) {
	if id == "" {
		err := fmt.Errorf("%w: user_id is empty", domain.ErrValidation)
		s.log.Warn("validate GetReviews failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty user_id"),
		)
		return nil, err
	}

	s.log.Info("listing reviews for user", slog.String("user_id", string(id)))

	prs, err := s.prs.ListByReviewer(ctx, id)
	if err != nil {
		s.log.Error("GetReviews failed",
			slog.String("user_id", string(id)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}

	return prs, nil
}
