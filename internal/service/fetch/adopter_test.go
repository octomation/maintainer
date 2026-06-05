package fetch_test

import (
	"context"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// makeClone initialises a repo at dir with a single origin remote URL.
func makeClone(t *testing.T, dir, originURL string) {
	t.Helper()
	repo, err := gogit.PlainInit(dir, false)
	require.NoError(t, err)
	_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{originURL}})
	require.NoError(t, err)
}

type fakeResolver struct{ ids map[string]int64 }

func (r fakeResolver) ResolveByName(_ context.Context, owner, name string) (int64, error) {
	return r.ids[owner+"/"+name], nil
}

func TestAdopter_Scan(t *testing.T) {
	root := t.TempDir()
	makeClone(t, filepath.Join(root, "public/acme/svc"), "git@github.com:acme/svc.git")
	makeClone(t, filepath.Join(root, "private/acme/secret"), "https://github.com/acme/secret.git")
	makeClone(t, filepath.Join(root, "public/acme/renamed"), "git@github.com:acme/renamed.git")
	makeClone(t, filepath.Join(root, "elsewhere/gitlab"), "git@gitlab.com:acme/x.git") // non-github → skipped

	snapshots := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "svc"},
		{ID: 2, Owner: "acme", Name: "secret"},
	}
	resolver := fakeResolver{ids: map[string]int64{"acme/renamed": 5}}

	a := NewAdopter(gitsvc.NewSync(), resolver)
	clones, err := a.Scan(context.Background(), root, snapshots, nil)
	require.NoError(t, err)

	byPath := map[string]DiskClone{}
	for _, c := range clones {
		byPath[c.Path] = c
	}
	require.Len(t, clones, 3) // gitlab clone excluded

	svc := byPath[filepath.Join(root, "public/acme/svc")]
	assert.Equal(t, int64(1), svc.ID)
	assert.Equal(t, "ssh", svc.Transport)
	assert.Equal(t, "git@github.com:acme/svc.git", svc.RemoteURL)

	secret := byPath[filepath.Join(root, "private/acme/secret")]
	assert.Equal(t, int64(2), secret.ID)
	assert.Equal(t, "https", secret.Transport)

	// resolved via the rename-redirect resolver.
	renamed := byPath[filepath.Join(root, "public/acme/renamed")]
	assert.Equal(t, int64(5), renamed.ID)
}
