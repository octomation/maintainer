package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
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

// actionKinds lists the actions of a plan as "kind id" pairs.
func actionKinds(actions []Action) []string {
	out := make([]string, 0, len(actions))
	for _, a := range actions {
		out = append(out, fmt.Sprintf("%s %d", a.Kind, a.ID))
	}
	return out
}

func TestPlanner_NameReuseNeverRebindsACheckout(t *testing.T) {
	const (
		old = "/work/public/acme/old"
		pin = "/home/op/old"
	)
	snapshots := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public},
		{ID: 2, Owner: "acme", Name: "old", Visibility: github.Public},
	}
	pinned := &config.Fetch{
		Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1},
		Repos:    []config.Repo{{Match: config.RepoMatch{Owner: "acme", Name: "old"}, Path: pin}},
	}
	for _, tc := range []struct {
		name     string
		cnf      *config.Fetch
		record   state.Record
		clones   []DiskClone
		known    map[string]bool
		expected []string
		conflict string // the path of the conflict that names both ids
	}{
		{
			name:     "the recorded remote is the origin: the recorded id moves",
			record:   state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: old, RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh"},
			clones:   []DiskClone{{ID: 2, Path: old, Owner: "acme", Name: "old", RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1}},
			expected: []string{"move 1", "fetch 1", "conflict 2"},
		},
		{
			name:     "the recorded remote differs: a conflict names both ids",
			record:   state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: old, RemoteURL: "git@github.com:acme/fork.git", CloneURL: "ssh"},
			clones:   []DiskClone{{ID: 2, Path: old, Owner: "acme", Name: "old", RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1}},
			expected: []string{"conflict 1", "conflict 2"},
			conflict: old,
		},
		{
			name:     "a pin rule: the recorded id keeps its pin",
			cnf:      pinned,
			record:   state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: pin, PinnedPath: pin, RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh"},
			clones:   []DiskClone{{ID: 2, Path: pin, Owner: "acme", Name: "old", RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1, Pinned: true}},
			expected: []string{"update_remote 1", "fetch 1", "conflict 2"},
		},
		{
			name:     "a pin rule with another remote: a conflict names both ids",
			cnf:      pinned,
			record:   state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: pin, PinnedPath: pin, RemoteURL: "git@github.com:acme/fork.git", CloneURL: "ssh"},
			clones:   []DiskClone{{ID: 2, Path: pin, Owner: "acme", Name: "old", RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1, Pinned: true}},
			expected: []string{"conflict 1", "conflict 2"},
			conflict: pin,
		},
		{
			name: "a previous path: a conflict names both ids, the active checkout is fetched",
			cnf: &config.Fetch{
				Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1},
				Repos:    []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: "/home/op/new"}},
			},
			record: state.Record{ID: 1, OwnerLogin: "acme", Name: "new", Path: "/home/op/new", PinnedPath: "/home/op/new",
				PreviousPaths: []string{old}, RemoteURL: "git@github.com:acme/new.git", CloneURL: "ssh"},
			clones: []DiskClone{
				{ID: 1, Path: "/home/op/new", Owner: "acme", Name: "new", RemoteURL: "git@github.com:acme/new.git", Transport: "ssh", Origins: 1, Pinned: true},
				{ID: 2, Path: old, Owner: "acme", Name: "old", RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1},
			},
			known:    map[string]bool{old: true},
			expected: []string{"fetch 1", "conflict 1", "conflict 2"},
			conflict: old,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := state.New()
			st.Upsert(tc.record)
			acts, err := newPlanner(t, tc.cnf).Plan(PlanInput{
				State: st, Snapshots: snapshots, Clones: tc.clones, KnownCheckouts: tc.known,
				Occupancy: map[string]Occupancy{old: OccupancyForeign, "/work/public/acme/new": OccupancyClear},
			})
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, actionKinds(acts))
			for _, a := range acts {
				assert.False(t, a.Kind == KindAdopt && a.ID == 2, "the checkout is adopted as the new repository")
				assert.False(t, a.Kind == KindClone && a.ID == 1, "the renamed repository is cloned afresh")
				if a.Kind == KindMove {
					assert.Equal(t, old, a.FromPath)
					assert.Equal(t, "/work/public/acme/new", a.ToPath)
				}
				if tc.conflict != "" && a.Kind == KindConflict && a.ID == 1 {
					assert.Equal(t, tc.conflict, a.Path)
					assert.Contains(t, a.Reason, "id=1")
					assert.Contains(t, a.Reason, "id=2")
				}
			}
		})
	}
}

// A checkout bound to its recorded id although its origin now names another
// repository follows the selection of the recorded one: outside it, the
// checkout is not a conflict.
func TestPlanner_DisputedCheckoutFollowsTheRecordedRepository(t *testing.T) {
	const old = "/work/public/acme/old"
	rec := state.Record{ID: 11, OwnerLogin: "acme", Name: "old", Path: old, RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh"}
	clone := DiskClone{ID: 13, Path: old, Owner: "acme", Name: "other", RemoteURL: "git@github.com:acme/other.git", Transport: "ssh", Origins: 1}
	snap := github.RepoSnapshot{ID: 13, Owner: "acme", Name: "other", Visibility: github.Public}
	for _, tc := range []struct {
		name       string
		deselected map[int64]string
		expected   []string
	}{
		{name: "selected", expected: []string{"clone 13", "conflict 11"}},
		{name: "outside the selection", deselected: map[int64]string{11: "transferred outside the selection to other/old"}, expected: []string{"clone 13", "noop 11"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cnf := &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1}}
			st := state.New()
			st.Upsert(rec)
			acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: []DiskClone{clone}, Deselected: tc.deselected})
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, actionKinds(acts))
		})
	}
}

// commitFile writes name and commits it on the current branch.
func commitFile(t *testing.T, repo *gogit.Repository, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	wt, err := repo.Worktree()
	require.NoError(t, err)
	_, err = wt.Add(name)
	require.NoError(t, err)
	_, err = wt.Commit(content, &gogit.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@e", When: time.Unix(2, 0)}})
	require.NoError(t, err)
}

// checkoutWithWork makes a checkout of originURL whose local branch work is
// one commit ahead of master, with an untracked file beside them.
func checkoutWithWork(t *testing.T, dir, originURL string) {
	t.Helper()
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	commitFile(t, repo, dir, "f", "base")
	wt, err := repo.Worktree()
	require.NoError(t, err)
	require.NoError(t, wt.Checkout(&gogit.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("work"), Create: true}))
	commitFile(t, repo, dir, "f", "work")
	require.NoError(t, wt.Checkout(&gogit.CheckoutOptions{Branch: plumbing.Master}))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "local.txt"), []byte("keep"), 0o644))
	_, err = repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{originURL}})
	require.NoError(t, err)
}

// checkoutState captures what reconciliation must leave as it was: HEAD,
// the local branches, the index and the working files.
func checkoutState(t *testing.T, dir string) map[string]string {
	t.Helper()
	snapshot := map[string]string{}
	for _, name := range []string{"HEAD", "index"} {
		data, err := os.ReadFile(filepath.Join(dir, ".git", name))
		require.NoError(t, err)
		snapshot[name] = string(data)
	}
	repo, err := gogit.PlainOpen(dir)
	require.NoError(t, err)
	branches, err := repo.Branches()
	require.NoError(t, err)
	require.NoError(t, branches.ForEach(func(ref *plumbing.Reference) error {
		snapshot[ref.Name().String()] = ref.Hash().String()
		return nil
	}))
	require.NoError(t, filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.Type().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(dir, path)
			snapshot["file "+rel] = string(data)
		}
		return nil
	}))
	return snapshot
}

func planKinds(t *testing.T, out []byte) []string {
	t.Helper()
	var plan planDocument
	require.NoError(t, json.Unmarshal(out, &plan), string(out))
	var got []string
	for _, a := range plan.Actions {
		got = append(got, a.Kind+" "+a.Owner+"/"+a.Name)
	}
	return got
}

// offlineSync reads a local origin as its GitHub URL and never fetches.
type offlineSync struct{ localOriginSync }

func (offlineSync) Fetch(context.Context, string, gitsvc.Auth) error { return nil }

func TestService_NameReuseMovesTheRecordedCheckout(t *testing.T) {
	root := t.TempDir()
	old, renamed := filepath.Join(root, "public/acme/old"), filepath.Join(root, "public/acme/new")
	checkoutWithWork(t, old, "https://github.com/acme/old.git")
	before := checkoutState(t, old)

	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: old,
		RemoteURL: "https://github.com/acme/old.git", CloneURL: "https", SourceProfile: "p"})
	require.NoError(t, store.Save(st))

	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
	origin := originWithCommit(t)
	snapshots := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public, SourceProfile: "p"},
		{ID: 2, Owner: "acme", Name: "old", Visibility: github.Public, SourceProfile: "p", HTTPSCloneURL: origin},
	}
	gitSync := offlineSync{localOriginSync{Sync: gitsvc.NewSync(), localURL: origin, canonicalURL: "https://github.com/acme/old.git"}}
	var out, errw bytes.Buffer
	run := func(apply bool) error {
		out.Reset()
		errw.Reset()
		svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
			Store: store, Discoverer: fakeDiscoverer{snaps: snapshots}, Resolver: fakeGitHubNameResolver{},
			GitSync: gitSync, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
		})
		require.NoError(t, err)
		return svc.Run(context.Background(), apply)
	}

	// The first apply moves the recorded checkout; the reused name waits
	// until its target is free.
	err := run(true)
	var coded exitpkg.Coder
	require.True(t, errors.As(err, &coded), "%v", err)
	assert.Equal(t, exitpkg.Partial, coded.ExitCode())
	assert.ElementsMatch(t, []string{"move acme/new", "fetch acme/new", "conflict acme/old"}, planKinds(t, out.Bytes()))
	_, statErr := os.Stat(old)
	assert.True(t, os.IsNotExist(statErr))
	assert.Equal(t, before, checkoutState(t, renamed))
	info, err := gitsvc.NewSync().Inspect(renamed)
	require.NoError(t, err)
	assert.Equal(t, []string{"https://github.com/acme/new.git"}, info.Origins)

	// The second apply clones the new repository at the freed path.
	require.NoError(t, run(true), errw.String())
	assert.ElementsMatch(t, []string{"clone acme/old", "fetch acme/new"}, planKinds(t, out.Bytes()))
	loaded, err := store.Load()
	require.NoError(t, err)
	rec, ok := loaded.ByID(1)
	require.True(t, ok)
	assert.Equal(t, renamed, rec.Path)
	assert.Equal(t, before, checkoutState(t, renamed))

	require.NoError(t, run(false))
	assert.ElementsMatch(t, []string{"fetch acme/old", "fetch acme/new"}, planKinds(t, out.Bytes()))
}

// caseInsensitive reports whether dir lives on a case-insensitive filesystem.
func caseInsensitive(t *testing.T, dir string) bool {
	t.Helper()
	probe := filepath.Join(dir, "case-probe")
	require.NoError(t, os.WriteFile(probe, nil, 0o644))
	defer func() { _ = os.Remove(probe) }()
	_, err := os.Lstat(filepath.Join(dir, "CASE-PROBE"))
	return err == nil
}

func TestService_CaseOnlyRename(t *testing.T) {
	root := t.TempDir()
	recorded, target := filepath.Join(root, "public/octocat/tool"), filepath.Join(root, "public/OctoCat/tool")
	makeClone(t, recorded, "https://github.com/octocat/tool.git")
	insensitive := caseInsensitive(t, root)
	if !insensitive {
		// A different checkout occupies the new spelling.
		makeClone(t, target, "https://github.com/someone/else.git")
	}

	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "octocat", Name: "tool", Path: recorded,
		RemoteURL: "https://github.com/octocat/tool.git", CloneURL: "https", SourceProfile: "p"})
	require.NoError(t, store.Save(st))
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
	snap := github.RepoSnapshot{ID: 1, Owner: "OctoCat", Name: "tool", Visibility: github.Public, SourceProfile: "p"}

	var out, errw bytes.Buffer
	run := func(apply bool) error {
		out.Reset()
		errw.Reset()
		svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"OctoCat"}}}, root, root, 1, Deps{
			Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: fakeGitHubNameResolver{},
			GitSync: noNetworkFetch{gitsvc.NewSync()}, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
		})
		require.NoError(t, err)
		return svc.Run(context.Background(), apply)
	}

	if !insensitive {
		require.NoError(t, run(false))
		assert.Equal(t, []string{"conflict OctoCat/tool"}, planKinds(t, out.Bytes()))
		return
	}
	require.NoError(t, run(true), errw.String())
	assert.Equal(t, []string{"update_remote OctoCat/tool", "fetch OctoCat/tool"}, planKinds(t, out.Bytes()))
	info, err := gitsvc.NewSync().Inspect(recorded)
	require.NoError(t, err)
	assert.Equal(t, []string{"https://github.com/OctoCat/tool.git"}, info.Origins)
	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, recorded, loaded.Repos[0].Path, "the record keeps the spelling on disk")
	assert.Equal(t, "OctoCat", loaded.Repos[0].OwnerLogin)

	require.NoError(t, run(false))
	assert.Equal(t, []string{"fetch OctoCat/tool"}, planKinds(t, out.Bytes()))
}

func TestPlanner_CaseAliasSettlesTheTarget(t *testing.T) {
	const recorded, target = "/work/public/octocat/tool", "/work/public/OctoCat/tool"
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "octocat", Name: "tool", Path: recorded, RemoteURL: "git@github.com:octocat/tool.git", CloneURL: "ssh"})
	snap := github.RepoSnapshot{ID: 1, Owner: "OctoCat", Name: "tool", Visibility: github.Public}
	clones := []DiskClone{{ID: 1, Path: recorded, RemoteURL: "git@github.com:octocat/tool.git", Transport: "ssh", Origins: 1}}
	occupied := map[string]Occupancy{target: OccupancyForeign}

	acts, err := newPlanner(t, nil).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: clones,
		Occupancy: occupied, Aliases: map[string]string{target: recorded}})
	require.NoError(t, err)
	a := reconciled(t, acts)
	assert.Equal(t, KindUpdateRemote, a.Kind)
	assert.Equal(t, recorded, a.Path)
	assert.Equal(t, "git@github.com:OctoCat/tool.git", a.RemoteURL)

	// Without the filesystem's confirmation, the spelling is another path.
	acts, err = newPlanner(t, nil).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: clones, Occupancy: occupied})
	require.NoError(t, err)
	require.Len(t, acts, 1)
	assert.Equal(t, KindConflict, acts[0].Kind)

	// After the remote update, the plan is settled.
	clones[0].RemoteURL = "git@github.com:OctoCat/tool.git"
	acts, err = newPlanner(t, nil).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: clones,
		Occupancy: occupied, Aliases: map[string]string{target: recorded}})
	require.NoError(t, err)
	require.Len(t, acts, 1)
	assert.Equal(t, KindFetch, acts[0].Kind)
}

func TestApplier_MoveRecordsThePathBeforeTheRemoteUpdate(t *testing.T) {
	const old, renamed = "/work/public/acme/old", "/work/public/acme/new"
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: old, RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh"})
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public}
	acts, err := newPlanner(t, nil).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap},
		Clones: []DiskClone{{ID: 1, Path: old, RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1}}})
	require.NoError(t, err)
	require.Equal(t, KindMove, reconciled(t, acts).Kind)

	g := &fakeGit{updateErr: errors.New("locked config")}
	require.Error(t, NewApplier(g, nil, clock).Execute(context.Background(), acts[0], st))
	rec, _ := st.ByID(1)
	assert.Equal(t, renamed, rec.Path)
	assert.Equal(t, "git@github.com:acme/old.git", rec.RemoteURL, "the remote is recorded only once it is updated")

	acts, err = newPlanner(t, nil).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap},
		Clones: []DiskClone{{ID: 1, Path: renamed, RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1}}})
	require.NoError(t, err)
	a := reconciled(t, acts)
	assert.Equal(t, KindUpdateRemote, a.Kind)
	assert.Equal(t, renamed, a.Path)
}

func TestApplier_FetchRecordsTheObservedMetadata(t *testing.T) {
	const path = "/work/public/acme/tool"
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: path, Visibility: "private", DefaultBranch: "master",
		RemoteURL: "git@github.com:acme/tool.git", CloneURL: "ssh"})
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public, DefaultBranch: "main",
		IsFork: true, IsTemplate: true, IsArchived: true}
	acts, err := newPlanner(t, nil).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap},
		Clones: []DiskClone{{ID: 1, Path: path, RemoteURL: "git@github.com:acme/tool.git", Transport: "ssh", Origins: 1}}})
	require.NoError(t, err)
	require.Len(t, acts, 1)
	require.Equal(t, KindFetch, acts[0].Kind)

	require.NoError(t, NewApplier(new(fakeGit), nil, clock).Execute(context.Background(), acts[0], st))
	rec, _ := st.ByID(1)
	assert.Equal(t, "public", rec.Visibility)
	assert.Equal(t, "main", rec.DefaultBranch)
	assert.True(t, rec.IsFork)
	assert.True(t, rec.IsTemplate)
	assert.True(t, rec.ArchivedOnGitHub)
	assert.Equal(t, clock(), rec.LastSeen)
	assert.Equal(t, path, rec.Path)
}

func TestPlanner_KeptRecordIsNotFetchedThroughAReusedName(t *testing.T) {
	const pin, active, managed = "/home/op/old", "/home/op/tool", "/work/public/acme/old"
	snapshots := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public},
		{ID: 2, Owner: "acme", Name: "old", Visibility: github.Public},
	}
	// The checkout's origin is still acme/old, the name id=2 holds now.
	origin := func(path string, pinned bool) DiskClone {
		return DiskClone{ID: 2, Path: path, Owner: "acme", Name: "old", RemoteURL: "git@github.com:acme/old.git",
			Transport: "ssh", Origins: 1, Pinned: pinned}
	}
	for _, tc := range []struct {
		name      string
		cnf       *config.Fetch
		record    state.Record
		clone     DiskClone
		occupancy map[string]Occupancy
		expected  []string
		next      []string // the reconciliation that updates the remote, then the fetch
	}{
		{
			name: "the pin goes back to a previous path",
			cnf: &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1},
				Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}},
			record: state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: active, PinnedPath: active,
				PreviousPaths: []string{pin}, RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh"},
			clone:     origin(pin, true),
			occupancy: map[string]Occupancy{managed: OccupancyClear},
			expected:  []string{"adopt 1", "clone 2"},
			next:      []string{"update_remote", "fetch"},
		},
		{
			name: "the record relocates to a previous path",
			record: state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: "/work/private/acme/old",
				PreviousPaths: []string{managed}, RemoteURL: "git@github.com:acme/old.git", CloneURL: "ssh"},
			clone:     origin(managed, false),
			occupancy: map[string]Occupancy{managed: OccupancyForeign},
			expected:  []string{"relocate 1", "conflict 2"},
			next:      []string{"move", "fetch"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := state.New()
			st.Upsert(tc.record)
			acts, err := newPlanner(t, tc.cnf).Plan(PlanInput{State: st, Snapshots: snapshots, Clones: []DiskClone{tc.clone}, Occupancy: tc.occupancy})
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expected, actionKinds(acts), "no fetch through an origin that names id=2 now")

			// Once the reconciliation is applied, the remote is updated first
			// and only then fetched.
			apply := NewApplier(new(fakeGit), nil, clock)
			for _, act := range acts {
				if act.ID == 1 {
					require.NoError(t, apply.Execute(context.Background(), act, st))
				}
			}
			acts, err = newPlanner(t, tc.cnf).Plan(PlanInput{State: st, Snapshots: snapshots, Clones: []DiskClone{tc.clone}, Occupancy: tc.occupancy})
			require.NoError(t, err)
			var own []string
			for _, act := range acts {
				if act.ID == 1 {
					own = append(own, string(act.Kind))
					assert.True(t, act.Kind != KindMove || act.UpdateRemote, "the move updates the remote before its refresh")
				}
			}
			assert.Equal(t, tc.next, own)
		})
	}
}

// localOnlyFetch fetches only from local origins; a GitHub URL is not
// contacted. It reads a local origin as its GitHub URL, like localOriginSync.
type localOnlyFetch struct{ localOriginSync }

func (s localOnlyFetch) Fetch(ctx context.Context, path string, auth gitsvc.Auth) error {
	info, err := s.Sync.Inspect(path)
	if err != nil || len(info.Origins) != 1 || !filepath.IsAbs(info.Origins[0]) {
		return err
	}
	return s.Sync.Fetch(ctx, path, auth)
}

func remoteRefs(t *testing.T, dir string) map[string]string {
	t.Helper()
	repo, err := gogit.PlainOpen(dir)
	require.NoError(t, err)
	refs, err := repo.References()
	require.NoError(t, err)
	out := map[string]string{}
	require.NoError(t, refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() {
			out[ref.Name().String()] = ref.Hash().String()
		}
		return nil
	}))
	return out
}

func TestService_KeptRecordNeverFetchesTheRepositoryThatTookItsName(t *testing.T) {
	root := t.TempDir()
	renamed := originWithCommit(t) // id=1, now acme/new
	reused := filepath.Join(t.TempDir(), "reused")
	repo, err := gogit.PlainInit(reused, false)
	require.NoError(t, err)
	commitFile(t, repo, reused, "g", "unrelated history") // id=2, now acme/old

	pin := filepath.Join(t.TempDir(), "tool")
	checkout, err := gogit.PlainClone(pin, false, &gogit.CloneOptions{URL: renamed})
	require.NoError(t, err)
	// The old name leads to id=2 now.
	require.NoError(t, checkout.DeleteRemote("origin"))
	_, err = checkout.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{reused}})
	require.NoError(t, err)
	before := remoteRefs(t, pin)
	require.NotEmpty(t, before)

	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	elsewhere := filepath.Join(t.TempDir(), "elsewhere") // the pin before it went back
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "old", Path: elsewhere, PinnedPath: elsewhere, PreviousPaths: []string{pin},
		RemoteURL: "https://github.com/acme/old.git", CloneURL: "https", SourceProfile: "p"})
	require.NoError(t, store.Save(st))
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}}
	snapshots := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public, SourceProfile: "p"},
		{ID: 2, Owner: "acme", Name: "old", Visibility: github.Public, SourceProfile: "p", HTTPSCloneURL: reused},
	}
	gitSync := localOnlyFetch{localOriginSync{Sync: gitsvc.NewSync(), localURL: reused, canonicalURL: "https://github.com/acme/old.git"}}
	var out, errw bytes.Buffer
	run := func() []string {
		out.Reset()
		errw.Reset()
		svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
			Store: store, Discoverer: fakeDiscoverer{snaps: snapshots}, Resolver: fakeGitHubNameResolver{},
			GitSync: gitSync, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
		})
		require.NoError(t, err)
		require.NoError(t, svc.Run(context.Background(), true), errw.String())
		return planKinds(t, out.Bytes())
	}

	assert.ElementsMatch(t, []string{"adopt acme/new", "clone acme/old"}, run())
	assert.Equal(t, before, remoteRefs(t, pin), "the checkout of id=1 is not fetched from id=2")

	assert.ElementsMatch(t, []string{"update_remote acme/new", "fetch acme/new", "fetch acme/old"}, run())
	assert.Equal(t, before, remoteRefs(t, pin))
	info, err := gitsvc.NewSync().Inspect(pin)
	require.NoError(t, err)
	assert.Equal(t, []string{"https://github.com/acme/new.git"}, info.Origins)
}
