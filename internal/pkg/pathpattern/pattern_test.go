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
