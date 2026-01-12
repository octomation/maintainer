package fetch

import (
	"context"
	"sync"
	"time"

	"go.octolab.org/toolset/maintainer/internal/config"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// AuthFunc resolves per-operation Git credentials for a profile/transport
// (§5.4). The service builds it from the resolved profile tokens.
type AuthFunc func(profile, transport string) gitsvc.Auth

// Applier is the only component with disk side effects. It executes one Action
// per call and updates the matching state record on success; the state write
// lock is never held across a network call (§9).
type Applier struct {
	git   gitsvc.GitSync
	auth  AuthFunc
	clock func() time.Time
	mu    sync.Mutex
}

// NewApplier wires an Applier from the Git port, an auth resolver and a clock.
func NewApplier(git gitsvc.GitSync, auth AuthFunc, clock func() time.Time) *Applier {
	if clock == nil {
		clock = time.Now
	}
	if auth == nil {
		auth = func(string, string) gitsvc.Auth { return gitsvc.Auth{} }
	}
	return &Applier{git: git, auth: auth, clock: clock}
}

// Execute performs one action's side effect (if any) and commits the state
// change. It is safe to call concurrently for distinct repositories.
func (a *Applier) Execute(ctx context.Context, act Action, st *state.State) error {
	now := a.clock()
	switch act.Kind {
	case KindAdopt:
		rec := *act.Record
		rec.FirstSeen, rec.LastSeen, rec.LastApply = now, now, now
		a.commit(func() { st.Upsert(rec) })

	case KindRelocate:
		a.commit(func() {
			if r, ok := st.ByID(act.ID); ok {
				r.Path = act.ToPath
				r.LastSeen, r.LastApply = now, now
			}
		})

	case KindUpdateRemote:
		if err := a.git.UpdateRemote(act.Path, act.RemoteURL); err != nil {
			return err
		}
		a.commit(func() {
			if r, ok := st.ByID(act.ID); ok {
				r.RemoteURL = act.RemoteURL
				r.OwnerLogin, r.Name = act.Owner, act.Name
				r.LastApply = now
			}
		})

	case KindMove:
		if err := a.git.Move(act.FromPath, act.ToPath); err != nil {
			return err
		}
		if act.UpdateRemote {
			if err := a.git.UpdateRemote(act.ToPath, act.RemoteURL); err != nil {
				return err
			}
		}
		a.commit(func() {
			if r, ok := st.ByID(act.ID); ok {
				r.Path = act.ToPath
				r.OwnerLogin, r.Name = act.Owner, act.Name
				if act.RemoteURL != "" {
					r.RemoteURL = act.RemoteURL
				}
				r.LastSeen, r.LastApply = now, now
			}
		})

	case KindClone:
		if err := a.git.Clone(ctx, gitsvc.CloneOptions{
			URL:  cloneURL(act),
			Path: act.Path,
			Auth: a.auth(act.Profile, act.Transport),
		}); err != nil {
			return err
		}
		rec := *act.Record
		if rec.FirstSeen.IsZero() {
			rec.FirstSeen = now
		}
		rec.LastSeen, rec.LastApply = now, now
		a.commit(func() { st.Upsert(rec) })

	case KindFetch:
		if err := a.git.Fetch(ctx, act.Path, a.auth(act.Profile, act.Transport)); err != nil {
			return err
		}
		a.commit(func() {
			if r, ok := st.ByID(act.ID); ok {
				r.LastSeen, r.LastApply = now, now
			}
		})

	default:
		// orphan / noop / conflict are report-only (§7.4).
	}
	return nil
}

func (a *Applier) commit(mutate func()) {
	a.mu.Lock()
	defer a.mu.Unlock()
	mutate()
}

// cloneURL picks the transport-appropriate clone URL from the snapshot; the
// PAT is never embedded (it is supplied via BasicAuth, §5.4).
func cloneURL(act Action) string {
	if act.Snapshot != nil {
		if act.Transport == config.TransportSSH && act.Snapshot.SSHCloneURL != "" {
			return act.Snapshot.SSHCloneURL
		}
		if act.Snapshot.HTTPSCloneURL != "" {
			return act.Snapshot.HTTPSCloneURL
		}
	}
	return act.RemoteURL
}
