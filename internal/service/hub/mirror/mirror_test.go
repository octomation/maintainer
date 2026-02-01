package mirror_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/hub/mirror"
)

const fixtures = "../testdata/mirror"

func read(t *testing.T, rel string) *mirror.Document {
	t.Helper()
	doc, err := mirror.Read(filepath.Join(fixtures, rel))
	require.NoError(t, err)
	return doc
}

func at(s string) *time.Time { return mirror.ParseTime(s) }

func TestParseIssueGenerations(t *testing.T) {
	t.Run("code", func(t *testing.T) {
		doc := read(t, "octomation/maintainer/.github/issues/MAIN-1.md")
		assert.False(t, doc.CRLF)
		assert.Regexp(t, `^sha256-[0-9a-f]{64}$`, doc.Version)
		issue := mirror.ParseIssue(doc)
		assert.Equal(t, mirror.GenerationCode, issue.Generation)
		assert.Equal(t, "MAIN-1", issue.Code)
		require.NotNil(t, issue.Number)
		assert.Equal(t, 1, *issue.Number)
		assert.Equal(t, "I_main1", issue.NodeID)
		assert.Equal(t, "fetch: adopt existing checkouts", issue.Title)
		require.NotNil(t, issue.IssueType)
		assert.Equal(t, "Feature", *issue.IssueType)
		assert.Equal(t, []string{"type: feature", "scope: code", "impact: high", "effort: easy"}, issue.Labels)
		assert.Equal(t, []string{"kamilsk"}, issue.Assignees)
		assert.Equal(t, "MAIN-M1", issue.Milestone)
		assert.Equal(t, map[string]any{"Priority": "High"}, issue.IssueFields)
		assert.Equal(t, "open", issue.State)
		assert.Nil(t, issue.StateReason)
		assert.Equal(t, at("2026-01-10T10:00:00Z"), issue.CreatedAt)
		assert.Equal(t, at("2026-10-01T12:00:00Z"), issue.UpdatedAt)
		assert.Nil(t, issue.ClosedAt)
		require.Len(t, issue.Comments, 2)
		assert.Equal(t, "https://github.com/octomation/maintainer/issues/1#issuecomment-1", issue.Comments[0].URL)
		assert.Equal(t, *at("2026-02-01T08:30Z"), issue.Comments[0].At)
		assert.Equal(t, "First comment with **bold** text.", issue.Comments[0].Body)
		assert.Equal(t, *at("2026-03-01T09:45:10Z"), issue.Comments[1].At)
		assert.Equal(t, "Second comment:\n\n- one\n- two", issue.Comments[1].Body)
		assert.Contains(t, issue.Body, "<!-- not a comment block: it stays in the body -->")
		assert.Contains(t, issue.Body, "https://example.com/not-a-comment", "a block inside fenced code is body")
		assert.NotContains(t, issue.Body, "First comment")
	})

	t.Run("local-only without dates", func(t *testing.T) {
		issue := mirror.ParseIssue(read(t, "octomation/maintainer/.github/issues/MAIN-3.md"))
		assert.Equal(t, mirror.GenerationCode, issue.Generation)
		assert.Nil(t, issue.Number)
		assert.Nil(t, issue.IssueType)
		assert.Equal(t, "", issue.NodeID)
		assert.Equal(t, "MAIN-1", issue.Parent)
		assert.Nil(t, issue.CreatedAt)
		assert.Nil(t, issue.UpdatedAt)
		assert.Equal(t, []string{"type:bug", "effort: hard"}, issue.Labels)
		assert.Empty(t, issue.Comments)
	})

	t.Run("closed with reasons", func(t *testing.T) {
		issue := mirror.ParseIssue(read(t, "octomation/maintainer/.github/issues/MAIN-2.md"))
		assert.Equal(t, "closed", issue.State)
		require.NotNil(t, issue.StateReason)
		assert.Equal(t, "completed", *issue.StateReason)
		assert.Equal(t, []string{}, issue.Labels)
		assert.Equal(t, []string{}, issue.Assignees)
		assert.Equal(t, at("2026-09-15T00:00:00Z"), issue.ClosedAt)
		issue = mirror.ParseIssue(read(t, "octomation/maintainer/.github/issues/MAIN-7.md"))
		assert.Equal(t, "not_planned", *issue.StateReason)
	})

	t.Run("crlf", func(t *testing.T) {
		doc := read(t, "octomation/maintainer/.github/issues/MAIN-4.md")
		assert.True(t, doc.CRLF)
		issue := mirror.ParseIssue(doc)
		assert.Equal(t, "docs: the service guide", issue.Title)
		assert.Equal(t, []string{"scope: docs", "impact: medium", "effort: medium"}, issue.Labels)
		require.Len(t, issue.Comments, 1)
		assert.Equal(t, "A CRLF comment.", issue.Comments[0].Body)
		assert.Contains(t, issue.Body, "Written with CRLF line endings.")
	})

	t.Run("old", func(t *testing.T) {
		issue := mirror.ParseIssue(read(t, "octomation/tool/.github/issues/issue-1.md"))
		assert.Equal(t, mirror.GenerationOld, issue.Generation)
		assert.Equal(t, "issue-1", issue.Code, "the file stem stands in for the empty code")
		require.NotNil(t, issue.Number)
		assert.Equal(t, 1, *issue.Number)
		assert.Equal(t, "milestone-1", issue.Milestone)
		assert.Equal(t, "open", issue.State)
		assert.Nil(t, issue.StateReason)
		assert.Equal(t, []string{}, issue.Assignees)
		assert.Equal(t, at("2023-06-14T10:05:54Z"), issue.CreatedAt)
		issue = mirror.ParseIssue(read(t, "octomation/tool/.github/issues/issue-2.md"))
		assert.Equal(t, "closed", issue.State)
		assert.Equal(t, "duplicate", *issue.StateReason)
	})

	t.Run("dotfiles", func(t *testing.T) {
		issue := mirror.ParseIssue(read(t, "kamilsk/dotfiles/.github/issues/DOT-231.md"))
		assert.Equal(t, mirror.GenerationDotfiles, issue.Generation)
		assert.Equal(t, "DOT-231", issue.Code)
		require.NotNil(t, issue.Number)
		assert.Equal(t, 232, *issue.Number, "id is the GitHub number")
		assert.Equal(t, "I_kwDOCTsueM5A44Zp", issue.NodeID)
		assert.Equal(t, "open", issue.State)
		assert.Nil(t, issue.StateReason)
		assert.Equal(t, "DOT-M1", issue.Milestone)
		assert.Equal(t, []string{"type: improvement", "scope: code", "impact: medium", "effort: easy"}, issue.Labels)
		assert.Equal(t, at("2021-12-25T20:33:41Z"), issue.CreatedAt)
		assert.Equal(t, at("2026-09-17T11:02:03Z"), issue.UpdatedAt)
		issue = mirror.ParseIssue(read(t, "kamilsk/dotfiles/.github/issues/DOT-1.md"))
		assert.Equal(t, "closed", issue.State)
		assert.Equal(t, "completed", *issue.StateReason)
		assert.Equal(t, 1, *issue.Number)
		require.Len(t, issue.Comments, 1)
	})

	t.Run("malformed", func(t *testing.T) {
		_, err := mirror.Read(filepath.Join(fixtures, "octomation/maintainer/.github/issues/MAIN-8.md"))
		require.Error(t, err)
		_, err = mirror.Parse([]byte("# no frontmatter\n"))
		assert.ErrorIs(t, err, mirror.ErrNoFrontmatter)
		_, err = mirror.Parse([]byte("---\ntitle: x\n"))
		assert.ErrorIs(t, err, mirror.ErrNoFrontmatter)
		doc, err := mirror.Parse([]byte("---\ntitle: x\n---"))
		require.NoError(t, err, "the delimiter may end the file")
		assert.Equal(t, "", doc.Body)
		doc, err = mirror.Parse([]byte("---\n---\nbody\n"))
		require.NoError(t, err, "an empty frontmatter")
		assert.Equal(t, "body\n", doc.Body)
	})
}

func TestSplitComments(t *testing.T) {
	body, comments := mirror.SplitComments("text\n\n<!-- 2026-01-01T00:00Z https://x/1\none\n-->\n\n<!-- 2026-01-02T00:00Z https://x/2\ntwo\n-->\n")
	assert.Equal(t, "text\n", body)
	require.Len(t, comments, 2)
	assert.Equal(t, "one", comments[0].Body)
	assert.Equal(t, "two", comments[1].Body)

	body, comments = mirror.SplitComments("<!-- 2026-01-01T00:00Z https://x/1\none\n-->\n\nmore text\n")
	assert.Empty(t, comments, "a block before other content is not trailing")
	assert.Contains(t, body, "one")

	body, comments = mirror.SplitComments("text\n<!-- unknown https://x/1\nnot a comment\n-->\n")
	assert.Empty(t, comments, "the timestamp must parse")
	assert.Contains(t, body, "not a comment")

	body, comments = mirror.SplitComments("```\n<!-- 2026-01-01T00:00Z https://x/1\nfenced\n-->\n```\n")
	assert.Empty(t, comments)
	assert.Contains(t, body, "fenced")

	_, comments = mirror.SplitComments("text\n\n<!-- 2026-01-01T00:00Z https://x/1\nunclosed\n")
	assert.Empty(t, comments)

	// A line that only starts with the marker does not close the fence.
	body, comments = mirror.SplitComments("```\n```not-a-close\n<!-- 2026-01-01T00:00Z https://x/1\nstill fenced\n-->\n```\n")
	assert.Empty(t, comments, "```not-a-close keeps the fence open")
	assert.Contains(t, body, "still fenced")
	body, comments = mirror.SplitComments("```\n    ```\n<!-- 2026-01-01T00:00Z https://x/1\nstill fenced\n-->\n```\n")
	assert.Empty(t, comments, "an over-indented marker keeps the fence open")
	assert.Contains(t, body, "still fenced")
	_, comments = mirror.SplitComments("````\n```\nstill fenced\n````\n\n<!-- 2026-01-01T00:00Z https://x/1\none\n-->\n")
	require.Len(t, comments, 1, "a shorter marker does not close a longer fence; the matching one does")
	_, comments = mirror.SplitComments("```\n  ``` \n\n<!-- 2026-01-01T00:00Z https://x/1\none\n-->\n")
	assert.Len(t, comments, 1, "up to three spaces and trailing whitespace close it")
}

func TestParseBoardFences(t *testing.T) {
	doc, err := mirror.Parse([]byte("---\ncode: X-P1\n---\n\n## To do\n\n```\n```not-a-close\n- [ ] [[X-1]] — fenced\n```\n- [ ] [[X-2]] — listed\n\n## Done\n\n```\n    ```\n- [x] [[X-3]] — fenced\n```\n- [x] [[X-4]] — listed\n"))
	require.NoError(t, err)
	b := mirror.ParseBoard(doc)
	require.Len(t, b.Sections, 2)
	assert.Equal(t, []mirror.Item{{Link: "X-2", Text: "[[X-2]] — listed"}}, b.Sections[0].Items)
	assert.Equal(t, []mirror.Item{{Checked: true, Link: "X-4", Text: "[[X-4]] — listed"}}, b.Sections[1].Items)
}

func TestParseRoadmap(t *testing.T) {
	t.Run("milestones", func(t *testing.T) {
		doc := read(t, "octomation/maintainer/.github/roadmap/MAIN-M1.md")
		require.True(t, mirror.IsMilestone(doc))
		m := mirror.ParseMilestone(doc)
		assert.Equal(t, mirror.Milestone{Code: "MAIN-M1", Number: 1, NodeID: "MI_main1", Title: "Workspace, v0.2", State: "open", DueOn: at("2026-11-01T00:00:00Z"), URL: "https://github.com/octomation/maintainer/milestone/1"}, *m)
		m = mirror.ParseMilestone(read(t, "octomation/maintainer/.github/roadmap/MAIN-M2.md"))
		assert.Nil(t, m.DueOn)
		doc = read(t, "octomation/tool/.github/roadmap/milestone-1.md")
		require.True(t, mirror.IsMilestone(doc))
		m = mirror.ParseMilestone(doc)
		assert.Equal(t, "milestone-1", m.Code)
		assert.Equal(t, 1, m.Number)
		assert.Nil(t, m.DueOn)
		doc = read(t, "kamilsk/dotfiles/.github/roadmap/DOT-M1.md")
		require.True(t, mirror.IsMilestone(doc))
		m = mirror.ParseMilestone(doc)
		assert.Equal(t, mirror.Milestone{Code: "DOT-M1", Number: 1, NodeID: "MI_kwDOCTsueM4AfH6f", Title: "Git flow", State: "open", DueOn: at("2026-12-31T00:00:00Z"), URL: "https://github.com/kamilsk/dotfiles/milestone/1"}, *m)
	})

	t.Run("boards", func(t *testing.T) {
		doc := read(t, "octomation/maintainer/.github/roadmap/MAIN-P1.md")
		require.False(t, mirror.IsMilestone(doc))
		b := mirror.ParseBoard(doc)
		assert.Equal(t, "MAIN-P1", b.Code)
		assert.Equal(t, "PVT_org1", b.NativeID)
		assert.Equal(t, "Platform", b.Title)
		var names []string
		for _, s := range b.Sections {
			names = append(names, s.Name)
		}
		assert.Equal(t, []string{"Backlog", "To do", "In progress", "Done", "Won't do"}, names)
		assert.Equal(t, []mirror.Item{
			{Link: "MAIN-1", Text: "[[MAIN-1]] — fetch: adopt existing checkouts"},
			{Ref: "octomation/tool#3", Text: "octomation/tool#3 — a task whose mirror is not in the collection"},
		}, b.Sections[0].Items, "the draft is skipped")
		assert.Equal(t, mirror.Item{Checked: true, Ref: "octomation/tool#2", Text: "octomation/tool#2 — tool: duplicate of the first"}, b.Sections[4].Items[1])

		b = mirror.ParseBoard(read(t, "octomation/tool/.github/roadmap/project-octomation-1.md"))
		assert.Equal(t, "project-octomation-1", b.Code)
		assert.Equal(t, "PVT_org1", b.NativeID, "the same organization board")
		require.Len(t, b.Sections, 1)
		assert.Equal(t, "Items", b.Sections[0].Name, "the count is stripped")
		assert.Equal(t, mirror.Item{Ref: "octomation/tool#1", Text: "octomation/tool#1 — tool: first task ([[issue-1]])"}, b.Sections[0].Items[0])

		b = mirror.ParseBoard(read(t, "kamilsk/dotfiles/.github/roadmap/DOT-P1.md"))
		assert.Equal(t, "PVT_kwHOABHIaM4AB1OK", b.NativeID, "node_id in the dotfiles shape")
		assert.Equal(t, "Items", b.Sections[0].Name)
	})

	t.Run("sections", func(t *testing.T) {
		for name, want := range map[string]string{"To do": "todo", "Todo": "todo", "TODO": "todo", "In progress": "in_progress", "In Progress": "in_progress", "Backlog": "backlog", "Done": "done", "Won't do": "wont_do", "Won’t do": "wont_do", "Items": "", "Idle": "", "Cancelled": ""} {
			assert.Equal(t, want, mirror.NormalizeSection(name), name)
		}
	})
}

func TestGenerate(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, mirror.Generate(root, "acme", "gen", "GEN", 60))
	entries, err := os.ReadDir(filepath.Join(root, ".github", "issues"))
	require.NoError(t, err)
	assert.Len(t, entries, 60)
	doc, err := mirror.Read(filepath.Join(root, ".github/issues/GEN-7.md"))
	require.NoError(t, err)
	issue := mirror.ParseIssue(doc)
	assert.Equal(t, "GEN-7", issue.Code)
	assert.Len(t, issue.Comments, 1)
	doc, err = mirror.Read(filepath.Join(root, ".github/roadmap/GEN-P1.md"))
	require.NoError(t, err)
	assert.Len(t, mirror.ParseBoard(doc).Sections, 5)
}
