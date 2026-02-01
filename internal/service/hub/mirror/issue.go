package mirror

import (
	"strings"
	"time"
)

// Generation names the frontmatter shape of a file.
type Generation string

// The three frontmatter generations of the issue mirrors.
const (
	// GenerationOld is issue-N.md: an empty code, [[milestone-N]] links, no
	// assignees, the state in upper case.
	GenerationOld Generation = "old"
	// GenerationCode is CODE-N.md: [[CODE-MN]] links, issueType, issueFields,
	// parent, duplicateOf, and number: null for a local-only issue.
	GenerationCode Generation = "code"
	// GenerationDotfiles has id as the GitHub number, status and
	// state_reason in lower case, and created_at and updated_at.
	GenerationDotfiles Generation = "dotfiles"
)

// Issue is a parsed issue file, normalized across the generations.
type Issue struct {
	Generation  Generation
	Code        string // the frontmatter code, else the file stem
	Number      *int   // nil for a local-only issue
	NodeID      string
	URL         string
	Title       string
	IssueType   *string // non-null and non-empty only
	Labels      []string
	Assignees   []string
	Milestone   string // the link target without [[ ]], or ""
	Parent      string
	DuplicateOf string
	IssueFields map[string]any
	State       string  // open | closed
	StateReason *string // completed | not_planned | duplicate
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	ClosedAt    *time.Time
	Body        string // without the trailing comment blocks
	Comments    []Comment
}

// ParseIssue reads an issue from a document; the generation is detected
// from the keys.
func ParseIssue(doc *Document) *Issue {
	issue := &Issue{
		Code:        strings.TrimSpace(doc.String("code")),
		NodeID:      doc.String("node_id"),
		URL:         strings.TrimSpace(doc.String("url")),
		Title:       doc.String("title"),
		Labels:      doc.Strings("labels"),
		Assignees:   doc.Strings("assignees"),
		Milestone:   WikiLink(doc.String("milestone")),
		Parent:      WikiLink(doc.String("parent")),
		DuplicateOf: WikiLink(doc.String("duplicateOf")),
		IssueFields: doc.Map("issueFields"),
	}
	if issue.Code == "" {
		issue.Code = Stem(doc.Path)
	}
	if issue.Labels == nil {
		issue.Labels = []string{}
	}
	if issue.Assignees == nil {
		issue.Assignees = []string{}
	}
	if issue.IssueFields == nil {
		issue.IssueFields = map[string]any{}
	}
	switch {
	case doc.Has("status") && !doc.Has("state"):
		issue.Generation = GenerationDotfiles
		if n, ok := doc.Int("id"); ok {
			issue.Number = &n
		}
		issue.State = normalizeState(doc.String("status"))
		issue.StateReason = normalizeReason(doc.String("state_reason"))
		issue.CreatedAt, issue.UpdatedAt, issue.ClosedAt = doc.Time("created_at"), doc.Time("updated_at"), doc.Time("closed_at")
	default:
		issue.Generation = GenerationCode
		if strings.TrimSpace(doc.String("code")) == "" {
			issue.Generation = GenerationOld
		}
		if n, ok := doc.Int("number"); ok {
			issue.Number = &n
		}
		if id := doc.String("id"); issue.NodeID == "" {
			issue.NodeID = id
		}
		issue.State = normalizeState(doc.String("state"))
		issue.StateReason = normalizeReason(doc.String("stateReason"))
		issue.CreatedAt, issue.UpdatedAt, issue.ClosedAt = doc.Time("createdAt"), doc.Time("updatedAt"), doc.Time("closedAt")
	}
	if t := strings.TrimSpace(doc.String("issueType")); t != "" {
		issue.IssueType = &t
	}
	issue.Body, issue.Comments = SplitComments(doc.Body)
	return issue
}

func normalizeState(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "closed":
		return "closed"
	default:
		return "open"
	}
}

func normalizeReason(s string) *string {
	switch r := strings.ToLower(strings.TrimSpace(s)); r {
	case "completed", "not_planned", "duplicate":
		return &r
	case "reopened":
		return nil
	default:
		return nil
	}
}
