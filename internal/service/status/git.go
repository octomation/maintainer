// Package status provides a read-only snapshot of local Git checkouts.
package status

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/giturl"
)

// Row is shared by the JSON, printable and interactive views. Counts describe
// local refs only; no network operation is performed. Added, Deleted and
// Binary are nil until the line count of the checkout completes, and Changed,
// Untracked and Conflicts until its git status does: unknown, never zero.
type Row struct {
	ID            int64  `json:"id,omitempty"`
	Repository    string `json:"repository"`
	Path          string `json:"path"`
	Branch        string `json:"branch"`
	DefaultBranch string `json:"default_branch,omitempty"`
	Commit        string `json:"commit,omitempty"`
	Upstream      string `json:"upstream,omitempty"`
	Added         *int   `json:"added"`
	Deleted       *int   `json:"deleted"`
	Changed       *int   `json:"changed_files"`
	Untracked     *int   `json:"untracked"`
	Binary        *int   `json:"binary"`
	Conflicts     *int   `json:"conflicts"`
	Ahead         int    `json:"ahead"`
	Behind        int    `json:"behind"`
	Status        string `json:"status"`
	PushLocked    bool   `json:"push_locked"`
	Pinned        bool   `json:"pinned"`
	// Visibility is the repository's visibility on GitHub as the fetch state
	// recorded it (public, private, internal); empty when no record knows it.
	Visibility      string     `json:"visibility,omitempty"`
	OrphanReason    string     `json:"orphan_reason,omitempty"`
	ActivePath      string     `json:"active_path,omitempty"`
	RemoteCheckedAt *time.Time `json:"remote_checked_at,omitempty"`
	Error           string     `json:"error,omitempty"`
	displayPath     string     // relative to workspace root, or home for external paths
	placeholder     bool       // a scan obstruction or an unresolved pin: no checkout to inspect
}

func git(ctx context.Context, path string, args ...string) (string, error) {
	argv := append([]string{"--no-optional-locks", "-C", path, "-c", "core.fsmonitor=false", "-c", "core.quotePath=false"}, args...)
	cmd := exec.CommandContext(ctx, "git", argv...)
	// A child of Git, such as a filter, may hold the output open after Git is
	// killed: stop waiting for it soon after.
	cmd.WaitDelay = time.Second
	// A maintainer launched from a hook must still inspect the requested repo.
	for _, env := range os.Environ() {
		key, _, _ := strings.Cut(env, "=")
		switch key {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_OPTIONAL_LOCKS", "GIT_NO_LAZY_FETCH":
			continue
		}
		cmd.Env = append(cmd.Env, env)
	}
	cmd.Env = append(cmd.Env, "GIT_OPTIONAL_LOCKS=0", "GIT_NO_LAZY_FETCH=1", "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		// The exit status alone cannot tell a broken gitfile from dubious
		// ownership or a corrupt index; Git explains the failure on stderr.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if msg := oneLine(exitErr.Stderr); msg != "" {
				return "", fmt.Errorf("git %s: %s (%w)", args[0], msg, err)
			}
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return string(out), nil
}

// oneLine joins the non-empty lines of a Git diagnostic for a table cell.
func oneLine(raw []byte) string {
	var lines []string
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, " ")
}

// RemoteName returns owner/name of a remote URL. A local repository, even a
// file:// URL with a host, names no repository.
func RemoteName(raw string) string {
	u, err := giturl.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || u.Scheme == "file" {
		return ""
	}
	name := strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
	if !strings.Contains(name, "/") {
		return ""
	}
	return name
}

// Inspect reads a single checkout. Metadata remains available on error.
func Inspect(ctx context.Context, row Row) Row {
	// Values derived from Git are recomputed rather than accumulated when a row
	// is inspected again; metadata supplied by the caller is kept.
	row.Branch, row.Commit, row.Upstream, row.Status, row.Error, row.PushLocked = "", "", "", "", "", false
	row.Added, row.Deleted, row.Binary, row.Changed, row.Untracked, row.Conflicts = nil, nil, nil, nil, nil, nil
	row.Ahead, row.Behind = 0, 0
	if _, err := os.Stat(filepath.Join(row.Path, ".git")); err != nil {
		row.Status, row.Error = "error", err.Error()
		return row
	}
	if err := inspect(ctx, &row); err != nil {
		row.Status, row.Error = "error", err.Error()
	}
	// The lock probe runs after git status and independently of it, so the
	// first failure is the one reported and a lock survives a later failure.
	locked, err := pushLocked(ctx, row.Path)
	switch {
	case err == nil:
		row.PushLocked = locked
	case row.Error == "":
		row.Status, row.Error = "error", err.Error()
	}
	row.forgetCounts()
	return row
}

// forgetCounts drops the file and line counts of an error row: they are
// unknown, as its cell shows them, whichever rule made it an error.
func (row *Row) forgetCounts() {
	if row.Error != "" {
		row.Added, row.Deleted, row.Binary, row.Changed, row.Untracked, row.Conflicts = nil, nil, nil, nil, nil, nil
	}
}

func inspect(ctx context.Context, row *Row) error {
	// Renames are detected explicitly, as in the line count below, so that
	// status.renames and diff.renames cannot change a row. Rename detection
	// in both shares one bound; past it, git status runs again without it, a
	// rename counts as a deleted and an added file, and the lines stay unknown.
	// That run has a bound of its own: past it, the branch and the commit come
	// from HEAD, which walks no tree, and the file counts stay unknown too.
	start := time.Now()
	renames, cancel := context.WithTimeout(ctx, renamesBound)
	defer cancel()
	statusArgs := []string{"status", "--porcelain=v2", "--branch", "-z", "--untracked-files=all", "--ignore-submodules=none"}
	raw, statusErr := git(renames, row.Path, append(statusArgs, "--find-renames")...)
	spent := time.Since(start)
	counted := true
	if statusErr != nil && ctx.Err() == nil && renames.Err() != nil {
		plain, cancel := context.WithTimeout(ctx, statusBound)
		defer cancel()
		raw, statusErr = git(plain, row.Path, append(statusArgs, "--no-renames")...)
		if statusErr != nil && ctx.Err() == nil && plain.Err() != nil {
			counted, statusErr = false, readHead(ctx, row)
		}
	}
	// Identity is best effort and remains available when git status fails.
	if origin, err := git(ctx, row.Path, "remote", "get-url", "origin"); err == nil {
		if name := RemoteName(origin); name != "" {
			row.Repository = name
		}
	}
	if head, err := git(ctx, row.Path, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD"); err == nil {
		row.DefaultBranch = strings.TrimPrefix(strings.TrimSpace(head), "refs/remotes/origin/")
	}
	if statusErr != nil {
		return statusErr
	}
	if counted {
		countStatus(row, raw)
	}
	row.Status = "no upstream"
	unborn := row.Commit == "(initial)"
	switch {
	case unborn:
		row.Commit, row.Status = "", "unborn"
	case row.Branch == "(detached)":
		row.Branch, row.Status = "detached", "detached"
	default:
		upstream, err := git(ctx, row.Path, "for-each-ref", "--format=%(upstream:short)", "refs/heads/"+row.Branch)
		if err != nil {
			return err
		}
		row.Upstream = strings.TrimSpace(upstream)
		if row.Upstream != "" {
			if _, err := git(ctx, row.Path, "rev-parse", "--verify", "@{upstream}"); err != nil {
				row.Status = "upstream gone"
			} else {
				counts, err := git(ctx, row.Path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
				if err != nil {
					return err
				}
				if _, err := fmt.Sscan(counts, &row.Ahead, &row.Behind); err != nil {
					return fmt.Errorf("git rev-list: %w", err)
				}
				row.Status = "synced"
				switch {
				case row.Ahead > 0 && row.Behind > 0:
					row.Status = "diverged"
				case row.Ahead > 0:
					row.Status = "ahead"
				case row.Behind > 0:
					row.Status = "behind"
				}
			}
		}
	}
	if !counted {
		return nil // past the bound of git status the counts stay unknown
	}
	args := []string{"diff", "--no-ext-diff", "--no-textconv", "--find-renames", "--numstat", "-z", "--ignore-submodules=none"}
	if unborn {
		args = append(args, "--cached")
	} else {
		args = append(args, "HEAD")
	}
	args = append(args, "--")
	// The line count gets what git status left of the bound: nothing once
	// git status passed it.
	lines, cancel := context.WithTimeout(ctx, renamesBound-spent)
	defer cancel()
	stats, err := git(lines, row.Path, args...)
	if err != nil {
		if ctx.Err() == nil && lines.Err() != nil {
			return nil // past the bound the line counts stay unknown
		}
		return err
	}
	if err := countLines(row, stats); err != nil {
		return fmt.Errorf("git diff: %w", err)
	}
	return nil
}

// readHead sets the branch and the commit of row as git status --branch names
// them, (detached) and (initial) included, from HEAD alone: it walks no tree.
func readHead(ctx context.Context, row *Row) error {
	ref, err := git(ctx, row.Path, "symbolic-ref", "--quiet", "HEAD")
	switch {
	case err == nil:
		row.Branch = strings.TrimPrefix(strings.TrimSpace(ref), "refs/heads/")
	case exitCode(err) == 1:
		row.Branch = "(detached)"
	default:
		return err
	}
	oid, err := git(ctx, row.Path, "rev-parse", "--verify", "--quiet", "HEAD")
	switch {
	case err == nil:
		row.Commit = strings.TrimSpace(oid)
	case exitCode(err) == 1:
		row.Commit = "(initial)" // an unborn branch
	default:
		return err
	}
	return nil
}

// exitCode is the exit status of a Git command that ran and failed, -1 otherwise.
func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

// countStatus sets the file counts of row, its branch and its commit from the
// entries of git status --porcelain=v2 --branch -z.
func countStatus(row *Row, raw string) {
	var changed, untracked, conflicts int
	entries := strings.Split(raw, "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		switch {
		case strings.HasPrefix(entry, "# branch.oid "):
			row.Commit = strings.TrimPrefix(entry, "# branch.oid ")
		case strings.HasPrefix(entry, "# branch.head "):
			row.Branch = strings.TrimPrefix(entry, "# branch.head ")
		case strings.HasPrefix(entry, "? "):
			untracked++
		case strings.HasPrefix(entry, "u "):
			conflicts++
			changed++
		case strings.HasPrefix(entry, "1 "), strings.HasPrefix(entry, "2 "):
			changed++
			if entry[0] == '2' {
				i++
			} // rename's original pathname is another NUL field
		}
	}
	row.Changed, row.Untracked, row.Conflicts = &changed, &untracked, &conflicts
}

// renamesBound limits rename detection in one checkout, in git status and the
// line count together. Over a large change it can take minutes, and one
// checkout must not hold the snapshot: past the bound its row keeps the rest
// and its line counts stay unknown.
var renamesBound = 10 * time.Second

// statusBound limits git status without rename detection in one checkout. In
// a very large tree it takes minutes too: past the bound the row keeps what
// HEAD, the refs and the configuration give, and its file counts stay unknown.
var statusBound = 10 * time.Second

// countLines sets the line and binary counts of row from the records of git
// diff --numstat -z. A rename leaves the path field empty and names the
// original and new paths in the next two NUL fields; it counts as one record.
func countLines(row *Row, raw string) error {
	var added, deleted, binary int
	records := strings.Split(raw, "\x00")
	for i := 0; i < len(records); i++ {
		if records[i] == "" {
			continue
		}
		fields := strings.SplitN(records[i], "\t", 3)
		if len(fields) != 3 {
			return fmt.Errorf("invalid Git numstat")
		}
		if fields[2] == "" {
			if i+2 >= len(records) || records[i+1] == "" || records[i+2] == "" {
				return fmt.Errorf("invalid Git numstat")
			}
			i += 2
		}
		if fields[0] == "-" {
			binary++
			continue
		}
		add, e1 := strconv.Atoi(fields[0])
		del, e2 := strconv.Atoi(fields[1])
		if e1 != nil || e2 != nil {
			return fmt.Errorf("invalid Git line counts")
		}
		added += add
		deleted += del
	}
	row.Added, row.Deleted, row.Binary = &added, &deleted, &binary
	return nil
}

// pushLocked recognizes the marker written by git lock [remote]. Read through
// Git so includes and worktree configuration use the same rules as git push.
func pushLocked(ctx context.Context, path string) (bool, error) {
	_, err := git(ctx, path, "config", "--get-regexp", `^remote\..*\.pushurl$`, "^no_push$")
	if err == nil {
		return true, nil
	}
	if exitCode(err) == 1 {
		return false, nil // no matching configuration entry
	}
	return false, err
}
