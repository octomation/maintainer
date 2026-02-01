package mirror

import (
	"regexp"
	"strings"
	"time"
)

// Milestone is a parsed milestone file of .github/roadmap.
type Milestone struct {
	Code   string // the frontmatter code, else the file stem: the wiki link target
	Number int
	NodeID string
	Title  string
	State  string // open | closed
	DueOn  *time.Time
	URL    string
}

// Board is a parsed project board file of .github/roadmap.
type Board struct {
	Code     string // the frontmatter code, else the file stem
	NativeID string // the ProjectV2 node id when the file has one
	Title    string
	URL      string
	Sections []Section
}

// Section is a heading of a board with its items.
type Section struct {
	Name  string
	Items []Item
}

// Item is one task line of a section.
type Item struct {
	Checked bool
	Link    string // a wiki link target, [[CODE-N]], resolved in the mirror's repository
	Ref     string // an owner/repo#N reference
	Text    string
}

// IsMilestone reports whether a roadmap document is a milestone: it has the
// due date key of its generation.
func IsMilestone(doc *Document) bool {
	return doc.Has("dueOn") || doc.Has("due_on") || doc.Has("openIssueCount") || doc.Has("open_issues")
}

// ParseMilestone reads a milestone from a document.
func ParseMilestone(doc *Document) *Milestone {
	m := &Milestone{Code: strings.TrimSpace(doc.String("code")), Title: doc.String("title"), URL: strings.TrimSpace(doc.String("url")), NodeID: doc.String("node_id")}
	if m.Code == "" {
		m.Code = Stem(doc.Path)
	}
	if doc.Has("status") && !doc.Has("state") {
		m.Number, _ = doc.Int("id")
		m.State = normalizeState(doc.String("status"))
		m.DueOn = doc.Time("due_on")
	} else {
		m.Number, _ = doc.Int("number")
		m.State = normalizeState(doc.String("state"))
		m.DueOn = doc.Time("dueOn")
		if m.NodeID == "" {
			m.NodeID = doc.String("id")
		}
	}
	return m
}

var (
	heading     = regexp.MustCompile(`^#{2,}\s+(.*?)\s*$`)
	headingSize = regexp.MustCompile(`\s*\(\d+\)$`)
	itemLine    = regexp.MustCompile(`^\s*[-*]\s+\[([ xX])\]\s+(.*)$`)
	wikiRef     = regexp.MustCompile(`\[\[([^\]|#]+)(?:[|#][^\]]*)?\]\]`)
	repoRef     = regexp.MustCompile(`^([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)#(\d+)`)
)

// ParseBoard reads a board from a document: its sections are the level-two
// headings of the body and their task lines, `- [ ] [[CODE-N]] — title` or
// `- [x] owner/repo#N — title ([[issue-N]])`.
func ParseBoard(doc *Document) *Board {
	b := &Board{Code: strings.TrimSpace(doc.String("code")), Title: doc.String("title"), URL: strings.TrimSpace(doc.String("url"))}
	if b.Code == "" {
		b.Code = Stem(doc.Path)
	}
	if id := doc.String("node_id"); id != "" {
		b.NativeID = id
	} else if id, ok := doc.Fields["id"].(string); ok {
		b.NativeID = strings.TrimSpace(id)
	}
	var current *Section
	var fence string
	for _, line := range strings.Split(doc.Body, "\n") {
		line = strings.TrimRight(line, "\r")
		if fence != "" {
			if closesFence(line, fence) {
				fence = ""
			}
			continue
		}
		if marker := fenceMarker(line); marker != "" {
			fence = marker
			continue
		}
		if m := heading.FindStringSubmatch(line); m != nil {
			name := headingSize.ReplaceAllString(strings.TrimSpace(m[1]), "")
			b.Sections = append(b.Sections, Section{Name: name})
			current = &b.Sections[len(b.Sections)-1]
			continue
		}
		if current == nil {
			continue
		}
		m := itemLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		item := Item{Checked: m[1] != " ", Text: strings.TrimSpace(m[2])}
		text := item.Text
		if r := repoRef.FindStringSubmatch(text); r != nil {
			item.Ref = r[1] + "#" + r[2]
		} else if strings.HasPrefix(text, "[[") {
			if w := wikiRef.FindStringSubmatch(text); w != nil {
				item.Link = strings.TrimSpace(w[1])
			}
		}
		if item.Ref == "" && item.Link == "" {
			continue // a draft or free text names no task
		}
		current.Items = append(current.Items, item)
	}
	return b
}

// NormalizeSection maps a section name to a column token: To do and Todo to
// todo, In progress to in_progress, Backlog to backlog, Done to done and
// Won't do to wont_do; any other name yields "".
func NormalizeSection(name string) string {
	key := strings.ToLower(strings.TrimSpace(name))
	key = strings.NewReplacer("’", "'", "‘", "'", "-", " ", "_", " ").Replace(key)
	key = strings.Join(strings.Fields(key), " ")
	switch key {
	case "to do", "todo":
		return "todo"
	case "in progress", "inprogress", "doing":
		return "in_progress"
	case "backlog":
		return "backlog"
	case "done":
		return "done"
	case "won't do", "wont do", "won t do":
		return "wont_do"
	default:
		return ""
	}
}
