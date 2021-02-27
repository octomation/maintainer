package config_test

import (
	"testing"

	"github.com/mitchellh/mapstructure"
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

func TestTool_Load(t *testing.T) {
	vars := env.Environment{
		env.Must("GIT_REMOTE", "git@github.com:octomation/maintainer.git"),
		env.Must("GITHUB_TOKEN", "secret"),
	}

	t.Run("load from file", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		f, err := fs.Create(".env")
		require.NoError(t, err)

		for _, v := range vars {
			unsafe.DoSilent(f.WriteString(v.String() + "\n"))
		}
		unsafe.Ignore(f.Close())

		var cnf Tool
		require.NoError(t, cnf.Load(fs))
		assert.Equal(t, "git@github.com:octomation/maintainer.git", cnf.Git.Remote)
		assert.Equal(t, config.Secret("secret"), cnf.GitHub.Token)
	})

	t.Run("load from env", func(t *testing.T) {
		// TODO:debt improve Viper <> Config <> Env experience
		bindings := make([]func(cnf *Tool, provider *viper.Viper) error, 0, len(vars))
		for _, v := range vars {
			name := v.Name()
			t.Setenv(v.Name(), v.Value())
			bindings = append(bindings, func(_ *Tool, provider *viper.Viper) error {
				return provider.BindEnv(name)
			})
		}

		var cnf Tool
		require.NoError(t, cnf.Load(afero.NewMemMapFs(), bindings...))
		assert.Equal(t, "git@github.com:octomation/maintainer.git", cnf.Git.Remote)
		assert.Equal(t, config.Secret("secret"), cnf.GitHub.Token)
	})

	t.Run("load from flags", func(t *testing.T) {
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flags.String("remote", "", "git remote url")
		flags.String("token", "", "github access token")

		bindings := []func(cnf *Tool, provider *viper.Viper) error{
			func(cnf *Tool, provider *viper.Viper) error {
				return provider.BindPFlag("git_remote", flags.Lookup("remote"))
			},
			func(cnf *Tool, provider *viper.Viper) error {
				return provider.BindPFlag("github_token", flags.Lookup("token"))
			},
		}

		var cnf Tool
		require.NoError(t, flags.Parse([]string{
			"--remote=git@github.com:octomation/maintainer.git",
			"--token=secret",
		}))
		require.NoError(t, cnf.Load(afero.NewMemMapFs(), bindings...))
		assert.Equal(t, "git@github.com:octomation/maintainer.git", cnf.Git.Remote)
		assert.Equal(t, config.Secret("secret"), cnf.GitHub.Token)
	})

	t.Run("squash with prefix", func(t *testing.T) {
		t.SkipNow()

		type Git struct {
			Remote string `mapstructure:"remote"`
		}

		type GitHub struct {
			Token config.Secret `mapstructure:"token"`
		}

		type Tool struct {
			Git    `mapstructure:"git,squash"`
			GitHub `mapstructure:"github,squash"`
		}

		var cnf Tool
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			DecodeHook:       nil,
			ErrorUnused:      false,
			ZeroFields:       false,
			WeaklyTypedInput: false,
			Squash:           false,
			Metadata:         nil,
			Result:           &cnf,
			TagName:          "",
			MatchName:        nil,
		})
		require.NoError(t, err)

		input := map[string]interface{}{
			"GIT_REMOTE":   "git@github.com:octomation/maintainer.git",
			"GITHUB_TOKEN": "secret",
		}

		require.NoError(t, decoder.Decode(input))
		assert.Equal(t, "git@github.com:octomation/maintainer.git", cnf.Git.Remote)
		assert.Equal(t, config.Secret("secret"), cnf.GitHub.Token)
	})
}
