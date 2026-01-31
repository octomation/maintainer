package skill_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/skill"
)

func version(t *testing.T, s string) Version {
	t.Helper()
	v, ok := ParseVersion(s)
	require.True(t, ok, s)
	return v
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"0.9.9", "0.10.0", -1},
		{"1.0.0-rc.1", "1.0.0", -1},
		{"1.0.0", "1.0.0-rc.1", 1},
		{"1.0.0-rc.2", "1.0.0-rc.10", -1},
		{"1.0.0-alpha", "1.0.0-1", 1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, version(t, tc.a).Compare(version(t, tc.b)), "%s vs %s", tc.a, tc.b)
	}
}

func TestParseVersion(t *testing.T) {
	for _, s := range []string{"1.0.0+build", "1.0.0+foo..bar", "01.0.0", "1.0.0-01", "1.0", ""} {
		_, ok := ParseVersion(s)
		assert.False(t, ok, s)
	}
	assert.Equal(t, "1.0.0-rc.1", version(t, "v1.0.0-rc.1").String())
}

func TestDefaultRange(t *testing.T) {
	assert.Equal(t, ">=1.4.2 <2.0.0", DefaultRange(version(t, "1.4.2")))
	assert.Equal(t, ">=0.4.2 <0.5.0", DefaultRange(version(t, "0.4.2")))
	assert.Equal(t, "=0.0.2", DefaultRange(version(t, "0.0.2")))
	assert.Equal(t, "=1.0.0-rc.1", DefaultRange(version(t, "1.0.0-rc.1")))
}

func TestRange_Contains(t *testing.T) {
	tests := []struct {
		version, rng string
		want         bool
	}{
		{"0.2.5", ">=0.2.0 <0.3.0", true},
		{"0.3.0", ">=0.2.0 <0.3.0", false},
		{"0.1.9", ">=0.2.0 <0.3.0", false},
		{"1.0.0", "1.0.0", true},
		{"2.0.0", ">=1.0.0 <2.0.0", false},
		{"0.3.0-rc.1", ">=0.2.0 <0.3.0", false},
		{"0.3.0-rc.2", ">=0.3.0-rc.1 <0.3.0", true},
		{"1.0.0-rc.1", "=1.0.0-rc.1", true},
	}
	for _, tc := range tests {
		r, ok := ParseRange(tc.rng)
		require.True(t, ok, tc.rng)
		assert.Equal(t, tc.want, r.Contains(version(t, tc.version)), "%s in %s", tc.version, tc.rng)
	}
	for _, s := range []string{"", ">=x", ">=v1.0.0"} {
		_, ok := ParseRange(s)
		assert.False(t, ok, s)
	}
}

func TestCheckRange(t *testing.T) {
	tests := []struct {
		version, rng, problem string
	}{
		{"0.2.0", ">=0.2.0 <0.3.0", ""},
		{"0.2.0", "<0.3.0 >=0.2.0", ""},
		{"0.2.3", ">=0.2.3 <0.2.5", ""},
		{"1.4.2", ">=1.4.2 <2.0.0", ""},
		{"0.0.2", "=0.0.2", ""},
		{"0.2.0", "=0.2.0", ""},
		{"1.4.2", "=1.4.2", ""},
		{"0.2.0", ">=0.1.0 <9.0.0", "start at the release"},
		{"0.2.0", ">=0.2.0 <9.0.0", "no later than <0.3.0"},
		{"0.2.0", ">=0.2.0", "look like"},
		{"0.2.0", ">=0.2.0 >=0.2.0", "look like"},
		{"0.0.2", ">=0.0.2 <0.1.0", "must be =0.0.2"},
		{"0.2.0", ">=0.2.0 <0.3.0-rc.1", "prereleases"},
		{"0.2.0", "", "version range"},
	}
	for _, tc := range tests {
		err := CheckRange(version(t, tc.version), tc.rng)
		if tc.problem == "" {
			assert.NoError(t, err, "%s for %s", tc.rng, tc.version)
			continue
		}
		if assert.Error(t, err, "%s for %s", tc.rng, tc.version) {
			assert.Contains(t, err.Error(), tc.problem)
		}
	}
}
