package domain

// ReviewerAssignmentStat represents statistics about the number of pull request
type ReviewerAssignmentStat struct {
	ReviewerID       UserID
	AssignmentsCount int
}

// PullRequestReviewersStat represents statistics about the number of reviewers
type PullRequestReviewersStat struct {
	PullRequestID  PullRequestID
	ReviewersCount int
}
