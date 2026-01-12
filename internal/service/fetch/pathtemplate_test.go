package fetch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/config"
	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

const (
	home = "/home/op"
	cwd  = "/work"
)

func snap() github.RepoSnapshot {
	return github.RepoSnapshot{
		ID: 1, Owner: "acme", Name: "service",
		Visibility: github.Public, DefaultBranch: "main",
	}
}

func TestPathRenderer_Render(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		tmpl     string
		external bool
		mutate   func(*github.RepoSnapshot)
		want     string
		wantErr  bool
	}{
		{
			name: "default template under relative root",
			root: ".", tmpl: config.DefaultPath,
			want: "/work/public/acme/service",
		},
		{
			name: "absolute root",
			root: "/srv/code", tmpl: "{{.Owner}}/{{.Repo}}",
			want: "/srv/code/acme/service",
		},
		{
			name: "private visibility",
			root: ".", tmpl: "{{.Visibility}}/{{.Owner}}/{{.Repo}}",
			mutate: func(s *github.RepoSnapshot) { s.Visibility = github.Private },
			want:   "/work/private/acme/service",
		},
		{
			name: "lower func",
			root: ".", tmpl: "{{.Owner | lower}}/{{.Repo}}",
			mutate: func(s *github.RepoSnapshot) { s.Owner = "ACME" },
			want:   "/work/acme/service",
		},
		{
			name: "external absolute path allowed",
			root: ".", tmpl: "/opt/special", external: true,
			want: "/opt/special",
		},
		{
			name: "external tilde expands to home",
			root: ".", tmpl: "~/.dotfiles", external: true,
			want: "/home/op/.dotfiles",
		},
		{
			name: "non-external absolute rejected by containment",
			root: "/srv", tmpl: "/etc/passwd", external: false,
			wantErr: true,
		},
		{
			name: "escape via parent rejected",
			root: "/srv", tmpl: "../../etc", external: false,
			wantErr: true,
		},
		{
			name: "bad template syntax",
			root: ".", tmpl: "{{.Nope", external: false,
			wantErr: true,
		},
		{
			name: "unknown variable errors",
			root: ".", tmpl: "{{.Bogus}}", external: false,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewPathRenderer(tc.root, home, cwd)
			require.NoError(t, err)
			s := snap()
			if tc.mutate != nil {
				tc.mutate(&s)
			}
			got, err := r.Render(tc.tmpl, s, tc.external)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPathResolver_OverridePrecedence(t *testing.T) {
	cnf := &config.Fetch{
		Defaults: config.Defaults{Root: ".", Path: config.DefaultPath, CloneURL: "ssh", Concurrency: 1},
		Owners:   []config.Owner{{Name: "acme", Path: "mirror/{{.Owner}}/{{.Repo}}"}},
		Repos: []config.Repo{
			{Match: config.RepoMatch{ID: 1}, Path: "~/Code/special"},
		},
	}
	cnf.Validate() // applies nothing but exercises the validator
	r, err := NewPathRenderer(cnf.Defaults.Root, home, cwd)
	require.NoError(t, err)
	resolver := NewPathResolver(cnf, r)

	// per-repo override (external, id match) wins.
	got, err := resolver.Resolve(snap())
	require.NoError(t, err)
	assert.Equal(t, "/home/op/Code/special", got)
	assert.True(t, resolver.External(snap()))

	// a snapshot without a repo override falls back to the per-owner template.
	other := github.RepoSnapshot{ID: 2, Owner: "acme", Name: "tool", Visibility: github.Public}
	got, err = resolver.Resolve(other)
	require.NoError(t, err)
	assert.Equal(t, "/work/mirror/acme/tool", got)
	assert.False(t, resolver.External(other))

	// a different owner falls back to defaults.path.
	third := github.RepoSnapshot{ID: 3, Owner: "globex", Name: "thing", Visibility: github.Private}
	got, err = resolver.Resolve(third)
	require.NoError(t, err)
	assert.Equal(t, "/work/private/globex/thing", got)
}
