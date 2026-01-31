package fetch_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

func TestPinnedPathNeedsSufficientContext(t *testing.T) {
	resolver := func(path string, match config.RepoMatch) *PathResolver {
		cnf := &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath},
			Repos: []config.Repo{{Match: match, Path: path}}}
		r, err := NewPathRenderer("/work", home, cwd)
		require.NoError(t, err)
		return NewPathResolver(cnf, r)
	}
	byID := config.RepoMatch{ID: 7}
	byName := config.RepoMatch{Owner: "acme", Name: "special"}
	rec := &state.Record{ID: 7, OwnerLogin: "acme", Name: "special", Visibility: "private", DefaultBranch: "main", IsFork: true}
	identity := github.RepoSnapshot{ID: 7, Owner: "acme", Name: "special"}
	for _, tc := range []struct {
		name  string
		path  string
		match config.RepoMatch
		snap  github.RepoSnapshot
		rec   *state.Record
		want  string
	}{
		{"literal without state", "/pins/x", byID, github.RepoSnapshot{ID: 7}, nil, "/pins/x"},
		{"root without state", "{{.Root}}/x", byID, github.RepoSnapshot{ID: 7}, nil, "/work/x"},
		{"name rule without state", "/pins/{{.Owner}}/{{.Repo}}", byName, github.RepoSnapshot{Owner: "acme", Name: "special"}, nil, "/pins/acme/special"},
		{"id rule with identity", "/pins/{{.Owner}}/{{.Repo}}", byID, identity, nil, "/pins/acme/special"},
		{"metadata from the record", "/pins/{{.Visibility}}/{{.DefaultBranch}}/{{if .IsFork}}fork{{end}}", byID, identity, rec, "/pins/private/main/fork"},
		{"metadata from GitHub", "/pins/{{.Visibility}}/{{.Repo}}", byID, github.RepoSnapshot{ID: 7, Owner: "acme", Name: "special", Visibility: github.Public}, rec, "/pins/public/special"},
		{"dot of a with block", "{{with .Root}}{{.}}/pins/tool{{end}}", byID, github.RepoSnapshot{ID: 7}, nil, "/work/pins/tool"},
		{"dot of a called template", `{{define "base"}}{{.}}/pins{{end}}{{template "base" .Root}}/tool`, byID, github.RepoSnapshot{ID: 7}, nil, "/work/pins/tool"},
		{"root variable in a with block", "{{with .Root}}{{$.Root}}/pins/tool{{end}}", byID, github.RepoSnapshot{ID: 7}, nil, "/work/pins/tool"},
		{"root variable of a called template", `{{define "base"}}{{$}}/pins{{end}}{{template "base" .Root}}/tool`, byID, github.RepoSnapshot{ID: 7}, nil, "/work/pins/tool"},
		{"root variable forwarded", `{{define "a"}}{{$}}{{end}}{{define "b"}}{{template "a" $}}/pins{{end}}{{template "b" .Root}}/tool`, byID, github.RepoSnapshot{ID: 7}, nil, "/work/pins/tool"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolver(tc.path, tc.match).PinnedPath(tc.snap, tc.rec)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
	for _, tc := range []struct {
		name, path string
		snap       github.RepoSnapshot
		want       string
	}{
		{"id rule without state", "/pins/{{.Owner}}/{{.Repo}}", github.RepoSnapshot{ID: 7}, "Owner, Repo"},
		{"identity is not metadata", "/pins/{{.Visibility}}/{{.Repo}}", identity, "Visibility"},
		{"condition on a flag", "/pins/{{if .IsArchived}}old/{{end}}{{.Repo}}", identity, "IsArchived"},
		{"bare dot", "/pins/{{.}}", identity, "DefaultBranch, IsArchived, IsFork, IsTemplate, Visibility"},
		{"dot of a with block", "/pins/{{with .Owner}}{{.}}{{end}}", github.RepoSnapshot{ID: 7}, "Owner"},
		{"whole context passed to a template", `{{define "v"}}{{.Visibility}}{{end}}/pins/{{template "v" .}}`, identity, "Visibility"},
		{"field passed to a template", `{{define "v"}}{{.}}{{end}}/pins/{{template "v" .IsFork}}`, identity, "IsFork"},
		{"root variable of the whole context", `{{define "v"}}{{$.Visibility}}{{end}}/pins/{{template "v" .}}`, identity, "Visibility"},
		{"root variable inside a with block", "/pins/{{with .Root}}{{$.IsTemplate}}{{end}}", identity, "IsTemplate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resolver(tc.path, byID).PinnedPath(tc.snap, nil)
			assert.ErrorIs(t, err, ErrUnresolvedPin)
			assert.ErrorContains(t, err, "references unknown "+tc.want)
		})
	}
	// An applied pin is the fallback while the rule cannot be rendered.
	legacy := &state.Record{ID: 7, OwnerLogin: "acme", Name: "special", PinnedPath: "/pins/public/special"}
	got, err := resolver("/pins/{{.Visibility}}/{{.Repo}}", byID).PinnedPath(identity, legacy)
	require.NoError(t, err)
	assert.Equal(t, legacy.PinnedPath, got)
}

func TestPlanner_PinAndPathErrorsStayPerRepository(t *testing.T) {
	cnf := &config.Fetch{
		Workspace: &config.Workspace{Root: "/work", Path: "{{.Owner}}/{{.Repo}}@{{.DefaultBranch}}"},
		Defaults:  config.Defaults{CloneURL: "ssh"},
		Repos:     []config.Repo{{Match: config.RepoMatch{ID: 9}, Path: "/pins/{{.Visibility}}/{{.Repo}}"}},
	}
	st := state.New()
	st.Upsert(state.Record{ID: 9, OwnerLogin: "acme", Name: "legacy", Path: "/pins/public/legacy"})
	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "ok", Visibility: github.Public, DefaultBranch: "main"},
		{ID: 2, Owner: "acme", Name: "odd", Visibility: github.Public, DefaultBranch: "release/v1"},
	}})
	require.NoError(t, err)
	require.Len(t, acts, 3)
	assert.Equal(t, KindClone, acts[0].Kind)
	assert.Equal(t, "/work/acme/ok@main", acts[0].Path)
	assert.Equal(t, KindConflict, acts[1].Kind)
	assert.Equal(t, int64(2), acts[1].ID)
	assert.Contains(t, acts[1].Reason, `cannot resolve the target path: rendered path "/work/acme/odd@release/v1" is outside workspace layout`)
	assert.Equal(t, KindConflict, acts[2].Kind)
	assert.Equal(t, int64(9), acts[2].ID)
	assert.Equal(t, `unresolved pin: path "/pins/{{.Visibility}}/{{.Repo}}" references unknown Visibility`, acts[2].Reason)
}

// A per-repo path that does not parse is a conflict of its repository, with
// or without a record: the applied pin is no fallback for a broken rule.
func TestPlanner_BrokenPinTemplateIsConflict(t *testing.T) {
	cnf := &config.Fetch{
		Workspace: &config.Workspace{Root: "/work", Path: "{{.Owner}}/{{.Repo}}"},
		Defaults:  config.Defaults{CloneURL: "ssh"},
		Repos:     []config.Repo{{Match: config.RepoMatch{Owner: "acme", Name: "broken"}, Path: "~/ext/{{.Owner"}},
	}
	snapshots := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "ok", Visibility: github.Public, DefaultBranch: "main"},
		{ID: 2, Owner: "acme", Name: "broken", Visibility: github.Public, DefaultBranch: "main"},
	}
	recorded := state.New()
	recorded.Upsert(state.Record{ID: 2, OwnerLogin: "acme", Name: "broken", Visibility: "public", Path: "/pins/broken", PinnedPath: "/pins/broken"})
	for _, st := range []*state.State{state.New(), recorded} {
		acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Snapshots: snapshots})
		require.NoError(t, err)
		require.Len(t, acts, 2)
		assert.Equal(t, KindClone, acts[0].Kind)
		assert.Equal(t, "/work/acme/ok", acts[0].Path)
		assert.Equal(t, KindConflict, acts[1].Kind)
		assert.Equal(t, int64(2), acts[1].ID)
		assert.Equal(t, `parse path template "~/ext/{{.Owner": template: path:1: unclosed action`, acts[1].Reason)
	}
}

// Changing a per-repo pin selects the other verified checkout; the previous
// one stays in the diagnostics: in previous_paths and as an orphan row.
func TestPlanner_ChangedPinSwitchesCheckout(t *testing.T) {
	const previous, next = "/pins/public/tool", "/pins/tool"
	cnf := &config.Fetch{Defaults: config.Defaults{Root: "/work", Path: config.DefaultPath, CloneURL: "ssh"},
		Repos: []config.Repo{{Match: config.RepoMatch{ID: 1}, Path: "/pins/{{.Repo}}"}}}
	st := state.New()
	st.Upsert(state.Record{ID: 1, OwnerLogin: "acme", Name: "tool", Visibility: "public", Path: previous, PinnedPath: previous,
		RemoteURL: "git@github.com:acme/tool.git", CloneURL: "ssh", FirstSeen: clock()})
	clones := []DiskClone{
		{ID: 1, Path: previous, Transport: "ssh", RemoteURL: "git@github.com:acme/tool.git", Pinned: true},
		{ID: 1, Path: next, Transport: "ssh", RemoteURL: "git@github.com:acme/tool.git", Pinned: true},
	}
	acts, err := newPlanner(t, cnf).Plan(PlanInput{State: st, Clones: clones, KnownCheckouts: map[string]bool{previous: true},
		Snapshots: []github.RepoSnapshot{{ID: 1, Owner: "acme", Name: "tool", Visibility: github.Public}}})
	require.NoError(t, err)
	require.Len(t, acts, 3)
	assert.Equal(t, KindAdopt, acts[0].Kind)
	assert.Equal(t, next, acts[0].Path)
	assert.Equal(t, next, acts[0].Record.PinnedPath)
	assert.Equal(t, []string{previous}, acts[0].Record.PreviousPaths)
	assert.Equal(t, clock(), acts[0].Record.FirstSeen)
	assert.Equal(t, KindAdopt, acts[1].After)
	assert.Equal(t, next, acts[1].Path)
	assert.Equal(t, KindOrphan, acts[2].Kind)
	assert.Equal(t, previous, acts[2].Path)
	assert.Equal(t, next, acts[2].ActivePath)
}

// A templated per-repo pin adopted on the first apply is refreshed in that
// apply and only fetched on the second: every run renders it from the record,
// never from {id, owner, name}.
func TestService_TemplatedPinSettlesAfterAdoption(t *testing.T) {
	root := t.TempDir()
	pin := filepath.Join(root, "pins/public/special")
	makeClone(t, pin, "git@github.com:acme/special.git")
	cnf := &config.Fetch{
		Workspace: &config.Workspace{Root: root},
		Defaults:  config.Defaults{CloneURL: "ssh", Concurrency: 1},
		Repos:     []config.Repo{{Match: config.RepoMatch{ID: 7}, Path: "pins/{{.Visibility}}/{{.Repo}}"}},
	}
	require.NoError(t, cnf.Validate())
	snap := github.RepoSnapshot{ID: 7, Owner: "acme", Name: "special", Visibility: github.Public, DefaultBranch: "main", SourceProfile: "p"}
	store := state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil)
	api := renamedRepository{snapshot: snap}
	for i, kinds := range [][]string{{"adopt", "fetch"}, {"fetch"}, {"adopt", "fetch"}, {"fetch"}} {
		if i == 2 {
			// The repository becomes private and its owner moves the checkout:
			// the current rule selects it, the old rendering is not scanned.
			snap.Visibility = github.Private
			api.snapshot = snap
			moved := filepath.Join(root, "pins/private/special")
			require.NoError(t, os.MkdirAll(filepath.Dir(moved), 0o755))
			require.NoError(t, os.Rename(pin, moved))
			pin = moved
		}
		var out, errw bytes.Buffer
		svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
			Store: store, Discoverer: fakeDiscoverer{snaps: []github.RepoSnapshot{snap}}, Confirmer: api, Resolver: api,
			GitSync: noNetworkFetch{gitsvc.NewSync()}, Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
		})
		require.NoError(t, err)
		require.NoError(t, svc.Run(context.Background(), true), "run %d: %s", i+1, out.String())
		plan := decodePlan(t, &out)
		require.Len(t, plan.Actions, len(kinds), out.String())
		for j, kind := range kinds {
			assert.Equal(t, kind, plan.Actions[j].Kind)
			assert.Equal(t, pin, plan.Actions[j].Path)
		}
		assert.Zero(t, plan.Summary.Conflict)
	}
}

func TestService_UnrenderablePathIsOneConflict(t *testing.T) {
	root := t.TempDir()
	cnf := &config.Fetch{
		Workspace: &config.Workspace{Root: root, Path: "{{.Owner}}/{{.Repo}}@{{.DefaultBranch}}"},
		Defaults:  config.Defaults{CloneURL: "https", Concurrency: 1},
	}
	require.NoError(t, cnf.Validate())
	snaps := []github.RepoSnapshot{
		{ID: 1, Owner: "acme", Name: "ok", Visibility: github.Public, DefaultBranch: "main", SourceProfile: "p"},
		{ID: 2, Owner: "acme", Name: "odd", Visibility: github.Public, DefaultBranch: "release/v1", SourceProfile: "p"},
	}
	var out, errw bytes.Buffer
	svc, err := NewService(cnf, []ResolvedProfile{{Name: "p", Owners: []string{"acme"}}}, root, root, 1, Deps{
		Store:      state.NewStore(afero.NewOsFs(), filepath.Join(t.TempDir(), "state.json"), nil),
		Discoverer: fakeDiscoverer{snaps: snaps}, Resolver: fakeGitHubNameResolver{}, GitSync: gitsvc.NewSync(),
		Reporter: NewReporter(&out, &errw, FormatJSON, 0, false),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Run(context.Background(), false))
	plan := decodePlan(t, &out)
	require.Len(t, plan.Actions, 2)
	assert.Equal(t, "clone", plan.Actions[0].Kind)
	assert.Equal(t, filepath.Join(root, "acme/ok@main"), plan.Actions[0].Path)
	assert.Equal(t, "conflict", plan.Actions[1].Kind)
	assert.Equal(t, int64(2), plan.Actions[1].ID)
	assert.Equal(t, 1, plan.Summary.Conflict)
}

func TestReporter_ShortensOnlyInsideRoot(t *testing.T) {
	var out bytes.Buffer
	plan := Plan{Root: "/work", Actions: []Action{
		{Kind: KindConflict, ID: 1, Owner: "acme", Name: "x", Path: "/workspace/acme/x", Reason: "r"},
		{Kind: KindConflict, ID: 2, Owner: "acme", Name: "y", Path: "/work/acme/y", Reason: "r"},
		{Kind: KindConflict, ID: 3, Owner: "acme", Name: "z", Path: "/work", Reason: "r"},
	}}
	require.NoError(t, NewReporter(&out, &out, FormatHuman, 0, false).Render(plan, false))
	assert.Contains(t, out.String(), "at /workspace/acme/x\n")
	assert.Contains(t, out.String(), "at <root>/acme/y\n")
	assert.Contains(t, out.String(), "at <root>\n")
	assert.NotContains(t, out.String(), "<root>space")
}
