package fetch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

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

// hubRepo is a repository of the fake GitHub; private ones are visible only to
// the listed tokens.
type hubRepo struct {
	id          int64
	owner, name string
	private     bool
	fork        bool
	unlisted    bool     // reachable by id or name, absent from listings
	former      []string // owner/name spellings that redirect to it
	tokens      []string
}

func (r *hubRepo) visible(token string) bool {
	if !r.private {
		return true
	}
	for _, t := range r.tokens {
		if t == token {
			return true
		}
	}
	return false
}

func (r *hubRepo) json() string {
	visibility := "public"
	if r.private {
		visibility = "private"
	}
	return fmt.Sprintf(`{"id":%d,"node_id":"N%d","name":%q,"owner":{"login":%q},"visibility":%q,"private":%t,"fork":%t,"default_branch":"main","clone_url":"https://github.com/%s/%s.git","ssh_url":"git@github.com:%s/%s.git"}`,
		r.id, r.id, r.name, r.owner, visibility, r.private, r.fork, r.owner, r.name, r.owner, r.name)
}

// fakeHub answers the REST calls of discovery, confirmation and name
// resolution case-insensitively, like GitHub, and logs every request.
type fakeHub struct {
	mu       sync.Mutex
	logins   map[string]string   // token → login
	orgs     map[string][]string // token → member organizations
	repos    []*hubRepo
	requests []string // "token GET /path?query"
}

func (h *fakeHub) factory(_ context.Context, token string) *gh.Client {
	client, _ := gh.NewClient(gh.WithHTTPClient(&http.Client{Transport: hubTransport{hub: h, token: token}}))
	return client
}

type hubTransport struct {
	hub   *fakeHub
	token string
}

func (t hubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	h := t.hub
	h.mu.Lock()
	defer h.mu.Unlock()
	h.requests = append(h.requests, t.token+" "+req.Method+" "+req.URL.Path+"?"+req.URL.RawQuery)
	status, body := h.route(t.token, req)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": {"application/json"}, "X-Ratelimit-Remaining": {"4999"}},
		Request:    req,
	}, nil
}

func (h *fakeHub) route(token string, req *http.Request) (int, string) {
	const notFound = `{"message":"Not Found"}`
	login := h.logins[token]
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	list := func(keep func(*hubRepo) bool) (int, string) {
		var items []string
		for _, r := range h.repos {
			if r.visible(token) && !r.unlisted && keep(r) {
				items = append(items, r.json())
			}
		}
		return http.StatusOK, "[" + strings.Join(items, ",") + "]"
	}
	isOrg := func(name string) bool {
		for _, orgs := range h.orgs {
			for _, org := range orgs {
				if strings.EqualFold(org, name) {
					return true
				}
			}
		}
		return false
	}
	switch {
	case len(parts) == 1 && parts[0] == "user":
		return http.StatusOK, fmt.Sprintf(`{"login":%q}`, login)
	case len(parts) == 2 && parts[0] == "user" && parts[1] == "orgs":
		var items []string
		for _, org := range h.orgs[token] {
			items = append(items, fmt.Sprintf(`{"login":%q}`, org))
		}
		return http.StatusOK, "[" + strings.Join(items, ",") + "]"
	case len(parts) == 2 && parts[0] == "user" && parts[1] == "repos":
		return list(func(r *hubRepo) bool { return strings.EqualFold(r.owner, login) })
	case len(parts) == 3 && parts[0] == "orgs" && parts[2] == "repos":
		if !isOrg(parts[1]) {
			return http.StatusNotFound, notFound
		}
		public := req.URL.Query().Get("type") == "public"
		return list(func(r *hubRepo) bool { return strings.EqualFold(r.owner, parts[1]) && (!public || !r.private) })
	case len(parts) == 3 && parts[0] == "users" && parts[2] == "repos":
		return list(func(r *hubRepo) bool { return strings.EqualFold(r.owner, parts[1]) && !r.private })
	case len(parts) == 2 && parts[0] == "repositories":
		id, _ := strconv.ParseInt(parts[1], 10, 64)
		for _, r := range h.repos {
			if r.id == id && r.visible(token) {
				return http.StatusOK, r.json()
			}
		}
	case len(parts) == 3 && parts[0] == "repos":
		name := parts[1] + "/" + parts[2]
		for _, r := range h.repos {
			matches := strings.EqualFold(r.owner+"/"+r.name, name)
			for _, former := range r.former {
				matches = matches || strings.EqualFold(former, name)
			}
			if matches && r.visible(token) {
				return http.StatusOK, r.json()
			}
		}
	}
	return http.StatusNotFound, notFound
}

// about lists the logged requests that mention any of the fragments.
func (h *fakeHub) about(fragments ...string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	var out []string
	for _, r := range h.requests {
		for _, f := range fragments {
			if strings.Contains(r, f) {
				out = append(out, r)
				break
			}
		}
	}
	return out
}

// recordingGit inspects real checkouts and records every operation instead of
// performing it.
type recordingGit struct {
	*gitsvc.Sync
	mu      sync.Mutex
	fetched map[string]string // path → token
	cloned  []string
	moved   [][2]string
	updated [][2]string
	failure error
}

func (g *recordingGit) Fetch(_ context.Context, path string, auth gitsvc.Auth) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.failure != nil {
		return g.failure
	}
	g.fetched[path] = auth.Token
	return nil
}

func (g *recordingGit) Clone(_ context.Context, opt gitsvc.CloneOptions) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cloned = append(g.cloned, opt.Path)
	return nil
}

func (g *recordingGit) Move(from, to string) error {
	g.moved = append(g.moved, [2]string{from, to})
	return nil
}

func (g *recordingGit) UpdateRemote(path, url string) error {
	g.updated = append(g.updated, [2]string{path, url})
	return nil
}

// selectionFixture is a workspace with tracked checkouts of acme and other,
// recorded by the profiles primary and beta, and the fake GitHub behind them.
type selectionFixture struct {
	root  string
	hub   *fakeHub
	git   *recordingGit
	store *state.Store
	cnf   *config.Fetch
}

func newSelectionFixture(t *testing.T, records ...state.Record) *selectionFixture {
	t.Helper()
	f := &selectionFixture{
		root: t.TempDir(),
		hub: &fakeHub{
			logins: map[string]string{"primary-token": "me", "beta-token": "beta-bot"},
			orgs:   map[string][]string{"primary-token": {"acme", "other"}, "beta-token": {"acme", "other"}},
		},
		git: &recordingGit{Sync: gitsvc.NewSync(), fetched: map[string]string{}},
		cnf: &config.Fetch{Workspace: &config.Workspace{Root: ""}, Defaults: config.Defaults{CloneURL: "ssh", Concurrency: 1}},
	}
	f.cnf.Workspace.Root = f.root
	f.store = state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	st := state.New()
	for _, rec := range records {
		if rec.Path == "" {
			rec.Path = f.path(rec.Visibility, rec.OwnerLogin, rec.Name)
		}
		if rec.RemoteURL == "" {
			rec.RemoteURL = "git@github.com:" + rec.OwnerLogin + "/" + rec.Name + ".git"
		}
		rec.CloneURL = "ssh"
		makeClone(t, rec.Path, rec.RemoteURL)
		st.Upsert(rec)
	}
	require.NoError(t, f.store.Save(st))
	return f
}

func (f *selectionFixture) path(visibility, owner, name string) string {
	if visibility == "" {
		visibility = "public"
	}
	return filepath.Join(f.root, visibility, owner, name)
}

// run plans or applies with the profiles as `fetch --owner/--profile` resolves
// them and returns the JSON plan and the human summary of the same run.
func (f *selectionFixture) run(t *testing.T, apply bool, sel Selection, profiles ...ResolvedProfile) (jsonPlan, error) {
	t.Helper()
	require.NoError(t, f.cnf.Validate())
	discoverer := github.NewRESTDiscoverer(f.hub.factory)
	var out, errw bytes.Buffer
	svc, err := NewService(f.cnf, profiles, f.root, f.root, 1, Deps{
		Store: f.store, Discoverer: discoverer, Confirmer: discoverer, Resolver: discoverer,
		GitSync: f.git, Clock: clock, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
		Selection: sel,
	})
	require.NoError(t, err)
	runErr := svc.Run(context.Background(), apply)
	var plan jsonPlan
	require.NoError(t, json.Unmarshal(out.Bytes(), &plan), out.String()+errw.String())
	return plan, runErr
}

func (f *selectionFixture) records(t *testing.T) map[int64]string {
	t.Helper()
	st, err := f.store.Load()
	require.NoError(t, err)
	out := make(map[int64]string, len(st.Repos))
	for _, rec := range st.Repos {
		raw, err := json.Marshal(rec)
		require.NoError(t, err)
		out[rec.ID] = string(raw)
	}
	return out
}

func actionOf(t *testing.T, plan jsonPlan, id int64) []string {
	t.Helper()
	var out []string
	for _, act := range plan.Actions {
		if act.ID == id {
			out = append(out, act.Kind+": "+act.Reason)
		}
	}
	return out
}

var (
	primary = ResolvedProfile{Name: "primary", Token: "primary-token"}
	beta    = ResolvedProfile{Name: "beta", Token: "beta-token"}
)

func withOwners(p ResolvedProfile, owners ...string) ResolvedProfile {
	p.Owners = owners
	return p
}

// --owner acme: repositories of other get no request, no fetch and no state
// change; they are listed as noop with the reason (rows 5 and 1).
func TestSelectionOwnerNarrowsTheRun(t *testing.T) {
	f := newSelectionFixture(t,
		state.Record{ID: 1, OwnerLogin: "acme", Name: "svc", SourceProfile: "primary"},
		state.Record{ID: 2, OwnerLogin: "other", Name: "tool", SourceProfile: "primary"},
		state.Record{ID: 3, OwnerLogin: "other", Name: "old", SourceProfile: "primary"},
		state.Record{ID: 4, OwnerLogin: "acme", Name: "moving", SourceProfile: "primary"},
	)
	f.hub.repos = []*hubRepo{
		{id: 1, owner: "acme", name: "svc"},
		{id: 2, owner: "other", name: "tool"},
		{id: 3, owner: "other", name: "new", former: []string{"other/old"}},
		{id: 4, owner: "other", name: "moving", former: []string{"acme/moving"}}, // transferred out
	}
	before := f.records(t)

	plan, err := f.run(t, true, Selection{Owners: []string{"acme"}}, withOwners(primary, "acme"))
	require.NoError(t, err)

	assert.Equal(t, []string{"fetch: "}, actionOf(t, plan, 1))
	assert.Equal(t, []string{"noop: outside the selection"}, actionOf(t, plan, 2))
	assert.Equal(t, []string{"noop: outside the selection"}, actionOf(t, plan, 3))
	assert.Equal(t, []string{"noop: transferred outside the selection to other/moving"}, actionOf(t, plan, 4))
	assert.Empty(t, f.hub.about("/repositories/2", "/repositories/3", "/repos/other/"))
	assert.Len(t, f.hub.about("/repositories/4"), 1, "the selected record is confirmed once")
	assert.Equal(t, map[string]string{f.path("", "acme", "svc"): "primary-token"}, f.git.fetched)
	assert.Empty(t, f.git.moved)
	assert.Empty(t, f.git.updated)
	after := f.records(t)
	for _, id := range []int64{2, 3, 4} {
		assert.Equal(t, before[id], after[id], "record %d", id)
	}
	assert.NotEqual(t, before[1], after[1], "the selected repository is fetched and recorded")
}

// The human output collapses the rows left alone and counts the selection.
func TestSelectionHumanSummary(t *testing.T) {
	f := newSelectionFixture(t,
		state.Record{ID: 1, OwnerLogin: "acme", Name: "svc", SourceProfile: "primary"},
		state.Record{ID: 2, OwnerLogin: "other", Name: "tool", SourceProfile: "primary"},
	)
	f.hub.repos = []*hubRepo{{id: 1, owner: "acme", name: "svc"}, {id: 2, owner: "other", name: "tool"}}
	discoverer := github.NewRESTDiscoverer(f.hub.factory)
	var out bytes.Buffer
	svc, err := NewService(f.cnf, []ResolvedProfile{withOwners(primary, "acme")}, f.root, f.root, 1, Deps{
		Store: f.store, Discoverer: discoverer, Confirmer: discoverer, Resolver: discoverer,
		GitSync: f.git, Reporter: NewReporter(&out, &out, FormatHuman, 0, false),
		Selection: Selection{Owners: []string{"ACME"}},
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	assert.Contains(t, out.String(), "plan: 2 repos total")
	assert.Contains(t, out.String(), "noop=1 conflict=0 errors=0 selected=1/2")
	assert.NotContains(t, out.String(), "other/tool")
}

// A tracked record of an excluded owner that a selected owner now lists was
// transferred in: it is moved and fetched without a confirmation (row 2).
func TestSelectionTransferredIn(t *testing.T) {
	f := newSelectionFixture(t, state.Record{ID: 11, OwnerLogin: "other", Name: "incoming", SourceProfile: "primary"})
	f.hub.repos = []*hubRepo{{id: 11, owner: "acme", name: "incoming", former: []string{"other/incoming"}}}

	plan, err := f.run(t, true, Selection{Owners: []string{"acme"}}, withOwners(primary, "acme"))
	require.NoError(t, err)
	assert.Equal(t, []string{"move: ", "fetch: "}, actionOf(t, plan, 11))
	assert.Empty(t, f.hub.about("/repositories/"))
	assert.Equal(t, [][2]string{{f.path("", "other", "incoming"), f.path("", "acme", "incoming")}}, f.git.moved)
	assert.Equal(t, [][2]string{{f.path("", "acme", "incoming"), "git@github.com:acme/incoming.git"}}, f.git.updated)
	st, err := f.store.Load()
	require.NoError(t, err)
	assert.Equal(t, "acme", st.Repos[0].OwnerLogin)
	assert.Equal(t, f.path("", "acme", "incoming"), st.Repos[0].Path)
}

// An id-only pin is explicit intent: it is confirmed once, and managed only
// when its owner is selected (row 3).
func TestSelectionIDOnlyPin(t *testing.T) {
	for _, tc := range []struct {
		owner string
		want  []string
	}{
		{"acme", []string{"adopt: ", "fetch: "}},
		{"other", []string{"noop: outside the selection"}},
	} {
		t.Run(tc.owner, func(t *testing.T) {
			f := newSelectionFixture(t)
			pin := filepath.Join(t.TempDir(), "special")
			makeClone(t, pin, "git@github.com:"+tc.owner+"/special.git")
			f.cnf.Repos = []config.Repo{{Match: config.RepoMatch{ID: 21}, Path: pin}}
			f.hub.repos = []*hubRepo{{id: 21, owner: tc.owner, name: "special", unlisted: true}}

			plan, err := f.run(t, true, Selection{Owners: []string{"acme"}}, withOwners(primary, "acme"))
			require.NoError(t, err)
			assert.Equal(t, tc.want, actionOf(t, plan, 21))
			assert.Len(t, f.hub.about("/repositories/21"), 1)
			st, err := f.store.Load()
			require.NoError(t, err)
			if tc.owner == "acme" {
				require.Len(t, st.Repos, 1)
				assert.Equal(t, pin, st.Repos[0].PinnedPath)
				assert.Equal(t, map[string]string{pin: "primary-token"}, f.git.fetched, "the adoption is refreshed in the same apply")
			} else {
				assert.Empty(t, st.Repos)
				assert.Empty(t, f.hub.about("/repos/"))
				assert.Empty(t, f.git.fetched)
			}
		})
	}
}

// A record of an excluded profile that a selected profile lists is fetched
// with the selected profile's credentials, which are recorded only on a
// successful apply (row 4).
func TestSelectionProfileListsAnotherRecord(t *testing.T) {
	newFixture := func(t *testing.T) *selectionFixture {
		f := newSelectionFixture(t, state.Record{ID: 31, OwnerLogin: "acme", Name: "shared", SourceProfile: "primary"})
		f.hub.repos = []*hubRepo{{id: 31, owner: "acme", name: "shared"}}
		return f
	}
	sel := Selection{Profiles: []string{"beta"}}

	f := newFixture(t)
	_, err := f.run(t, false, sel, withOwners(beta, "acme"))
	require.NoError(t, err)
	f.git.failure = fmt.Errorf("injected fetch failure")
	_, err = f.run(t, true, sel, withOwners(beta, "acme"))
	require.Error(t, err)
	st, err := f.store.Load()
	require.NoError(t, err)
	assert.Equal(t, "primary", st.Repos[0].SourceProfile, "neither a plan nor a failed apply records the profile")

	f.git.failure = nil
	plan, err := f.run(t, true, sel, withOwners(beta, "acme"))
	require.NoError(t, err)
	assert.Equal(t, []string{"fetch: "}, actionOf(t, plan, 31))
	assert.Equal(t, "beta-token", f.git.fetched[f.path("", "acme", "shared")])
	st, err = f.store.Load()
	require.NoError(t, err)
	assert.Equal(t, "beta", st.Repos[0].SourceProfile)
	assert.Empty(t, f.hub.about("/repositories/"))
}

// --profile beta leaves a private record of primary alone: no confirmation
// with beta's token, and no "gone" observation.
func TestSelectionProfileKeepsOtherProfileRecords(t *testing.T) {
	f := newSelectionFixture(t, state.Record{ID: 41, OwnerLogin: "acme", Name: "private", Visibility: "private", SourceProfile: "primary"})
	f.hub.repos = []*hubRepo{{id: 41, owner: "acme", name: "private", private: true, tokens: []string{"primary-token"}}}
	before := f.records(t)

	plan, err := f.run(t, true, Selection{Profiles: []string{"beta"}}, withOwners(beta, "acme"))
	require.NoError(t, err)
	assert.Equal(t, []string{"noop: outside the selection"}, actionOf(t, plan, 41))
	assert.Empty(t, f.hub.about("/repositories/41", "/repos/acme/private"))
	assert.Equal(t, before, f.records(t))
}

// --profile beta --owner acme: a repository of other that beta's listing
// returns gets no request, no mutation and no state change.
func TestSelectionProfileAndOwnerTogether(t *testing.T) {
	f := newSelectionFixture(t,
		state.Record{ID: 51, OwnerLogin: "acme", Name: "svc", SourceProfile: "beta"},
		state.Record{ID: 52, OwnerLogin: "other", Name: "listed", SourceProfile: "beta"},
	)
	f.hub.logins["beta-token"] = "other" // beta's own repositories belong to other
	f.hub.orgs["beta-token"] = []string{"acme"}
	f.hub.repos = []*hubRepo{{id: 51, owner: "acme", name: "svc"}, {id: 52, owner: "other", name: "listed"}}
	before := f.records(t)

	plan, err := f.run(t, true, Selection{Owners: []string{"acme"}, Profiles: []string{"beta"}}, withOwners(beta, "acme"))
	require.NoError(t, err)
	assert.Equal(t, []string{"noop: outside the selection"}, actionOf(t, plan, 52))
	assert.Empty(t, f.hub.about("/repositories/52", "/repos/other/"))
	assert.NotContains(t, f.git.fetched, f.path("", "other", "listed"))
	assert.Equal(t, before[52], f.records(t)[52])
}

// A 404 by id is "not accessible" only from the record's own profile: with
// another token the record stays unverified and nothing is saved.
func TestConfirmationNeedsTheRecordProfile(t *testing.T) {
	f := newSelectionFixture(t, state.Record{ID: 61, OwnerLogin: "acme", Name: "private", Visibility: "private", SourceProfile: "retired"})
	f.hub.repos = []*hubRepo{{id: 61, owner: "acme", name: "private", private: true, tokens: []string{"retired-token"}}}
	before := f.records(t)

	plan, err := f.run(t, true, Selection{}, withOwners(primary, "acme"))
	require.NoError(t, err)
	assert.Equal(t, []string{`noop: not verified: checked with profile "primary", not with its own profile "retired"`}, actionOf(t, plan, 61))
	assert.Len(t, f.hub.about("/repositories/61"), 1)
	assert.Equal(t, before, f.records(t))
	assert.Equal(t, "unverified", plan.Actions[0].Flag)

	// The human plan shows the row and its reason.
	discoverer := github.NewRESTDiscoverer(f.hub.factory)
	var human bytes.Buffer
	svc, err := NewService(f.cnf, []ResolvedProfile{withOwners(primary, "acme")}, f.root, f.root, 1, Deps{
		Store: f.store, Discoverer: discoverer, Confirmer: discoverer, Resolver: discoverer,
		GitSync: f.git, Reporter: NewReporter(&human, &human, FormatHuman, 0, false),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	assert.Contains(t, human.String(), `acme/private`)
	assert.Contains(t, human.String(), `not verified: checked with profile "primary"`)

	// The record's own profile establishes it.
	f.hub.repos[0].tokens = nil
	f.hub.repos = nil
	_, err = f.run(t, true, Selection{}, ResolvedProfile{Name: "retired", Token: "retired-token", Owners: []string{"acme"}})
	require.NoError(t, err)
	st, err := f.store.Load()
	require.NoError(t, err)
	assert.Equal(t, "gone", st.Repos[0].RemoteStatus)
}

// Configured owners, owner rules and repository rules match regardless of
// case; a full run still manages checkouts outside include_owners.
func TestSelectionCaseInsensitiveOwners(t *testing.T) {
	f := newSelectionFixture(t,
		state.Record{ID: 71, OwnerLogin: "acme", Name: "svc", SourceProfile: "primary"},
		state.Record{ID: 72, OwnerLogin: "other", Name: "kept", SourceProfile: "primary"},
	)
	f.hub.repos = []*hubRepo{{id: 71, owner: "acme", name: "svc"}, {id: 72, owner: "other", name: "kept"}}
	f.cnf.Repos = []config.Repo{{Match: config.RepoMatch{Owner: "ACME", Name: "SVC"}, Ignore: true}}

	plan, err := f.run(t, true, Selection{Owners: []string{"Other"}}, withOwners(primary, "OTHER"))
	require.NoError(t, err)
	assert.Empty(t, actionOf(t, plan, 71), "the ignore rule matches acme/svc")
	assert.Equal(t, []string{"fetch: "}, actionOf(t, plan, 72))
	assert.Contains(t, f.hub.about("/orgs/other/repos"), "primary-token GET /orgs/other/repos?per_page=100&type=all")
}
