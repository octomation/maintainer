package contribution_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
)

func TestHistogram(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{Transport: calendarTransport{}})

	tests := map[string]struct {
		args     []string
		expected string
	}{
		"a month": {
			args: []string{"2025-01"},
			expected: "" +
				"Contributions  Days  Distribution\n" +
				"14                1  #\n" +
				"15               30  ##############################\n",
		},
		// the week from 2024-12-29 has three days before the January of the calendar
		"a week with zero": {
			args: []string{"--with-zero", "2025-01-03"},
			expected: "" +
				"Contributions  Days  Distribution\n" +
				"0                 3  ###\n" +
				"14                1  #\n" +
				"15                3  ###\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			cmd := Histogram(&cobra.Command{Use: "test"}, &config.Tool{Token: "test-token"})
			cmd.SetArgs(test.args)
			cmd.SetOut(&out)
			cmd.SetErr(&diagnostic)
			require.NoError(t, cmd.ExecuteContext(ctx))
			assert.Equal(t, test.expected, out.String())
			assert.Empty(t, diagnostic.String())
		})
	}
}

func TestHistogramView(t *testing.T) {
	type row = contribution.HistogramByCountRow
	const header = "Contributions  Days  Distribution\n"

	tests := map[string]struct {
		rows     []row
		zero     bool
		expected string
	}{
		"empty": {
			expected: header,
		},
		"empty with zero": {
			zero:     true,
			expected: header,
		},
		"only zero": {
			rows:     []row{{Count: 0, Frequency: 365}},
			expected: header,
		},
		"only zero with zero": {
			rows:     []row{{Count: 0, Frequency: 365}},
			zero:     true,
			expected: header + "0               365  " + strings.Repeat("#", 40) + "\n",
		},
		"a # per day up to the limit": {
			rows: []row{{Count: 1, Frequency: 4}, {Count: 3, Frequency: 1}, {Count: 7, Frequency: 40}},
			expected: header +
				"1                 4  ####\n" +
				"3                 1  #\n" +
				"7                40  " + strings.Repeat("#", 40) + "\n",
		},
		"skewed": {
			rows: []row{
				{Count: 0, Frequency: 200},
				{Count: 1, Frequency: 2},
				{Count: 5, Frequency: 152},
				{Count: 6, Frequency: 76},
				{Count: 1234, Frequency: 1},
			},
			expected: header +
				"1                 2  #\n" +
				"5               152  " + strings.Repeat("#", 40) + "\n" +
				"6                76  " + strings.Repeat("#", 20) + "\n" +
				"1234              1  #\n",
		},
		"skewed with zero": {
			rows: []row{
				{Count: 0, Frequency: 200},
				{Count: 1, Frequency: 2},
				{Count: 5, Frequency: 152},
				{Count: 6, Frequency: 76},
			},
			zero: true,
			expected: header +
				"0               200  " + strings.Repeat("#", 40) + "\n" +
				"1                 2  #\n" +
				"5               152  " + strings.Repeat("#", 30) + "\n" +
				"6                76  " + strings.Repeat("#", 15) + "\n",
		},
		"rounded to the nearest": {
			rows: []row{{Count: 1, Frequency: 80}, {Count: 2, Frequency: 3}, {Count: 3, Frequency: 59}},
			expected: header +
				"1                80  " + strings.Repeat("#", 40) + "\n" +
				"2                 3  ##\n" +
				"3                59  " + strings.Repeat("#", 30) + "\n",
		},
		"order is kept": {
			rows: []row{{Count: 7, Frequency: 1}, {Count: 0, Frequency: 2}, {Count: 3, Frequency: 3}},
			zero: true,
			expected: header +
				"7                 1  #\n" +
				"0                 2  ##\n" +
				"3                 3  ###\n",
		},
		"wide columns": {
			rows: []row{{Count: 12345678901234, Frequency: 12345}},
			expected: "" +
				"Contributions    Days  Distribution\n" +
				"12345678901234  12345  " + strings.Repeat("#", 40) + "\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			HistogramView(&buf, test.rows, test.zero)
			assert.Equal(t, test.expected, buf.String())
			for _, line := range strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")[1:] {
				bar := strings.TrimLeft(line[strings.LastIndex(line, "  ")+2:], " ")
				assert.LessOrEqual(t, len(bar), HistogramBar, line)
				assert.NotEmpty(t, bar, line)
			}
		})
	}
}

func TestHistogramScope(t *testing.T) {
	tests := []struct {
		args     []string
		from, to string // no from means an error
	}{
		// the current week
		{args: nil, from: "2026-09-20", to: "now"},
		// a year
		{args: []string{"2021"}, from: "2021-01-01", to: "2021-12-31"},
		{args: []string{"2026"}, from: "2026-01-01", to: "now"},
		{args: []string{"1970"}, from: "1970-01-01", to: "1970-12-31"},
		// a month
		{args: []string{"2021-02"}, from: "2021-02-01", to: "2021-02-28"},
		{args: []string{"2021-12"}, from: "2021-12-01", to: "2021-12-31"},
		{args: []string{"2026-09"}, from: "2026-09-01", to: "now"},
		// the week of a day
		{args: []string{"2021-02-10"}, from: "2021-02-07", to: "2021-02-13"},
		{args: []string{"2022-01-02"}, from: "2022-01-02", to: "2022-01-08"},
		{args: []string{"2021-12-31"}, from: "2021-12-26", to: "2022-01-01"},
		{args: []string{"2026-09-25"}, from: "2026-09-20", to: "now"},
		{args: []string{"2026-09-26"}, from: "2026-09-20", to: "now"},
		// the future
		{args: []string{"2027"}},
		{args: []string{"2030"}},
		{args: []string{"2026-10"}},
		{args: []string{"2026-09-27"}},
		{args: []string{"2030-01-01"}},
		// out of the grammar
		{args: []string{""}},
		{args: []string{"now"}},
		{args: []string{"git"}},
		{args: []string{"2021/3"}},
		{args: []string{"2021-02-01T00:00:00Z"}},
	}
	for _, input := range grammar {
		tests = append(tests, struct {
			args     []string
			from, to string
		}{args: []string{input}})
	}

	for _, test := range tests {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			scope, err := HistogramScope(test.args, now)
			if test.from == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			sane(t, scope, now, false)
			expect(t, scope, now, test.from, test.to)
		})
	}
}
