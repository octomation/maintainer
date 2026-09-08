package status

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
)

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
	return []string{safeText(row.Repository), safeText(branch), safeText(changes), safeText(status)}
}

type table struct {
	widths []int
	rows   [][]string
}

func newTable(rows []Row) table {
	t := table{widths: make([]int, 4), rows: [][]string{{"REPOSITORY", "BRANCH", "UNCOMMITTED", "STATUS"}}}
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
		if i == 0 {
			if _, err := fmt.Fprintln(w, strings.Repeat("─", runewidth.StringWidth(t.line(0)))); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(w, "\n* default branch · +/- tracked lines · f changed files · ? untracked · ! conflicts")
	return err
}

// window clips by terminal cell width, including wide Unicode characters.
func window(s string, offset, width int) string {
	if width <= 0 {
		return ""
	}
	var out strings.Builder
	position, used := 0, 0
	for _, r := range s {
		n := runewidth.RuneWidth(r)
		if position < offset {
			position += n
			continue
		}
		if used+n > width {
			break
		}
		out.WriteRune(r)
		used += n
	}
	return out.String()
}

// Selection is independent of terminal I/O so a future detail view can retain
// the selected row and scroll position while opening/closing another screen.
type Selection struct{ Index, Top, Left int }

func (s *Selection) Move(key string, count, height int) {
	height = max(height, 1)
	switch key {
	case "j", "\x1b[B":
		s.Index++
	case "k", "\x1b[A":
		s.Index--
	case "\x1b[6~":
		s.Index += height
	case "\x1b[5~":
		s.Index -= height
	case "g", "\x1b[H", "\x1b[1~", "\x1bOH":
		s.Index = 0
	case "G", "\x1b[F", "\x1b[4~", "\x1bOF":
		s.Index = count - 1
	case "h", "\x1b[D":
		s.Left = max(0, s.Left-8)
	case "l", "\x1b[C":
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

func frame(t table, rows []Row, sel *Selection, width, height int) string {
	width, height = max(width, 1), max(height, 1)
	page := max(1, height-6)
	sel.Move("", len(rows), page)
	sel.Left = min(sel.Left, max(0, runewidth.StringWidth(t.line(0))+2-width))
	lines := []string{
		fmt.Sprintf("Repository status · %d checkouts · local refs (no fetch)", len(rows)),
		window("  "+t.line(0), sel.Left, width),
	}
	for i := 0; i < page; i++ {
		idx := sel.Top + i
		line := ""
		if idx < len(rows) {
			prefix := "  "
			if idx == sel.Index {
				prefix = "› "
			}
			line = window(prefix+t.line(idx+1), sel.Left, width)
			if idx == sel.Index {
				line = "\x1b[7m" + line + "\x1b[0m"
			}
		}
		lines = append(lines, line)
	}
	if len(rows) > 0 {
		r := rows[sel.Index]
		pin := ""
		if r.Pinned {
			pin = " [pinned]"
		}
		lines = append(lines, window(fmt.Sprintf("%d/%d  %s%s", sel.Index+1, len(rows), safeText(r.Path), pin), 0, width))
		lines = append(lines, window(fmt.Sprintf("HEAD %.12s  upstream %s", safeText(r.Commit), safeText(r.Upstream)), 0, width))
	} else {
		lines = append(lines, "No repositories found.", "")
	}
	lines = append(lines, "↑↓/jk select · PgUp/PgDn · g/G first/last · ←→/hl scroll · q quit", "* default · +/- lines · f files · ? untracked · ! conflicts")
	var out strings.Builder
	out.WriteString("\x1b[H")
	for i, line := range lines[:min(height, len(lines))] {
		if i > 0 {
			out.WriteString("\r\n")
		}
		// Selected rows were clipped before adding ANSI; other lines are plain.
		if !strings.HasPrefix(line, "\x1b[7m") {
			line = window(line, 0, width)
		}
		out.WriteString(line)
		out.WriteString("\x1b[K")
	}
	out.WriteString("\x1b[J")
	return out.String()
}
