package fetch

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// templateVars are the variables available to a path template (§4.3).
type templateVars struct {
	Root          string
	Owner         string
	Repo          string
	Visibility    string
	DefaultBranch string
	IsFork        bool
	IsTemplate    bool
	IsArchived    bool
}

// tmplFuncs ships lower/upper helpers but applies neither by default (§15).
var tmplFuncs = template.FuncMap{
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
}

// PathRenderer renders a single template string into a cleaned absolute path
// per the resolution rule in §4.3. It is pure: $HOME and root are injected.
type PathRenderer struct {
	root string // expanded, absolute root
	home string
}

// NewPathRenderer expands defaults.root (supports ~ and relative-to-cwd) into
// an absolute path. cwd is used to resolve a relative root; home expands ~.
func NewPathRenderer(root, home, cwd string) (*PathRenderer, error) {
	expanded, err := expand(root, home, cwd)
	if err != nil {
		return nil, fmt.Errorf("expand root %q: %w", root, err)
	}
	return &PathRenderer{root: expanded, home: home}, nil
}

// Root returns the expanded absolute root.
func (r *PathRenderer) Root() string { return r.root }

// Render renders tmpl for snap and resolves the result. external allows the
// rendered path to be absolute or ~-rooted (per-repo overrides only); a
// non-external result that escapes root is an error (§4.3).
func (r *PathRenderer) Render(tmpl string, snap github.RepoSnapshot, external bool) (string, error) {
	t, err := template.New("path").Funcs(tmplFuncs).Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse path template %q: %w", tmpl, err)
	}
	var buf strings.Builder
	vars := templateVars{
		Root:          r.root,
		Owner:         snap.Owner,
		Repo:          snap.Name,
		Visibility:    string(snap.Visibility),
		DefaultBranch: snap.DefaultBranch,
		IsFork:        snap.IsFork,
		IsTemplate:    snap.IsTemplate,
		IsArchived:    snap.IsArchived,
	}
	if err := t.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("render path template %q: %w", tmpl, err)
	}

	resolved, err := expand(buf.String(), r.home, r.root)
	if err != nil {
		return "", err
	}
	if !external {
		within, err := filepath.Rel(r.root, resolved)
		if err != nil || within == ".." || strings.HasPrefix(within, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("rendered path %q escapes root %q", resolved, r.root)
		}
	}
	return resolved, nil
}

// expand applies the §4.3 resolution rule to a rendered string:
// absolute → as-is; ~ → relative to home; otherwise join with base.
func expand(s, home, base string) (string, error) {
	switch {
	case filepath.IsAbs(s):
		return filepath.Clean(s), nil
	case s == "~" || strings.HasPrefix(s, "~/"):
		if home == "" {
			return "", fmt.Errorf("cannot expand %q: $HOME is empty", s)
		}
		return filepath.Clean(filepath.Join(home, strings.TrimPrefix(s, "~"))), nil
	case strings.HasPrefix(s, "~"):
		return "", fmt.Errorf("cannot expand %q: ~user form is not supported", s)
	default:
		if base == "" {
			return filepath.Clean(s), nil
		}
		return filepath.Clean(filepath.Join(base, s)), nil
	}
}

// PathResolver picks the applicable template from the override chain
// (per-repo → per-owner → defaults) and renders it (§4.3). It is pure.
type PathResolver struct {
	cnf      *config.Fetch
	renderer *PathRenderer
}

// NewPathResolver wires a resolver from config and an expanded renderer.
func NewPathResolver(cnf *config.Fetch, renderer *PathRenderer) *PathResolver {
	return &PathResolver{cnf: cnf, renderer: renderer}
}

// Resolve returns the absolute target path for a snapshot.
func (pr *PathResolver) Resolve(snap github.RepoSnapshot) (string, error) {
	// Per-repo override wins; it may be absolute/~ (external).
	if r, ok := pr.cnf.RepoOverride(snap.ID, snap.Owner, snap.Name); ok && r.Path != "" {
		external := filepath.IsAbs(r.Path) || strings.HasPrefix(r.Path, "~")
		return pr.renderer.Render(r.Path, snap, external)
	}
	// Per-owner override next.
	if o, ok := pr.cnf.OwnerOverride(snap.Owner); ok && o.Path != "" {
		return pr.renderer.Render(o.Path, snap, false)
	}
	// Defaults.
	return pr.renderer.Render(pr.cnf.Defaults.Path, snap, false)
}

// External reports whether the resolved path for a snapshot lives outside root
// via a per-repo override (adopt/fetch-only, never auto-moved — §16).
func (pr *PathResolver) External(snap github.RepoSnapshot) bool {
	if r, ok := pr.cnf.RepoOverride(snap.ID, snap.Owner, snap.Name); ok && r.Path != "" {
		return filepath.IsAbs(r.Path) || strings.HasPrefix(r.Path, "~")
	}
	return false
}
