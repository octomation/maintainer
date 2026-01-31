package status

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func TestWorkspaceBroadPinDuplicatesAndExplicitSelection(t *testing.T) {
	root := t.TempDir()
	managed := checkout(t, filepath.Join(root, "public/acme"))
	pin := checkout(t, filepath.Join(root, "prototyping"))
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Pins: []string{"prototyping"}}}
	rows, err := Collect(context.Background(), cnf, state.New(), root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	for _, row := range rows {
		assert.Contains(t, row.Error, "multiple checkouts")
	}
	cnf.Repos = []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: pin}}
	rows, err = Collect(context.Background(), cnf, state.New(), root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, managed, rows[1].Path)
	assert.Equal(t, "public/acme/repo", rows[1].displayPath)
	assert.Equal(t, "duplicate-pin", rows[1].OrphanReason)
	assert.Equal(t, pin, rows[0].Path)
	assert.Equal(t, "prototyping/repo", rows[0].displayPath)
	assert.NotEqual(t, managed, rows[0].Path)
	assert.True(t, rows[0].Pinned)
	assert.Empty(t, rows[0].Error)
}

func TestWorkspaceRemovedPinDoesNotReenterThroughStateOrTemplate(t *testing.T) {
	root := t.TempDir()
	path := checkout(t, filepath.Join(root, "public/acme"))
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: path, PinnedPath: path, PinSource: "workspace"})
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}}
	rows, err := Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "out-of-scope", rows[0].OrphanReason)
	assert.False(t, rows[0].Pinned)
	cnf.Workspace.Pins = []string{"public/acme"}
	rows, err = Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, rows[0].Pinned)
	assert.Equal(t, "workspace", st.Repos[0].PinSource, "status must not mutate state")
}

func TestWorkspaceMissingPinIsVisible(t *testing.T) {
	root := t.TempDir()
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Pins: []string{"missing"}}}
	rows, err := Collect(context.Background(), cnf, state.New(), root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.True(t, rows[0].Pinned)
	assert.Equal(t, "error", rows[0].Status)
	assert.Equal(t, "missing", rows[0].displayPath)
}

func TestWorkspacePinStaleCopyIsNotDuplicate(t *testing.T) {
	root := t.TempDir()
	old := checkout(t, filepath.Join(root, "public/acme"))
	pin := checkout(t, filepath.Join(root, "prototyping"))
	write(t, filepath.Join(old, "text"), "stale work\n")
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Path: pin, PinSource: "workspace", PreviousPaths: []string{old}})
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Pins: []string{"prototyping"}}}
	rows, err := Collect(context.Background(), cnf, st, root, root, nil, 2)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, pin, rows[0].Path)
	assert.True(t, rows[0].Pinned)
	assert.Equal(t, "synced", rows[0].Status)
	assert.Empty(t, rows[0].Error)
	assert.Equal(t, old, rows[1].Path)
	assert.Equal(t, "duplicate-pin", rows[1].OrphanReason)
	assert.Equal(t, pin, rows[1].ActivePath)
	assert.Equal(t, "synced", rows[1].Status)
	assert.Equal(t, 1, rows[1].Added)
	assert.Empty(t, rows[1].Error)
}

func TestWorkspaceObstructionsAreErrorRows(t *testing.T) {
	root := t.TempDir()
	repo := checkout(t, filepath.Join(root, "public/o"))
	runGit(t, filepath.Join(root, "public/o"), "init", "-q")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}}
	rows, err := Collect(context.Background(), cnf, state.New(), root, root, []string{"acme"}, 1)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, repo, rows[0].Path, "the checkout below the obstruction stays visible")
	assert.Empty(t, rows[0].Error)
	assert.Equal(t, filepath.Join(root, "public/o"), rows[1].Path)
	assert.Equal(t, "public/o", rows[1].displayPath)
	assert.Equal(t, "error", rows[1].Status)
	assert.Equal(t, "scan: a Git repository at an unmanaged level of the layout", rows[1].Error)

	// A remembered checkout at a prefix level keeps its row and the error.
	st := state.New()
	st.Upsert(state.Record{ID: 9, OwnerLogin: "acme", Name: "o", Path: filepath.Join(root, "public/o")})
	rows, err = Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, filepath.Join(root, "public/o"), rows[0].Path)
	assert.Equal(t, "out-of-scope", rows[0].OrphanReason)
	assert.Equal(t, "error", rows[0].Status)
	assert.Equal(t, "scan: a Git repository at an unmanaged level of the layout", rows[0].Error)
	// The same identity below it is no duplicate of the out-of-scope copy;
	// the scan reason stays.
	st.Repos[0].Name = "tool"
	rows, err = Collect(context.Background(), cnf, st, root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, filepath.Join(root, "public/o"), rows[0].Path)
	assert.Equal(t, "scan: a Git repository at an unmanaged level of the layout", rows[0].Error)
	assert.Empty(t, rows[1].Error, "an orphan copy makes no duplicates")

	if os.Geteuid() == 0 {
		return // permissions do not restrict root
	}
	pgdata := filepath.Join(root, "prototyping/app/pgdata")
	require.NoError(t, os.MkdirAll(pgdata, 0o755))
	require.NoError(t, os.Chmod(pgdata, 0o000))
	t.Cleanup(func() { _ = os.Chmod(pgdata, 0o700) })
	pinned := checkout(t, filepath.Join(root, "prototyping/tool"))
	runGit(t, pinned, "remote", "set-url", "origin", "https://github.com/acme/pinned.git")
	cnf.Workspace.Pins = []string{"prototyping"}
	rows, err = Collect(context.Background(), cnf, state.New(), root, root, nil, 1)
	require.NoError(t, err)
	paths := map[string]Row{}
	for _, row := range rows {
		paths[row.Path] = row
	}
	require.Len(t, paths, 4)
	assert.True(t, paths[pinned].Pinned)
	assert.Empty(t, paths[pinned].Error)
	assert.Equal(t, "error", paths[pgdata].Status)
	assert.Equal(t, "scan: lstat "+filepath.Join(pgdata, ".git")+": permission denied", paths[pgdata].Error)
}

func TestWorkspaceTemplatedPinsNeedSufficientContext(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	ext := filepath.Join(home, "ext")
	// The path that empty fields would collapse to is a real repository.
	runGit(t, home, "init", "-q", "ext")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 42}, Path: ext + "/{{.Owner}}/{{.Repo}}"}}}
	rows, err := Collect(context.Background(), cnf, state.New(), home, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, StatusUnresolvedPin, rows[0].Status)
	assert.Equal(t, "id=42", rows[0].Repository)
	assert.Equal(t, ext+"/{{.Owner}}/{{.Repo}}", rows[0].Path)
	assert.Empty(t, rows[0].Error, "an unresolved pin is not a failure")
	assert.Empty(t, rows[0].Branch, "nothing is inspected in its place")
	assert.True(t, rows[0].Pinned)
	// A record without metadata, as an early fetch saved it, does not help.
	legacy := state.New()
	legacy.Upsert(state.Record{ID: 42, OwnerLogin: "acme", Name: "repo", Path: filepath.Join(ext, "old")})
	cnf.Repos[0].Path = ext + "/{{.Visibility}}/{{.Repo}}"
	rows, err = Collect(context.Background(), cnf, legacy, home, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, StatusUnresolvedPin, rows[0].Status)
	assert.Equal(t, "acme/repo", rows[0].Repository)

	// A state record supplies the fields; a name-based rule supplies its own.
	byOwner, byVisibility := checkout(t, filepath.Join(ext, "acme")), checkout(t, filepath.Join(ext, "public"))
	for _, tc := range []struct {
		path, pinned string
		match        config.RepoMatch
		recorded     bool
	}{
		{ext + "/{{.Owner}}/{{.Repo}}", byOwner, config.RepoMatch{ID: 42}, true},
		{ext + "/{{.Owner}}/{{.Repo}}", byOwner, config.RepoMatch{Owner: "acme", Name: "repo"}, false},
		{ext + "/{{.Visibility}}/{{.Repo}}", byVisibility, config.RepoMatch{Owner: "acme", Name: "repo"}, true},
	} {
		runGit(t, tc.pinned, "remote", "set-url", "origin", "https://github.com/acme/repo.git")
		st := state.New()
		if tc.recorded {
			st.Upsert(state.Record{ID: 42, OwnerLogin: "acme", Name: "repo", Visibility: "public", Path: tc.pinned, PinnedPath: tc.pinned})
		}
		cnf.Repos[0].Match, cnf.Repos[0].Path = tc.match, tc.path
		rows, err = Collect(context.Background(), cnf, st, home, root, nil, 1)
		require.NoError(t, err)
		require.Len(t, rows, 1, tc.path)
		assert.Equal(t, tc.pinned, rows[0].Path)
		assert.Equal(t, "main", rows[0].Branch)
		assert.True(t, rows[0].Pinned)
		assert.Empty(t, rows[0].Error)
	}
}

// A rule whose path does not parse is an error of its repository only: the
// other checkouts are inspected, nothing is inspected in its place.
func TestWorkspaceBrokenPinTemplateIsErrorRow(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	tool := checkout(t, filepath.Join(root, "public/acme"))
	broken := checkout(t, filepath.Join(home, "ext/acme"))
	runGit(t, broken, "remote", "set-url", "origin", "https://github.com/acme/broken.git")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root},
		Repos: []config.Repo{{Match: config.RepoMatch{Owner: "acme", Name: "broken"}, Path: "~/ext/{{.Owner"}}}
	recorded := state.New()
	recorded.Upsert(state.Record{ID: 42, OwnerLogin: "acme", Name: "broken", Visibility: "public", Path: broken, PinnedPath: broken})
	for _, st := range []*state.State{state.New(), recorded} {
		rows, err := Collect(context.Background(), cnf, st, home, root, nil, 1)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, "acme/broken", rows[0].Repository)
		assert.Equal(t, "~/ext/{{.Owner", rows[0].Path)
		assert.Equal(t, "error", rows[0].Status)
		assert.Equal(t, `parse path template "~/ext/{{.Owner": template: path:1: unclosed action`, rows[0].Error)
		assert.Empty(t, rows[0].Branch, "nothing is inspected in its place")
		assert.Equal(t, tool, rows[1].Path)
		assert.Equal(t, "main", rows[1].Branch)
		assert.Empty(t, rows[1].Error)
	}
	// Only the applicable rule reports it: an id rule wins over a name rule.
	byID := config.Repo{Match: config.RepoMatch{ID: 42}, Path: "~/ext/{{.Owner"}
	cnf.Repos = []config.Repo{{Match: config.RepoMatch{Owner: "acme", Name: "broken"}, Path: broken}, byID}
	rows, err := Collect(context.Background(), cnf, recorded, home, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "~/ext/{{.Owner", rows[0].Path)
	assert.Equal(t, "error", rows[0].Status)
	assert.Equal(t, tool, rows[1].Path)
}

// A scan obstruction and an unresolved pin are rows, never checkouts: the
// header counts only what can be inspected, and the obstruction as an error.
func TestWorkspaceSummaryCountsOnlyCheckouts(t *testing.T) {
	root := t.TempDir()
	checkout(t, filepath.Join(root, "public/o"))
	runGit(t, filepath.Join(root, "public/o"), "init", "-q")
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 42}, Path: root + "/pins/{{.Owner}}/{{.Repo}}"}}}
	rows, err := Collect(context.Background(), cnf, state.New(), root, root, nil, 1)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	var out bytes.Buffer
	require.NoError(t, Plain(&out, rows))
	assert.Equal(t, "Repository status · 1 checkout · 0 uncommitted · 0 ahead · 0 behind · 0 locked · 0 orphan · 1 error · local refs (no fetch)",
		strings.SplitN(out.String(), "\n", 2)[0])
}
