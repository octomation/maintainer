package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/pkg/buildinfo"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/skill"
	"go.octolab.org/toolset/maintainer/skills"
)

var agentNames = map[string]string{skill.ClaudeCode: "Claude Code", skill.Codex: "Codex"}

// bundle is a skill that ships with this binary.
type bundle struct {
	fsys    fs.FS
	name    string
	tool    string
	version string
	rng     string
	digest  string
}

func loadBundle(fsys fs.FS) (bundle, error) {
	b := bundle{fsys: fsys}
	text, err := fs.ReadFile(fsys, "SKILL.md")
	if err != nil {
		return b, err
	}
	fm, _, err := skill.ParseFrontmatter(string(text))
	if err != nil {
		return b, err
	}
	b.name, _ = fm.String("name")
	meta, _ := fm.Map("metadata")
	b.tool, b.version, b.rng = meta["tool"], meta["version"], meta["tool-version-range"]
	b.digest, err = skill.Digest(fsys)
	return b, err
}

// loadBundles reads the skills of this binary, the main skill first.
func loadBundles(list []fs.FS) ([]bundle, error) {
	all := make([]bundle, 0, len(list))
	for _, fsys := range list {
		b, err := loadBundle(fsys)
		if err != nil {
			return nil, err
		}
		all = append(all, b)
	}
	if len(all) == 0 {
		return nil, errors.New("the binary ships no skill")
	}
	return all, nil
}

// pick returns the skills named, in the binary's order; no names means all.
func pick(all []bundle, names []string) ([]bundle, error) {
	known := make([]string, 0, len(all))
	for _, b := range all {
		known = append(known, b.name)
	}
	for _, name := range names {
		if !contains(known, name) {
			return nil, exit.WithUser(fmt.Errorf("unknown skill %q; use %s", name, strings.Join(known, " or ")))
		}
	}
	if len(names) == 0 {
		return all, nil
	}
	var list []bundle
	for _, b := range all {
		if contains(names, b.name) {
			list = append(list, b)
		}
	}
	return list, nil
}

var skillNames = strings.Join(skills.Names, ", ")

func skillCommand(list ...fs.FS) *cobra.Command {
	command := cobra.Command{
		Use:   "skill",
		Short: "show or install the agent skills for this release",
		Long: "Shows or installs the agent skills for this release.\n\n" +
			"The maintainer skill teaches coding agents such as Claude Code and Codex to\n" +
			"use maintainer; the review skill writes review reports in a fixed form. Both\n" +
			"ship inside the binary and match its version. These commands work offline,\n" +
			"need no token and never read .env.",
		// Overrides the root hook: no .env, no token, no GitHub.
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	}
	load := func() ([]bundle, error) { return loadBundles(list) }
	command.AddCommand(
		skillShowCommand(load),
		skillInfoCommand(load),
		skillExportCommand(load),
		skillInstallCommand(load),
		skillStatusCommand(load),
	)
	return &command
}

func skillShowCommand(load func() ([]bundle, error)) *cobra.Command {
	var file, name string
	command := cobra.Command{
		Use:   "show",
		Short: "print a skill's SKILL.md or one of its files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, err := load()
			if err != nil {
				return err
			}
			picked, err := pick(all, []string{name})
			if err != nil {
				return err
			}
			b := picked[0]
			data, err := fs.ReadFile(b.fsys, file)
			if err != nil {
				files, _ := skill.Files(b.fsys)
				return exit.WithUser(fmt.Errorf("no file %q in the %s skill; choose one of: %s", file, b.name, strings.Join(files, ", ")))
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
	command.Flags().StringVar(&file, "file", "SKILL.md", "file to print, e.g. references/cli.md")
	command.Flags().StringVar(&name, "skill", skills.Names[0], "skill to print: "+skillNames)
	return &command
}

func skillInfoCommand(load func() ([]bundle, error)) *cobra.Command {
	var (
		names  []string
		asJSON bool
	)
	command := cobra.Command{
		Use:   "info",
		Short: "print the version, compatible range and digest of each skill",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, err := load()
			if err != nil {
				return err
			}
			picked, err := pick(all, names)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := newEncoder(out)
				for _, b := range picked {
					err := enc.Encode(struct {
						Name            string `json:"name"`
						Version         string `json:"version"`
						Range           string `json:"tool_version_range"`
						Digest          string `json:"digest"`
						DigestAlgorithm string `json:"digest_algorithm"`
						BinaryVersion   string `json:"binary_version"`
						BinaryCommit    string `json:"binary_commit"`
					}{b.name, b.version, b.rng, b.digest, skill.DigestAlgorithm, buildinfo.Version, buildinfo.Commit})
					if err != nil {
						return err
					}
				}
				return nil
			}
			for _, b := range picked {
				_, _ = fmt.Fprintf(out, "%s skill:\n  version     : %s\n  range       : %s\n  digest      : %s\n",
					b.name, b.version, b.rng, b.digest)
			}
			_, err = fmt.Fprintf(out, "binary:\n  version     : %s\n  git hash    : %s\n", buildinfo.Version, buildinfo.Commit)
			return err
		},
	}
	command.Flags().StringSliceVar(&names, "skill", nil, "skills to print: "+skillNames+" (default all)")
	command.Flags().BoolVar(&asJSON, "json", false, "print one JSON object per skill")
	return &command
}

func skillExportCommand(load func() ([]bundle, error)) *cobra.Command {
	var (
		output  string
		names   []string
		replace bool
	)
	command := cobra.Command{
		Use:   "export",
		Short: "write the skills into a directory",
		Long: "Write the skills into a directory, each into a folder named after it, with an\n" +
			"ownership marker next to its SKILL.md: -o skills writes skills/maintainer and\n" +
			"skills/review.\n\n" +
			"A new or empty folder is written. --replace replaces a copy that maintainer\n" +
			"installed and nobody changed since; any other folder is left alone. A review\n" +
			"folder of another tool is skipped unless --skill names review. A directory\n" +
			"that holds a skill itself is refused: pass the directory that holds it.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, err := load()
			if err != nil {
				return err
			}
			picked, err := pick(all, names)
			if err != nil {
				return err
			}
			dir, err := filepath.Abs(output)
			if err != nil {
				return err
			}
			// Before v0.6.0, -o named the skill's own folder: writing the
			// skills into such a copy would change it.
			for _, name := range []string{"SKILL.md", skill.MarkerFile} {
				if _, err := os.Lstat(filepath.Join(dir, name)); err == nil {
					return &skill.RefusedError{Reason: fmt.Sprintf(
						"%s holds a skill itself; pass the directory that holds the skills instead: maintainer skill export -o %s",
						dir, filepath.Dir(dir))}
				}
			}
			out, named := cmd.OutOrStdout(), cmd.Flags().Changed("skill")
			failed := false
			for _, b := range picked {
				if _, err := place(out, all[0], b, filepath.Join(dir, b.name), replace, named); err != nil {
					_, _ = fmt.Fprintf(out, "%v\n", err)
					failed = true
				}
			}
			if failed {
				return errors.New("the skills were not exported everywhere")
			}
			return nil
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "", "directory to write the skills into")
	command.Flags().StringSliceVar(&names, "skill", nil, "skills to write: "+skillNames+" (default all)")
	command.Flags().BoolVar(&replace, "replace", false, "replace a copy that maintainer installed earlier")
	_ = command.MarkFlagRequired("output")
	return &command
}

// place writes the skill b into dir. A folder named like an extra skill
// that holds another tool's skill is skipped with a note, unless --skill
// named the skill: then it is refused like any folder of another installer.
// It reports whether the skill is in place.
func place(out io.Writer, main, b bundle, dir string, replace, named bool) (bool, error) {
	if b.name != main.name && !named && occupied(dir) && !skill.OfTool(dir, main.tool) {
		_, err := fmt.Fprintf(out, "left %s alone: it holds a %s skill of another tool\n", dir, b.name)
		return false, err
	}
	marker := skill.Marker{Version: b.version, Commit: buildinfo.Commit, Digest: b.digest}
	outcome, err := skill.Export(b.fsys, dir, marker, replace)
	if err != nil {
		return false, err
	}
	var format string
	switch outcome {
	case skill.Installed:
		format = "installed the %s skill %s into %s\n"
	case skill.Replaced:
		format = "replaced the skill in %[3]s with the %[1]s skill %[2]s\n"
	case skill.Unchanged:
		format = "%[3]s already holds the %[1]s skill %[2]s\n"
	}
	_, err = fmt.Fprintf(out, format, b.name, b.version, dir)
	return true, err
}

// occupied reports whether path exists and is not an empty directory.
func occupied(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return true
	}
	entries, err := os.ReadDir(path)
	return err != nil || len(entries) > 0
}

func skillInstallCommand(load func() ([]bundle, error)) *cobra.Command {
	var (
		agents  []string
		names   []string
		project bool
		replace bool
	)
	command := cobra.Command{
		Use:   "install",
		Short: "install the skills for Claude Code or Codex",
		Long: "Install the skills for Claude Code or Codex: all of them, or the ones --skill\n" +
			"names.\n\n" +
			"Writes ~/.claude/skills/<skill> (or $CLAUDE_CONFIG_DIR/skills/<skill>) for\n" +
			"Claude Code and ~/.agents/skills/<skill> for Codex; with --project, the same\n" +
			"under the current directory. After upgrading maintainer, run it again with\n" +
			"--replace. A directory from another installer is left alone; a review\n" +
			"directory of another tool is skipped unless --skill names review.",
		Example: `  maintainer skill install --agent claude-code
  maintainer skill install --agent claude-code,codex --replace
  maintainer skill install --agent codex --skill review`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, agent := range agents {
				if _, ok := agentNames[agent]; !ok {
					return exit.WithUser(fmt.Errorf("unknown agent %q; use %s", agent, strings.Join(skill.Agents, " or ")))
				}
			}
			all, err := load()
			if err != nil {
				return err
			}
			picked, err := pick(all, names)
			if err != nil {
				return err
			}
			env, err := skill.LookupEnv()
			if err != nil {
				return err
			}
			out, named := cmd.OutOrStdout(), cmd.Flags().Changed("skill")
			var (
				failed bool
				loaded []string
			)
			for _, agent := range agents {
				for _, b := range picked {
					dir := env.Dir(agent, b.name, project)
					_, _ = fmt.Fprintf(out, "%s: %s\n", agentNames[agent], dir)
					ok, err := place(out, all[0], b, dir, replace, named)
					if err != nil {
						_, _ = fmt.Fprintf(out, "%v\n", err)
						failed = true
						continue
					}
					if ok && !contains(loaded, b.name) {
						loaded = append(loaded, b.name)
					}
				}
			}
			if failed {
				return errors.New("the skills were not installed everywhere")
			}
			if len(loaded) == 0 {
				return nil
			}
			_, err = fmt.Fprintf(out, "Start a new agent session to load the skills: /%s in Claude Code, $%s in Codex.\n",
				strings.Join(loaded, " and /"), strings.Join(loaded, " and $"))
			return err
		},
	}
	command.Flags().StringSliceVar(&agents, "agent", nil, "agents to install for: claude-code, codex")
	command.Flags().StringSliceVar(&names, "skill", nil, "skills to install: "+skillNames+" (default all)")
	command.Flags().BoolVar(&project, "project", false, "install into the current directory instead of the user's skills")
	command.Flags().BoolVar(&replace, "replace", false, "replace a copy that maintainer installed earlier")
	_ = command.MarkFlagRequired("agent")
	return &command
}

// pluginGap is what a plugin of the tool lacks or holds beyond the skills of
// this binary.
type pluginGap struct {
	plugin         skill.Plugin
	missing, extra []string
}

func skillStatusCommand(load func() ([]bundle, error)) *cobra.Command {
	var (
		agents []string
		names  []string
		exact  bool
		asJSON bool
	)
	command := cobra.Command{
		Use:   "status",
		Short: "check the installed copies of the skills against this binary",
		Long: "Check the installed copies of the skills against this binary.\n\n" +
			"Looks in the user's and the project's skills of Claude Code and Codex, and\n" +
			"in their maintainer plugins. Fails when a copy is incompatible with this\n" +
			"binary, was changed after maintainer installed it, or no copy exists for a\n" +
			"requested --agent: of the maintainer skill, or of a skill --skill names.\n" +
			"With --exact, also when a copy differs from this release or a plugin does\n" +
			"not hold exactly the skills of this binary. Duplicate copies, a review\n" +
			"folder of another tool and a missing review skill are only reported.\n" +
			"A dev or pseudo-version binary leaves compatibility unknown.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, agent := range agents {
				if _, ok := agentNames[agent]; !ok {
					return exit.WithUser(fmt.Errorf("unknown agent %q; use %s", agent, strings.Join(skill.Agents, " or ")))
				}
			}
			all, err := load()
			if err != nil {
				return err
			}
			picked, err := pick(all, names)
			if err != nil {
				return err
			}
			env, err := skill.LookupEnv()
			if err != nil {
				return err
			}
			if len(agents) == 0 {
				agents = skill.Agents
			}
			main := all[0]
			var (
				copies  []skill.Copy
				foreign []skill.Location
			)
			for _, b := range picked {
				for _, loc := range skill.Locations(env, main.tool, b.name) {
					switch {
					case !contains(agents, loc.Agent):
					case b.name != main.name && loc.Scope != "plugin" && !skill.OfTool(loc.Path, main.tool):
						foreign = append(foreign, loc)
					default:
						copies = append(copies, skill.Inspect(loc, b.digest, buildinfo.Version))
					}
				}
			}
			var gaps []pluginGap
			for _, p := range skill.Plugins(env, main.tool) {
				if !contains(agents, p.Agent) {
					continue
				}
				// A plugin holds exactly the skills of this binary, whatever
				// --skill picks.
				g := pluginGap{plugin: p}
				for _, b := range all {
					if !contains(p.Skills, b.name) {
						g.missing = append(g.missing, b.name)
					}
				}
				for _, name := range p.Skills {
					if _, err := pick(all, []string{name}); err != nil {
						g.extra = append(g.extra, name)
					}
				}
				if len(g.missing)+len(g.extra) > 0 {
					gaps = append(gaps, g)
				}
			}
			// A requested agent needs the maintainer skill and the skills
			// --skill names; other skills it lacks are only reported.
			required := []string{main.name}
			if cmd.Flags().Changed("skill") {
				required = names
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := newEncoder(out)
				for _, c := range copies {
					if err := enc.Encode(c); err != nil {
						return err
					}
				}
			} else {
				printStatus(out, picked, required, env, copies, foreign, gaps, agents)
			}
			return statusError(copies, gaps, agents, required, exact, cmd.Flags().Changed("agent"))
		},
	}
	command.Flags().StringSliceVar(&agents, "agent", nil, "agents to check: claude-code, codex (default all)")
	command.Flags().StringSliceVar(&names, "skill", nil, "skills to check: "+skillNames+" (default all)")
	command.Flags().BoolVar(&exact, "exact", false, "also fail when a copy differs from this release or a plugin lacks a skill")
	command.Flags().BoolVar(&asJSON, "json", false, "print one JSON record per copy")
	return &command
}

func printStatus(out io.Writer, picked []bundle, required []string, env skill.Env,
	copies []skill.Copy, foreign []skill.Location, gaps []pluginGap, agents []string) {
	for _, b := range picked {
		_, _ = fmt.Fprintf(out, "maintainer %s ships the %s skill %s, %s\n", buildinfo.Version, b.name, b.version, b.digest)
	}
	if len(copies) == 0 {
		_, _ = fmt.Fprintln(out, "\nNo installed copies. Install them: maintainer skill install --agent claude-code,codex")
	}
	home := func(path string) string { return strings.Replace(path, env.Home, "~", 1) }
	perSkill := make(map[[2]string]int)
	for _, c := range copies {
		perSkill[[2]string{c.Agent, c.Skill}]++
		where := home(c.Path)
		if c.Plugin != "" {
			where = c.Plugin + " " + where
		}
		if c.Symlink {
			where += " (symbolic link)"
		}
		_, _ = fmt.Fprintf(out, "\n%s %s: %s\n  %s\n", agentNames[c.Agent], c.Scope, where, describe(c))
		if len(c.AmbiguousVersions) > 0 {
			_, _ = fmt.Fprintf(out, "  Codex's pick among %s depends on the order on disk; remove the stale ones\n",
				strings.Join(c.AmbiguousVersions, ", "))
		}
		switch {
		case c.Integrity == "modified":
			_, _ = fmt.Fprintln(out, "  changed after maintainer installed it; move it away and run maintainer skill install")
		case c.Compatibility == "incompatible" && c.Ownership == "maintainer":
			_, _ = fmt.Fprintf(out, "  update it: maintainer skill install --agent %s --replace\n", c.Agent)
		case c.Compatibility == "incompatible":
			_, _ = fmt.Fprintln(out, "  update it with the tool that installed it, or remove it and run maintainer skill install")
		}
	}
	for _, loc := range foreign {
		_, _ = fmt.Fprintf(out, "\n%s %s: %s\n  a %s skill of another tool, not checked\n",
			agentNames[loc.Agent], loc.Scope, home(loc.Path), loc.Skill)
	}
	for _, g := range gaps {
		p := g.plugin
		for _, name := range g.missing {
			_, _ = fmt.Fprintf(out, "\n%s plugin %s holds no %s skill; update the plugin with %s's plugin manager.\n",
				agentNames[p.Agent], p.ID, name, agentNames[p.Agent])
		}
		for _, name := range g.extra {
			_, _ = fmt.Fprintf(out, "\n%s plugin %s holds the %s skill, which this binary does not ship.\n",
				agentNames[p.Agent], p.ID, name)
		}
	}
	// A skill that is not required is a note when it is missing for an agent
	// that has a required one; a plugin that lacks it is reported above.
	for _, agent := range agents {
		has := false
		for _, name := range required {
			has = has || perSkill[[2]string{agent, name}] > 0
		}
		for _, g := range gaps {
			has = has && g.plugin.Agent != agent
		}
		for _, b := range picked {
			if !has || contains(required, b.name) || perSkill[[2]string{agent, b.name}] > 0 || isForeign(foreign, agent, b.name) {
				continue
			}
			_, _ = fmt.Fprintf(out, "\n%s has no copy of the %s skill; add it: maintainer skill install --agent %s --skill %s\n",
				agentNames[agent], b.name, agent, b.name)
		}
	}
	for _, agent := range skill.Agents {
		for _, b := range picked {
			if n := perSkill[[2]string{agent, b.name}]; n > 1 {
				_, _ = fmt.Fprintf(out, "\n%s loads %d copies of the %s skill; keep one to avoid mixing versions.\n",
					agentNames[agent], n, b.name)
			}
		}
	}
}

func isForeign(foreign []skill.Location, agent, name string) bool {
	for _, loc := range foreign {
		if loc.Agent == agent && loc.Skill == name {
			return true
		}
	}
	return false
}

func describe(c skill.Copy) string {
	facts := []string{"skill " + c.Version}
	if c.Version == "" {
		facts[0] = "unreadable SKILL.md"
	}
	facts = append(facts, map[string]string{
		"compatible":   "compatible with this binary",
		"incompatible": "incompatible with this binary",
		"unknown":      "compatibility unknown",
	}[c.Compatibility])
	switch c.ReleaseMatch {
	case "exact":
		facts = append(facts, "same as this release")
	case "unknown":
		facts = append(facts, "match with this release unknown")
	default:
		facts = append(facts, "differs from this release")
	}
	facts = append(facts, map[string]string{
		"maintainer": "installed by maintainer",
		"external":   "installed by another tool",
		"unknown":    "owner unknown",
	}[c.Ownership])
	if c.Integrity != "unknown" {
		facts = append(facts, c.Integrity)
	}
	return strings.Join(facts, ", ")
}

func statusError(copies []skill.Copy, gaps []pluginGap, agents, required []string, exact, requested bool) error {
	var problems []string
	found := make(map[[2]string]bool)
	for _, c := range copies {
		found[[2]string{c.Agent, c.Skill}] = true
		switch {
		case c.Compatibility == "incompatible":
			problems = append(problems, c.Path+" is incompatible with this binary")
		case c.Integrity == "modified":
			problems = append(problems, c.Path+" was changed after maintainer installed it")
		case exact && len(c.AmbiguousVersions) > 0:
			problems = append(problems, c.Path+" may not be the copy Codex loads")
		case exact && c.ReleaseMatch != "exact":
			problems = append(problems, c.Path+" differs from this release")
		}
	}
	if exact {
		for _, g := range gaps {
			for _, name := range g.missing {
				problems = append(problems, fmt.Sprintf("the %s plugin %s holds no %s skill", agentNames[g.plugin.Agent], g.plugin.ID, name))
			}
			for _, name := range g.extra {
				problems = append(problems, fmt.Sprintf("the %s plugin %s holds the %s skill, which this binary does not ship", agentNames[g.plugin.Agent], g.plugin.ID, name))
			}
		}
	}
	if requested {
		for _, agent := range agents {
			for _, name := range required {
				if !found[[2]string{agent, name}] {
					problems = append(problems, fmt.Sprintf("no copy of the %s skill for %s", name, agentNames[agent]))
				}
			}
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return errors.New(strings.Join(problems, "; "))
}

func newEncoder(w io.Writer) *json.Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
