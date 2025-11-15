package domain

type UserID string

// User represents an application user that can be part of teams
// and participate in pull requests as an author or reviewer.
type User struct {
	ID       UserID
	Username string
	IsActive bool
}
