package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	gh "github.com/google/go-github/v93/github"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// restTransport answers GitHub REST paths with canned responses and records
// every request it receives.
type restTransport struct {
	mu     sync.Mutex
	routes map[string]restResponse
	seen   []string
}

type restResponse struct {
	status int
	body   string
}

func (f *restTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen = append(f.seen, req.Method+" "+req.URL.Path)
	r, ok := f.routes[req.URL.Path]
	if !ok {
		r = restResponse{status: http.StatusNotFound, body: `{"message":"Not Found"}`}
	}
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("X-RateLimit-Remaining", "4999")
	return &http.Response{StatusCode: r.status, Header: h, Body: io.NopCloser(strings.NewReader(r.body)), Request: req}, nil
}

func (f *restTransport) requests() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.seen...)
}

// restAPI is the real REST discoverer behind the real RateGuard, over fake.
func restAPI(fake http.RoundTripper) *github.RESTDiscoverer {
	guard := github.NewRateGuard(fake, github.WithSleeper(func(context.Context, time.Duration) error { return nil }))
	return github.NewRESTDiscoverer(func(context.Context, string) *gh.Client {
		c, _ := gh.NewClient(gh.WithHTTPClient(&http.Client{Transport: guard}))
		return c
	})
}

type planDocument struct {
	Actions []struct {
		Kind   string `json:"kind"`
		ID     int64  `json:"id"`
		Owner  string `json:"owner"`
		Name   string `json:"name"`
		Path   string `json:"path"`
		Reason string `json:"reason"`
	} `json:"actions"`
}

func TestService_RefusedCheckoutIsOneConflict(t *testing.T) {
	const sso = `{"message":"Resource protected by organization SAML enforcement. You must grant your Personal Access token access to this organization."}`
	for _, ignored := range []bool{false, true} {
		t.Run(map[bool]string{false: "refused", true: "ignored"}[ignored], func(t *testing.T) {
			root := t.TempDir()
			refused := filepath.Join(root, "public/sso-org/tool")
			makeClone(t, refused, "git@github.com:sso-org/tool.git")
			cnf := &config.Fetch{Workspace: &config.Workspace{Root: root}, Defaults: config.Defaults{CloneURL: "ssh", Concurrency: 1}}
			if ignored {
				cnf.Repos = []config.Repo{{Match: config.RepoMatch{Owner: "sso-org", Name: "tool"}, Ignore: true}}
			}
			fake := &restTransport{routes: map[string]restResponse{
				"/repos/sso-org/tool": {status: http.StatusForbidden, body: sso},
			}}
			api := restAPI(fake)
			snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public, SourceProfile: "p"}

			var out, errw bytes.Buffer
			svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
				Store:      state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil),
				Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}},
				Confirmer:  api, Resolver: api, GitSync: gitsvc.NewSync(),
				Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
			})
			require.NoError(t, err)
			require.NoError(t, svc.Run(context.Background(), false), errw.String())

			var plan planDocument
			require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String())
			kinds := map[string]string{}
			for _, a := range plan.Actions {
				kinds[a.Owner+"/"+a.Name] = a.Kind
				if a.Kind == "conflict" {
					assert.Equal(t, refused, a.Path)
					assert.Contains(t, a.Reason, "403")
				}
			}
			assert.Equal(t, "clone", kinds["acme/svc"], "the rest of the run is planned")
			if ignored {
				assert.NotContains(t, kinds, "sso-org/tool")
				assert.Empty(t, fake.requests(), "an ignored checkout is never resolved")
			} else {
				assert.Equal(t, "conflict", kinds["sso-org/tool"])
				assert.Equal(t, []string{"GET /repos/sso-org/tool"}, fake.requests(), "an ordinary 403 is not retried")
			}
		})
	}
}
