package fetch

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	giturls "github.com/whilp/git-urls"

	"go.octolab.org/toolset/maintainer/internal/config"
	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/service/workspace"
)

const githubHost = "github.com"

// incompleteClone explains the conflict of a clone that stopped before it
// finished (§7.5).
const incompleteClone = "incomplete clone: an earlier clone stopped before it finished; remove the directory to clone it again"

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
	cnf      *config.Fetch // ignore rules; nil ignores nothing
}

// NewAdopter wires an Adopter from the Git port and a redirect resolver.
func NewAdopter(git gitsvc.GitSync, resolver NameResolver) *Adopter {
	return &Adopter{git: git, resolver: resolver}
}

// Ignoring makes the scan skip the name lookup of a checkout whose origin
// matches an ignore rule of cnf, so an ignored repository cannot fail it.
func (a *Adopter) Ignoring(cnf *config.Fetch) *Adopter {
	a.cnf = cnf
	return a
}

// Scan walks root plus the explicit per-repo override paths (which live outside
// root and would otherwise never be re-discovered), inspects every clone, and
// resolves each to a stable id. Snapshots are consulted first to avoid an API
// round trip; only an (owner,name) miss falls back to the redirect resolver.
func (a *Adopter) Scan(ctx context.Context, root string, snapshots []github.RepoSnapshot, cnf *config.Fetch, extraPaths ...string) ([]DiskClone, error) {
	if cnf == nil {
		cnf = &config.Fetch{}
	}
	home, _ := os.UserHomeDir()
	scope, err := workspace.New(cnf, root, home)
	if err != nil {
		return nil, err
	}
	scoped := *a
	scoped.cnf = cnf
	return scoped.ScanScoped(ctx, scope, snapshots, append(externalPaths(cnf), extraPaths...)...)
}

// ScanScoped consumes the same bounded local inventory as status.
func (a *Adopter) ScanScoped(ctx context.Context, scope *workspace.Scope, snapshots []github.RepoSnapshot, extraPaths ...string) ([]DiskClone, error) {
	// GitHub names are case-insensitive: an origin spelled in another case
	// names the same repository.
	byName := make(map[string]int64, len(snapshots))
	for _, s := range snapshots {
		byName[strings.ToLower(s.Owner+"/"+s.Name)] = s.ID
	}

	seen := make(map[string]bool)
	var clones []DiskClone

	add := func(dir string, pinned bool) error {
		if seen[dir] {
			return nil
		}
		seen[dir] = true
		clone, ok, err := a.inspect(ctx, dir, byName)
		if err != nil {
			return err
		}
		if ok {
			clone.Pinned = pinned
			if pinned && clone.ID == 0 && clone.Error == "" {
				clone.Error = "pinned checkout identity could not be verified"
			}
			clones = append(clones, clone)
		} else if pinned {
			clones = append(clones, DiskClone{Path: dir, Pinned: true, Error: "pinned path is missing or is not a supported GitHub checkout"})
		}
		return nil
	}

	// An obstruction becomes a report-only conflict: the scan is incomplete
	// there, and the plan must say so rather than pass for a complete one.
	obstruct := func(dir string, reason error) error {
		seen[dir] = true
		clones = append(clones, DiskClone{Path: dir, Error: "scan: " + reason.Error()})
		return nil
	}
	if err := scope.Walk(ctx, add, obstruct); err != nil {
		return nil, err
	}
	for _, ext := range extraPaths {
		if err := add(ext, true); err != nil {
			return nil, err
		}
	}
	return clones, nil
}

func (a *Adopter) inspect(ctx context.Context, dir string, byName map[string]int64) (DiskClone, bool, error) {
	info, err := a.git.Inspect(dir)
	if err != nil {
		return DiskClone{}, false, nil // not a usable git repo; skip silently
	}
	if info.Incomplete {
		// An interrupted clone is neither adopted nor fetched: it is reported,
		// attributed by name when its origin says which repository it was.
		clone := DiskClone{Path: dir, Origins: len(info.Origins), Error: incompleteClone}
		if len(info.Origins) == 1 {
			if host, owner, name, _, ok := parseRemote(info.Origins[0]); ok && host == githubHost {
				clone.Owner, clone.Name = owner, name
				clone.ID = byName[strings.ToLower(owner+"/"+name)]
			}
		}
		return clone, true, nil
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
		Worktrees: info.Worktrees,
	}
	if clone.Origins > 1 {
		return clone, true, nil // ambiguous; ID stays 0 → Planner flags a conflict
	}

	if id, found := byName[strings.ToLower(owner+"/"+name)]; found {
		clone.ID = id
		return clone, true, nil
	}
	if a.cnf != nil && a.cnf.Ignored(0, owner, name) {
		return clone, true, nil // ID stays 0: never looked up, never planned
	}
	if a.resolver != nil {
		id, err := a.resolver.ResolveByName(ctx, owner, name)
		if err != nil {
			switch github.HTTPStatus(err) {
			case http.StatusUnauthorized, http.StatusForbidden, http.StatusUnavailableForLegalReasons:
				// Access to this one repository is denied: its checkout becomes
				// a conflict, and the rest of the run proceeds.
				clone.Error = fmt.Sprintf("checkout identity could not be verified: resolve %s/%s: %v", owner, name, err)
				return clone, true, nil
			}
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
