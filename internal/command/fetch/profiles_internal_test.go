package fetch

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	fetchsvc "go.octolab.org/toolset/maintainer/internal/service/fetch"
)

// --token reaches the first selected profile, never the first configured one
// or every profile; --owner matches configured owners regardless of case.
func TestResolveProfilesSelection(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("ALPHA_TOKEN", "alpha")
	t.Setenv("BETA_TOKEN", "beta")
	fcfg := &config.Fetch{Source: "fetch.toml", Profiles: map[string]config.Profile{
		"alpha": {TokenEnv: "ALPHA_TOKEN", IncludeOwners: []string{"Acme"}},
		"beta":  {TokenEnv: "BETA_TOKEN", IncludeOwners: []string{"acme-bot", "ACME"}},
	}}
	for name, tc := range map[string]struct {
		opts options
		want []fetchsvc.ResolvedProfile
	}{
		"--profile beta --token T": {
			opts: options{profiles: []string{"beta"}, token: "T"},
			want: []fetchsvc.ResolvedProfile{{Name: "beta", Token: "T", Owners: []string{"acme-bot", "ACME"}}},
		},
		"--token T": {
			opts: options{token: "T"},
			want: []fetchsvc.ResolvedProfile{
				{Name: "alpha", Token: "T", Owners: []string{"Acme"}},
				{Name: "beta", Token: "beta", Owners: []string{"acme-bot", "ACME"}},
			},
		},
		"--owner acme-bot --token T": {
			opts: options{owners: []string{"Acme-Bot"}, token: "T"},
			want: []fetchsvc.ResolvedProfile{{Name: "beta", Token: "T", Owners: []string{"acme-bot"}}},
		},
		"--owner acme": {
			opts: options{owners: []string{"acme"}},
			want: []fetchsvc.ResolvedProfile{
				{Name: "alpha", Token: "alpha", Owners: []string{"Acme"}},
				{Name: "beta", Token: "beta", Owners: []string{"ACME"}},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := resolveProfiles(afero.NewMemMapFs(), fcfg, &tc.opts, nil)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
