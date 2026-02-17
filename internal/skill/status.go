package skill

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Agents that maintainer installs the skill for, named as the skills CLI names
// them.
const (
	ClaudeCode = "claude-code"
	Codex      = "codex"
)

// Agents lists the supported agents.
var Agents = []string{ClaudeCode, Codex}

// Env holds the directories where agents look for skills.
type Env struct {
	Home      string // the user's home directory
	ClaudeDir string // CLAUDE_CONFIG_DIR, or ~/.claude
	CodexHome string // CODEX_HOME, or ~/.codex
	Cwd       string // the current directory
}

// LookupEnv reads Env from the process environment.
func LookupEnv() (Env, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Env{}, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return Env{}, err
	}
	env := Env{Home: home, ClaudeDir: os.Getenv("CLAUDE_CONFIG_DIR"), CodexHome: os.Getenv("CODEX_HOME"), Cwd: cwd}
	if env.ClaudeDir == "" {
		env.ClaudeDir = filepath.Join(home, ".claude")
	}
	if env.CodexHome == "" {
		env.CodexHome = filepath.Join(home, ".codex")
	}
	return env, nil
}

// Dir returns where maintainer skill install writes the skill for the agent:
// the user's skills, or with project set, the current directory's.
func (env Env) Dir(agent, name string, project bool) string {
	switch {
	case agent == ClaudeCode && project:
		return filepath.Join(env.Cwd, ".claude", "skills", name)
	case agent == ClaudeCode:
		return filepath.Join(env.ClaudeDir, "skills", name)
	case project:
		return filepath.Join(env.Cwd, ".agents", "skills", name)
	}
	return filepath.Join(env.Home, ".agents", "skills", name)
}

// Location is a place where an agent may load a copy of a skill.
type Location struct {
	Agent  string `json:"agent"`
	Skill  string `json:"skill"`
	Scope  string `json:"scope"` // user, project, plugin or legacy
	Path   string `json:"path"`
	Plugin string `json:"plugin,omitempty"` // a Claude Code or Codex plugin id
	// PluginSkills lists the skills the plugin holds.
	PluginSkills []string `json:"plugin_skills,omitempty"`
	// AmbiguousVersions lists the version directories of a Codex plugin when
	// which of them Codex loads depends on the order on disk.
	AmbiguousVersions []string `json:"ambiguous_versions,omitempty"`
}

// Plugin is an installed Claude Code or Codex plugin of the tool, in the
// version the agent loads.
type Plugin struct {
	Agent  string
	ID     string
	Path   string   // the plugin's directory, which holds skills/<name>
	Skills []string // the skills under Path, sorted
	// Ambiguous lists the version directories of a Codex plugin when which
	// of them Codex loads depends on the order on disk.
	Ambiguous []string
}

// Plugins lists the plugins of the tool that the agents load: Claude Code's
// user plugins and project plugins from the current directory up to the
// repository root, then Codex's enabled plugins. A plugin is found by the
// tool's name, whatever skills it holds.
func Plugins(env Env, tool string) []Plugin {
	var list []Plugin
	projects := projectDirs(env.Cwd)
	for _, p := range claudePlugins(env.ClaudeDir, tool) {
		if p.Scope == "user" || contains(projects, p.ProjectPath) {
			list = append(list, Plugin{Agent: ClaudeCode, ID: p.ID, Path: p.InstallPath, Skills: pluginSkills(p.InstallPath)})
		}
	}
	for _, p := range codexPlugins(env.CodexHome, tool) {
		list = append(list, Plugin{Agent: Codex, ID: p.ID, Path: p.InstallPath, Skills: pluginSkills(p.InstallPath), Ambiguous: p.Ambiguous})
	}
	return list
}

// pluginSkills lists the directories under the plugin's skills that hold a
// SKILL.md.
func pluginSkills(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(dir, "skills"))
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if _, err := os.Stat(filepath.Join(dir, "skills", e.Name(), "SKILL.md")); err == nil {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

// Locations lists the existing copies of the tool's skill name that the
// agents load: the user's skills, project skills from the current directory
// up to the repository root, the plugins of the tool and the legacy Codex
// directory.
func Locations(env Env, tool, name string) []Location {
	var list []Location
	seen := make(map[string]bool)
	add := func(agent, scope, path string, plugin *Plugin) {
		// In the home directory, project skills are the user's.
		if seen[path] {
			return
		}
		loc := Location{Agent: agent, Skill: name, Scope: scope, Path: path}
		if plugin != nil {
			loc.Plugin, loc.PluginSkills, loc.AmbiguousVersions = plugin.ID, plugin.Skills, plugin.Ambiguous
		}
		// An ambiguous Codex plugin is kept without a readable copy too:
		// another of its versions may be the one Codex loads.
		if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err == nil || len(loc.AmbiguousVersions) > 0 {
			seen[path] = true
			list = append(list, loc)
		}
	}
	projects := projectDirs(env.Cwd)
	plugins := Plugins(env, tool)
	addPlugins := func(agent string) {
		for i := range plugins {
			if p := &plugins[i]; p.Agent == agent {
				add(agent, "plugin", filepath.Join(p.Path, "skills", name), p)
			}
		}
	}

	add(ClaudeCode, "user", filepath.Join(env.ClaudeDir, "skills", name), nil)
	for _, dir := range projects {
		add(ClaudeCode, "project", filepath.Join(dir, ".claude", "skills", name), nil)
	}
	addPlugins(ClaudeCode)

	add(Codex, "user", filepath.Join(env.Home, ".agents", "skills", name), nil)
	for _, dir := range projects {
		add(Codex, "project", filepath.Join(dir, ".agents", "skills", name), nil)
	}
	add(Codex, "legacy", filepath.Join(env.CodexHome, "skills", name), nil)
	addPlugins(Codex)
	return list
}

// projectDirs returns cwd and its parents up to the repository root, or cwd
// alone outside a repository.
func projectDirs(cwd string) []string {
	var dirs []string
	for dir := cwd; ; dir = filepath.Dir(dir) {
		dirs = append(dirs, dir)
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dirs
		}
		if filepath.Dir(dir) == dir {
			return dirs[:1]
		}
	}
}

type claudePlugin struct {
	ID          string
	Scope       string `json:"scope"`
	ProjectPath string `json:"projectPath"`
	InstallPath string `json:"installPath"`
}

// claudePlugins reads the plugins named like the tool from Claude Code's
// installed_plugins.json, read-only.
func claudePlugins(claudeDir, tool string) []claudePlugin {
	data, err := os.ReadFile(filepath.Join(claudeDir, "plugins", "installed_plugins.json"))
	if err != nil {
		return nil
	}
	var registry struct {
		Plugins map[string][]claudePlugin `json:"plugins"`
	}
	if json.Unmarshal(data, &registry) != nil {
		return nil
	}
	var list []claudePlugin
	for id, installs := range registry.Plugins {
		if plugin, _, _ := strings.Cut(id, "@"); plugin != tool {
			continue
		}
		for _, p := range installs {
			if p.InstallPath != "" {
				p.ID = id
				list = append(list, p)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID+list[i].InstallPath < list[j].ID+list[j].InstallPath })
	return list
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

// Copy describes one installed copy of the skill.
type Copy struct {
	Location
	Version       string `json:"version,omitempty"`
	Range         string `json:"tool_version_range,omitempty"`
	ReleaseMatch  string `json:"release_match"` // exact, different, or unknown for an ambiguous copy
	Compatibility string `json:"compatibility"` // compatible, incompatible or unknown
	Integrity     string `json:"integrity"`     // intact, modified or unknown
	Ownership     string `json:"ownership"`     // maintainer, external or unknown
	Symlink       bool   `json:"symlink,omitempty"`
}

// Inspect compares a copy with the skill the binary ships: digest is the
// embedded skill's, binary the binary's version.
func Inspect(loc Location, digest, binary string) Copy {
	c := Copy{
		Location:      loc,
		ReleaseMatch:  "different",
		Compatibility: "unknown",
		Integrity:     "unknown",
		Ownership:     "external",
	}
	if info, err := os.Lstat(loc.Path); err == nil && info.Mode()&fs.ModeSymlink != 0 {
		c.Symlink = true
	}

	current, digestErr := CopyDigest(os.DirFS(loc.Path))
	switch {
	case len(loc.AmbiguousVersions) > 0:
		// The agent may load another copy: never claim an exact match.
		c.ReleaseMatch = "unknown"
	case digestErr == nil && current == digest:
		c.ReleaseMatch = "exact"
	}
	marker, err := ReadMarker(loc.Path)
	switch {
	case err == nil && !c.Symlink:
		c.Ownership = "maintainer"
		c.Integrity = "modified"
		if digestErr == nil && current == marker.Digest {
			c.Integrity = "intact"
		}
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		c.Ownership = "unknown"
	}

	text, err := os.ReadFile(filepath.Join(loc.Path, "SKILL.md"))
	if err != nil {
		return c
	}
	fm, _, err := ParseFrontmatter(string(text))
	if err != nil {
		return c
	}
	meta, _ := fm.Map("metadata")
	c.Version, c.Range = meta["version"], meta["tool-version-range"]
	// dev, a pseudo-version or an unparsable version names no release:
	// compatibility stays unknown.
	v, okV := ParseVersion(binary)
	r, okR := ParseRange(c.Range)
	if okV && okR && !v.Pseudo() {
		c.Compatibility = "incompatible"
		if r.Contains(v) {
			c.Compatibility = "compatible"
		}
	}
	return c
}
