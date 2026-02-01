package hub

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/hub/mirror"
)

type kind int

const (
	kindBroken kind = iota
	kindIssue
	kindMilestone
	kindBoard
)

// entry is one parsed mirror file. A snapshot is built from entries, so a
// write replaces one entry and rebuilds, and a scan that lost the race
// takes the entries of the writes it missed.
type entry struct {
	path      string
	root      Root
	kind      kind
	version   string
	issue     *mirror.Issue
	milestone *mirror.Milestone
	board     *mirror.Board
	stale     bool
	err       string
}

// Snapshot is an immutable index of the collection.
type Snapshot struct {
	Generation uint64
	IndexedAt  time.Time
	Roots      []Root
	Tasks      []*Task
	Catalog    Catalog
	Warnings   []string
	byKey      map[string]*Task
	entries    map[string]*entry
}

// Task returns the task of a key.
func (s *Snapshot) Task(key string) *Task {
	if s == nil {
		return nil
	}
	return s.byKey[key]
}

// readEntry parses one mirror file; a failure yields a broken entry.
func readEntry(root Root, path string) *entry {
	e := &entry{path: path, root: root}
	doc, err := mirror.Read(path)
	if err != nil {
		e.err = err.Error()
		return e
	}
	e.version = doc.Version
	switch filepath.Base(filepath.Dir(path)) {
	case "roadmap":
		if mirror.IsMilestone(doc) {
			e.kind, e.milestone = kindMilestone, mirror.ParseMilestone(doc)
		} else {
			e.kind, e.board = kindBoard, mirror.ParseBoard(doc)
		}
	default:
		e.kind, e.issue = kindIssue, mirror.ParseIssue(doc)
	}
	return e
}

// readRoot parses every mirror file of a root.
func readRoot(root Root) []*entry {
	var out []*entry
	for _, dir := range []string{"issues", "roadmap"} {
		paths, _ := filepath.Glob(filepath.Join(root.Path, ".github", dir, "*.md"))
		sort.Strings(paths)
		for _, path := range paths {
			if info, err := os.Lstat(path); err != nil || !info.Mode().IsRegular() {
				continue
			}
			out = append(out, readEntry(root, path))
		}
	}
	return out
}

// HasMirror reports whether a checkout has an issue or roadmap mirror.
func HasMirror(path string) bool {
	for _, dir := range []string{"issues", "roadmap"} {
		if info, err := os.Stat(filepath.Join(path, ".github", dir)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// build derives the tasks, boards, columns, scores and the catalog from the
// entries at one UTC time.
func build(entries map[string]*entry, roots []Root, now time.Time, weights config.Weights, repoWeight func(Root) float64) *Snapshot {
	now = now.UTC()
	snap := &Snapshot{IndexedAt: now, Roots: roots, byKey: map[string]*Task{}, entries: entries, Warnings: []string{}}
	paths := make([]string, 0, len(entries))
	for path := range entries {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	// Milestones by repository and code: the wiki links resolve through them.
	type repoCode struct{ repo, code string }
	milestones := map[repoCode]*Milestone{}
	var catalogMilestones []Milestone
	for _, path := range paths {
		e := entries[path]
		if e.kind != kindMilestone {
			continue
		}
		m := Milestone{
			Key: MilestoneKey(e.root.Owner, e.root.Name, e.milestone.Number), Repository: e.root.Repository(),
			Code: e.milestone.Code, Number: e.milestone.Number, Title: e.milestone.Title, State: e.milestone.State, DueOn: e.milestone.DueOn,
		}
		if e.milestone.URL != "" {
			u := e.milestone.URL
			m.URL = &u
		}
		rc := repoCode{strings.ToLower(e.root.Repository()), e.milestone.Code}
		if _, dup := milestones[rc]; dup {
			continue
		}
		milestones[rc] = &m
		catalogMilestones = append(catalogMilestones, m)
	}

	// Tasks, the first mirror of a key wins; a second one flags it.
	byCode := map[repoCode]*Task{}
	labels := map[string]bool{}
	for _, path := range paths {
		e := entries[path]
		switch e.kind {
		case kindBroken:
			snap.Warnings = append(snap.Warnings, fmt.Sprintf("%s: %s", path, e.err))
			continue
		case kindIssue:
		default:
			continue
		}
		issue := e.issue
		key := TaskKey(e.root.Owner, e.root.Name, issue.Number, issue.Code)
		if existing, dup := snap.byKey[key]; dup {
			existing.Flags.DuplicateMirror = true
			snap.Warnings = append(snap.Warnings, fmt.Sprintf("%s: duplicate mirror of %s (%s)", path, key, existing.path))
			continue
		}
		task := &Task{
			Key: key, ID: EncodeID(key), Repository: e.root.Repository(), Owner: e.root.Owner,
			Code: issue.Code, Number: issue.Number, Title: issue.Title, State: issue.State, StateReason: issue.StateReason,
			Labels: issue.Labels, Assignees: issue.Assignees, IssueFields: issue.IssueFields,
			CreatedAt: issue.CreatedAt, UpdatedAt: issue.UpdatedAt, ClosedAt: issue.ClosedAt,
			Boards: []Membership{}, Version: e.version, Flags: Flags{Stale: e.stale},
			path: path, root: e.root, body: issue.Body,
		}
		if issue.URL != "" {
			u := issue.URL
			task.URL = &u
		}
		for _, c := range issue.Comments {
			task.comments = append(task.comments, comment{At: c.At, URL: c.URL, Body: c.Body})
		}
		task.Type = taskType(issue)
		if issue.Milestone != "" {
			if m, ok := milestones[repoCode{strings.ToLower(task.Repository), issue.Milestone}]; ok {
				task.Milestone = &MilestoneRef{Key: m.Key, Title: m.Title}
				task.milestoneDue = m.DueOn
			} else {
				task.Milestone = &MilestoneRef{Key: fmt.Sprintf("github:%s/milestones/%s", task.Repository, issue.Milestone), Title: issue.Milestone}
			}
		}
		for _, label := range issue.Labels {
			labels[label] = true
		}
		snap.byKey[key] = task
		byCode[repoCode{strings.ToLower(task.Repository), issue.Code}] = task
		snap.Tasks = append(snap.Tasks, task)
	}
	byLowerKey := make(map[string]*Task, len(snap.byKey))
	for key, task := range snap.byKey {
		byLowerKey[strings.ToLower(key)] = task
	}

	// Boards merge by native id; memberships attach by task key.
	boards := map[string]*Board{}
	var boardKeys []string
	for _, path := range paths {
		e := entries[path]
		if e.kind != kindBoard {
			continue
		}
		b := e.board
		key := b.NativeID
		if key == "" {
			key = e.root.Repository() + ":" + b.Code
		}
		merged, ok := boards[key]
		if !ok {
			merged = &Board{Key: key, Title: b.Title, Repositories: []string{}, Sections: []string{}}
			if b.URL != "" {
				u := b.URL
				merged.URL = &u
			}
			boards[key] = merged
			boardKeys = append(boardKeys, key)
		}
		if !contains(merged.Repositories, e.root.Repository()) {
			merged.Repositories = append(merged.Repositories, e.root.Repository())
		}
		for _, section := range b.Sections {
			if !contains(merged.Sections, section.Name) {
				merged.Sections = append(merged.Sections, section.Name)
			}
			for _, item := range section.Items {
				var task *Task
				switch {
				case item.Ref != "":
					task = byLowerKey[strings.ToLower("github:"+item.Ref)]
				case item.Link != "":
					task = byCode[repoCode{strings.ToLower(e.root.Repository()), item.Link}]
				}
				if task == nil {
					continue
				}
				membership := Membership{Board: key, Section: section.Name}
				if !containsMembership(task.Boards, membership) {
					task.Boards = append(task.Boards, membership)
				}
			}
		}
	}

	// Columns, flags and scores.
	for _, task := range snap.Tasks {
		task.Column, task.Flags.Ambiguous, task.Flags.Contradiction = deriveColumn(task)
		task.Score = score(task, now, weights, repoWeight(task.root))
	}
	sort.Slice(snap.Tasks, func(i, j int) bool { return snap.Tasks[i].Key < snap.Tasks[j].Key })

	// The catalog.
	counts, open := map[string]int{}, map[string]int{}
	for _, task := range snap.Tasks {
		counts[task.Repository]++
		if task.State == "open" {
			open[task.Repository]++
		}
	}
	snap.Catalog.Repositories = []Repository{}
	for _, root := range roots {
		snap.Catalog.Repositories = append(snap.Catalog.Repositories, Repository{
			Repository: root.Repository(), Owner: root.Owner, Name: root.Name, Weight: repoWeight(root), Path: root.Path, Tasks: counts[root.Repository()], Open: open[root.Repository()],
			Private: root.Visibility == "private" || root.Visibility == "internal",
		})
	}
	sort.Slice(snap.Catalog.Repositories, func(i, j int) bool {
		return strings.ToLower(snap.Catalog.Repositories[i].Repository) < strings.ToLower(snap.Catalog.Repositories[j].Repository)
	})
	sort.Slice(catalogMilestones, func(i, j int) bool { return catalogMilestones[i].Key < catalogMilestones[j].Key })
	snap.Catalog.Milestones = catalogMilestones
	if snap.Catalog.Milestones == nil {
		snap.Catalog.Milestones = []Milestone{}
	}
	sort.Strings(boardKeys)
	snap.Catalog.Boards = []Board{}
	for _, key := range boardKeys {
		b := boards[key]
		sort.Strings(b.Repositories)
		snap.Catalog.Boards = append(snap.Catalog.Boards, *b)
	}
	snap.Catalog.Labels = []string{}
	for label := range labels {
		snap.Catalog.Labels = append(snap.Catalog.Labels, label)
	}
	sort.Strings(snap.Catalog.Labels)
	snap.Catalog.Columns = append([]string(nil), Columns...)
	return snap
}

// taskType is the non-null, non-empty issueType, else the single type label.
func taskType(issue *mirror.Issue) *string {
	if issue.IssueType != nil {
		t := strings.ToLower(strings.TrimSpace(*issue.IssueType))
		if t != "" && t != "null" {
			return &t
		}
	}
	var found []string
	for _, label := range issue.Labels {
		if rest, ok := strings.CutPrefix(strings.ToLower(label), "type:"); ok {
			found = append(found, strings.TrimSpace(rest))
		}
	}
	if len(found) == 1 && found[0] != "" {
		return &found[0]
	}
	return nil
}

// deriveColumn is total: the state wins, then the normalized sections of
// the task's boards. An "Items" section is the old flat board format, a
// listing without columns: it keeps the membership but casts no vote. Any
// other section that names no column counts as backlog, so it disagrees
// with a To do or In progress section of another board; disagreeing boards
// yield backlog and ambiguous; an open task under Done or Won't do gets
// contradiction.
func deriveColumn(task *Task) (column string, ambiguous, contradiction bool) {
	if task.State == "closed" {
		if task.StateReason != nil && (*task.StateReason == "not_planned" || *task.StateReason == "duplicate") {
			return ColumnWontDo, false, false
		}
		return ColumnDone, false, false
	}
	seen := map[string]bool{}
	for _, m := range task.Boards {
		if isListing(m.Section) {
			continue
		}
		switch c := mirror.NormalizeSection(m.Section); c {
		case ColumnDone, ColumnWontDo:
			contradiction = true
		case "":
			seen[ColumnBacklog] = true
		default:
			seen[c] = true
		}
	}
	switch len(seen) {
	case 0:
		return ColumnBacklog, false, contradiction
	case 1:
		for c := range seen {
			return c, false, contradiction
		}
	}
	return ColumnBacklog, true, contradiction
}

// isListing reports the section of the old flat board format, "Items",
// optionally with a count in parentheses.
func isListing(section string) bool {
	name := strings.ToLower(strings.TrimSpace(section))
	if i := strings.Index(name, "("); i >= 0 {
		name = strings.TrimSpace(name[:i])
	}
	return name == "items"
}

// score is repo_weight × (impact·w_impact + ease·w_effort + type·w_type +
// age·w_age + due·w_due) with every input clamped to [0, 1].
func score(task *Task, now time.Time, w config.Weights, repoWeight float64) Score {
	s := Score{RepoWeight: repoWeight}
	switch Impact(task) {
	case "high":
		s.Impact = 1
	case "medium":
		s.Impact = 0.5
	case "low":
		s.Impact = 0.25
	}
	switch Effort(task) {
	case "easy":
		s.Ease = 1
	case "medium":
		s.Ease = 0.5
	case "hard":
		s.Ease = 0.25
	}
	if task.Type != nil {
		switch *task.Type {
		case "bug":
			s.Type = 1
		case "feature":
			s.Type = 0.5
		}
	}
	if task.CreatedAt != nil && w.AgeHorizonDays > 0 {
		days := now.Sub(*task.CreatedAt).Hours() / 24
		s.Age = clamp(days / float64(w.AgeHorizonDays))
	}
	if task.Milestone != nil && w.DueHorizonDays > 0 {
		if due := task.milestoneDue; due != nil {
			days := due.Sub(now).Hours() / 24
			s.Due = clamp(1 - days/float64(w.DueHorizonDays))
		}
	}
	s.Total = repoWeight * (s.Impact*w.Impact + s.Ease*w.Effort + s.Type*w.Type + s.Age*w.Age + s.Due*w.Due)
	if math.IsNaN(s.Total) || math.IsInf(s.Total, 0) {
		s.Total = 0
	}
	return s
}

func clamp(v float64) float64 {
	if math.IsNaN(v) {
		return 0
	}
	return math.Max(0, math.Min(1, v))
}

// Impact returns the value of the task's impact label, or "".
func Impact(task *Task) string { return labelValue(task.Labels, "impact") }

// Effort returns the value of the task's effort label, or "".
func Effort(task *Task) string { return labelValue(task.Labels, "effort") }

func labelValue(labels []string, prefix string) string {
	value := ""
	for _, label := range labels {
		if rest, ok := strings.CutPrefix(strings.ToLower(label), prefix+":"); ok {
			if value != "" {
				return "" // several labels decide nothing
			}
			value = strings.TrimSpace(rest)
		}
	}
	return value
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func containsMembership(list []Membership, m Membership) bool {
	for _, item := range list {
		if item == m {
			return true
		}
	}
	return false
}
