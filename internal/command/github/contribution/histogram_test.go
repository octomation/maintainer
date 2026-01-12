package contribution_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
)

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
