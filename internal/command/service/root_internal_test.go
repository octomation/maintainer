package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/status"
)

func TestOriginName(t *testing.T) {
	config := "[core]\n\tbare = false\n[remote \"origin\"]\n\turl = git@github.com:octomation/maintainer.git\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n"

	t.Run("a checkout", func(t *testing.T) {
		repo := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repo, ".git", "config"), []byte(config), 0o644))
		assert.Equal(t, "octomation/maintainer", originName(repo))
	})

	t.Run("a linked worktree", func(t *testing.T) {
		main := t.TempDir()
		gitDir := filepath.Join(main, ".git")
		wtDir := filepath.Join(gitDir, "worktrees", "feature")
		require.NoError(t, os.MkdirAll(wtDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(gitDir, "config"), []byte(config), 0o644))
		// The worktree's own directory has no config: commondir points up.
		require.NoError(t, os.WriteFile(filepath.Join(wtDir, "commondir"), []byte("../..\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(wtDir, "HEAD"), []byte("ref: refs/heads/feature\n"), 0o644))
		worktree := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(worktree, ".git"), []byte("gitdir: "+wtDir+"\n"), 0o644))
		assert.Equal(t, "octomation/maintainer", originName(worktree))
	})

	t.Run("no origin", func(t *testing.T) {
		repo := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".git"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(repo, ".git", "config"), []byte("[core]\n\tbare = false\n"), 0o644))
		assert.Equal(t, "", originName(repo))
		assert.Equal(t, "", originName(t.TempDir()))
	})
}

func TestMirrorRootsCarryVisibility(t *testing.T) {
	mirror := func() string {
		path := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(path, ".github", "issues"), 0o755))
		return path
	}
	private, unknown := mirror(), mirror()
	roots := mirrorRoots([]status.Row{
		{ID: 1, Repository: "acme/secret", Path: private, Visibility: "private"},
		{ID: 2, Repository: "acme/open", Path: unknown},
	})
	require.Len(t, roots, 2)
	assert.Equal(t, "private", roots[0].Visibility)
	assert.Equal(t, "", roots[1].Visibility)
}

func TestOpenURLThroughARedirectPage(t *testing.T) {
	state := filepath.Join(t.TempDir(), "maintainer", "service")
	const url = "http://0123456789abcdef0123456789abcdef.maintainer.localhost:8080/"
	var args []string
	var page string
	launch = func(ctx context.Context, path string) (*exec.Cmd, error) {
		cmd := exec.CommandContext(ctx, "true")
		args = append(args, cmd.Args...)
		args = append(args, path)
		page = path
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
		info, err = os.Stat(filepath.Dir(path))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
		raw, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Contains(t, string(raw), `content="0;url=`+url+`"`)
		return cmd, cmd.Start()
	}
	redirectWait = 10 * time.Millisecond
	t.Cleanup(func() { launch, redirectWait = nil, 5*time.Second })

	require.NoError(t, openURL(context.Background(), state, url, 8080))
	require.NotEmpty(t, page)
	assert.Equal(t, state, filepath.Dir(filepath.Dir(page)), "the page is under the state folder")
	assert.NotContains(t, strings.Join(args, " "), "0123456789abcdef", "the label is no process argument")
	_, err := os.Stat(filepath.Dir(page))
	assert.True(t, os.IsNotExist(err), "the folder is removed after the wait")
}

func TestRedirectPageVisitsThePlainAddressWithoutTheLabel(t *testing.T) {
	const label = "0123456789abcdef0123456789abcdef"
	target := "http://" + label + ".maintainer.localhost:8080/"
	plain := "http://maintainer.localhost:8080/status?back=1"
	page := redirectPage(target, plain)

	// The plain page is requested as a top-level visit, without the label.
	assert.Contains(t, page, `location.assign(plain)`)
	assert.Contains(t, page, `const plain = "`+plain+`";`)
	assert.NotContains(t, plain, label)
	// The label is only in the final navigation (and its no-script form).
	assert.Equal(t, 2, strings.Count(page, label))
	assert.Contains(t, page, `location.replace("`+target+`")`)
	assert.Contains(t, page, `content="0;url=`+target+`"`)
	// No URL holds both, and nothing hands the label to the plain page.
	for _, line := range strings.Split(page, "\n") {
		assert.False(t, strings.Contains(line, label) && strings.Contains(line, "maintainer.localhost:8080/status"), line)
	}
	assert.Contains(t, page, `<meta name="referrer" content="no-referrer">`)
	for _, leak := range []string{"window.name", "postMessage", "document.cookie", "localStorage", "sessionStorage", "?" + label, "#" + label} {
		assert.NotContains(t, page, leak)
	}
}
