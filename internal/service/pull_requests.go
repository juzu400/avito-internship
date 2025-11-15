package service

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/juzu400/avito-internship/internal/domain"
)

// Create creates a new pull request for the given author and automatically
// assigns reviewers from the author's team. At most two reviewers are assigned.
// If required fields are missing or business rules are violated, ErrValidation is returned.
func (s *PullRequestService) Create(
	ctx context.Context,
	id domain.PullRequestID,
	name string,
	authorID domain.UserID,
) (*domain.PullRequest, error) {
	if id == "" || name == "" || authorID == "" {
		err := fmt.Errorf("%w: missing fields (id/name/author)", domain.ErrValidation)
		s.log.Warn("validate Create failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty id/name/author_id"),
			slog.String("pull_request_id", string(id)),
			slog.String("author_id", string(authorID)),
		)
		return nil, err
	}

	s.log.Info("creating pull request",
		slog.String("pull_request_id", string(id)),
		slog.String("author_id", string(authorID)),
	)

	team, err := s.teams.GetByMemberID(ctx, authorID)
	if err != nil {
		s.log.Error("get team for author failed",
			slog.String("pull_request_id", string(id)),
			slog.String("author_id", string(authorID)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}

	reviewers := pickReviewersFromTeam(team, authorID, 2)

	pr := &domain.PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now().UTC(),
	}

	if err := s.prs.Create(ctx, pr); err != nil {
		s.log.Error("Create pull request failed",
			slog.String("pull_request_id", string(id)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}

	return pr, nil
}

// Merge marks a pull request as merged in an idempotent way.
// If the pull request is already merged, the existing state is returned without error.
// If the pull request does not exist, domain.ErrNotFound is returned.
func (s *PullRequestService) Merge(
	ctx context.Context,
	id domain.PullRequestID,
) (*domain.PullRequest, error) {
	if id == "" {
		err := fmt.Errorf("%w: pull_request_id is empty", domain.ErrValidation)
		s.log.Warn("validate Merge failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty pull_request_id"),
		)
		return nil, err
	}

	s.log.Info("merging pull request",
		slog.String("pull_request_id", string(id)),
	)

	now := time.Now().UTC()

	pr, err := s.prs.Merge(ctx, id, now)
	if err != nil {
		s.log.Error("Merge failed",
			slog.String("pull_request_id", string(id)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, err
	}
	if pr.IsMerged() {
		s.log.Info("pull request merged (idempotent)",
			slog.String("pull_request_id", string(id)),
		)
	} else {
		s.log.Warn("Merge: pull request not in MERGED status after merge",
			slog.String("pull_request_id", string(id)),
			slog.String("status", string(pr.Status)),
		)
	}

	return pr, nil
}

// ReassignReviewer replaces an existing reviewer of a pull request with another
// candidate from the same team. It skips inactive users, the author and already
// assigned reviewers. If the pull request is merged, has no such reviewer or
// there are no suitable candidates, a corresponding domain error is returned.
func (s *PullRequestService) ReassignReviewer(
	ctx context.Context,
	prID domain.PullRequestID,
	oldReviewerID domain.UserID,
) (*domain.PullRequest, *domain.User, error) {
	if prID == "" || oldReviewerID == "" {
		err := fmt.Errorf("%w: empty prID or oldReviewerID", domain.ErrValidation)
		s.log.Warn("validate ReassignReviewer failed",
			slog.String("error_code", ErrCodeValidation),
			slog.String("reason", "empty pull_request_id or old_user_id"),
		)
		return nil, nil, err
	}

	s.log.Info("reassigning reviewer",
		slog.String("pull_request_id", string(prID)),
		slog.String("old_reviewer_id", string(oldReviewerID)),
	)

	pr, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("GetByID in ReassignReviewer failed",
			slog.String("pull_request_id", string(prID)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, nil, err
	}

	if pr.IsMerged() {
		err := domain.ErrPullRequestAlreadyMerged
		s.log.Warn("ReassignReviewer on merged PR",
			slog.String("pull_request_id", string(prID)),
			slog.String("error_code", ErrCodePullRequestAlreadyMerged),
		)
		return nil, nil, err
	}

	foundIdx := -1
	for i, id := range pr.AssignedReviewers {
		if id == oldReviewerID {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		err := domain.ErrReviewerNotAssigned
		s.log.Warn("ReassignReviewer: old reviewer not assigned",
			slog.String("pull_request_id", string(prID)),
			slog.String("old_reviewer_id", string(oldReviewerID)),
			slog.String("error_code", ErrCodeReviewerNotAssigned),
		)
		return nil, nil, err
	}

	team, err := s.teams.GetByMemberID(ctx, oldReviewerID)
	if err != nil {
		s.log.Error("GetByMemberID in ReassignReviewer failed",
			slog.String("pull_request_id", string(prID)),
			slog.String("old_reviewer_id", string(oldReviewerID)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, nil, err
	}

	candidates := make([]domain.User, 0)
	for _, m := range team.Members {
		if !m.IsActive {
			continue
		}
		if m.ID == oldReviewerID {
			continue
		}
		if m.ID == pr.AuthorID {
			continue
		}
		if containsUserID(pr.AssignedReviewers, m.ID) {
			continue
		}
		candidates = append(candidates, m)
	}

	if len(candidates) == 0 {
		err := domain.ErrNoReviewerCandidates
		s.log.Warn("ReassignReviewer: no reviewer candidates",
			slog.String("pull_request_id", string(prID)),
			slog.String("old_reviewer_id", string(oldReviewerID)),
			slog.String("error_code", ErrCodeNoReviewerCandidates),
		)
		return nil, nil, err
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	newIdx := r.Intn(len(candidates))
	newReviewer := candidates[newIdx]

	pr.AssignedReviewers[foundIdx] = newReviewer.ID

	if err := s.prs.Update(ctx, pr); err != nil {
		s.log.Error("Update in ReassignReviewer failed",
			slog.String("pull_request_id", string(prID)),
			slog.String("error_code", ErrorCode(err)),
			slog.Any("err", err),
		)
		return nil, nil, err
	}

	return pr, &newReviewer, nil
}

// pickReviewersFromTeam selects up to maxCount active team members as reviewers,
// excluding the author. If there are fewer candidates than maxCount, all of them
// are returned. Selection is randomized to avoid always picking the same users.
func pickReviewersFromTeam(team *domain.Team, authorID domain.UserID, maxCount int) []domain.UserID {
	candidates := make([]domain.UserID, 0, len(team.Members))

	for _, m := range team.Members {
		if !m.IsActive {
			continue
		}
		if m.ID == authorID {
			continue
		}
		candidates = append(candidates, m.ID)
	}

	if len(candidates) == 0 {
		return nil
	}

	if len(candidates) <= maxCount {
		return candidates
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	return candidates[:maxCount]
}

// containsUserID reports whether the given user ID is present in the list.
func containsUserID(list []domain.UserID, id domain.UserID) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}
