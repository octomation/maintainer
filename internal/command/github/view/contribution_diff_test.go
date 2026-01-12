package view_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/github/view"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestContributionDiff(t *testing.T) {
	Nov2013 := xtime.UTC().Year(2013).Month(time.November)

	tests := []struct {
		name     string
		diff     []contribution.DayDiff
		baseOnly []time.Time
		headOnly []time.Time
		expected []string
	}{
		{
			name: "changed days",
			diff: []contribution.DayDiff{
				{Day: Nov2013.Day(13).Time(), Before: 1, After: 5, InBase: true, InHead: true},
				{Day: Nov2013.Day(14).Time(), Before: 5, After: 2, InBase: true, InHead: true},
				{Day: Nov2013.Day(15).Time(), Before: 0, After: 12, InBase: true, InHead: true},
				{Day: Nov2013.Day(16).Time(), Before: 3, InBase: true},
				{Day: Nov2013.Day(17).Time(), After: 4, InHead: true},
			},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2013-11-13     1        5      +4",
				" 2013-11-14     5        2      -3",
				" 2013-11-15     0       12     +12",
				" 2013-11-16     3        -      -3",
				" 2013-11-17     -        4      +4",
				`The diff between base{"file:before.json"} and head{"file:after.json"}`,
			},
		},
		{
			name: "no diff",
			diff: nil,
			expected: []string{
				`There is no diff between base{"file:before.json"} and head{"file:after.json"}`,
			},
		},
		{
			name:     "head covers extra days without contributions",
			headOnly: []time.Time{Nov2013.Day(14).Time(), Nov2013.Day(15).Time(), Nov2013.Day(17).Time()},
			expected: []string{
				`There is no diff between base{"file:before.json"} and head{"file:after.json"}`,
				"head also has 3 days that base lacks, 2013-11-14…2013-11-17, without contributions",
			},
		},
		{
			name: "head covers extra days with contributions",
			diff: []contribution.DayDiff{
				{Day: Nov2013.Day(13).Time(), Before: 1, After: 5, InBase: true, InHead: true},
				{Day: Nov2013.Day(15).Time(), After: 4, InHead: true},
				{Day: Nov2013.Day(16).Time(), After: 2, InHead: true},
			},
			headOnly: []time.Time{Nov2013.Day(14).Time(), Nov2013.Day(15).Time(), Nov2013.Day(16).Time(), Nov2013.Day(17).Time()},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2013-11-13     1        5      +4",
				" 2013-11-15     -        4      +4",
				" 2013-11-16     -        2      +2",
				`The diff between base{"file:before.json"} and head{"file:after.json"}`,
				"head also has 4 days that base lacks, 2013-11-14…2013-11-17, 2 with contributions listed above",
			},
		},
		{
			name: "base covers extra days",
			diff: []contribution.DayDiff{
				{Day: Nov2013.Day(20).Time(), Before: 3, InBase: true},
			},
			baseOnly: []time.Time{Nov2013.Day(18).Time(), Nov2013.Day(20).Time()},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2013-11-20     3        -      -3",
				`The diff between base{"file:before.json"} and head{"file:after.json"}`,
				"base also has 2 days that head lacks, 2013-11-18…2013-11-20, 1 with contributions listed above",
			},
		},
		{
			name: "disjoint years",
			diff: []contribution.DayDiff{
				{Day: xtime.UTC().Year(2012).Month(time.March).Day(1).Time(), Before: 2, InBase: true},
			},
			baseOnly: []time.Time{
				xtime.UTC().Year(2012).Month(time.January).Day(1).Time(),
				xtime.UTC().Year(2012).Month(time.March).Day(1).Time(),
			},
			headOnly: []time.Time{
				xtime.UTC().Year(2013).Month(time.January).Day(1).Time(),
				xtime.UTC().Year(2013).Month(time.December).Day(31).Time(),
			},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2012-03-01     2        -      -2",
				`The diff between base{"file:before.json"} and head{"file:after.json"}`,
				"base also has 2 days that head lacks, 2012-01-01…2012-03-01, 1 with contributions listed above",
				"head also has 2 days that base lacks, 2013-01-01…2013-12-31, without contributions",
			},
		},
		{
			name:     "one extra day",
			headOnly: []time.Time{Nov2013.Day(25).Time()},
			expected: []string{
				`There is no diff between base{"file:before.json"} and head{"file:after.json"}`,
				"head also has 1 day that base lacks, 2013-11-25, without contributions",
			},
		},
		{
			name: "same periods",
			diff: []contribution.DayDiff{
				{Day: Nov2013.Day(13).Time(), Before: 1, After: 5, InBase: true, InHead: true},
			},
			expected: []string{
				" Day          before   after   diff",
				"------------ -------- ------- ------",
				" 2013-11-13     1        5      +4",
				`The diff between base{"file:before.json"} and head{"file:after.json"}`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := new(cobra.Command)
			cmd.SetOut(&buf)

			require.NoError(t, ContributionDiff(cmd, test.diff, test.baseOnly, test.headOnly, "file:before.json", "file:after.json"))

			// the table pads its cells, the trailing spaces are invisible
			lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			for i := range lines {
				lines[i] = strings.TrimRight(lines[i], " ")
			}
			assert.Equal(t, test.expected, lines)
		})
	}
}
