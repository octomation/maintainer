package contribution_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestParseDate(t *testing.T) {
	start := xtime.UTC().Year(2021).Month(time.February).Day(8).Hour(9).Minute(16).Second(3)
	now := time.Now().UTC()

	tests := []struct {
		name   string
		arg    string
		fDate  time.Time
		fWeeks int
		health func(require.TestingT, error, ...interface{})
		assert func(testing.TB, xtime.Range)
	}{
		{
			name:   "RFC3339 with default range",
			arg:    start.Format(time.RFC3339),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-27")
			},
		},
		{
			name:   "RFC3339 with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-20")
			},
		},
		{
			name:   "RFC3339 with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-13")
			},
		},
		{
			name:   "RFC3339 with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-02-07")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-03-06")
			},
		},
		{
			name:   "DateOnly with default range",
			arg:    start.Format(xtime.DateOnly),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-27")
			},
		},
		{
			name:   "DateOnly with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.DateOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-20")
			},
		},
		{
			name:   "DateOnly with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.DateOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-13")
			},
		},
		{
			name:   "DateOnly with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.DateOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-02-07")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-03-06")
			},
		},
		{
			name:   "YearAndMonth with default range",
			arg:    start.Format(xtime.YearAndMonth),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-20")
			},
		},
		{
			name:   "YearAndMonth with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.YearAndMonth)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-13")
			},
		},
		{
			name:   "YearAndMonth with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.YearAndMonth)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-10")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-06")
			},
		},
		{
			name:   "YearAndMonth with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.YearAndMonth)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-27")
			},
		},
		{
			name:   "YearOnly with default range",
			arg:    start.Format(xtime.YearOnly),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-13")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-16")
			},
		},
		{
			name:   "YearOnly with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.YearOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-20")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-09")
			},
		},
		{
			name:   "YearOnly with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.YearOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-06")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-02")
			},
		},
		{
			name:   "YearOnly with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.YearOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-27")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-23")
			},
		},
		{
			name:   "author date, strict ISO 8601 format",
			arg:    "2022-07-24T08:40:56+03:00/-3",
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2022-07-03")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2022-07-30")
			},
		},
		{
			name:   "current time",
			arg:    "/-3",
			fDate:  now,
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts, err := ParseDate([]string{test.arg}, test.fDate, test.fWeeks)
			test.health(t, err)
			test.assert(t, contribution.LookupRange(opts))
		})
	}
}
