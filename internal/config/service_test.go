package config_test

import (
	"os"
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
			"empty listen":     "listen = []\n",
			"named listen":     "listen = [\"localhost:17841\"]\n",
			"any listen":       "listen = [\"[::]:17841\"]\n",
			"listen port":      "listen = [\"127.0.0.1:65536\"]\n",
			"listen service":   "listen = [\"127.0.0.1:http\"]\n",
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

func TestLoadServiceListen(t *testing.T) {
	cnf, err := LoadService(afero.NewOsFs(), filepath.Join("testdata", "service", "valid", "listen.toml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"127.0.0.1:8080", "[::1]:8080"}, cnf.Listen)

	cnf, err = LoadService(afero.NewMemMapFs(), "")
	require.NoError(t, err)
	assert.Nil(t, cnf.Listen, "no listen keeps the defaults of the command")
}

func TestCheckLoopback(t *testing.T) {
	for _, address := range []string{"127.0.0.1:17841", "127.0.0.2:0", "[::1]:17841"} {
		assert.NoError(t, CheckLoopback(address), address)
	}
	for _, address := range []string{"0.0.0.0:17841", "localhost:17841", "192.0.2.1:17841", ":17841", "127.0.0.1", "[::]:17841",
		"127.0.0.1:", "127.0.0.1:65536", "127.0.0.1:-1", "127.0.0.1:http", "[::1]:+80"} {
		assert.Error(t, CheckLoopback(address), address)
	}
}

func TestServiceTemplate(t *testing.T) {
	// TestServiceExampleFixtures keeps the template and its uncommented form
	// as fixtures for make config-vet; here they load to the defaults.
	fs := afero.NewMemMapFs()
	live, err := os.ReadFile(filepath.Join("testdata", "service", "valid", "template-uncommented.toml"))
	require.NoError(t, err)
	require.NoError(t, afero.WriteFile(fs, "commented.toml", ServiceTemplate, 0o600))
	require.NoError(t, afero.WriteFile(fs, "live.toml", live, 0o600))
	defaults, err := LoadService(fs, "")
	require.NoError(t, err)

	commented, err := LoadService(fs, "commented.toml")
	require.NoError(t, err)
	assert.Equal(t, defaults.Weights(), commented.Weights())
	assert.Nil(t, commented.Listen)
	assert.Empty(t, commented.Repos)
	assert.Empty(t, commented.Fetch.Config)

	uncommented, err := LoadService(fs, "live.toml")
	require.NoError(t, err)
	assert.Equal(t, defaults.Weights(), uncommented.Weights(), "the commented values are the defaults")
	assert.Equal(t, []string{"127.0.0.1:17841", "[::1]:17841"}, uncommented.Listen)
	require.Len(t, uncommented.Repos, 1)
	assert.Equal(t, DefaultRepoWeight, uncommented.RepoWeight(0, "octomation", "maintainer"))
	for _, cnf := range []*Service{defaults, uncommented} {
		path, explicit := cnf.FetchConfigPath(afero.NewMemMapFs(), func(string) string { return "" }, "/home/me")
		assert.False(t, explicit, "the template keeps the usual lookup of the fetch config")
		assert.Empty(t, path)
	}
}

func TestEnsureServiceFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "maintainer", "service.toml")
	created, err := EnsureServiceFile(path)
	require.NoError(t, err)
	assert.True(t, created)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, ServiceTemplate, data)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	entries, err := os.ReadDir(filepath.Dir(path))
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temporary file is left behind")

	for _, content := range []string{"", "listen = [\"127.0.0.1:8080\"]\n"} {
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		created, err = EnsureServiceFile(path)
		require.NoError(t, err)
		assert.False(t, created)
		data, err = os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, content, string(data), "an existing file is never replaced")
	}
}

func TestEnsureServiceFile_Concurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.toml")
	results := make(chan bool, 8)
	for range cap(results) {
		go func() {
			created, err := EnsureServiceFile(path)
			assert.NoError(t, err)
			results <- created
		}()
	}
	won := 0
	for range cap(results) {
		if <-results {
			won++
		}
	}
	assert.Equal(t, 1, won, "exactly one process writes the file")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, ServiceTemplate, data)
}

func TestEnsureServiceFile_Unwritable(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	created, err := EnsureServiceFile(filepath.Join(dir, "maintainer", "service.toml"))
	assert.Error(t, err)
	assert.False(t, created)
}

func TestServiceConfigLookup(t *testing.T) {
	fs := afero.NewMemMapFs()
	getenv := func(string) string { return "" }
	path, required := ServiceConfigLookup{Getenv: getenv, Home: "/home/me"}.Resolve(fs)
	assert.Equal(t, "", path)
	assert.False(t, required)
	assert.Equal(t, "/home/me/.config/maintainer/service.toml", ServiceConfigLookup{Getenv: getenv, Home: "/home/me"}.Default(), "the default path, even without a file")
	assert.Empty(t, ServiceConfigLookup{Getenv: getenv}.Default())

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
