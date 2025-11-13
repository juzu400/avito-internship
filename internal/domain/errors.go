package domain

import "errors"

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrTeamNotFound             = errors.New("team not found")
	ErrPullRequestNotFound      = errors.New("pull request not found")
	ErrPullRequestAlreadyMerged = errors.New("pull request already merged")
	ErrReviewerNotAssigned      = errors.New("reviewer not assigned to pull request")
	ErrNoReviewerCandidates     = errors.New("no reviewer candidates available")
	ErrValidation               = errors.New("validation error")
)
