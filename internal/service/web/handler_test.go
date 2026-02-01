package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/hub"
	"go.octolab.org/toolset/maintainer/internal/service/status"
	"go.octolab.org/toolset/maintainer/internal/service/web"
)

var assets = fstest.MapFS{
	"index.html":                {Data: []byte("<!doctype html><title>maintainer</title><script type=\"module\" src=\"/assets/app-abc123.js\"></script>")},
	"assets/app-abc123.js":      {Data: []byte("console.log('hub')")},
	"assets/app-abc123.css":     {Data: []byte("body{}")},
	"assets/inter-abc123.woff2": {Data: []byte("wOF2")},
}

// fixtureRoots copies the hub fixture mirror into a temporary directory.
func fixtureRoots(t *testing.T) []hub.Root {
	t.Helper()
	base := t.TempDir()
	var roots []hub.Root
	for _, r := range []hub.Root{
		{Owner: "octomation", Name: "maintainer", Path: "octomation/maintainer"},
		{Owner: "octomation", Name: "tool", Path: "octomation/tool"},
		{Owner: "kamilsk", Name: "dotfiles", Path: "kamilsk/dotfiles"},
		{Owner: "kamilsk", Name: "dotfiles", Path: "kamilsk/dotfiles-copy"},
	} {
		src := filepath.Join("..", "hub", "testdata", "mirror", r.Path)
		dst := filepath.Join(base, r.Path)
		require.NoError(t, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(src, path)
			if info.IsDir() {
				return os.MkdirAll(filepath.Join(dst, rel), 0o755)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dst, rel), raw, 0o644)
		}))
		r.Path = dst
		roots = append(roots, r)
	}
	return roots
}

type fixture struct {
	server    *httptest.Server
	index     *hub.Index
	checkouts *web.Checkouts
	collect   chan []status.Row
	port      string
}

func (f *fixture) url(path string) string { return f.server.URL + path }

// newFixture serves the handler on a real loopback listener, so the Host
// check sees the bound address, with the fixture index scanned.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{collect: make(chan []status.Row, 1)}
	roots := fixtureRoots(t)
	// MAIN-9 becomes a flow list over several lines: the one shape the
	// writer refuses, for the 422 case.
	nine := filepath.Join(roots[0].Path, ".github", "issues", "MAIN-9.md")
	raw, err := os.ReadFile(nine)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(nine, []byte(strings.Replace(string(raw), `labels: ["impact: high", "scope: code"]`, "labels: [\"impact: high\",\n  \"scope: code\"]", 1)), 0o644))
	weights := config.Weights{Impact: 1, Effort: 1, Type: 1, Age: 1, Due: 1, AgeHorizonDays: 365, DueHorizonDays: 30}
	f.index = hub.New(func(context.Context) ([]hub.Root, error) { return roots, nil }, weights, nil)
	require.NoError(t, f.index.Scan(context.Background()))
	f.checkouts = web.NewCheckouts(func(ctx context.Context) ([]status.Row, error) {
		select {
		case rows := <-f.collect:
			return rows, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	f.checkouts.Start(ctx)
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	handler, err := web.NewHandler(web.Options{Addrs: []net.Addr{listener.Addr()}, Assets: assets, Index: f.index, Checkouts: f.checkouts})
	require.NoError(t, err)
	f.server = httptest.NewUnstartedServer(handler)
	_ = f.server.Listener.Close()
	f.server.Listener = listener
	f.server.Start()
	t.Cleanup(f.server.Close)
	_, f.port, _ = net.SplitHostPort(listener.Addr().String())
	return f
}

type response struct {
	*http.Response
	body []byte
}

func (r response) json(t *testing.T, target any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(r.body, target), string(r.body))
}

func (r response) envelope(t *testing.T) (web.Envelope, json.RawMessage) {
	t.Helper()
	var env struct {
		web.Envelope
		Data json.RawMessage `json:"data"`
	}
	r.json(t, &env)
	return env.Envelope, env.Data
}

func (r response) errorCode(t *testing.T) string {
	t.Helper()
	var body web.ErrorBody
	r.json(t, &body)
	return body.Error.Code
}

type option func(*http.Request)

func header(key, value string) option { return func(r *http.Request) { r.Header.Set(key, value) } }
func host(h string) option            { return func(r *http.Request) { r.Host = h } }

func mutation(f *fixture) option {
	return func(r *http.Request) {
		r.Header.Set("Origin", "http://"+r.Host)
		r.Header.Set("X-Requested-With", "maintainer")
	}
}

func (f *fixture) do(t *testing.T, method, path string, body []byte, opts ...option) response {
	t.Helper()
	var reader io.Reader = http.NoBody
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, f.url(path), reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, opt := range opts {
		opt(req)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return response{Response: resp, body: raw}
}

func TestPagesAndAssets(t *testing.T) {
	f := newFixture(t)
	for _, path := range web.Pages {
		resp := f.do(t, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode, path)
		assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"), path)
		assert.Equal(t, "no-store", resp.Header.Get("Cache-Control"), path)
		assert.Contains(t, string(resp.body), "<title>maintainer</title>", path)
		head := f.do(t, http.MethodHead, path, nil)
		assert.Equal(t, http.StatusOK, head.StatusCode, path)
		assert.Empty(t, head.body)
	}
	root := f.do(t, http.MethodGet, "/", nil)
	assert.Equal(t, http.StatusFound, root.StatusCode)
	assert.Equal(t, "/status", root.Header.Get("Location"))

	asset := f.do(t, http.MethodGet, "/assets/app-abc123.js", nil)
	assert.Equal(t, http.StatusOK, asset.StatusCode)
	assert.Equal(t, "public, max-age=31536000, immutable", asset.Header.Get("Cache-Control"))
	assert.Contains(t, asset.Header.Get("Content-Type"), "javascript")
	assert.Equal(t, "console.log('hub')", string(asset.body))
	css := f.do(t, http.MethodGet, "/assets/app-abc123.css", nil)
	assert.Contains(t, css.Header.Get("Content-Type"), "text/css")
	font := f.do(t, http.MethodGet, "/assets/inter-abc123.woff2", nil)
	assert.Equal(t, http.StatusOK, font.StatusCode)
	assert.Equal(t, "font/woff2", font.Header.Get("Content-Type"))

	for _, path := range []string{"/index.html", "/assets/", "/assets/missing.js", "/status/x", "/og.png", "/favicon.ico", "/api/v0/tasks/"} {
		resp := f.do(t, http.MethodGet, path, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, path)
	}
	escaped := f.do(t, http.MethodGet, "/assets/../index.html", nil)
	assert.Equal(t, http.StatusTemporaryRedirect, escaped.StatusCode, "the mux cleans the path")
	assert.Equal(t, "/index.html", escaped.Header.Get("Location"))
	missing := f.do(t, http.MethodGet, "/api/v0/nothing", nil)
	assert.Equal(t, http.StatusNotFound, missing.StatusCode)
	assert.Equal(t, "not_found", missing.errorCode(t))

	t.Run("headers on every answer", func(t *testing.T) {
		for _, path := range []string{"/status", "/", "/assets/app-abc123.js", "/missing", "/api/v0/catalog", "/api/v0/nothing"} {
			resp := f.do(t, http.MethodGet, path, nil)
			assert.Equal(t, "default-src 'self'; img-src 'self' data:; frame-ancestors 'none'", resp.Header.Get("Content-Security-Policy"), path)
			assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), path)
			assert.Equal(t, "no-referrer", resp.Header.Get("Referrer-Policy"), path)
		}
		api := f.do(t, http.MethodGet, "/api/v0/catalog", nil)
		assert.Equal(t, "no-store", api.Header.Get("Cache-Control"))
	})

	t.Run("method not allowed", func(t *testing.T) {
		resp := f.do(t, http.MethodPost, "/status", nil, mutation(f))
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		assert.Equal(t, "GET, HEAD", resp.Header.Get("Allow"))
		resp = f.do(t, http.MethodGet, "/api/v0/refresh", nil)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		assert.Equal(t, "POST", resp.Header.Get("Allow"))
		resp = f.do(t, http.MethodDelete, "/api/v0/tasks", nil, mutation(f))
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		assert.Equal(t, "GET, HEAD", resp.Header.Get("Allow"))
		assert.Equal(t, "method_not_allowed", resp.errorCode(t))
		resp = f.do(t, http.MethodPost, "/api/v0/tasks/"+hub.EncodeID("github:a/b#1")+"/triage", nil, mutation(f))
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		assert.Equal(t, "PATCH", resp.Header.Get("Allow"))
	})
}

func TestHostAndOrigin(t *testing.T) {
	f := newFixture(t)
	t.Run("host", func(t *testing.T) {
		for _, h := range []string{"maintainer.localhost:" + f.port, "localhost:" + f.port, "127.0.0.1:" + f.port, "MAINTAINER.LOCALHOST:" + f.port} {
			resp := f.do(t, http.MethodGet, "/api/v0/catalog", nil, host(h))
			assert.Equal(t, http.StatusOK, resp.StatusCode, h)
		}
		for _, h := range []string{"evil.example:" + f.port, "maintainer.localhost", "maintainer.localhost:1", "127.0.0.2:" + f.port, "[::1]:" + f.port, "10.0.0.1:" + f.port, "maintainer.localhost.evil.example:" + f.port} {
			resp := f.do(t, http.MethodGet, "/api/v0/catalog", nil, host(h))
			assert.Equal(t, http.StatusMisdirectedRequest, resp.StatusCode, h)
			assert.Equal(t, "misdirected", resp.errorCode(t), h)
		}
		forwarded := f.do(t, http.MethodGet, "/api/v0/catalog", nil, host("evil.example:"+f.port), header("X-Forwarded-Host", "maintainer.localhost:"+f.port))
		assert.Equal(t, http.StatusMisdirectedRequest, forwarded.StatusCode, "forwarded headers are ignored")
	})

	t.Run("origin", func(t *testing.T) {
		with := func(origin string) option { return header("Origin", origin) }
		requested := header("X-Requested-With", "maintainer")
		valid := "http://maintainer.localhost:" + f.port
		h := host("maintainer.localhost:" + f.port)
		for name, opts := range map[string][]option{
			"missing origin":          {h, requested},
			"null origin":             {h, requested, with("null")},
			"foreign origin":          {h, requested, with("http://evil.example")},
			"other host of this port": {h, requested, with("http://localhost:" + f.port)},
			"https origin":            {h, requested, with("https://maintainer.localhost:" + f.port)},
			"missing requested-with":  {h, with(valid)},
			"wrong requested-with":    {h, with(valid), header("X-Requested-With", "XMLHttpRequest")},
		} {
			resp := f.do(t, http.MethodPost, "/api/v0/refresh", nil, opts...)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode, name)
			assert.Equal(t, "forbidden", resp.errorCode(t), name)
		}
		resp := f.do(t, http.MethodPost, "/api/v0/refresh", nil, h, requested, with(valid))
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		env, _ := resp.envelope(t)
		assert.True(t, env.Refreshing)
		resp = f.do(t, http.MethodPost, "/api/v0/checkouts/refresh", nil, h, requested, with(strings.ToUpper(valid)))
		assert.Equal(t, http.StatusAccepted, resp.StatusCode, "the origin compares case-insensitively")
		get := f.do(t, http.MethodGet, "/api/v0/catalog", nil, h, with("http://evil.example"))
		assert.Equal(t, http.StatusOK, get.StatusCode, "reads need no Origin")
	})
}

func TestCheckoutsAPI(t *testing.T) {
	f := newFixture(t)
	first := f.do(t, http.MethodGet, "/api/v0/checkouts", nil)
	assert.Equal(t, http.StatusAccepted, first.StatusCode)
	env, data := first.envelope(t)
	assert.True(t, env.Refreshing)
	assert.Nil(t, env.CollectedAt)
	assert.Equal(t, "null", string(data))
	assert.Equal(t, []string{}, env.Warnings)

	f.collect <- []status.Row{
		{Repository: "acme/tool", Path: "/w/acme/tool", Branch: "main", DefaultBranch: "main", Added: 3, Status: "synced"},
		{Repository: "acme/lib", Path: "/w/acme/lib", Branch: "feature", Ahead: 2, Status: "ahead", PushLocked: true},
		{Repository: "zeta/app", Path: "/w/zeta/app", Status: "error", Error: "boom"},
	}
	require.Eventually(t, func() bool {
		return f.do(t, http.MethodGet, "/api/v0/checkouts", nil).StatusCode == http.StatusOK
	}, 5*time.Second, 10*time.Millisecond)

	resp := f.do(t, http.MethodGet, "/api/v0/checkouts?sort=-status,repository&q=acme", nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	env, data = resp.envelope(t)
	assert.False(t, env.Refreshing)
	require.NotNil(t, env.CollectedAt)
	var checkouts web.CheckoutsData
	require.NoError(t, json.Unmarshal(data, &checkouts))
	assert.Equal(t, 3, checkouts.Total)
	assert.Equal(t, "acme", checkouts.Query)
	assert.Equal(t, []web.SortSpec{{Column: "status", Desc: true}, {Column: "repository"}}, checkouts.Sort)
	require.Len(t, checkouts.Rows, 2)
	assert.Equal(t, "acme/lib", checkouts.Rows[0].Repository, "ahead 2 sorts before synced descending")
	assert.Equal(t, []string{"acme/lib", "feature", "+0/-0", "ahead 2", "🔒", "/w/acme/lib"}, checkouts.Rows[0].Cells)
	assert.Equal(t, status.Summary{Checkouts: 2, Uncommitted: 1, Ahead: 1, Locked: 1}, checkouts.Summary)

	bad := f.do(t, http.MethodGet, "/api/v0/checkouts?sort=changes", nil)
	assert.Equal(t, http.StatusBadRequest, bad.StatusCode)
	assert.Equal(t, "bad_request", bad.errorCode(t))

	refresh := f.do(t, http.MethodPost, "/api/v0/checkouts/refresh", nil, mutation(f))
	assert.Equal(t, http.StatusAccepted, refresh.StatusCode)
	env, _ = refresh.envelope(t)
	assert.True(t, env.Refreshing)
	assert.NotNil(t, env.CollectedAt)
	during := f.do(t, http.MethodGet, "/api/v0/checkouts", nil)
	assert.Equal(t, http.StatusOK, during.StatusCode, "the cached rows stay during a refresh")
	env, _ = during.envelope(t)
	assert.True(t, env.Refreshing)
	f.collect <- []status.Row{{Repository: "acme/tool", Path: "/w/acme/tool", Branch: "main", Status: "synced"}}
	require.Eventually(t, func() bool {
		env, _ := f.do(t, http.MethodGet, "/api/v0/checkouts", nil).envelope(t)
		return !env.Refreshing
	}, 5*time.Second, 10*time.Millisecond)
	_, data = f.do(t, http.MethodGet, "/api/v0/checkouts", nil).envelope(t)
	require.NoError(t, json.Unmarshal(data, &checkouts))
	assert.Equal(t, 1, checkouts.Total)
}

func TestTasksAPI(t *testing.T) {
	f := newFixture(t)
	t.Run("list", func(t *testing.T) {
		resp := f.do(t, http.MethodGet, "/api/v0/tasks?repo=octomation/maintainer&state=all&limit=5&sort=updated", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		env, data := resp.envelope(t)
		assert.Equal(t, uint64(1), env.Generation)
		assert.NotNil(t, env.CollectedAt)
		assert.Len(t, env.Warnings, 2)
		var page hub.Page
		require.NoError(t, json.Unmarshal(data, &page))
		assert.Equal(t, 8, page.Total)
		assert.Len(t, page.Tasks, 5)
		assert.Equal(t, 5, page.Limit)
		assert.Equal(t, "github:octomation/maintainer#4", page.Tasks[0].Key)
		assert.NotContains(t, string(data), "body_html")
		assert.Contains(t, string(data), `"issue_fields"`)
		assert.Contains(t, string(data), `"columns":[`)
		for _, field := range []string{`"labels":null`, `"assignees":null`, `"boards":null`, `"tasks":null`, `"warnings":null`} {
			assert.NotContains(t, string(resp.body), field, "arrays are never null")
		}
		for name, query := range map[string]string{
			"unknown sort": "sort=title", "bad state": "state=done", "bad limit": "limit=ten", "negative offset": "offset=-1", "bad column": "column=doing",
		} {
			bad := f.do(t, http.MethodGet, "/api/v0/tasks?"+query, nil)
			assert.Equal(t, http.StatusBadRequest, bad.StatusCode, name)
			assert.Equal(t, "bad_request", bad.errorCode(t), name)
		}
		closed := f.do(t, http.MethodGet, "/api/v0/tasks?state=closed&column=wont_do&label=type:%20chore", nil)
		_, data = closed.envelope(t)
		require.NoError(t, json.Unmarshal(data, &page))
		assert.Equal(t, 1, page.Total)
	})

	t.Run("catalog", func(t *testing.T) {
		resp := f.do(t, http.MethodGet, "/api/v0/catalog", nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		_, data := resp.envelope(t)
		var catalog hub.Catalog
		require.NoError(t, json.Unmarshal(data, &catalog))
		assert.Len(t, catalog.Repositories, 4)
		assert.Len(t, catalog.Boards, 3)
		assert.Equal(t, hub.Columns, catalog.Columns)
	})

	t.Run("detail by both id forms", func(t *testing.T) {
		for _, key := range []string{"github:octomation/maintainer#1", "github:octomation/maintainer:MAIN-3"} {
			resp := f.do(t, http.MethodGet, "/api/v0/tasks/"+hub.EncodeID(key), nil)
			require.Equal(t, http.StatusOK, resp.StatusCode, key)
			_, data := resp.envelope(t)
			var detail web.TaskDetail
			require.NoError(t, json.Unmarshal(data, &detail))
			assert.Equal(t, key, detail.Key)
			assert.Equal(t, `"`+detail.Version+`"`, resp.Header.Get("ETag"))
			assert.Contains(t, detail.BodyHTML, "<h1>")
			assert.True(t, strings.HasSuffix(detail.Provenance.Path, ".md"))
			assert.NotNil(t, detail.Comments)
		}
		missing := f.do(t, http.MethodGet, "/api/v0/tasks/"+hub.EncodeID("github:nobody/nothing#1"), nil)
		assert.Equal(t, http.StatusNotFound, missing.StatusCode)
		bad := f.do(t, http.MethodGet, "/api/v0/tasks/not-an-id!", nil)
		assert.Equal(t, http.StatusBadRequest, bad.StatusCode)
	})

	t.Run("triage", func(t *testing.T) {
		key := "github:octomation/maintainer#1"
		path := "/api/v0/tasks/" + hub.EncodeID(key) + "/triage"
		current := f.do(t, http.MethodGet, "/api/v0/tasks/"+hub.EncodeID(key), nil)
		etag := current.Header.Get("ETag")
		patch := func(body string, opts ...option) response {
			t.Helper()
			return f.do(t, http.MethodPatch, path, []byte(body), append([]option{mutation(f)}, opts...)...)
		}

		resp := patch(`{"impact":"low"}`)
		assert.Equal(t, http.StatusPreconditionRequired, resp.StatusCode)
		assert.Equal(t, "precondition_required", resp.errorCode(t))

		resp = patch(`{"impact":"low"}`, header("If-Match", `"sha256-stale"`))
		assert.Equal(t, http.StatusPreconditionFailed, resp.StatusCode)
		assert.Equal(t, "precondition_failed", resp.errorCode(t))

		resp = patch(`{"impact":"low"}`, header("If-Match", `W/`+etag))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "a weak ETag")

		resp = f.do(t, http.MethodPatch, path, nil, mutation(f), header("If-Match", etag), header("Content-Type", "text/plain"))
		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)

		for name, body := range map[string]string{
			"unknown key":   `{"priority":"high"}`,
			"duplicate key": `{"impact":"low","impact":"high"}`,
			"bad value":     `{"impact":"urgent"}`,
			"bad effort":    `{"effort":"low"}`,
			"not an object": `["impact"]`,
			"trailing":      `{"impact":"low"} {}`,
			"number":        `{"impact":1}`,
		} {
			resp := patch(body, header("If-Match", etag))
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, name)
			assert.Equal(t, "bad_request", resp.errorCode(t), name)
		}
		huge := patch(`{"impact":"low","effort":"`+strings.Repeat("x", web.MaxBodyBytes)+`"}`, header("If-Match", etag))
		assert.Equal(t, http.StatusBadRequest, huge.StatusCode, "the body limit")

		resp = patch(`{"impact":"low","effort":null}`, header("If-Match", etag))
		require.Equal(t, http.StatusOK, resp.StatusCode, string(resp.body))
		env, data := resp.envelope(t)
		assert.Equal(t, uint64(2), env.Generation)
		var detail web.TaskDetail
		require.NoError(t, json.Unmarshal(data, &detail))
		assert.Equal(t, []string{"type: feature", "scope: code", "impact: low"}, detail.Labels)
		assert.NotEqual(t, etag, resp.Header.Get("ETag"))
		assert.Equal(t, `"`+detail.Version+`"`, resp.Header.Get("ETag"))
		assert.Equal(t, 0.25, detail.Score.Impact)
		raw, err := os.ReadFile(detail.Provenance.Path)
		require.NoError(t, err)
		assert.Contains(t, string(raw), "  - \"impact: low\"\n")
		assert.NotContains(t, string(raw), "effort: easy")

		stale := patch(`{"impact":"high"}`, header("If-Match", etag))
		assert.Equal(t, http.StatusPreconditionFailed, stale.StatusCode, "the old ETag is stale")

		again := patch(`{"effort":"hard"}`, header("If-Match", resp.Header.Get("ETag")))
		assert.Equal(t, http.StatusOK, again.StatusCode)

		dup := f.do(t, http.MethodGet, "/api/v0/tasks/"+hub.EncodeID("github:kamilsk/dotfiles#1"), nil)
		conflict := f.do(t, http.MethodPatch, "/api/v0/tasks/"+hub.EncodeID("github:kamilsk/dotfiles#1")+"/triage", []byte(`{"impact":"low"}`), mutation(f), header("If-Match", dup.Header.Get("ETag")))
		assert.Equal(t, http.StatusConflict, conflict.StatusCode)
		assert.Equal(t, "conflict", conflict.errorCode(t))

		flow := f.do(t, http.MethodGet, "/api/v0/tasks/"+hub.EncodeID("github:octomation/maintainer#9"), nil)
		unprocessable := f.do(t, http.MethodPatch, "/api/v0/tasks/"+hub.EncodeID("github:octomation/maintainer#9")+"/triage", []byte(`{"impact":"low"}`), mutation(f), header("If-Match", flow.Header.Get("ETag")))
		assert.Equal(t, http.StatusUnprocessableEntity, unprocessable.StatusCode)
		assert.Equal(t, "unprocessable", unprocessable.errorCode(t))

		missing := f.do(t, http.MethodPatch, "/api/v0/tasks/"+hub.EncodeID("github:nobody/nothing#1")+"/triage", []byte(`{"impact":"low"}`), mutation(f), header("If-Match", `"sha256-x"`))
		assert.Equal(t, http.StatusNotFound, missing.StatusCode)
	})

	t.Run("an external edit: 412, then the reload gets the current ETag", func(t *testing.T) {
		key := "github:octomation/maintainer#4"
		id := hub.EncodeID(key)
		current := f.do(t, http.MethodGet, "/api/v0/tasks/"+id, nil)
		_, data := current.envelope(t)
		var detail web.TaskDetail
		require.NoError(t, json.Unmarshal(data, &detail))
		raw, err := os.ReadFile(detail.Provenance.Path)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(detail.Provenance.Path, append(raw, "\r\nEdited outside.\r\n"...), 0o644))
		stale := f.do(t, http.MethodPatch, "/api/v0/tasks/"+id+"/triage", []byte(`{"impact":"high"}`), mutation(f), header("If-Match", current.Header.Get("ETag")))
		assert.Equal(t, http.StatusPreconditionFailed, stale.StatusCode)
		reloaded := f.do(t, http.MethodGet, "/api/v0/tasks/"+id, nil)
		require.Equal(t, http.StatusOK, reloaded.StatusCode)
		assert.NotEqual(t, current.Header.Get("ETag"), reloaded.Header.Get("ETag"), "the reload sees the file as it is")
		ok := f.do(t, http.MethodPatch, "/api/v0/tasks/"+id+"/triage", []byte(`{"impact":"high"}`), mutation(f), header("If-Match", reloaded.Header.Get("ETag")))
		require.Equal(t, http.StatusOK, ok.StatusCode, string(ok.body))
		after, err := os.ReadFile(detail.Provenance.Path)
		require.NoError(t, err)
		assert.Contains(t, string(after), "impact: high")
		assert.Contains(t, string(after), "Edited outside.")
	})
}

func TestBeforeTheFirstIndex(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	index := hub.New(func(ctx context.Context) ([]hub.Root, error) {
		close(started)
		<-release
		return nil, nil
	}, config.Weights{AgeHorizonDays: 365, DueHorizonDays: 30}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	index.Start(ctx, nil)
	<-started
	handler, err := web.NewHandler(web.Options{Assets: assets, Index: index})
	require.NoError(t, err)
	local := func(path string) *http.Request {
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
		req.Host = "maintainer.localhost"
		return req
	}
	for _, path := range []string{"/api/v0/tasks", "/api/v0/catalog", "/api/v0/tasks/" + hub.EncodeID("github:a/b#1")} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, local(path))
		assert.Equal(t, http.StatusAccepted, w.Code, path)
		var env web.Envelope
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
		assert.True(t, env.Refreshing)
		assert.Nil(t, env.Data)
	}
	close(release)
	require.Eventually(t, func() bool { return index.Snapshot() != nil }, 5*time.Second, 10*time.Millisecond)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, local("/api/v0/tasks"))
	assert.Equal(t, http.StatusOK, w.Code)

	t.Run("without index or checkouts", func(t *testing.T) {
		handler, err := web.NewHandler(web.Options{Assets: assets})
		require.NoError(t, err)
		for _, path := range []string{"/api/v0/tasks", "/api/v0/checkouts"} {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, local(path))
			assert.Equal(t, http.StatusServiceUnavailable, w.Code, path)
		}
		_, err = web.NewHandler(web.Options{Assets: fstest.MapFS{}})
		assert.Error(t, err, "no index.html")
	})
}
