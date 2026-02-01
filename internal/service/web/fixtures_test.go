package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/hub"
	"go.octolab.org/toolset/maintainer/internal/service/status"
	"go.octolab.org/toolset/maintainer/internal/service/web"
)

var update = flag.Bool("update", false, "rewrite the JSON fixtures of the frontend from the Go types")

const fixturesDir = "../../../web/src/api/fixtures"

// The JSON fixtures the frontend tests run against are what the Go handler
// answers for the fixture mirror: `go test ./internal/service/web -run
// TestFrontendFixtures -update` rewrites them, and without the flag every
// fixture must decode into the Go types without unknown fields and encode
// back to the same bytes.
func TestFrontendFixtures(t *testing.T) {
	base := t.TempDir()
	roots := fixtureRootsUnder(t, base)
	weights := config.Weights{Impact: 1, Effort: 1, Type: 1, Age: 1, Due: 1, AgeHorizonDays: 365, DueHorizonDays: 30}
	index := hub.New(func(context.Context) ([]hub.Root, error) { return roots, nil }, weights, func(r hub.Root) float64 {
		if r.Name == "tool" {
			return 2
		}
		return 1
	})
	now := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	index.SetNow(func() time.Time { return now })
	require.NoError(t, index.Scan(context.Background()))
	collected := make(chan []status.Row, 1)
	collected <- []status.Row{
		{ID: 1, Repository: "octomation/maintainer", Path: "/home/me/Development/public/octomation/maintainer", Branch: "main", DefaultBranch: "main", Commit: "ce2936a0000000000000000000000000000000000", Upstream: "origin/main", Added: 12, Deleted: 3, Changed: 2, Untracked: 1, Status: "ahead", Ahead: 1},
		{ID: 2, Repository: "octomation/tool", Path: "/home/me/Development/private/octomation/tool", Branch: "feature", DefaultBranch: "main", Commit: "abc1234000000000000000000000000000000000", Upstream: "origin/feature", Status: "diverged", Ahead: 2, Behind: 1, PushLocked: true, Pinned: true, Visibility: "private"},
		{Repository: "kamilsk/dotfiles", Path: "/home/me/.dotfiles", Branch: "main", DefaultBranch: "main", Status: "synced", OrphanReason: "duplicate-pin", ActivePath: "/home/me/Development/public/kamilsk/dotfiles"},
		{Repository: "broken", Path: "/home/me/Development/broken", Status: "error", Error: "git status: not a git repository"},
	}
	checkouts := web.NewCheckouts(func(context.Context) ([]status.Row, error) { return <-collected, nil })
	checkouts.Refresh()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	checkouts.Start(ctx)
	require.Eventually(t, func() bool { _, _, ok, _, _ := checkouts.Snapshot(); return ok }, 5*time.Second, 10*time.Millisecond)
	checkouts.SetCollectedAt(now)

	handler, err := web.NewHandler(web.Options{Assets: assets, Index: index, Checkouts: checkouts})
	require.NoError(t, err)
	pending, err := web.NewHandler(web.Options{Assets: assets, Checkouts: web.NewCheckouts(func(context.Context) ([]status.Row, error) { select {} })})
	require.NoError(t, err)

	taskID := hub.EncodeID("github:octomation/maintainer#1")
	type route struct {
		handler http.Handler
		method  string
		path    string
		body    string
		want    int
		decode  any
	}
	routes := map[string]route{
		"tasks.json":             {handler, http.MethodGet, "/api/v0/tasks?repo=octomation/maintainer&state=all&sort=score", "", http.StatusOK, &envelopeOf[hub.Page]{}},
		"task.json":              {handler, http.MethodGet, "/api/v0/tasks/" + taskID, "", http.StatusOK, &envelopeOf[web.TaskDetail]{}},
		"catalog.json":           {handler, http.MethodGet, "/api/v0/catalog", "", http.StatusOK, &envelopeOf[hub.Catalog]{}},
		"checkouts.json":         {handler, http.MethodGet, "/api/v0/checkouts?sort=-status", "", http.StatusOK, &envelopeOf[web.CheckoutsData]{}},
		"checkouts-pending.json": {pending, http.MethodGet, "/api/v0/checkouts", "", http.StatusAccepted, &envelopeOf[*web.CheckoutsData]{}},
		"error.json":             {handler, http.MethodGet, "/api/v0/tasks/not-an-id!", "", http.StatusBadRequest, &web.ErrorBody{}},
	}
	for name, r := range routes {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(ctx, r.method, r.path, strings.NewReader(r.body))
			req.Host = "maintainer.localhost"
			req = req.WithContext(context.WithValue(req.Context(), http.LocalAddrContextKey, &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 80}))
			r.handler.ServeHTTP(w, req)
			require.Equal(t, r.want, w.Code, w.Body.String())
			var pretty bytes.Buffer
			require.NoError(t, json.Indent(&pretty, bytes.ReplaceAll(w.Body.Bytes(), []byte(base), []byte("/home/me/Development")), "", "  "))
			pretty.WriteByte('\n')
			path := filepath.Join(fixturesDir, name)
			if *update {
				require.NoError(t, os.MkdirAll(fixturesDir, 0o755))
				require.NoError(t, os.WriteFile(path, pretty.Bytes(), 0o644))
			}
			raw, err := os.ReadFile(path)
			require.NoError(t, err, "run go test ./internal/service/web -run TestFrontendFixtures -update")
			assert.Equal(t, pretty.String(), string(raw), "the fixture is stale: run go test ./internal/service/web -run TestFrontendFixtures -update")
			// The fixture decodes into the Go types without unknown fields
			// and encodes back to the same document.
			dec := json.NewDecoder(bytes.NewReader(raw))
			dec.DisallowUnknownFields()
			require.NoError(t, dec.Decode(r.decode), name)
			back, err := json.Marshal(r.decode)
			require.NoError(t, err)
			var want, got any
			require.NoError(t, json.Unmarshal(raw, &want))
			require.NoError(t, json.Unmarshal(back, &got))
			assert.Equal(t, want, got, "%s does not round-trip through the Go types", name)
		})
	}
}

type envelopeOf[T any] struct {
	Generation  uint64     `json:"generation"`
	CollectedAt *time.Time `json:"collected_at"`
	Refreshing  bool       `json:"refreshing"`
	Warnings    []string   `json:"warnings"`
	Data        T          `json:"data"`
}

// fixtureRootsUnder copies the hub fixture mirror under base.
func fixtureRootsUnder(t *testing.T, base string) []hub.Root {
	t.Helper()
	var roots []hub.Root
	for _, r := range []hub.Root{
		{Owner: "octomation", Name: "maintainer", Path: "public/octomation/maintainer", ID: 1},
		{Owner: "octomation", Name: "tool", Path: "private/octomation/tool", ID: 2, Visibility: "private"},
		{Owner: "kamilsk", Name: "dotfiles", Path: "public/kamilsk/dotfiles", ID: 3},
		{Owner: "kamilsk", Name: "dotfiles", Path: "public/kamilsk/dotfiles-copy"},
	} {
		src := filepath.Join("..", "hub", "testdata", "mirror", strings.TrimPrefix(strings.TrimPrefix(r.Path, "public/"), "private/"))
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
