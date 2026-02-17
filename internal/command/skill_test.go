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
	"go.octolab.org/toolset/maintainer/internal/pkg/buildinfo"
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

	out, err = run(t, "skill", "show", "--skill", "review")
	require.NoError(t, err)
	assert.Contains(t, out, "name: review")
	for _, cmd := range [][]string{{"show"}, {"info"}, {"install", "--agent", "codex"}, {"status"}, {"export", "-o", "x"}} {
		_, err = run(t, append(append([]string{"skill"}, cmd...), "--skill", "bogus")...)
		assert.Equal(t, exit.User, code(err), "an unknown skill is a usage error: %v", cmd)
	}

	_, err = run(t, "skill", "install")
	assert.Equal(t, exit.User, code(err), "a missing required flag is a usage error")
	_, err = run(t, "skill", "bogus")
	assert.Equal(t, exit.User, code(err))
}

// records decodes one JSON object per line.
func records[T any](t *testing.T, out string) []T {
	t.Helper()
	var list []T
	for _, line := range bytes.Split(bytes.TrimSpace([]byte(out)), []byte("\n")) {
		var v T
		require.NoError(t, json.Unmarshal(line, &v))
		list = append(list, v)
	}
	return list
}

func TestSkill_info(t *testing.T) {
	out, err := run(t, "skill", "info", "--json")
	require.NoError(t, err)
	infos := records[map[string]string](t, out)
	require.Len(t, infos, len(skills.Names), "one object per skill")
	for i, info := range infos {
		digest, err := skill.Digest(skills.All()[i])
		require.NoError(t, err)
		assert.Equal(t, skills.Names[i], info["name"])
		assert.Equal(t, digest, info["digest"])
		assert.Equal(t, skill.DigestAlgorithm, info["digest_algorithm"])
		version, ok := skill.ParseVersion(info["version"])
		require.True(t, ok, info["version"])
		// A stable release declares >=X.Y.Z <..., a prerelease exactly =X.Y.Z-rcN.
		assert.NoError(t, skill.CheckRange(version, info["tool_version_range"]))
	}

	out, err = run(t, "skill", "info", "--skill", "review")
	require.NoError(t, err)
	assert.Contains(t, out, "review skill:")
	assert.NotContains(t, out, "maintainer skill:")
	assert.Contains(t, out, "binary:")
}

func TestSkill_export(t *testing.T) {
	sandbox(t)
	out, err := run(t, "skill", "export", "-o", "skills")
	require.NoError(t, err)
	for _, name := range skills.Names {
		assert.Contains(t, out, "installed the "+name+" skill")
		got, err := skill.CopyDigest(os.DirFS(filepath.Join("skills", name)))
		require.NoError(t, err)
		want, err := skill.Digest(skills.Skill(name))
		require.NoError(t, err)
		assert.Equal(t, want, got, name)
	}
	out, err = run(t, "skill", "export", "-o", "one", "--skill", "review")
	require.NoError(t, err)
	assert.Contains(t, out, filepath.Join("one", "review"))
	assert.NoDirExists(t, filepath.Join("one", "maintainer"))

	// -o of an earlier release named the skill's own folder: a copy there is
	// refused, not filled with more skills.
	copyDir := filepath.Join("skills", "maintainer")
	before, err := skill.CopyDigest(os.DirFS(copyDir))
	require.NoError(t, err)
	out, err = run(t, "skill", "export", "-o", copyDir, "--replace")
	assert.Equal(t, exit.Transport, code(err))
	assert.ErrorContains(t, err, "holds a skill itself; pass the directory that holds the skills instead")
	assert.NoDirExists(t, filepath.Join(copyDir, "maintainer"))
	after, err := skill.CopyDigest(os.DirFS(copyDir))
	require.NoError(t, err)
	assert.Equal(t, before, after)
	require.NoError(t, os.Remove(filepath.Join(copyDir, "SKILL.md")))
	_, err = run(t, "skill", "export", "-o", copyDir)
	assert.ErrorContains(t, err, "holds a skill itself", "the marker alone marks a copy")
	assert.Empty(t, out)
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
	copies := records[skill.Copy](t, out)
	require.Len(t, copies, 2*len(skills.Names), "every skill for both agents")
	for _, c := range copies {
		assert.Equal(t, "exact", c.ReleaseMatch)
		assert.Equal(t, "intact", c.Integrity)
		assert.Equal(t, "maintainer", c.Ownership)
		assert.Equal(t, "user", c.Scope)
	}
	out, err = run(t, "skill", "status", "--json", "--skill", "review")
	require.NoError(t, err)
	assert.Len(t, records[skill.Copy](t, out), 2)

	_, err = run(t, "skill", "install", "--agent", "claude-code", "--project")
	require.NoError(t, err)
	out, err = run(t, "skill", "status", "--agent", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "Claude Code loads 2 copies")

	_, err = run(t, "skill", "install", "--agent", "cursor")
	assert.Equal(t, exit.User, code(err))
}

// A binary built from an untagged commit reports a pseudo-version: the
// compatibility of its copies is unknown, and status does not fail on it.
func TestSkill_statusPseudoVersion(t *testing.T) {
	sandbox(t)
	original := buildinfo.Version
	t.Cleanup(func() { buildinfo.Version = original })
	buildinfo.Version = "v0.5.1-0.20261125120000-1848546abcde"

	_, err := run(t, "skill", "install", "--agent", "claude-code")
	require.NoError(t, err)
	out, err := run(t, "skill", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "compatibility unknown, same as this release")
	assert.NotContains(t, out, "incompatible")
	assert.NotContains(t, out, "--replace")

	_, err = run(t, "skill", "status", "--exact", "--agent", "claude-code,codex")
	assert.Equal(t, exit.Transport, code(err), "a missing copy for --agent still fails")
}

// plugin unpacks the named skills of this binary into a plugin directory,
// as the catalog lays them out.
func plugin(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		target := filepath.Join(dir, "skills", name)
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
		require.NoError(t, os.CopyFS(target, skills.Skill(name)))
	}
}

func TestSkill_statusCodexPlugin(t *testing.T) {
	home := sandbox(t)
	codex := filepath.Join(home, ".codex")
	dir := filepath.Join(codex, "plugins", "cache", "octolab", "maintainer", "0.2.2")
	plugin(t, dir, skills.Names...)
	require.NoError(t, os.WriteFile(filepath.Join(codex, "config.toml"),
		[]byte("[plugins.\"maintainer@octolab\"]\nenabled = true\n"), 0o644))

	out, err := run(t, "skill", "status", "--agent", "codex", "--exact")
	require.NoError(t, err, "the plugin copies are the agent's copies")
	assert.Contains(t, out, "Codex plugin: maintainer@octolab ~/.codex/plugins/cache/octolab/maintainer/0.2.2/skills/maintainer")
	assert.Contains(t, out, "Codex plugin: maintainer@octolab ~/.codex/plugins/cache/octolab/maintainer/0.2.2/skills/review")
	assert.Contains(t, out, "same as this release, installed by another tool")
	assert.NotContains(t, out, "holds no")

	_, err = run(t, "skill", "install", "--agent", "codex", "--skill", "maintainer")
	require.NoError(t, err)
	out, err = run(t, "skill", "status", "--agent", "codex")
	require.NoError(t, err)
	assert.Contains(t, out, "Codex loads 2 copies of the maintainer skill")
	assert.NotContains(t, out, "copies of the review skill")
}

// A plugin of an earlier release holds the maintainer skill alone: only
// --exact fails on it.
func TestSkill_statusPluginWithoutReview(t *testing.T) {
	for _, agent := range []string{"claude-code", "codex"} {
		t.Run(agent, func(t *testing.T) {
			home := sandbox(t)
			var dir string
			if agent == "codex" {
				dir = filepath.Join(home, ".codex", "plugins", "cache", "octolab", "maintainer", "0.5.2")
				require.NoError(t, os.MkdirAll(filepath.Join(home, ".codex"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(home, ".codex", "config.toml"),
					[]byte("[plugins.\"maintainer@octolab\"]\n"), 0o644))
			} else {
				dir = filepath.Join(home, ".claude", "plugins", "cache", "octolab", "maintainer", "0.5.2")
				claudePlugin(t, home, dir)
			}
			plugin(t, dir, "maintainer")

			out, err := run(t, "skill", "status", "--agent", agent)
			require.NoError(t, err, "a plugin without the review skill is a note")
			assert.Regexp(t, `plugin maintainer@octolab holds no review skill; update the plugin with (Claude Code|Codex)'s plugin manager`, out)
			assert.NotContains(t, out, "has no copy of the review skill", "the plugin note covers it")

			_, err = run(t, "skill", "status", "--agent", agent, "--exact")
			assert.Equal(t, exit.Transport, code(err), "the plugin must hold every skill of the binary")
			assert.ErrorContains(t, err, "holds no review skill")

			_, err = run(t, "skill", "status", "--agent", agent, "--exact", "--skill", "maintainer")
			assert.ErrorContains(t, err, "holds no review skill", "--skill does not narrow what a plugin must hold")

			_, err = run(t, "skill", "status", "--agent", agent, "--skill", "review")
			assert.Equal(t, exit.Transport, code(err), "a skill named by --skill must have a copy")
		})
	}
}

func TestSkill_statusClaudePlugin(t *testing.T) {
	home := sandbox(t)
	dir := filepath.Join(home, ".claude", "plugins", "cache", "octolab", "maintainer", "0.6.0")
	claudePlugin(t, home, dir)
	plugin(t, dir, skills.Names...)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "skills", "extra"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "skills", "extra", "SKILL.md"), []byte("x"), 0o644))

	out, err := run(t, "skill", "status", "--agent", "claude-code")
	require.NoError(t, err)
	assert.Contains(t, out, "Claude Code plugin: maintainer@octolab ~/.claude/plugins/cache/octolab/maintainer/0.6.0/skills/review")
	assert.Contains(t, out, "Claude Code plugin maintainer@octolab holds the extra skill, which this binary does not ship.")

	_, err = run(t, "skill", "status", "--agent", "claude-code", "--exact")
	assert.ErrorContains(t, err, "holds the extra skill", "a plugin holds exactly the skills of the binary")

	require.NoError(t, os.RemoveAll(filepath.Join(dir, "skills", "extra")))
	out, err = run(t, "skill", "status", "--agent", "claude-code", "--exact", "--json")
	require.NoError(t, err)
	for _, c := range records[skill.Copy](t, out) {
		assert.Equal(t, "plugin", c.Scope)
		assert.Equal(t, skills.Names, c.PluginSkills)
	}
}

// claudePlugin registers a user plugin maintainer@octolab installed in dir.
func claudePlugin(t *testing.T, home, dir string) {
	t.Helper()
	registry := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(registry), 0o755))
	data, err := json.Marshal(map[string]any{"version": 2, "plugins": map[string]any{
		"maintainer@octolab": []map[string]string{{"scope": "user", "installPath": dir}},
	}})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(registry, data, 0o644))
}

// A review skill of another tool is left alone and not checked.
func TestSkill_foreignReview(t *testing.T) {
	home := sandbox(t)
	foreign := filepath.Join(home, ".claude", "skills", "review")
	require.NoError(t, os.MkdirAll(foreign, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(foreign, "SKILL.md"),
		[]byte("---\nname: review\ndescription: Another review.\n---\n"), 0o644))

	out, err := run(t, "skill", "install", "--agent", "claude-code")
	require.NoError(t, err, "a foreign review is skipped")
	assert.Contains(t, out, "left "+foreign+" alone: it holds a review skill of another tool")
	assert.Contains(t, out, "installed the maintainer skill")
	assert.Contains(t, out, "/maintainer in Claude Code")
	assert.NotContains(t, out, "/review in")

	out, err = run(t, "skill", "install", "--agent", "claude-code", "--skill", "review")
	assert.Equal(t, exit.Transport, code(err), "a review named by --skill is refused")
	assert.Contains(t, out, "was not installed by maintainer")

	out, err = run(t, "skill", "status", "--agent", "claude-code", "--exact")
	require.NoError(t, err, "the foreign review is not a copy")
	assert.Contains(t, out, "Claude Code user: ~/.claude/skills/review\n  a review skill of another tool, not checked")
	assert.NotContains(t, out, "has no copy of the review skill")
	out, err = run(t, "skill", "status", "--json")
	require.NoError(t, err)
	assert.Len(t, records[skill.Copy](t, out), 1)

	// A missing review next to the maintainer skill is a note.
	require.NoError(t, os.RemoveAll(foreign))
	out, err = run(t, "skill", "status", "--agent", "claude-code", "--exact")
	require.NoError(t, err)
	assert.Contains(t, out, "Claude Code has no copy of the review skill; add it: maintainer skill install --agent claude-code --skill review")
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

	out, err := run(t, "skill", "status", "--agent", "codex", "--skill", "maintainer")
	require.NoError(t, err, "without --exact an ambiguous copy is only reported")
	assert.Contains(t, out, "Codex plugin: maintainer@octolab")
	assert.Contains(t, out, "match with this release unknown")
	assert.Contains(t, out, "Codex's pick among 1.10.0, 1.11x, 1.9.0 depends on the order on disk; remove the stale ones")

	_, err = run(t, "skill", "status", "--agent", "codex", "--exact", "--skill", "maintainer")
	assert.Equal(t, exit.Transport, code(err), "an ambiguous copy is never an exact match")

	out, err = run(t, "skill", "status", "--agent", "codex", "--json", "--skill", "maintainer")
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

	out, err := run(t, "skill", "status", "--agent", "codex", "--json", "--skill", "maintainer")
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

	out, err = run(t, "skill", "status", "--agent", "codex", "--exact", "--skill", "maintainer")
	assert.Equal(t, exit.Transport, code(err), "an exact user copy does not hide the ambiguity")
	assert.Contains(t, out, "unreadable SKILL.md")
	assert.Contains(t, out, "depends on the order on disk")
}
