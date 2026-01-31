package fetch_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gogit "github.com/go-git/go-git/v5"
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

// digest records every entry below dir with its mode and content, so a test
// can prove that a tree is byte-identical before and after a run.
func digest(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	require.NoError(t, filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		require.NoError(t, err)
		rel, err := filepath.Rel(dir, path)
		require.NoError(t, err)
		info, err := d.Info()
		require.NoError(t, err)
		entry := info.Mode().String()
		switch {
		case info.Mode()&fs.ModeSymlink != 0:
			target, err := os.Readlink(path)
			require.NoError(t, err)
			entry += " -> " + target
		case info.Mode().IsRegular():
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			entry += fmt.Sprintf(" %x", sha256.Sum256(data))
		}
		out[rel] = entry
		return nil
	}))
	return out
}

type planJSON struct {
	Actions []struct {
		Kind   string `json:"kind"`
		ID     int64  `json:"id"`
		Path   string `json:"path"`
		Reason string `json:"reason"`
	} `json:"actions"`
	Summary Summary `json:"summary"`
}

func decodePlan(t *testing.T, out *bytes.Buffer) planJSON {
	t.Helper()
	var plan planJSON
	require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
	return plan
}

// hookWriter runs fn once, when the service announces the apply phase: it
// changes the disk between planning and the first mutation.
type hookWriter struct {
	bytes.Buffer
	fn func()
}

func (w *hookWriter) Write(p []byte) (int, error) {
	if w.fn != nil && strings.HasPrefix(string(p), "applying ") {
		w.fn()
		w.fn = nil
	}
	return w.Buffer.Write(p)
}

func boundaryService(t *testing.T, root string, snap github.RepoSnapshot, store *state.Store, git gitsvc.GitSync, out *bytes.Buffer, errw *hookWriter) *Service {
	t.Helper()
	cnf := &config.Fetch{
		Workspace: &config.Workspace{Root: root, Path: "{{.Owner}}/{{.Repo}}"},
		Defaults:  config.Defaults{CloneURL: "https", Concurrency: 1},
	}
	cnf.Profiles = map[string]config.Profile{}
	require.NoError(t, cnf.Validate())
	api := renamedRepository{snapshot: snap}
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{snap.Owner}}}, root, root, 1, Deps{
		Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}},
		Confirmer: api, Resolver: api, GitSync: git,
		Reporter: NewReporter(out, errw, FormatJSON, 1, false),
	})
	require.NoError(t, err)
	return svc
}

func requirePartial(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var coded exitpkg.Coder
	require.True(t, errors.As(err, &coded), err)
	assert.Equal(t, exitpkg.Partial, coded.ExitCode())
}

func TestBoundary_CloneThroughSymlinkParentIsConflict(t *testing.T) {
	origin := originWithCommit(t)
	root, outside := t.TempDir(), t.TempDir()
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "acme")))
	target := filepath.Join(root, "acme/svc")
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public, HTTPSCloneURL: origin, SourceProfile: "p"}
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	git := localOriginSync{Sync: gitsvc.NewSync(), localURL: origin, canonicalURL: "https://github.com/acme/svc.git"}
	before := digest(t, outside)

	var out bytes.Buffer
	var errw hookWriter
	svc := boundaryService(t, root, snap, store, git, &out, &errw)
	require.NoError(t, svc.Run(context.Background(), false))
	plan := decodePlan(t, &out)
	require.Len(t, plan.Actions, 1)
	assert.Equal(t, "conflict", plan.Actions[0].Kind)
	assert.Equal(t, target, plan.Actions[0].Path)
	assert.Contains(t, plan.Actions[0].Reason, "symbolic link")
	assert.Equal(t, 1, plan.Summary.Conflict)

	out.Reset()
	requirePartial(t, svc.Run(context.Background(), true))
	assert.Equal(t, 0, decodePlan(t, &out).Summary.Clone)
	assert.Equal(t, before, digest(t, outside), "nothing is written outside the root")
	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Empty(t, loaded.Repos)
}

func TestBoundary_MoveThroughSymlinkParentIsConflict(t *testing.T) {
	origin := originWithCommit(t)
	root, outside := t.TempDir(), t.TempDir()
	source := filepath.Join(root, "old-owner/svc")
	_, err := gogit.PlainClone(source, false, &gogit.CloneOptions{URL: origin})
	require.NoError(t, err)
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "acme")))
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	st.Upsert(state.Record{
		ID: 1, OwnerLogin: "old-owner", Name: "svc", Path: source, Visibility: "public",
		RemoteURL: "https://github.com/old-owner/svc.git", CloneURL: "https", SourceProfile: "p",
	})
	require.NoError(t, store.Save(st))
	// The repository was transferred: only the new destination's parent is a link.
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public, HTTPSCloneURL: origin, SourceProfile: "p"}
	git := localOriginSync{Sync: gitsvc.NewSync(), localURL: origin, canonicalURL: "https://github.com/old-owner/svc.git"}
	sourceBefore, outsideBefore := digest(t, source), digest(t, outside)

	var out bytes.Buffer
	var errw hookWriter
	svc := boundaryService(t, root, snap, store, git, &out, &errw)
	require.NoError(t, svc.Run(context.Background(), false))
	plan := decodePlan(t, &out)
	require.Len(t, plan.Actions, 1)
	assert.Equal(t, "conflict", plan.Actions[0].Kind)
	assert.Equal(t, filepath.Join(root, "acme/svc"), plan.Actions[0].Path)
	assert.Contains(t, plan.Actions[0].Reason, "symbolic link")

	out.Reset()
	requirePartial(t, svc.Run(context.Background(), true))
	assert.Equal(t, 0, decodePlan(t, &out).Summary.Move)
	assert.Equal(t, sourceBefore, digest(t, source), "the source checkout is untouched")
	assert.Equal(t, outsideBefore, digest(t, outside), "nothing is written outside the root")
	loaded, err := store.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Repos, 1)
	assert.Equal(t, source, loaded.Repos[0].Path)
	assert.Equal(t, "old-owner", loaded.Repos[0].OwnerLogin)
}

// A symlink that appears after the plan is refused at the mutation point.
func TestBoundary_ApplyChecksTheDestinationAgain(t *testing.T) {
	origin := originWithCommit(t)
	root, outside := t.TempDir(), t.TempDir()
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public, HTTPSCloneURL: origin, SourceProfile: "p"}
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	git := localOriginSync{Sync: gitsvc.NewSync(), localURL: origin, canonicalURL: "https://github.com/acme/svc.git"}

	var out bytes.Buffer
	errw := hookWriter{fn: func() { require.NoError(t, os.Symlink(outside, filepath.Join(root, "acme"))) }}
	svc := boundaryService(t, root, snap, store, git, &out, &errw)
	requirePartial(t, svc.Run(context.Background(), true))
	plan := decodePlan(t, &out)
	assert.Equal(t, 1, plan.Summary.Clone, "the plan saw a clear destination")
	assert.Equal(t, 1, plan.Summary.Errors)
	assert.Contains(t, errw.String(), "refusing to clone")
	assert.Contains(t, errw.String(), "symbolic link")
	entries, err := os.ReadDir(outside)
	require.NoError(t, err)
	assert.Empty(t, entries, "nothing is written outside the root")
}

func TestBoundary_ObstructionsAreConflicts(t *testing.T) {
	root := t.TempDir()
	_, err := gogit.PlainInit(filepath.Join(root, "public/o"), false)
	require.NoError(t, err)
	makeClone(t, filepath.Join(root, "public/o/r"), "git@github.com:o/r.git")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "ssh", Concurrency: 1}}
	require.NoError(t, cnf.Validate())
	snap := github.RepoSnapshot{ID: 1, Owner: "o", Name: "r", Visibility: github.Public, SourceProfile: "p"}
	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"o"}}}, root, root, 1, Deps{
		Store:      state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil),
		Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: fakeGitHubNameResolver{},
		GitSync: gitsvc.NewSync(), Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	plan := decodePlan(t, &out)
	require.Len(t, plan.Actions, 3, out.String())
	assert.Equal(t, "adopt", plan.Actions[0].Kind, "the checkout below the obstruction stays visible")
	assert.Equal(t, filepath.Join(root, "public/o/r"), plan.Actions[0].Path)
	assert.Equal(t, "fetch", plan.Actions[1].Kind, "the adoption is refreshed")
	assert.Equal(t, filepath.Join(root, "public/o/r"), plan.Actions[1].Path)
	assert.Equal(t, "conflict", plan.Actions[2].Kind)
	assert.Equal(t, filepath.Join(root, "public/o"), plan.Actions[2].Path)
	assert.Equal(t, "scan: a Git repository at an unmanaged level of the layout", plan.Actions[2].Reason)
	assert.Equal(t, 1, plan.Summary.Conflict)
}

func TestPlanner_NestedTargetsConflict(t *testing.T) {
	// Validate rejects these layouts; the planner still refuses to nest.
	cnf := &config.Fetch{
		Defaults: config.Defaults{Root: "/work", Path: "{{.Owner}}/{{.Repo}}", CloneURL: "ssh", Concurrency: 1},
		Owners:   []config.Owner{{Name: "acme", Path: "{{.Repo}}"}, {Name: "zed", Path: "acme/bob/{{.Repo}}"}},
	}
	snaps := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "bob", Visibility: github.Public},
		{ID: 2, Owner: "bob", Name: "x", Visibility: github.Public},
	}
	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: state.New(), Snapshots: snaps})
	require.NoError(t, err)
	require.Len(t, acts, 2)
	assert.Equal(t, KindConflict, acts[0].Kind)
	assert.Equal(t, `target path overlaps the target path "/work/bob/x" of bob/x`, acts[0].Reason)
	assert.Equal(t, KindConflict, acts[1].Kind)
	assert.Equal(t, `target path overlaps the target path "/work/bob" of acme/bob`, acts[1].Reason)

	// A tracked checkout and a destination inside it are a conflict for both.
	st := state.New()
	st.Upsert(state.Record{ID: 3, OwnerLogin: "acme", Name: "bob", Path: "/work/acme/bob", RemoteURL: "git@github.com:acme/bob.git", CloneURL: "ssh"})
	cnf.Owners = cnf.Owners[1:]
	snaps = []github.RepoSnapshot{
		{ID: 3, Owner: "acme", Name: "bob", Visibility: github.Public},
		{ID: 4, Owner: "zed", Name: "x", Visibility: github.Public},
	}
	clones := []DiskClone{{ID: 3, Path: "/work/acme/bob", RemoteURL: "git@github.com:acme/bob.git", Transport: "ssh", Owner: "acme", Name: "bob", Origins: 1}}
	acts, err = newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: snaps, Clones: clones})
	require.NoError(t, err)
	require.Len(t, acts, 2)
	assert.Equal(t, KindConflict, acts[0].Kind)
	assert.Equal(t, `checkout overlaps the target path "/work/acme/bob/x" of zed/x`, acts[0].Reason)
	assert.Equal(t, KindConflict, acts[1].Kind)
	assert.Equal(t, `target path overlaps the existing checkout "/work/acme/bob"`, acts[1].Reason)

	// A checkout at root contains the whole workspace: it is not a collision.
	cnf.Owners = nil
	acts, err = newPlanner(t, cnf).Plan(PlanInput{State: state.New(), Snapshots: snaps[1:], Clones: []DiskClone{{Path: "/work", Error: "scan: a Git repository at an unmanaged level of the layout"}}})
	require.NoError(t, err)
	require.Len(t, acts, 2)
	assert.Equal(t, KindClone, acts[0].Kind)
	assert.Equal(t, KindConflict, acts[1].Kind)
	assert.Equal(t, "/work", acts[1].Path)
}

// An unreadable owner directory is a scan obstruction, and a target below it
// is a conflict for that repository; the rest of the plan goes on.
func TestBoundary_UnreadableTargetParentIsOneConflict(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permissions do not restrict root")
	}
	root := t.TempDir()
	locked := filepath.Join(root, "public/acme")
	require.NoError(t, os.MkdirAll(locked, 0o755))
	require.NoError(t, os.Chmod(locked, 0o000))
	t.Cleanup(func() { _ = os.Chmod(locked, 0o700) })
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
	require.NoError(t, cnf.Validate())
	snaps := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public, SourceProfile: "p"},
		{ID: 2, Owner: "bob", Name: "lib", Visibility: github.Public, SourceProfile: "p"},
	}
	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme", "bob"}}}, root, root, 1, Deps{
		Store:      state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil),
		Discoverer: fakeDiscoverer{snaps: snaps}, Resolver: fakeGitHubNameResolver{},
		GitSync: gitsvc.NewSync(), Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	plan := decodePlan(t, &out)
	require.Len(t, plan.Actions, 3, out.String())
	assert.Equal(t, "clone", plan.Actions[0].Kind)
	assert.Equal(t, int64(2), plan.Actions[0].ID)
	assert.Equal(t, "conflict", plan.Actions[1].Kind)
	assert.Equal(t, locked, plan.Actions[1].Path)
	assert.Contains(t, plan.Actions[1].Reason, "scan: lstat")
	assert.Equal(t, "conflict", plan.Actions[2].Kind)
	assert.Equal(t, int64(1), plan.Actions[2].ID)
	assert.Equal(t, "the target path cannot be inspected; refusing to write into it", plan.Actions[2].Reason)
}

func TestPlanner_RefusedTargetStillConflictsWithNestedCheckout(t *testing.T) {
	const pin = "/work/acme/foo/bar"
	cnf := &config.Fetch{
		Defaults: config.Defaults{Root: "/work", Path: "{{.Owner}}/{{.Repo}}", CloneURL: "ssh", Concurrency: 1},
		Repos:    []config.Repo{{Match: config.RepoMatch{ID: 2}, Path: pin}},
	}
	st := state.New()
	st.Upsert(state.Record{ID: 2, OwnerLogin: "acme", Name: "bar", Path: pin, PinnedPath: pin, RemoteURL: "git@github.com:acme/bar.git", CloneURL: "ssh"})
	acts, err := newPlanner(t, cnf).Plan(PlanInput{
		State: st,
		Snapshots: []github.RepoSnapshot{
			{ID: 1, Owner: "acme", Name: "foo", Visibility: github.Public},
			{ID: 2, Owner: "acme", Name: "bar", Visibility: github.Public},
		},
		Clones:    []DiskClone{{ID: 2, Path: pin, RemoteURL: "git@github.com:acme/bar.git", Transport: "ssh", Owner: "acme", Name: "bar", Origins: 1, Pinned: true}},
		Occupancy: map[string]Occupancy{"/work/acme/foo": OccupancyForeign},
	})
	require.NoError(t, err)
	require.Len(t, acts, 2)
	assert.Equal(t, KindConflict, acts[0].Kind)
	assert.Contains(t, acts[0].Reason, "already exists", "the first reason stays")
	assert.Equal(t, KindConflict, acts[1].Kind)
	assert.Equal(t, `checkout overlaps the target path "/work/acme/foo" of acme/foo`, acts[1].Reason)
}
