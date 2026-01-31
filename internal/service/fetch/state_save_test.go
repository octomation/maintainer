package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

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

// faultyFs fails every write while broken is set, and the next `once` writes,
// so a state save fails on demand while loading keeps working.
type faultyFs struct {
	afero.Fs
	broken atomic.Bool
	once   atomic.Int32
}

func (f *faultyFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|os.O_RDWR) != 0 && (f.broken.Load() || f.once.Add(-1) >= 0) {
		return nil, &os.PathError{Op: "open", Path: name, Err: errors.New("injected write failure")}
	}
	return f.Fs.OpenFile(name, flag, perm)
}

// faultyGit moves for real and fails remote updates on demand; clones and
// fetches are recorded and never reach the network.
type faultyGit struct {
	*gitsvc.Sync
	failRemote  bool
	cloned      []string
	fetched     []string
	beforeFetch func(path string)
}

func (g *faultyGit) UpdateRemote(path, url string) error {
	if g.failRemote {
		return errors.New("injected remote update failure")
	}
	return g.Sync.UpdateRemote(path, url)
}

func (g *faultyGit) Clone(_ context.Context, opt gitsvc.CloneOptions) error {
	g.cloned = append(g.cloned, opt.Path)
	return errors.New("clone is not expected")
}

func (g *faultyGit) Fetch(_ context.Context, path string, _ gitsvc.Auth) error {
	if g.beforeFetch != nil {
		g.beforeFetch(path)
	}
	g.fetched = append(g.fetched, path)
	return nil
}

// redirects resolves a name like GitHub does, following renames to the id.
type redirects map[string]int64

func (r redirects) ResolveByName(_ context.Context, _ github.Profile, owner, name string) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{ID: r[owner+"/"+name], Owner: owner, Name: name}, nil
}

type jsonPlan struct {
	Actions []struct {
		Kind     string `json:"kind"`
		ID       int64  `json:"id"`
		Path     string `json:"path"`
		FromPath string `json:"from_path"`
		ToPath   string `json:"to_path"`
		Flag     string `json:"flag"`
		Reason   string `json:"reason"`
	} `json:"actions"`
}

type renameFixture struct {
	root, old, renamed string
	fs                 *faultyFs
	store              *state.Store
	git                *faultyGit
	cnf                *config.Fetch
	names              redirects
}

// newRenameFixture tracks acme/old at its layout path; GitHub now calls it
// acme/new, so a run plans a move with a remote update.
func newRenameFixture(t *testing.T) *renameFixture {
	t.Helper()
	f := &renameFixture{
		root: t.TempDir(), fs: &faultyFs{Fs: afero.NewOsFs()}, git: &faultyGit{Sync: gitsvc.NewSync()},
		names: redirects{"acme/old": 1, "acme/new": 1},
	}
	f.old = filepath.Join(f.root, "public/acme/old")
	f.renamed = filepath.Join(f.root, "public/acme/new")
	makeClone(t, f.old, "git@github.com:acme/old.git")
	f.store = state.NewStore(f.fs, filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "acme", Name: "old", Visibility: "public", Path: f.old,
		RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh", SourceProfile: "p",
	})
	require.NoError(t, f.store.Save(st))
	f.cnf = &config.Fetch{Workspace: &config.Workspace{Root: f.root}, Defaults: config.Defaults{CloneURL: "ssh", Concurrency: 1}}
	require.NoError(t, f.cnf.Validate())
	return f
}

func (f *renameFixture) run(t *testing.T, apply bool, snaps ...github.RepoSnapshot) (jsonPlan, string, error) {
	t.Helper()
	var out, errw bytes.Buffer
	svc, err := NewService(f.cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, f.root, f.root, 1, Deps{
		Store: f.store, Discoverer: fakeDiscoverer{snaps: snaps}, Resolver: f.names,
		GitSync: f.git, Clock: clock, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
	})
	require.NoError(t, err)
	runErr := svc.Run(context.Background(), apply)
	var plan jsonPlan
	require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
	return plan, errw.String(), runErr
}

func (f *renameFixture) record(t *testing.T) state.Record {
	t.Helper()
	st, err := f.store.Load()
	require.NoError(t, err)
	require.Len(t, st.Repos, 1)
	return st.Repos[0]
}

func renamed(id int64, name string) github.RepoSnapshot {
	return github.RepoSnapshot{
		ID: id, Owner: "acme", Name: name, Visibility: github.Public, SourceProfile: "p",
		SSHCloneURL: "git@github.com:acme/" + name + ".git",
	}
}

func requireExit(t *testing.T, err error, code int) {
	t.Helper()
	require.Error(t, err)
	got := exitpkg.Transport
	var coded exitpkg.Coder
	if errors.As(err, &coded) {
		got = coded.ExitCode()
	}
	assert.Equal(t, code, got, err.Error())
}

func kinds(plan jsonPlan) []string {
	var out []string
	for _, act := range plan.Actions {
		out = append(out, act.Kind)
	}
	return out
}

func TestApplyStateSaveFailureAfterMove(t *testing.T) {
	f := newRenameFixture(t)
	f.fs.broken.Store(true)
	_, errs, err := f.run(t, true, renamed(1, "new"))
	requireExit(t, err, exitpkg.Partial)
	assert.Contains(t, errs, "save state:")
	assert.Contains(t, errs, "injected write failure")
	assert.Contains(t, errs, "not recorded: acme/new")
	assert.NotContains(t, errs, "acme/new, acme/new", "the refresh names the repository once")
	assert.Equal(t, []string{f.renamed}, f.git.fetched, "the move is refreshed in the same apply")

	// The checkout moved and carries the new origin; the state file did not.
	info, inspectErr := gitsvc.NewSync().Inspect(f.renamed)
	require.NoError(t, inspectErr)
	assert.Equal(t, []string{"git@github.com:acme/new.git"}, info.Origins)
	assert.NoDirExists(t, f.old)
	f.fs.broken.Store(false)
	assert.Equal(t, f.old, f.record(t).Path)

	// The same origin and the same id: the rerun relocates without a clone.
	plan, _, err := f.run(t, false, renamed(1, "new"))
	require.NoError(t, err)
	require.Equal(t, []string{"relocate", "fetch"}, kinds(plan))
	assert.Equal(t, f.old, plan.Actions[0].FromPath)
	assert.Equal(t, f.renamed, plan.Actions[0].ToPath)
	_, _, err = f.run(t, true, renamed(1, "new"))
	require.NoError(t, err)
	assert.Equal(t, f.renamed, f.record(t).Path)
	assert.Empty(t, f.git.cloned)
}

func TestApplyStateSaveFailureCountsOnceSaved(t *testing.T) {
	f := newRenameFixture(t)
	f.fs.once.Store(1) // the save after the move fails, the final one succeeds
	_, errs, err := f.run(t, true, renamed(1, "new"))
	requireExit(t, err, exitpkg.Partial)
	assert.Contains(t, errs, "not recorded: acme/new")
	assert.Equal(t, f.renamed, f.record(t).Path, "the final save records the move")
}

func TestApplyStateSaveFailureWithoutActions(t *testing.T) {
	f := newRenameFixture(t)
	f.cnf.Repos = []config.Repo{{Match: config.RepoMatch{ID: 1}, Ignore: true}}
	f.fs.broken.Store(true)
	plan, errs, err := f.run(t, true)
	assert.Empty(t, plan.Actions)
	requireExit(t, err, exitpkg.Transport)
	assert.Contains(t, err.Error(), "state not saved")
	assert.Contains(t, errs, "save state:")
}

func TestApplyStateAndRemoteFailureAfterMove(t *testing.T) {
	f := newRenameFixture(t)
	f.fs.broken.Store(true)
	f.git.failRemote = true
	_, errs, err := f.run(t, true, renamed(1, "new"))
	requireExit(t, err, exitpkg.Partial)
	assert.Contains(t, errs, "injected remote update failure")
	assert.Contains(t, errs, "save state:")
	assert.Contains(t, errs, "not recorded: acme/new", "the move stands, and so does its record")
	assert.Empty(t, f.git.fetched, "a failed move is not refreshed")

	// The moved checkout still carries the old origin, and its recorded path
	// is absent.
	info, inspectErr := gitsvc.NewSync().Inspect(f.renamed)
	require.NoError(t, inspectErr)
	assert.Equal(t, []string{"git@github.com:acme/old.git"}, info.Origins)
	f.fs.broken.Store(false)
	f.git.failRemote = false
	assert.Equal(t, f.old, f.record(t).Path)

	// While GitHub still resolves the old name to the recorded id, the
	// checkout is relocated.
	plan, _, err := f.run(t, false, renamed(1, "new"))
	require.NoError(t, err)
	require.Equal(t, []string{"relocate", "fetch"}, kinds(plan))
	assert.Equal(t, f.renamed, plan.Actions[0].ToPath)

	// A new repository has taken the old name since: the checkout is not
	// adopted as it, and the renamed one is not cloned over it.
	f.names["acme/old"] = 2
	plan, _, err = f.run(t, true, renamed(1, "new"), renamed(2, "old"))
	requireExit(t, err, exitpkg.Partial)
	require.NotEmpty(t, plan.Actions)
	var guarded bool
	for _, act := range plan.Actions {
		assert.Equal(t, "conflict", act.Kind, act)
		if act.ID == 2 {
			guarded = true
			assert.Equal(t, f.renamed, act.Path)
			assert.Contains(t, act.Reason, "recorded for id=1 at "+f.old)
		}
	}
	assert.True(t, guarded, kinds(plan))
	assert.Empty(t, f.git.cloned)
	assert.Equal(t, f.old, f.record(t).Path)
	info, inspectErr = gitsvc.NewSync().Inspect(f.renamed)
	require.NoError(t, inspectErr)
	assert.Equal(t, []string{"git@github.com:acme/old.git"}, info.Origins)
}

// A move whose remote update fails stands: its new path is saved at once, not
// only by the save after the routine fetches.
func TestApplyRemoteFailureSavesTheMoveAtOnce(t *testing.T) {
	f := newRenameFixture(t)
	other := filepath.Join(f.root, "public/acme/other")
	makeClone(t, other, "git@github.com:acme/other.git")
	st, err := f.store.Load()
	require.NoError(t, err)
	st.Upsert(state.Record{ID: 2, OwnerLogin: "acme", Name: "other", Visibility: "public", Path: other,
		RemoteURL: "git@github.com:acme/other.git", CloneURL: "ssh", SourceProfile: "p"})
	require.NoError(t, f.store.Save(st))
	f.names["acme/other"] = 2
	f.git.failRemote = true
	var saved string
	f.git.beforeFetch = func(path string) {
		st, err := f.store.Load()
		require.NoError(t, err)
		rec, ok := st.ByID(1)
		require.True(t, ok)
		saved = rec.Path
	}
	plan, errs, err := f.run(t, true, renamed(1, "new"), renamed(2, "other"))
	requireExit(t, err, exitpkg.Partial)
	assert.ElementsMatch(t, []string{"move", "fetch", "fetch"}, kinds(plan))
	assert.Contains(t, errs, "injected remote update failure")
	assert.NotContains(t, errs, "save state:")
	assert.Equal(t, []string{other}, f.git.fetched, "the failed move is not refreshed")
	assert.Equal(t, f.renamed, saved, "the move is in the state file before the fetches run")
}

// Two records with missing paths share the origin of a checkout: binding it
// to either id is unsafe, whatever their order in the state.
func TestUnsavedMoveGuardKeepsAmbiguousOrigins(t *testing.T) {
	for _, order := range [][]int64{{1, 2}, {2, 1}} {
		t.Run(fmt.Sprint(order), func(t *testing.T) {
			root := t.TempDir()
			checkout := filepath.Join(root, "public/acme/new")
			makeClone(t, checkout, "git@github.com:acme/old.git")
			records := map[int64]state.Record{
				1: {ID: 1, OwnerLogin: "acme", Name: "old", Path: filepath.Join(root, "public/acme/old"), RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh", SourceProfile: "p"},
				2: {ID: 2, OwnerLogin: "acme", Name: "old", Path: filepath.Join(root, "public/acme/elsewhere"), RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh", SourceProfile: "p"},
			}
			st := state.New()
			for _, id := range order {
				st.Upsert(records[id])
			}
			store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
			require.NoError(t, store.Save(st))
			cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "ssh", Concurrency: 1}}
			var out, errw bytes.Buffer
			svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
				Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{renamed(1, "new"), renamed(2, "old")}},
				Resolver: redirects{"acme/old": 2}, GitSync: &faultyGit{Sync: gitsvc.NewSync()}, Clock: clock,
				Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
			})
			require.NoError(t, err)
			require.NoError(t, svc.Run(context.Background(), false))
			var plan jsonPlan
			require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
			guarded := false
			for _, act := range plan.Actions {
				assert.NotEqual(t, "relocate", act.Kind, act)
				assert.NotEqual(t, "adopt", act.Kind, act)
				if act.Path == checkout && act.ID == 2 {
					guarded = true
					assert.Equal(t, "conflict", act.Kind, act)
					assert.Contains(t, act.Reason, "recorded for id=1")
				}
			}
			assert.True(t, guarded, kinds(plan))
		})
	}
}

// A final save that fails names the repositories whose new observation it
// loses, even when no action ran.
func TestApplyFinalSaveFailureNamesObservations(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "public/acme/tool")
	makeClone(t, path, "git@github.com:acme/tool.git")
	fs := &faultyFs{Fs: afero.NewOsFs()}
	store := state.NewStore(fs, filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: path, SourceProfile: "p"})
	require.NoError(t, store.Save(st))
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{Concurrency: 1}}
	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p"}}, root, root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{}, Resolver: absentRepo{}, Confirmer: absentRepo{},
		GitSync: gitsvc.NewSync(), Clock: clock, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
	})
	require.NoError(t, err)
	fs.broken.Store(true)
	err = svc.Run(context.Background(), true)
	requireExit(t, err, exitpkg.Transport)
	assert.Contains(t, errw.String(), "save state:")
	assert.Contains(t, errw.String(), "remote observations of acme/tool")
	fs.broken.Store(false)
	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded.Repos[0].RemoteStatus)
}
