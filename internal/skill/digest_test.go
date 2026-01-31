package skill_test

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/skill"
)

func TestDigest(t *testing.T) {
	// The test vector shared with the octolab/skills catalog.
	digest, err := Digest(fstest.MapFS{
		"SKILL.md":        {Data: []byte("hello\n")},
		"references/a.md": {Data: []byte("a")},
	})
	require.NoError(t, err)
	assert.Equal(t, "sha256:64ce802d5cef2c789ce95d306c9fcce6779c6e8166339ae47634e7f8b265a34b", digest)

	// Sizes delimit the contents.
	a, err := Digest(fstest.MapFS{"a": {Data: []byte("bc")}, "b": {Data: []byte("")}})
	require.NoError(t, err)
	b, err := Digest(fstest.MapFS{"a": {Data: []byte("b")}, "b": {Data: []byte("c")}})
	require.NoError(t, err)
	assert.NotEqual(t, a, b)
}

func TestFiles(t *testing.T) {
	files, err := Files(fstest.MapFS{
		"SKILL.md":        {},
		"references/a.md": {},
		"a.md":            {},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"SKILL.md", "a.md", "references/a.md"}, files)

	_, err = Files(fstest.MapFS{".maintainer-skill.json": {}})
	assert.ErrorContains(t, err, "dotfiles")

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), nil, 0o644))
	require.NoError(t, os.Symlink("SKILL.md", filepath.Join(dir, "link.md")))
	_, err = Files(os.DirFS(dir))
	assert.ErrorContains(t, err, "symbolic links")
}

func TestCopyDigest(t *testing.T) {
	skill := fstest.MapFS{"SKILL.md": {Data: []byte("hello\n")}}
	want, err := Digest(skill)
	require.NoError(t, err)

	// The marker and Finder's litter are not part of the copy.
	digest, err := CopyDigest(fstest.MapFS{
		"SKILL.md":             {Data: []byte("hello\n")},
		MarkerFile:             {Data: []byte("{}")},
		".DS_Store":            {},
		"references/.DS_Store": {},
	})
	require.NoError(t, err)
	assert.Equal(t, want, digest)

	// Any other dotfile is, and so is a marker away from the top.
	for _, extra := range []string{".git/config", ".hidden", "references/" + MarkerFile} {
		digest, err := CopyDigest(fstest.MapFS{
			"SKILL.md": {Data: []byte("hello\n")},
			extra:      {},
		})
		require.NoError(t, err)
		assert.NotEqual(t, want, digest, extra)
	}
}
