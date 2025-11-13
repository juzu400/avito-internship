package domain

import "time"

type PullRequestID string

type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "OPEN"
	PRStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID                PullRequestID
	Name              string
	AuthorID          UserID
	Status            PullRequestStatus
	AssignedReviewers []UserID
	CreatedAt         time.Time
	MergedAt          *time.Time
}

func (p *PullRequest) IsMerged() bool {
	return p.Status == PRStatusMerged
}
