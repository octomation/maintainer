package status

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func press(m *screen, code rune, text string, mod tea.KeyMod) tea.Cmd {
	_, cmd := m.Update(tea.KeyPressMsg{Code: code, Text: text, Mod: mod})
	return cmd
}

func TestKeyboardSortingAndFilter(t *testing.T) {
	m := newScreen([]Row{{Repository: "acme/maintainer", Path: "/one", Branch: "main"}, {Repository: "acme/dotfiles", Path: "/two", Branch: "main"}})
	press(m, tea.KeyTab, "", 0)
	assert.Equal(t, BranchColumn, m.column)
	press(m, 's', "s", 0)
	press(m, tea.KeyTab, "", 0)
	press(m, 's', "S", tea.ModShift)
	assert.Equal(t, []sortKey{{Column: BranchColumn}, {Column: ChangesColumn}}, m.keys)
	press(m, 'S', "S", 0) // legacy terminal delivers uppercase rather than a modifier
	assert.True(t, m.keys[1].Desc)
	press(m, tea.KeyEnter, "", tea.ModShift)
	assert.Equal(t, []sortKey{{Column: BranchColumn}}, m.keys)
	press(m, '0', "0", 0)
	assert.Empty(t, m.keys)
	press(m, '/', "/", 0)
	require.True(t, m.search.Focused())
	for _, r := range "dtf" {
		press(m, r, string(r), 0)
	}
	require.Len(t, m.visible, 1)
	assert.Equal(t, "acme/dotfiles", m.rows[m.visible[0]].Repository)
	press(m, tea.KeyEnter, "", 0)
	assert.False(t, m.search.Focused())
	press(m, tea.KeyEscape, "", 0)
	assert.Len(t, m.visible, 2)
	assert.Equal(t, "", m.search.Value())
	press(m, '/', "/", 0)
	press(m, 'q', "q", 0)
	assert.Equal(t, "q", m.search.Value(), "quit shortcut must be text in the input")
	press(m, 'u', "", tea.ModCtrl)
	assert.Empty(t, m.search.Value())
	assert.Len(t, m.visible, 2)
}

func TestMouseSortingAndSelectionAcrossResize(t *testing.T) {
	var rows []Row
	for i := 0; i < 80; i++ {
		rows = append(rows, Row{Repository: fmt.Sprintf("repo-%02d", i), Path: fmt.Sprint(i), Added: 80 - i})
	}
	m := newScreen(rows)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	m.Update(tea.MouseClickMsg{X: 2, Y: 4, Button: tea.MouseLeft})
	assert.Equal(t, []sortKey{{Column: RepositoryColumn}}, m.keys)
	m.Update(tea.MouseClickMsg{X: m.widths[0] + 5, Y: 4, Button: tea.MouseLeft, Mod: tea.ModShift})
	assert.Equal(t, []sortKey{{Column: RepositoryColumn}, {Column: BranchColumn}}, m.keys)
	m.move("G")
	assert.Equal(t, 79, m.anchor)
	m.column = ChangesColumn
	m.sort(false)
	assert.Equal(t, 79, m.visible[m.selection.Index], "sorting preserves the selected checkout")
	assert.Equal(t, 0, m.selection.Index, "sort covers rows beyond the previous viewport")
	m.search.SetValue("repo-00")
	m.rebuild()
	require.Len(t, m.visible, 1)
	assert.Equal(t, 79, m.anchor, "hidden selection is remembered")
	m.search.SetValue("")
	m.rebuild()
	assert.Equal(t, 79, m.visible[m.selection.Index])
	m.Update(tea.WindowSizeMsg{Width: 35, Height: 12})
	m.focusColumn(1)
	assert.Positive(t, m.selection.Left)
	for _, line := range strings.Split(m.View().Content, "\n") {
		assert.LessOrEqual(t, ansi.StringWidth(line), 35)
	}
	assert.LessOrEqual(t, len(strings.Split(m.View().Content, "\n")), 12)
}

func TestScreenStylesEmptyAndControlCharacters(t *testing.T) {
	m := newScreen([]Row{{Repository: "evil\x1b[2J\nname", Path: "path\x1b", Branch: "界面", Error: "failure\r\n"}})
	m.resize(110, 20)
	m.sort(false)
	view := m.View()
	assert.True(t, view.AltScreen)
	assert.Equal(t, tea.MouseModeCellMotion, view.MouseMode)
	plain := ansi.Strip(view.Content)
	assert.Contains(t, plain, "Repository ↑1")
	assert.Contains(t, plain, "╭")
	assert.Contains(t, plain, "evil�[2J�name")
	assert.NotContains(t, view.Content, "evil\x1b")
	m.search.SetValue("does-not-match")
	m.rebuild()
	assert.Contains(t, ansi.Strip(m.View().Content), "No matches")
	m.resize(1, 1)
	assert.LessOrEqual(t, ansi.StringWidth(m.View().Content), 1)
}
