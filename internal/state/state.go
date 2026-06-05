// Package state owns the local reconciliation state file: an on-disk JSON
// document that remembers what `maintainer fetch` has materialised, keyed by
// the stable GitHub repo id (fetch plan §6).
package state

import "time"

// Version is the only schema version supported in the PoC (§6.4).
const Version = 1

// State is the whole state document (§6.2 top-level fields).
type State struct {
	Version  int                  `json:"version"`
	Profiles map[string]time.Time `json:"profiles"`
	Repos    []Record             `json:"repos"`
}

// Record is one tracked repository (§6.2). The id is the primary key; every
// other field is a last-observed value.
type Record struct {
	ID               int64     `json:"id"`
	NodeID           string    `json:"node_id"`
	OwnerLogin       string    `json:"owner_login"`
	Name             string    `json:"name"`
	Visibility       string    `json:"visibility"`
	Path             string    `json:"path"`
	RemoteURL        string    `json:"remote_url"`
	CloneURL         string    `json:"clone_url"`
	SourceProfile    string    `json:"source_profile"`
	DefaultBranch    string    `json:"default_branch"`
	IsFork           bool      `json:"is_fork"`
	IsTemplate       bool      `json:"is_template"`
	ArchivedOnGitHub bool      `json:"archived_on_github"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
	LastApply        time.Time `json:"last_apply"`
}

// New returns an empty state at the current schema version.
func New() *State {
	return &State{Version: Version, Profiles: map[string]time.Time{}}
}

// ByID returns a pointer to the record with the given id for in-place mutation.
func (s *State) ByID(id int64) (*Record, bool) {
	for i := range s.Repos {
		if s.Repos[i].ID == id {
			return &s.Repos[i], true
		}
	}
	return nil, false
}

// Upsert inserts or replaces the record with the matching id.
func (s *State) Upsert(rec Record) {
	if existing, ok := s.ByID(rec.ID); ok {
		*existing = rec
		return
	}
	s.Repos = append(s.Repos, rec)
}

// Remove drops the record with the given id; it reports whether one was found.
func (s *State) Remove(id int64) bool {
	for i := range s.Repos {
		if s.Repos[i].ID == id {
			s.Repos = append(s.Repos[:i], s.Repos[i+1:]...)
			return true
		}
	}
	return false
}
