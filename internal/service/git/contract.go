package git

import (
	"context"

	"github.com/go-git/go-git/v5"
)

//go:generate mockgen -source $GOFILE -destination mocks_test.go -package ${GOPACKAGE}_test

// Repository represents a Git repository.
type Repository interface {
	Remotes() ([]*git.Remote, error)
}

// Auth carries per-operation credentials for clone/fetch (§5.4). Nothing
// secret is ever persisted; credentials are supplied per operation.
type Auth struct {
	Transport string // ssh | https
	Token     string // PAT for https + BasicAuth
}

// CloneOptions parameterises a clone.
type CloneOptions struct {
	URL  string
	Path string
	Auth Auth
}

// CloneInfo is the on-disk view of a clone the Adopter and Planner need (§4.4).
type CloneInfo struct {
	Origins   []string // origin remote URLs (usually exactly one)
	HeadShort string   // short HEAD sha, for the fetch display line
}

// GitSync is the port the fetcher drives for all disk-mutating Git work (§2.3).
// The default implementation wraps go-git; a future shell-out fallback can
// replace it without touching the Planner.
type GitSync interface {
	// Clone materialises opt.URL into opt.Path (no submodules, no working-tree
	// surprises). It honours context cancellation.
	Clone(ctx context.Context, opt CloneOptions) error
	// Fetch runs the equivalent of `git fetch --prune` on origin: it updates
	// remote-tracking refs only and never touches the working tree (§16).
	Fetch(ctx context.Context, path string, auth Auth) error
	// Move renames a directory the fetcher owns (same-volume rename only).
	Move(from, to string) error
	// UpdateRemote rewrites origin's URL to the given credential-free URL.
	UpdateRemote(path, url string) error
	// Inspect reads a clone's origin URLs and HEAD for adoption/planning.
	Inspect(path string) (CloneInfo, error)
}
