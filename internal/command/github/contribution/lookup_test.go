package contribution_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

var (
	// now is Friday, 2026-09-25T12:34:56Z, in the zone of a user.
	now = time.Date(2026, time.September, 25, 15, 34, 56, 0, time.FixedZone("", 3*60*60))
	// anchor is the date of HEAD, Wednesday, 2026-09-02T10:00:00Z.
	anchor = time.Date(2026, time.September, 2, 13, 0, 0, 0, time.FixedZone("", 3*60*60))
)

// grammar are the inputs a command must decline with an ordinary error,
// whatever its grammar is: they are malformed.
var grammar = []string{
	"22", "2022-1", "2022-13", "2022-02-30", "2022-02-01T00:00:00", "2022-02-01T00:00:00.Z",
	"2022-02-01T00:00:00.5", "2022-02-01T00:00:00+3", "2022-02-01Tgarbage", "2022-02-01 00:00:00Z", "tomorrow", "NOW", "Git",
	"0000", "0001", "1969-12-31", "1970-01-01T01:00:00+02:00",
	"2022/54", "2022/+54", "2022/-54", "2022/-99999999", "2022/x", "2022/+", "2022/-", "2022/3/3", "2022//3",
	"0001/+1", "0001-01-01T00:00:00Z/1",
	"2026-01-12T00:00:00+03:60", "2026-01-12T00:00:00+24:00", "2026-01-12T1:00:00Z", "2026-01-12T00:00:00,5Z",
	"2026-01-12t00:00:00Z", "2026-01-12T00:00:00z", "2026/", "/", "now/", "2026-01-12T00:00:00Z/",
}

func TestLookupScope(t *testing.T) {
	tests := []struct {
		args     []string
		from, to string // no from means an error
	}{
		// the anchor, the date of HEAD
		{args: nil, from: "2026-08-23", to: "2026-09-05"},
		{args: []string{""}, from: "2026-08-23", to: "2026-09-05"},
		{args: []string{"/3"}, from: "2026-08-23", to: "2026-09-12"},
		{args: []string{"/+3"}, from: "2026-08-30", to: "now"},
		{args: []string{"/-3"}, from: "2026-08-09", to: "2026-09-05"},
		{args: []string{"git"}, from: "2026-08-23", to: "2026-09-05"},
		{args: []string{"git/3"}, from: "2026-08-23", to: "2026-09-12"},
		{args: []string{"git/+3"}, from: "2026-08-30", to: "now"},
		{args: []string{"git/-3"}, from: "2026-08-09", to: "2026-09-05"},
		// now
		{args: []string{"now"}, from: "2026-09-13", to: "now"},
		{args: []string{"now/3"}, from: "2026-09-13", to: "now"},
		{args: []string{"now/+3"}, from: "2026-09-20", to: "now"},
		{args: []string{"now/-3"}, from: "2026-08-30", to: "now"},
		// a year, Friday
		{args: []string{"2021"}, from: "2020-12-20", to: "2021-01-02"},
		{args: []string{"2021/3"}, from: "2020-12-20", to: "2021-01-09"},
		{args: []string{"2021/+3"}, from: "2020-12-27", to: "2021-01-23"},
		{args: []string{"2021/-3"}, from: "2020-12-06", to: "2021-01-02"},
		// a month, Saturday
		{args: []string{"2022-01"}, from: "2021-12-19", to: "2022-01-01"},
		{args: []string{"2022-01/3"}, from: "2021-12-19", to: "2022-01-08"},
		{args: []string{"2022-01/+3"}, from: "2021-12-26", to: "2022-01-22"},
		{args: []string{"2022-01/-3"}, from: "2021-12-05", to: "2022-01-01"},
		// a day, Sunday
		{args: []string{"2022-01-02"}, from: "2021-12-26", to: "2022-01-08"},
		{args: []string{"2022-01-02/3"}, from: "2021-12-26", to: "2022-01-15"},
		{args: []string{"2022-01-02/+3"}, from: "2022-01-02", to: "2022-01-29"},
		{args: []string{"2022-01-02/-3"}, from: "2021-12-12", to: "2022-01-08"},
		// a day at the boundary of years
		{args: []string{"2021-12-31"}, from: "2021-12-19", to: "2022-01-01"},
		{args: []string{"2021-12-31/+1"}, from: "2021-12-26", to: "2022-01-08"},
		// RFC 3339, Sunday
		{args: []string{"2022-07-24T08:40:56Z"}, from: "2022-07-17", to: "2022-07-30"},
		{args: []string{"2022-07-24T08:40:56Z/3"}, from: "2022-07-17", to: "2022-08-06"},
		{args: []string{"2022-07-24T08:40:56Z/+3"}, from: "2022-07-24", to: "2022-08-20"},
		{args: []string{"2022-07-24T08:40:56Z/-3"}, from: "2022-07-03", to: "2022-07-30"},
		// RFC 3339 with fractional seconds or an offset, Sunday in UTC
		{args: []string{"2022-07-24T08:40:56.5Z"}, from: "2022-07-17", to: "2022-07-30"},
		{args: []string{"2022-07-24T08:40:56.1234Z"}, from: "2022-07-17", to: "2022-07-30"},
		{args: []string{"2022-07-24T08:40:56.123456789+03:00"}, from: "2022-07-17", to: "2022-07-30"},
		{args: []string{"2022-07-23T20:00:00-07:00"}, from: "2022-07-17", to: "2022-07-30"},
		// RFC 3339 with an offset, Sunday there, but Saturday in UTC
		{args: []string{"2022-07-24T02:40:56+03:00"}, from: "2022-07-10", to: "2022-07-23"},
		{args: []string{"2022-07-24T02:40:56+03:00/3"}, from: "2022-07-10", to: "2022-07-30"},
		{args: []string{"2022-07-24T02:40:56+03:00/+3"}, from: "2022-07-17", to: "2022-08-13"},
		{args: []string{"2022-07-24T02:40:56+03:00/-3"}, from: "2022-06-26", to: "2022-07-23"},
		// the longest span and the earliest year
		{args: []string{"2022/53"}, from: "2021-06-27", to: "2022-07-02"},
		{args: []string{"2022/-53"}, from: "2020-12-20", to: "2022-01-01"},
		{args: []string{"1970"}, from: "1969-12-21", to: "1970-01-03"},
		// the future
		{args: []string{"2026-09-26"}, from: "2026-09-13", to: "now"},
		{args: []string{"2026-09-27"}, from: "2026-09-20", to: "now"},
		{args: []string{"2026-09-27/-3"}, from: "2026-09-06", to: "now"},
		{args: []string{"2026-10"}, from: "2026-09-20", to: "now"},
		{args: []string{"2026-09-27/+1"}},
		{args: []string{"2026-12-01/3"}},
		{args: []string{"2026-11"}},
		{args: []string{"2030"}},
		{args: []string{"2030-01-01"}},
		{args: []string{"2030-01-01/-1"}},
		{args: []string{"2030-01-01T00:00:00Z/-3"}},
	}
	for _, input := range grammar {
		tests = append(tests, struct {
			args     []string
			from, to string
		}{args: []string{input}})
	}

	for _, test := range tests {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			scope, err := LookupScope(test.args, anchor, now)
			if test.from == "" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			sane(t, scope, now, true)
			expect(t, scope, now, test.from, test.to)
		})
	}

	t.Run("the week starts now", func(t *testing.T) {
		sunday := xtime.UTC().Year(2026).Month(time.September).Day(20).Time()
		_, err := LookupScope([]string{"now/+1"}, anchor, sunday)
		assert.ErrorIs(t, err, xtime.ErrFuturePeriod)
	})
}

// sane checks the invariants of a range to show or to request: it is not
// empty, never ends after now, and starts at midnight UTC of Sunday if the
// range is made of weeks.
func sane(t testing.TB, scope xtime.Range, now time.Time, weeks bool) {
	t.Helper()

	assert.True(t, scope.From().Before(scope.To()), "from %s is not before to %s", scope.From(), scope.To())
	assert.False(t, scope.To().After(now), "to %s is after now", scope.To())
	if weeks {
		from := scope.From().UTC()
		assert.Equal(t, time.Sunday, from.Weekday())
		assert.Equal(t, xtime.TruncateToDay(from), from)
	}
}

// expect checks that the range starts at midnight UTC of the day from
// and lasts up to the end of the day to, or up to now.
func expect(t testing.TB, scope xtime.Range, now time.Time, from, to string) {
	t.Helper()

	first, err := time.Parse(xtime.DateOnly, from)
	require.NoError(t, err)
	assert.Equal(t, first, scope.From(), "from")

	end := now
	if to != "now" {
		last, err := time.Parse(xtime.DateOnly, to)
		require.NoError(t, err)
		end = last.Add(xtime.Day - time.Nanosecond)
	}
	assert.True(t, end.Equal(scope.To()), "to is %s, expected %s", scope.To(), end)
}
