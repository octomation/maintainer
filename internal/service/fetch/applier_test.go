package fetch_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

type fakeGit struct {
	cloned   []gitsvc.CloneOptions
	fetched  []string
	moved    [][2]string
	updated  [][2]string
	cloneErr error
}

func (f *fakeGit) Clone(_ context.Context, opt gitsvc.CloneOptions) error {
	if f.cloneErr != nil {
		return f.cloneErr
	}
	f.cloned = append(f.cloned, opt)
	return nil
}
func (f *fakeGit) Fetch(_ context.Context, path string, _ gitsvc.Auth) error {
	f.fetched = append(f.fetched, path)
	return nil
}
func (f *fakeGit) Move(from, to string) error {
	f.moved = append(f.moved, [2]string{from, to})
	return nil
}
func (f *fakeGit) UpdateRemote(path, url string) error {
	f.updated = append(f.updated, [2]string{path, url})
	return nil
}
func (f *fakeGit) Inspect(string) (gitsvc.CloneInfo, error) { return gitsvc.CloneInfo{}, nil }

var clock = func() time.Time { return time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC) }

func auth(profile, transport string) gitsvc.Auth {
	return gitsvc.Auth{Transport: transport, Token: "tok"}
}

func TestApplier_Clone(t *testing.T) {
	g := &fakeGit{}
	a := NewApplier(g, auth, clock)
	st := state.New()
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", SSHCloneURL: "git@github.com:acme/svc.git"}
	act := Action{
		Kind: KindClone, ID: 1, Owner: "acme", Name: "svc", Path: "/work/acme/svc",
		Transport: "ssh", Profile: "primary", Snapshot: &snap,
		Record: &state.Record{ID: 1, OwnerLogin: "acme", Name: "svc", Path: "/work/acme/svc"},
	}
	require.NoError(t, a.Execute(context.Background(), act, st))

	require.Len(t, g.cloned, 1)
	assert.Equal(t, "git@github.com:acme/svc.git", g.cloned[0].URL)
	assert.Equal(t, "/work/acme/svc", g.cloned[0].Path)
	rec, ok := st.ByID(1)
	require.True(t, ok)
	assert.Equal(t, clock(), rec.FirstSeen)
	assert.Equal(t, clock(), rec.LastApply)
}

func TestApplier_MoveWithUpdateRemote(t *testing.T) {
	g := &fakeGit{}
	a := NewApplier(g, auth, clock)
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "dotfiles", Path: "/work/acme/dotfiles"})
	act := Action{
		Kind: KindMove, ID: 1, Owner: "acme", Name: "configs",
		FromPath: "/work/acme/dotfiles", ToPath: "/work/acme/configs",
		UpdateRemote: true, RemoteURL: "git@github.com:acme/configs.git",
	}
	require.NoError(t, a.Execute(context.Background(), act, st))

	require.Len(t, g.moved, 1)
	assert.Equal(t, [2]string{"/work/acme/dotfiles", "/work/acme/configs"}, g.moved[0])
	require.Len(t, g.updated, 1)
	rec, _ := st.ByID(1)
	assert.Equal(t, "/work/acme/configs", rec.Path)
	assert.Equal(t, "configs", rec.Name)
	assert.Equal(t, "git@github.com:acme/configs.git", rec.RemoteURL)
}

func TestApplier_AdoptAndRelocateNoGitOps(t *testing.T) {
	g := &fakeGit{}
	a := NewApplier(g, auth, clock)
	st := state.New()

	adopt := Action{Kind: KindAdopt, ID: 7, Record: &state.Record{ID: 7, OwnerLogin: "acme", Name: "tool", Path: "/work/acme/tool"}}
	require.NoError(t, a.Execute(context.Background(), adopt, st))
	_, ok := st.ByID(7)
	assert.True(t, ok)

	st.Upsert(state.Record{ID: 8, Path: "/work/old"})
	reloc := Action{Kind: KindRelocate, ID: 8, FromPath: "/work/old", ToPath: "/work/new"}
	require.NoError(t, a.Execute(context.Background(), reloc, st))
	rec, _ := st.ByID(8)
	assert.Equal(t, "/work/new", rec.Path)

	assert.Empty(t, g.cloned)
	assert.Empty(t, g.moved)
}

func TestApplier_CloneErrorPropagates(t *testing.T) {
	g := &fakeGit{cloneErr: errors.New("boom")}
	a := NewApplier(g, auth, clock)
	st := state.New()
	act := Action{Kind: KindClone, ID: 1, Snapshot: &github.RepoSnapshot{ID: 1}, Record: &state.Record{ID: 1}, Transport: "https"}
	err := a.Execute(context.Background(), act, st)
	require.Error(t, err)
	_, ok := st.ByID(1)
	assert.False(t, ok) // no record written on failure
}

func TestApplier_ReportOnlyKindsAreNoops(t *testing.T) {
	g := &fakeGit{}
	a := NewApplier(g, auth, clock)
	st := state.New()
	for _, k := range []Kind{KindOrphan, KindNoop, KindConflict} {
		require.NoError(t, a.Execute(context.Background(), Action{Kind: k, ID: 1}, st))
	}
	assert.Empty(t, g.cloned)
	assert.Empty(t, g.fetched)
}
