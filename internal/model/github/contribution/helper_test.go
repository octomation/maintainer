package contribution_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/safe"
	"go.octolab.org/unsafe"

	. "go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestLookupRange(t *testing.T) {
	chm := load(t, "testdata/kamilsk.2021.json")

	t.Run("MAIN-99 (issue#124): correct centering", func(t *testing.T) {
		opts := DateOptions{
			Value: xtime.UTC().Year(2021).Month(time.January).Day(30).Time(),
			Weeks: 3, Half: true,
		}
		scope := LookupRange(opts).Until(time.Now())
		schedule, target := xtime.Everyday(xtime.Hours(5, 19, 0)), uint(5)
		suggestion := Suggest(chm, scope.Since(opts.Value), schedule, target)

		opts.Value = suggestion.Time
		scope = LookupRange(opts)
		assert.Equal(t, "2021-01-17", scope.From().Format(xtime.DateOnly))
		assert.Equal(t, "2021-02-06", scope.To().Format(xtime.DateOnly))
	})
}

// load reads a heat map snapshot, the same format the snapshot command writes.
func load(t testing.TB, name string) HeatMap {
	f, err := os.Open(name)
	require.NoError(t, err)
	defer safe.Close(f, unsafe.Ignore)

	chm := make(HeatMap)
	require.NoError(t, json.NewDecoder(f).Decode(&chm))

	return chm
}
