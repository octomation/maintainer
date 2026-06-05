package config_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/config"
)

const sampleTOML = `
[defaults]
root = "/srv/code"
clone_url = "ssh"
concurrency = 8

[filters]
exclude_archived = true

[profiles.primary]
token_env = "GITHUB_TOKEN"
include_owners = ["acme", "acme-user"]

[profiles.bot]
token_env = "BOT_TOKEN"
include_owners = ["acme-bot"]
clone_url = "https"

[[owners]]
name = "acme"
path = "mirror/{{.Owner}}/{{.Repo}}"

[[repos]]
match = { id = 42 }
path = "~/Code/special"
clone_url = "https"

[[repos]]
match = { owner = "acme-user", name = "dotfiles" }
path = "~/.dotfiles"
`

const sampleYAML = `
defaults:
  root: /srv/code
  clone_url: ssh
  concurrency: 8
filters:
  exclude_archived: true
profiles:
  primary:
    token_env: GITHUB_TOKEN
    include_owners: [acme, acme-user]
  bot:
    token_env: BOT_TOKEN
    include_owners: [acme-bot]
    clone_url: https
owners:
  - name: acme
    path: mirror/{{.Owner}}/{{.Repo}}
repos:
  - match: { id: 42 }
    path: ~/Code/special
    clone_url: https
  - match: { owner: acme-user, name: dotfiles }
    path: ~/.dotfiles
`

func TestLoadFetch_TOMLandYAML(t *testing.T) {
	for _, tc := range []struct{ name, file, body string }{
		{"toml", "fetch.toml", sampleTOML},
		{"yaml", "fetch.yaml", sampleYAML},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			require.NoError(t, afero.WriteFile(fs, tc.file, []byte(tc.body), 0o600))

			cnf, err := LoadFetch(fs, tc.file)
			require.NoError(t, err)
			require.NoError(t, cnf.Validate())

			assert.Equal(t, "/srv/code", cnf.Defaults.Root)
			assert.Equal(t, 8, cnf.Defaults.Concurrency)
			assert.True(t, cnf.Filters.ExcludeArchived)
			assert.Len(t, cnf.Profiles, 2)
			assert.Equal(t, "https", cnf.Profiles["bot"].CloneURL)
			require.Len(t, cnf.Owners, 1)
			assert.Equal(t, "acme", cnf.Owners[0].Name)
			require.Len(t, cnf.Repos, 2)
			assert.Equal(t, int64(42), cnf.Repos[0].Match.ID)
		})
	}
}

func TestLoadFetch_DefaultsApplied(t *testing.T) {
	fs := afero.NewMemMapFs()
	cnf, err := LoadFetch(fs, "") // empty path → single-run defaults
	require.NoError(t, err)
	assert.Equal(t, DefaultRoot, cnf.Defaults.Root)
	assert.Equal(t, DefaultPath, cnf.Defaults.Path)
	assert.Equal(t, DefaultCloneURL, cnf.Defaults.CloneURL)
	assert.Equal(t, DefaultConcurrency, cnf.Defaults.Concurrency)
}

func TestValidate_Rejections(t *testing.T) {
	tests := map[string]func(*Fetch){
		"escape via parent":  func(f *Fetch) { f.Defaults.Path = "../{{.Repo}}" },
		"absolute defaults":  func(f *Fetch) { f.Defaults.Path = "/etc/{{.Repo}}" },
		"bad transport":      func(f *Fetch) { f.Defaults.CloneURL = "ftp" },
		"zero concurrency":   func(f *Fetch) { f.Defaults.Concurrency = 0 },
		"owner path escapes": func(f *Fetch) { f.Owners = []Owner{{Name: "a", Path: "../x"}} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			f := &Fetch{Defaults: Defaults{Root: ".", Path: DefaultPath, CloneURL: "ssh", Concurrency: 4}}
			mutate(f)
			assert.Error(t, f.Validate())
		})
	}
}

func TestCloneURLFor_Precedence(t *testing.T) {
	f := &Fetch{
		Defaults: Defaults{Root: ".", Path: DefaultPath, CloneURL: "ssh", Concurrency: 1},
		Profiles: map[string]Profile{"bot": {CloneURL: "https"}},
		Owners:   []Owner{{Name: "acme", CloneURL: "ssh"}},
		Repos:    []Repo{{Match: RepoMatch{ID: 1}, CloneURL: "https"}},
	}
	assert.Equal(t, "ssh", f.CloneURLFor("primary", "globex", 0, "x")) // defaults
	assert.Equal(t, "https", f.CloneURLFor("bot", "globex", 0, "x"))   // profile override
	assert.Equal(t, "ssh", f.CloneURLFor("bot", "acme", 0, "x"))       // owner override beats profile
	assert.Equal(t, "https", f.CloneURLFor("bot", "acme", 1, "x"))     // repo override beats all
}

func TestResolveToken(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/secret.token", []byte("  filetoken\n"), 0o600))
	require.NoError(t, afero.WriteFile(fs, "/loose.token", []byte("x"), 0o644))

	t.Run("token_file trimmed", func(t *testing.T) {
		tok, warn, err := Profile{TokenFile: "/secret.token"}.ResolveToken(fs, func(string) string { return "" }, "")
		require.NoError(t, err)
		assert.Equal(t, "filetoken", tok)
		assert.Empty(t, warn)
	})
	t.Run("token_file too permissive", func(t *testing.T) {
		_, _, err := Profile{TokenFile: "/loose.token"}.ResolveToken(fs, func(string) string { return "" }, "")
		assert.Error(t, err)
	})
	t.Run("token_env", func(t *testing.T) {
		tok, _, err := Profile{TokenEnv: "MY"}.ResolveToken(fs, func(k string) string { return map[string]string{"MY": "envtok"}[k] }, "")
		require.NoError(t, err)
		assert.Equal(t, "envtok", tok)
	})
	t.Run("fallback env for first profile", func(t *testing.T) {
		tok, _, err := Profile{}.ResolveToken(fs, func(k string) string { return map[string]string{"GITHUB_TOKEN": "gh"}[k] }, "GITHUB_TOKEN")
		require.NoError(t, err)
		assert.Equal(t, "gh", tok)
	})
	t.Run("inline warns", func(t *testing.T) {
		tok, warn, err := Profile{Token: "inline"}.ResolveToken(fs, func(string) string { return "" }, "")
		require.NoError(t, err)
		assert.Equal(t, "inline", tok)
		assert.NotEmpty(t, warn)
	})
	t.Run("none resolves to error", func(t *testing.T) {
		_, _, err := Profile{}.ResolveToken(fs, func(string) string { return "" }, "")
		assert.Error(t, err)
	})
}

func TestFetchConfigLookup(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/work/fetch.toml", []byte("x"), 0o600))
	require.NoError(t, afero.WriteFile(fs, "/home/op/.config/maintainer/fetch.yaml", []byte("x"), 0o600))

	t.Run("explicit empty disables", func(t *testing.T) {
		path, disabled := FetchConfigLookup{ExplicitSet: true, Explicit: ""}.Resolve(fs)
		assert.Empty(t, path)
		assert.True(t, disabled)
	})
	t.Run("cwd wins over xdg", func(t *testing.T) {
		path, _ := FetchConfigLookup{WorkDir: "/work", Home: "/home/op"}.Resolve(fs)
		assert.Equal(t, "/work/fetch.toml", path)
	})
	t.Run("xdg fallback", func(t *testing.T) {
		path, _ := FetchConfigLookup{WorkDir: "/nope", Home: "/home/op"}.Resolve(fs)
		assert.Equal(t, "/home/op/.config/maintainer/fetch.yaml", path)
	})
	t.Run("env override", func(t *testing.T) {
		getenv := func(k string) string { return map[string]string{EnvFetchConfig: "/work/fetch.toml"}[k] }
		path, _ := FetchConfigLookup{Getenv: getenv, WorkDir: "/nope", Home: "/nope"}.Resolve(fs)
		assert.Equal(t, "/work/fetch.toml", path)
	})
}
