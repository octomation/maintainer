package fetch_test

import (
	"bytes"
	"context"
	"errors"
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
	exitpkg "go.octolab.org/toolset/maintainer/internal/pkg/exit"
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

type fakeGitHubNameResolver struct{}

func (fakeGitHubNameResolver) ResolveByName(
	_ context.Context,
	_ github.Profile,
	owner, name string,
) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{Owner: owner, Name: name}, nil
}

type localOriginSync struct {
	*gitsvc.Sync
	localURL     string
	canonicalURL string
}

func (s localOriginSync) Inspect(path string) (gitsvc.CloneInfo, error) {
	info, err := s.Sync.Inspect(path)
	if err != nil {
		return info, err
	}
	for i, origin := range info.Origins {
		if origin == s.localURL {
			info.Origins[i] = s.canonicalURL
		}
	}
	return info, nil
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

func TestService_ReconcilesEmptyRemoteAcrossApplyRuns(t *testing.T) {
	origin := filepath.Join(t.TempDir(), "empty.git")
	_, err := gogit.PlainInit(origin, true)
	require.NoError(t, err)
	root := t.TempDir()
	target := filepath.Join(root, "acme/empty")
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)

	cnf := &config.Fetch{Defaults: config.Defaults{
		Root: root, Path: "{{.Owner}}/{{.Repo}}", CloneURL: "https", Concurrency: 1,
	}, Profiles: map[string]config.Profile{}}
	require.NoError(t, cnf.Validate())
	snap := github.RepoSnapshot{
		ID: 1, Owner: "acme", Name: "empty", Visibility: github.Private,
		HTTPSCloneURL: origin, SourceProfile: "p",
	}
	now := time.Unix(1000, 0).UTC()
	var out, errw bytes.Buffer
	gitSync := localOriginSync{
		Sync: gitsvc.NewSync(), localURL: origin,
		canonicalURL: "https://github.com/acme/empty.git",
	}
	deps := Deps{
		Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}},
		Resolver: fakeGitHubNameResolver{}, GitSync: gitSync,
		Reporter: NewReporter(&out, &errw, FormatHuman, 0, false),
		Clock:    func() time.Time { return now }, IDGen: func() string { return "EMPTY" },
	}
	profiles := []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}

	// First apply materialises the empty checkout and records it.
	svc, err := NewService(cnf, profiles, "/home/op", root, 1, deps)
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), true), out.String())
	_, err = os.Stat(filepath.Join(target, ".git"))
	require.NoError(t, err)

	// A later apply fetches the still-empty upstream without becoming partial.
	now = time.Unix(2000, 0).UTC()
	out.Reset()
	errw.Reset()
	svc, err = NewService(cnf, profiles, "/home/op", root, 1, deps)
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), true), out.String())
	assert.Contains(t, out.String(), "summary: clone=0 fetch=1")
	assert.Contains(t, out.String(), "errors=0")
	assert.NotContains(t, errw.String(), "error:")

	loaded, err := store.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Repos, 1)
	assert.Equal(t, now, loaded.Repos[0].LastApply)
}

func TestService_PlanRecognisesCheckoutWithPushLock(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "acme/svc")
	makeClone(t, target, "git@github.com:acme/svc.git")
	setPushURL(t, target, "no_push")

	statePath := filepath.Join(t.TempDir(), "state.json")
	store := state.NewStore(afero.NewOsFs(), statePath, nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "svc", Path: target,
		RemoteURL: "git@github.com:acme/svc.git", CloneURL: "ssh", SourceProfile: "p",
	})
	require.NoError(t, store.Save(st))

	cnf := &config.Fetch{Defaults: config.Defaults{
		Root: root, Path: "{{.Owner}}/{{.Repo}}", CloneURL: "ssh", Concurrency: 1,
	}, Profiles: map[string]config.Profile{}}
	require.NoError(t, cnf.Validate())
	snap := github.RepoSnapshot{
		ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public,
		SSHCloneURL: "git@github.com:acme/svc.git", SourceProfile: "p",
	}

	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, "/home/op", root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}},
		Resolver: fakeGitHubNameResolver{}, GitSync: gitsvc.NewSync(),
		Reporter: NewReporter(&out, &errw, FormatHuman, 0, false),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	assert.Contains(t, out.String(), "summary: clone=0 fetch=1")
	assert.NotContains(t, out.String(), "+ clone")

	info, err := gitsvc.NewSync().Inspect(target)
	require.NoError(t, err)
	assert.Equal(t, []string{"no_push"}, info.PushURLs)
}

func TestService_PlanReportsOccupiedCloneTargetAsConflict(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "acme/svc")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "local-only.txt"), []byte("keep"), 0o644))

	cnf := &config.Fetch{Defaults: config.Defaults{
		Root: root, Path: "{{.Owner}}/{{.Repo}}", CloneURL: "https", Concurrency: 1,
	}, Profiles: map[string]config.Profile{}}
	require.NoError(t, cnf.Validate())
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public, SourceProfile: "p"}

	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, "/home/op", root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}},
		Resolver: fakeGitHubNameResolver{}, GitSync: gitsvc.NewSync(),
		Reporter: NewReporter(&out, &errw, FormatHuman, 0, false),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	assert.Contains(t, out.String(), "conflict")
	assert.Contains(t, out.String(), "refusing to overwrite")
	assert.Contains(t, out.String(), "summary: clone=0 fetch=0")
	assert.Contains(t, out.String(), "resolve 1 conflict(s) before applying")
	content, readErr := os.ReadFile(filepath.Join(target, "local-only.txt"))
	require.NoError(t, readErr)
	assert.Equal(t, "keep", string(content))

	out.Reset()
	err = svc.Run(context.Background(), true)
	require.Error(t, err)
	var coded exitpkg.Coder
	require.True(t, errors.As(err, &coded))
	assert.Equal(t, exitpkg.Partial, coded.ExitCode())
	assert.Contains(t, out.String(), "1 unresolved conflict(s)")
	content, readErr = os.ReadFile(filepath.Join(target, "local-only.txt"))
	require.NoError(t, readErr)
	assert.Equal(t, "keep", string(content))
}
