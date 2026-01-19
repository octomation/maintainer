package contribution_test

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
	"go.octolab.org/toolset/maintainer/internal/pkg/time/jitter"
)

func TestSuggestScope(t *testing.T) {
	tests := []struct {
		args []string
		from string // the data always lasts up to now; no from means an error
		text string // what the error tells
	}{
		// the anchor, the date of HEAD
		{args: nil, from: "2026-08-16"},
		{args: []string{""}, from: "2026-08-16"},
		{args: []string{"/3"}, from: "2026-08-23"},
		{args: []string{"/+3"}, from: "2026-08-30"},
		{args: []string{"/-3"}, from: "2026-08-09"},
		{args: []string{"git"}, from: "2026-08-16"},
		{args: []string{"git/3"}, from: "2026-08-23"},
		{args: []string{"git/+3"}, from: "2026-08-30"},
		{args: []string{"git/-3"}, from: "2026-08-09"},
		// now
		{args: []string{"now"}, from: "2026-09-06"},
		{args: []string{"now/3"}, from: "2026-09-13"},
		{args: []string{"now/+3"}, from: "2026-09-20"},
		{args: []string{"now/-3"}, from: "2026-08-30"},
		// a year, Friday
		{args: []string{"2021"}, from: "2020-12-13"},
		{args: []string{"2021/3"}, from: "2020-12-20"},
		{args: []string{"2021/+3"}, from: "2020-12-27"},
		{args: []string{"2021/-3"}, from: "2020-12-06"},
		// a month, Saturday
		{args: []string{"2022-01"}, from: "2021-12-12"},
		{args: []string{"2022-01/3"}, from: "2021-12-19"},
		{args: []string{"2022-01/+3"}, from: "2021-12-26"},
		{args: []string{"2022-01/-3"}, from: "2021-12-05"},
		// a day, Sunday
		{args: []string{"2022-01-02"}, from: "2021-12-19"},
		{args: []string{"2022-01-02/3"}, from: "2021-12-26"},
		{args: []string{"2022-01-02/+3"}, from: "2022-01-02"},
		{args: []string{"2022-01-02/-3"}, from: "2021-12-12"},
		// a day at the boundary of years
		{args: []string{"2021-12-31"}, from: "2021-12-12"},
		{args: []string{"2021-12-31/1"}, from: "2021-12-26"},
		{args: []string{"2022-01-05/1"}, from: "2022-01-02"},
		// RFC 3339, Sunday
		{args: []string{"2022-07-24T08:40:56Z"}, from: "2022-07-10"},
		{args: []string{"2022-07-24T08:40:56Z/3"}, from: "2022-07-17"},
		{args: []string{"2022-07-24T08:40:56Z/+3"}, from: "2022-07-24"},
		{args: []string{"2022-07-24T08:40:56Z/-3"}, from: "2022-07-03"},
		// RFC 3339 with an offset, Sunday there, but Saturday in UTC
		{args: []string{"2022-07-24T02:40:56+03:00"}, from: "2022-07-03"},
		{args: []string{"2022-07-24T02:40:56+03:00/3"}, from: "2022-07-10"},
		{args: []string{"2022-07-24T02:40:56+03:00/+3"}, from: "2022-07-17"},
		{args: []string{"2022-07-24T02:40:56+03:00/-3"}, from: "2022-06-26"},
		// the past and the future of now
		{args: []string{"2026-09-25T11:34:56Z"}, from: "2026-09-06"},
		{args: []string{"2026-09-25T12:34:56Z"}, from: "2026-09-06"},
		{args: []string{"2026-09-25T12:34:57Z"}, text: "2026-09-25T12:34:57Z"},
		{args: []string{"2026-09-25T13:34:56Z"}, text: "2026-09-25T13:34:56Z"},
		{args: []string{"2026-09-26"}, text: "2026-09-26T00:00:00Z"},
		{args: []string{"2026-09-26/-3"}, text: "2026-09-26T00:00:00Z"},
		{args: []string{"2027-09-25T15:34:56+03:00"}, text: "2027-09-25T15:34:56+03:00"},
		{args: []string{"2026-12-01/3"}, text: "2026-12-01T00:00:00Z"},
		{args: []string{"2030"}, text: "2030-01-01T00:00:00Z"},
		{args: []string{"2030-01-01"}, text: "2030-01-01T00:00:00Z"},
	}
	for _, input := range grammar {
		tests = append(tests, struct {
			args []string
			from string
			text string
		}{args: []string{input}})
	}

	for _, test := range tests {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			opts, scope, err := SuggestScope(test.args, anchor, now)
			if test.from == "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.text)
				return
			}
			require.NoError(t, err)
			sane(t, scope, now, true)
			expect(t, scope, now, test.from, "now")
			assert.False(t, opts.Value.After(now), "the anchor is in the future")
			assert.True(t, contribution.LookupRange(opts).From().Equal(scope.From()))
		})
	}

	t.Run("the week starts now", func(t *testing.T) {
		sunday := xtime.UTC().Year(2026).Month(time.September).Day(20).Time()
		_, _, err := SuggestScope([]string{"now/+1"}, anchor, sunday)
		assert.ErrorIs(t, err, xtime.ErrFuturePeriod)
	})
}

// TestSuggestScope_Anchor covers issue#133 and issue#148: HEAD dated after
// now leaves nothing to suggest, and it is an error before any request.
func TestSuggestScope_Anchor(t *testing.T) {
	tests := map[string]struct {
		head  time.Time
		valid bool
	}{
		"HEAD in the past":     {head: now.Add(-time.Hour), valid: true},
		"HEAD at now":          {head: now, valid: true},
		"HEAD an hour ahead":   {head: now.Add(time.Hour)},
		"HEAD a day ahead":     {head: now.AddDate(0, 0, 1)},
		"HEAD a year ahead":    {head: now.AddDate(1, 0, 0)},
		"HEAD a second ahead":  {head: now.Add(time.Second)},
		"HEAD ahead in UTC":    {head: now.Add(time.Hour).UTC()},
		"HEAD of the last day": {head: now.AddDate(0, 0, -1), valid: true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("GIT_DIR", "")
			t.Chdir(repository(t, test.head))

			head, err := FallbackDate([]string{"git/3"}, now)
			require.NoError(t, err)
			require.True(t, test.head.Equal(head))

			opts, scope, err := SuggestScope([]string{"git/3"}, head, now)
			if !test.valid {
				require.Error(t, err)
				assert.Contains(t, err.Error(), head.Format(time.RFC3339))
				return
			}
			require.NoError(t, err)
			assert.True(t, head.Equal(opts.Value))
			assert.True(t, scope.To().Equal(now))
		})
	}
}

// TestSuggestScope_Data covers the data of a suggestion: HEAD is weeks
// before now, so the lookup around it ends long before now, but any day
// up to now may be suggested, and a day without data looks like a gap.
func TestSuggestScope_Data(t *testing.T) {
	head := xtime.UTC().Year(2026).Month(time.August).Day(26).Hour(10).Time() // Wednesday
	opts, scope, err := SuggestScope([]string{"git/3"}, head, now)
	require.NoError(t, err)

	lookup := contribution.LookupRange(opts)
	expect(t, lookup, now, "2026-08-16", "2026-09-05")
	expect(t, scope, now, "2026-08-16", "now")

	// every day is on target, except the only gap after the lookup
	gap := xtime.UTC().Year(2026).Month(time.September).Day(8).Time() // Tuesday
	upstream := make(contribution.HeatMap)
	for day := xtime.UTC().Year(2026).Month(time.August).Day(9).Time(); day.Before(now); day = day.Add(xtime.Day) {
		upstream.SetCount(day, 5)
	}
	upstream.SetCount(gap, 2)

	schedule, target := xtime.Everyday(xtime.Hours(5, 19, 0)), uint(5)
	suggestion := contribution.Suggest(upstream.Subset(scope), opts.Value, now, schedule, target)
	assert.Equal(t, contribution.Suggestion{Time: gap.Add(5 * time.Hour), Actual: 2, Target: 5}, suggestion)

	// the data of the lookup alone would suggest the next day after it,
	// which has contributions
	suggestion = contribution.Suggest(upstream.Subset(lookup), opts.Value, now, schedule, target)
	assert.Equal(t, xtime.UTC().Year(2026).Month(time.September).Day(6).Hour(5).Time(), suggestion.Time)
	assert.Equal(t, uint(5), upstream.Count(xtime.TruncateToDay(suggestion.Time)))
}

func TestJitter(t *testing.T) {
	day := xtime.UTC().Year(2026).Month(time.September).Day(25)
	hours := xtime.Everyday(xtime.Hours(5, 19, 0))
	upmost := jitter.Transformation(func(d time.Duration) time.Duration { return d - 1 })
	random := jitter.FullCustom(rand.New(rand.NewSource(1)))

	tests := map[string]struct {
		time, now time.Time
		limit     time.Time // the latest jitter is right before it
	}{
		"within an hour": {
			time:  day.Hour(10).Time(),
			now:   day.Hour(23).Time(),
			limit: day.Hour(11).Time(),
		},
		"up to now": {
			time:  day.Hour(10).Time(),
			now:   day.Hour(10).Minute(20).Time(),
			limit: day.Hour(10).Minute(20).Time(),
		},
		"up to the end of the schedule": {
			time:  day.Hour(18).Minute(50).Time(),
			now:   day.Day(26).Hour(12).Time(),
			limit: day.Hour(19).Time(),
		},
		"no room before now": {
			time:  day.Hour(10).Time(),
			now:   day.Hour(10).Time(),
			limit: day.Hour(10).Time(),
		},
		"no room in the schedule": {
			time:  day.Hour(19).Time(),
			now:   day.Hour(23).Time(),
			limit: day.Hour(19).Time(),
		},
		"outside the schedule": {
			time:  day.Hour(20).Time(),
			now:   day.Hour(23).Time(),
			limit: day.Hour(20).Time(),
		},
		"after now": {
			time:  day.Hour(10).Time(),
			now:   day.Hour(9).Time(),
			limit: day.Hour(10).Time(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			expected := test.time
			if test.limit.After(test.time) {
				expected = test.limit.Add(-time.Nanosecond)
			}
			assert.Equal(t, expected, Jitter(test.time, test.now, hours, upmost))

			for range 1000 {
				moment := Jitter(test.time, test.now, hours, random)
				assert.False(t, moment.Before(test.time), "before the suggestion")
				assert.False(t, moment.After(expected), "after the limit")
			}
		})
	}

	t.Run("no random without room", func(t *testing.T) {
		broken := jitter.Transformation(func(time.Duration) time.Duration {
			t.Fatal("random is not needed")
			return 0
		})
		assert.Equal(t, day.Hour(10).Time(), Jitter(day.Hour(10).Time(), day.Hour(10).Time(), hours, broken))
		assert.Equal(t, day.Hour(19).Time(), Jitter(day.Hour(19).Time(), day.Hour(23).Time(), hours, broken))
	})
}

// A moment is printed to the second, so fractional seconds of the anchor
// or of now must not move the printed moment out of the window or out of
// the schedule.
func TestSuggest_Printed(t *testing.T) {
	clock := xtime.UTC().Year(2026).Month(time.January).Day(14).Hour(12).Time().Add(987 * time.Millisecond)
	day := xtime.UTC().Year(2026).Month(time.January)
	schedule, target := xtime.Everyday(xtime.Hours(5, 19, 0)), uint(1000000)
	upstream := make(contribution.HeatMap)
	for ts := day.Day(1).Time(); ts.Before(clock); ts = ts.Add(xtime.Day) {
		upstream.SetCount(ts, 1)
	}

	tests := map[string]struct {
		args     []string
		head     time.Time
		earliest time.Time // the suggestion before the jitter
	}{
		"anchor at the end of the schedule": {
			args:     []string{"2026-01-12T19:00:00.1234Z/+0"},
			earliest: day.Day(13).Hour(5).Time(),
		},
		"anchor at now": {
			args:     []string{"now/+0"},
			earliest: day.Day(14).Hour(12).Time(),
		},
		"anchor at HEAD": {
			args:     []string{"git/+0"},
			head:     day.Day(13).Hour(10).Time(),
			earliest: day.Day(13).Hour(10).Time(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// the same as the command does
			now := clock.Truncate(time.Second)
			opts, scope, err := SuggestScope(test.args, test.head, now)
			require.NoError(t, err)

			suggestion := contribution.Suggest(upstream.Subset(scope), opts.Value, now, schedule, target)
			require.Equal(t, test.earliest, suggestion.Time)

			// the anchor as given, before it is rounded up
			anchor, err := time.Parse(time.RFC3339, strings.Split(test.args[0], "/")[0])
			if err != nil {
				anchor = opts.Value
			}
			random := jitter.FullCustom(rand.New(rand.NewSource(1)))
			for range 1000 {
				moment := Moment(suggestion.Time, now, schedule, random)
				printed, err := time.Parse(time.RFC3339, moment.Local().Format(time.RFC3339))
				require.NoError(t, err)

				assert.False(t, printed.Before(anchor), "before the anchor")
				assert.False(t, printed.After(now), "after now")
				assert.False(t, printed.Before(test.earliest), "before the suggestion")
				start := xtime.TruncateToDay(printed).Add(5 * time.Hour)
				assert.False(t, printed.Before(start), "before the schedule")
				assert.False(t, printed.After(start.Add(14*time.Hour)), "after the schedule")
			}
		})
	}

	t.Run("anchor rounded up after now", func(t *testing.T) {
		now := clock.Truncate(time.Second)
		_, _, err := SuggestScope([]string{"2026-01-14T12:00:00.5Z"}, time.Time{}, now)
		assert.ErrorContains(t, err, "in the future")
	})
}

// The delta is printed next to the moment, and they must agree to the second.
func TestMoment_Delta(t *testing.T) {
	day := xtime.UTC().Year(2026).Month(time.January)
	schedule := xtime.Everyday(xtime.Hours(5, 19, 0))
	now := day.Day(14).Hour(12).Time()
	fraction := jitter.Transformation(func(d time.Duration) time.Duration { return d/3 + 500*time.Millisecond })
	random := jitter.FullCustom(rand.New(rand.NewSource(1)))

	for _, suggested := range []time.Time{day.Day(13).Hour(10).Time(), day.Day(14).Hour(11).Minute(50).Time()} {
		for _, tf := range []jitter.Transformation{fraction, random} {
			for range 100 {
				moment := Moment(suggested, now, schedule, tf)
				require.Equal(t, moment, moment.Truncate(time.Second), "fractional seconds")
				assert.False(t, moment.Before(suggested), "before the suggestion")
				assert.False(t, moment.After(now), "after now")

				printed, err := time.Parse(time.RFC3339, moment.Local().Format(time.RFC3339))
				require.NoError(t, err)
				shift := now.Sub(printed)
				expected := "-" + strings.ToUpper(shift.String())
				if days := shift / xtime.Day; days > 0 {
					expected = fmt.Sprintf("-%dd%s", days, strings.ToUpper((shift % xtime.Day).String()))
				}
				assert.Equal(t, expected, Datetime(moment.Local(), now))
			}
		}
	}
}

// TestMissed covers the explanation of nothing to suggest: without it,
// a week filled up to now looked like a failure to retry, since the table
// that shows it was not printed.
func TestMissed(t *testing.T) {
	day := xtime.UTC().Year(2026).Month(time.October)
	schedule, target := xtime.Everyday(xtime.Hours(5, 19, 0)), uint(5)
	local := func(t time.Time) string { return t.Local().Format(time.RFC3339) }
	upstream := func(thursday uint) contribution.HeatMap {
		chm := make(contribution.HeatMap)
		for ts := xtime.UTC().Year(2026).Month(time.September).Day(27).Time(); ts.Before(day.Day(11).Time()); ts = ts.Add(xtime.Day) {
			chm.SetCount(ts, 25)
		}
		for ts := day.Day(11).Time(); ts.Before(day.Day(15).Time()); ts = ts.Add(xtime.Day) {
			chm.SetCount(ts, 20)
		}
		chm.SetCount(day.Day(15).Time(), thursday)
		return chm
	}

	tests := map[string]struct {
		heats  contribution.HeatMap
		args   []string
		anchor time.Time
		now    time.Time

		reason   string
		expected []string // what stderr tells
		absent   []string // what stderr does not tell
	}{
		"caught up": {
			heats:  upstream(20),
			anchor: day.Day(11).Hour(9).Minute(18).Second(9).Time(),
			now:    day.Day(15).Hour(16).Minute(50).Second(36).Time(),
			reason: "caught up",
			expected: []string{
				"Oct 15", "distribution{20: 5, 25: 14}",
				"Every day since Oct 11 has what its week requires, 20 in #42.\n",
				"The next gap opens at " + local(day.Day(16).Hour(5).Time()) + ", in +12H9M24S.\n",
				"Or raise the target: maintainer suggest --target=21\n",
			},
		},
		"caught up across weeks": {
			heats:  upstream(20),
			anchor: day.Day(6).Hour(9).Time(), // week #41 requires 25
			now:    day.Day(15).Hour(16).Minute(50).Second(36).Time(),
			reason: "caught up",
			expected: []string{
				"Every day since Oct 6 has what its week requires, 20 in #42.\n",
				"Or raise the target: maintainer suggest --target=21\n",
			},
		},
		"caught up, short": {
			heats:    upstream(20),
			args:     []string{"--short", "2026-10-11"},
			anchor:   time.Time{},
			now:      day.Day(15).Hour(16).Minute(50).Second(36).Time(),
			reason:   "caught up",
			expected: []string{"Or raise the target: maintainer suggest --short --target=21 2026-10-11\n"},
			absent:   []string{"Day / Week"},
		},
		"the schedule has not opened yet": {
			heats:  upstream(0),
			anchor: day.Day(11).Hour(9).Time(),
			now:    day.Day(15).Hour(3).Time(),
			reason: "the schedule has not opened yet",
			expected: []string{
				"Oct 15 has 0 of 20, but its schedule opens at " + local(day.Day(15).Hour(5).Time()) + ", in +2H0M0S.\n",
				"Or raise the target: maintainer suggest --target=21\n", // Sunday has 20
			},
			absent: []string{"next gap"},
		},
		"the schedule has not opened at the start of a week": {
			heats: func() contribution.HeatMap {
				chm := upstream(20)
				chm.SetCount(day.Day(16).Time(), 20)
				chm.SetCount(day.Day(17).Time(), 20)
				return chm
			}(),
			args:   []string{"2026-10-11/0"},
			now:    day.Day(18).Time(), // Sunday midnight, the week of it starts at now
			reason: "the schedule has not opened yet",
			expected: []string{
				"Oct 18 has 0 of 5, but its schedule opens at " + local(day.Day(18).Hour(5).Time()) + ", in +5H0M0S.\n",
				"Or raise the target: maintainer suggest --target=21 2026-10-11/0\n", // the week before has 20
			},
			absent: []string{"Day / Week"}, // no weeks to show up to now
		},
		"the schedule is over by the anchor": {
			heats:  upstream(19),
			anchor: day.Day(15).Hour(20).Time(),
			now:    day.Day(15).Hour(21).Time(),
			reason: "the schedule is over by the anchor",
			expected: []string{
				"Oct 15 has 19 of 20, but its schedule is over by the anchor.\n",
				"The next gap opens at " + local(day.Day(16).Hour(5).Time()) + ", in +8H0M0S.\n",
			},
			absent: []string{"raise the target"}, // no other day is in the window
		},
		"the schedule is over by the anchor, other days on target": {
			heats: func() contribution.HeatMap {
				chm := upstream(25)
				for ts := day.Day(12).Time(); ts.Before(day.Day(15).Time()); ts = ts.Add(xtime.Day) {
					chm.SetCount(ts, 25)
				}
				return chm
			}(),
			anchor: day.Day(11).Hour(20).Time(),
			now:    day.Day(15).Hour(16).Time(),
			reason: "the schedule is over by the anchor",
			expected: []string{
				"Oct 11 has 20 of 25, but its schedule is over by the anchor.\n",
				"The next gap opens at " + local(day.Day(16).Hour(5).Time()) + ", in +13H0M0S.\n",
				"Or raise the target: maintainer suggest --target=26\n", // 21 is closed on Oct 11
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			root := &cobra.Command{Use: "maintainer"}
			cmd := Suggest(&cobra.Command{Use: "suggest"}, new(config.Tool))
			root.AddCommand(cmd)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			require.NoError(t, cmd.ParseFlags(test.args))
			args := cmd.Flags().Args()
			short, err := cmd.Flags().GetBool("short")
			require.NoError(t, err)

			opts, scope, err := SuggestScope(args, test.anchor, test.now)
			require.NoError(t, err)
			chm := test.heats.Subset(scope)
			require.True(t, contribution.Suggest(chm, opts.Value, test.now, schedule, target).Time.IsZero())

			err = Missed(cmd, args, chm, opts, test.now, schedule, target, short)
			require.Error(t, err)
			assert.True(t, strings.HasSuffix(err.Error(), ": "+test.reason), err.Error())
			var coder exit.Coder
			require.True(t, errors.As(err, &coder))
			assert.Equal(t, exit.Nothing, coder.ExitCode())

			assert.Empty(t, stdout.String())
			for _, text := range test.expected {
				assert.Contains(t, stderr.String(), text)
			}
			for _, text := range test.absent {
				assert.NotContains(t, stderr.String(), text)
			}
		})
	}
}

// TestSuggest_ExitCodes covers the exit codes of input errors, which the flags
// and arguments decline before the command runs, and the command itself.
func TestSuggest_ExitCodes(t *testing.T) {
	tests := map[string][]string{
		"a malformed target": {"--target=abc"},
		"an unknown flag":    {"--unknown"},
		"two arguments":      {"2026", "2027"},
		"a zero target":      {"--target=0"},
		"a malformed date":   {"2021-02-30"},
		"an anchor ahead":    {"9999-01-01"},
		"no token":           {"2026-01-01"},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := Suggest(&cobra.Command{Use: "suggest", SilenceUsage: true}, new(config.Tool))
			cmd.SetArgs(args)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			err := cmd.Execute()
			var coder exit.Coder
			require.True(t, errors.As(err, &coder), "%v", err)
			assert.Equal(t, exit.User, coder.ExitCode())
			assert.Empty(t, stdout.String())
		})
	}
}

func TestRerun(t *testing.T) {
	tests := map[string]struct {
		args     []string
		expected string
	}{
		"no arguments":     {nil, "maintainer suggest --target=21"},
		"the target given": {[]string{"--target=3", "git/3"}, "maintainer suggest --target=21 git/3"},
		"flags kept":       {[]string{"--delta", "--short=false", "2013-11/-10"}, "maintainer suggest --delta --short=false --target=21 2013-11/-10"},
		"an empty date":    {[]string{""}, "maintainer suggest --target=21 ''"},
		"a quote":          {[]string{"it's"}, `maintainer suggest --target=21 'it'\''s'`},
		"a spaced date":    {[]string{"2026-10-11 10:00"}, "maintainer suggest --target=21 '2026-10-11 10:00'"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			root := &cobra.Command{Use: "maintainer"}
			root.PersistentFlags().String("token", "", "")
			cmd := Suggest(&cobra.Command{Use: "suggest"}, new(config.Tool))
			root.AddCommand(cmd)
			require.NoError(t, cmd.ParseFlags(append([]string{"--token=secret"}, test.args...)))

			assert.Equal(t, test.expected, Rerun(cmd, cmd.Flags().Args(), 21))
		})
	}
}
