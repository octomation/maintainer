package vanity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fill(t *testing.T) {
	tests := map[string]struct {
		prefix   string
		packages []string
		expected []string
	}{
		"issue#6": {
			prefix: "go.octolab.org/toolkit/cli",
			packages: []string{
				"go.octolab.org/toolkit/cli/cobra",
				"go.octolab.org/toolkit/cli/debugger",
			},
			expected: []string{
				"go.octolab.org/toolkit/cli",
				"go.octolab.org/toolkit/cli/cobra",
				"go.octolab.org/toolkit/cli/debugger",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, fill(test.prefix, test.packages))
		})
	}
}
