package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.octolab.org/safe"
)

func TestExecution(t *testing.T) {
	args, code, errs, outs := os.Args, exit, stderr, stdout
	t.Cleanup(func() { os.Args, exit, stderr, stdout = args, code, errs, outs })

	// fetch must not find a real config, state or token if it gets that far,
	// and the root command must not read a .env of the working directory.
	t.Chdir(t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("MAINTAINER_FETCH_CONFIG", "")
	t.Setenv("GITHUB_TOKEN", "")

	t.Run("success", func(t *testing.T) {
		for _, args := range [][]string{{"version"}, {}, {"github"}, {"fetch", "state"}, {"skill", "info"}} {
			code, out, _ := execute(args...)
			assert.Equalf(t, -1, code, "%v must not call exit", args)
			assert.NotEmptyf(t, out, "%v prints its output", args)
		}
	})

	t.Run("usage errors", func(t *testing.T) {
		for _, test := range []struct {
			args    []string
			message string
		}{
			{[]string{"unknown"}, `unknown command "unknown" for "maintainer"`},
			{[]string{"fecth"}, "Did you mean this?\n\tfetch"},
			{[]string{"github", "contribution", "bogus"}, `unknown command "bogus" for "maintainer github contribution"`},
			{[]string{"fetch", "state", "bogus"}, `unknown command "bogus" for "maintainer fetch state"`},
			{[]string{"fetch", "bogus"}, `unknown command "bogus" for "maintainer fetch"`},
			{[]string{"completion", "bogus"}, `unknown command "bogus" for "maintainer completion"`},
			{[]string{"version", "extra"}, `unknown command "extra" for "maintainer version"`},
			{[]string{"github", "contribution", "diff", "a"}, "accepts 2 arg(s), received 1"},
			{[]string{"github", "contribution", "lookup", "bogus"}, `parse date "bogus"`},
			{[]string{"github", "contribution", "snapshot", "bogus"}, `invalid argument "bogus"`},
			{[]string{"github", "contribution", "histogram", "2021"}, "a GitHub token is required"},
			{[]string{"--bogus"}, "unknown flag: --bogus"},
			{[]string{"status", "--concurrency=abc"}, `invalid argument "abc" for "--concurrency" flag`},
			{[]string{"fetch", "--timeout=soon"}, `invalid argument "soon" for "--timeout" flag`},
			{[]string{"fetch", "--config=", "--concurrency=-1"}, "concurrency and timeout must not be negative"},
			{[]string{"fetch", "--config=", "--format=jsno"}, `unknown format "jsno"`},
			{[]string{"skill", "install"}, `required flag(s) "agent" not set`},
			{[]string{"skill", "status", "--agent=cursor"}, `unknown agent "cursor"`},
		} {
			code, out, errs := execute(test.args...)
			assert.Equalf(t, 2, code, "%v exits with a usage error", test.args)
			assert.Containsf(t, errs, test.message, "%v names the cause in stderr", test.args)
			assert.Emptyf(t, out, "%v prints no help to stdout", test.args)
		}
	})

	t.Run("shutdown with panic", func(t *testing.T) {
		code := -1
		exit = func(c int) { code = c }
		stderr, stdout = io.Discard, io.Discard
		safe.Do(func() error { panic("test") }, shutdown)
		assert.Equal(t, 1, code)
	})
}

// execute runs main with the arguments and returns the code it passed to
// exit, or -1 if it did not call exit, as on success, with stdout and stderr.
func execute(args ...string) (int, string, string) {
	code := -1
	var out, errs bytes.Buffer
	exit = func(c int) { code = c }
	stderr, stdout = &errs, &out
	os.Args = append([]string{"root"}, args...)
	main()
	return code, out.String(), errs.String()
}
