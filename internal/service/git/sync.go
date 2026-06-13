package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// defaultRefSpec fetches remote-tracking refs only, never local branches (§16).
const defaultRefSpec config.RefSpec = "+refs/heads/*:refs/remotes/origin/*"

// Sync is the go-git implementation of the GitSync port.
type Sync struct{}

// NewSync returns a go-git-backed GitSync.
func NewSync() *Sync { return new(Sync) }

var _ GitSync = (*Sync)(nil)

// Clone materialises a repository with submodules disabled (§2 non-goals).
func (s *Sync) Clone(ctx context.Context, opt CloneOptions) error {
	auth, err := authMethod(opt.Auth)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(opt.Path), 0o755); err != nil {
		return fmt.Errorf("prepare clone parent: %w", err)
	}
	// Reserve the final path ourselves before handing it to go-git. This is an
	// atomic existence check: a clone never starts in a path that predates this
	// invocation, even if the plan became stale between review and apply.
	if err := os.Mkdir(opt.Path, 0o755); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("clone target %s already exists; refusing to modify it", opt.Path)
		}
		return fmt.Errorf("reserve clone target %s: %w", opt.Path, err)
	}
	_, err = git.PlainCloneContext(ctx, opt.Path, false, &git.CloneOptions{
		URL:               opt.URL,
		Auth:              auth,
		RecurseSubmodules: git.NoRecurseSubmodules,
	})
	if err != nil {
		// An empty upstream (created but never pushed) is not a failure: mirror
		// `git clone` of an empty repo — init the dir + origin, no checkout. It
		// becomes tracked and fills in on a later fetch once commits land.
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			if initErr := s.initEmpty(opt); initErr == nil {
				return nil
			} else {
				return cloneFailure(opt, initErr)
			}
		}
		return cloneFailure(opt, err)
	}
	return nil
}

// cloneFailure removes only an empty directory reserved by Clone. A partial
// checkout is deliberately retained for inspection: recursive cleanup cannot
// prove that every entry still belongs to this invocation.
func cloneFailure(opt CloneOptions, cause error) error {
	if err := os.Remove(opt.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(
			"clone %s: %w; partial checkout retained at %s (refusing recursive cleanup: %v)",
			opt.URL, cause, opt.Path, err,
		)
	}
	return fmt.Errorf("clone %s: %w", opt.URL, cause)
}

// initEmpty materialises an empty clone (dir + origin remote, no checkout),
// matching `git clone` of a repository with no commits.
func (s *Sync) initEmpty(opt CloneOptions) error {
	if err := os.MkdirAll(opt.Path, 0o755); err != nil {
		return fmt.Errorf("init empty %s: %w", opt.Path, err)
	}
	repo, err := git.PlainInit(opt.Path, false)
	if err != nil {
		return fmt.Errorf("init empty %s: %w", opt.Path, err)
	}
	if _, err := repo.CreateRemote(&config.RemoteConfig{
		Name:  git.DefaultRemoteName,
		URLs:  []string{opt.URL},
		Fetch: []config.RefSpec{defaultRefSpec},
	}); err != nil {
		return fmt.Errorf("init empty %s: set origin: %w", opt.Path, err)
	}
	return nil
}

// Fetch updates remote-tracking refs with prune; the working tree is untouched.
func (s *Sync) Fetch(ctx context.Context, path string, auth Auth) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	method, err := authMethod(auth)
	if err != nil {
		return err
	}
	err = repo.FetchContext(ctx, &git.FetchOptions{
		Auth:     method,
		RefSpecs: []config.RefSpec{defaultRefSpec},
		Prune:    true,
	})
	// A reachable zero-ref upstream is a valid no-op. Retry it on every run so
	// the first branch created later is discovered without persistent flags.
	if err != nil &&
		!errors.Is(err, git.NoErrAlreadyUpToDate) &&
		!errors.Is(err, transport.ErrEmptyRemoteRepository) {
		return fmt.Errorf("fetch %s: %w", path, err)
	}
	return nil
}

// Move renames from→to. A cross-device rename (EXDEV) is reported as an error;
// there is no copy+remove fallback in the PoC (§7.5).
func (s *Sync) Move(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return fmt.Errorf("prepare move target: %w", err)
	}
	if _, err := os.Lstat(to); err == nil {
		return fmt.Errorf("move target %s already exists; refusing to overwrite it", to)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect move target %s: %w", to, err)
	}
	if err := os.Rename(from, to); err != nil {
		if errors.Is(err, syscall.EXDEV) {
			return fmt.Errorf("cross-device move %s → %s is not supported in the PoC: %w", from, to, err)
		}
		return fmt.Errorf("move %s → %s: %w", from, to, err)
	}
	return nil
}

// UpdateRemote rewrites origin's URL.
func (s *Sync) UpdateRemote(path, url string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	remote, ok := cfg.Remotes[git.DefaultRemoteName]
	if !ok {
		return fmt.Errorf("%s has no %q remote", path, git.DefaultRemoteName)
	}
	// go-git exposes remote.*.url and remote.*.pushurl in one URLs slice.
	// Rewrite the fetch endpoint only and leave push endpoints (including a
	// `git lock` sentinel) untouched in the raw config.
	remote.URLs = []string{url}
	if cfg.Raw != nil && cfg.Raw.HasSection("remote") {
		section := cfg.Raw.Section("remote")
		if section.HasSubsection(git.DefaultRemoteName) {
			section.Subsection(git.DefaultRemoteName).SetOption("url", url)
		}
	}
	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("update remote %s: %w", path, err)
	}
	return nil
}

// Inspect reads origin fetch/push URLs separately and the short HEAD sha.
func (s *Sync) Inspect(path string) (CloneInfo, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return CloneInfo{}, fmt.Errorf("open %s: %w", path, err)
	}
	var info CloneInfo
	cfg, err := repo.Config()
	if err != nil {
		return CloneInfo{}, fmt.Errorf("read config %s: %w", path, err)
	}
	if _, ok := cfg.Remotes[git.DefaultRemoteName]; ok {
		if cfg.Raw == nil || !cfg.Raw.HasSection("remote") {
			return CloneInfo{}, fmt.Errorf("read config %s: origin has no inspectable raw configuration", path)
		}
		section := cfg.Raw.Section("remote")
		if !section.HasSubsection(git.DefaultRemoteName) {
			return CloneInfo{}, fmt.Errorf("read config %s: origin has no inspectable raw configuration", path)
		}
		origin := section.Subsection(git.DefaultRemoteName)
		info.Origins = append(info.Origins, origin.OptionAll("url")...)
		info.PushURLs = append(info.PushURLs, origin.OptionAll("pushurl")...)
	}
	if head, err := repo.Head(); err == nil {
		sha := head.Hash().String()
		if len(sha) > 7 {
			sha = sha[:7]
		}
		info.HeadShort = sha
	}
	return info, nil
}

// authMethod resolves per-operation credentials (§5.4): HTTPS+PAT via
// BasicAuth (token never written into the remote URL); SSH best-effort via a
// running ssh-agent with strict known-hosts checking.
func authMethod(a Auth) (transport.AuthMethod, error) {
	switch a.Transport {
	case "https":
		if a.Token == "" {
			return nil, nil // public over https
		}
		return &githttp.BasicAuth{Username: "x-access-token", Password: a.Token}, nil
	case "ssh", "":
		method, err := gitssh.NewSSHAgentAuth("git")
		if err != nil {
			return nil, fmt.Errorf("ssh-agent auth (configure ssh-agent outside maintainer): %w", err)
		}
		return method, nil
	default:
		return nil, fmt.Errorf("unsupported transport %q", a.Transport)
	}
}
