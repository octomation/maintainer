package command_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/command"
	"go.octolab.org/toolset/maintainer/internal/pkg/exit"
)

func TestEnforceUsage(t *testing.T) {
	// tree builds root → group → {leaf, runnable → nested}, and counts the
	// hooks and runs, which a usage error must not reach.
	tree := func(calls *int) *cobra.Command {
		count := func(*cobra.Command, []string) error { *calls++; return nil }
		root := &cobra.Command{Use: "root", Args: cobra.NoArgs, PersistentPreRunE: count, SilenceUsage: true}
		group := &cobra.Command{Use: "group", Args: cobra.NoArgs}
		leaf := &cobra.Command{Use: "leaf", Args: cobra.ExactArgs(1), RunE: count}
		leaf.Flags().Int("number", 0, "an integer")
		runnable := &cobra.Command{Use: "runnable", Args: cobra.NoArgs, RunE: count}
		nested := &cobra.Command{Use: "nested"}
		failing := &cobra.Command{Use: "failing", RunE: func(*cobra.Command, []string) error {
			return errors.New("runtime failure")
		}}
		runnable.AddCommand(nested)
		group.AddCommand(leaf, runnable, failing)
		root.AddCommand(group)
		return root
	}
	execute := func(args ...string) (string, string, int, error) {
		var calls int
		root := tree(&calls)
		var out, errs bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errs)
		EnforceUsage(root)
		root.SetArgs(args)
		err := root.Execute()
		return out.String(), errs.String(), calls, err
	}

	t.Run("usage errors", func(t *testing.T) {
		for _, test := range []struct {
			args    []string
			message string
		}{
			{[]string{"bogus"}, `unknown command "bogus" for "root"`},
			{[]string{"grop"}, "Did you mean this?\n\tgroup"},
			{[]string{"group", "laef"}, `unknown command "laef" for "root group"` + "\n\nDid you mean this?\n\tleaf"},
			{[]string{"group", "runnable", "bogus"}, `unknown command "bogus" for "root group runnable"`},
			{[]string{"group", "runnable", "nested", "bogus"}, `unknown command "bogus" for "root group runnable nested"`},
			{[]string{"group", "leaf"}, "accepts 1 arg(s), received 0"},
			{[]string{"group", "leaf", "x", "--number=abc"}, `invalid argument "abc" for "--number" flag`},
			{[]string{"group", "--bogus"}, "unknown flag: --bogus"},
			{[]string{"completion", "bogus"}, `unknown command "bogus" for "root completion"`},
			{[]string{"completion", "bash", "extra"}, `unknown command "extra" for "root completion bash"`},
		} {
			out, errs, calls, err := execute(test.args...)
			require.Errorf(t, err, "%v", test.args)
			var coded exit.Coder
			require.ErrorAsf(t, err, &coded, "%v", test.args)
			assert.Equalf(t, exit.User, coded.ExitCode(), "%v", test.args)
			assert.Containsf(t, err.Error(), test.message, "%v", test.args)
			assert.Containsf(t, errs, test.message, "%v reports to stderr", test.args)
			assert.Emptyf(t, out, "%v prints no help", test.args)
			assert.Zerof(t, calls, "%v runs no hook", test.args)
		}
	})

	t.Run("groups without arguments show their help", func(t *testing.T) {
		for _, args := range [][]string{{}, {"group"}, {"completion"}} {
			out, _, calls, err := execute(args...)
			require.NoErrorf(t, err, "%v", args)
			assert.Containsf(t, out, "Available Commands:", "%v", args)
			assert.Zerof(t, calls, "%v runs no hook", args)
		}
	})

	t.Run("valid runs and runtime errors keep their codes", func(t *testing.T) {
		_, _, calls, err := execute("group", "leaf", "x", "--number=1")
		require.NoError(t, err)
		assert.Equal(t, 2, calls, "the hook and the run")

		_, _, calls, err = execute("group", "runnable")
		require.NoError(t, err)
		assert.Equal(t, 2, calls, "the hook and the run")

		_, _, _, err = execute("group", "failing", "arbitrary")
		require.EqualError(t, err, "runtime failure")
		var coded exit.Coder
		assert.False(t, errors.As(err, &coded), "a runtime error is not a usage error")
	})
}
