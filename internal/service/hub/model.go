// Package hub indexes the issue mirrors of the collection into one model of
// tasks, boards and milestones, and answers the queries of the pages.
package hub

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"
)

// Column tokens of the basic board.
const (
	ColumnBacklog    = "backlog"
	ColumnTodo       = "todo"
	ColumnInProgress = "in_progress"
	ColumnDone       = "done"
	ColumnWontDo     = "wont_do"
)

// Columns lists the column tokens in board order.
var Columns = []string{ColumnBacklog, ColumnTodo, ColumnInProgress, ColumnDone, ColumnWontDo}

// Root is a mirror root: a real checkout of owner/name whose .github holds
// the issues and the roadmap.
type Root struct {
	Path  string `json:"path"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
	ID    int64  `json:"id,omitempty"`
	// Visibility is the repository's visibility on GitHub from the fetch
	// state (public, private, internal); empty when unknown.
	Visibility string `json:"visibility,omitempty"`
	// Real is the path with every symlink resolved, as seen at discovery: a
	// write refuses a root whose real path changed since, so a root or an
	// ancestor of it replaced by a symlink redirects nothing.
	Real string `json:"-"`
	// Info is the identity of the root directory as seen at discovery: the
	// writer requires the directory it opens to be the same file.
	Info os.FileInfo `json:"-"`
}

// Repository returns owner/name.
func (r Root) Repository() string { return r.Owner + "/" + r.Name }

// Task is one issue of the collection.
type Task struct {
	Key         string         `json:"key"`
	ID          string         `json:"id"`
	Repository  string         `json:"repository"`
	Owner       string         `json:"owner"`
	Code        string         `json:"code"`
	Number      *int           `json:"number"`
	URL         *string        `json:"url"`
	Title       string         `json:"title"`
	State       string         `json:"state"`
	StateReason *string        `json:"state_reason"`
	Type        *string        `json:"type"`
	Labels      []string       `json:"labels"`
	Assignees   []string       `json:"assignees"`
	Milestone   *MilestoneRef  `json:"milestone"`
	IssueFields map[string]any `json:"issue_fields"`
	CreatedAt   *time.Time     `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	ClosedAt    *time.Time     `json:"closed_at"`
	Boards      []Membership   `json:"boards"`
	Column      string         `json:"column"`
	Flags       Flags          `json:"flags"`
	Score       Score          `json:"score"`
	Version     string         `json:"version"`

	// Provenance and content stay out of the list responses.
	path         string
	root         Root
	body         string
	comments     []comment
	milestoneDue *time.Time
}

// Path is the provenance of the task: the mirror file it was read from.
func (t *Task) Path() string { return t.path }

// Root is the mirror root the task was read from.
func (t *Task) Root() Root { return t.root }

// Body is the Markdown of the task without the comment blocks.
func (t *Task) Body() string { return t.body }

type comment struct {
	At   time.Time
	URL  string
	Body string
}

// MilestoneRef names the milestone of a task.
type MilestoneRef struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}

// Membership is a board the task is listed on, with the section as written.
type Membership struct {
	Board   string `json:"board"`
	Section string `json:"section"`
}

// Flags are the derived warnings of a task.
type Flags struct {
	Ambiguous       bool `json:"ambiguous"`
	Contradiction   bool `json:"contradiction"`
	DuplicateMirror bool `json:"duplicate_mirror"`
	Stale           bool `json:"stale"`
}

// Score is the total and its components.
type Score struct {
	Total      float64 `json:"total"`
	RepoWeight float64 `json:"repo_weight"`
	Impact     float64 `json:"impact"`
	Ease       float64 `json:"ease"`
	Type       float64 `json:"type"`
	Age        float64 `json:"age"`
	Due        float64 `json:"due"`
}

// Milestone is a milestone of the catalog.
type Milestone struct {
	Key        string     `json:"key"`
	Repository string     `json:"repository"`
	Code       string     `json:"code"`
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	State      string     `json:"state"`
	DueOn      *time.Time `json:"due_on"`
	URL        *string    `json:"url"`
}

// Board is a board of the catalog, merged across its partial mirrors.
type Board struct {
	Key          string   `json:"key"`
	Title        string   `json:"title"`
	URL          *string  `json:"url"`
	Repositories []string `json:"repositories"`
	Sections     []string `json:"sections"`
}

// Repository is a repository of the catalog.
type Repository struct {
	Repository string  `json:"repository"`
	Owner      string  `json:"owner"`
	Name       string  `json:"name"`
	Weight     float64 `json:"weight"`
	Path       string  `json:"path"`
	Tasks      int     `json:"tasks"`
	// Open counts the open tasks of the repository.
	Open int `json:"open"`
	// Private marks a repository that is not public on GitHub (private or
	// internal) by the fetch state; false when the state does not know.
	Private bool `json:"private"`
}

// Catalog lists what the filters can select.
type Catalog struct {
	Repositories []Repository `json:"repositories"`
	Milestones   []Milestone  `json:"milestones"`
	Boards       []Board      `json:"boards"`
	Labels       []string     `json:"labels"`
	Columns      []string     `json:"columns"`
}

// TaskKey builds the source key of a task: github:<owner>/<repo>#<number>,
// or github:<owner>/<repo>:<code> without a number.
func TaskKey(owner, name string, number *int, code string) string {
	if number != nil {
		return fmt.Sprintf("github:%s/%s#%d", owner, name, *number)
	}
	return fmt.Sprintf("github:%s/%s:%s", owner, name, code)
}

// MilestoneKey builds github:<owner>/<repo>/milestones/<N>.
func MilestoneKey(owner, name string, number int) string {
	return fmt.Sprintf("github:%s/%s/milestones/%d", owner, name, number)
}

// EncodeID is the HTTP id of a key: unpadded base64url, opaque to clients.
func EncodeID(key string) string { return base64.RawURLEncoding.EncodeToString([]byte(key)) }

// DecodeID returns the key of an HTTP id.
func DecodeID(id string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil || !strings.HasPrefix(string(raw), "github:") {
		return "", fmt.Errorf("malformed task id")
	}
	return string(raw), nil
}
