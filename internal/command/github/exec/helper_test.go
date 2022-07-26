package exec_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/command/github/exec"
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
				assert.Equal(t, start.Time(), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-27")
			},
		},
		{
			name:   "RFC3339 with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, start.Time(), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-20")
			},
		},
		{
			name:   "RFC3339 with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, start.Time(), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-13")
			},
		},
		{
			name:   "RFC3339 with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, start.Time(), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-02-07")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-03-06")
			},
		},
		{
			name:   "RFC3339Day with default range",
			arg:    start.Format(xtime.RFC3339Day),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToDay(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-27")
			},
		},
		{
			name:   "RFC3339Day with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.RFC3339Day)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToDay(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-20")
			},
		},
		{
			name:   "RFC3339Day with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.RFC3339Day)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToDay(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-13")
			},
		},
		{
			name:   "RFC3339Day with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.RFC3339Day)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToDay(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-02-07")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-03-06")
			},
		},
		{
			name:   "RFC3339Month with default range",
			arg:    start.Format(xtime.RFC3339Month),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToMonth(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-20")
			},
		},
		{
			name:   "RFC3339Month with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.RFC3339Month)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToMonth(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-13")
			},
		},
		{
			name:   "RFC3339Month with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.RFC3339Month)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToMonth(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-10")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-06")
			},
		},
		{
			name:   "RFC3339Month with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.RFC3339Month)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToMonth(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-02-27")
			},
		},
		{
			name:   "RFC3339Year with default range",
			arg:    start.Format(xtime.RFC3339Year),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToYear(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2020-12-13")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-01-16")
			},
		},
		{
			name:   "RFC3339Year with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.RFC3339Year)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToYear(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2020-12-20")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-01-09")
			},
		},
		{
			name:   "RFC3339Year with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.RFC3339Year)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToYear(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2020-12-06")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-01-02")
			},
		},
		{
			name:   "RFC3339Year with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.RFC3339Year)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, xtime.TruncateToYear(start.Time()), lr.Base())
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2020-12-27")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2021-01-23")
			},
		},
		{
			name:   "author date, strict ISO 8601 format",
			arg:    "2022-07-24T08:40:56+03:00/-3",
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.RFC3339Day), "2022-07-03")
				assert.Equal(t, lr.To().Format(xtime.RFC3339Day), "2022-07-30")
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
			opts, err := exec.ParseDate([]string{test.arg}, test.fDate, test.fWeeks)
			test.health(t, err)
			test.assert(t, contribution.LookupRange(opts))
		})
	}
}
