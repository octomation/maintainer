package fetch_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

type renamedRepository struct{ snapshot github.RepoSnapshot }

func (r renamedRepository) ConfirmByID(context.Context, github.Profile, int64) (github.RepoSnapshot, error) {
	return r.snapshot, nil
}

func (r renamedRepository) ResolveByName(context.Context, github.Profile, string, string) (github.RepoSnapshot, error) {
	return r.snapshot, nil
}

func TestServiceDiscoversPersistedPinAfterTransfer(t *testing.T) {
	root := t.TempDir()
	pin := filepath.Join(t.TempDir(), ".dotfiles")
	repo, err := git.PlainInit(pin, false)
	require.NoError(t, err)
	_, err = repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{"git@github.com:acme/dotfiles.git"}})
	require.NoError(t, err)
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "dotfiles", Path: pin, PinnedPath: pin})
	require.NoError(t, store.Save(st))
	cnf := &config.Fetch{Defaults: config.Defaults{Root: root, Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1}}
	api := renamedRepository{snapshot: github.RepoSnapshot{ID: 1, Owner: "other", Name: "super-dotfiles", SourceProfile: "p"}}
	var output bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{}, Confirmer: api, Resolver: api, GitSync: noNetworkFetch{gitsvc.NewSync()},
		Reporter: NewReporter(&output, &output, FormatJSON, 0, false),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), true), output.String())
	loaded, err := store.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Repos, 1)
	assert.Equal(t, pin, loaded.Repos[0].Path)
	assert.Equal(t, pin, loaded.Repos[0].PinnedPath)
	assert.Equal(t, "other", loaded.Repos[0].OwnerLogin)
	info, err := gitsvc.NewSync().Inspect(pin)
	require.NoError(t, err)
	assert.Equal(t, []string{"git@github.com:other/super-dotfiles.git"}, info.Origins)
}

func TestPinnedRenameOnDisk(t *testing.T) {
	root := t.TempDir()
	pin := filepath.Join(root, ".dotfiles")
	duplicate := filepath.Join(root, "public", "acme", "dotfiles")
	for _, path := range []string{pin, duplicate} {
		repo, err := git.PlainInit(path, false)
		require.NoError(t, err)
		_, err = repo.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{"git@github.com:acme/dotfiles.git"}})
		require.NoError(t, err)
	}
	cnf := &config.Fetch{Defaults: config.Defaults{Root: root, Path: config.DefaultPath, CloneURL: "ssh"},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "dotfiles", Path: duplicate})
	sync := gitsvc.NewSync()
	clones, err := NewAdopter(sync, nil).Scan(context.Background(), root,
		[]github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "dotfiles"}}, cnf)
	require.NoError(t, err)
	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Clones: clones,
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "other", Name: "super-dotfiles"}}})
	require.NoError(t, err)
	require.Len(t, acts, 3)
	assert.Equal(t, KindAdopt, acts[1].After)
	assert.Equal(t, KindOrphan, acts[2].Kind)
	assert.False(t, acts[2].Executable())
	require.NoError(t, NewApplier(sync, nil, clock).Execute(context.Background(), acts[0], st))
	info, err := sync.Inspect(pin)
	require.NoError(t, err)
	assert.Equal(t, []string{"git@github.com:acme/dotfiles.git"}, info.Origins, "adoption keeps the origin")

	// The next apply updates the origin of the selected pin only.
	acts, err = newPlanner(t, cnf).Plan(PlanInput{State: st, Clones: clones,
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "other", Name: "super-dotfiles"}}})
	require.NoError(t, err)
	require.Len(t, acts, 3)
	assert.Equal(t, KindUpdateRemote, acts[0].Kind)
	assert.Equal(t, pin, acts[0].Path)
	require.NoError(t, NewApplier(sync, nil, clock).Execute(context.Background(), acts[0], st))
	info, err = sync.Inspect(pin)
	require.NoError(t, err)
	assert.Equal(t, []string{"git@github.com:other/super-dotfiles.git"}, info.Origins)
	info, err = sync.Inspect(duplicate)
	require.NoError(t, err)
	assert.Equal(t, []string{"git@github.com:acme/dotfiles.git"}, info.Origins)
	_, err = os.Stat(filepath.Join(root, "public", "other", "super-dotfiles"))
	assert.True(t, os.IsNotExist(err))
}

func TestPinnedCheckoutLifecycle(t *testing.T) {
	const pin = "/home/alice/.dotfiles"
	const duplicate = "/work/public/acme/dotfiles"
	cnf := &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh"},
		Repos: []config.Repo{{Match: config.RepoMatch{Owner: "acme", Name: "dotfiles"}, Path: pin}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "dotfiles", Path: duplicate, FirstSeen: clock()})
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "dotfiles"}
	clones := []DiskClone{
		{ID: 1, Path: duplicate, Transport: "ssh", RemoteURL: "git@github.com:acme/dotfiles.git"},
		{ID: 1, Path: pin, Transport: "ssh", RemoteURL: "git@github.com:acme/dotfiles.git"},
	}
	g := new(fakeGit)
	apply := NewApplier(g, nil, clock)
	for i, kinds := range [][]Kind{{KindAdopt, KindFetch}, {KindUpdateRemote, KindFetch}, {KindFetch}} {
		in := PlanInput{State: st, Clones: clones}
		if i == 1 {
			// A transfer leaves the discovery scope and the old name-based rule.
			snap.Owner, snap.Name = "other", "super-dotfiles"
			in.Confirmations = map[int64]Confirmation{1: {Status: ConfirmFound, Snapshot: &snap}}
		} else {
			in.Snapshots = []github.RepoSnapshot{snap}
		}
		acts, err := newPlanner(t, cnf).Plan(in)
		require.NoError(t, err)
		require.Len(t, acts, len(kinds)+1)
		orphan := acts[len(kinds)]
		assert.Equal(t, KindOrphan, orphan.Kind)
		assert.Equal(t, duplicate, orphan.Path)
		for j, kind := range kinds {
			assert.Equal(t, kind, acts[j].Kind)
			assert.Equal(t, pin, acts[j].Path)
			require.NoError(t, apply.Execute(context.Background(), acts[j], st))
		}
		if i == 1 {
			clones[1].RemoteURL = "git@github.com:other/super-dotfiles.git"
		}
	}
	rec, _ := st.ByID(1)
	assert.Equal(t, pin, rec.Path)
	assert.Equal(t, pin, rec.PinnedPath)
	assert.Equal(t, clock(), rec.FirstSeen)
	assert.Empty(t, g.moved)
	assert.Empty(t, g.cloned)
	assert.Equal(t, []string{pin, pin, pin}, g.fetched)
	assert.Equal(t, [][2]string{{pin, "git@github.com:other/super-dotfiles.git"}}, g.updated)
}

func TestPinnedSelection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		clones []DiskClone
		kind   Kind
	}{
		{"duplicate preferred", []DiskClone{{ID: 1, Path: "/work/special"}, {ID: 1, Path: "/work/old"}}, KindAdopt},
		{"missing with duplicate", []DiskClone{{ID: 1, Path: "/work/old"}}, KindConflict},
		{"foreign", []DiskClone{{ID: 2, Path: "/work/special"}}, KindConflict},
		{"ambiguous", []DiskClone{{ID: 1, Path: "/work/special", Origins: 2}}, KindConflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cnf := &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh"},
				Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: "special"}}}
			acts, err := newPlanner(t, cnf).Plan(PlanInput{State: state.New(), Clones: tc.clones,
				Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "dotfiles"}}})
			require.NoError(t, err)
			var selected []Action
			for _, act := range acts {
				if act.After != "" {
					assert.Equal(t, KindAdopt, tc.kind, "only a reconciliation is refreshed")
					continue
				}
				if act.Kind == KindOrphan {
					assert.Equal(t, "duplicate-pin", act.OrphanReason)
					assert.False(t, act.Executable())
				} else {
					selected = append(selected, act)
				}
			}
			require.Len(t, selected, 1)
			assert.Equal(t, tc.kind, selected[0].Kind)
			assert.Equal(t, "/work/special", selected[0].Path)
		})
	}
}

func TestConfirmationWithoutRenameRemainsNoop(t *testing.T) {
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: "/work/public/acme/tool"})
	acts, err := newPlanner(t, nil).Plan(PlanInput{State: st,
		Confirmations: map[int64]Confirmation{1: {Status: ConfirmFound,
			Snapshot: &github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool"}}}})
	require.NoError(t, err)
	require.Len(t, acts, 1)
	assert.Equal(t, KindNoop, acts[0].Kind)
}

func TestPinnedAdoptionOfARenamedRepositoryIsAdditive(t *testing.T) {
	const pin = "/home/op/tool"
	cnf := &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh"},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}}
	st := state.New()
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public}
	clone := DiskClone{ID: 1, Path: pin, Owner: "acme", Name: "old", RemoteURL: "git@github.com:acme/old.git", Transport: "ssh", Origins: 1, Pinned: true}
	g := new(fakeGit)
	apply := NewApplier(g, nil, clock)

	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: []DiskClone{clone}})
	require.NoError(t, err)
	adopt := reconciled(t, acts)
	assert.Equal(t, KindAdopt, adopt.Kind)
	assert.False(t, adopt.UpdateRemote)
	assert.Equal(t, "git@github.com:acme/old.git", adopt.Record.RemoteURL, "the observed origin is recorded")
	for _, act := range acts {
		require.NoError(t, apply.Execute(context.Background(), act, st))
	}
	assert.Empty(t, g.updated, "adoption changes no remote")

	acts, err = newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{snap}, Clones: []DiskClone{clone}})
	require.NoError(t, err)
	update := reconciled(t, acts)
	assert.Equal(t, KindUpdateRemote, update.Kind)
	assert.Equal(t, pin, update.Path)
	assert.Equal(t, "git@github.com:acme/new.git", update.RemoteURL)
}

func TestServicePinnedAdoptionKeepsTheConfig(t *testing.T) {
	root := t.TempDir()
	pin := filepath.Join(t.TempDir(), "tool")
	makeClone(t, pin, "git@github.com:acme/old.git")
	setPushURL(t, pin, "no_push")
	config0, err := os.ReadFile(filepath.Join(pin, ".git", "config"))
	require.NoError(t, err)

	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "ssh", Concurrency: 1},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}}
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "new", Visibility: github.Public, SourceProfile: "p"}
	api := renamedRepository{snapshot: snap}
	var out, errw bytes.Buffer
	run := func(apply bool) []string {
		out.Reset()
		errw.Reset()
		svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
			Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Confirmer: api, Resolver: api,
			GitSync: noNetworkFetch{gitsvc.NewSync()}, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
		})
		require.NoError(t, err)
		require.NoError(t, svc.Run(context.Background(), apply), errw.String())
		return planKinds(t, out.Bytes())
	}

	assert.Equal(t, []string{"adopt acme/new", "fetch acme/new"}, run(true))
	config1, err := os.ReadFile(filepath.Join(pin, ".git", "config"))
	require.NoError(t, err)
	assert.Equal(t, string(config0), string(config1), "the first apply leaves the Git configuration byte-identical")

	assert.Equal(t, []string{"update_remote acme/new", "fetch acme/new"}, run(true))
	info, err := gitsvc.NewSync().Inspect(pin)
	require.NoError(t, err)
	assert.Equal(t, []string{"git@github.com:acme/new.git"}, info.Origins)
	assert.Equal(t, []string{"no_push"}, info.PushURLs)

	assert.Equal(t, []string{"fetch acme/new"}, run(false))
}
