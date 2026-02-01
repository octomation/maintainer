package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Generate writes a synthetic mirror of n issues of the code generation under
// root/.github for owner/name: a benchmark and a development fixture. Labels,
// states, types, dates, milestones and board memberships vary
// deterministically with the issue number.
func Generate(root, owner, name, code string, n int) error {
	issues := filepath.Join(root, ".github", "issues")
	roadmap := filepath.Join(root, ".github", "roadmap")
	for _, dir := range []string{issues, roadmap} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	impacts := []string{"high", "medium", "low", ""}
	efforts := []string{"easy", "medium", "hard", ""}
	types := []string{"Feature", "Bug", "Task", "null"}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	sections := map[string][]string{"Backlog": nil, "To do": nil, "In progress": nil, "Done": nil, "Won't do": nil}
	order := []string{"Backlog", "To do", "In progress", "Done", "Won't do"}
	for i := 1; i <= n; i++ {
		var labels []string
		if v := impacts[i%4]; v != "" {
			labels = append(labels, fmt.Sprintf("%q", "impact: "+v))
		}
		if v := efforts[(i/4)%4]; v != "" {
			labels = append(labels, fmt.Sprintf("%q", "effort: "+v))
		}
		labels = append(labels, fmt.Sprintf("%q", "scope: "+[]string{"code", "docs", "test"}[i%3]))
		state, reason, closed := "OPEN", "null", "null"
		created := base.Add(time.Duration(i) * 7 * time.Hour)
		updated := created.Add(time.Duration(i%30) * 24 * time.Hour)
		if i%5 == 0 {
			state, reason, closed = "CLOSED", "COMPLETED", updated.Format(time.RFC3339)
			if i%10 == 0 {
				reason = "NOT_PLANNED"
			}
		}
		milestone := "null"
		if i%3 != 0 {
			milestone = fmt.Sprintf("%q", fmt.Sprintf("[[%s-M%d]]", code, i%3))
		}
		number := fmt.Sprint(i)
		if i%50 == 0 {
			number = "null"
		}
		body := fmt.Sprintf("# task %d\n\nA generated task about **%s** number %d.\n\n- item one\n- item two\n", i, []string{"fetch", "status", "service"}[i%3], i)
		comments := ""
		if i%7 == 0 {
			comments = fmt.Sprintf("\n<!-- %s https://github.com/%s/%s/issues/%d#issuecomment-%d\nA comment on task %d.\n-->\n", updated.Format("2006-01-02T15:04Z"), owner, name, i, i, i)
		}
		content := fmt.Sprintf(`---
code: %s-%d
id: I_gen%d
databaseId: %d
number: %s
url: https://github.com/%s/%s/issues/%d
title: "task %d: %s"
issueType: %s
labels:
  - %s
assignees:
  - kamilsk
milestone: %s
parent: null
issueFields: {}
state: %s
stateReason: %s
duplicateOf: null
createdAt: %s
updatedAt: %s
lastEditedAt: null
closedAt: %s
---

%s%s`, code, i, i, i, number, owner, name, i, i, []string{"fetch", "status", "service"}[i%3], types[i%4], strings.Join(labels, "\n  - "), milestone, state, reason,
			created.Format(time.RFC3339), updated.Format(time.RFC3339), closed, body, comments)
		if err := os.WriteFile(filepath.Join(issues, fmt.Sprintf("%s-%d.md", code, i)), []byte(content), 0o644); err != nil {
			return err
		}
		if state == "OPEN" {
			section := order[i%3]
			sections[section] = append(sections[section], fmt.Sprintf("- [ ] [[%s-%d]] — task %d", code, i, i))
		} else {
			section := "Done"
			if reason == "NOT_PLANNED" {
				section = "Won't do"
			}
			sections[section] = append(sections[section], fmt.Sprintf("- [x] [[%s-%d]] — task %d", code, i, i))
		}
	}
	for m := 1; m <= 2; m++ {
		due := "null"
		if m == 1 {
			due = base.AddDate(1, 0, 0).Format(time.RFC3339)
		}
		content := fmt.Sprintf("---\ncode: %s-M%d\nid: MI_gen%d\ndatabaseId: %d\nnumber: %d\nurl: https://github.com/%s/%s/milestone/%d\ntitle: \"Milestone %d\"\nstate: OPEN\nopenIssueCount: 0\nclosedIssueCount: 0\ndueOn: %s\ncreatedAt: %s\nupdatedAt: %s\nclosedAt: null\n---\n\n# Milestone %d\n", code, m, m, m, m, owner, name, m, m, due, base.Format(time.RFC3339), base.Format(time.RFC3339), m)
		if err := os.WriteFile(filepath.Join(roadmap, fmt.Sprintf("%s-M%d.md", code, m)), []byte(content), 0o644); err != nil {
			return err
		}
	}
	var board strings.Builder
	fmt.Fprintf(&board, "---\ncode: %s-P1\nid: PVT_gen1\nnumber: 1\nurl: https://github.com/orgs/%s/projects/1\ntitle: \"Generated board\"\nclosed: false\n---\n\n# Generated board\n", code, owner)
	for _, section := range order {
		fmt.Fprintf(&board, "\n## %s\n\n%s\n", section, strings.Join(sections[section], "\n"))
	}
	return os.WriteFile(filepath.Join(roadmap, code+"-P1.md"), []byte(board.String()), 0o644)
}
