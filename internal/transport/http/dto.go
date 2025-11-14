package http

import "time"

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type TeamAddResponse struct {
	Team TeamDTO `json:"team"`
}

type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User UserDTO `json:"user"`
}

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type GetUserReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt"`
}

type PullRequestResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type ReassignReviewerResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"` // user_id нового ревьювера
}
