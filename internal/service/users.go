package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/juzu400/avito-internship/internal/domain"
)

// validateUserID checks that user ID is not empty and logs validation errors
// with the given operation name.
func (s *UsersService) validateUserID(op string, id domain.UserID) error {
	if id == "" {
		err := fmt.Errorf("%w: user_id is empty", domain.ErrValidation)
		s.log.Warn("validate "+op+" failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty user_id"),
		)
		return err
	}
	return nil
}

// SetIsActive updates the is_active flag for the given user.
// If the user does not exist, domain.ErrNotFound is returned from the repository.
func (s *UsersService) SetIsActive(ctx context.Context, id domain.UserID, active bool) error {
	if err := s.validateUserID("SetIsActive", id); err != nil {
		return err
	}

	s.log.Info("setting user active flag",
		slog.String("user_id", string(id)),
		slog.Bool("is_active", active),
	)

	if err := s.users.SetIsActive(ctx, id, active); err != nil {
		s.log.Error("SetIsActive failed",
			slog.String("user_id", string(id)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return err
	}

	return nil
}

// GetByID returns a user by ID.
// If the user does not exist, domain.ErrNotFound is returned.
func (s *UsersService) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if err := s.validateUserID("GetByID", id); err != nil {
		return nil, err
	}

	s.log.Info("get user by id", slog.String("user_id", string(id)))

	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		s.log.Error("GetByID failed",
			slog.String("user_id", string(id)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}

	return u, nil
}

// GetReviews returns pull requests where the given user is assigned as a reviewer.
// It first ensures that the user exists by calling GetByID.
func (s *UsersService) GetReviews(ctx context.Context, id domain.UserID) ([]*domain.PullRequest, error) {
	if _, err := s.GetByID(ctx, id); err != nil {
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
