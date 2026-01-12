package contribution_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
)

func TestSnapshotScope(t *testing.T) {
	tests := []struct {
		args     []string
		from, to string // no from means an error
	}{
		// the current year
		{args: nil, from: "2026-01-01", to: "now"},
		// a year
		{args: []string{"2021"}, from: "2021-01-01", to: "2021-12-31"},
		{args: []string{"2026"}, from: "2026-01-01", to: "now"},
		{args: []string{"1970"}, from: "1970-01-01", to: "1970-12-31"},
		// the future
		{args: []string{"2027"}},
		{args: []string{"2030"}},
		// out of the grammar
		{args: []string{""}},
		{args: []string{"now"}},
		{args: []string{"abcd"}},
		{args: []string{"2021-02"}},
		{args: []string{"2021-02-10"}},
		{args: []string{"2021/1"}},
	}
	for _, input := range grammar {
		tests = append(tests, struct {
			args     []string
			from, to string
		}{args: []string{input}})
	}

	for _, test := range tests {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			scope, err := SnapshotScope(test.args, now)
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
