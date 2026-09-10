package status

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
)

func relativePath(path, root, home string) string {
	for _, base := range []struct{ path, prefix string }{{root, ""}, {home, "~"}} {
		if base.path == "" {
			continue
		}
		rel, err := filepath.Rel(base.path, path)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return filepath.Join(base.prefix, rel)
		}
	}
	return path
}

func (row Row) pathCell() string {
	if row.displayPath != "" {
		return safeText(row.displayPath)
	}
	return safeText(row.Path)
}

func safeText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.In(r, unicode.Cf) {
			return '�'
		}
		return r
	}, s)
}

func cells(row Row) []string {
	branch := row.Branch
	if branch == "" {
		branch = "—"
	}
	if row.Branch != "" && row.Branch == row.DefaultBranch {
		branch += "*"
	}
	changes := fmt.Sprintf("+%d/-%d", row.Added, row.Deleted)
	if row.Changed > 0 {
		changes += fmt.Sprintf(" · %df", row.Changed)
	}
	if row.Untracked > 0 {
		changes += fmt.Sprintf(" · ?%d", row.Untracked)
	}
	if row.Binary > 0 {
		changes += fmt.Sprintf(" · %d binary", row.Binary)
	}
	if row.Conflicts > 0 {
		changes += fmt.Sprintf(" · !%d", row.Conflicts)
	}
	status := row.Status
	if row.Error != "" {
		changes = "?"
		status = "error: " + row.Error
	}
	if row.Status == "diverged" || row.Status == "ahead" || row.Status == "behind" {
		var parts []string
		if row.Ahead > 0 {
			parts = append(parts, fmt.Sprintf("ahead %d", row.Ahead))
		}
		if row.Behind > 0 {
			parts = append(parts, fmt.Sprintf("behind %d", row.Behind))
		}
		status = strings.Join(parts, " · ")
	}
	if row.OrphanReason != "" {
		status = "orphan · " + status
	}
	lock := ""
	if row.PushLocked {
		lock = "🔒"
	}
	return []string{safeText(row.Repository), safeText(branch), safeText(changes), safeText(status), lock, row.pathCell()}
}

type table struct {
	widths []int
	rows   [][]string
}

func newTable(rows []Row) table {
	headers := make([]string, len(columnNames))
	for i, name := range columnNames {
		headers[i] = strings.ToUpper(name)
	}
	t := table{widths: make([]int, len(headers)), rows: [][]string{headers}}
	for _, row := range rows {
		t.rows = append(t.rows, cells(row))
	}
	for _, row := range t.rows {
		for i, cell := range row {
			t.widths[i] = max(t.widths[i], runewidth.StringWidth(cell))
		}
	}
	return t
}

func (t table) line(i int) string {
	var out strings.Builder
	for col, cell := range t.rows[i] {
		if col > 0 {
			out.WriteString(" │ ")
		}
		out.WriteString(cell)
		out.WriteString(strings.Repeat(" ", t.widths[col]-runewidth.StringWidth(cell)))
	}
	return out.String()
}

// Plain writes every row, without terminal controls or interactive input.
func Plain(w io.Writer, rows []Row) error {
	t := newTable(rows)
	if _, err := fmt.Fprintf(w, "Repository status · %d checkouts · local refs (no fetch)\n\n", len(rows)); err != nil {
		return err
	}
	for i := range t.rows {
		if _, err := fmt.Fprintln(w, t.line(i)); err != nil {
			return err
		}
		if i > 0 && rows[i-1].OrphanReason != "" {
			row := rows[i-1]
			if _, err := fmt.Fprintf(w, "  path: %s\n", safeText(row.Path)); err != nil {
				return err
			}
			if row.ActivePath != "" {
				if _, err := fmt.Fprintf(w, "  active: %s\n", safeText(row.ActivePath)); err != nil {
					return err
				}
			}
			if row.RemoteCheckedAt != nil {
				if _, err := fmt.Fprintf(w, "  GitHub check (cached): %s\n", row.RemoteCheckedAt.UTC().Format("2006-01-02T15:04:05Z")); err != nil {
					return err
				}
			}
		}
		if i == 0 {
			if _, err := fmt.Fprintln(w, strings.Repeat("─", runewidth.StringWidth(t.line(0)))); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(w, "\n* default branch · +/- tracked lines · f changed files · ? untracked · ! conflicts · 🔒 push locked")
	return err
}

// Selection retains row and viewport positions independently of the renderer.
type Selection struct{ Index, Top, Left int }

func (s *Selection) Move(key string, count, height int) {
	height = max(height, 1)
	switch key {
	case "j", "down":
		s.Index++
	case "k", "up":
		s.Index--
	case "pgdown":
		s.Index += height
	case "pgup":
		s.Index -= height
	case "g", "home":
		s.Index = 0
	case "G", "end":
		s.Index = count - 1
	case "h":
		s.Left = max(0, s.Left-8)
	case "l":
		s.Left += 8
	}
	s.Index = max(0, min(s.Index, count-1))
	if s.Index < s.Top {
		s.Top = s.Index
	}
	if s.Index >= s.Top+height {
		s.Top = s.Index - height + 1
	}
	s.Top = max(0, min(s.Top, count-height))
}
