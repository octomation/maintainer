package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareCodexVersions(t *testing.T) {
	for _, tt := range []struct{ a, b string }{
		{"0.2.2", "0.2.10"},
		{"0.2.10-rc.1", "0.2.10"},
		{"0.2.10-alpha", "0.2.10-alpha.1"},
		{"0.2.10-alpha.2", "0.2.10-alpha.10"},
		{"0.2.10-1", "0.2.10-alpha"},
		{"0.2.10", "0.2.10+build"},
		{"0.2.10+1", "0.2.10+01"},
		{"0.2.10+9", "0.2.10+10"},
		{"0.2.10+1", "0.2.10+a"},
		{"0.2.10", "nightly"},   // not semver: compared as strings
		{"01.2.3", "1.2.3"},     // a leading zero is not semver
		{"1.2.3-01", "1.2.3-1"}, // nor in a numeric prerelease
		{"latest", "nightly"},
	} {
		assert.Equal(t, -1, compareCodexVersions(tt.a, tt.b), "%s < %s", tt.a, tt.b)
		assert.Equal(t, 1, compareCodexVersions(tt.b, tt.a), "%s > %s", tt.b, tt.a)
	}
	assert.Zero(t, compareCodexVersions("1.2.3+x", "1.2.3+x"))
}

func TestCodexActiveVersion(t *testing.T) {
	cases := map[string]struct {
		dirs      []string
		want      string
		ambiguous []string
	}{
		"newest semver":         {dirs: []string{"0.2.1", "0.2.10", "0.2.2"}, want: "0.2.10"},
		"release over rc":       {dirs: []string{"0.2.10-rc.1", "0.2.10"}, want: "0.2.10"},
		"build metadata":        {dirs: []string{"0.2.2+1", "0.2.2+2", "0.2.2"}, want: "0.2.2+2"},
		"local wins":            {dirs: []string{"0.3.0", "local", "0.2.2", "nightly"}, want: "local"},
		"only other names":      {dirs: []string{"beta", "nightly"}, want: "nightly"},
		"invalid names skipped": {dirs: []string{"0.2.2", "v 1", "a=b"}, want: "0.2.2"},
		"none":                  {},
		"mixed names are ambiguous": {
			dirs: []string{"0.2.2", "nightly"}, want: "nightly", ambiguous: []string{"0.2.2", "nightly"},
		},
		// 1.9.0 < 1.10.0 as versions, 1.10.0 < 1.11x and 1.11x < 1.9.0 as
		// strings: no order exists, Codex's pick depends on the disk.
		"a cycle is ambiguous": {
			dirs: []string{"1.9.0", "1.10.0", "1.11x"}, ambiguous: []string{"1.10.0", "1.11x", "1.9.0"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			for _, dir := range tc.dirs {
				require.NoError(t, os.Mkdir(filepath.Join(base, dir), 0o755))
			}
			require.NoError(t, os.WriteFile(filepath.Join(base, "9.9.9"), nil, 0o644), "files are not versions")
			got, ambiguous := codexActiveVersion(base)
			switch {
			case tc.want != "":
				assert.Equal(t, tc.want, got)
			case tc.dirs == nil:
				assert.Empty(t, got)
			default:
				assert.Contains(t, tc.dirs, got, "some copy is still listed")
			}
			assert.Equal(t, tc.ambiguous, ambiguous)
		})
	}
}

func TestParseCodexPluginID(t *testing.T) {
	_, _, ok := parseCodexPluginID("my.plugin@team@octolab")
	assert.False(t, ok, "split at the last @, the plugin name holds an @")

	plugin, marketplace, ok := parseCodexPluginID("maintainer@octo_lab-1")
	require.True(t, ok)
	assert.Equal(t, "maintainer", plugin)
	assert.Equal(t, "octo_lab-1", marketplace)

	for _, id := range []string{"maintainer@..", "maintainer@.", "maintainer@a/b", `maintainer@a\b`, "maintainer@a.b", "maintainer@", "@octolab", "maintainer", "..@octolab", "a..b@octolab"} {
		_, _, ok := parseCodexPluginID(id)
		assert.False(t, ok, id)
	}
}
