package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/juzu400/avito-internship/internal/domain"
)

type userRepositoryPG struct {
	db *DB
}

func NewUserRepository(db *DB) *userRepositoryPG {
	return &userRepositoryPG{db: db}
}

func (r *userRepositoryPG) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	row := r.db.Pool.QueryRow(ctx, `
        SELECT user_id, username, is_active
        FROM users
        WHERE user_id = $1
    `, string(id))

	var u domain.User
	if err := row.Scan(&u.ID, &u.Username, &u.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id %s: %w", id, err)
	}

	return &u, nil
}

func (r *userRepositoryPG) SetIsActive(ctx context.Context, id domain.UserID, active bool) error {
	cmd, err := r.db.Pool.Exec(ctx, `
        UPDATE users
        SET is_active = $2
        WHERE user_id = $1
    `, string(id), active)
	if err != nil {
		return fmt.Errorf("set is_active for %s: %w", id, err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
