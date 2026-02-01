package config_test

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/config"
)

// The fixtures are the ones `make config-vet` checks against service.cue, so
// the loader and the schema accept and reject the same files.
func TestServiceSchemaFixtures(t *testing.T) {
	for dir, valid := range map[string]bool{"valid": true, "invalid": false} {
		paths, err := filepath.Glob(filepath.Join("testdata", "service", dir, "*.toml"))
		require.NoError(t, err)
		require.NotEmpty(t, paths)
		for _, path := range paths {
			_, err := LoadService(afero.NewOsFs(), path)
			assert.Equal(t, valid, err == nil, "%s: %v", path, err)
		}
	}
}

func TestLoadService(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cnf, err := LoadService(afero.NewMemMapFs(), "")
		require.NoError(t, err)
		assert.Equal(t, Weights{Impact: 1, Effort: 1, Type: 1, Age: 1, Due: 1, AgeHorizonDays: 365, DueHorizonDays: 30}, cnf.Weights())
		assert.Equal(t, 1.0, cnf.RepoWeight(1, "a", "b"))
	})

	t.Run("full", func(t *testing.T) {
		cnf, err := LoadService(afero.NewOsFs(), filepath.Join("testdata", "service", "valid", "full.toml"))
		require.NoError(t, err)
		assert.Equal(t, Weights{Impact: 2, Effort: 1.5, Type: 0, Age: 0.5, Due: 3, AgeHorizonDays: 180, DueHorizonDays: 14}, cnf.Weights())
		assert.Equal(t, 2.0, cnf.RepoWeight(42, "other", "name"), "the id wins")
		assert.Equal(t, 1.5, cnf.RepoWeight(0, "OctoMation", "Maintainer"), "owner and name match case-insensitively")
		assert.Equal(t, 0.0, cnf.RepoWeight(7, "octopot", "idle"), "an explicit zero stays zero")
		assert.Equal(t, 1.0, cnf.RepoWeight(0, "octopot", "other"))
		assert.Equal(t, ".config/maintainer/fetch.toml", cnf.Fetch.Config)
		path, explicit := cnf.FetchConfigPath(afero.NewMemMapFs(), func(string) string { return "" }, "/home/me")
		assert.True(t, explicit)
		assert.Equal(t, "/home/me/.config/maintainer/fetch.toml", path)
	})

	t.Run("missing explicit file", func(t *testing.T) {
		_, err := LoadService(afero.NewMemMapFs(), "/nowhere/service.toml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read service config")
	})

	t.Run("rejections", func(t *testing.T) {
		for name, body := range map[string]string{
			"unknown key":      "[score]\nimpakt = 1\n",
			"negative weight":  "[[repos]]\nmatch = { id = 1 }\nweight = -1\n",
			"zero horizon":     "[score]\ndue_horizon_days = 0\n",
			"owner-only match": "[[repos]]\nmatch = { owner = \"acme\" }\n",
			"infinite weight":  "[score]\nage = inf\n",
			"nan weight":       "[[repos]]\nmatch = { id = 1 }\nweight = nan\n",
			"yaml extension":   "",
		} {
			fs := afero.NewMemMapFs()
			path := "service.toml"
			if name == "yaml extension" {
				path = "service.yaml"
			}
			require.NoError(t, afero.WriteFile(fs, path, []byte(body), 0o600))
			_, err := LoadService(fs, path)
			assert.Error(t, err, name)
		}
	})

	t.Run("unknown key suggestion", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "service.toml", []byte("[score]\nimpakt = 1\n"), 0o600))
		_, err := LoadService(fs, "service.toml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown key impakt in [score]; did you mean impact?`)
	})
}

func TestServiceConfigLookup(t *testing.T) {
	fs := afero.NewMemMapFs()
	getenv := func(string) string { return "" }
	path, required := ServiceConfigLookup{Getenv: getenv, Home: "/home/me"}.Resolve(fs)
	assert.Equal(t, "", path)
	assert.False(t, required)

	require.NoError(t, afero.WriteFile(fs, "/home/me/.config/maintainer/service.toml", []byte(""), 0o600))
	path, required = ServiceConfigLookup{Getenv: getenv, Home: "/home/me"}.Resolve(fs)
	assert.Equal(t, "/home/me/.config/maintainer/service.toml", path)
	assert.False(t, required)

	require.NoError(t, afero.WriteFile(fs, "/xdg/maintainer/service.toml", []byte(""), 0o600))
	path, _ = ServiceConfigLookup{Getenv: func(key string) string {
		if key == "XDG_CONFIG_HOME" {
			return "/xdg"
		}
		return ""
	}, Home: "/home/me"}.Resolve(fs)
	assert.Equal(t, "/xdg/maintainer/service.toml", path)

	path, required = ServiceConfigLookup{Explicit: "/elsewhere/service.toml", ExplicitSet: true, Home: "/home/me"}.Resolve(fs)
	assert.Equal(t, "/elsewhere/service.toml", path)
	assert.True(t, required)
}

func TestResolveAgainst(t *testing.T) {
	assert.Equal(t, "/home/me/x/y", ResolveAgainst("/home/me", "x/y"))
	assert.Equal(t, "/home/me/x", ResolveAgainst("/home/me", "~/x"))
	assert.Equal(t, "/home/me", ResolveAgainst("/home/me", "~"))
	assert.Equal(t, "/abs/x", ResolveAgainst("/home/me", "/abs/x"))
}
