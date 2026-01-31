package workspace

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
)

func TestBoundedWalkAndPins(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{"public/a/tool", "private/b/secret", "internal/c/corp", "research/x/unrelated", "prototyping/deep/local", "public/a/tool/submodule", "public/a/extra/too/deep"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, path, ".git"), 0o755))
	}
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Path: config.DefaultPath}}
	collect := func() map[string]bool {
		s, err := New(cnf, root, root)
		require.NoError(t, err)
		got := map[string]bool{}
		require.NoError(t, s.Walk(context.Background(), func(path string, pin bool) error {
			rel, err := filepath.Rel(root, path)
			require.NoError(t, err)
			_, duplicate := got[rel]
			assert.False(t, duplicate)
			got[rel] = pin
			return nil
		}, unexpected(t)))
		return got
	}
	want := map[string]bool{"public/a/tool": false, "private/b/secret": false, "internal/c/corp": false}
	assert.Equal(t, want, collect())
	// Arbitrary root mutations and an unreadable unrelated directory do not
	// change discovery; chmod would make a broad walk fail for a normal user.
	require.NoError(t, os.Mkdir(filepath.Join(root, "future"), 0o000))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, "future"), 0o700) })
	assert.Equal(t, want, collect())
	cnf.Workspace.Pins = []string{"prototyping", "prototyping", "prototyping/deep", "public/a/tool"}
	want["prototyping/deep/local"], want["public/a/tool"] = true, true
	assert.Equal(t, want, collect())
}

func TestExternalAndMissingPins(t *testing.T) {
	root, home := t.TempDir(), t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(home, "code/repo/.git"), 0o755))
	cnf := &config.Fetch{Workspace: &config.Workspace{Pins: []string{"~/code", "absent"}}}
	s, err := New(cnf, root, home)
	require.NoError(t, err)
	var got []string
	require.NoError(t, s.Walk(context.Background(), func(path string, pinned bool) error { assert.True(t, pinned); got = append(got, path); return nil }, unexpected(t)))
	want := []string{filepath.Join(home, "code/repo"), filepath.Join(root, "absent")}
	sort.Strings(got)
	sort.Strings(want)
	assert.Equal(t, want, got)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, s.Walk(ctx, func(string, bool) error { t.Fatal("visited after cancellation"); return nil }, unexpected(t)), context.Canceled)
}

func TestOwnerOverridesAndSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(outside, "tool/.git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "public"), 0o755))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "public/alice")))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "mirror/acme/tool/.git"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "mirror/other/tool/.git"), 0o755))
	s, err := New(&config.Fetch{Owners: []config.Owner{{Name: "acme", Path: "mirror/{{.Owner}}/{{.Repo}}"}}}, root, root)
	require.NoError(t, err)
	var paths []string
	require.NoError(t, s.Walk(context.Background(), func(path string, _ bool) error { paths = append(paths, path); return nil }, unexpected(t)))
	assert.Equal(t, []string{filepath.Join(root, "mirror/acme/tool")}, paths)
}

func unexpected(t *testing.T) func(string, error) error {
	return func(path string, reason error) error {
		t.Errorf("unexpected obstruction %s: %v", path, reason)
		return nil
	}
}

// walk collects checkouts and obstructions relative to root.
func walk(t *testing.T, s *Scope, root string) (found []string, obstructions map[string]error) {
	t.Helper()
	obstructions = map[string]error{}
	rel := func(path string) string {
		r, err := filepath.Rel(root, path)
		require.NoError(t, err)
		return r
	}
	require.NoError(t, s.Walk(context.Background(), func(path string, _ bool) error {
		found = append(found, rel(path))
		return nil
	}, func(path string, reason error) error {
		_, duplicate := obstructions[rel(path)]
		assert.False(t, duplicate, path)
		obstructions[rel(path)] = reason
		return nil
	}))
	sort.Strings(found)
	return found, obstructions
}

func TestUnreadableDirectoriesAreObstructions(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permissions do not restrict root")
	}
	root := t.TempDir()
	for _, path := range []string{"prototyping/app/pgdata", "prototyping/app/listing", "prototyping/deep/tool/.git", "public/a/tool/.git", "public/b"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, path), 0o755))
	}
	// pgdata cannot even be searched for .git, listing can be searched but
	// not read, and public/b blocks an owner level of the layout.
	for path, mode := range map[string]os.FileMode{"prototyping/app/pgdata": 0o000, "prototyping/app/listing": 0o300, "public/b": 0o000} {
		require.NoError(t, os.Chmod(filepath.Join(root, path), mode))
		t.Cleanup(func() { _ = os.Chmod(filepath.Join(root, path), 0o700) })
	}
	cnf := &config.Fetch{Workspace: &config.Workspace{Root: root, Path: config.DefaultPath, Pins: []string{"prototyping", "prototyping/app"}}}
	s, err := New(cnf, root, root)
	require.NoError(t, err)
	found, obstructions := walk(t, s, root)
	assert.Equal(t, []string{"prototyping/deep/tool", "public/a/tool"}, found)
	require.Len(t, obstructions, 3)
	for path := range obstructions {
		assert.ErrorIs(t, obstructions[path], fs.ErrPermission, path)
		assert.Contains(t, obstructions[path].Error(), filepath.Join(root, path))
	}
}

func TestPrefixLevelCheckoutsDoNotHideCheckoutsBelow(t *testing.T) {
	root := t.TempDir()
	// public/p/.git/.git would look like a checkout if Git metadata were walked.
	for _, path := range []string{".git", "public/o/.git", "public/o/r/.git", "public/p/.git/.git", "research/x/.git"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, path), 0o755))
	}
	s, err := New(&config.Fetch{Workspace: &config.Workspace{Root: root, Path: config.DefaultPath}}, root, root)
	require.NoError(t, err)
	found, obstructions := walk(t, s, root)
	assert.Equal(t, []string{"public/o/r"}, found)
	assert.Equal(t, map[string]error{".": ErrUnmanagedCheckout, "public/o": ErrUnmanagedCheckout, "public/p": ErrUnmanagedCheckout}, obstructions)
}

func TestDestination(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "real/owner"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "file"), nil, 0o644))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "linked")))
	require.NoError(t, os.Symlink(filepath.Join(root, "real"), filepath.Join(root, "inner")))
	for path, want := range map[string]error{
		"real/owner/repo":     nil,
		"real/new/owner/repo": nil,
		"absent/repo":         nil,
		"real/owner":          fs.ErrExist,
		"linked":              fs.ErrExist,
		"linked/repo":         ErrSymlinkParent,
		"linked/owner/repo":   ErrSymlinkParent,
		"inner/owner/repo":    ErrSymlinkParent,
		"file/repo":           ErrFileParent,
		"file/owner/repo":     ErrFileParent,
	} {
		err := Destination(root, filepath.Join(root, path))
		if want == nil {
			assert.NoError(t, err, path)
			continue
		}
		assert.ErrorIs(t, err, want, path)
	}
	for _, path := range []string{root, filepath.Dir(root), filepath.Join(outside, "repo"), root + "-sibling/repo"} {
		assert.ErrorContains(t, Destination(root, path), "outside the workspace root", path)
	}
	entries, err := os.ReadDir(outside)
	require.NoError(t, err)
	assert.Empty(t, entries, "a check never creates anything")
}
