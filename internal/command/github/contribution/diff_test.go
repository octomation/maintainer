package contribution_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/config"
)

func TestDiff(t *testing.T) {
	fs := afero.NewMemMapFs()
	for name, snapshot := range map[string]string{
		// 2022-01-02 is Sunday, 2022-01-03 is Monday
		"a.json": `{"2021-12-31T00:00:00Z":2,"2022-01-01T00:00:00Z":1,"2022-01-02T00:00:00Z":5,"2022-01-03T00:00:00Z":3}`,
		"b.json": `{"2022-01-01T00:00:00Z":1,"2022-01-02T00:00:00Z":2,"2022-01-03T00:00:00Z":7,"2022-01-04T00:00:00Z":4}`,
		"d.json": `{"2022-01-01T00:00:00Z":1,"2022-01-02T00:00:00Z":5,"2022-01-03T00:00:00Z":0,"2022-01-04T00:00:00Z":0}`,
		"c.json": `{"2021-12-31T00:00:00Z":2,"2022-01-01T00:00:00Z":1,"2022-01-02T00:00:00Z":5,"2022-01-03T00:00:00Z":3}`,
	} {
		require.NoError(t, afero.WriteFile(fs, name, []byte(snapshot), 0644))
	}

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name: "base then head",
			args: []string{"a.json", "b.json"},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2021-12-31     2        -      -2",
				" 2022-01-02     5        2      -3",
				" 2022-01-03     3        7      +4",
				" 2022-01-04     -        4      +4",
				`The diff between base{"file:a.json"} and head{"file:b.json"}`,
				"base also has 1 day that head lacks, 2021-12-31, 1 with contributions listed above",
				"head also has 1 day that base lacks, 2022-01-04, 1 with contributions listed above",
			},
		},
		{
			name: "swapped sources flip the sign",
			args: []string{"b.json", "a.json"},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2021-12-31     -        2      +2",
				" 2022-01-02     2        5      +3",
				" 2022-01-03     7        3      -4",
				" 2022-01-04     4        -      -4",
				`The diff between base{"file:b.json"} and head{"file:a.json"}`,
				"base also has 1 day that head lacks, 2022-01-04, 1 with contributions listed above",
				"head also has 1 day that base lacks, 2021-12-31, 1 with contributions listed above",
			},
		},
		{
			name: "different periods",
			args: []string{"d.json", "c.json"},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2021-12-31     -        2      +2",
				" 2022-01-03     0        3      +3",
				`The diff between base{"file:d.json"} and head{"file:c.json"}`,
				"base also has 1 day that head lacks, 2022-01-04, without contributions",
				"head also has 1 day that base lacks, 2021-12-31, 1 with contributions listed above",
			},
		},
		{
			name: "identical snapshots",
			args: []string{"a.json", "c.json"},
			expected: []string{
				`There is no diff between base{"file:a.json"} and head{"file:c.json"}`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := Diff(&cobra.Command{Use: "diff"}, &config.Tool{FS: fs})
			cmd.SetArgs(test.args)
			cmd.SetOut(&buf)

			require.NoError(t, cmd.ExecuteContext(offline(t)))

			// the table pads its cells, the trailing spaces are invisible
			lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			for i := range lines {
				lines[i] = strings.TrimRight(lines[i], " ")
			}
			assert.Equal(t, test.expected, lines)
		})
	}
}

// offline returns a context whose HTTP client fails the test on any request,
// so a command runs without a token and without GitHub.
func offline(t testing.TB) context.Context {
	return context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: forbidden{t},
	})
}

type forbidden struct{ t testing.TB }

func (rt forbidden) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.t.Errorf("unexpected request: %s %s", req.Method, req.URL)
	return nil, fmt.Errorf("unexpected request to %s", req.URL)
}

// TestCommands_NoToken runs the commands that need GitHub data without
// a token: they fail before any request and name both ways to pass it.
func TestCommands_NoToken(t *testing.T) {
	type command = func(*cobra.Command, *config.Tool) *cobra.Command

	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "a.json", []byte(`{"2022-01-01T00:00:00Z":1}`), 0644))

	tests := map[string]struct {
		command command
		args    []string
	}{
		"lookup":             {Lookup, []string{"2022/+10"}},
		"lookup by default":  {Lookup, nil},
		"suggest":            {Suggest, []string{"2022-01-05/1"}},
		"histogram":          {Histogram, []string{"2021"}},
		"snapshot":           {Snapshot, []string{"2021"}},
		"diff of a year":     {Diff, []string{"a.json", "2026"}},
		"diff of two years":  {Diff, []string{"2025", "2026"}},
		"diff with the base": {Diff, []string{"2026", "a.json"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := test.command(&cobra.Command{Use: "test", SilenceUsage: true}, &config.Tool{FS: fs})
			cmd.SetArgs(test.args)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			err := cmd.ExecuteContext(offline(t))
			require.ErrorIs(t, err, ErrNoToken)
			assert.Contains(t, err.Error(), "GITHUB_TOKEN")
			assert.Contains(t, err.Error(), "--token")
			assert.Empty(t, stdout.String())
		})
	}
}
