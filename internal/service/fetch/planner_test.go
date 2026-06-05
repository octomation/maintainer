package fetch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func newPlanner(t *testing.T, cnf *config.Fetch) *Planner {
	t.Helper()
	if cnf == nil {
		cnf = &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1}}
	}
	r, err := NewPathRenderer(cnf.Defaults.Root, home, cwd)
	require.NoError(t, err)
	return NewPlanner(cnf, NewPathResolver(cnf, r))
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
	require.Len(t, actions, 1)
	a := actions[0]
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
	require.Len(t, actions, 1)
	a := actions[0]
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
	require.Len(t, actions, 1)
	a := actions[0]
	assert.Equal(t, KindUpdateRemote, a.Kind)
	assert.Equal(t, "git@github.com:acme/service.git", a.RemoteURL)
}

func TestPlanner_Relocate(t *testing.T) {
	p := newPlanner(t, nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "service", Path: "/work/old/acme/service",
		RemoteURL: "git@github.com:acme/service.git", CloneURL: "ssh",
	})
	in := PlanInput{
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "service", Visibility: github.Public}},
		State:     st,
		Clones:    []DiskClone{{Path: "/work/public/acme/service", ID: 1}}, // moved manually
	}
	actions, err := p.Plan(in)
	require.NoError(t, err)
	require.Len(t, actions, 1)
	a := actions[0]
	assert.Equal(t, KindRelocate, a.Kind)
	assert.Equal(t, "/work/old/acme/service", a.FromPath)
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
