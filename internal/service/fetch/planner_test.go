package fetch_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func newPlanner(t *testing.T, cnf *config.Fetch) *Planner {
	t.Helper()
	if cnf == nil {
		cnf = &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1}}
	}
	r, err := NewPathRenderer(cnf.WorkspaceConfig().Root, home, cwd)
	require.NoError(t, err)
	return NewPlanner(cnf, NewPathResolver(cnf, r))
}

// reconciled checks that a plan holds one reconciliation followed by the
// fetch that refreshes the same checkout, and returns the reconciliation.
func reconciled(t *testing.T, actions []Action) Action {
	t.Helper()
	require.Len(t, actions, 2)
	a, refresh := actions[0], actions[1]
	assert.Equal(t, KindFetch, refresh.Kind)
	assert.Equal(t, a.Kind, refresh.After)
	assert.Equal(t, a.ID, refresh.ID)
	assert.Equal(t, a.Path, refresh.Path)
	return a
}

func byID(actions []Action) map[int64]Action {
	m := make(map[int64]Action, len(actions))
	for _, a := range actions {
		m[a.ID] = a
	}
	return m
}

func TestPlanner_Clone(t *testing.T) {
	p := newPlanner(t, nil)
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     state.New(),
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	require.Len(t, actions, 1)
	a := actions[0]
	assert.Equal(t, KindClone, a.Kind)
	assert.Equal(t, "/work/public/acme/service", a.Path)
	assert.Equal(t, "git@github.com:acme/service.git", a.RemoteURL)
	assert.NotNil(t, a.Record)
}

func TestPlanner_Fetch(t *testing.T) {
	p := newPlanner(t, nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "service", Path: "/work/public/acme/service",
		RemoteURL: "git@github.com:acme/service.git", CloneURL: "ssh",
	})
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     st,
		Clones:    []DiskClone{{Path: "/work/public/acme/service", ID: 1}},
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	require.Len(t, actions, 1)
	assert.Equal(t, KindFetch, actions[0].Kind)
}

func TestPlanner_TrackedMissingCloneConflictsWhenTargetExists(t *testing.T) {
	p := newPlanner(t, nil)
	const path = "/work/public/acme/service"
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "service", Path: path,
		RemoteURL: "git@github.com:acme/service.git", CloneURL: "ssh",
	})
	actions, err := p.Plan(PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     st,
		Occupancy: map[string]Occupancy{path: OccupancyForeign},
	})
	require.NoError(t, err)
	require.Len(t, actions, 1)
	assert.Equal(t, KindConflict, actions[0].Kind)
	assert.Contains(t, actions[0].Reason, "refusing to overwrite")
}

func TestPlanner_UsesObservedFetchRemoteForDrift(t *testing.T) {
	p := newPlanner(t, nil)
	const path = "/work/public/acme/service"
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "service", Path: path,
		RemoteURL: "git@github.com:acme/service.git", CloneURL: "ssh",
	})
	actions, err := p.Plan(PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     st,
		Clones: []DiskClone{{
			Path: path, ID: 1, Transport: "ssh", RemoteURL: "git@github.com:acme/old-name.git",
		}},
	})
	require.NoError(t, err)
	a := reconciled(t, actions)
	assert.Equal(t, KindUpdateRemote, a.Kind)
	assert.Equal(t, "git@github.com:acme/service.git", a.RemoteURL)
}

func TestPlanner_MoveOnRename(t *testing.T) {
	p := newPlanner(t, nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "dotfiles", Path: "/work/public/acme/dotfiles",
		RemoteURL: "git@github.com:acme/dotfiles.git", CloneURL: "ssh",
	})
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "configs", Visibility: github.Public}},
		State:     st,
		Clones:    []DiskClone{{Path: "/work/public/acme/dotfiles", ID: 1}},
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	a := reconciled(t, actions)
	assert.Equal(t, KindMove, a.Kind)
	assert.Equal(t, "/work/public/acme/dotfiles", a.FromPath)
	assert.Equal(t, "/work/public/acme/configs", a.ToPath)
	assert.True(t, a.UpdateRemote)
	assert.Equal(t, "git@github.com:acme/configs.git", a.RemoteURL)
}

func TestPlanner_Adopt(t *testing.T) {
	p := newPlanner(t, nil)
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 7, Owner: "acme", Name: "tool", Visibility: github.Public}},
		State:     state.New(),
		Clones: []DiskClone{{
			Path: "/work/public/acme/tool", ID: 7,
			RemoteURL: "git@github.com:acme/tool.git", Transport: "ssh", Owner: "acme", Name: "tool",
		}},
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	a := reconciled(t, actions)
	assert.Equal(t, KindAdopt, a.Kind)
	assert.Equal(t, "/work/public/acme/tool", a.Path)
	require.NotNil(t, a.Record)
	assert.Equal(t, "git@github.com:acme/tool.git", a.Record.RemoteURL)
}

func TestPlanner_UpdateRemoteOnly(t *testing.T) {
	p := newPlanner(t, nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "service", Path: "/work/public/acme/service",
		RemoteURL: "https://github.com/acme/service.git", CloneURL: "ssh", // stale remote vs ssh transport
	})
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     st,
		Clones:    []DiskClone{{Path: "/work/public/acme/service", ID: 1}},
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	a := reconciled(t, actions)
	assert.Equal(t, KindUpdateRemote, a.Kind)
	assert.Equal(t, "git@github.com:acme/service.git", a.RemoteURL)
}

func TestPlanner_Relocate(t *testing.T) {
	p := newPlanner(t, nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "service", Path: "/work/private/acme/service",
		RemoteURL: "git@github.com:acme/service.git", CloneURL: "ssh",
	})
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     st,
		Clones:    []DiskClone{{Path: "/work/public/acme/service", ID: 1}}, // moved manually
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	a := reconciled(t, actions)
	assert.Equal(t, KindRelocate, a.Kind)
	assert.Equal(t, "/work/private/acme/service", a.FromPath)
	assert.Equal(t, "/work/public/acme/service", a.ToPath)
}

func TestPlanner_OrphanAndConfirmations(t *testing.T) {
	p := newPlanner(t, nil)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "gone", Path: "/work/public/acme/gone"})
	st.Upsert(state.Record{ID: 2, OwnerLogin: "acme", Name: "hidden", Path: "/work/public/acme/hidden"})
	st.Upsert(state.Record{ID: 3, OwnerLogin: "acme", Name: "locked", Path: "/work/public/acme/locked"})

	in := PlanInput{
		State: st,
		Confirmations: map[int64]Confirmation{
			1: {Status: ConfirmGone},
			2: {Status: ConfirmInaccessible},
			3: {Status: ConfirmFound},
		},
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	m := byID(actions)
	assert.Equal(t, KindOrphan, m[1].Kind)
	assert.Equal(t, KindNoop, m[2].Kind)
	assert.Equal(t, FlagInaccessible, m[2].Flag)
	assert.Equal(t, KindNoop, m[3].Kind)
}

func TestPlanner_TransientLeavesNoAction(t *testing.T) {
	p := newPlanner(t, nil)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "flaky", Path: "/work/public/acme/flaky"})
	in := PlanInput{State: st, Confirmations: map[int64]Confirmation{1: {Status: ConfirmTransient}}}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	assert.Empty(t, actions)
}

func TestPlanner_TargetCollision(t *testing.T) {
	// Two distinct ids whose per-repo overrides resolve to the same path.
	cnf := &config.Fetch{
		Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1},
		Repos: []config.Repo{
			{Match: config.RepoMatch{ID: 1}, Path: "/shared/spot"},
			{Match: config.RepoMatch{ID: 2}, Path: "/shared/spot"},
		},
	}
	p := newPlanner(t, cnf)
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{
			{ID: 1, Owner: "acme", Name: "a", Visibility: github.Public},
			{ID: 2, Owner: "acme", Name: "b", Visibility: github.Public},
		},
		State: state.New(),
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	m := byID(actions)
	assert.Equal(t, KindConflict, m[1].Kind)
	assert.Equal(t, KindConflict, m[2].Kind)
}

func TestPlanner_IgnoreSuppresses(t *testing.T) {
	cnf := &config.Fetch{
		Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1},
		Repos:    []config.Repo{{Match: config.RepoMatch{ID: 1}, Ignore: true}},
	}
	p := newPlanner(t, cnf)
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     state.New(),
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	assert.Empty(t, actions)
}

func TestPlanner_FilterGatesNewCloneNotTracked(t *testing.T) {
	cnf := &config.Fetch{
		Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1},
		Filters:  config.Filters{ExcludeArchived: true},
	}
	p := newPlanner(t, cnf)

	// A new archived repo is skipped (no clone).
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "old", Visibility: github.Public, IsArchived: true}},
		State:     state.New(),
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	assert.Empty(t, actions)

	// An already-tracked archived repo stays tracked, flagged filtered.
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "old", Path: "/work/public/acme/old",
		RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh",
	})
	in.State = st
	in.Clones = []DiskClone{{Path: "/work/public/acme/old", ID: 1}}
	actions, err = p.Plan(in)
	require.NoError(t, err)
	require.Len(t, actions, 1)
	assert.Equal(t, KindFetch, actions[0].Kind)
	assert.True(t, actions[0].Filtered)
}

// Filters gate adoption as they gate clones, on every path that can adopt an
// untracked checkout; tracked repositories stay tracked.
func TestPlanner_FilterGatesAdoption(t *testing.T) {
	fork := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "fork", Visibility: github.Public, IsFork: true}
	for name, tc := range map[string]struct {
		path, pinned, source string
		setup                func(*config.Fetch)
	}{
		"workspace checkout": {path: "/work/public/acme/fork"},
		"per-repository pin": {path: "/elsewhere/fork", pinned: "/elsewhere/fork", setup: func(c *config.Fetch) {
			c.Repos = []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: "/elsewhere/fork"}}
		}},
		"workspace pin": {path: "/work/prototyping/fork", pinned: "/work/prototyping/fork", source: "workspace", setup: func(c *config.Fetch) {
			c.Workspace.Pins = []string{"prototyping"}
		}},
	} {
		t.Run(name, func(t *testing.T) {
			cnf := &config.Fetch{
				Workspace: &config.Workspace{Root: "/work"},
				Defaults:  config.Defaults{CloneURL: "ssh", Concurrency: 1},
				Filters:   config.Filters{ExcludeForks: true},
			}
			if tc.setup != nil {
				tc.setup(cnf)
			}
			clone := DiskClone{ID: 1, Path: tc.path, Owner: "acme", Name: "fork", RemoteURL: "git@github.com:acme/fork.git", Transport: "ssh"}
			actions, err := newPlanner(t, cnf).Plan(PlanInput{Snapshots: []github.RepoSnapshot{fork}, State: state.New(), Clones: []DiskClone{clone}})
			require.NoError(t, err)
			require.Len(t, actions, 1)
			assert.Equal(t, KindNoop, actions[0].Kind)
			assert.Equal(t, tc.path, actions[0].Path)
			assert.True(t, actions[0].Filtered)
			assert.Contains(t, actions[0].Reason, "not adopted")

			st := state.New()
			st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "fork", Path: tc.path, PinnedPath: tc.pinned, PinSource: tc.source,
				RemoteURL: "git@github.com:acme/fork.git", CloneURL: "ssh", IsFork: true})
			actions, err = newPlanner(t, cnf).Plan(PlanInput{Snapshots: []github.RepoSnapshot{fork}, State: st, Clones: []DiskClone{clone}})
			require.NoError(t, err)
			require.Len(t, actions, 1)
			assert.Equal(t, KindFetch, actions[0].Kind, "a tracked fork stays tracked")
		})
	}
}

// An owner override applies whatever the letter case of its name, and its
// layout still contains the path rendered with GitHub's spelling.
func TestPlanner_OwnerOverrideIgnoresCase(t *testing.T) {
	root := t.TempDir()
	cnf := &config.Fetch{
		Workspace: &config.Workspace{Root: root},
		Defaults:  config.Defaults{CloneURL: "ssh", Concurrency: 1},
		Owners:    []config.Owner{{Name: "ACME", Path: "mirror/{{.Owner}}/{{.Repo}}", CloneURL: "https"}},
	}
	snaps := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public},
		{ID: 2, Owner: "acme", Name: "kept", Visibility: github.Public},
	}
	kept := filepath.Join(root, "mirror", "acme", "kept")
	makeClone(t, kept, "https://github.com/acme/kept.git")
	clones, err := NewAdopter(gitsvc.NewSync(), nil).Scan(context.Background(), root, snaps, cnf)
	require.NoError(t, err)
	require.Len(t, clones, 1, "the scan covers the override layout")

	st := state.New()
	st.Upsert(state.Record{ID: 2, OwnerLogin: "acme", Name: "kept", Path: kept, RemoteURL: "https://github.com/acme/kept.git", CloneURL: "https"})
	actions, err := newPlanner(t, cnf).Plan(PlanInput{Snapshots: snaps, State: st, Clones: clones})
	require.NoError(t, err)
	got := byID(actions)
	assert.Equal(t, KindClone, got[1].Kind)
	assert.Equal(t, filepath.Join(root, "mirror", "acme", "tool"), got[1].Path)
	assert.Equal(t, "https://github.com/acme/tool.git", got[1].RemoteURL)
	assert.Equal(t, KindFetch, got[2].Kind)
}
