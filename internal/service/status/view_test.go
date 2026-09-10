package status

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushLockViews(t *testing.T) {
	rows := []Row{{Repository: "acme/locked", Path: "/locked", PushLocked: true}, {Repository: "acme/open", Path: "/open"}}
	var out bytes.Buffer
	require.NoError(t, Plain(&out, rows))
	lines := strings.Split(out.String(), "\n")
	assert.Contains(t, lines[2], "│ LOCK │ PATH")
	assert.Contains(t, lines[4], "│ 🔒   │ /locked")
	assert.Contains(t, lines[5], "│      │ /open")
	assert.Equal(t, ansi.StringWidth(lines[4]), ansi.StringWidth(lines[5]))
	raw, err := json.Marshal(rows)
	require.NoError(t, err)
	var decoded []map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, true, decoded[0]["push_locked"])
	assert.Equal(t, false, decoded[1]["push_locked"])
}

func TestViews(t *testing.T) {
	rows := []Row{{Repository: "acme/tool", Path: "/work/public/acme/tool", displayPath: "public/acme/tool", Branch: "main", DefaultBranch: "main", Added: 12, Deleted: 1, Untracked: 3, Ahead: 1, Behind: 2, Status: "diverged"},
		{Repository: "evil\x1b[2J\nname", Path: "path\x1b", Error: "failure\r\n"}}
	var out bytes.Buffer
	require.NoError(t, Plain(&out, rows))
	assert.Contains(t, out.String(), "main*")
	assert.Contains(t, out.String(), "+12/-1")
	assert.Contains(t, out.String(), "ahead 1 · behind 2")
	assert.Contains(t, out.String(), "?3")
	assert.NotContains(t, out.String(), "\x1b")
	lines := strings.Split(out.String(), "\n")
	assert.True(t, strings.HasSuffix(strings.TrimSpace(lines[2]), "│ PATH"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(lines[4]), "│ public/acme/tool"))
	assert.NotContains(t, out.String(), "/work/")
}

func TestRelativePath(t *testing.T) {
	for _, tc := range []struct{ name, path, root, home, want string }{
		{"workspace", "/home/me/work/public/acme/tool", "/home/me/work", "/home/me", "public/acme/tool"},
		{"external pin", "/home/me/.dotfiles", "/home/me/work", "/home/me", "~/.dotfiles"},
		{"outside home", "/srv/tool", "/home/me/work", "/home/me", "/srv/tool"},
		{"root prefix", "/home/me/work-old/tool", "/home/me/work", "/home/me", "~/work-old/tool"},
		{"home prefix", "/home/merlin/tool", "/home/me/work", "/home/me", "/home/merlin/tool"},
		{"root checkout", "/home/me/work", "/home/me/work", "/home/me", "."},
		{"home checkout", "/home/me", "/home/me/work", "/home/me", "~"},
		{"unknown home", "/srv/tool", "/work", "", "/srv/tool"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, relativePath(tc.path, tc.root, tc.home))
		})
	}
}

func TestSelectionAndKeys(t *testing.T) {
	s := new(Selection)
	s.Move("G", 158, 20)
	assert.Equal(t, 157, s.Index)
	assert.Equal(t, 138, s.Top)
	s.Move("pgup", 158, 20)
	assert.Equal(t, 137, s.Index)
	s.Move("g", 158, 20)
	assert.Equal(t, 0, s.Index)
	assert.Equal(t, 0, s.Top)
	s.Move("down", 158, 20)
	assert.Equal(t, 1, s.Index)
	s.Move("l", 158, 20)
	assert.Equal(t, 8, s.Left)
	s.Move("h", 158, 20)
	assert.Zero(t, s.Left)
	s.Move("G", 0, 0)
	assert.Zero(t, s.Index)
}
