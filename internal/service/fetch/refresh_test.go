package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

// originOnMain makes a non-bare origin whose branch main has one commit.
func originOnMain(t *testing.T) (string, *gogit.Repository) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "origin")
	repo, err := gogit.PlainInitWithOptions(dir, &gogit.PlainInitOptions{
		InitOptions: gogit.InitOptions{DefaultBranch: plumbing.NewBranchReferenceName("main")},
	})
	require.NoError(t, err)
	commitFile(t, repo, dir, "f", "first")
	return dir, repo
}

func remoteMain(t *testing.T, dir string) plumbing.Hash {
	t.Helper()
	repo, err := gogit.PlainOpen(dir)
	require.NoError(t, err)
	ref, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", "main"), false)
	require.NoError(t, err)
	return ref.Hash()
}

func headOf(t *testing.T, repo *gogit.Repository) plumbing.Hash {
	t.Helper()
	ref, err := repo.Head()
	require.NoError(t, err)
	return ref.Hash()
}

// hidingSync hides the origin right before the next fetch, as if it became
// unreadable between the adoption and the refresh.
type hidingSync struct {
	localOriginSync
	hide   *bool
	origin string
}

func (s hidingSync) Fetch(ctx context.Context, path string, auth gitsvc.Auth) error {
	if *s.hide {
		if err := os.Rename(s.origin, s.origin+".off"); err != nil {
			return err
		}
	}
	return s.localOriginSync.Fetch(ctx, path, auth)
}

type planSummary struct {
	Actions []struct {
		Kind string `json:"kind"`
	} `json:"actions"`
	Summary Summary `json:"summary"`
}

func TestService_AdoptionRefreshesRefsInTheSameApply(t *testing.T) {
	for _, broken := range []bool{false, true} {
		t.Run(map[bool]string{false: "fetch succeeds", true: "fetch fails"}[broken], func(t *testing.T) {
			origin, upstream := originOnMain(t)
			root := t.TempDir()
			checkout := filepath.Join(root, "public/acme/tool")
			_, err := gogit.PlainClone(checkout, false, &gogit.CloneOptions{URL: origin})
			require.NoError(t, err)
			setPushURL(t, checkout, "no_push")
			commitFile(t, upstream, origin, "f", "second") // the checkout's origin/main is stale now
			latest := headOf(t, upstream)
			require.NotEqual(t, latest, remoteMain(t, checkout))
			before := checkoutState(t, checkout)

			cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 2}}
			store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
			snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public, SourceProfile: "p"}
			hide := broken
			gitSync := hidingSync{
				localOriginSync: localOriginSync{Sync: gitsvc.NewSync(), localURL: origin, canonicalURL: "https://github.com/acme/tool.git"},
				hide:            &hide, origin: origin,
			}
			var out, errw bytes.Buffer
			run := func(apply bool) (planSummary, error) {
				out.Reset()
				errw.Reset()
				svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 2, Deps{
					Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: fakeGitHubNameResolver{},
					GitSync: gitSync, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
				})
				require.NoError(t, err)
				runErr := svc.Run(context.Background(), apply)
				var plan planSummary
				require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
				return plan, runErr
			}

			// A plan-only run shows the refresh.
			plan, err := run(false)
			require.NoError(t, err)
			assert.Equal(t, Summary{Adopt: 1, Fetch: 1}, plan.Summary)

			plan, err = run(true)
			assert.Equal(t, 1, plan.Summary.Adopt)
			assert.Equal(t, 1, plan.Summary.Fetch)
			loaded, loadErr := store.Load()
			require.NoError(t, loadErr)
			require.Len(t, loaded.Repos, 1, "the adoption is recorded")
			if broken {
				var coded exitpkg.Coder
				require.True(t, errors.As(err, &coded), "%v", err)
				assert.Equal(t, exitpkg.Partial, coded.ExitCode())
				assert.Equal(t, 1, plan.Summary.Errors)
				assert.NotEqual(t, latest, remoteMain(t, checkout))
				hide = false
				require.NoError(t, os.Rename(origin+".off", origin))
			} else {
				require.NoError(t, err, errw.String())
				assert.Equal(t, latest, remoteMain(t, checkout))
			}
			assert.Equal(t, before, checkoutState(t, checkout))
			info, err := gitsvc.NewSync().Inspect(checkout)
			require.NoError(t, err)
			assert.Equal(t, []string{"no_push"}, info.PushURLs)

			// The next apply only fetches.
			plan, err = run(true)
			require.NoError(t, err, errw.String())
			assert.Equal(t, Summary{Fetch: 1}, plan.Summary)
			assert.Equal(t, latest, remoteMain(t, checkout))
			assert.Equal(t, before, checkoutState(t, checkout))
		})
	}
}

func TestService_FailedReconciliationSkipsItsRefresh(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "public/acme/tool")
	makeClone(t, path, "https://github.com/acme/old-name.git")
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: path,
		RemoteURL: "https://github.com/acme/old-name.git", CloneURL: "https", SourceProfile: "p"})
	require.NoError(t, store.Save(st))
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public, SourceProfile: "p"}
	g := &failingUpdate{Sync: gitsvc.NewSync(), fetches: new(int)}

	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: renamedRepository{snapshot: snap},
		GitSync: g, Reporter: NewReporter(&out, &errw, FormatHuman, 0, false),
	})
	require.NoError(t, err)
	err = svc.Run(context.Background(), true)
	var coded exitpkg.Coder
	require.True(t, errors.As(err, &coded), "%v", err)
	assert.Equal(t, exitpkg.Partial, coded.ExitCode())
	assert.Zero(t, *g.fetches, "the refresh of a failed update runs")
	assert.Contains(t, out.String(), "update_remote=1 adopt=0")
	assert.Contains(t, out.String(), "+ fetch remote-tracking refs")
	assert.Contains(t, out.String(), "errors=1")
}

// failingUpdate is the real Git adapter whose remote updates fail; it counts
// fetches instead of running them.
type failingUpdate struct {
	*gitsvc.Sync
	fetches *int
}

func (failingUpdate) UpdateRemote(string, string) error { return errors.New("config is locked") }

func (f *failingUpdate) Fetch(context.Context, string, gitsvc.Auth) error {
	*f.fetches++
	return nil
}
