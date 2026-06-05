package fetch_test

import (
	"testing"

	. "go.octolab.org/toolset/maintainer/internal/service/fetch"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// FuzzPathRenderer asserts the renderer rejects bad templates without panicking
// (§14 fuzz target). A non-nil error is fine; a panic is not.
func FuzzPathRenderer(f *testing.F) {
	for _, seed := range []string{
		"{{.Owner}}/{{.Repo}}",
		"{{.Visibility}}/{{.Owner}}/{{.Repo}}",
		"{{.Owner | lower}}",
		"{{.Bogus}}",
		"{{.Owner",
		"../escape",
		"~/abs",
		"",
		"{{range}}",
	} {
		f.Add(seed)
	}

	r, err := NewPathRenderer(".", "/home/op", "/work")
	if err != nil {
		f.Fatal(err)
	}
	snap := github.RepoSnapshot{ID: 1, Owner: "acme", Name: "svc", Visibility: github.Public, DefaultBranch: "main"}

	f.Fuzz(func(t *testing.T, tmpl string) {
		// external=false and external=true must both stay panic-free.
		_, _ = r.Render(tmpl, snap, false)
		_, _ = r.Render(tmpl, snap, true)
	})
}
