package status

import (
	"context"
	"os"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Interactive delegates terminal lifecycle, input decoding and resize to Bubble
// Tea. The snapshot and all sort/filter state remain local to this invocation.
func Interactive(ctx context.Context, in, out *os.File, rows []Row) error {
	_, err := tea.NewProgram(newScreen(rows), tea.WithContext(ctx), tea.WithInput(in), tea.WithOutput(out)).Run()
	return err
}

type screen struct {
	rows          []Row
	visible       []int
	keys          []sortKey
	column        Column
	selection     Selection
	anchor        int
	search        textinput.Model
	width, height int
	widths        []int
	dark          bool
	quitting      bool
}

// Keep the input reader alive briefly after rendering mouse reporting off.
// Bubble Tea stops its reader before restoring terminal modes on Quit, so an
// immediate quit can leave in-flight mouse reports for the parent shell.
const exitDrainTime = 150 * time.Millisecond

func (m *screen) quit() (tea.Model, tea.Cmd) {
	m.quitting = true
	m.search.Blur()
	return m, tea.Tick(exitDrainTime, func(time.Time) tea.Msg { return tea.QuitMsg{} })
}

func newScreen(rows []Row) *screen {
	input := textinput.New()
	input.Prompt = "/ "
	input.Placeholder = "Fuzzy filter: repository, branch, path, status…"
	input.CharLimit = 256
	m := &screen{rows: slices.Clone(rows), search: input, width: 100, height: 24, anchor: -1, dark: true}
	m.widths = newTable(rows).widths
	for i := range m.widths {
		m.widths[i] = max(m.widths[i], len(columnNames[i])+7)
	}
	m.widths[0] += 2 // reserve an always-visible selection marker, including NO_COLOR
	m.resize(100, 24)
	m.rebuild()
	if len(m.visible) > 0 {
		m.anchor = m.visible[0]
	}
	return m
}

func (m *screen) Init() tea.Cmd { return tea.RequestBackgroundColor }

func (m *screen) pageSize() int { return max(1, m.height-11) }

func (m *screen) resize(width, height int) {
	m.width, m.height = max(1, width), max(1, height)
	m.search.SetWidth(max(1, width-4))
	m.selection.Move("", len(m.visible), m.pageSize())
	m.selection.Left = min(m.selection.Left, max(0, m.tableWidth()-max(1, m.width-2)))
}

func (m *screen) rebuild() {
	m.visible = queryRows(m.rows, m.keys, m.search.Value())
	m.selection.Index = max(0, slices.Index(m.visible, m.anchor))
	m.selection.Move("", len(m.visible), m.pageSize())
}

func (m *screen) move(key string) {
	m.selection.Move(key, len(m.visible), m.pageSize())
	if len(m.visible) > 0 {
		m.anchor = m.visible[m.selection.Index]
	}
	m.resize(m.width, m.height)
}

func (m *screen) sort(additive bool) {
	m.keys = cycleSort(m.keys, m.column, additive)
	m.rebuild()
}

func (m *screen) focusColumn(delta int) {
	m.column = Column((int(m.column) + delta + len(columnNames)) % len(columnNames))
	start := 0
	for i := 0; i < int(m.column); i++ {
		start += m.widths[i] + 3
	}
	end := start + m.widths[m.column]
	available := max(1, m.width-2)
	if start < m.selection.Left {
		m.selection.Left = start
	}
	if end > m.selection.Left+available {
		m.selection.Left = max(start, end-available)
	}
	m.resize(m.width, m.height)
}

func (m *screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.quitting {
		// Consume trailing keys, paste and mouse reports without changing the
		// snapshot or scheduling more commands during the drain window.
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.search.SetStyles(textinput.DefaultStyles(m.dark))
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m.quit()
		}
		if m.search.Focused() {
			switch key {
			case "enter", "esc":
				m.search.Blur()
				return m, nil
			case "ctrl+u":
				m.search.SetValue("")
				m.rebuild()
				return m, nil
			}
			before := m.search.Value()
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			clean := safeText(m.search.Value())
			if clean != m.search.Value() {
				m.search.SetValue(clean)
			}
			if m.search.Value() != before {
				m.rebuild()
			}
			return m, cmd
		}
		switch key {
		case "q", "ctrl+d":
			return m.quit()
		case "esc":
			if m.search.Value() != "" {
				m.search.SetValue("")
				m.rebuild()
			} else {
				return m.quit()
			}
		case "/":
			return m, m.search.Focus()
		case "tab", "right":
			m.focusColumn(1)
		case "shift+tab", "left":
			m.focusColumn(-1)
		case "0":
			m.keys = nil
			m.rebuild()
		case "s", "S", "shift+s", "enter", "shift+enter":
			m.sort(msg.Mod.Contains(tea.ModShift) || key == "S" || strings.HasPrefix(key, "shift+"))
		case "up", "down", "j", "k", "pgup", "pgdown", "home", "end", "g", "G", "h", "l":
			m.move(key)
		}
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			break
		}
		if msg.Y == 2 {
			return m, m.search.Focus()
		}
		if msg.Y == 4 && msg.X > 0 && msg.X < m.width-1 {
			m.search.Blur()
			x, start := msg.X-1+m.selection.Left, 0
			for i, width := range m.widths {
				if x >= start && x < start+width {
					m.column = Column(i)
					m.sort(msg.Mod.Contains(tea.ModShift))
					break
				}
				start += width + 3
			}
		} else if msg.Y >= 6 && msg.Y < 6+m.pageSize() && msg.X > 0 && msg.X < m.width-1 {
			m.search.Blur()
			index := m.selection.Top + msg.Y - 6
			if index < len(m.visible) {
				m.selection.Index = index
				m.anchor = m.visible[index]
			}
		}
	case tea.MouseWheelMsg:
		if msg.Button == tea.MouseWheelUp {
			m.move("up")
		}
		if msg.Button == tea.MouseWheelDown {
			m.move("down")
		}
	default:
		before := m.search.Value()
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		if m.search.Value() != before {
			m.search.SetValue(safeText(m.search.Value()))
			m.rebuild()
		}
		return m, cmd
	}
	return m, nil
}
