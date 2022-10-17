package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestSchedule_Suggest(t *testing.T) {
	tests := []struct {
		name     string
		hours    Schedule
		time     time.Time
		expected time.Time
	}{
		{
			name:     "empty schedule",
			hours:    make(Schedule),
			time:     UTC().Hour(8).Minute(22).Second(18).Time(),
			expected: time.Time{},
		},
		{
			name:     "before interval",
			hours:    Everyday(Hours(9, 12, 0), Hours(14, 18, 0), Hours(20, 22, 0)),
			time:     UTC().Hour(8).Minute(22).Second(18).Time(),
			expected: UTC().Hour(9).Time(),
		},
		{
			name:     "inside first interval",
			hours:    Everyday(Hours(9, 12, 0), Hours(14, 18, 0), Hours(20, 22, 0)),
			time:     UTC().Hour(10).Minute(22).Second(18).Time(),
			expected: UTC().Hour(10).Minute(22).Second(18).Time(),
		},
		{
			name:     "between first and second interval",
			hours:    Everyday(Hours(9, 12, 0), Hours(14, 18, 0), Hours(20, 22, 0)),
			time:     UTC().Hour(12).Minute(22).Second(18).Time(),
			expected: UTC().Hour(14).Time(),
		},
		{
			name:     "inside second interval",
			hours:    Everyday(Hours(9, 12, 0), Hours(14, 18, 0), Hours(20, 22, 0)),
			time:     UTC().Hour(15).Minute(22).Second(18).Time(),
			expected: UTC().Hour(15).Minute(22).Second(18).Time(),
		},
		{
			name:     "between second and third interval",
			hours:    Everyday(Hours(9, 12, 0), Hours(14, 18, 0), Hours(20, 22, 0)),
			time:     UTC().Hour(19).Minute(22).Second(18).Time(),
			expected: UTC().Hour(20).Time(),
		},
		{
			name:     "outside intervals",
			hours:    Everyday(Hours(9, 12, 0), Hours(14, 18, 0), Hours(20, 22, 0)),
			time:     UTC().Hour(23).Minute(22).Second(18).Time(),
			expected: time.Time{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.hours.Suggest(test.time))
		})
	}
}
