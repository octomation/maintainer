package fetch_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

func fixedPlan() Plan {
	return Plan{
		ID:          "01HZTESTTESTTESTTESTTEST",
		GeneratedAt: time.Date(2026, 5, 31, 12, 34, 56, 0, time.UTC),
		Root:        "/work",
		StatePath:   "/work/.state/state.json",
		StateCount:  3,
		Discoveries: []DiscoverySummary{
			{Profile: "primary", Count: 2, Endpoints: []github.EndpointStat{{Endpoint: "/user/repos"}}},
		},
		Actions: []Action{
			{Kind: KindClone, ID: 999, Owner: "acme-labs", Name: "new-thing", Path: "/work/public/acme-labs/new-thing"},
			{Kind: KindFetch, ID: 100, Owner: "acme", Name: "service", Path: "/work/public/acme/service"},
			{
				Kind: KindMove, ID: 111, Owner: "acme-user", Name: "configs", FromName: "acme-user/dotfiles",
				FromPath: "/work/public/acme-user/dotfiles", ToPath: "/work/public/acme-user/configs", UpdateRemote: true,
				RemoteURL: "git@github.com:acme-user/configs.git",
			},
			{Kind: KindOrphan, ID: 222, Owner: "acme-tools", Name: "old-experiment", Path: "/work/public/acme-tools/old-experiment", Reason: "gone on GitHub (404); local clone retained"},
		},
	}
}

func TestReporter_Human(t *testing.T) {
	var out, errw bytes.Buffer
	r := NewReporter(&out, &errw, FormatHuman, 0, false)
	require.NoError(t, r.Render(fixedPlan(), false))

	got := out.String()
	// Routine fetch is collapsed; drift lines are shown.
	assert.Contains(t, got, "profile=primary   discovered=2 repos (/user/repos)")
	assert.Contains(t, got, "+ clone      acme-labs/new-thing")
	assert.NotContains(t, got, "~ fetch      acme/service") // collapsed
	assert.Contains(t, got, "↻ move       acme-user/dotfiles")
	assert.Contains(t, got, "+ update remote.origin.url")
	assert.Contains(t, got, "! orphan     acme-tools/old-experiment")
	assert.Contains(t, got, "<root>/public/acme-tools/old-experiment")
	assert.Contains(t, got, "summary: clone=1 fetch=1 move=1 relocate=0 update_remote=0 adopt=0 orphan=1 noop=0 conflict=0 errors=0")
	assert.Contains(t, got, "run with --apply to execute")
	assert.Empty(t, errw.String())
}

func TestReporter_ErrorfPlainWhenNotTTY(t *testing.T) {
	var out, errw bytes.Buffer
	r := NewReporter(&out, &errw, FormatHuman, 0, false)
	r.Errorf("clone %s: %v", "withsparkle/vscode", "remote repository is empty")

	got := errw.String()
	assert.Equal(t, "error: clone withsparkle/vscode: remote repository is empty\n", got)
	assert.NotContains(t, got, "\x1b[") // a buffer is not a terminal → no ANSI
}

func TestReporter_JSON(t *testing.T) {
	var out, errw bytes.Buffer
	r := NewReporter(&out, &errw, FormatJSON, 0, false)
	require.NoError(t, r.Render(fixedPlan(), false))

	var doc struct {
		PlanID      string `json:"plan_id"`
		GeneratedAt string `json:"generated_at"`
		Actions     []struct {
			Kind     string `json:"kind"`
			ID       int64  `json:"id"`
			From     string `json:"from"`
			To       string `json:"to"`
			FromPath string `json:"from_path"`
		} `json:"actions"`
		Summary struct {
			Clone, Fetch, Move, Orphan int
		} `json:"summary"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &doc))
	assert.Equal(t, "01HZTESTTESTTESTTESTTEST", doc.PlanID)
	assert.Equal(t, "2026-05-31T12:34:56Z", doc.GeneratedAt)
	// JSON lists every action including routine fetches.
	require.Len(t, doc.Actions, 4)
	assert.Equal(t, 1, doc.Summary.Fetch)
	assert.Equal(t, 1, doc.Summary.Clone)
	// move carries from/to identifiers and paths.
	for _, a := range doc.Actions {
		if a.Kind == "move" {
			assert.Equal(t, "acme-user/dotfiles", a.From)
			assert.Equal(t, "acme-user/configs", a.To)
			assert.Equal(t, "/work/public/acme-user/dotfiles", a.FromPath)
		}
	}
}
