package skill_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/skill"
)

func release(t *testing.T, version, text string) (fstest.MapFS, Marker) {
	t.Helper()
	fsys := fstest.MapFS{
		"SKILL.md":        {Data: []byte(text)},
		"references/a.md": {Data: []byte("a")},
	}
	digest, err := Digest(fsys)
	require.NoError(t, err)
	return fsys, Marker{Version: version, Commit: "abc", Digest: digest}
}

func refused(t *testing.T, err error, reason string) {
	t.Helper()
	var r *RefusedError
	require.True(t, errors.As(err, &r), "want a refusal, got %v", err)
	assert.Contains(t, r.Reason, reason)
}

func TestExport(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "skills", "maintainer")
	v1, m1 := release(t, "0.2.0", "one\n")
	v2, m2 := release(t, "0.3.0", "two\n")

	outcome, err := Export(v1, dir, m1, false)
	require.NoError(t, err)
	assert.Equal(t, Installed, outcome)
	marker, err := ReadMarker(dir)
	require.NoError(t, err)
	assert.Equal(t, m1, marker)
	digest, err := CopyDigest(os.DirFS(dir))
	require.NoError(t, err)
	assert.Equal(t, m1.Digest, digest)

	outcome, err = Export(v1, dir, m1, true)
	require.NoError(t, err)
	assert.Equal(t, Unchanged, outcome)

	_, err = Export(v2, dir, m2, false)
	refused(t, err, "pass --replace to replace it with 0.3.0")

	outcome, err = Export(v2, dir, m2, true)
	require.NoError(t, err)
	assert.Equal(t, Replaced, outcome)
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	require.NoError(t, err)
	assert.Equal(t, "two\n", string(data))
	entries, err := os.ReadDir(filepath.Dir(dir))
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temporary directories are left behind")

	// Finder litter doesn't count as a change.
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".DS_Store"), []byte("x"), 0o644))
	outcome, err = Export(v2, dir, m2, false)
	require.NoError(t, err)
	assert.Equal(t, Unchanged, outcome)

	// Any other dotfile is the user's, and replacing the copy would delete it.
	git := filepath.Join(dir, ".git", "config")
	require.NoError(t, os.MkdirAll(filepath.Dir(git), 0o755))
	require.NoError(t, os.WriteFile(git, []byte("x"), 0o644))
	_, err = Export(v1, dir, m1, true)
	refused(t, err, "was changed after maintainer installed it")
	assert.FileExists(t, git)
	require.NoError(t, os.RemoveAll(filepath.Dir(git)))

	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("edited\n"), 0o644))
	_, err = Export(v1, dir, m1, true)
	refused(t, err, "was changed after maintainer installed it")
}

func TestExport_leavesOthersAlone(t *testing.T) {
	v1, m1 := release(t, "0.2.0", "one\n")

	foreign := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(foreign, "SKILL.md"), []byte("theirs"), 0o644))
	_, err := Export(v1, foreign, m1, true)
	refused(t, err, "was not installed by maintainer")

	link := filepath.Join(t.TempDir(), "maintainer")
	require.NoError(t, os.Symlink(foreign, link))
	_, err = Export(v1, link, m1, true)
	refused(t, err, "symbolic link")

	file := filepath.Join(t.TempDir(), "maintainer")
	require.NoError(t, os.WriteFile(file, nil, 0o644))
	_, err = Export(v1, file, m1, true)
	refused(t, err, "not a directory")

	broken := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(broken, MarkerFile), []byte("{"), 0o644))
	_, err = Export(v1, broken, m1, true)
	refused(t, err, "unreadable marker")

	empty := t.TempDir()
	outcome, err := Export(v1, empty, m1, false)
	require.NoError(t, err)
	assert.Equal(t, Installed, outcome)
}
