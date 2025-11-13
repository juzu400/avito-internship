package domain

type UserID string

type User struct {
	ID       UserID
	Username string
	IsActive bool
}
