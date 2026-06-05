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
	_, err = git.PlainCloneContext(ctx, opt.Path, false, &git.CloneOptions{
		URL:               opt.URL,
		Auth:              auth,
		RecurseSubmodules: git.NoRecurseSubmodules,
	})
	if err != nil {
		// go-git leaves a partial directory on cancel; remove it so the next
		// run is a clean clone (§11).
		_ = os.RemoveAll(opt.Path)
		// An empty upstream (created but never pushed) is not a failure: mirror
		// `git clone` of an empty repo — init the dir + origin, no checkout. It
		// becomes tracked and fills in on a later fetch once commits land.
		if errors.Is(err, transport.ErrEmptyRemoteRepository) {
			return s.initEmpty(opt)
		}
		return fmt.Errorf("clone %s: %w", opt.URL, err)
	}
	return nil
}

// initEmpty materialises an empty clone (dir + origin remote, no checkout),
// matching `git clone` of a repository with no commits.
func (s *Sync) initEmpty(opt CloneOptions) error {
	if err := os.MkdirAll(opt.Path, 0o755); err != nil {
		return fmt.Errorf("init empty %s: %w", opt.Path, err)
	}
	repo, err := git.PlainInit(opt.Path, false)
	if err != nil {
		_ = os.RemoveAll(opt.Path)
		return fmt.Errorf("init empty %s: %w", opt.Path, err)
	}
	if _, err := repo.CreateRemote(&config.RemoteConfig{
		Name:  git.DefaultRemoteName,
		URLs:  []string{opt.URL},
		Fetch: []config.RefSpec{defaultRefSpec},
	}); err != nil {
		_ = os.RemoveAll(opt.Path)
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
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
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
	remote.URLs = []string{url}
	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("update remote %s: %w", path, err)
	}
	return nil
}

// Inspect reads origin URLs and the short HEAD sha.
func (s *Sync) Inspect(path string) (CloneInfo, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return CloneInfo{}, fmt.Errorf("open %s: %w", path, err)
	}
	var info CloneInfo
	if remote, err := repo.Remote(git.DefaultRemoteName); err == nil {
		info.Origins = append(info.Origins, remote.Config().URLs...)
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
