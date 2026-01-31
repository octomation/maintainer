package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"go.octolab.org/toolset/maintainer/internal/pkg/buildinfo"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
	"go.octolab.org/toolset/maintainer/internal/skill"
)

var agentNames = map[string]string{skill.ClaudeCode: "Claude Code", skill.Codex: "Codex"}

// bundle is the skill that ships with this binary.
type bundle struct {
	fsys    fs.FS
	name    string
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
	b.version, b.rng = meta["version"], meta["tool-version-range"]
	b.digest, err = skill.Digest(fsys)
	return b, err
}

func skillCommand(fsys fs.FS) *cobra.Command {
	command := cobra.Command{
		Use:   "skill",
		Short: "show or install the agent skill for this release",
		Long: "Shows or installs the agent skill for this release.\n\n" +
			"The skill teaches coding agents such as Claude Code and Codex to use\n" +
			"maintainer. It ships inside the binary and matches its version. These\n" +
			"commands work offline, need no token and never read .env.",
		// Overrides the root hook: no .env, no token, no GitHub.
		PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	}
	load := func() (bundle, error) { return loadBundle(fsys) }
	command.AddCommand(
		skillShowCommand(load),
		skillInfoCommand(load),
		skillExportCommand(load),
		skillInstallCommand(load),
		skillStatusCommand(load),
	)
	return &command
}

func skillShowCommand(load func() (bundle, error)) *cobra.Command {
	var file string
	command := cobra.Command{
		Use:   "show",
		Short: "print the skill's SKILL.md or one of its files",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := load()
			if err != nil {
				return err
			}
			data, err := fs.ReadFile(b.fsys, file)
			if err != nil {
				files, _ := skill.Files(b.fsys)
				return exit.WithUser(fmt.Errorf("no file %q in the skill; choose one of: %s", file, strings.Join(files, ", ")))
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
	command.Flags().StringVar(&file, "file", "SKILL.md", "file to print, e.g. references/cli.md")
	return &command
}

func skillInfoCommand(load func() (bundle, error)) *cobra.Command {
	var asJSON bool
	command := cobra.Command{
		Use:   "info",
		Short: "print the skill's version, compatible range and digest",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := load()
			if err != nil {
				return err
			}
			if asJSON {
				return newEncoder(cmd.OutOrStdout()).Encode(struct {
					Name            string `json:"name"`
					Version         string `json:"version"`
					Range           string `json:"tool_version_range"`
					Digest          string `json:"digest"`
					DigestAlgorithm string `json:"digest_algorithm"`
					BinaryVersion   string `json:"binary_version"`
					BinaryCommit    string `json:"binary_commit"`
				}{b.name, b.version, b.rng, b.digest, skill.DigestAlgorithm, buildinfo.Version, buildinfo.Commit})
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(),
				"%s skill:\n  version     : %s\n  range       : %s\n  digest      : %s\n  binary      : %s\n  git hash    : %s\n",
				b.name, b.version, b.rng, b.digest, buildinfo.Version, buildinfo.Commit)
			return err
		},
	}
	command.Flags().BoolVar(&asJSON, "json", false, "print one JSON object")
	return &command
}

func skillExportCommand(load func() (bundle, error)) *cobra.Command {
	var (
		output  string
		replace bool
	)
	command := cobra.Command{
		Use:   "export",
		Short: "write the skill into a directory",
		Long: "Write the skill into a directory, with an ownership marker next to SKILL.md.\n\n" +
			"A new or empty directory is written. --replace replaces a copy that maintainer\n" +
			"installed and nobody changed since; any other directory is left alone.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := load()
			if err != nil {
				return err
			}
			dir, err := filepath.Abs(output)
			if err != nil {
				return err
			}
			return export(cmd.OutOrStdout(), b, dir, replace)
		},
	}
	command.Flags().StringVarP(&output, "output", "o", "", "directory to write the skill into")
	command.Flags().BoolVar(&replace, "replace", false, "replace a copy that maintainer installed earlier")
	_ = command.MarkFlagRequired("output")
	return &command
}

func export(out io.Writer, b bundle, dir string, replace bool) error {
	marker := skill.Marker{Version: b.version, Commit: buildinfo.Commit, Digest: b.digest}
	outcome, err := skill.Export(b.fsys, dir, marker, replace)
	if err != nil {
		return err
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
	return err
}

func skillInstallCommand(load func() (bundle, error)) *cobra.Command {
	var (
		agents  []string
		project bool
		replace bool
	)
	command := cobra.Command{
		Use:   "install",
		Short: "install the skill for Claude Code or Codex",
		Long: "Install the skill for Claude Code or Codex.\n\n" +
			"Writes ~/.claude/skills/maintainer (or $CLAUDE_CONFIG_DIR/skills/maintainer) for\n" +
			"Claude Code and ~/.agents/skills/maintainer for Codex; with --project, the same\n" +
			"under the current directory. After upgrading maintainer, run it again with\n" +
			"--replace. A directory from another installer is left alone.",
		Example: `  maintainer skill install --agent claude-code
  maintainer skill install --agent claude-code,codex --replace`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, agent := range agents {
				if _, ok := agentNames[agent]; !ok {
					return exit.WithUser(fmt.Errorf("unknown agent %q; use %s", agent, strings.Join(skill.Agents, " or ")))
				}
			}
			b, err := load()
			if err != nil {
				return err
			}
			env, err := skill.LookupEnv()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			var failed []error
			for _, agent := range agents {
				dir := env.Dir(agent, b.name, project)
				_, _ = fmt.Fprintf(out, "%s: %s\n", agentNames[agent], dir)
				if err := export(out, b, dir, replace); err != nil {
					_, _ = fmt.Fprintf(out, "%v\n", err)
					failed = append(failed, err)
					continue
				}
			}
			if len(failed) > 0 {
				return errors.New("the skill was not installed everywhere")
			}
			_, err = fmt.Fprintln(out, "Start a new agent session to load the skill: /maintainer in Claude Code, $maintainer in Codex.")
			return err
		},
	}
	command.Flags().StringSliceVar(&agents, "agent", nil, "agents to install for: claude-code, codex")
	command.Flags().BoolVar(&project, "project", false, "install into the current directory instead of the user's skills")
	command.Flags().BoolVar(&replace, "replace", false, "replace a copy that maintainer installed earlier")
	_ = command.MarkFlagRequired("agent")
	return &command
}

func skillStatusCommand(load func() (bundle, error)) *cobra.Command {
	var (
		agents []string
		exact  bool
		asJSON bool
	)
	command := cobra.Command{
		Use:   "status",
		Short: "check the installed copies of the skill against this binary",
		Long: "Check the installed copies of the skill against this binary.\n\n" +
			"Looks in the user's and the project's skills of Claude Code and Codex, and\n" +
			"in their plugins. Fails when a copy is incompatible with this binary, was\n" +
			"changed after maintainer installed it, or no copy exists for a requested\n" +
			"--agent; with --exact, also when a copy differs from this release.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, agent := range agents {
				if _, ok := agentNames[agent]; !ok {
					return exit.WithUser(fmt.Errorf("unknown agent %q; use %s", agent, strings.Join(skill.Agents, " or ")))
				}
			}
			b, err := load()
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
			var copies []skill.Copy
			for _, loc := range skill.Locations(env, b.name) {
				if contains(agents, loc.Agent) {
					copies = append(copies, skill.Inspect(loc, b.digest, buildinfo.Version))
				}
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
				printStatus(out, b, env, copies)
			}
			return statusError(copies, agents, exact, cmd.Flags().Changed("agent"))
		},
	}
	command.Flags().StringSliceVar(&agents, "agent", nil, "agents to check: claude-code, codex (default all)")
	command.Flags().BoolVar(&exact, "exact", false, "also fail when a copy differs from this release")
	command.Flags().BoolVar(&asJSON, "json", false, "print one JSON record per copy")
	return &command
}

func printStatus(out io.Writer, b bundle, env skill.Env, copies []skill.Copy) {
	_, _ = fmt.Fprintf(out, "maintainer %s ships the %s skill %s, %s\n", buildinfo.Version, b.name, b.version, b.digest)
	if len(copies) == 0 {
		_, _ = fmt.Fprintln(out, "\nNo installed copies. Install one: maintainer skill install --agent claude-code,codex")
		return
	}
	perAgent := make(map[string]int)
	for _, c := range copies {
		perAgent[c.Agent]++
		where := strings.Replace(c.Path, env.Home, "~", 1)
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
	for _, agent := range skill.Agents {
		if perAgent[agent] > 1 {
			_, _ = fmt.Fprintf(out, "\n%s loads %d copies of the skill; keep one to avoid mixing versions.\n",
				agentNames[agent], perAgent[agent])
		}
	}
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

func statusError(copies []skill.Copy, agents []string, exact, requested bool) error {
	var problems []string
	found := make(map[string]bool)
	for _, c := range copies {
		found[c.Agent] = true
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
	if requested {
		for _, agent := range agents {
			if !found[agent] {
				problems = append(problems, "no copy for "+agentNames[agent])
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
