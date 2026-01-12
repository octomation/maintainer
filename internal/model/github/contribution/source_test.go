package contribution_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestFileSource_Fetch(t *testing.T) {
	day := xtime.UTC().Year(2022).Month(time.January).Day(3).Time()

	tests := map[string]struct {
		file    string
		content string
		counts  map[time.Time]uint
		err     []string
	}{
		"midnight UTC": {
			content: `{"2022-01-03T00:00:00Z": 5, "2022-01-04T00:00:00Z": 2}`,
			counts:  map[time.Time]uint{day: 5, day.AddDate(0, 0, 1): 2},
		},
		"midnight with a zero offset": {
			content: `{"2022-01-03T00:00:00+00:00": 5}`,
			counts:  map[time.Time]uint{day: 5},
		},
		"empty snapshot": {
			content: `{}`,
			counts:  map[time.Time]uint{},
		},
		"noon": {
			content: `{"2022-01-03T00:00:00Z": 5, "2022-01-03T12:00:00Z": 2}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T12:00:00Z"`},
		},
		"midnight in another zone": {
			content: `{"2022-01-03T00:00:00+03:00": 5}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00+03:00"`},
		},
		"fraction of a second": {
			content: `{"2022-01-03T00:00:00.5Z": 5}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00.5Z"`},
		},
		"the same day twice": {
			content: `{"2022-01-03T00:00:00Z": 5, "2022-01-03T00:00:00+00:00": 6}`,
			err:     []string{`"snapshot.json"`, "2022-01-03"},
		},
		"not a snapshot": {
			content: `[1, 2, 3]`,
			err:     []string{`"snapshot.json"`},
		},
		"trailing data": {
			content: `{} THIS IS NOT JSON`,
			err:     []string{`"snapshot.json"`, "after the object"},
		},
		"second object": {
			content: `{} {}`,
			err:     []string{`"snapshot.json"`, "after the object"},
		},
		"trailing spaces": {
			content: "{\"2022-01-03T00:00:00Z\": 5}\n\t \n",
			counts:  map[time.Time]uint{day: 5},
		},
		"null": {
			content: `null`,
			err:     []string{`"snapshot.json"`, "not an object"},
		},
		"scalar": {
			content: `5`,
			err:     []string{`"snapshot.json"`, "not an object"},
		},
		"empty file": {
			content: ``,
			err:     []string{`"snapshot.json"`, "empty"},
		},
		"duplicate key": {
			content: `{"2022-01-03T00:00:00Z": 1, "2022-01-03T00:00:00Z": 5}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00Z"`, "duplicated"},
		},
		"fractional count": {
			content: `{"2022-01-03T00:00:00Z": 1.5}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00Z"`, "1.5"},
		},
		"negative count": {
			content: `{"2022-01-03T00:00:00Z": -1}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00Z"`},
		},
		"string count": {
			content: `{"2022-01-03T00:00:00Z": "5"}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00Z"`},
		},
		"boolean count": {
			content: `{"2022-01-03T00:00:00Z": true}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00Z"`},
		},
		"max count": {
			content: `{"2022-01-03T00:00:00Z": 9223372036854775807}`,
			counts:  map[time.Time]uint{day: math.MaxInt64},
		},
		"count above max": {
			content: `{"2022-01-03T00:00:00Z": 9223372036854775808}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00Z"`},
		},
		"max uint count": {
			content: `{"2022-01-03T00:00:00Z": 18446744073709551615}`,
			err:     []string{`"snapshot.json"`, `"2022-01-03T00:00:00Z"`},
		},
		"not a time": {
			content: `{"yesterday": 5}`,
			err:     []string{`"snapshot.json"`, `"yesterday"`},
		},
		"yaml": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: 5\n\"2022-01-04T00:00:00Z\": 2\n",
			counts:  map[time.Time]uint{day: 5, day.AddDate(0, 0, 1): 2},
		},
		"yaml empty mapping": {
			file:    "snapshot.yaml",
			content: "{}\n",
			counts:  map[time.Time]uint{},
		},
		"yaml max count": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: 9223372036854775807\n",
			counts:  map[time.Time]uint{day: math.MaxInt64},
		},
		"yaml count above max": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: 9223372036854775808\n",
			err:     []string{`"snapshot.yml"`, `"2022-01-03T00:00:00Z"`},
		},
		"yaml fractional count": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: 1.5\n",
			err:     []string{`"snapshot.yml"`, `"2022-01-03T00:00:00Z"`},
		},
		"yaml negative count": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: -1\n",
			err:     []string{`"snapshot.yml"`, `"2022-01-03T00:00:00Z"`},
		},
		"yaml string count": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: five\n",
			err:     []string{`"snapshot.yml"`, `"2022-01-03T00:00:00Z"`},
		},
		"yaml duplicate key": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: 1\n2022-01-03T00:00:00Z: 5\n",
			err:     []string{`"snapshot.yml"`, "2022-01-03T00:00:00Z"},
		},
		"yaml second document": {
			file:    "snapshot.yml",
			content: "2022-01-03T00:00:00Z: 1\n---\n2022-01-04T00:00:00Z: 5\n",
			err:     []string{`"snapshot.yml"`, "document"},
		},
		"yaml null": {
			file:    "snapshot.yml",
			content: "null\n",
			err:     []string{`"snapshot.yml"`, "not a mapping"},
		},
		"yaml sequence": {
			file:    "snapshot.yml",
			content: "- 1\n- 2\n",
			err:     []string{`"snapshot.yml"`, "not a mapping"},
		},
		"yaml empty file": {
			file:    "snapshot.yml",
			content: "",
			err:     []string{`"snapshot.yml"`, "empty"},
		},
		"unknown extension": {
			file:    "snapshot.txt",
			content: `{}`,
			err:     []string{`"snapshot.txt"`, "unsupported format"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			file := test.file
			if file == "" {
				file = "snapshot.json"
			}
			fs := afero.NewMemMapFs()
			require.NoError(t, afero.WriteFile(fs, file, []byte(test.content), 0o644))

			chm, err := NewFileSource(fs, file).Fetch(context.Background())
			if test.err != nil {
				require.Error(t, err)
				for _, part := range test.err {
					assert.Contains(t, err.Error(), part)
				}
				return
			}
			require.NoError(t, err)
			assert.Len(t, chm, len(test.counts))
			for ts, count := range test.counts {
				assert.Equal(t, count, chm.Count(ts), ts.Format(xtime.DateOnly))
			}
		})
	}
}

func TestUpstreamSource_Fetch(t *testing.T) {
	now := xtime.UTC().Year(2026).Month(time.September).Day(25).Hour(12).Minute(34).Second(56).Time()

	tests := map[string]struct {
		year  time.Time
		scope xtime.Range
		err   error
	}{
		"past year": {
			year:  xtime.UTC().Year(2021).Time(),
			scope: xtime.RangeByYears(xtime.UTC().Year(2021).Time(), 0, false),
		},
		"current year": {
			year:  xtime.UTC().Year(2026).Time(),
			scope: xtime.NewRange(xtime.UTC().Year(2026).Time(), now),
		},
		"future year": {
			year: xtime.UTC().Year(2030).Time(),
			err:  xtime.ErrFuturePeriod,
		},
		"too early year": {
			year: xtime.UTC().Year(1).Time(),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var requested []xtime.Range
			src := NewUpstreamSource(contributor(func(_ context.Context, scope xtime.Range) (HeatMap, error) {
				requested = append(requested, scope)
				return make(HeatMap), nil
			}), test.year)
			src.Now = func() time.Time { return now }

			chm, err := src.Fetch(context.Background())
			if test.scope.IsZero() {
				require.Error(t, err)
				if test.err != nil {
					assert.ErrorIs(t, err, test.err)
				}
				assert.Contains(t, err.Error(), src.Location())
				assert.Empty(t, requested, "nothing is requested")
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, chm)
			assert.Equal(t, []xtime.Range{test.scope}, requested)
		})
	}
}

// contributor provides a heat map with a function.
type contributor func(context.Context, xtime.Range) (HeatMap, error)

func (fn contributor) ContributionHeatMap(ctx context.Context, scope xtime.Range) (HeatMap, error) {
	return fn(ctx, scope)
}
