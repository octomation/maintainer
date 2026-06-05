package fetch_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

type fakeDiscoverer struct{ snaps []github.RepoSnapshot }

func (f fakeDiscoverer) List(_ context.Context, p github.Profile) (github.Discovery, error) {
	return github.Discovery{
		Profile:   p.Name,
		Endpoints: []github.EndpointStat{{Endpoint: "/user/repos", Count: len(f.snaps)}},
		Snapshots: f.snaps,
	}, nil
}

func originWithCommit(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "origin")
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644))
	wt, err := repo.Worktree()
	require.NoError(t, err)
	_, err = wt.Add("f")
	require.NoError(t, err)
	_, err = wt.Commit("init", &gogit.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@e", When: time.Unix(1, 0)}})
	require.NoError(t, err)
	return dir
}

func TestService_PlanThenApply(t *testing.T) {
	origin := originWithCommit(t)
	root := t.TempDir()
	statePath := filepath.Join(t.TempDir(), "state.json")

	cnf := &config.Fetch{Defaults: config.Defaults{
		Root: root, Path: "{{.Owner}}/{{.Repo}}", CloneURL: "https", Concurrency: 2,
	}}
	cnf.Profiles = map[string]config.Profile{}
	require.NoError(t, cnf.Validate())

	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public, HTTPSCloneURL: origin, SourceProfile: "p"}
	store := state.NewStore(afero.NewOsFs(), statePath, nil)

	var out, errw bytes.Buffer
	deps := Deps{
		Store:      store,
		Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}},
		GitSync:    gitsvc.NewSync(),
		Reporter:   NewReporter(&out, &errw, FormatHuman, 0, false),
		Clock:      func() time.Time { return time.Unix(1000, 0).UTC() },
		IDGen:      func() string { return "PLAN1" },
	}
	profiles := []ResolvedProfile{{Name: "p", Token: "", Owners: []string{"acme"}}}

	// Plan-only must not touch the disk.
	svc, err := NewService(cnf, profiles, "/home/op", root, 2, deps)
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	assert.Contains(t, out.String(), "clone")
	_, statErr := os.Stat(statePath)
	assert.True(t, os.IsNotExist(statErr), "plan-only wrote a state file")
	_, cloneErr := os.Stat(filepath.Join(root, "acme/svc"))
	assert.True(t, os.IsNotExist(cloneErr), "plan-only created a clone")

	// Apply clones and persists state.
	out.Reset()
	errw.Reset()
	svc, err = NewService(cnf, profiles, "/home/op", root, 2, deps)
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), true))

	info, err := os.Stat(filepath.Join(root, "acme/svc", ".git"))
	require.NoError(t, err, "apply did not clone")
	assert.True(t, info.IsDir())

	loaded, err := store.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Repos, 1)
	assert.Equal(t, int64(1), loaded.Repos[0].ID)
	assert.Equal(t, filepath.Join(root, "acme/svc"), loaded.Repos[0].Path)
	assert.Equal(t, time.Unix(1000, 0).UTC(), loaded.Repos[0].FirstSeen)
	// the state file is 0600.
	info, err = os.Stat(statePath)
	require.NoError(t, err)
	assert.Equal(t, "-rw-------", info.Mode().String())
}
