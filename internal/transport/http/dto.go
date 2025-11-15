package http

import "time"

// ErrorBody represents a structured error payload returned in HTTP responses.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps error details under the "error" field.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// TeamMemberDTO represents a team member in HTTP requests and responses.
type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// TeamDTO represents a team with its members in HTTP responses.
type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

// TeamAddResponse is the response body for successful team creation or update.
type TeamAddResponse struct {
	Team TeamDTO `json:"team"`
}

// SetUserActiveRequest is the request body for toggling user activity.
type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// UserDTO represents a user together with their team in HTTP responses.
type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// UserResponse wraps a single user under the "user" field.
type UserResponse struct {
	User UserDTO `json:"user"`
}

// PullRequestShortDTO is a compact representation of a pull request
// used in lists, e.g. for user reviews.
type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

// GetUserReviewResponse is the response body for listing pull requests
// where a user is assigned as a reviewer.
type GetUserReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

// CreatePullRequestRequest is the request body for creating a new pull request.
type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// MergePullRequestRequest is the request body for merging a pull request.
type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// ReassignReviewerRequest is the request body for reassigning a reviewer
// on a pull request.
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// PullRequestDTO represents a detailed pull request in HTTP responses.
type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt"`
}

// PullRequestResponse wraps a single pull request under the "pr" field.
type PullRequestResponse struct {
	PR PullRequestDTO `json:"pr"`
}

// ReassignReviewerResponse is the response body for successful reviewer reassignment.
type ReassignReviewerResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

// ReviewerStatsItemDTO represents statistics for a single reviewer.
type ReviewerStatsItemDTO struct {
	ReviewerID  string `json:"reviewer_id"`
	Assignments int    `json:"assignments"`
}

// ReviewerStatsResponse is the response body for reviewer statistics.
type ReviewerStatsResponse struct {
	Items []ReviewerStatsItemDTO `json:"items"`
}

// PullRequestStatsItemDTO represents statistics for a single pull request.
type PullRequestStatsItemDTO struct {
	PullRequestID string `json:"pull_request_id"`
	Reviewers     int    `json:"reviewers"`
}

// PullRequestStatsResponse is the response body for pull request statistics.
type PullRequestStatsResponse struct {
	Items []PullRequestStatsItemDTO `json:"items"`
}
