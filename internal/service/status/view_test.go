package status

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViews(t *testing.T) {
	rows := []Row{{Repository: "acme/tool", Branch: "main", DefaultBranch: "main", Added: 12, Deleted: 1, Untracked: 3, Ahead: 1, Behind: 2, Status: "diverged"},
		{Repository: "evil\x1b[2J\nname", Path: "path\x1b", Error: "failure\r\n"}}
	var out bytes.Buffer
	require.NoError(t, Plain(&out, rows))
	assert.Contains(t, out.String(), "main*")
	assert.Contains(t, out.String(), "+12/-1")
	assert.Contains(t, out.String(), "ahead 1 · behind 2")
	assert.Contains(t, out.String(), "?3")
	assert.NotContains(t, out.String(), "\x1b")
	sel := &Selection{Index: 1}
	f := frame(newTable(rows), rows, sel, 80, 12)
	assert.Contains(t, f, "\x1b[7m› evil�[2J�name")
	assert.NotContains(t, f, "evil\x1b")
	for _, line := range strings.Split(frame(newTable(rows), rows, sel, 4, 3), "\r\n") {
		assert.NotContains(t, line, "acme/tool")
	}
	assert.Equal(t, "界", window("a界b", 1, 2))
	assert.Equal(t, "", window("界", 0, 1))
}

func TestSelectionAndKeys(t *testing.T) {
	s := new(Selection)
	s.Move("G", 158, 20)
	assert.Equal(t, 157, s.Index)
	assert.Equal(t, 138, s.Top)
	s.Move("\x1b[5~", 158, 20)
	assert.Equal(t, 137, s.Index)
	s.Move("g", 158, 20)
	assert.Equal(t, 0, s.Index)
	assert.Equal(t, 0, s.Top)
	s.Move("\x1b[B", 158, 20)
	assert.Equal(t, 1, s.Index)
	s.Move("l", 158, 20)
	assert.Equal(t, 8, s.Left)
	s.Move("h", 158, 20)
	assert.Zero(t, s.Left)
	s.Move("G", 0, 0)
	assert.Zero(t, s.Index)
	key, rest, ok := nextKey("\x1b[6~q")
	assert.True(t, ok)
	assert.Equal(t, "\x1b[6~", key)
	assert.Equal(t, "q", rest)
	_, _, ok = nextKey("\x1b[")
	assert.False(t, ok)
}
