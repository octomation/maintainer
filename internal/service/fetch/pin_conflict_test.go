package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	gh "github.com/google/go-github/v93/github"
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

// answering is a GitHub that answers every check of a pinned repository with
// one HTTP status.
type answering int

func (a answering) err() error {
	return &gh.ErrorResponse{Response: &http.Response{StatusCode: int(a)}, Message: http.StatusText(int(a))}
}

func (a answering) status() string {
	return fmt.Sprintf("%d %s", int(a), http.StatusText(int(a)))
}

func (a answering) ConfirmByID(context.Context, github.Profile, int64) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{}, a.err()
}

func (a answering) ResolveByName(context.Context, github.Profile, string, string) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{}, a.err()
}

// limited is a GitHub that answers every check with a rate limit error.
type limited struct{ err error }

func (l limited) ConfirmByID(context.Context, github.Profile, int64) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{}, l.err
}

func (l limited) ResolveByName(context.Context, github.Profile, string, string) (github.RepoSnapshot, error) {
	return github.RepoSnapshot{}, l.err
}

type pinChecker interface {
	github.Confirmer
	github.NameResolver
}

// A pin that GitHub does not confirm is a conflict of its repository: the
// rest of the plan is built and applied, and the run exits as with any other
// conflict. A failure that is not about the pinned repository still stops
// the run.
func TestService_UnconfirmedPinIsConflictOfItsRepository(t *testing.T) {
	origin := originWithCommit(t)
	setup := func(t *testing.T, answer pinChecker, format string) (*Service, *bytes.Buffer, string, string, string) {
		t.Helper()
		root, pins := t.TempDir(), t.TempDir()
		byName, byID := filepath.Join(pins, "gone"), filepath.Join(pins, "lost")
		cnf := &config.Fetch{
			Workspace: &config.Workspace{Root: root},
			Defaults:  config.Defaults{CloneURL: "https", Concurrency: 1},
			Repos: []config.Repo{
				{Match: config.RepoMatch{Owner: "acme", Name: "gone"}, Path: byName},
				{Match: config.RepoMatch{ID: 77}, Path: byID},
			},
		}
		snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public, HTTPSCloneURL: origin, SourceProfile: "p"}
		var out, errw bytes.Buffer
		svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
			Store:      state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil),
			Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}},
			Confirmer:  answer, Resolver: answer,
			GitSync:  gitsvc.NewSync(),
			Reporter: NewReporter(&out, &errw, format, 0, false),
		})
		require.NoError(t, err)
		return svc, &out, root, byName, byID
	}

	for _, answer := range []answering{http.StatusNotFound, http.StatusForbidden, http.StatusUnavailableForLegalReasons} {
		t.Run(http.StatusText(int(answer)), func(t *testing.T) {
			svc, out, _, byName, byID := setup(t, answer, FormatJSON)
			require.NoError(t, svc.Run(context.Background(), false), "a plan with conflicts exits 0")
			var plan planDocument
			require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
			kinds := map[string]string{}
			for _, a := range plan.Actions {
				kinds[a.Owner+"/"+a.Name] = a.Kind
				switch {
				case a.Owner == "acme" && a.Name == "gone":
					assert.Equal(t, byName, a.Path)
					assert.Equal(t, "GitHub does not confirm the pinned repository: "+answer.status(), a.Reason)
				case a.ID == 77:
					assert.Equal(t, byID, a.Path)
					assert.Equal(t, "GitHub does not confirm the pinned repository: "+answer.status(), a.Reason)
				}
			}
			assert.Equal(t, map[string]string{"acme/tool": "clone", "acme/gone": "conflict", "/": "conflict"}, kinds)

			svc, out, root, _, _ := setup(t, answer, FormatHuman)
			require.NoError(t, svc.Run(context.Background(), false))
			assert.Contains(t, out.String(), "  ✗ conflict   acme/gone                    → GitHub does not confirm the pinned repository: "+answer.status()+"\n")
			assert.Contains(t, out.String(), "  ✗ conflict   id=77                        → GitHub does not confirm the pinned repository: "+answer.status()+"\n")
			assert.Contains(t, out.String(), "  + clone      acme/tool")

			err := svc.Run(context.Background(), true)
			var coded exitpkg.Coder
			require.True(t, errors.As(err, &coded), "%v", err)
			assert.Equal(t, exitpkg.Partial, coded.ExitCode())
			_, statErr := os.Stat(filepath.Join(root, "public/acme/tool/.git"))
			assert.NoError(t, statErr, "the other repository is cloned")
		})
	}

	forbidden := func(header http.Header, message string) error {
		return &gh.ErrorResponse{Response: &http.Response{StatusCode: http.StatusForbidden, Header: header}, Message: message}
	}
	failures := map[string]pinChecker{
		"Unauthorized":         answering(http.StatusUnauthorized),
		"Bad Gateway":          answering(http.StatusBadGateway),
		"primary rate limit":   limited{forbidden(http.Header{"X-Ratelimit-Remaining": {"0"}}, "API rate limit exceeded")},
		"Retry-After":          limited{forbidden(http.Header{"Retry-After": {"60"}}, "Forbidden")},
		"secondary rate limit": limited{forbidden(http.Header{}, "You have exceeded a secondary rate limit.")},
		"rate limit error":     limited{&gh.RateLimitError{Response: &http.Response{StatusCode: http.StatusForbidden}, Message: "API rate limit exceeded"}},
	}
	for name, answer := range failures {
		t.Run(name, func(t *testing.T) {
			svc, out, _, _, _ := setup(t, answer, FormatJSON)
			err := svc.Run(context.Background(), false)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot verify explicit checkout")
			assert.Empty(t, out.String())
		})
	}
}
