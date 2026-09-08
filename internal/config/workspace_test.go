package config_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/config"
)

func TestWorkspaceConfig(t *testing.T) {
	for _, tc := range []struct{ file, body string }{
		{"fetch.toml", "[workspace]\nroot = '/code'\npins = ['prototyping']\n"},
		{"fetch.yaml", "workspace:\n  root: /code\n  pins: [prototyping]\n"},
	} {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, tc.file, []byte(tc.body), 0o600))
		cnf, err := LoadFetch(fs, tc.file)
		require.NoError(t, err)
		require.NoError(t, cnf.Validate())
		assert.Equal(t, Workspace{Root: "/code", Path: DefaultPath, Pins: []string{"prototyping"}}, cnf.WorkspaceConfig())
		assert.Empty(t, cnf.Defaults.Root)
	}
}

func TestWorkspaceConfigRejected(t *testing.T) {
	for _, body := range []string{
		"[defaults]\nroot = '/old'\n[workspace]\nroot = '/new'",
		"[workspaces.one]\nroot = '/code'",
		"[workspaces]\n",
		"[workspace]\npins = ['']",
		"[workspace]\npins = ['~someone/code']",
		"[workspace]\npath = '{{if .IsFork}}fork{{end}}/{{.Repo}}'",
	} {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "fetch.toml", []byte(body), 0o600))
		cnf, err := LoadFetch(fs, "fetch.toml")
		if err == nil {
			err = cnf.Validate()
		}
		assert.Error(t, err, body)
	}
}

func TestWorkspaceTemplateAndLegacy(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "fetch.toml", []byte(FetchConfigTemplate), 0o600))
	cnf, err := LoadFetch(fs, "fetch.toml")
	require.NoError(t, err)
	require.NoError(t, cnf.Validate())
	legacy, err := LoadFetch(fs, "")
	require.NoError(t, err)
	assert.Equal(t, legacy.WorkspaceConfig().Root, cnf.WorkspaceConfig().Root)
	assert.Equal(t, legacy.WorkspaceConfig().Path, cnf.WorkspaceConfig().Path)
}
