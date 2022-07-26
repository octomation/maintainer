package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
				assert.Equal(t, b, r.Base())
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
