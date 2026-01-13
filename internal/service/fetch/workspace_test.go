package fetch_test

import (
	"bytes"
	"context"
	"path/filepath"
	"sort"
	"testing"

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

func TestWorkspacePinLifecycle(t *testing.T) {
	const path = "/work/prototyping/deep/tool"
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: "/work", Pins: []string{"prototyping"}}, Defaults: config.Defaults{CloneURL: "ssh"}}
	st := state.New()
	snap := github.RepoSnapshot{ID: 1, Owner: "alice", Name: "tool", Visibility: github.Public}
	clone := DiskClone{ID: 1, Path: path, RemoteURL: "git@github.com:alice/tool.git", Transport: "ssh", Pinned: true}
	g := new(fakeGit)
	apply := NewApplier(g, nil, clock)
	for i, kind := range []Kind{KindAdopt, KindFetch, KindUpdateRemote, KindFetch} {
		if i == 2 {
			snap.Owner, snap.Name = "bob", "renamed"
		}
		if i == 3 {
			snap.Visibility = github.Private
			clone.RemoteURL = "git@github.com:bob/renamed.git"
		}
		acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: []DiskClone{clone}})
		require.NoError(t, err)
		require.Len(t, acts, 1)
		assert.Equal(t, kind, acts[0].Kind)
		assert.Equal(t, path, acts[0].Path)
		require.NoError(t, apply.Execute(context.Background(), acts[0], st))
	}
	rec, ok := st.ByID(1)
	require.True(t, ok)
	assert.Equal(t, "workspace", rec.PinSource)
	assert.Equal(t, path, rec.PinnedPath)
	assert.Equal(t, "private", rec.Visibility)
	assert.Empty(t, g.moved)
	assert.Empty(t, g.cloned)
	// Removing the broad pin suspends it, including when a new layout would
	// otherwise match the remembered location. No automatic unpin/move.
	cnf.Workspace.Pins = nil
	cnf.Workspace.Path = "prototyping/{{.Owner}}/{{.Repo}}"
	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: []DiskClone{clone}})
	require.NoError(t, err)
	require.Len(t, acts, 1)
	assert.Equal(t, KindOrphan, acts[0].Kind)
	assert.Equal(t, "out-of-scope", acts[0].OrphanReason)
}

func TestWorkspacePinsDoNotChooseDuplicates(t *testing.T) {
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: "/work", Pins: []string{"prototyping"}}}
	snap := github.RepoSnapshot{ID: 1, Owner: "alice", Name: "tool", Visibility: github.Public}
	clones := []DiskClone{{ID: 1, Path: "/work/prototyping/tool"}, {ID: 1, Path: "/work/public/alice/tool"}}
	for _, tracked := range []bool{false, true} {
		st := state.New()
		if tracked {
			st.Upsert(state.Record{ID: 1, OwnerLogin: "alice", Name: "tool", Path: clones[0].Path, PinnedPath: clones[0].Path, PinSource: "workspace"})
		}
		acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: clones})
		require.NoError(t, err)
		require.Len(t, acts, 1)
		assert.Equal(t, KindConflict, acts[0].Kind)
	}
	cnf.Repos = []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: "prototyping/tool"}}
	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: state.New(), Snapshots: []github.RepoSnapshot{snap}, Clones: clones})
	require.NoError(t, err)
	require.Len(t, acts, 2)
	assert.Equal(t, KindOrphan, acts[1].Kind)
	assert.Equal(t, KindAdopt, acts[0].Kind)
	assert.Empty(t, acts[0].Record.PinSource, "a specific repo pin wins over a broad workspace pin")
}

func TestWorkspaceOldStateIsNotAuthority(t *testing.T) {
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "alice", Name: "tool", Path: "/work/research/alice/tool"})
	for _, snap := range []github.RepoSnapshot{
		{ID: 1, Owner: "alice", Name: "tool", Visibility: github.Public},
		{ID: 1, Owner: "bob", Name: "renamed", Visibility: github.Private},
	} {
		acts, err := newPlanner(t, nil).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}})
		require.NoError(t, err)
		require.Len(t, acts, 1)
		assert.Equal(t, KindOrphan, acts[0].Kind)
		assert.Equal(t, "out-of-scope", acts[0].OrphanReason)
	}
}

func TestWorkspacePinsNeverMaterialiseTargets(t *testing.T) {
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: "/work", Pins: []string{"public"}}}
	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: state.New(), Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "alice", Name: "new", Visibility: github.Public}}})
	require.NoError(t, err)
	require.Len(t, acts, 1)
	assert.Equal(t, KindConflict, acts[0].Kind)
	assert.Contains(t, acts[0].Reason, "only existing pinned")
}

func TestWorkspaceTargetMustMatchDiscoveryLayout(t *testing.T) {
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: "/work", Path: "{{.DefaultBranch}}/{{.Owner}}/{{.Repo}}"}}
	_, err := newPlanner(t, cnf).Plan(PlanInput{State: state.New(), Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "alice", Name: "tool", DefaultBranch: "feature/work"}}})
	assert.ErrorContains(t, err, "outside workspace layout")
}

func TestWorkspaceFetchStatusSameLocalScope(t *testing.T) {
	root := t.TempDir()
	makeClone(t, filepath.Join(root, "public/alice/tool"), "git@github.com:alice/tool.git")
	makeClone(t, filepath.Join(root, "prototyping/deep/local"), "git@github.com:bob/local.git")
	makeClone(t, filepath.Join(root, "research/outsider/research"), "git@github.com:outsider/research.git")
	makeClone(t, filepath.Join(root, "future/outsider/future"), "git@github.com:outsider/future.git")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Pins: []string{"prototyping"}}}
	st := state.New()
	st.Upsert(state.Record{ID: 99, OwnerLogin: "outsider", Name: "research", Path: filepath.Join(root, "research/outsider/research")})
	resolver := &countingNames{t: t}
	clones, err := NewAdopter(gitsvc.NewSync(), resolver).Scan(context.Background(), root, nil, cnf)
	require.NoError(t, err)
	assert.Equal(t, []string{"bob/local", "alice/tool"}, resolver.names)
	rows, err := statussvc.Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	var fetchPaths, statusPaths []string
	for _, c := range clones {
		fetchPaths = append(fetchPaths, c.Path)
	}
	for _, row := range rows {
		if row.OrphanReason == "" {
			statusPaths = append(statusPaths, row.Path)
		}
		assert.Empty(t, row.Error)
	}
	sort.Strings(fetchPaths)
	sort.Strings(statusPaths)
	assert.Equal(t, fetchPaths, statusPaths)
	require.Len(t, rows, 3)
	assert.Equal(t, "out-of-scope", rows[2].OrphanReason)
	assert.False(t, rows[0].Pinned)
	assert.True(t, rows[1].Pinned)
}

type countingNames struct {
	t     *testing.T
	names []string
}

func (r *countingNames) ResolveByName(_ context.Context, owner, name string) (int64, error) {
	require.NotEqual(r.t, "outsider", owner, "out-of-scope branch reached the network resolver")
	r.names = append(r.names, owner+"/"+name)
	if owner == "alice" {
		return 1, nil
	}
	return 2, nil
}

func TestServiceWorkspacePinOutsideRemoteOwners(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "prototyping/old/tool")
	makeClone(t, path, "git@github.com:old/tool.git")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Pins: []string{"prototyping"}}, Defaults: config.Defaults{Concurrency: 1, CloneURL: "ssh"}}
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	api := renamedRepository{snapshot: github.RepoSnapshot{ID: 1, Owner: "new", Name: "renamed", Visibility: github.Private}}
	var out bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "primary", Owners: []string{"unrelated"}}}, root, root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{}, Resolver: api, Confirmer: api, GitSync: gitsvc.NewSync(), Reporter: NewReporter(&out, &out, FormatJSON, 0, false),
	})
	require.NoError(t, err)
	// Adoption updates the real disposable origin without contacting GitHub.
	require.NoError(t, svc.Run(context.Background(), true), out.String())
	loaded, err := store.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Repos, 1)
	assert.Equal(t, path, loaded.Repos[0].Path)
	assert.Equal(t, "workspace", loaded.Repos[0].PinSource)
	assert.Equal(t, "private", loaded.Repos[0].Visibility)
	info, err := gitsvc.NewSync().Inspect(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"git@github.com:new/renamed.git"}, info.Origins)
	out.Reset()
	require.NoError(t, svc.Run(context.Background(), false))
	assert.Contains(t, out.String(), `"kind": "fetch"`)
	assert.NotContains(t, out.String(), `"kind": "move"`)
}
