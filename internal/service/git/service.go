package git

// New returns a new Git service.
func New(repo Repository) *service {
	srv := new(service)
	srv.repo = repo

	return srv
}

type service struct {
	repo Repository
}
