package github

import "time"

// WithClock replaces the wall clock of the service; tests only.
func (srv *Service) WithClock(now func() time.Time) *Service {
	srv.now = now
	return srv
}
