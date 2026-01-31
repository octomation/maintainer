package skills_test

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/command"
	"go.octolab.org/toolset/maintainer/internal/skill"
	"go.octolab.org/toolset/maintainer/skills"
)

const reference = "references/cli.md"

var maintainer = os.DirFS("maintainer")

func TestSkill(t *testing.T) {
	for _, err := range skill.Check("maintainer", maintainer) {
		t.Error(err)
	}
}

// TestEmbedded makes sure the binary carries every file of the skill.
func TestEmbedded(t *testing.T) {
	want, err := skill.Digest(maintainer)
	require.NoError(t, err)
	got, err := skill.Digest(skills.Maintainer())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestReference(t *testing.T) {
	committed, err := fs.ReadFile(maintainer, reference)
	require.NoError(t, err)
	if string(committed) != skill.Reference(command.New()) {
		t.Errorf("%s is stale; run go generate ./skills", reference)
	}
}

// TestExamples parses every maintainer command line in the skill's code blocks
// against the command tree without running anything.
func TestExamples(t *testing.T) {
	for _, ex := range examples(t) {
		root := command.New()
		// Unknown subcommands of a group are errors, as in the binary.
		command.EnforceUsage(root)
		root.SetArgs(ex.args)
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		disarm(root)
		_, err := root.ExecuteC()
		assert.NoError(t, err, "%s: %s", ex.at, ex.line)
	}
}

// TestCoverage requires the skill to mention every command that does work.
func TestCoverage(t *testing.T) {
	var text strings.Builder
	for _, name := range []string{"SKILL.md", "references/workflows.md"} {
		data, err := fs.ReadFile(maintainer, name)
		require.NoError(t, err)
		text.Write(data)
	}
	for _, cmd := range skill.Leaves(command.New()) {
		// maintainer fetch config validate is mentioned as config validate,
		// maintainer skill status as skill status, and maintainer version in
		// full.
		name := strings.TrimPrefix(cmd.CommandPath(), "maintainer ")
		if path := strings.Fields(name); len(path) > 2 {
			name = strings.Join(path[len(path)-2:], " ")
		}
		mentioned := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
		if !mentioned.MatchString(text.String()) {
			t.Errorf("the skill never mentions %s", cmd.CommandPath())
		}
	}
}

// TestRanges keeps every compatibility range in the text equal to
// metadata.tool-version-range.
func TestRanges(t *testing.T) {
	text, err := fs.ReadFile(maintainer, "SKILL.md")
	require.NoError(t, err)
	fm, _, err := skill.ParseFrontmatter(string(text))
	require.NoError(t, err)
	meta, _ := fm.Map("metadata")
	declared := meta["tool-version-range"]
	compatibility, _ := fm.String("compatibility")
	assert.Contains(t, compatibility, declared)

	const version = `\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?`
	ranges := regexp.MustCompile(`>=` + version + ` <` + version)
	files, err := skill.Files(maintainer)
	require.NoError(t, err)
	for _, name := range files {
		data, err := fs.ReadFile(maintainer, name)
		require.NoError(t, err)
		for _, found := range ranges.FindAllString(string(data), -1) {
			assert.Equal(t, declared, found, "%s mentions another range", name)
		}
	}
}

type example struct {
	at   string
	line string
	args []string
}

// examples collects the command lines that start with maintainer from fenced
// code blocks, joining continuation lines and stopping at pipes and
// redirections. The generated reference repeats the command tree's own
// examples, so the tree's Example fields are checked instead.
func examples(t *testing.T) []example {
	t.Helper()
	var list []example
	files, err := skill.Files(maintainer)
	require.NoError(t, err)
	for _, name := range files {
		if !strings.HasSuffix(name, ".md") || name == reference {
			continue
		}
		data, err := fs.ReadFile(maintainer, name)
		require.NoError(t, err)
		var (
			fence   bool
			pending string
			start   int
		)
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for n := 1; scanner.Scan(); n++ {
			line := scanner.Text()
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				fence = !fence
				continue
			}
			if !fence {
				continue
			}
			if pending == "" {
				start = n
			}
			pending += strings.TrimSpace(line)
			if strings.HasSuffix(pending, `\`) {
				pending = strings.TrimSpace(strings.TrimSuffix(pending, `\`)) + " "
				continue
			}
			list = append(list, parse(t, name, start, pending)...)
			pending = ""
		}
	}
	for _, cmd := range commands(command.New()) {
		for _, line := range strings.Split(cmd.Example, "\n") {
			list = append(list, parse(t, cmd.CommandPath()+" example", 0, strings.TrimSpace(line))...)
		}
	}
	require.NotEmpty(t, list)
	return list
}

func parse(t *testing.T, file string, line int, text string) []example {
	t.Helper()
	words, err := shellWords(text)
	require.NoError(t, err, "%s:%d: %s", file, line, text)
	if len(words) == 0 || words[0] != "maintainer" {
		return nil
	}
	at := file
	if line > 0 {
		at += ":" + strconv.Itoa(line)
	}
	return []example{{at: at, line: text, args: words[1:]}}
}

// shellWords splits a command line like a POSIX shell, up to the first
// unquoted pipe, redirection or command separator.
func shellWords(s string) ([]string, error) {
	var (
		words []string
		word  strings.Builder
		quote rune
		have  bool
	)
	for i := 0; i < len(s); i++ {
		c := rune(s[i])
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
				continue
			}
			word.WriteRune(c)
		case c == '\'' || c == '"':
			quote, have = c, true
		case c == '\\' && i+1 < len(s):
			i++
			word.WriteByte(s[i])
			have = true
		case c == ' ' || c == '\t':
			if have {
				words = append(words, word.String())
				word.Reset()
				have = false
			}
		case strings.ContainsRune("|&;<>", c):
			i = len(s)
		default:
			word.WriteRune(c)
			have = true
		}
	}
	if quote != 0 {
		return nil, io.ErrUnexpectedEOF
	}
	if have {
		words = append(words, word.String())
	}
	return words, nil
}

// disarm leaves only argument and flag parsing: no hook reads .env or a
// config, and no command does its work.
func disarm(cmd *cobra.Command) {
	cmd.PersistentPreRun, cmd.PersistentPreRunE = nil, nil
	cmd.PreRun, cmd.PreRunE = nil, nil
	cmd.PostRun, cmd.PostRunE = nil, nil
	cmd.PersistentPostRun, cmd.PersistentPostRunE = nil, nil
	if cmd.Runnable() {
		cmd.Run, cmd.RunE = nil, func(*cobra.Command, []string) error { return nil }
	}
	for _, sub := range cmd.Commands() {
		disarm(sub)
	}
}

func commands(cmd *cobra.Command) []*cobra.Command {
	list := []*cobra.Command{cmd}
	for _, sub := range cmd.Commands() {
		list = append(list, commands(sub)...)
	}
	return list
}
