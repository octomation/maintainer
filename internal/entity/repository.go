package entity

// Repository represents a GitHub repository.
type Repository struct {
	ID     int64
	URN    RepositoryURN
	Labels []Label
}
