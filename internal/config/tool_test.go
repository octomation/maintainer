package config_test

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.octolab.org/env"
	"go.octolab.org/toolkit/config"
	"go.octolab.org/unsafe"

	. "go.octolab.org/toolset/maintainer/internal/config"
)

// TODO:replace https://github.com/octolab/pkg/issues/65
func Get(src env.Environment, key string) string {
	for _, v := range src {
		if v.Name() == key {
			return v.Value()
		}
	}
	return ""
}

func TestTool_Load(t *testing.T) {
	const (
		envGitRemote   = "GIT_REMOTE"
		envGithubToken = "GITHUB_TOKEN"
	)

	vars := env.Environment{
		env.Must(envGitRemote, "git@github.com:octomation/maintainer.git"),
		env.Must(envGithubToken, "secret"),
	}

	t.Run("load from file", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		f, err := fs.Create(".env")
		require.NoError(t, err)
		for _, v := range vars {
			unsafe.DoSilent(f.WriteString(v.String() + "\n"))
		}
		require.NoError(t, f.Close())

		var cnf Tool
		require.NoError(t, cnf.Load(fs))
		assert.Equal(t, Get(vars, envGitRemote), cnf.Git.Remote)
		assert.Equal(t, config.Secret(Get(vars, envGithubToken)), cnf.GitHub.Token)
	})

	t.Run("load from env", func(t *testing.T) {
		bindings := make([]func(provider *viper.Viper) error, 0, len(vars))
		for _, v := range vars {
			t.Setenv(v.Name(), v.Value())

			name := v.Name()
			bindings = append(bindings, func(v *viper.Viper) error { return v.BindEnv(name) })
		}

		var cnf Tool
		require.NoError(t, cnf.Load(afero.NewMemMapFs(), bindings...))
		assert.Equal(t, Get(vars, envGitRemote), cnf.Git.Remote)
		assert.Equal(t, config.Secret(Get(vars, envGithubToken)), cnf.GitHub.Token)
	})

	t.Run("load from flags", func(t *testing.T) {
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flags.String("remote", "", "git remote url")
		flags.String("token", "", "github access token")
		bindings := []func(provider *viper.Viper) error{
			func(v *viper.Viper) error { return v.BindPFlag(envGitRemote, flags.Lookup("remote")) },
			func(v *viper.Viper) error { return v.BindPFlag(envGithubToken, flags.Lookup("token")) },
		}

		var cnf Tool
		require.NoError(t, flags.Parse([]string{
			fmt.Sprintf("--remote=%s", Get(vars, envGitRemote)),
			fmt.Sprintf("--token=%s", Get(vars, envGithubToken)),
		}))
		require.NoError(t, cnf.Load(afero.NewMemMapFs(), bindings...))
		assert.Equal(t, Get(vars, envGitRemote), cnf.Git.Remote)
		assert.Equal(t, config.Secret(Get(vars, envGithubToken)), cnf.GitHub.Token)
	})
}
