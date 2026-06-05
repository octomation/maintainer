package github_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	gh "github.com/google/go-github/v88/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// fakeTransport replays canned responses keyed by "METHOD path?page=N".
type fakeTransport struct {
	routes map[string]fakeResponse
	seen   map[string]int
}

type fakeResponse struct {
	status int
	body   string
	next   int // emit a Link rel="next" to this page when > 0
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	page := req.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	key := fmt.Sprintf("%s %s?page=%s", req.Method, req.URL.Path, page)
	f.seen[key]++
	r, ok := f.routes[key]
	if !ok {
		return jsonResponse(http.StatusNotFound, `{"message":"Not Found"}`, 0, req), nil
	}
	return jsonResponse(r.status, r.body, r.next, req), nil
}

func jsonResponse(status int, body string, next int, req *http.Request) *http.Response {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("X-RateLimit-Remaining", "4999")
	if next > 0 {
		u := *req.URL
		q := u.Query()
		q.Set("page", fmt.Sprint(next))
		u.RawQuery = q.Encode()
		h.Set("Link", fmt.Sprintf(`<%s>; rel="next"`, u.String()))
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     h,
		Request:    req,
	}
}

func repoJSON(id int64, owner, name, visibility string, private bool) string {
	return fmt.Sprintf(
		`{"id":%d,"node_id":"N%d","name":%q,"owner":{"login":%q},"visibility":%q,"private":%t,"default_branch":"main","clone_url":"https://github.com/%s/%s.git","ssh_url":"git@github.com:%s/%s.git"}`,
		id, id, name, owner, visibility, private, owner, name, owner, name,
	)
}

func TestRESTDiscoverer_List(t *testing.T) {
	fake := &fakeTransport{seen: map[string]int{}, routes: map[string]fakeResponse{
		"GET /user?page=1":      {status: 200, body: `{"login":"me"}`},
		"GET /user/orgs?page=1": {status: 200, body: `[{"login":"acme"}]`},
		"GET /user/repos?page=1": {status: 200, next: 2, body: "[" +
			repoJSON(1, "me", "a", "public", false) + "," +
			repoJSON(10, "acme", "svc", "public", false) + "," +
			repoJSON(99, "stranger", "x", "public", false) + "]"},
		"GET /user/repos?page=2": {status: 200, body: "[" +
			repoJSON(2, "me", "b", "private", true) + "]"},
		"GET /orgs/acme/repos?page=1": {status: 200, body: "[" +
			repoJSON(10, "acme", "svc", "private", true) + "]"},
	}}

	factory := func(ctx context.Context, token string) *gh.Client {
		c, _ := gh.NewClient(gh.WithHTTPClient(&http.Client{Transport: fake}))
		return c
	}
	d := github.NewRESTDiscoverer(factory)

	got, err := d.List(context.Background(), github.Profile{Name: "primary", Token: "t", Owners: []string{"me", "acme"}})
	require.NoError(t, err)

	byID := map[int64]github.RepoSnapshot{}
	for _, s := range got.Snapshots {
		byID[s.ID] = s
	}
	// stranger is dropped by the include_owners allowlist.
	require.Len(t, got.Snapshots, 3)
	assert.Equal(t, github.Public, byID[1].Visibility)
	assert.Equal(t, github.Private, byID[2].Visibility)
	// id=10 seen as public via /user/repos and private via the org endpoint:
	// the broader (private) snapshot wins (§5.1).
	assert.Equal(t, github.Private, byID[10].Visibility)
	assert.Equal(t, "primary", byID[10].SourceProfile)
	// pagination actually walked two pages.
	assert.Equal(t, 1, fake.seen["GET /user/repos?page=2"])
}

func TestRESTDiscoverer_WildcardExpandsToMemberOrgs(t *testing.T) {
	fake := &fakeTransport{seen: map[string]int{}, routes: map[string]fakeResponse{
		"GET /user?page=1":      {status: 200, body: `{"login":"me"}`},
		"GET /user/orgs?page=1": {status: 200, body: `[{"login":"acme"},{"login":"globex"}]`},
		"GET /user/repos?page=1": {status: 200, body: "[" +
			repoJSON(1, "me", "dotfiles", "private", true) + "]"},
		"GET /orgs/acme/repos?page=1": {status: 200, body: "[" +
			repoJSON(2, "acme", "svc", "private", true) + "]"},
		"GET /orgs/globex/repos?page=1": {status: 200, body: "[" +
			repoJSON(3, "globex", "tool", "public", false) + "]"},
	}}
	factory := func(_ context.Context, _ string) *gh.Client {
		c, _ := gh.NewClient(gh.WithHTTPClient(&http.Client{Transport: fake}))
		return c
	}
	d := github.NewRESTDiscoverer(factory)

	// "*" must pull the user plus every membership without listing them.
	got, err := d.List(context.Background(), github.Profile{Name: "p", Token: "t", Owners: []string{"*"}})
	require.NoError(t, err)
	require.Len(t, got.Snapshots, 3)
	assert.Equal(t, 1, fake.seen["GET /orgs/acme/repos?page=1"])
	assert.Equal(t, 1, fake.seen["GET /orgs/globex/repos?page=1"])

	// An empty owner list behaves identically.
	got, err = d.List(context.Background(), github.Profile{Name: "p", Token: "t"})
	require.NoError(t, err)
	assert.Len(t, got.Snapshots, 3)
}

func TestRESTDiscoverer_UserFallbackOn404(t *testing.T) {
	fake := &fakeTransport{seen: map[string]int{}, routes: map[string]fakeResponse{
		"GET /user?page=1":             {status: 200, body: `{"login":"me"}`},
		"GET /user/orgs?page=1":        {status: 200, body: `[]`},
		"GET /orgs/ghost/repos?page=1": {status: 404, body: `{"message":"Not Found"}`},
		"GET /users/ghost/repos?page=1": {status: 200, body: "[" +
			repoJSON(5, "ghost", "pub", "public", false) + "]"},
	}}
	factory := func(ctx context.Context, token string) *gh.Client {
		c, _ := gh.NewClient(gh.WithHTTPClient(&http.Client{Transport: fake}))
		return c
	}
	d := github.NewRESTDiscoverer(factory)

	got, err := d.List(context.Background(), github.Profile{Name: "p", Token: "t", Owners: []string{"ghost"}})
	require.NoError(t, err)
	require.Len(t, got.Snapshots, 1)
	assert.Equal(t, "ghost", got.Snapshots[0].Owner)
	assert.Equal(t, 1, fake.seen["GET /users/ghost/repos?page=1"]) // fell back to the user endpoint
}
