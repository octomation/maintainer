package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// defaultRefSpec fetches remote-tracking refs only, never local branches (§16).
const defaultRefSpec config.RefSpec = "+refs/heads/*:refs/remotes/origin/*"

// IncompleteMarker is the file in the Git directory of a clone in progress. It
// is written before the transfer starts and removed once the checkout is
// complete, so a clone cut by a timeout, a signal or a dropped connection
// stays recognisable as incomplete.
const IncompleteMarker = "maintainer-clone-incomplete"

// Sync is the go-git implementation of the GitSync port; UpdateRemote edits
// the config through the git CLI.
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
	// Mark the clone incomplete before it starts. The marker also keeps go-git
	// from removing the reserved directory's contents recursively on failure.
	marker := filepath.Join(opt.Path, git.GitDirName, IncompleteMarker)
	if err := os.Mkdir(filepath.Dir(marker), 0o755); err != nil {
		return cloneFailure(opt, err)
	}
	if err := os.WriteFile(marker, []byte("maintainer fetch: this clone has not finished\n"), 0o644); err != nil {
		return cloneFailure(opt, err)
	}
	_, err = git.PlainCloneContext(ctx, opt.Path, false, &git.CloneOptions{
		URL:               opt.URL,
		Auth:              auth,
		RecurseSubmodules: git.NoRecurseSubmodules,
	})
	// An empty upstream (created but never pushed) is not a failure: mirror
	// `git clone` of an empty repo — init the dir + origin, no checkout. It
	// becomes tracked and fills in on a later fetch once commits land.
	if errors.Is(err, transport.ErrEmptyRemoteRepository) {
		err = s.initEmpty(opt)
	}
	if err != nil {
		return cloneFailure(opt, err)
	}
	if err := os.Remove(marker); err != nil {
		return fmt.Errorf("clone %s: complete %s: %w", opt.URL, opt.Path, err)
	}
	return nil
}

// cloneFailure removes what a failed clone provably owns alone: the reserved
// directory with the bare skeleton of a Git directory that received no data.
// Anything else is retained for inspection, marked incomplete: recursive
// cleanup cannot prove that every entry still belongs to this invocation.
func cloneFailure(opt CloneOptions, cause error) error {
	if removeSkeleton(opt.Path) {
		return fmt.Errorf("clone %s: %w", opt.URL, cause)
	}
	return fmt.Errorf("clone %s: %w; incomplete clone retained at %s", opt.URL, cause, opt.Path)
}

// skeleton lists, deepest first, the entries an initialised but empty clone
// consists of; nothing else may be present for removeSkeleton to act.
var skeleton = []string{
	filepath.Join(git.GitDirName, "objects", "info"),
	filepath.Join(git.GitDirName, "objects", "pack"),
	filepath.Join(git.GitDirName, "objects"),
	filepath.Join(git.GitDirName, "refs", "heads"),
	filepath.Join(git.GitDirName, "refs", "tags"),
	filepath.Join(git.GitDirName, "refs"),
	filepath.Join(git.GitDirName, "HEAD"),
	filepath.Join(git.GitDirName, "config"),
	filepath.Join(git.GitDirName, IncompleteMarker),
	git.GitDirName,
}

// removeSkeleton removes path when it holds nothing but the skeleton, one
// entry at a time and never recursively; it reports whether path is gone.
func removeSkeleton(path string) bool {
	allowed := make(map[string]bool, len(skeleton))
	for _, entry := range skeleton {
		allowed[entry] = true
	}
	foreign := false
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(path, p)
		if rel == "." {
			return nil
		}
		if !allowed[rel] || d.Type()&fs.ModeSymlink != 0 {
			foreign = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) || foreign {
		return false
	}
	for _, entry := range skeleton {
		if err := os.Remove(filepath.Join(path, entry)); err != nil && !os.IsNotExist(err) {
			return false
		}
	}
	return os.Remove(path) == nil || isNotExist(path)
}

func isNotExist(path string) bool {
	_, err := os.Lstat(path)
	return os.IsNotExist(err)
}

// initEmpty completes the clone of a repository with no commits, matching
// `git clone` of it: the directory and its origin remote, no checkout.
func (s *Sync) initEmpty(opt CloneOptions) error {
	repo, err := git.PlainOpen(opt.Path)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = git.PlainInit(opt.Path, false)
	}
	if err != nil {
		return fmt.Errorf("init empty %s: %w", opt.Path, err)
	}
	if _, err := repo.Remote(git.DefaultRemoteName); err == nil {
		return nil
	} else if !errors.Is(err, git.ErrRemoteNotFound) {
		return fmt.Errorf("init empty %s: read origin: %w", opt.Path, err)
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

// UpdateRemote rewrites origin's fetch URL. It goes through Git rather than
// go-git's SetConfig, which re-encodes the whole file: Git replaces only the
// remote.origin.url lines, so push URLs (including a `git lock` sentinel),
// other remotes, valueless keys, comments and insteadOf rules stay as written.
// The raw value is read back the way Inspect reads it, without insteadOf.
func (s *Sync) UpdateRemote(path, url string) error {
	const key = "remote." + git.DefaultRemoteName + ".url"
	urls, err := localValues(path, key)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if len(urls) == 0 {
		return fmt.Errorf("%s has no %q remote", path, git.DefaultRemoteName)
	}
	if slices.Equal(urls, []string{url}) {
		return nil
	}
	if _, err := localConfig(path, "--replace-all", "--", key, url); err != nil {
		return fmt.Errorf("update remote %s: %w", path, err)
	}
	urls, err = localValues(path, key)
	if err != nil {
		return fmt.Errorf("update remote %s: read back: %w", path, err)
	}
	if !slices.Equal(urls, []string{url}) {
		return fmt.Errorf("update remote %s: %s is %q after the update, want %q", path, key, urls, url)
	}
	return nil
}

// localValues returns every raw value of key in the repository's own config
// file, without includes or insteadOf rewriting; a missing key yields none.
func localValues(path, key string) ([]string, error) {
	out, err := localConfig(path, "--null", "--get-all", "--", key)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return nil, nil // no matching configuration entry
	}
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSuffix(out, "\x00"), "\x00"), nil
}

// localConfig runs `git config --local` on the repository at path. The
// explicit --git-dir disables discovery, so a path without its own .git never
// reaches a parent repository; variables that a Git hook exports and that
// would redirect the command to another repository or file are dropped.
func localConfig(path string, args ...string) (string, error) {
	return command("config", append([]string{"--git-dir", filepath.Join(path, ".git"), "config", "--local"}, args...)...)
}

// command runs Git with argv and returns its standard output; name labels its
// errors. Variables that a Git hook exports and that would redirect the
// command to another repository or file are dropped.
func command(name string, argv ...string) (string, error) {
	cmd := exec.Command("git", argv...) //nolint:noctx // the GitSync port takes no context; these are short local calls
	for _, env := range os.Environ() {
		key, _, _ := strings.Cut(env, "=")
		switch key {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_CONFIG":
			continue
		}
		cmd.Env = append(cmd.Env, env)
	}
	cmd.Env = append(cmd.Env, "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %w: %s", name, err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", name, err)
	}
	return string(out), nil
}

// linkedWorktrees lists the linked worktrees of the checkout at path the way
// Git records them: a gitdir file per worktree under .git/worktrees names the
// worktree's own .git file. A checkout that is itself a linked worktree, or
// has none, yields none.
func linkedWorktrees(path string) ([]string, error) {
	records, _, err := worktreeRecords(path)
	worktrees := make([]string, 0, len(records))
	for _, record := range records {
		worktrees = append(worktrees, filepath.Dir(record.gitfile))
	}
	return worktrees, err
}

// worktreeRecord is the administrative directory of one linked worktree and
// the worktree's .git file it links back to.
type worktreeRecord struct {
	dir     string // <checkout>/.git/worktrees/<id>
	gitfile string // <worktree>/.git
}

// worktreeRecords reads the worktree records of the checkout at path. A record
// whose gitdir file cannot be read names no worktree: it comes back as a
// *WorktreeError for that record, so a move reports the worktree it cannot
// find instead of passing over it.
func worktreeRecords(path string) ([]worktreeRecord, []error, error) {
	dir := filepath.Join(path, git.GitDirName, "worktrees")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var (
		records    []worktreeRecord
		unreadable []error
	)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		record := worktreeRecord{dir: filepath.Join(dir, entry.Name())}
		raw, err := os.ReadFile(filepath.Join(record.dir, "gitdir"))
		if err == nil && strings.TrimSpace(string(raw)) == "" {
			err = errors.New("it is empty")
		}
		if err != nil {
			unreadable = append(unreadable, &WorktreeError{Path: record.dir, Unknown: true, Err: err})
			continue
		}
		record.gitfile = strings.TrimSpace(string(raw))
		if !filepath.IsAbs(record.gitfile) {
			record.gitfile = filepath.Join(record.dir, record.gitfile)
		}
		record.gitfile = filepath.Clean(record.gitfile)
		records = append(records, record)
	}
	slices.SortFunc(records, func(a, b worktreeRecord) int { return strings.Compare(a.gitfile, b.gitfile) })
	return records, unreadable, nil
}

// RepairWorktrees reconnects the linked worktrees of a checkout moved from
// one path to another, as `git worktree repair` run in the moved checkout
// does: each worktree's .git file is pointed at the moved checkout, and a
// worktree inside the checkout, which moved with it, is recorded at its new
// path. Each worktree is repaired on its own, so that one Git cannot write to
// does not leave the others broken, and is then opened by Git: one that is
// still unusable, its directory gone or its .git file not writable, is a
// *WorktreeError.
func (s *Sync) RepairWorktrees(from, to string) []error {
	records, errs, err := worktreeRecords(to)
	if err != nil {
		return []error{fmt.Errorf("list the linked worktrees of %s: %w", to, err)}
	}
	if len(records) == 0 {
		return errs
	}
	// Git records worktrees by their real path; the old path no longer exists,
	// so only its parent can be resolved.
	sources := []string{from}
	if parent, err := filepath.EvalSymlinks(filepath.Dir(from)); err == nil {
		sources = append(sources, filepath.Join(parent, filepath.Base(from)))
	}
	for _, record := range records {
		gitfile := record.gitfile
		for _, source := range sources {
			if rel, err := filepath.Rel(source, gitfile); err == nil && rel != ".." &&
				!strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				gitfile = filepath.Join(to, rel)
				break
			}
		}
		worktree := filepath.Dir(gitfile)
		if err := repairWorktree(record, gitfile, sources); err != nil {
			errs = append(errs, &WorktreeError{Path: worktree, Err: err})
			continue
		}
		if _, err := command("rev-parse", "-C", worktree, "rev-parse", "--git-dir"); err != nil {
			errs = append(errs, &WorktreeError{Path: worktree, Err: err})
		}
	}
	return errs
}

// repairWorktree writes the two links of one linked worktree in the format of
// gitrepository-layout(5): the worktree's .git file names its administrative
// directory, and the directory's gitdir file names the .git file back. It
// writes nothing unless the worktree's .git is a regular file that links to
// this administrative directory, at its old location under one of sources or
// already at the new one: a path reused by another repository's worktree, or
// a .git that Git did not write for this one, is never relinked.
func repairWorktree(record worktreeRecord, gitfile string, sources []string) error {
	if info, err := os.Stat(filepath.Dir(gitfile)); err != nil || !info.IsDir() {
		return fmt.Errorf("its directory is gone")
	}
	if dir, err := filepath.EvalSymlinks(filepath.Dir(gitfile)); err == nil {
		gitfile = filepath.Join(dir, filepath.Base(gitfile)) // Git records real paths
	}
	if info, err := os.Lstat(gitfile); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("its .git is not the link file Git writes")
	}
	raw, err := os.ReadFile(gitfile)
	if err != nil {
		return err
	}
	linked, ok := strings.CutPrefix(strings.TrimSpace(string(raw)), "gitdir: ")
	if !ok {
		return fmt.Errorf("its .git is not the link file Git writes")
	}
	if !filepath.IsAbs(linked) {
		linked = filepath.Join(filepath.Dir(gitfile), linked)
	}
	linked = filepath.Clean(linked)
	admin, err := filepath.EvalSymlinks(record.dir)
	if err != nil {
		return err
	}
	current, err := filepath.EvalSymlinks(linked)
	relinked := err == nil && current == admin
	if !relinked {
		owned := false
		for _, source := range sources {
			owned = owned || linked == filepath.Join(source, git.GitDirName, "worktrees", filepath.Base(record.dir))
		}
		if !owned {
			return fmt.Errorf("its .git links to another repository: %s", linked)
		}
	}
	if gitfile != record.gitfile {
		if err := os.WriteFile(filepath.Join(record.dir, "gitdir"), []byte(gitfile+"\n"), 0o644); err != nil {
			return err
		}
	}
	if relinked {
		return nil // already linked to the moved checkout
	}
	return os.WriteFile(gitfile, []byte("gitdir: "+admin+"\n"), 0o644)
}

// WorktreeError names a linked worktree that a move could not bring back.
type WorktreeError struct {
	Path    string
	Unknown bool // Path is the worktree's record, which does not say where it is
	Err     error
}

func (e *WorktreeError) Error() string {
	// Git reports the links it repaired on the same stream; the first line
	// says what failed.
	reason, _, _ := strings.Cut(e.Err.Error(), "\n")
	if e.Unknown {
		return fmt.Sprintf("the linked worktree recorded at %s cannot be found to bring it back after the move: %s", e.Path, reason)
	}
	return fmt.Sprintf("linked worktree %s is not usable after the move: %s", e.Path, reason)
}

func (e *WorktreeError) Unwrap() error { return e.Err }

// Inspect reads origin fetch/push URLs separately, the short HEAD sha and the
// linked worktrees. A checkout that is itself a linked worktree is read with
// the configuration of its main repository, as `git config --local` reads it.
// An incomplete clone is reported as one even when it stopped too early to
// open as a repository; then nothing else about it is known.
func (s *Sync) Inspect(path string) (CloneInfo, error) {
	_, err := os.Lstat(filepath.Join(path, git.GitDirName, IncompleteMarker))
	incomplete := err == nil
	info, err := s.inspect(path)
	if incomplete {
		info.Incomplete = true
		return info, nil
	}
	return info, err
}

func (s *Sync) inspect(path string) (CloneInfo, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{EnableDotGitCommonDir: true})
	if err != nil {
		return CloneInfo{}, fmt.Errorf("open %s: %w", path, err)
	}
	var info CloneInfo
	// The list only informs the plan; a move reads it again and reports
	// what it cannot repair, so an unreadable list never hides the checkout.
	info.Worktrees, _ = linkedWorktrees(path)
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
