package hub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/hub/mirror"
)

var now = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)

var weights = config.Weights{Impact: 1, Effort: 1, Type: 1, Age: 1, Due: 1, AgeHorizonDays: 365, DueHorizonDays: 30}

// fixtureRoots copies the fixture mirror into a temporary directory so that
// writes never touch the repository, and returns its roots.
func fixtureRoots(t *testing.T) []Root {
	t.Helper()
	base := t.TempDir()
	var roots []Root
	for _, r := range []Root{
		{Owner: "octomation", Name: "maintainer", Path: "octomation/maintainer"},
		{Owner: "octomation", Name: "tool", Path: "octomation/tool", Visibility: "private"},
		{Owner: "kamilsk", Name: "dotfiles", Path: "kamilsk/dotfiles"},
		{Owner: "kamilsk", Name: "dotfiles", Path: "kamilsk/dotfiles-copy"},
	} {
		src := filepath.Join("testdata", "mirror", r.Path)
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

func newIndex(t *testing.T, roots []Root) *Index {
	t.Helper()
	ix := New(func(context.Context) ([]Root, error) { return roots, nil }, weights, func(r Root) float64 {
		if r.Name == "tool" {
			return 2
		}
		return 1
	})
	ix.now = func() time.Time { return now }
	require.NoError(t, ix.Scan(context.Background()))
	return ix
}

func TestBuildFixtureMirror(t *testing.T) {
	ix := newIndex(t, fixtureRoots(t))
	snap := ix.Snapshot()
	require.NotNil(t, snap)
	assert.Equal(t, uint64(1), snap.Generation)
	assert.Equal(t, now, snap.IndexedAt)

	task := func(key string) *Task {
		t.Helper()
		task := snap.Task(key)
		require.NotNil(t, task, key)
		return task
	}
	keys := make([]string, 0, len(snap.Tasks))
	for _, task := range snap.Tasks {
		keys = append(keys, task.Key)
	}
	assert.Equal(t, []string{
		"github:kamilsk/dotfiles#1", "github:kamilsk/dotfiles#232",
		"github:octomation/maintainer#1", "github:octomation/maintainer#2", "github:octomation/maintainer#4", "github:octomation/maintainer#5",
		"github:octomation/maintainer#6", "github:octomation/maintainer#7", "github:octomation/maintainer#9", "github:octomation/maintainer:MAIN-3",
		"github:octomation/tool#1", "github:octomation/tool#2",
	}, keys)

	t.Run("ids", func(t *testing.T) {
		for _, key := range []string{"github:octomation/maintainer#1", "github:octomation/maintainer:MAIN-3"} {
			id := task(key).ID
			assert.NotContains(t, id, "=")
			assert.NotContains(t, id, "/")
			assert.NotContains(t, id, "+")
			back, err := DecodeID(id)
			require.NoError(t, err)
			assert.Equal(t, key, back)
		}
		_, err := DecodeID("not base64!")
		assert.Error(t, err)
		_, err = DecodeID(EncodeID("other:thing"))
		assert.Error(t, err)
	})

	t.Run("columns and flags", func(t *testing.T) {
		for key, want := range map[string]struct {
			column string
			flags  Flags
		}{
			"github:octomation/maintainer#1":      {ColumnBacklog, Flags{}},
			"github:octomation/maintainer#2":      {ColumnDone, Flags{}},
			"github:octomation/maintainer:MAIN-3": {ColumnBacklog, Flags{}},
			"github:octomation/maintainer#4":      {ColumnInProgress, Flags{}},
			"github:octomation/maintainer#5":      {ColumnBacklog, Flags{Contradiction: true}},
			"github:octomation/maintainer#6":      {ColumnBacklog, Flags{Ambiguous: true}},
			"github:octomation/maintainer#7":      {ColumnWontDo, Flags{}},
			"github:octomation/maintainer#9":      {ColumnTodo, Flags{}},
			"github:octomation/tool#1":            {ColumnTodo, Flags{}}, // To do on the board; Items on its partial copy casts no vote
			"github:octomation/tool#2":            {ColumnWontDo, Flags{}},
			"github:kamilsk/dotfiles#1":           {ColumnDone, Flags{DuplicateMirror: true}},
			"github:kamilsk/dotfiles#232":         {ColumnBacklog, Flags{}},
		} {
			assert.Equal(t, want.column, task(key).Column, key)
			assert.Equal(t, want.flags, task(key).Flags, key)
		}
	})

	t.Run("memberships merge across partial copies", func(t *testing.T) {
		assert.Equal(t, []Membership{{Board: "PVT_org1", Section: "To do"}, {Board: "PVT_org1", Section: "Items"}}, task("github:octomation/tool#1").Boards)
		assert.Equal(t, []Membership{{Board: "PVT_org1", Section: "To do"}, {Board: "PVT_repo2", Section: "In progress"}}, task("github:octomation/maintainer#6").Boards)
		assert.Equal(t, []Membership{{Board: "PVT_org1", Section: "In progress"}, {Board: "PVT_repo2", Section: "In progress"}}, task("github:octomation/maintainer#4").Boards, "agreeing boards")
		assert.Equal(t, []Membership{{Board: "PVT_kwHOABHIaM4AB1OK", Section: "Items"}}, task("github:kamilsk/dotfiles#232").Boards, "a reference resolves across the dotfiles code and number")
		var org *Board
		for i := range snap.Catalog.Boards {
			if snap.Catalog.Boards[i].Key == "PVT_org1" {
				org = &snap.Catalog.Boards[i]
			}
		}
		require.NotNil(t, org)
		assert.Equal(t, []string{"octomation/maintainer", "octomation/tool"}, org.Repositories)
		assert.Equal(t, []string{"Backlog", "To do", "In progress", "Done", "Won't do", "Items"}, org.Sections)
		assert.Equal(t, "Platform", org.Title)
		assert.Len(t, snap.Catalog.Boards, 3)
	})

	t.Run("types", func(t *testing.T) {
		want := map[string]string{
			"github:octomation/maintainer#1":      "feature",
			"github:octomation/maintainer#2":      "bug",
			"github:octomation/maintainer:MAIN-3": "bug", // type:bug label without issueType
			"github:octomation/maintainer#6":      "feature",
			"github:octomation/maintainer#7":      "chore",
			"github:octomation/tool#1":            "feature",
		}
		for key, typ := range want {
			require.NotNil(t, task(key).Type, key)
			assert.Equal(t, typ, *task(key).Type, key)
		}
		assert.Nil(t, task("github:octomation/tool#2").Type)
		assert.Nil(t, task("github:kamilsk/dotfiles#1").Type)
		assert.Equal(t, "improvement", *task("github:kamilsk/dotfiles#232").Type)
	})

	t.Run("milestones", func(t *testing.T) {
		assert.Equal(t, &MilestoneRef{Key: "github:octomation/maintainer/milestones/1", Title: "Workspace, v0.2"}, task("github:octomation/maintainer#1").Milestone)
		assert.Equal(t, &MilestoneRef{Key: "github:octomation/tool/milestones/1", Title: "tool 1.0"}, task("github:octomation/tool#1").Milestone)
		assert.Equal(t, &MilestoneRef{Key: "github:kamilsk/dotfiles/milestones/1", Title: "Git flow"}, task("github:kamilsk/dotfiles#232").Milestone)
		assert.Nil(t, task("github:octomation/maintainer#6").Milestone)
		require.Len(t, snap.Catalog.Milestones, 4)
		assert.Equal(t, "github:kamilsk/dotfiles/milestones/1", snap.Catalog.Milestones[0].Key)
		assert.Equal(t, mirror.ParseTime("2026-12-31T00:00:00Z"), snap.Catalog.Milestones[0].DueOn)
		assert.Nil(t, snap.Catalog.Milestones[2].DueOn, "MAIN-M2 has no due date")
	})

	t.Run("score", func(t *testing.T) {
		s := task("github:octomation/maintainer#1").Score
		assert.Equal(t, 1.0, s.Impact)
		assert.Equal(t, 1.0, s.Ease)
		assert.Equal(t, 0.5, s.Type)
		assert.InDelta(t, 283.0/365, s.Age, 0.01, "283 days since 2026-01-10")
		assert.InDelta(t, 1-11.5/30, s.Due, 0.001, "11.5 days before the due date over a 30-day horizon")
		assert.InDelta(t, 1+1+0.5+283.0/365+(1-11.5/30), s.Total, 0.02)
		assert.Equal(t, 1.0, s.RepoWeight)
		s = task("github:octomation/tool#1").Score
		assert.Equal(t, 2.0, s.RepoWeight)
		assert.Equal(t, 0.25, s.Impact)
		assert.Equal(t, 1.0, s.Age, "older than the horizon")
		assert.Equal(t, 0.0, s.Due, "no due date")
		assert.InDelta(t, 2*(0.25+1+0.5+1), s.Total, 0.001)
		s = task("github:octomation/maintainer:MAIN-3").Score
		assert.Equal(t, 0.0, s.Age, "no created date")
		assert.Equal(t, 0.25, s.Ease)
		assert.Equal(t, 1.0, s.Type)
		assert.Equal(t, 0.0, s.Impact)
		s = task("github:kamilsk/dotfiles#232").Score
		assert.Equal(t, 0.0, s.Due, "the due date is more than 30 days away")
		s = task("github:octomation/maintainer#2").Score
		assert.Equal(t, 0.0, s.Due, "MAIN-M2 has no due date")
	})

	t.Run("warnings and stale", func(t *testing.T) {
		require.Len(t, snap.Warnings, 2, snap.Warnings)
		assert.Contains(t, snap.Warnings[0], "DOT-1.md: duplicate mirror of github:kamilsk/dotfiles#1")
		assert.Contains(t, snap.Warnings[1], "MAIN-8.md")
		assert.False(t, task("github:kamilsk/dotfiles#1").Flags.Stale)
	})

	t.Run("catalog", func(t *testing.T) {
		assert.Len(t, snap.Catalog.Repositories, 4)
		assert.Equal(t, Columns, snap.Catalog.Columns)
		assert.Contains(t, snap.Catalog.Labels, "impact: high")
		assert.Contains(t, snap.Catalog.Labels, "type:bug")
		assert.Equal(t, 2.0, snap.Catalog.Repositories[3].Weight)
		assert.Equal(t, 2, snap.Catalog.Repositories[3].Tasks)
		open := 0
		for _, task := range snap.Tasks {
			if task.Repository == "octomation/tool" && task.State == "open" {
				open++
			}
		}
		assert.Equal(t, open, snap.Catalog.Repositories[3].Open, "the open tasks of octomation/tool")
		assert.True(t, snap.Catalog.Repositories[3].Private, "octomation/tool is private by the fetch state")
		assert.False(t, snap.Catalog.Repositories[2].Private, "an unknown visibility shows no lock")
	})

	t.Run("comments and body", func(t *testing.T) {
		one := task("github:octomation/maintainer#1")
		assert.Len(t, one.comments, 2)
		assert.Contains(t, one.Body(), "<!-- not a comment block")
		assert.True(t, strings.HasSuffix(one.Path(), "MAIN-1.md"))
		assert.Equal(t, "octomation/maintainer", one.Root().Repository())
	})
}

func TestDeriveColumn(t *testing.T) {
	reason := func(s string) *string { return &s }
	open := func(sections ...string) *Task {
		task := &Task{State: "open"}
		for i, section := range sections {
			task.Boards = append(task.Boards, Membership{Board: fmt.Sprintf("B%d", i), Section: section})
		}
		return task
	}
	for name, tc := range map[string]struct {
		task                     *Task
		column                   string
		ambiguous, contradiction bool
	}{
		"no boards":                       {open(), ColumnBacklog, false, false},
		"items only":                      {open("Items"), ColumnBacklog, false, false},
		"items with a count":              {open("Items (16)"), ColumnBacklog, false, false},
		"to do and items":                 {open("To do", "ITEMS"), ColumnTodo, false, false},
		"in progress and items":           {open("In progress", "Items (3)"), ColumnInProgress, false, false},
		"to do and idle":                  {open("To do", "Idle"), ColumnBacklog, true, false},
		"idle only":                       {open("Idle"), ColumnBacklog, false, false},
		"idle and cancelled":              {open("Idle", "Cancelled"), ColumnBacklog, false, false},
		"two boards agree":                {open("Todo", "To do"), ColumnTodo, false, false},
		"two boards disagree":             {open("To do", "In progress"), ColumnBacklog, true, false},
		"done for an open task":           {open("Done", "To do"), ColumnTodo, false, true},
		"won't do and items":              {open("Won't do", "Items"), ColumnBacklog, false, true},
		"closed completed under to do":    {&Task{State: "closed", StateReason: reason("completed"), Boards: []Membership{{Board: "B", Section: "To do"}}}, ColumnDone, false, false},
		"closed not planned":              {&Task{State: "closed", StateReason: reason("not_planned")}, ColumnWontDo, false, false},
		"closed duplicate":                {&Task{State: "closed", StateReason: reason("duplicate")}, ColumnWontDo, false, false},
		"closed without a reason is done": {&Task{State: "closed"}, ColumnDone, false, false},
	} {
		column, ambiguous, contradiction := deriveColumn(tc.task)
		assert.Equal(t, tc.column, column, name)
		assert.Equal(t, tc.ambiguous, ambiguous, "%s: ambiguous", name)
		assert.Equal(t, tc.contradiction, contradiction, "%s: contradiction", name)
	}
}

func TestQuery(t *testing.T) {
	snap := newIndex(t, fixtureRoots(t)).Snapshot()
	query := func(q Query) Page {
		t.Helper()
		require.NoError(t, q.Validate())
		return snap.Query(q)
	}
	keys := func(p Page) []string {
		out := make([]string, 0, len(p.Tasks))
		for _, task := range p.Tasks {
			out = append(out, task.Key)
		}
		return out
	}

	t.Run("open by default, score descending", func(t *testing.T) {
		p := query(Query{})
		assert.Equal(t, 8, p.Total)
		assert.Equal(t, 100, p.Limit)
		assert.Equal(t, keys(p)[0], "github:octomation/tool#1", "the weighted repository wins")
		assert.Equal(t, []Facet{{Value: "open", Count: 8}}, p.Facets.States)
		assert.Equal(t, []Facet{{Value: ColumnBacklog, Count: 5}, {Value: ColumnTodo, Count: 2}, {Value: ColumnInProgress, Count: 1}}, p.Facets.Columns)
	})

	t.Run("states", func(t *testing.T) {
		assert.Equal(t, 4, query(Query{State: "closed"}).Total)
		assert.Equal(t, 12, query(Query{State: "all"}).Total)
		assert.Equal(t, []Facet{{Value: "open", Count: 8}, {Value: "closed", Count: 4}}, query(Query{State: "all"}).Facets.States)
	})

	t.Run("dimensions are AND, values OR, labels all", func(t *testing.T) {
		assert.Equal(t, 7, query(Query{Owners: []string{"OctoMation"}}).Total)
		assert.Equal(t, 8, query(Query{Owners: []string{"octomation", "kamilsk"}}).Total)
		assert.Equal(t, 1, query(Query{Repos: []string{"octomation/tool"}}).Total)
		assert.Equal(t, 2, query(Query{State: "all", Repos: []string{"octomation/tool"}, Owners: []string{"octomation"}}).Total)
		assert.Equal(t, 2, query(Query{Labels: []string{"impact: high", "scope: code"}}).Total)
		assert.Equal(t, 1, query(Query{Labels: []string{"impact: high", "effort: easy"}}).Total)
		assert.Equal(t, 2, query(Query{Labels: []string{"impact: high"}}).Total)
		assert.Equal(t, 4, query(Query{Types: []string{"feature"}}).Total)
		assert.Equal(t, 3, query(Query{Assignees: []string{"kamilsk"}}).Total)
		assert.Equal(t, 1, query(Query{Assignees: []string{"octocat"}}).Total)
		assert.Equal(t, 2, query(Query{Milestones: []string{"github:octomation/maintainer/milestones/1"}}).Total)
		assert.Equal(t, 2, query(Query{State: "all", Columns: []string{ColumnDone}}).Total)
		assert.Equal(t, 4, query(Query{State: "all", Columns: []string{ColumnDone, ColumnWontDo}}).Total)
	})

	t.Run("facets ignore column, limit and offset", func(t *testing.T) {
		p := query(Query{State: "all", Columns: []string{ColumnDone}, Limit: 1, Offset: 1})
		assert.Equal(t, 2, p.Total)
		assert.Len(t, p.Tasks, 1)
		assert.Equal(t, 1, p.Offset)
		assert.Equal(t, 1, p.Limit)
		assert.Equal(t, []Facet{{Value: ColumnBacklog, Count: 5}, {Value: ColumnDone, Count: 2}, {Value: ColumnTodo, Count: 2}, {Value: ColumnWontDo, Count: 2}, {Value: ColumnInProgress, Count: 1}}, p.Facets.Columns)
		assert.Equal(t, []Facet{{Value: "octomation/maintainer", Count: 8}, {Value: "kamilsk/dotfiles", Count: 2}, {Value: "octomation/tool", Count: 2}}, p.Facets.Repositories)
		assert.Equal(t, []Facet{{Value: "kamilsk", Count: 3}, {Value: "octocat", Count: 1}}, p.Facets.Assignees)
	})

	t.Run("q", func(t *testing.T) {
		assert.Equal(t, []string{"github:octomation/maintainer#4"}, keys(query(Query{Q: "GUIDE"})))
		assert.Equal(t, 2, query(Query{Q: "impact: high"}).Total, "labels match")
		assert.Equal(t, 1, query(Query{Q: "dotfiles"}).Total, "the repository matches")
		assert.Equal(t, 1, query(Query{Q: "main-3"}).Total, "the code matches")
	})

	t.Run("orders", func(t *testing.T) {
		assert.Equal(t, []string{
			"github:octomation/maintainer#1", "github:octomation/maintainer#9", // high, updated desc
			"github:octomation/maintainer#4", "github:kamilsk/dotfiles#232", // medium
			"github:octomation/maintainer#5", "github:octomation/tool#1", // low
			"github:octomation/maintainer#6", "github:octomation/maintainer:MAIN-3", // unset: by updated, missing last
		}, keys(query(Query{Sort: SortImportance})))
		updated := keys(query(Query{Sort: SortUpdated}))
		assert.Equal(t, "github:octomation/maintainer#4", updated[0])
		assert.Equal(t, "github:octomation/maintainer:MAIN-3", updated[len(updated)-1], "a missing date is last")
		created := keys(query(Query{Sort: SortCreated}))
		assert.Equal(t, "github:octomation/maintainer#6", created[0])
		assert.Equal(t, "github:octomation/maintainer:MAIN-3", created[len(created)-1])
	})

	t.Run("validation", func(t *testing.T) {
		assert.Error(t, (&Query{State: "done"}).Validate())
		assert.Error(t, (&Query{Sort: "title"}).Validate())
		assert.Error(t, (&Query{Columns: []string{"doing"}}).Validate())
		assert.Error(t, (&Query{Limit: -1}).Validate())
		q := Query{Limit: 9000}
		require.NoError(t, q.Validate())
		assert.Equal(t, MaxLimit, q.Limit)
		var empty *Snapshot
		assert.Equal(t, 0, empty.Query(Query{}).Total)
		assert.Nil(t, empty.Task("x"))
	})
}

// multiline rewrites the labels of a copied MAIN-9 as a flow list over
// several lines, which the writer refuses.
func multiline(t *testing.T, roots []Root) {
	t.Helper()
	path := filepath.Join(roots[0].Path, ".github", "issues", "MAIN-9.md")
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	next := strings.Replace(string(raw), `labels: ["impact: high", "scope: code"]`, "labels: [\"impact: high\",\n  \"scope: code\"]", 1)
	require.NotEqual(t, string(raw), next)
	require.NoError(t, os.WriteFile(path, []byte(next), 0o644))
}

func TestIndexWrites(t *testing.T) {
	roots := fixtureRoots(t)
	multiline(t, roots)
	ix := newIndex(t, roots)
	key := "github:octomation/maintainer#1"
	before := ix.Snapshot().Task(key)

	t.Run("patch publishes a new generation", func(t *testing.T) {
		task, snap, err := ix.Patch(key, before.Version, mirror.Triage{Impact: mirror.Change{Set: true, Value: "low"}})
		require.NoError(t, err)
		assert.Equal(t, uint64(2), snap.Generation)
		assert.Same(t, snap, ix.Snapshot())
		assert.NotEqual(t, before.Version, task.Version)
		assert.Equal(t, "low", Impact(task))
		assert.Equal(t, 0.25, task.Score.Impact)
		assert.Equal(t, before.Boards, task.Boards, "the rebuilt task keeps its memberships")
		_, _, err = ix.Patch(key, before.Version, mirror.Triage{Impact: mirror.Change{Set: true, Value: "high"}})
		assert.ErrorIs(t, err, mirror.ErrVersionMismatch, "the old version is stale")
	})

	t.Run("errors", func(t *testing.T) {
		_, _, err := ix.Patch("github:nobody/nothing#1", "v", mirror.Triage{})
		assert.ErrorIs(t, err, ErrNotFound)
		dup := ix.Snapshot().Task("github:kamilsk/dotfiles#1")
		_, _, err = ix.Patch(dup.Key, dup.Version, mirror.Triage{Effort: mirror.Change{Set: true, Value: "easy"}})
		assert.ErrorIs(t, err, ErrDuplicateMirror)
		flow := ix.Snapshot().Task("github:octomation/maintainer#9")
		_, _, err = ix.Patch(flow.Key, flow.Version, mirror.Triage{Effort: mirror.Change{Set: true, Value: "easy"}})
		assert.ErrorIs(t, err, mirror.ErrNotEditable)
		empty := New(func(context.Context) ([]Root, error) { return nil, nil }, weights, nil)
		_, _, err = empty.Patch(key, "v", mirror.Triage{})
		assert.ErrorIs(t, err, ErrNotIndexed)
	})

	t.Run("a scan racing a write cannot publish the older parse", func(t *testing.T) {
		current := ix.Snapshot().Task(key)
		// The scan reads the files before the write, publishes after it.
		started := make(chan struct{})
		release := make(chan struct{})
		slow := New(func(ctx context.Context) ([]Root, error) {
			close(started)
			<-release
			return roots, nil
		}, weights, nil)
		slow.now = ix.now
		slow.current.Store(ix.Snapshot())
		// The write goes through the index under test, which shares the
		// snapshot: simulate it on the slow index by copying its state.
		done := make(chan error, 1)
		go func() { done <- slow.Scan(context.Background()) }()
		<-started
		// Meanwhile a write lands on disk and in the slow index's snapshot.
		written, snap, err := slow.Patch(key, current.Version, mirror.Triage{Effort: mirror.Change{Set: true, Value: "hard"}})
		require.NoError(t, err)
		assert.Equal(t, "hard", Effort(written))
		gen := snap.Generation
		close(release)
		require.NoError(t, <-done)
		after := slow.Snapshot()
		assert.Equal(t, gen+1, after.Generation)
		assert.Equal(t, "hard", Effort(after.Task(key)), "the write survives the scan")
		assert.Equal(t, written.Version, after.Task(key).Version)
	})

	t.Run("an ancestor of the root replaced by a symlink after discovery", func(t *testing.T) {
		base := t.TempDir()
		src := filepath.Join(base, "work", "octomation", "maintainer")
		require.NoError(t, os.MkdirAll(filepath.Dir(src), 0o755))
		require.NoError(t, os.CopyFS(src, os.DirFS(roots[0].Path)))
		own := newIndex(t, []Root{{Owner: "octomation", Name: "maintainer", Path: src}})
		task := own.Snapshot().Task(key)
		require.NotNil(t, task)
		assert.NotEmpty(t, task.Root().Real)
		before, err := os.ReadFile(task.Path())
		require.NoError(t, err)
		// ~/Development itself replaced by a symlink to a copy.
		require.NoError(t, os.Rename(filepath.Join(base, "work"), filepath.Join(base, "work-real")))
		require.NoError(t, os.CopyFS(filepath.Join(base, "copy"), os.DirFS(filepath.Join(base, "work-real"))))
		require.NoError(t, os.Symlink(filepath.Join(base, "copy"), filepath.Join(base, "work")))
		_, _, err = own.Patch(key, task.Version, mirror.Triage{Impact: mirror.Change{Set: true, Value: "high"}})
		assert.ErrorIs(t, err, mirror.ErrNotEditable)
		for _, tree := range []string{"work-real", "copy"} {
			after, err := os.ReadFile(filepath.Join(base, tree, "octomation", "maintainer", ".github", "issues", "MAIN-1.md"))
			require.NoError(t, err)
			assert.Equal(t, before, after, tree)
		}
		// A stable symlinked ancestor, as /var to /private/var, is fine: it
		// was already resolved at discovery.
		stable := newIndex(t, []Root{{Owner: "octomation", Name: "maintainer", Path: filepath.Join(base, "work", "octomation", "maintainer")}})
		copyTask := stable.Snapshot().Task(key)
		_, _, err = stable.Patch(key, copyTask.Version, mirror.Triage{Impact: mirror.Change{Set: true, Value: "high"}})
		require.NoError(t, err)
	})

	t.Run("a mismatch on disk publishes the current file", func(t *testing.T) {
		task := ix.Snapshot().Task(key)
		raw, err := os.ReadFile(task.Path())
		require.NoError(t, err)
		edited := strings.Replace(string(raw), "fetch: adopt existing checkouts", "fetch: adopt existing checkouts (edited)", 1)
		require.NoError(t, os.WriteFile(task.Path(), []byte(edited), 0o644))
		before := ix.Snapshot().Generation
		_, _, err = ix.Patch(key, task.Version, mirror.Triage{Impact: mirror.Change{Set: true, Value: "high"}})
		assert.ErrorIs(t, err, mirror.ErrVersionMismatch)
		current := ix.Snapshot().Task(key)
		assert.Equal(t, before+1, ix.Snapshot().Generation, "the current parse is published")
		assert.NotEqual(t, task.Version, current.Version)
		assert.Equal(t, mirror.Hash([]byte(edited)), current.Version)
		assert.Contains(t, current.Title, "(edited)")
		written, _, err := ix.Patch(key, current.Version, mirror.Triage{Impact: mirror.Change{Set: true, Value: "high"}})
		require.NoError(t, err, "the reload's version writes")
		assert.Equal(t, "high", Impact(written))
	})

	t.Run("a malformed file keeps its entry, flagged stale", func(t *testing.T) {
		task := ix.Snapshot().Task(key)
		raw, err := os.ReadFile(task.Path())
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(task.Path(), []byte("---\nbroken: [\n"), 0o644))
		require.NoError(t, ix.Scan(context.Background()))
		stale := ix.Snapshot().Task(key)
		require.NotNil(t, stale)
		assert.True(t, stale.Flags.Stale)
		assert.Equal(t, task.Title, stale.Title)
		var found bool
		for _, w := range ix.Snapshot().Warnings {
			found = found || strings.Contains(w, "MAIN-1.md") && strings.Contains(w, "previous entry is kept")
		}
		assert.True(t, found, ix.Snapshot().Warnings)
		require.NoError(t, os.WriteFile(task.Path(), raw, 0o644))
		require.NoError(t, ix.Scan(context.Background()))
		assert.False(t, ix.Snapshot().Task(key).Flags.Stale)
	})

	t.Run("refresh coalesces", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var errs []error
		fresh := New(func(context.Context) ([]Root, error) { return roots, nil }, weights, nil)
		fresh.Start(ctx, func(err error) { errs = append(errs, err) })
		require.Eventually(t, func() bool { return fresh.Snapshot() != nil }, 5*time.Second, 10*time.Millisecond)
		first := fresh.Snapshot().Generation
		for i := 0; i < 5; i++ {
			fresh.Refresh()
		}
		require.Eventually(t, func() bool { return fresh.Snapshot().Generation > first && !fresh.Refreshing() }, 5*time.Second, 10*time.Millisecond)
		assert.LessOrEqual(t, fresh.Snapshot().Generation, first+2)
		assert.Empty(t, errs)
	})
}

func TestRenderer(t *testing.T) {
	snap := newIndex(t, fixtureRoots(t)).Snapshot()
	r := NewRenderer(2)
	hostile := snap.Task("github:octomation/maintainer#7")
	out := r.Render(hostile)
	assert.NotContains(t, out.BodyHTML, "<script")
	assert.NotContains(t, out.BodyHTML, "javascript:")
	assert.Contains(t, out.BodyHTML, "<!-- raw HTML omitted -->")
	assert.Contains(t, out.BodyHTML, `href="mailto:a@example.com"`)
	assert.Contains(t, out.BodyHTML, `href="https://example.com/"`)
	assert.Contains(t, out.BodyHTML, `<a href="">click</a>`)
	assert.Contains(t, out.BodyHTML, `src="https://example.com/image.png"`, "remote images are left to the CSP")
	assert.Contains(t, out.BodyHTML, "title: not a delimiter", "the fenced delimiter is code")
	require.Len(t, out.Comments, 1)
	assert.NotContains(t, out.Comments[0].HTML, "<b>")
	assert.NotContains(t, out.Comments[0].HTML, "javascript:")
	assert.Same(t, out, r.Render(hostile), "cached by version")

	one := snap.Task("github:octomation/maintainer#1")
	first := r.Render(one)
	assert.Contains(t, first.BodyHTML, `<a href="https://example.com/docs">link</a>`)
	assert.Contains(t, one.Body(), "<!-- not a comment block: it stays in the body -->")
	assert.Contains(t, first.BodyHTML, "<!-- raw HTML omitted -->", "an HTML comment in the body is raw HTML")
	assert.Contains(t, first.BodyHTML, "https://example.com/not-a-comment", "the block inside fenced code is code")
	assert.Len(t, first.Comments, 2)
	assert.Contains(t, first.Comments[0].HTML, "<strong>bold</strong>")
	r.Render(snap.Task("github:octomation/maintainer#2"))
	assert.NotSame(t, out, r.Render(hostile), "evicted after the limit")

	for url, ok := range map[string]bool{"https://x": true, "http://x": true, "mailto:x@y": true, "javascript:alert(1)": false, "data:text/html,x": false, "vbscript:x": false, "/relative/path": true, "../up.md": true, "#anchor": true, "page?x=a:b": true, "JAVASCRIPT:x": false} {
		assert.Equal(t, ok, AllowedURL(url), url)
	}
}

func BenchmarkQueryGenerated(b *testing.B) {
	root := b.TempDir()
	require.NoError(b, mirror.Generate(root, "acme", "big", "BIG", 3000))
	ix := New(func(context.Context) ([]Root, error) { return []Root{{Path: root, Owner: "acme", Name: "big"}}, nil }, weights, nil)
	require.NoError(b, ix.Scan(context.Background()))
	snap := ix.Snapshot()
	require.Len(b, snap.Tasks, 3000)
	q := Query{Repos: []string{"acme/big"}, Labels: []string{"scope: code"}, Q: "fetch", State: "all"}
	require.NoError(b, q.Validate())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		page := snap.Query(q)
		if page.Total == 0 {
			b.Fatal("no tasks")
		}
	}
}

func TestGeneratedMirrorAnswersFast(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	root := t.TempDir()
	require.NoError(t, mirror.Generate(root, "acme", "big", "BIG", 3000))
	ix := New(func(context.Context) ([]Root, error) { return []Root{{Path: root, Owner: "acme", Name: "big"}}, nil }, weights, nil)
	require.NoError(t, ix.Scan(context.Background()))
	snap := ix.Snapshot()
	require.Len(t, snap.Tasks, 3000)
	assert.Empty(t, snap.Warnings)
	q := Query{Labels: []string{"scope: code"}, Q: "fetch", State: "all", Sort: SortImportance}
	require.NoError(t, q.Validate())
	snap.Query(q) // warm
	start := time.Now()
	page := snap.Query(q)
	elapsed := time.Since(start)
	assert.Greater(t, page.Total, 0)
	assert.Less(t, elapsed, 100*time.Millisecond, "a filtered query over 3000 tasks")
}
