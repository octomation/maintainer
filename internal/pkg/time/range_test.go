package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestRangeByWeeks(t *testing.T) {
	start := UTC().Year(2021).Month(time.February).Day(8).Hour(9).Minute(16).Second(3)

	tests := []struct {
		name  string
		date  time.Time
		weeks int
		half  bool
		check func(t testing.TB, b time.Time, r Range)
	}{
		{
			name:  "beginning, one week ahead",
			date:  start.Time(),
			weeks: 1,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Monday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(2*Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "beginning, one week behind",
			date:  start.Time(),
			weeks: -1,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Monday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b).Add(-Week), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "beginning, half week",
			date:  start.Time(),
			weeks: 1,
			half:  true,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Monday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "midweek, one week ahead",
			date:  start.Day(10).Time(),
			weeks: 1,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Wednesday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(2*Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "midweek, one week behind",
			date:  start.Day(10).Time(),
			weeks: -1,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Wednesday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b).Add(-Week), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "midweek, half week",
			date:  start.Day(10).Time(),
			weeks: 1,
			half:  true,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Wednesday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "ending, one week ahead",
			date:  start.Day(14).Time(),
			weeks: 1,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Sunday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(2*Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "ending, one week behind",
			date:  start.Day(14).Time(),
			weeks: -1,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Sunday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b).Add(-Week), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "ending, half week",
			date:  start.Day(14).Time(),
			weeks: 1,
			half:  true,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Sunday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "odd weeks ahead",
			date:  start.Time(),
			weeks: 5,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Monday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(6*Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "odd weeks behind",
			date:  start.Day(10).Time(),
			weeks: -5,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Wednesday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b).Add(-5*Week), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "half odd weeks",
			date:  start.Day(14).Time(),
			weeks: 5,
			half:  true,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Sunday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b).Add(-2*Week), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(3*Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "even weeks ahead",
			date:  start.Time(),
			weeks: 4,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Monday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(5*Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "even weeks behind",
			date:  start.Day(10).Time(),
			weeks: -4,
			half:  false,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Wednesday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b).Add(-4*Week), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(Week-time.Nanosecond), r.To())
			},
		},
		{
			name:  "half even weeks",
			date:  start.Day(14).Time(),
			weeks: 4,
			half:  true,
			check: func(t testing.TB, b time.Time, r Range) {
				assert.Equal(t, time.Sunday, b.Weekday())
				assert.Equal(t, TruncateToWeek(b).Add(-2*Week), r.From())
				assert.Equal(t, TruncateToWeek(b).Add(3*Week-time.Nanosecond), r.To())
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.check(t, test.date, RangeByWeeks(test.date, test.weeks, test.half))
		})
	}
}

func TestGregorianWeeks(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected Range
	}{
		{
			name: "sunday",
			date: UTC().Year(2021).Month(time.December).Day(12).Hour(8).Time(),
			expected: NewRange(
				UTC().Year(2021).Month(time.December).Day(12).Hour(0).Time(),
				UTC().Year(2021).Month(time.December).Day(19).Add(-time.Nanosecond).Time(),
			),
		},
		{
			name: "monday",
			date: UTC().Year(2021).Month(time.December).Day(13).Hour(8).Time(),
			expected: NewRange(
				UTC().Year(2021).Month(time.December).Day(12).Hour(0).Time(),
				UTC().Year(2021).Month(time.December).Day(19).Add(-time.Nanosecond).Time(),
			),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, GregorianWeeks(test.date, 0, false))
		})
	}
}

func TestRange_ExcludeFuture(t *testing.T) {
	now := UTC().Year(2026).Month(time.September).Day(25).Hour(12).Minute(34).Second(56).Time()
	week := func(from time.Time) Range { return NewRange(from, from.Add(Week-time.Nanosecond)) }

	tests := map[string]struct {
		scope    Range
		expected Range
		err      error
	}{
		"in the past": {
			scope:    week(UTC().Year(2026).Month(time.September).Day(13).Time()),
			expected: week(UTC().Year(2026).Month(time.September).Day(13).Time()),
		},
		"ends at now": {
			scope:    NewRange(UTC().Year(2026).Month(time.September).Day(20).Time(), now),
			expected: NewRange(UTC().Year(2026).Month(time.September).Day(20).Time(), now),
		},
		"through now": {
			scope:    week(UTC().Year(2026).Month(time.September).Day(20).Time()),
			expected: NewRange(UTC().Year(2026).Month(time.September).Day(20).Time(), now),
		},
		"starts at now": {
			scope: week(now),
			err:   ErrFuturePeriod,
		},
		"starts after now": {
			scope: week(UTC().Year(2026).Month(time.September).Day(27).Time()),
			err:   ErrFuturePeriod,
		},
		"far in the future": {
			scope: RangeByYears(UTC().Year(2030).Time(), 0, false),
			err:   ErrFuturePeriod,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			scope, err := test.scope.ExcludeFuture(now)
			if test.err != nil {
				require.ErrorIs(t, err, test.err)
				assert.Contains(t, err.Error(), test.scope.From().Format(time.RFC3339))
				assert.True(t, scope.IsZero())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, scope)
		})
	}
}

func TestGregorianWeek(t *testing.T) {
	tests := map[string]struct {
		date time.Time
		year int
		week int
	}{
		"2020-12-27, contains January 1":    {UTC().Year(2020).Month(time.December).Day(27).Time(), 2021, 1},
		"2021-01-03, after a long ISO year": {UTC().Year(2021).Month(time.January).Day(3).Time(), 2021, 2},
		"2021-12-26, contains January 1":    {UTC().Year(2021).Month(time.December).Day(26).Time(), 2022, 1},
		"2022-01-02":                        {UTC().Year(2022).Month(time.January).Day(2).Time(), 2022, 2},
		"2022-06-05":                        {UTC().Year(2022).Month(time.June).Day(5).Time(), 2022, 24},
		"2024-12-29, contains January 1":    {UTC().Year(2024).Month(time.December).Day(29).Time(), 2025, 1},
		"2025-12-28, contains January 1":    {UTC().Year(2025).Month(time.December).Day(28).Time(), 2026, 1},
		"2020-12-31, Thursday":              {UTC().Year(2020).Month(time.December).Day(31).Hour(23).Time(), 2021, 1},
		"2021-01-02, Saturday":              {UTC().Year(2021).Month(time.January).Day(2).Time(), 2021, 1},
		"2022-06-08, Wednesday":             {UTC().Year(2022).Month(time.June).Day(8).Time(), 2022, 24},
		"2022-06-11, Saturday":              {UTC().Year(2022).Month(time.June).Day(11).Hour(23).Time(), 2022, 24},
		"2022-12-24, the 52nd one":          {UTC().Year(2022).Month(time.December).Day(24).Time(), 2022, 52},
		"2022-12-31, the 53rd one":          {UTC().Year(2022).Month(time.December).Day(31).Time(), 2022, 53},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			year, week := GregorianWeek(test.date)
			assert.Equal(t, test.year, year)
			assert.Equal(t, test.week, week)
		})
	}
}
