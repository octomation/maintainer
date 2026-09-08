package status

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *screen) tableWidth() int {
	width := 3 * (len(m.widths) - 1)
	for _, n := range m.widths {
		width += n
	}
	return width
}

func (m *screen) sortLabel(column Column) string {
	for i, key := range m.keys {
		if key.Column == column {
			arrow := "↑"
			if key.Desc {
				arrow = "↓"
			}
			return fmt.Sprintf(" %s%d", arrow, i+1)
		}
	}
	return ""
}

func (m *screen) sortSummary() string {
	if len(m.keys) == 0 {
		if strings.TrimSpace(m.search.Value()) != "" {
			return "Sort: relevance · 0 resets sorting"
		}
		return "Sort: default (repository / path)"
	}
	parts := make([]string, len(m.keys))
	for i, key := range m.keys {
		parts[i] = columnNames[key.Column] + m.sortLabel(key.Column)
	}
	return "Sort: " + strings.Join(parts, "  →  ")
}

func (m *screen) View() tea.View {
	lightDark := lipgloss.LightDark(m.dark)
	accent := lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#5B21B6"), lipgloss.Color("#C4B5FD")))
	muted := lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#52525B"), lipgloss.Color("#A1A1AA")))
	focus := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#6D28D9"))
	selected := accent.Reverse(true)
	borderColor := lightDark(lipgloss.Color("#A1A1AA"), lipgloss.Color("#52525B"))
	width := max(1, m.width-2)
	clip := func(s string) string { return ansi.Truncate(s, m.width, "") }
	cell := func(value string, col int) string {
		value = ansi.Truncate(value, m.widths[col], "…")
		return value + strings.Repeat(" ", max(0, m.widths[col]-ansi.StringWidth(value)))
	}
	var headers []string
	for i, name := range columnNames {
		if i == 0 {
			name = "  " + name
		}
		label := cell(name+m.sortLabel(Column(i)), i)
		if Column(i) == m.column {
			label = focus.Render(label)
		} else {
			label = accent.Bold(true).Render(label)
		}
		headers = append(headers, label)
	}
	line := func(s string) string { return ansi.Cut(s, m.selection.Left, m.selection.Left+width) }
	content := []string{line(strings.Join(headers, " │ ")), muted.Render(strings.Repeat("─", min(width, m.tableWidth())))}
	for i := 0; i < m.pageSize(); i++ {
		index := m.selection.Top + i
		if index >= len(m.visible) {
			text := ""
			if i == 0 {
				text = "No matches · Esc clears the filter"
			}
			content = append(content, muted.Render(ansi.Truncate(text, width, "")))
			continue
		}
		row := m.rows[m.visible[index]]
		values := cells(row)
		prefix := "  "
		if index == m.selection.Index {
			prefix = "› "
		}
		values[0] = prefix + values[0]
		for col := range values {
			values[col] = cell(values[col], col)
		}
		text := line(strings.Join(values, " │ "))
		text += strings.Repeat(" ", max(0, min(width, m.tableWidth())-ansi.StringWidth(text)))
		if index == m.selection.Index {
			text = selected.Render(text)
		} else if row.Error != "" {
			text = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(text)
		} else if row.OrphanReason != "" {
			text = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render(text)
		}
		content = append(content, text)
	}
	// Lip Gloss includes the border in Width. The content already reserves its
	// two cells; subtracting them again would wrap headers and shift mouse rows.
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderColor).Width(m.width).Render(strings.Join(content, "\n"))
	lines := []string{
		clip(accent.Bold(true).Render(" MAINTAINER ") + fmt.Sprintf("  %d / %d repositories", len(m.visible), len(m.rows)) + muted.Render(" · local refs")),
		clip(muted.Render(m.sortSummary())),
		clip(m.search.View()),
		box,
	}
	if len(m.visible) > 0 {
		row := m.rows[m.visible[m.selection.Index]]
		pin := ""
		if row.Pinned {
			pin = " · pinned"
		}
		lines = append(lines, clip(accent.Render(fmt.Sprintf("%d/%d  %s%s", m.selection.Index+1, len(m.visible), safeText(row.Path), pin))))
		detail := fmt.Sprintf("HEAD %.12s · upstream %s", safeText(row.Commit), safeText(row.Upstream))
		if row.OrphanReason != "" {
			detail = "orphan [" + safeText(row.OrphanReason) + "] · " + detail
		}
		if row.ActivePath != "" {
			detail = "active: " + safeText(row.ActivePath)
		}
		if row.RemoteCheckedAt != nil && row.OrphanReason != "" {
			detail += " · GitHub cached " + row.RemoteCheckedAt.UTC().Format("2006-01-02T15:04Z")
		}
		lines = append(lines, clip(muted.Render(detail)))
	} else {
		lines = append(lines, clip("No repository selected"), "")
	}
	help := "↑↓/jk rows · Tab/←→ column · s sort · Shift+s append · 0 reset"
	if m.search.Focused() {
		help = "Type to filter · Enter/Esc return to table · Ctrl+u clear"
	}
	lines = append(lines, clip(muted.Render(help)), clip(muted.Render("/ search · Esc clear · q quit · h/l scroll · PgUp/PgDn · g/G")))
	all := strings.Split(strings.Join(lines, "\n"), "\n")
	for i := range all {
		all[i] = clip(all[i])
	}
	v := tea.NewView(strings.Join(all[:min(len(all), m.height)], "\n"))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
