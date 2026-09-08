package fetch

import (
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// PinnedPath resolves an explicit repo path, including an old-name rule during
// a rename. Persisted pins keep name-based rules safe on subsequent runs.
func (pr *PathResolver) PinnedPath(snap github.RepoSnapshot, rec *state.Record) (string, error) {
	rule, ok := pr.cnf.RepoOverride(snap.ID, snap.Owner, snap.Name)
	if !ok && rec != nil {
		rule, ok = pr.cnf.RepoOverride(rec.ID, rec.OwnerLogin, rec.Name)
	}
	if ok && rule.Path != "" {
		return pr.renderer.Render(rule.Path, snap, true)
	}
	if rec != nil {
		if rec.PinSource != "workspace" && rec.PinnedPath != "" {
			return rec.PinnedPath, nil
		}
		if pr.scope != nil && pr.scope.Pinned(rec.Path) {
			return rec.Path, nil
		}
	}
	return "", nil
}

// Authorised refuses to turn remembered paths into additional scan roots.
func (pr *PathResolver) Authorised(rec state.Record, snap github.RepoSnapshot) bool {
	if snap.ID == 0 {
		snap = github.RepoSnapshot{ID: rec.ID, Owner: rec.OwnerLogin, Name: rec.Name}
	}
	if pin, err := pr.PinnedPath(snap, &rec); err == nil && pin != "" {
		return true
	}
	if rec.PinSource == "workspace" {
		return false
	} // removing a pin never enables moves
	return pr.scope != nil && pr.scope.Contains(rec.Path)
}

func (pr *PathResolver) WorkspacePin(snap github.RepoSnapshot, rec *state.Record) bool {
	rule, ok := pr.cnf.RepoOverride(snap.ID, snap.Owner, snap.Name)
	if !ok && rec != nil {
		rule, ok = pr.cnf.RepoOverride(rec.ID, rec.OwnerLogin, rec.Name)
	}
	if ok && rule.Path != "" {
		return false
	}
	return rec != nil && (rec.PinSource == "workspace" || rec.PinnedPath == "") && pr.scope.Pinned(rec.Path)
}

// An explicit path selects the active clone even when an old duplicate exists.
// It never moves or creates directories: changing the pin selects an existing,
// identity-verified checkout, or reports a conflict.
func planPinned(base Action, target string, rec *state.Record, clones []DiskClone) *Action {
	base.Path = target
	c, found := cloneAt(clones, target)
	if !found || c.Origins > 1 {
		base.Kind = KindConflict
		base.Reason = "pinned path is missing or is not an unambiguous clone of this repository"
		return &base
	}
	if c.Transport != "" {
		base.Transport = c.Transport
	}
	base.RemoteURL = canonicalURL(base.Transport, base.Owner, base.Name)
	base.Record = recordFromSnapshot(*base.Snapshot, target, base.RemoteURL, base.Transport)
	base.Record.PinnedPath = target
	if rec != nil {
		base.Record.FirstSeen = rec.FirstSeen
	}
	base.UpdateRemote = c.RemoteURL != "" && c.RemoteURL != base.RemoteURL
	switch {
	case rec == nil || rec.Path != target:
		base.Kind = KindAdopt
	case base.UpdateRemote:
		base.Kind = KindUpdateRemote
	default:
		base.Kind = KindFetch
	}
	return &base
}
