package domain

import "errors"

var (
	ErrNotFound                 = errors.New("resource not found")
	ErrPullRequestAlreadyMerged = errors.New("pull request already merged")
	ErrReviewerNotAssigned      = errors.New("reviewer not assigned to pull request")
	ErrNoReviewerCandidates     = errors.New("no reviewer candidates available")
	ErrPullRequestAlreadyExists = errors.New("pull request already exists")
	ErrTeamAlreadyExists        = errors.New("team already exists")
	ErrValidation               = errors.New("validation error")
)
