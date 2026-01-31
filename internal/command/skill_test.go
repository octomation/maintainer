package command_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/skill"
	"go.octolab.org/toolset/maintainer/skills"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	root := New()
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	EnforceUsage(root)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

// code is the exit code main resolves err to.
func code(err error) int {
	var coder exit.Coder
	switch {
	case err == nil:
		return exit.OK
	case errors.As(err, &coder):
		return coder.ExitCode()
	}
	return exit.Transport
}

// sandbox points every agent directory into a temporary home.
func sandbox(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("CODEX_HOME", "")
	t.Chdir(t.TempDir())
	return home
}

func TestSkill_offline(t *testing.T) {
	// No token, config or .env is needed.
	sandbox(t)
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	out, err := run(t, "skill", "show")
	require.NoError(t, err)
	assert.Contains(t, out, "name: maintainer")

	out, err = run(t, "skill", "show", "--file", "references/cli.md")
	require.NoError(t, err)
	assert.Contains(t, out, "## maintainer skill install")

	_, err = run(t, "skill", "show", "--file", "missing.md")
	assert.Equal(t, exit.User, code(err))

	_, err = run(t, "skill", "install")
	assert.Equal(t, exit.User, code(err), "a missing required flag is a usage error")
	_, err = run(t, "skill", "bogus")
	assert.Equal(t, exit.User, code(err))
}

func TestSkill_info(t *testing.T) {
	out, err := run(t, "skill", "info", "--json")
	require.NoError(t, err)
	var info map[string]string
	require.NoError(t, json.Unmarshal([]byte(out), &info))
	digest, err := skill.Digest(skills.Maintainer())
	require.NoError(t, err)
	assert.Equal(t, digest, info["digest"])
	assert.Equal(t, skill.DigestAlgorithm, info["digest_algorithm"])
	assert.Equal(t, "maintainer", info["name"])
	version, ok := skill.ParseVersion(info["version"])
	require.True(t, ok, info["version"])
	// A stable release declares >=X.Y.Z <..., a prerelease exactly =X.Y.Z-rcN.
	assert.NoError(t, skill.CheckRange(version, info["tool_version_range"]))
}

func TestSkill_installAndStatus(t *testing.T) {
	home := sandbox(t)

	_, err := run(t, "skill", "status", "--agent", "codex")
	assert.Equal(t, exit.Transport, code(err), "a requested agent without a copy fails")

	out, err := run(t, "skill", "install", "--agent", "claude-code,codex")
	require.NoError(t, err)
	assert.Contains(t, out, filepath.Join(home, ".claude", "skills", "maintainer"))
	assert.Contains(t, out, filepath.Join(home, ".agents", "skills", "maintainer"))

	out, err = run(t, "skill", "install", "--agent", "codex")
	require.NoError(t, err)
	assert.Contains(t, out, "already holds")

	out, err = run(t, "skill", "status", "--json", "--exact")
	require.NoError(t, err)
	var copies []skill.Copy
	for _, line := range bytes.Split(bytes.TrimSpace([]byte(out)), []byte("\n")) {
		var c skill.Copy
		require.NoError(t, json.Unmarshal(line, &c))
		copies = append(copies, c)
	}
	require.Len(t, copies, 2)
	for _, c := range copies {
		assert.Equal(t, "exact", c.ReleaseMatch)
		assert.Equal(t, "intact", c.Integrity)
		assert.Equal(t, "maintainer", c.Ownership)
	}

	_, err = run(t, "skill", "install", "--agent", "claude-code", "--project")
	require.NoError(t, err)
	out, err = run(t, "skill", "status", "--agent", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "Claude Code loads 2 copies")

	_, err = run(t, "skill", "install", "--agent", "cursor")
	assert.Equal(t, exit.User, code(err))
}

func TestSkill_statusCodexPlugin(t *testing.T) {
	home := sandbox(t)
	codex := filepath.Join(home, ".codex")
	dir := filepath.Join(codex, "plugins", "cache", "octolab", "maintainer", "0.2.2", "skills", "maintainer")
	require.NoError(t, os.MkdirAll(filepath.Dir(dir), 0o755))
	require.NoError(t, os.CopyFS(dir, skills.Maintainer()))
	require.NoError(t, os.WriteFile(filepath.Join(codex, "config.toml"),
		[]byte("[plugins.\"maintainer@octolab\"]\nenabled = true\n"), 0o644))

	out, err := run(t, "skill", "status", "--agent", "codex", "--exact")
	require.NoError(t, err, "the plugin copy is the agent's copy")
	assert.Contains(t, out, "Codex plugin: maintainer@octolab ~/.codex/plugins/cache/octolab/maintainer/0.2.2/skills/maintainer")
	assert.Contains(t, out, "same as this release, installed by another tool")

	_, err = run(t, "skill", "install", "--agent", "codex")
	require.NoError(t, err)
	out, err = run(t, "skill", "status", "--agent", "codex")
	require.NoError(t, err)
	assert.Contains(t, out, "Codex loads 2 copies")
}

func TestSkill_statusAmbiguousCodexPlugin(t *testing.T) {
	home := sandbox(t)
	codex := filepath.Join(home, ".codex")
	for _, version := range []string{"1.9.0", "1.10.0", "1.11x"} {
		dir := filepath.Join(codex, "plugins", "cache", "octolab", "maintainer", version, "skills", "maintainer")
		require.NoError(t, os.MkdirAll(filepath.Dir(dir), 0o755))
		require.NoError(t, os.CopyFS(dir, skills.Maintainer()))
	}
	require.NoError(t, os.WriteFile(filepath.Join(codex, "config.toml"),
		[]byte("[plugins.\"maintainer@octolab\"]\n"), 0o644))

	out, err := run(t, "skill", "status", "--agent", "codex")
	require.NoError(t, err, "without --exact an ambiguous copy is only reported")
	assert.Contains(t, out, "Codex plugin: maintainer@octolab")
	assert.Contains(t, out, "match with this release unknown")
	assert.Contains(t, out, "Codex's pick among 1.10.0, 1.11x, 1.9.0 depends on the order on disk; remove the stale ones")

	_, err = run(t, "skill", "status", "--agent", "codex", "--exact")
	assert.Equal(t, exit.Transport, code(err), "an ambiguous copy is never an exact match")

	out, err = run(t, "skill", "status", "--agent", "codex", "--json")
	require.NoError(t, err)
	var c skill.Copy
	require.NoError(t, json.Unmarshal([]byte(out), &c))
	assert.Equal(t, []string{"1.10.0", "1.11x", "1.9.0"}, c.AmbiguousVersions)
	assert.Equal(t, "unknown", c.ReleaseMatch)
}

func TestSkill_statusAmbiguousCodexPluginWithoutCopy(t *testing.T) {
	home := sandbox(t)
	codex := filepath.Join(home, ".codex")
	cache := filepath.Join(codex, "plugins", "cache", "octolab", "maintainer")
	for _, version := range []string{"1.10.0", "1.11x"} {
		dir := filepath.Join(cache, version, "skills", "maintainer")
		require.NoError(t, os.MkdirAll(filepath.Dir(dir), 0o755))
		require.NoError(t, os.CopyFS(dir, skills.Maintainer()))
	}
	// The pick of the cycle holds no skill at all.
	require.NoError(t, os.MkdirAll(filepath.Join(cache, "1.9.0"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(codex, "config.toml"),
		[]byte("[plugins.\"maintainer@octolab\"]\n"), 0o644))
	_, err := run(t, "skill", "install", "--agent", "codex")
	require.NoError(t, err)

	out, err := run(t, "skill", "status", "--agent", "codex", "--json")
	require.NoError(t, err)
	var plugin *skill.Copy
	for _, line := range bytes.Split(bytes.TrimSpace([]byte(out)), []byte("\n")) {
		var c skill.Copy
		require.NoError(t, json.Unmarshal(line, &c))
		if c.Scope == "plugin" {
			plugin = &c
		}
	}
	require.NotNil(t, plugin, "the ambiguous plugin stays listed without a copy")
	assert.Equal(t, filepath.Join(cache, "1.9.0", "skills", "maintainer"), plugin.Path)
	assert.Equal(t, []string{"1.10.0", "1.11x", "1.9.0"}, plugin.AmbiguousVersions)

	out, err = run(t, "skill", "status", "--agent", "codex", "--exact")
	assert.Equal(t, exit.Transport, code(err), "an exact user copy does not hide the ambiguity")
	assert.Contains(t, out, "unreadable SKILL.md")
	assert.Contains(t, out, "depends on the order on disk")
}
