package service

import (
	"errors"

	"github.com/juzu400/avito-internship/internal/domain"
)

const (
	ErrCodeValidation               = "VALIDATION_ERROR"
	ErrCodeUserNotFound             = "USER_NOT_FOUND"
	ErrCodeTeamNotFound             = "TEAM_NOT_FOUND"
	ErrCodePullRequestNotFound      = "PULL_REQUEST_NOT_FOUND"
	ErrCodePullRequestAlreadyMerged = "PULL_REQUEST_ALREADY_MERGED"
	ErrCodeReviewerNotAssigned      = "REVIEWER_NOT_ASSIGNED"
	ErrCodeNoReviewerCandidates     = "NO_CANDIDATE" // как в openapi
	ErrCodeInternal                 = "INTERNAL_ERROR"
)

func ErrorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return ErrCodeUserNotFound
	case errors.Is(err, domain.ErrTeamNotFound):
		return ErrCodeTeamNotFound
	case errors.Is(err, domain.ErrPullRequestNotFound):
		return ErrCodePullRequestNotFound
	case errors.Is(err, domain.ErrPullRequestAlreadyMerged):
		return ErrCodePullRequestAlreadyMerged
	case errors.Is(err, domain.ErrReviewerNotAssigned):
		return ErrCodeReviewerNotAssigned
	case errors.Is(err, domain.ErrNoReviewerCandidates):
		return ErrCodeNoReviewerCandidates
	case errors.Is(err, domain.ErrValidation):
		return ErrCodeValidation
	default:
		return ErrCodeInternal
	}
}
