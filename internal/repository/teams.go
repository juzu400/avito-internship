package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// UpsertTeam creates a new team with the given members or fails if a team
// with the same name already exists. User records are upserted into the users
// table.
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

// GetByMemberID returns a team and its members for the given user ID.
// If the user does not belong to any team, domain.ErrNotFound is returned.
func (r *teamRepositoryPG) GetByName(ctx context.Context, name string) (*domain.Team, error) {
	var teamID int64
	row := r.db.Pool.QueryRow(ctx, `
        SELECT id
        FROM teams
        WHERE team_name = $1
    `, name)

	if err := row.Scan(&teamID); err != nil {
		return nil, domain.ErrNotFound
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

// GetTeamsByMemberIDs returns teams for the given user IDs keyed by user ID.
// Users that are not members of any team are not present in the result map.
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
		return nil, domain.ErrNotFound
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

func (r *teamRepositoryPG) GetTeamsByMemberIDs(ctx context.Context, userIDs []domain.UserID) (map[domain.UserID]*domain.Team, error) {
	if len(userIDs) == 0 {
		return map[domain.UserID]*domain.Team{}, nil
	}

	args := make([]any, len(userIDs))
	placeholders := make([]string, len(userIDs))
	for i, id := range userIDs {
		args[i] = string(id)
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	rows, err := r.db.Pool.Query(ctx, fmt.Sprintf(`
        SELECT tm.user_id, t.team_name
        FROM team_members tm
        JOIN teams t ON t.id = tm.team_id
        WHERE tm.user_id IN (%s)
    `, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, fmt.Errorf("get teams by member ids: %w", err)
	}
	defer rows.Close()

	res := make(map[domain.UserID]*domain.Team, len(userIDs))
	for rows.Next() {
		var userID string
		var teamName string

		if err := rows.Scan(&userID, &teamName); err != nil {
			return nil, fmt.Errorf("scan team by member id: %w", err)
		}

		res[domain.UserID(userID)] = &domain.Team{
			Name: teamName,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teams by member ids: %w", err)
	}

	return res, nil
}
