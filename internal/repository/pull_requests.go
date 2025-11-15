package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/juzu400/avito-internship/internal/domain"
)

type pullRequestRepositoryPG struct {
	db *DB
}

func NewPullRequestRepository(db *DB) *pullRequestRepositoryPG {
	return &pullRequestRepositoryPG{db: db}
}

// Create inserts a new pull request and its reviewers into the database.
// If a pull request with the same ID already exists, ErrPullRequestAlreadyExists is returned.
func (r *pullRequestRepositoryPG) Create(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if pr.CreatedAt.IsZero() {
		pr.CreatedAt = time.Now().UTC()
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `,
		string(pr.ID),
		pr.Name,
		string(pr.AuthorID),
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrPullRequestAlreadyExists
		}
		return fmt.Errorf("insert pull_request: %w", err)
	}

	if err := saveReviewers(ctx, tx, pr.ID, pr.AssignedReviewers); err != nil {
		return fmt.Errorf("save reviewers: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// Update updates pull request fields and completely replaces its reviewers.
// If the pull request does not exist, ErrNotFound is returned.
func (r *pullRequestRepositoryPG) Update(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	cmd, err := tx.Exec(ctx, `
        UPDATE pull_requests
        SET pull_request_name = $2,
            author_id = $3,
            status = $4,
            merged_at = $5
        WHERE pull_request_id = $1
    `,
		string(pr.ID),
		pr.Name,
		string(pr.AuthorID),
		string(pr.Status),
		pr.MergedAt,
	)
	if err != nil {
		return fmt.Errorf("update pull_request: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err := tx.Exec(ctx, `
        DELETE FROM pull_request_reviewers
        WHERE pull_request_id = $1
    `, string(pr.ID)); err != nil {
		return fmt.Errorf("delete reviewers: %w", err)
	}

	if err := saveReviewers(ctx, tx, pr.ID, pr.AssignedReviewers); err != nil {
		return fmt.Errorf("save reviewers: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// GetByID retrieves a pull request with its reviewers by ID.
// If the pull request does not exist, ErrNotFound is returned.
func (r *pullRequestRepositoryPG) GetByID(ctx context.Context, id domain.PullRequestID) (*domain.PullRequest, error) {
	row := r.db.Pool.QueryRow(ctx, `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests
        WHERE pull_request_id = $1
    `, string(id))

	var pr domain.PullRequest
	var status string
	var mergedAt *time.Time
	if err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &status, &pr.CreatedAt, &mergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get pull_request by id %s: %w", id, err)
	}
	pr.Status = domain.PullRequestStatus(status)
	pr.MergedAt = mergedAt

	rows, err := r.db.Pool.Query(ctx, `
        SELECT reviewer_id
        FROM pull_request_reviewers
        WHERE pull_request_id = $1
        ORDER BY reviewer_id
    `, string(id))
	if err != nil {
		return nil, fmt.Errorf("query reviewers: %w", err)
	}
	defer rows.Close()

	pr.AssignedReviewers = make([]domain.UserID, 0)
	for rows.Next() {
		var rid domain.UserID
		if err := rows.Scan(&rid); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, rid)
	}

	return &pr, nil
}

// ListByReviewer returns pull requests where the given user is assigned as a reviewer,
// ordered by creation time in descending order.
func (r *pullRequestRepositoryPG) ListByReviewer(ctx context.Context, reviewerID domain.UserID) ([]*domain.PullRequest, error) {
	rows, err := r.db.Pool.Query(ctx, `
        SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status, p.created_at, p.merged_at
        FROM pull_requests p
        JOIN pull_request_reviewers r ON r.pull_request_id = p.pull_request_id
        WHERE r.reviewer_id = $1
        ORDER BY p.created_at DESC
    `, string(reviewerID))
	if err != nil {
		return nil, fmt.Errorf("query pull_requests by reviewer: %w", err)
	}
	defer rows.Close()

	result := make([]*domain.PullRequest, 0)

	for rows.Next() {
		var pr domain.PullRequest
		var status string
		var mergedAt *time.Time

		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &status, &pr.CreatedAt, &mergedAt); err != nil {
			return nil, fmt.Errorf("scan pull_request: %w", err)
		}
		pr.Status = domain.PullRequestStatus(status)
		pr.MergedAt = mergedAt

		result = append(result, &pr)
	}

	return result, nil
}

// saveReviewers stores reviewer assignments for the given pull request inside the transaction.
func saveReviewers(ctx context.Context, tx pgx.Tx, prID domain.PullRequestID, reviewers []domain.UserID) error {
	for _, rid := range reviewers {
		if _, err := tx.Exec(ctx, `
            INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
            VALUES ($1, $2)
        `, string(prID), string(rid)); err != nil {
			return fmt.Errorf("insert reviewer %s: %w", rid, err)
		}
	}
	return nil
}

// Merge atomically marks a pull request as merged.
// If the pull request is already merged, it returns the existing state without error.
// If the pull request does not exist, ErrNotFound is returned.
func (r *pullRequestRepositoryPG) Merge(
	ctx context.Context,
	id domain.PullRequestID,
	mergedAt time.Time,
) (*domain.PullRequest, error) {
	cmdTag, err := r.db.Pool.Exec(ctx, `
        UPDATE pull_requests
        SET status = $2,
            merged_at = $3
        WHERE pull_request_id = $1
          AND status = $4
    `,
		id,
		domain.PRStatusMerged,
		mergedAt,
		domain.PRStatusOpen,
	)
	if err != nil {
		return nil, fmt.Errorf("merge pull request %s: %w", id, err)
	}

	if cmdTag.RowsAffected() == 0 {
		pr, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if pr.IsMerged() {
			return pr, nil
		}
		return nil, fmt.Errorf("merge pull request %s: unexpected status %s", id, pr.Status)
	}

	pr, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return pr, nil
}
