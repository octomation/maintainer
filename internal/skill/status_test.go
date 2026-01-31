package skill_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/skill"
)

func writeSkill(t *testing.T, dir, text string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(text), 0o644))
}

func TestLocations(t *testing.T) {
	root := t.TempDir()
	env := Env{
		Home:      filepath.Join(root, "home"),
		ClaudeDir: filepath.Join(root, "claude"),
		CodexHome: filepath.Join(root, "codex"),
		Cwd:       filepath.Join(root, "repo", "sub"),
	}
	require.NoError(t, os.MkdirAll(filepath.Join(root, "repo", ".git"), 0o755))
	for _, dir := range []string{
		filepath.Join(env.ClaudeDir, "skills", "maintainer"),
		filepath.Join(root, "repo", ".claude", "skills", "maintainer"),
		filepath.Join(env.Cwd, ".agents", "skills", "maintainer"),
		filepath.Join(env.Home, ".agents", "skills", "maintainer"),
		filepath.Join(env.CodexHome, "skills", "maintainer"),
		filepath.Join(root, "cache", "user", "skills", "maintainer"),
		filepath.Join(root, "cache", "here", "skills", "maintainer"),
		filepath.Join(root, "cache", "elsewhere", "skills", "maintainer"),
		filepath.Join(root, "cache", "other", "skills", "maintainer"),
	} {
		writeSkill(t, dir, "x")
	}
	registry := map[string]any{"version": 2, "plugins": map[string]any{
		"maintainer@octolab": []map[string]string{
			{"scope": "user", "installPath": filepath.Join(root, "cache", "user")},
			{"scope": "project", "projectPath": filepath.Join(root, "repo"), "installPath": filepath.Join(root, "cache", "here")},
			{"scope": "project", "projectPath": filepath.Join(root, "elsewhere"), "installPath": filepath.Join(root, "cache", "elsewhere")},
		},
		"other@octolab": []map[string]string{{"scope": "user", "installPath": filepath.Join(root, "cache", "other")}},
	}}
	data, err := json.Marshal(registry)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(env.ClaudeDir, "plugins"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(env.ClaudeDir, "plugins", "installed_plugins.json"), data, 0o644))

	var got [][3]string
	for _, loc := range Locations(env, "maintainer") {
		rel, err := filepath.Rel(root, loc.Path)
		require.NoError(t, err)
		got = append(got, [3]string{loc.Agent, loc.Scope, rel})
	}
	assert.Equal(t, [][3]string{
		{ClaudeCode, "user", "claude/skills/maintainer"},
		{ClaudeCode, "project", "repo/.claude/skills/maintainer"},
		{ClaudeCode, "plugin", "cache/here/skills/maintainer"},
		{ClaudeCode, "plugin", "cache/user/skills/maintainer"},
		{Codex, "user", "home/.agents/skills/maintainer"},
		{Codex, "project", "repo/sub/.agents/skills/maintainer"},
		{Codex, "legacy", "codex/skills/maintainer"},
	}, got)
}

func TestLocations_codexPlugins(t *testing.T) {
	root := t.TempDir()
	env := Env{
		Home:      filepath.Join(root, "home"),
		ClaudeDir: filepath.Join(root, "claude"),
		CodexHome: filepath.Join(root, "codex"),
		Cwd:       filepath.Join(root, "home"),
	}
	cache := filepath.Join(env.CodexHome, "plugins", "cache")
	for _, dir := range []string{
		filepath.Join(cache, "octolab", "maintainer", "0.2.1", "skills", "maintainer"),
		filepath.Join(cache, "octolab", "maintainer", "0.2.10", "skills", "maintainer"),
		filepath.Join(cache, "octolab", "maintainer", "0.2.2", "skills", "maintainer"),
		filepath.Join(cache, "off", "maintainer", "0.2.2", "skills", "maintainer"),
		filepath.Join(cache, "octolab", "other", "0.1.0", "skills", "maintainer"),
	} {
		writeSkill(t, dir, "x")
	}
	config := `model = "gpt"

[marketplaces.octolab]
source_type = "git"
source = "https://github.com/octolab/skills.git"

[plugins]
auto_update = true

[plugins."maintainer@octolab"]

[plugins."maintainer@off"]
enabled = false

[plugins."maintainer@missing"]
enabled = true

[plugins."maintainer@.."]
enabled = true

[plugins."other@octolab"]
enabled = true
`
	require.NoError(t, os.MkdirAll(env.CodexHome, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(env.CodexHome, "config.toml"), []byte(config), 0o644))
	writeSkill(t, filepath.Join(env.CodexHome, "plugins", "maintainer", "0.9.0", "skills", "maintainer"), "x")

	locations := Locations(env, "maintainer")
	require.Len(t, locations, 1, "only the enabled plugin named like the skill, in the version Codex loads")
	assert.Equal(t, Location{
		Agent:  Codex,
		Scope:  "plugin",
		Path:   filepath.Join(cache, "octolab", "maintainer", "0.2.10", "skills", "maintainer"),
		Plugin: "maintainer@octolab",
	}, locations[0], "a table without enabled is enabled; a non-table value and a dot marketplace are skipped")

	require.NoError(t, os.WriteFile(filepath.Join(env.CodexHome, "config.toml"), []byte("not = [toml"), 0o644))
	assert.Empty(t, Locations(env, "maintainer"), "an unreadable config lists no plugin")
}

func TestLocations_home(t *testing.T) {
	home := t.TempDir()
	env := Env{Home: home, ClaudeDir: filepath.Join(home, ".claude"), CodexHome: filepath.Join(home, ".codex"), Cwd: home}
	writeSkill(t, filepath.Join(home, ".claude", "skills", "maintainer"), "x")
	locations := Locations(env, "maintainer")
	require.Len(t, locations, 1)
	assert.Equal(t, "user", locations[0].Scope)
}

func TestInspect(t *testing.T) {
	v1, m1 := release(t, "0.2.0", skillMD)
	dir := filepath.Join(t.TempDir(), "maintainer")
	_, err := Export(v1, dir, m1, false)
	require.NoError(t, err)
	loc := Location{Agent: ClaudeCode, Scope: "user", Path: dir}

	c := Inspect(loc, m1.Digest, "0.2.3")
	assert.Equal(t, "0.2.0", c.Version)
	assert.Equal(t, ">=0.2.0 <0.3.0", c.Range)
	assert.Equal(t, "exact", c.ReleaseMatch)
	assert.Equal(t, "compatible", c.Compatibility)
	assert.Equal(t, "intact", c.Integrity)
	assert.Equal(t, "maintainer", c.Ownership)

	assert.Equal(t, "incompatible", Inspect(loc, m1.Digest, "0.3.0").Compatibility)
	assert.Equal(t, "unknown", Inspect(loc, m1.Digest, "dev").Compatibility)
	assert.Equal(t, "different", Inspect(loc, "sha256:other", "0.2.0").ReleaseMatch)

	require.NoError(t, os.WriteFile(filepath.Join(dir, "references", "a.md"), []byte("b"), 0o644))
	assert.Equal(t, "modified", Inspect(loc, m1.Digest, "0.2.0").Integrity)

	link := filepath.Join(t.TempDir(), "maintainer")
	require.NoError(t, os.Symlink(dir, link))
	c = Inspect(Location{Agent: Codex, Scope: "user", Path: link}, m1.Digest, "0.2.0")
	assert.True(t, c.Symlink)
	assert.Equal(t, "external", c.Ownership)
	assert.Equal(t, "unknown", c.Integrity)

	foreign := filepath.Join(t.TempDir(), "maintainer")
	writeSkill(t, foreign, "no frontmatter")
	c = Inspect(Location{Agent: Codex, Scope: "user", Path: foreign}, m1.Digest, "0.2.0")
	assert.Equal(t, "external", c.Ownership)
	assert.Equal(t, "unknown", c.Compatibility)
	assert.Empty(t, c.Version)
}

func TestEnv_Dir(t *testing.T) {
	env := Env{Home: "/h", ClaudeDir: "/c", CodexHome: "/x", Cwd: "/p"}
	assert.Equal(t, "/c/skills/maintainer", env.Dir(ClaudeCode, "maintainer", false))
	assert.Equal(t, "/p/.claude/skills/maintainer", env.Dir(ClaudeCode, "maintainer", true))
	assert.Equal(t, "/h/.agents/skills/maintainer", env.Dir(Codex, "maintainer", false))
	assert.Equal(t, "/p/.agents/skills/maintainer", env.Dir(Codex, "maintainer", true))
}
