package fetch_test

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"testing"

	gh "github.com/google/go-github/v91/github"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	statussvc "go.octolab.org/toolset/maintainer/internal/service/status"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func TestOrphanSurvivesAdoptWithoutMutation(t *testing.T) {
	root := t.TempDir()
	old, pin := filepath.Join(t.TempDir(), "old"), filepath.Join(t.TempDir(), "active")
	for _, path := range []string{old, pin} {
		makeClone(t, path, "git@github.com:acme/tool.git")
	}
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: old, FirstSeen: clock()})
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public}
	g := new(fakeGit)
	for _, expected := range []Kind{KindAdopt, KindFetch} {
		// The old external path is a known diagnostic, never an adoption candidate.
		acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, KnownCheckouts: map[string]bool{old: true}, Clones: []DiskClone{{ID: 1, Path: pin, RemoteURL: "git@github.com:acme/tool.git", Transport: "ssh"}}})
		require.NoError(t, err)
		require.Len(t, acts, 2)
		assert.Equal(t, expected, acts[0].Kind)
		assert.Equal(t, KindOrphan, acts[1].Kind)
		assert.Equal(t, "duplicate-pin", acts[1].OrphanReason)
		assert.Equal(t, old, acts[1].Path)
		assert.Equal(t, pin, acts[1].ActivePath)
		apply := NewApplier(g, nil, clock)
		for _, act := range acts {
			require.NoError(t, apply.Execute(context.Background(), act, st))
		}
		assert.Equal(t, []string{old}, st.Repos[0].PreviousPaths)
	}
	assert.Empty(t, g.cloned)
	assert.Empty(t, g.moved)
	assert.Empty(t, g.updated)
	assert.Equal(t, []string{pin}, g.fetched)
	rows, err := statussvc.Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	var orphans int
	for _, row := range rows {
		if row.OrphanReason != "" {
			orphans++
			assert.Equal(t, old, row.Path)
			assert.Equal(t, pin, row.ActivePath)
			assert.Empty(t, row.Error)
		}
	}
	assert.Equal(t, 1, orphans)
	// Historical paths and orphan metadata round-trip in the same schema.
	store := state.NewStore(afero.NewMemMapFs(), "/state.json", nil)
	require.NoError(t, store.Save(st))
	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, st, loaded)
}

type absentRepo struct{}

type noNetworkFetch struct{ *gitsvc.Sync }

func (noNetworkFetch) Fetch(context.Context, string, gitsvc.Auth) error { return nil }

func (absentRepo) ConfirmByID(context.Context, github.Profile, int64) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{}, &gh.ErrorResponse{Response: &http.Response{StatusCode: 404}, Message: "Not Found"}
}
func (absentRepo) ResolveByName(context.Context, github.Profile, string, string) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{}, &gh.ErrorResponse{Response: &http.Response{StatusCode: 404}, Message: "Not Found"}
}

func TestRemoteOrphanOnlyCachedOnApply(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "public/acme/tool")
	makeClone(t, path, "git@github.com:acme/tool.git")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{Concurrency: 1}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: path, SourceProfile: "p"})
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	require.NoError(t, store.Save(st))
	var out bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p"}}, root, root, 1, Deps{Store: store, Discoverer: fakeDiscoverer{}, Resolver: absentRepo{}, Confirmer: absentRepo{}, GitSync: gitsvc.NewSync(), Clock: clock, Reporter: NewReporter(&out, &out, FormatJSON, 0, false)})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	assert.Contains(t, out.String(), `"orphan_reason": "remote-gone"`)
	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded.Repos[0].RemoteStatus)
	require.NoError(t, svc.Run(context.Background(), true))
	loaded, err = store.Load()
	require.NoError(t, err)
	assert.Equal(t, "gone", loaded.Repos[0].RemoteStatus)
	require.NotNil(t, loaded.Repos[0].RemoteCheckedAt)
	assert.Equal(t, clock(), *loaded.Repos[0].RemoteCheckedAt)
	rows, err := statussvc.Collect(context.Background(), cnf, loaded, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "remote-gone", rows[0].OrphanReason)
	assert.Empty(t, rows[0].Error)
	assert.Equal(t, "unborn", rows[0].Status)
	// A later positive observation clears the warning only when saved by apply.
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public, SourceProfile: "p"}
	restored, err := NewService(cnf, []ResolvedProfile{{Name: "p"}}, root, root, 1, Deps{Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: renamedRepository{snapshot: snap}, GitSync: noNetworkFetch{gitsvc.NewSync()}, Clock: clock, Reporter: NewReporter(&out, &out, FormatJSON, 0, false)})
	require.NoError(t, err)
	require.NoError(t, restored.Run(context.Background(), false))
	loaded, err = store.Load()
	require.NoError(t, err)
	assert.Equal(t, "gone", loaded.Repos[0].RemoteStatus)
	require.NoError(t, restored.Run(context.Background(), true))
	loaded, err = store.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded.Repos[0].RemoteStatus)
}

func TestOrphanReporterCountsIdentitiesAndShowsPaths(t *testing.T) {
	plan := Plan{Actions: []Action{
		{Kind: KindFetch, ID: 1, Owner: "acme", Name: "tool", Path: "/active"},
		{Kind: KindOrphan, ID: 1, Owner: "acme", Name: "tool", Path: "/old", ActivePath: "/active", OrphanReason: "duplicate-pin", Reason: "inactive checkout; active pin: /active"},
	}}
	var out bytes.Buffer
	require.NoError(t, NewReporter(&out, &out, FormatHuman, 0, false).Render(plan, false))
	assert.Contains(t, out.String(), "plan: 1 repos total")
	assert.Contains(t, out.String(), "[duplicate-pin]")
	assert.Contains(t, out.String(), "at /old")
	assert.Contains(t, out.String(), "/active")
	assert.Equal(t, 1, plan.Summary().Orphan)
}

func TestNoConfirmationIsNotRemoteDeletion(t *testing.T) {
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: "/work/public/acme/tool"})
	acts, err := newPlanner(t, nil).Plan(PlanInput{State: st})
	require.NoError(t, err)
	assert.Empty(t, acts)
}
