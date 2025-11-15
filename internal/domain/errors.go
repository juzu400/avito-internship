package domain

import "errors"

// Common domain-level errors. They are used to distinguish between different
// business situations (validation issues, conflicts, missing resources, etc.),
// and are later mapped to HTTP error codes in the transport layer.
var (
	ErrNotFound                 = errors.New("resource not found")
	ErrPullRequestAlreadyMerged = errors.New("pull request already merged")
	ErrReviewerNotAssigned      = errors.New("reviewer not assigned to pull request")
	ErrNoReviewerCandidates     = errors.New("no reviewer candidates available")
	ErrPullRequestAlreadyExists = errors.New("pull request already exists")
	ErrTeamAlreadyExists        = errors.New("team already exists")
	ErrValidation               = errors.New("validation error")
)
