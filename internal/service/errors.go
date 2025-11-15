package service

import (
	"errors"

	"github.com/juzu400/avito-internship/internal/domain"
)

const (
	ErrCodeValidation               = "VALIDATION_ERROR"
	ErrCodeInternal                 = "INTERNAL_ERROR"
	ErrCodeNotFound                 = "NOT_FOUND"
	ErrCodeTeamAlreadyExists        = "TEAM_EXISTS"
	ErrCodePullRequestAlreadyExists = "PR_EXISTS"
	ErrCodePullRequestAlreadyMerged = "PR_MERGED"
	ErrCodeReviewerNotAssigned      = "NOT_ASSIGNED"
	ErrCodeNoReviewerCandidates     = "NO_CANDIDATE"
)

func ErrorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return ErrCodeNotFound
	case errors.Is(err, domain.ErrPullRequestAlreadyMerged):
		return ErrCodePullRequestAlreadyMerged
	case errors.Is(err, domain.ErrReviewerNotAssigned):
		return ErrCodeReviewerNotAssigned
	case errors.Is(err, domain.ErrNoReviewerCandidates):
		return ErrCodeNoReviewerCandidates
	case errors.Is(err, domain.ErrPullRequestAlreadyExists):
		return ErrCodePullRequestAlreadyExists
	case errors.Is(err, domain.ErrTeamAlreadyExists):
		return ErrCodeTeamAlreadyExists
	case errors.Is(err, domain.ErrValidation):
		return ErrCodeValidation
	default:
		return ErrCodeInternal
	}
}
