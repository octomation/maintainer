package fetch

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	giturls "github.com/whilp/git-urls"

	"go.octolab.org/toolset/maintainer/internal/config"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
)

const githubHost = "github.com"

// NameResolver resolves an on-disk owner/name to a stable GitHub id, following
// the rename redirect (GET /repos/{owner}/{name}); it returns 0 on a 404 (§4.4).
type NameResolver interface {
	ResolveByName(ctx context.Context, owner, name string) (int64, error)
}

// Adopter walks the checkout tree and matches clones to GitHub repositories by
// their remote URL, producing the DiskClone facts the Planner consumes (§9).
type Adopter struct {
	git      gitsvc.GitSync
	resolver NameResolver
}

// NewAdopter wires an Adopter from the Git port and a redirect resolver.
func NewAdopter(git gitsvc.GitSync, resolver NameResolver) *Adopter {
	return &Adopter{git: git, resolver: resolver}
}

// Scan walks root plus the explicit per-repo override paths (which live outside
// root and would otherwise never be re-discovered), inspects every clone, and
// resolves each to a stable id. Snapshots are consulted first to avoid an API
// round trip; only an (owner,name) miss falls back to the redirect resolver.
func (a *Adopter) Scan(ctx context.Context, root string, snapshots []github.RepoSnapshot, cnf *config.Fetch) ([]DiskClone, error) {
	byName := make(map[string]int64, len(snapshots))
	for _, s := range snapshots {
		byName[s.Owner+"/"+s.Name] = s.ID
	}

	seen := make(map[string]bool)
	var clones []DiskClone

	add := func(dir string) error {
		if seen[dir] {
			return nil
		}
		seen[dir] = true
		clone, ok, err := a.inspect(ctx, dir, byName)
		if err != nil {
			return err
		}
		if ok {
			clones = append(clones, clone)
		}
		return nil
	}

	if root != "" {
		if err := a.walk(root, add); err != nil {
			return nil, err
		}
	}
	for _, ext := range externalPaths(cnf) {
		if info, err := os.Stat(ext); err == nil && info.IsDir() {
			if err := add(ext); err != nil {
				return nil, err
			}
		}
	}
	return clones, nil
}

// walk descends root, treating any directory that contains a .git entry as a
// clone and not descending into it.
func (a *Adopter) walk(root string, add func(string) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if _, statErr := os.Stat(filepath.Join(path, ".git")); statErr == nil {
			if aerr := add(path); aerr != nil {
				return aerr
			}
			return filepath.SkipDir
		}
		return nil
	})
}

func (a *Adopter) inspect(ctx context.Context, dir string, byName map[string]int64) (DiskClone, bool, error) {
	info, err := a.git.Inspect(dir)
	if err != nil {
		return DiskClone{}, false, nil // not a usable git repo; skip silently
	}
	if len(info.Origins) == 0 {
		return DiskClone{}, false, nil
	}

	host, owner, name, transport, ok := parseRemote(info.Origins[0])
	if !ok || host != githubHost {
		return DiskClone{}, false, nil
	}

	clone := DiskClone{
		Path:      dir,
		RemoteURL: canonicalURL(transport, owner, name),
		Transport: transport,
		Owner:     owner,
		Name:      name,
		Origins:   len(info.Origins),
	}
	if clone.Origins > 1 {
		return clone, true, nil // ambiguous; ID stays 0 → Planner flags a conflict
	}

	if id, found := byName[owner+"/"+name]; found {
		clone.ID = id
		return clone, true, nil
	}
	if a.resolver != nil {
		id, err := a.resolver.ResolveByName(ctx, owner, name)
		if err != nil {
			return DiskClone{}, false, fmt.Errorf("resolve %s/%s: %w", owner, name, err)
		}
		clone.ID = id
	}
	return clone, true, nil
}

// parseRemote normalises a remote URL into (host, owner, name, transport).
func parseRemote(raw string) (host, owner, name, transport string, ok bool) {
	u, err := giturls.Parse(raw)
	if err != nil {
		return "", "", "", "", false
	}
	host = u.Host
	transport = config.TransportSSH
	if u.Scheme == "http" || u.Scheme == "https" {
		transport = config.TransportHTTPS
	}
	path := strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), ".git")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", "", false
	}
	return host, parts[0], parts[1], transport, true
}

// externalPaths returns per-repo override paths that point outside root.
func externalPaths(cnf *config.Fetch) []string {
	if cnf == nil {
		return nil
	}
	var out []string
	for i := range cnf.Repos {
		p := cnf.Repos[i].Path
		if p == "" {
			continue
		}
		if filepath.IsAbs(p) {
			out = append(out, filepath.Clean(p))
			continue
		}
		if strings.HasPrefix(p, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				out = append(out, filepath.Clean(filepath.Join(home, strings.TrimPrefix(p, "~"))))
			}
		}
	}
	return out
}
