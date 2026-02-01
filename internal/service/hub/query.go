package hub

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Query selects and orders tasks. Repeated values are OR within a dimension,
// dimensions are AND, Labels is "all selected".
type Query struct {
	Owners     []string
	Repos      []string
	State      string // open (default), closed, all
	Milestones []string
	Labels     []string
	Types      []string
	Assignees  []string
	Columns    []string
	Q          string
	Sort       string // score (default), importance, updated, created
	Limit      int    // 100 by default, at most 500
	Offset     int
}

// Sort orders.
const (
	SortScore      = "score"
	SortImportance = "importance"
	SortUpdated    = "updated"
	SortCreated    = "created"
)

// Paging limits.
const (
	DefaultLimit = 100
	MaxLimit     = 500
)

// Facet is one value of a dimension with its count over the selection.
type Facet struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// Facets count the selection, ignoring Columns, Limit and Offset.
type Facets struct {
	Owners       []Facet `json:"owners"`
	Repositories []Facet `json:"repositories"`
	Milestones   []Facet `json:"milestones"`
	Labels       []Facet `json:"labels"`
	Types        []Facet `json:"types"`
	Assignees    []Facet `json:"assignees"`
	Columns      []Facet `json:"columns"`
	States       []Facet `json:"states"`
}

// Page is the answer of a query.
type Page struct {
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Facets Facets  `json:"facets"`
	Tasks  []*Task `json:"tasks"`
}

// Validate normalizes the query and rejects unknown values.
func (q *Query) Validate() error {
	switch q.State {
	case "":
		q.State = "open"
	case "open", "closed", "all":
	default:
		return fmt.Errorf("state must be open, closed or all, not %q", q.State)
	}
	switch q.Sort {
	case "":
		q.Sort = SortScore
	case SortScore, SortImportance, SortUpdated, SortCreated:
	default:
		return fmt.Errorf("sort must be score, importance, updated or created, not %q", q.Sort)
	}
	for _, c := range q.Columns {
		if !contains(Columns, c) {
			return fmt.Errorf("unknown column %q", c)
		}
	}
	if q.Limit < 0 || q.Offset < 0 {
		return fmt.Errorf("limit and offset must not be negative")
	}
	if q.Limit == 0 {
		q.Limit = DefaultLimit
	}
	if q.Limit > MaxLimit {
		q.Limit = MaxLimit
	}
	return nil
}

// Query answers q over the snapshot.
func (s *Snapshot) Query(q Query) Page {
	page := Page{Limit: q.Limit, Offset: q.Offset, Tasks: []*Task{}}
	if s == nil {
		return page
	}
	needle := strings.ToLower(strings.TrimSpace(q.Q))
	selection := make([]*Task, 0, len(s.Tasks))
	for _, task := range s.Tasks {
		if q.State != "all" && task.State != q.State {
			continue
		}
		if len(q.Owners) > 0 && !containsFold(q.Owners, task.Owner) {
			continue
		}
		if len(q.Repos) > 0 && !containsFold(q.Repos, task.Repository) {
			continue
		}
		if len(q.Milestones) > 0 && (task.Milestone == nil || !contains(q.Milestones, task.Milestone.Key)) {
			continue
		}
		if len(q.Types) > 0 && (task.Type == nil || !containsFold(q.Types, *task.Type)) {
			continue
		}
		if len(q.Assignees) > 0 && !anyFold(q.Assignees, task.Assignees) {
			continue
		}
		if len(q.Labels) > 0 && !allFold(q.Labels, task.Labels) {
			continue
		}
		if needle != "" && !matches(task, needle) {
			continue
		}
		selection = append(selection, task)
	}
	page.Facets = facets(selection)
	filtered := selection
	if len(q.Columns) > 0 {
		filtered = make([]*Task, 0, len(selection))
		for _, task := range selection {
			if contains(q.Columns, task.Column) {
				filtered = append(filtered, task)
			}
		}
	}
	sortTasks(filtered, q.Sort)
	page.Total = len(filtered)
	if q.Offset < len(filtered) {
		end := q.Offset + q.Limit
		if end > len(filtered) {
			end = len(filtered)
		}
		page.Tasks = filtered[q.Offset:end]
	}
	return page
}

func matches(task *Task, needle string) bool {
	if strings.Contains(strings.ToLower(task.Code), needle) || strings.Contains(strings.ToLower(task.Title), needle) ||
		strings.Contains(strings.ToLower(task.Repository), needle) {
		return true
	}
	for _, label := range task.Labels {
		if strings.Contains(strings.ToLower(label), needle) {
			return true
		}
	}
	return false
}

// sortTasks orders in place: every order is descending and deterministic,
// a missing date is last, ties break by updated_at descending with a
// missing date last, then by key ascending.
func sortTasks(tasks []*Task, order string) {
	less := func(a, b *Task) int {
		switch order {
		case SortImportance:
			if c := compareFloat(impactRank(a), impactRank(b)); c != 0 {
				return c
			}
		case SortUpdated:
			if c := compareTime(a.UpdatedAt, b.UpdatedAt); c != 0 {
				return c
			}
		case SortCreated:
			if c := compareTime(a.CreatedAt, b.CreatedAt); c != 0 {
				return c
			}
		default:
			if c := compareFloat(a.Score.Total, b.Score.Total); c != 0 {
				return c
			}
		}
		if c := compareTime(a.UpdatedAt, b.UpdatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.Key, b.Key)
	}
	sort.SliceStable(tasks, func(i, j int) bool { return less(tasks[i], tasks[j]) < 0 })
}

// compareFloat orders descending.
func compareFloat(a, b float64) int {
	switch {
	case a > b:
		return -1
	case a < b:
		return 1
	}
	return 0
}

// compareTime orders descending with nil last.
func compareTime(a, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	case a.After(*b):
		return -1
	case b.After(*a):
		return 1
	}
	return 0
}

func impactRank(task *Task) float64 {
	switch Impact(task) {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	}
	return 0
}

func facets(selection []*Task) Facets {
	count := func(get func(*Task) []string) []Facet {
		counts := map[string]int{}
		for _, task := range selection {
			for _, v := range get(task) {
				counts[v]++
			}
		}
		out := make([]Facet, 0, len(counts))
		for v, n := range counts {
			out = append(out, Facet{Value: v, Count: n})
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].Count != out[j].Count {
				return out[i].Count > out[j].Count
			}
			return out[i].Value < out[j].Value
		})
		return out
	}
	return Facets{
		Owners:       count(func(t *Task) []string { return []string{t.Owner} }),
		Repositories: count(func(t *Task) []string { return []string{t.Repository} }),
		Milestones: count(func(t *Task) []string {
			if t.Milestone == nil {
				return nil
			}
			return []string{t.Milestone.Key}
		}),
		Labels: count(func(t *Task) []string { return t.Labels }),
		Types: count(func(t *Task) []string {
			if t.Type == nil {
				return nil
			}
			return []string{*t.Type}
		}),
		Assignees: count(func(t *Task) []string { return t.Assignees }),
		Columns:   count(func(t *Task) []string { return []string{t.Column} }),
		States:    count(func(t *Task) []string { return []string{t.State} }),
	}
}

func containsFold(list []string, s string) bool {
	for _, item := range list {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

func anyFold(wanted, have []string) bool {
	for _, w := range wanted {
		if containsFold(have, w) {
			return true
		}
	}
	return false
}

func allFold(wanted, have []string) bool {
	for _, w := range wanted {
		if !containsFold(have, w) {
			return false
		}
	}
	return true
}
