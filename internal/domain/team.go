package domain

// Team represents a logical group of users that can review pull requests together.
type Team struct {
	Name    string
	Members []User
}
