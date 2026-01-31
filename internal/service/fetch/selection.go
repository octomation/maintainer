package fetch

import (
	"fmt"
	"strings"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// Selection narrows one invocation to the explicit --owner and --profile
// values (§5.1). Empty fields select everything: the configured
// include_owners narrow only the discovery of new repositories, and a full
// run still verifies and manages existing checkouts of other owners.
//
// A repository outside the selection gets no per-repository request
// (confirmation by id, resolution by name, clone, fetch), no mutation and no
// new observation; a bulk listing may still return it.
type Selection struct {
	Owners   []string
	Profiles []string
}

// Reasons of the report-only rows for repositories this invocation leaves
// alone.
const (
	reasonOutside     = "outside the selection"
	reasonTransferred = "transferred outside the selection to %s"
)

// owner reports whether a login is selected; logins compare
// case-insensitively, as GitHub does.
func (sel Selection) owner(login string) bool {
	if config.IsAllOwners(sel.Owners) {
		return true
	}
	for _, o := range sel.Owners {
		if strings.EqualFold(o, login) {
			return true
		}
	}
	return false
}

func (sel Selection) profile(name string) bool {
	if len(sel.Profiles) == 0 {
		return true
	}
	for _, p := range sel.Profiles {
		if p == name {
			return true
		}
	}
	return false
}

// deselect returns the tracked repositories outside the selection: neither
// listed by a selected profile nor of a selected owner and profile.
func (s *Service) deselect(st *state.State, listed []github.RepoSnapshot) map[int64]string {
	out := make(map[int64]string)
	present := make(map[int64]bool, len(listed))
	for _, snap := range listed {
		present[snap.ID] = true
	}
	for _, rec := range st.Repos {
		if present[rec.ID] {
			continue // the listing of a selected profile selected it
		}
		if !s.deps.Selection.owner(rec.OwnerLogin) || !s.deps.Selection.profile(rec.SourceProfile) {
			out[rec.ID] = reasonOutside
		}
	}
	return out
}

// selectedNames tells the name resolver which checkout origins it may resolve:
// those of a selected owner or of a selected record, and never the recorded
// remote of a repository outside the selection. Nil means all of them.
func (s *Service) selectedNames(st *state.State, deselected map[int64]string) func(owner, name string) bool {
	sel := s.deps.Selection
	if len(sel.Owners) == 0 && len(sel.Profiles) == 0 {
		return nil
	}
	outside := make(map[string]bool)
	selected := make(map[string]bool)
	for _, rec := range st.Repos {
		set := selected
		if _, ok := deselected[rec.ID]; ok {
			set = outside
		}
		for _, name := range recordedNames(rec) {
			set[name] = true
		}
	}
	return func(owner, name string) bool {
		key := strings.ToLower(owner + "/" + name)
		return !outside[key] && (sel.owner(owner) || selected[key])
	}
}

// recordedNames are the lower-cased owner/name spellings a record is known by:
// its GitHub name and the origin it recorded.
func recordedNames(rec state.Record) []string {
	names := []string{strings.ToLower(rec.OwnerLogin + "/" + rec.Name)}
	if _, owner, name, _, ok := parseRemote(rec.RemoteURL); ok {
		names = append(names, strings.ToLower(owner+"/"+name))
	}
	return names
}

// transferredOut names the new owner of a selected record that a confirmation
// found outside the selected owners.
func transferredOut(snap github.RepoSnapshot) string {
	return fmt.Sprintf(reasonTransferred, snap.FullName())
}
