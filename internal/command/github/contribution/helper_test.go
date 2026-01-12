package contribution_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command/github/contribution"
	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/model/github/contribution"
	xtime "go.octolab.org/toolset/maintainer/internal/pkg/time"
)

func TestParseDate(t *testing.T) {
	start := xtime.UTC().Year(2021).Month(time.February).Day(8).Hour(9).Minute(16).Second(3)
	now := xtime.UTC().Year(2026).Month(time.September).Day(25).Hour(12).Minute(34).Second(56).Time()

	tests := []struct {
		name   string
		arg    string
		fDate  time.Time
		fWeeks int
		health func(require.TestingT, error, ...interface{})
		assert func(testing.TB, xtime.Range)
	}{
		{
			name:   "RFC3339 with default range",
			arg:    start.Format(time.RFC3339),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-27")
			},
		},
		{
			name:   "RFC3339 with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-20")
			},
		},
		{
			name:   "RFC3339 with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-13")
			},
		},
		{
			name:   "RFC3339 with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(time.RFC3339)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-02-07")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-03-06")
			},
		},
		{
			name:   "DateOnly with default range",
			arg:    start.Format(xtime.DateOnly),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-27")
			},
		},
		{
			name:   "DateOnly with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.DateOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-20")
			},
		},
		{
			name:   "DateOnly with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.DateOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-13")
			},
		},
		{
			name:   "DateOnly with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.DateOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-02-07")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-03-06")
			},
		},
		{
			name:   "YearAndMonth with default range",
			arg:    start.Format(xtime.YearAndMonth),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-17")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-20")
			},
		},
		{
			name:   "YearAndMonth with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.YearAndMonth)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-24")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-13")
			},
		},
		{
			name:   "YearAndMonth with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.YearAndMonth)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-10")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-06")
			},
		},
		{
			name:   "YearAndMonth with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.YearAndMonth)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2021-01-31")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-02-27")
			},
		},
		{
			name:   "YearOnly with default range",
			arg:    start.Format(xtime.YearOnly),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-13")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-16")
			},
		},
		{
			name:   "YearOnly with specified range",
			arg:    fmt.Sprintf("%s/3", start.Format(xtime.YearOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-20")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-09")
			},
		},
		{
			name:   "YearOnly with specified range behind",
			arg:    fmt.Sprintf("%s/-3", start.Format(xtime.YearOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-06")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-02")
			},
		},
		{
			name:   "YearOnly with specified range ahead",
			arg:    fmt.Sprintf("%s/+3", start.Format(xtime.YearOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, lr.From().Format(xtime.DateOnly), "2020-12-27")
				assert.Equal(t, lr.To().Format(xtime.DateOnly), "2021-01-23")
			},
		},
		{
			name:   "author date, strict ISO 8601 format",
			arg:    "2022-07-24T08:40:56+03:00/-3",
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, "2022-07-03", lr.From().Format(xtime.DateOnly))
				assert.Equal(t, "2022-07-30", lr.To().Format(xtime.DateOnly))
			},
		},
		{
			name:   "current time",
			arg:    "/-3",
			fDate:  now,
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {},
		},
		{
			name:   "issue#155: default range behind",
			arg:    start.Format(xtime.DateOnly),
			fDate:  time.Time{},
			fWeeks: -1,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, "2021-01-31", lr.From().Format(xtime.DateOnly))
				assert.Equal(t, "2021-02-13", lr.To().Format(xtime.DateOnly))
			},
		},
		{
			name:   "issue#155: default range behind of fallback",
			arg:    "",
			fDate:  start.Time(),
			fWeeks: -1,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, "2021-01-31", lr.From().Format(xtime.DateOnly))
				assert.Equal(t, "2021-02-13", lr.To().Format(xtime.DateOnly))
			},
		},
		{
			name:   "now keyword",
			arg:    "now/+0",
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, "2026-09-20", lr.From().Format(xtime.DateOnly))
				assert.Equal(t, "2026-09-26", lr.To().Format(xtime.DateOnly))
			},
		},
		{
			name:   "the longest range",
			arg:    fmt.Sprintf("%s/-53", start.Format(xtime.DateOnly)),
			fDate:  time.Time{},
			fWeeks: 5,
			health: require.NoError,
			assert: func(t testing.TB, lr xtime.Range) {
				assert.Equal(t, "2020-02-02", lr.From().Format(xtime.DateOnly))
				assert.Equal(t, "2021-02-13", lr.To().Format(xtime.DateOnly))
			},
		},
		{
			name:   "too long range",
			arg:    fmt.Sprintf("%s/54", start.Format(xtime.DateOnly)),
			fWeeks: 5,
			health: require.Error,
		},
		{
			name:   "too long range ahead",
			arg:    fmt.Sprintf("%s/+54", start.Format(xtime.DateOnly)),
			fWeeks: 5,
			health: require.Error,
		},
		{
			name:   "too long range behind",
			arg:    fmt.Sprintf("%s/-99999999", start.Format(xtime.DateOnly)),
			fWeeks: 5,
			health: require.Error,
		},
		{
			name:   "too many parts",
			arg:    fmt.Sprintf("%s/3/3", start.Format(xtime.DateOnly)),
			fWeeks: 5,
			health: require.Error,
		},
		{
			name:   "the zero time instant",
			arg:    "0001/+1",
			fWeeks: 5,
			health: require.Error,
		},
		{
			name:   "fallback before 1970",
			arg:    "git",
			fDate:  xtime.UTC().Year(1969).Month(time.December).Day(31).Time(),
			fWeeks: 5,
			health: require.Error,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts, err := ParseDate([]string{test.arg}, test.fDate, test.fWeeks, now)
			test.health(t, err)
			if err == nil {
				test.assert(t, contribution.LookupRange(opts))
			}
		})
	}
}

func TestFallbackDate(t *testing.T) {
	now := xtime.UTC().Year(2026).Month(time.September).Day(25).Hour(12).Minute(34).Second(56).Time()
	head := time.Date(2026, time.September, 2, 13, 14, 15, 0, time.FixedZone("", 3*60*60))

	t.Run("explicit date", func(t *testing.T) {
		t.Setenv("GIT_DIR", filepath.Join(t.TempDir(), "missing")) // not even opened
		t.Chdir(repository(t, head))

		for _, arg := range []string{"now", "2022-01-02", "2022-01-02/-3", "garbage/x"} {
			anchor, err := FallbackDate([]string{arg}, now)
			require.NoError(t, err, arg)
			assert.Equal(t, now, anchor, arg)
		}
	})

	t.Run("working directory inside a repository", func(t *testing.T) {
		t.Setenv("GIT_DIR", "")
		dir := filepath.Join(repository(t, head), "nested", "dir")
		require.NoError(t, os.MkdirAll(dir, 0o755))
		t.Chdir(dir)

		for _, args := range [][]string{nil, {""}, {"git"}, {"/-3"}, {"git/3"}} {
			anchor, err := FallbackDate(args, now)
			require.NoError(t, err, args)
			assert.True(t, head.Equal(anchor), "%v: %s", args, anchor)
		}
	})

	t.Run("issue#189: GIT_DIR from another directory", func(t *testing.T) {
		repo := repository(t, head)
		t.Chdir(t.TempDir())
		t.Setenv("GIT_DIR", filepath.Join(repo, ".git"))

		anchor, err := FallbackDate([]string{"git/1"}, now)
		require.NoError(t, err)
		assert.True(t, head.Equal(anchor), anchor)
	})

	t.Run("GIT_DIR wins over the working directory", func(t *testing.T) {
		other := head.AddDate(-1, 0, 0)
		t.Chdir(repository(t, other))
		t.Setenv("GIT_DIR", filepath.Join(repository(t, head), ".git"))

		anchor, err := FallbackDate(nil, now)
		require.NoError(t, err)
		assert.True(t, head.Equal(anchor), anchor)
	})

	t.Run("GIT_DIR is not a repository", func(t *testing.T) {
		t.Chdir(repository(t, head))
		for _, dir := range []string{t.TempDir(), filepath.Join(t.TempDir(), "missing")} {
			t.Setenv("GIT_DIR", dir)

			_, err := FallbackDate(nil, now)
			require.Error(t, err, dir)
			assert.Contains(t, err.Error(), "GIT_DIR")
			assert.Contains(t, err.Error(), dir)
		}
	})

	t.Run("repository without commits", func(t *testing.T) {
		repo := repository(t, time.Time{})

		t.Setenv("GIT_DIR", "")
		t.Chdir(repo)
		anchor, err := FallbackDate(nil, now)
		require.NoError(t, err)
		assert.Equal(t, now, anchor)

		t.Chdir(t.TempDir())
		t.Setenv("GIT_DIR", filepath.Join(repo, ".git"))
		anchor, err = FallbackDate(nil, now)
		require.NoError(t, err)
		assert.Equal(t, now, anchor)
	})

	t.Run("outside a repository", func(t *testing.T) {
		t.Setenv("GIT_DIR", "")
		t.Chdir(t.TempDir())

		anchor, err := FallbackDate([]string{"git/3"}, now)
		require.NoError(t, err)
		assert.Equal(t, now, anchor)
	})

	t.Run("linked worktree", func(t *testing.T) {
		repo := repository(t, head)
		r, err := git.PlainOpen(repo)
		require.NoError(t, err)
		ref, err := r.Head()
		require.NoError(t, err)

		// the layout of git worktree add: the branch lives in the main repository
		linked, admin := t.TempDir(), filepath.Join(repo, ".git", "worktrees", "linked")
		require.NoError(t, os.MkdirAll(admin, 0o755))
		for name, content := range map[string]string{
			filepath.Join(admin, "HEAD"):      "ref: " + ref.Name().String(),
			filepath.Join(admin, "commondir"): "../..",
			filepath.Join(admin, "gitdir"):    filepath.Join(linked, ".git"),
			filepath.Join(linked, ".git"):     "gitdir: " + admin,
		} {
			require.NoError(t, os.WriteFile(name, []byte(content+"\n"), 0o644))
		}
		t.Setenv("GIT_DIR", "")
		t.Chdir(linked)

		anchor, err := FallbackDate(nil, now)
		require.NoError(t, err)
		assert.True(t, head.Equal(anchor), anchor)
	})
}

// TestCommands_Decline runs the commands with the input they must decline
// before any request: an ordinary error, nothing in stdout, and no panic.
// The future is the year 9999 here, because the commands take the clock.
func TestCommands_Decline(t *testing.T) {
	type command = func(*cobra.Command, *config.Tool) *cobra.Command

	tests := map[string]struct {
		command command
		args    []string
		err     string
	}{
		"lookup in the future":      {Lookup, []string{"9999-12-01/-1"}, xtime.ErrFuturePeriod.Error()},
		"lookup too long":           {Lookup, []string{"2022/-99999999"}, "weeks"},
		"lookup at the zero time":   {Lookup, []string{"0001/+1"}, "1970"},
		"suggest with zero target":  {Suggest, []string{"--target", "0", "2022-01-05/1"}, "target"},
		"suggest with zero only":    {Suggest, []string{"--target=0"}, "target"},
		"suggest in the future":     {Suggest, []string{"9999-01-01"}, "9999-01-01T00:00:00Z"},
		"suggest behind the future": {Suggest, []string{"9999-01-01/-3"}, "9999-01-01T00:00:00Z"},
		"suggest at the zero time":  {Suggest, []string{"0001/1"}, "1970"},
		"histogram in the future":   {Histogram, []string{"9999"}, xtime.ErrFuturePeriod.Error()},
		"histogram of a malformed":  {Histogram, []string{"2021-02-30"}, "2021-02-30"},
		"snapshot in the future":    {Snapshot, []string{"9999"}, xtime.ErrFuturePeriod.Error()},
		"snapshot of the zero time": {Snapshot, []string{"0001"}, "1970"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := test.command(&cobra.Command{Use: "test", SilenceUsage: true}, new(config.Tool))
			cmd.SetArgs(test.args)
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)

			err := cmd.ExecuteContext(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.err)
			assert.Empty(t, stdout.String())
			assert.Contains(t, stderr.String(), err.Error())
		})
	}
}

// repository creates a repository with a commit authored at the time,
// or without commits if the time is zero, and returns its work tree.
func repository(t *testing.T, when time.Time) string {
	t.Helper()

	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	require.NoError(t, err)
	if when.IsZero() {
		return dir
	}

	tree, err := repo.Worktree()
	require.NoError(t, err)
	author := &object.Signature{Name: "Maintainer", Email: "maintainer@octolab.org", When: when}
	_, err = tree.Commit("init", &git.CommitOptions{AllowEmptyCommits: true, Author: author})
	require.NoError(t, err)
	return dir
}

func TestTableView(t *testing.T) {
	sunday := xtime.UTC().Year(2021).Month(time.January).Day(3)

	heats := make(contribution.HeatMap)
	heats.SetCount(sunday.Time(), 5)
	heats.SetCount(sunday.Day(4).Time(), 5)
	heats.SetCount(sunday.Day(5).Time(), 10)
	heats.SetCount(sunday.Day(14).Time(), 3) // after the scope, shown as "?"

	var buf bytes.Buffer
	cmd := new(cobra.Command)
	cmd.SetErr(&buf)
	TableView(cmd, heats, xtime.NewRange(sunday.Time(), sunday.Day(12).Hour(23).Time()))

	// ten days up to the end of the scope: three with contributions, seven without
	assert.Contains(t, buf.String(), "distribution{0: 7, 5: 2, 10: 1}")
}

func TestTableView_Weeks(t *testing.T) {
	tests := map[string]struct {
		from     time.Time
		expected []string
	}{
		"after a long ISO year": {xtime.UTC().Year(2021).Month(time.January).Day(3).Time(), []string{"#02", "#03"}},
		"across the new year":   {xtime.UTC().Year(2022).Month(time.December).Day(25).Time(), []string{"#53", "#01"}},
		"the new year week":     {xtime.UTC().Year(2025).Month(time.December).Day(28).Time(), []string{"#01", "#02"}},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := new(cobra.Command)
			cmd.SetErr(&buf)
			TableView(cmd, make(contribution.HeatMap), xtime.NewRange(test.from, test.from.Add(2*xtime.Week-time.Nanosecond)))

			header := strings.Fields(strings.SplitN(strings.TrimSpace(buf.String()), "\n", 2)[0])
			assert.Equal(t, append([]string{"Day", "/", "Week"}, append(test.expected, "Date")...), header)
		})
	}
}
