package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestTruncateToWeek(t *testing.T) {
	tests := map[string]struct {
		day  time.Time
		want time.Time
	}{
		time.Monday.String(): {
			time.Date(2022, 07, 4, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 07, 4, 0, 0, 0, 0, time.UTC),
		},
		time.Tuesday.String(): {
			time.Date(2022, 7, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 7, 4, 0, 0, 0, 0, time.UTC),
		},
		time.Wednesday.String(): {
			time.Date(2022, 7, 6, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 7, 4, 0, 0, 0, 0, time.UTC),
		},
		time.Thursday.String(): {
			time.Date(2022, 7, 7, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 7, 4, 0, 0, 0, 0, time.UTC),
		},
		time.Friday.String(): {
			time.Date(2022, 7, 5, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 7, 4, 0, 0, 0, 0, time.UTC),
		},
		time.Saturday.String(): {
			time.Date(2022, 7, 9, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 7, 4, 0, 0, 0, 0, time.UTC),
		},
		time.Sunday.String(): {
			time.Date(2022, 7, 10, 0, 0, 0, 0, time.UTC),
			time.Date(2022, 7, 4, 0, 0, 0, 0, time.UTC),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := TruncateToWeek(test.day)
			assert.Equal(t, test.want, got)
		})
	}
}
