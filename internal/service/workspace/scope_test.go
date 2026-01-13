package workspace

import (
	"context"
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
		}))
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
	require.NoError(t, s.Walk(context.Background(), func(path string, pinned bool) error { assert.True(t, pinned); got = append(got, path); return nil }))
	want := []string{filepath.Join(home, "code/repo"), filepath.Join(root, "absent")}
	sort.Strings(got)
	sort.Strings(want)
	assert.Equal(t, want, got)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, s.Walk(ctx, func(string, bool) error { t.Fatal("visited after cancellation"); return nil }), context.Canceled)
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
	require.NoError(t, s.Walk(context.Background(), func(path string, _ bool) error { paths = append(paths, path); return nil }))
	assert.Equal(t, []string{filepath.Join(root, "mirror/acme/tool")}, paths)
}
