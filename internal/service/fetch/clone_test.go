package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

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

func TestService_InterruptedCloneIsNeitherAdoptedNorFetched(t *testing.T) {
	marker := func(t *testing.T, target string) {
		t.Helper()
		require.NoError(t, os.MkdirAll(filepath.Join(target, ".git"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(target, ".git", gitsvc.IncompleteMarker), nil, 0o644))
	}
	for _, tc := range []struct {
		name    string
		prepare func(t *testing.T, target, origin string)
	}{
		{"stopped before Git initialised it", func(t *testing.T, target, _ string) { marker(t, target) }},
		{"stopped while Git wrote its configuration", func(t *testing.T, target, _ string) {
			marker(t, target)
			require.NoError(t, os.WriteFile(filepath.Join(target, ".git", "HEAD"), []byte("ref: refs/heads/master\n"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(target, ".git", "config"), []byte("[remote \"origin"), 0o644))
		}},
		{"stopped after the transfer", func(t *testing.T, target, origin string) {
			require.NoError(t, gitsvc.NewSync().Clone(context.Background(), gitsvc.CloneOptions{URL: origin, Path: target, Auth: gitsvc.Auth{Transport: "https"}}))
			marker(t, target)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			origin := originWithCommit(t)
			root := t.TempDir()
			target := filepath.Join(root, "public/acme/tool")
			tc.prepare(t, target, origin)
			cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "https", Concurrency: 1}}
			store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
			snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public, HTTPSCloneURL: origin, SourceProfile: "p"}
			var out, errw bytes.Buffer
			for _, apply := range []bool{false, true} {
				out.Reset()
				errw.Reset()
				svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
					Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Resolver: fakeGitHubNameResolver{},
					GitSync:  localOriginSync{Sync: gitsvc.NewSync(), localURL: origin, canonicalURL: "https://github.com/acme/tool.git"},
					Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
				})
				require.NoError(t, err)
				err = svc.Run(context.Background(), apply)
				if apply {
					var coded exitpkg.Coder
					require.True(t, errors.As(err, &coded), "%v", err)
					assert.Equal(t, exitpkg.Partial, coded.ExitCode())
				} else {
					require.NoError(t, err)
				}
				var plan planDocument
				require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
				require.NotEmpty(t, plan.Actions)
				for _, a := range plan.Actions {
					assert.Equal(t, "conflict", a.Kind, "no adopt, fetch or clone for an incomplete clone")
					assert.Equal(t, target, a.Path)
					assert.Contains(t, a.Reason, "incomplete clone")
				}
			}
			loaded, err := store.Load()
			require.NoError(t, err)
			assert.Empty(t, loaded.Repos, "an incomplete clone is not recorded")
			_, err = os.Lstat(filepath.Join(target, ".git", gitsvc.IncompleteMarker))
			assert.NoError(t, err, "nothing is removed")
		})
	}
}
