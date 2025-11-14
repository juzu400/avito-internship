package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/juzu400/avito-internship/internal/domain"
)

type teamRepositoryPG struct {
	db *DB
}

func NewTeamRepository(db *DB) *teamRepositoryPG {
	return &teamRepositoryPG{db: db}
}

func (r *teamRepositoryPG) UpsertTeam(ctx context.Context, team *domain.Team) error {
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var teamID int64

	err = tx.QueryRow(ctx, `
        INSERT INTO teams (team_name)
        VALUES ($1)
        RETURNING id
    `, team.Name).Scan(&teamID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrTeamAlreadyExists
		}
		return fmt.Errorf("insert team: %w", err)
	}

	for _, m := range team.Members {
		_, err := tx.Exec(ctx, `
            INSERT INTO users (user_id, username, is_active)
            VALUES ($1, $2, $3)
            ON CONFLICT (user_id)
            DO UPDATE SET username = EXCLUDED.username,
                          is_active = EXCLUDED.is_active
        `, string(m.ID), m.Username, m.IsActive)
		if err != nil {
			return fmt.Errorf("upsert user %s: %w", m.ID, err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM team_members WHERE team_id = $1`, teamID); err != nil {
		return fmt.Errorf("delete old team members: %w", err)
	}

	for _, m := range team.Members {
		if _, err := tx.Exec(ctx, `
            INSERT INTO team_members (team_id, user_id)
            VALUES ($1, $2)
        `, teamID, string(m.ID)); err != nil {
			return fmt.Errorf("insert team member %s: %w", m.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *teamRepositoryPG) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	var teamID int64
	row := r.db.Pool.QueryRow(ctx, `
        SELECT id
        FROM teams
        WHERE team_name = $1
    `, name)

	if err := row.Scan(&teamID); err != nil {
		return nil, domain.ErrTeamNotFound
	}

	team := &domain.Team{
		Name:    name,
		Members: make([]domain.User, 0),
	}

	rows, err := r.db.Pool.Query(ctx, `
        SELECT u.user_id, u.username, u.is_active
        FROM team_members tm
        JOIN users u ON u.user_id = tm.user_id
        WHERE tm.team_id = $1
        ORDER BY u.user_id
    `, teamID)
	if err != nil {
		return nil, fmt.Errorf("query team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		team.Members = append(team.Members, u)
	}

	return team, nil
}

func (r *teamRepositoryPG) GetByMemberID(ctx context.Context, userID domain.UserID) (*domain.Team, error) {
	var teamID int64
	var teamName string

	row := r.db.Pool.QueryRow(ctx, `
        SELECT t.id, t.team_name
        FROM team_members tm
        JOIN teams t ON t.id = tm.team_id
        WHERE tm.user_id = $1
        LIMIT 1
    `, string(userID))

	if err := row.Scan(&teamID, &teamName); err != nil {
		return nil, domain.ErrTeamNotFound
	}

	team := &domain.Team{
		Name:    teamName,
		Members: make([]domain.User, 0),
	}

	rows, err := r.db.Pool.Query(ctx, `
        SELECT u.user_id, u.username, u.is_active
        FROM team_members tm
        JOIN users u ON u.user_id = tm.user_id
        WHERE tm.team_id = $1
        ORDER BY u.user_id
    `, teamID)
	if err != nil {
		return nil, fmt.Errorf("query team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		team.Members = append(team.Members, u)
	}

	return team, nil
}
