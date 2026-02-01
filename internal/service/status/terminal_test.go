package status

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func press(m *screen, code rune, text string, mod tea.KeyMod) tea.Cmd {
	_, cmd := m.Update(tea.KeyPressMsg{Code: code, Text: text, Mod: mod})
	return cmd
}

func TestQuitDrainsInputBeforeReturningToShell(t *testing.T) {
	for _, key := range []tea.KeyPressMsg{
		{Code: tea.KeyEscape},
		{Code: 'q', Text: "q"},
		{Code: 'c', Mod: tea.ModCtrl},
		{Code: 'd', Mod: tea.ModCtrl},
	} {
		t.Run(key.String(), func(t *testing.T) {
			m := newScreen([]Row{{Repository: "one"}, {Repository: "two"}})
			_, cmd := m.Update(key)
			require.NotNil(t, cmd)
			require.True(t, m.quitting)
			view := m.View()
			assert.True(t, view.AltScreen, "keep the TUI open while consuming trailing input")
			assert.Equal(t, tea.MouseModeNone, view.MouseMode)
			assert.True(t, view.DisableBracketedPasteMode)
			for _, trailing := range []tea.Msg{
				tea.MouseWheelMsg{Button: tea.MouseWheelDown},
				tea.KeyPressMsg{Code: '/'},
				tea.KeyPressMsg{Code: 'q'},
				tea.PasteMsg{Content: "65;65;53M"},
			} {
				_, next := m.Update(trailing)
				assert.Nil(t, next, "trailing input must not schedule commands")
			}
			assert.Zero(t, m.selection.Index)
			assert.False(t, m.search.Focused())
			assert.Empty(t, m.search.Value())
			assert.IsType(t, tea.QuitMsg{}, cmd())
		})
	}
}

func TestEscapeInSearchDoesNotStartExit(t *testing.T) {
	m := newScreen(nil)
	m.search.Focus()
	m.search.SetValue("query")
	assert.Nil(t, press(m, tea.KeyEscape, "", 0))
	assert.False(t, m.search.Focused())
	assert.Equal(t, "query", m.search.Value())
	assert.False(t, m.quitting)
	assert.Nil(t, press(m, tea.KeyEscape, "", 0))
	assert.Empty(t, m.search.Value())
	assert.False(t, m.quitting)
}

func TestInteractiveConsumesReportsAfterDisablingMouse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	in, input, err := os.Pipe()
	require.NoError(t, err)
	defer in.Close()
	defer input.Close()
	output, out, err := os.Pipe()
	require.NoError(t, err)
	defer output.Close()
	defer out.Close()

	enabled, disabled := make(chan struct{}), make(chan struct{})
	go func() {
		var seen bytes.Buffer
		buf := make([]byte, 4096)
		on, off := false, false
		for {
			n, err := output.Read(buf)
			seen.Write(buf[:n])
			if !on && strings.Contains(seen.String(), ansi.SetModeMouseExtSgr) {
				on = true
				close(enabled)
			}
			if on && !off && strings.Contains(seen.String(), ansi.ResetModeMouseExtSgr) {
				off = true
				close(disabled)
			}
			if err != nil {
				return
			}
		}
	}()
	done := make(chan error, 1)
	go func() { done <- Interactive(ctx, in, out, []Row{{Repository: "acme/tool"}}) }()
	await := func(ch <-chan struct{}) {
		t.Helper()
		select {
		case <-ch:
		case <-ctx.Done():
			t.Fatal("timed out waiting for terminal mode change")
		}
	}
	await(enabled)
	_, err = input.WriteString("\x1b")
	require.NoError(t, err)
	await(disabled)
	// Simulate reports already in flight when the terminal processes mouse-off.
	_, err = input.WriteString("\x1b[<65;65;53M\x1b[<65;65;53Mtrailing text")
	require.NoError(t, err)
	select {
	case err = <-done:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("interactive status did not exit")
	}
	require.NoError(t, input.Close())
	remaining, err := io.ReadAll(in)
	require.NoError(t, err)
	assert.Empty(t, remaining, "TUI input must not be left for the shell")
}

func TestKeyboardSortingAndFilter(t *testing.T) {
	m := newScreen([]Row{{Repository: "acme/maintainer", Path: "/one", Branch: "main"}, {Repository: "acme/dotfiles", Path: "/two", Branch: "main"}})
	press(m, tea.KeyTab, "", 0)
	assert.Equal(t, BranchColumn, m.column)
	press(m, 's', "s", 0)
	press(m, tea.KeyTab, "", 0)
	press(m, 's', "S", tea.ModShift)
	assert.Equal(t, []SortKey{{Column: BranchColumn}, {Column: ChangesColumn}}, m.keys)
	press(m, 'S', "S", 0) // legacy terminal delivers uppercase rather than a modifier
	assert.True(t, m.keys[1].Desc)
	press(m, tea.KeyEnter, "", tea.ModShift)
	assert.Equal(t, []SortKey{{Column: BranchColumn}}, m.keys)
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

func TestPathColumnNavigationAndSorting(t *testing.T) {
	m := newScreen([]Row{{Repository: "acme/tool", Path: "/work/public/acme/tool", displayPath: "public/acme/tool"}})
	m.resize(60, 24)
	press(m, tea.KeyTab, "", tea.ModShift)
	assert.Equal(t, PathColumn, m.column)
	assert.Positive(t, m.selection.Left)
	press(m, 's', "s", 0)
	assert.Equal(t, []SortKey{{Column: PathColumn}}, m.keys)
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "Path ↑1")
	assert.Contains(t, view, "public/acme/tool")
	press(m, tea.KeyTab, "", 0)
	assert.Equal(t, RepositoryColumn, m.column)
	assert.Zero(t, m.selection.Left)
}

func TestLockColumnNavigationAndSorting(t *testing.T) {
	m := newScreen([]Row{{Repository: "acme/open", Path: "/open"}, {Repository: "acme/locked", Path: "/locked", PushLocked: true}})
	m.resize(65, 16)
	press(m, tea.KeyTab, "", tea.ModShift)
	press(m, tea.KeyTab, "", tea.ModShift)
	assert.Equal(t, LockColumn, m.column)
	assert.Positive(t, m.selection.Left)
	press(m, 's', "s", 0)
	press(m, 's', "s", 0)
	assert.Equal(t, []SortKey{{Column: LockColumn, Desc: true}}, m.keys)
	assert.Equal(t, []int{1, 0}, m.visible)
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "Lock ↓1")
	assert.Contains(t, view, "🔒")
	for _, line := range strings.Split(view, "\n") {
		assert.LessOrEqual(t, ansi.StringWidth(line), 65)
	}
	x := 1 - m.selection.Left
	for i := 0; i < int(LockColumn); i++ {
		x += m.widths[i] + 3
	}
	m.Update(tea.MouseClickMsg{X: x, Y: 5, Button: tea.MouseLeft})
	assert.Empty(t, m.keys, "clicking the descending lock header removes sorting")
	press(m, tea.KeyTab, "", 0)
	assert.Equal(t, PathColumn, m.column)
}

func TestMouseSortingAndSelectionAcrossResize(t *testing.T) {
	var rows []Row
	for i := 0; i < 80; i++ {
		rows = append(rows, Row{Repository: fmt.Sprintf("repo-%02d", i), Path: fmt.Sprint(i), Added: 80 - i})
	}
	m := newScreen(rows)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	m.Update(tea.MouseClickMsg{X: 2, Y: 5, Button: tea.MouseLeft})
	assert.Equal(t, []SortKey{{Column: RepositoryColumn}}, m.keys)
	m.Update(tea.MouseClickMsg{X: m.widths[0] + 5, Y: 5, Button: tea.MouseLeft, Mod: tea.ModShift})
	assert.Equal(t, []SortKey{{Column: RepositoryColumn}, {Column: BranchColumn}}, m.keys)
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
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	assert.True(t, strings.HasPrefix(lines[6], "│─"), "separator stays immediately below the header")
	assert.Contains(t, lines[9], "/80", "wrapping must not push the selected-path footer off screen")
	assert.Contains(t, lines[10], "HEAD", "metadata remains visible after resize")
	m.keys = nil
	m.selection.Left = 10
	m.search.Focus()
	m.Update(tea.MouseClickMsg{X: m.widths[0] + 4 - 10, Y: 5, Button: tea.MouseLeft})
	assert.Equal(t, []SortKey{{Column: BranchColumn}}, m.keys, "header hit testing accounts for horizontal scroll")
	assert.False(t, m.search.Focused(), "header activation returns focus to the table")
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
	m.resize(62, 16)
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	assert.Contains(t, lines[7], "No matches", "empty message has the same row coordinate as table data")
	assert.Contains(t, lines[12], "No repository selected")
	m.resize(1, 1)
	assert.LessOrEqual(t, ansi.StringWidth(m.View().Content), 1)
}

func TestSortingUnderFilterKeepsHighlightedRow(t *testing.T) {
	m := newScreen([]Row{{Repository: "alpha", Path: "/a"}, {Repository: "beta1", Path: "/b1"}, {Repository: "beta2", Path: "/b2"}, {Repository: "gamma", Path: "/g"}})
	highlighted := func() string { return m.rows[m.visible[m.selection.Index]].Repository }
	press(m, 'G', "G", 0)
	require.Equal(t, "gamma", highlighted())
	press(m, '/', "/", 0)
	for _, r := range "beta" {
		press(m, r, string(r), 0)
	}
	press(m, tea.KeyEnter, "", 0)
	require.Len(t, m.visible, 2)
	assert.Equal(t, "beta1", highlighted())
	press(m, 's', "s", 0)
	assert.Equal(t, "beta1", highlighted())
	press(m, 's', "s", 0)
	require.Equal(t, []SortKey{{Column: RepositoryColumn, Desc: true}}, m.keys)
	assert.Equal(t, []int{2, 1}, m.visible)
	assert.Equal(t, "beta1", highlighted(), "sorting keeps the highlighted row")
	press(m, '0', "0", 0)
	assert.Equal(t, "beta1", highlighted(), "resetting the sort keeps it too")
	press(m, tea.KeyEscape, "", 0)
	require.Len(t, m.visible, 4)
	assert.Equal(t, "gamma", highlighted(), "clearing the filter restores the selection")
	// Selecting a row under the filter replaces the selection to restore.
	press(m, '/', "/", 0)
	for _, r := range "beta" {
		press(m, r, string(r), 0)
	}
	press(m, tea.KeyEnter, "", 0)
	press(m, 'j', "j", 0)
	assert.Equal(t, "beta2", highlighted())
	press(m, tea.KeyEscape, "", 0)
	assert.Equal(t, "beta2", highlighted())
}

func TestMouseCoordinatesBelowSummary(t *testing.T) {
	m := newScreen([]Row{{Repository: "one", Path: "/1"}, {Repository: "two", Path: "/2"}})
	m.resize(80, 24)
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	require.Contains(t, lines[1], "all: ")
	require.Contains(t, lines[3], "Fuzzy filter")
	require.Contains(t, lines[5], "Repository")
	require.Contains(t, lines[8], "two")
	m.Update(tea.MouseClickMsg{X: 4, Y: 8, Button: tea.MouseLeft})
	assert.Equal(t, 1, m.anchor, "a click selects the row under the pointer")
	m.Update(tea.MouseClickMsg{X: 4, Y: 5, Button: tea.MouseLeft})
	assert.Equal(t, []SortKey{{Column: RepositoryColumn}}, m.keys, "a click on the column headers sorts")
	m.Update(tea.MouseClickMsg{X: 4, Y: 3, Button: tea.MouseLeft})
	assert.True(t, m.search.Focused(), "a click on the search line focuses it")
}
