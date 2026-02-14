package web_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/web"
)

func TestLabel(t *testing.T) {
	a, err := web.NewLabel()
	require.NoError(t, err)
	b, err := web.NewLabel()
	require.NoError(t, err)
	assert.True(t, web.ValidLabel(a))
	assert.NotEqual(t, a, b)
	for _, bad := range []string{"", a[:31], a + "0", "0123456789ABCDEF0123456789ABCDEF", "0123456789abcdef0123456789abcdeg", "../../../../etc/passwd0000000000"} {
		assert.False(t, web.ValidLabel(bad), bad)
	}
	assert.Equal(t, "http://"+a+".maintainer.localhost:8080/", web.LabelURL(a, 8080))
	assert.Equal(t, "/s/maintainer/service", web.StateDir(func(string) string { return "/s" }, "/h"))
	assert.Equal(t, "/h/.local/state/maintainer/service", web.StateDir(func(string) string { return "" }, "/h"))
	assert.Len(t, web.Proof(a, []byte("nonce")), 64)
	assert.NotEqual(t, web.Proof(a, []byte("nonce")), web.Proof(b, []byte("nonce")))
}

func TestLabelFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state", "maintainer", "service")
	const first, second = "0123456789abcdef0123456789abcdef", "fedcba9876543210fedcba9876543210"

	_, err := web.ReadLabel(dir, 8080)
	require.ErrorIs(t, err, web.ErrNoLabel)

	require.NoError(t, web.WriteLabel(dir, 8080, first))
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	info, err = os.Stat(web.LabelPath(dir, 8080))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	got, err := web.ReadLabel(dir, 8080)
	require.NoError(t, err)
	assert.Equal(t, first, got)

	// A stale file is replaced; another port keeps its own.
	require.NoError(t, web.WriteLabel(dir, 8080, second))
	require.NoError(t, web.WriteLabel(dir, 8081, first))
	got, _ = web.ReadLabel(dir, 8080)
	assert.Equal(t, second, got)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 2, "no temporary file stays")

	// Removal only when the file still holds the label of the run.
	require.NoError(t, web.RemoveLabel(dir, 8080, first))
	got, _ = web.ReadLabel(dir, 8080)
	assert.Equal(t, second, got)
	require.NoError(t, web.RemoveLabel(dir, 8080, second))
	_, err = os.Stat(web.LabelPath(dir, 8080))
	assert.True(t, os.IsNotExist(err))
	require.NoError(t, web.RemoveLabel(dir, 8080, second), "a missing file is fine")

	// A file without a label is no label.
	require.NoError(t, os.WriteFile(web.LabelPath(dir, 8082), []byte("not a label\n"), 0o600))
	_, err = web.ReadLabel(dir, 8082)
	require.ErrorIs(t, err, web.ErrNoLabel)
}

func TestPrepareStateDir(t *testing.T) {
	base := t.TempDir()
	open := filepath.Join(base, "open")
	require.NoError(t, os.Mkdir(open, 0o755))
	require.NoError(t, os.Chmod(open, 0o755))
	require.NoError(t, web.PrepareStateDir(open))
	info, err := os.Stat(open)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm(), "the mode is closed")

	link := filepath.Join(base, "link")
	require.NoError(t, os.Symlink(open, link))
	assert.Error(t, web.PrepareStateDir(link), "a link is not the folder")

	file := filepath.Join(base, "file")
	require.NoError(t, os.WriteFile(file, nil, 0o600))
	assert.Error(t, web.PrepareStateDir(file))
	assert.Error(t, web.PrepareStateDir(""))
}
