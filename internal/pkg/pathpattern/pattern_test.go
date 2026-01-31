package pathpattern

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisibilityIsNotWildcard(t *testing.T) {
	p, err := Compile("{{.Visibility}}/{{.Owner}}/{{.Repo}}", "")
	require.NoError(t, err)
	for _, path := range []string{"public/old-owner/old-name", "private/a/b", "internal/a/b"} {
		assert.True(t, p.Match(path), path)
	}
	for _, path := range []string{"research/a/b", "prototyping/a/b", "public/a/b/nested", "public/a", "PUBLIC/a/b", "../a/b"} {
		assert.False(t, p.Match(path), path)
	}
	assert.True(t, p.Prefix("."))
	assert.True(t, p.Prefix("public"))
	assert.True(t, p.Prefix("public/alice"))
	assert.False(t, p.Prefix("research"))
	assert.False(t, p.Prefix("public/a/b/nested"))
}

func TestReorderedAndBoundOwner(t *testing.T) {
	p, err := Compile("mirror/{{lower .Owner}}/{{.Visibility | upper}}/repo-{{.Repo}}", "Acme")
	require.NoError(t, err)
	assert.True(t, p.Match("mirror/acme/PRIVATE/repo-tool"))
	assert.False(t, p.Match("mirror/other/PRIVATE/repo-tool"))
	assert.False(t, p.Match("mirror/acme/RESEARCH/repo-tool"))
	p, err = Compile("{{.IsFork}}/{{.IsTemplate}}/{{.IsArchived}}/{{.DefaultBranch}}/{{.Repo}}", "")
	require.NoError(t, err)
	assert.True(t, p.Match("true/false/true/main/tool"))
	assert.False(t, p.Match("maybe/false/true/main/tool"))
}

func TestUnsupportedFailsClosed(t *testing.T) {
	for _, source := range []string{"", "../{{.Repo}}", "/{{.Repo}}", "{{.Root}}/{{.Repo}}", "{{if .IsFork}}forks{{end}}/{{.Repo}}", "{{printf \"%s\" .Repo}}", "{{.Unknown}}", "{{.Repo}}//x", "{{.Repo}}/./x", "{{.Repo | lower | upper}}", "{{define \"x\"}}x{{end}}{{template \"x\"}}"} {
		_, err := Compile(source, "")
		assert.Error(t, err, source)
	}
}

func TestNested(t *testing.T) {
	compile := func(source, owner string) *Pattern {
		p, err := Compile(source, owner)
		require.NoError(t, err)
		return p
	}
	layout := compile("{{.Visibility}}/{{.Owner}}/{{.Repo}}", "")
	for _, tc := range []struct {
		layout, owner string
		nested        bool
	}{
		{"{{.Owner}}/{{.Repo}}", "{{.Repo}}", true},                 // acme/bob at bob, bob/x at bob/x
		{"{{.Owner}}/{{.Repo}}", "{{.Repo}}-mirror", true},          // acme/x at x-mirror, owner x-mirror below
		{"{{.Repo}}", "{{.Owner}}-{{.Repo}}/src", true},             // a repo named acme-x above acme/x
		{"{{.Owner}}/{{.Repo}}", "{{.Owner}}/{{.Repo}}/src", false}, // the layout never serves acme
		{"{{.Visibility}}/{{.Visibility}}/{{.Repo}}", "public/private/{{.Repo}}/src", false},
		{"{{.Visibility}}/{{.Visibility}}/{{.Repo}}", "public/public/{{.Repo}}/src", true},
		{"{{.Repo}}/{{.Repo}}", "a/b/{{.Repo}}", false},
		{"{{.Repo}}/{{.Repo}}", "a/a/{{.Repo}}", true},
		{"{{.Repo}}-{{.Repo}}", "a-b/{{.Repo}}", false},
		{"{{.Repo}}-{{.Repo}}", "a-a/{{.Repo}}", true},
		{"{{.Repo}}-x/{{.Repo}}", "{{.Repo}}-x/{{.Repo}}/src", true}, // solved from the second component
		{"{{lower .Repo}}", "Upper/{{.Repo}}", false},
		{"{{.Visibility}}/{{.Owner}}/{{.Repo}}", "{{.Visibility}}/{{.Owner}}", false},
		{"{{.Visibility}}/{{.Owner}}/{{.Repo}}", "public/{{.Repo}}", true},
	} {
		got := Nested(compile(tc.layout, ""), compile(tc.owner, "acme"), []string{"acme"})
		assert.Equal(t, tc.nested, got, "%s vs %s", tc.layout, tc.owner)
	}
	assert.False(t, Nested(layout, compile("mirror/{{.Owner}}/{{.Repo}}", "acme"), []string{"acme"}), "equal depth")
	assert.True(t, Nested(compile("mirror/{{.Repo}}", "acme"), compile("mirror/{{.Owner}}/{{.Repo}}", "bob"), []string{"acme", "bob"}))
	assert.False(t, Nested(compile("{{.Visibility}}/{{lower .Owner}}", "Acme"), layout, []string{"Acme"}))
}
