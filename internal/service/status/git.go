// Package status provides a read-only snapshot of local Git checkouts.
package status

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	giturls "github.com/whilp/git-urls"
)

// Row is shared by the JSON, printable and interactive views. Counts describe
// local refs only; no network operation is performed.
type Row struct {
	ID              int64      `json:"id,omitempty"`
	Repository      string     `json:"repository"`
	Path            string     `json:"path"`
	Branch          string     `json:"branch"`
	DefaultBranch   string     `json:"default_branch,omitempty"`
	Commit          string     `json:"commit,omitempty"`
	Upstream        string     `json:"upstream,omitempty"`
	Added           int        `json:"added"`
	Deleted         int        `json:"deleted"`
	Changed         int        `json:"changed_files"`
	Untracked       int        `json:"untracked"`
	Binary          int        `json:"binary"`
	Conflicts       int        `json:"conflicts"`
	Ahead           int        `json:"ahead"`
	Behind          int        `json:"behind"`
	Status          string     `json:"status"`
	Pinned          bool       `json:"pinned"`
	OrphanReason    string     `json:"orphan_reason,omitempty"`
	ActivePath      string     `json:"active_path,omitempty"`
	RemoteCheckedAt *time.Time `json:"remote_checked_at,omitempty"`
	Error           string     `json:"error,omitempty"`
	displayPath     string     // relative to workspace root, or home for external paths
}

func git(ctx context.Context, path string, args ...string) (string, error) {
	argv := append([]string{"--no-optional-locks", "-C", path, "-c", "core.fsmonitor=false", "-c", "core.quotePath=false"}, args...)
	cmd := exec.CommandContext(ctx, "git", argv...)
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
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return string(out), nil
}

func remoteName(raw string) string {
	u, err := giturls.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
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
	fail := func(err error) Row { row.Status, row.Error = "error", err.Error(); return row }
	if _, err := os.Stat(filepath.Join(row.Path, ".git")); err != nil {
		return fail(err)
	}
	if raw, err := git(ctx, row.Path, "remote", "get-url", "origin"); err == nil {
		if name := remoteName(raw); name != "" {
			row.Repository = name
		}
	}
	if raw, err := git(ctx, row.Path, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD"); err == nil {
		row.DefaultBranch = strings.TrimPrefix(strings.TrimSpace(raw), "refs/remotes/origin/")
	}
	raw, err := git(ctx, row.Path, "status", "--porcelain=v2", "--branch", "-z", "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return fail(err)
	}
	entries := strings.Split(raw, "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		switch {
		case strings.HasPrefix(entry, "# branch.oid "):
			row.Commit = strings.TrimPrefix(entry, "# branch.oid ")
		case strings.HasPrefix(entry, "# branch.head "):
			row.Branch = strings.TrimPrefix(entry, "# branch.head ")
		case strings.HasPrefix(entry, "? "):
			row.Untracked++
		case strings.HasPrefix(entry, "u "):
			row.Conflicts++
			row.Changed++
		case strings.HasPrefix(entry, "1 "), strings.HasPrefix(entry, "2 "):
			row.Changed++
			if entry[0] == '2' {
				i++
			} // rename's original pathname is another NUL field
		}
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
			return fail(err)
		}
		row.Upstream = strings.TrimSpace(upstream)
		if row.Upstream != "" {
			if _, err := git(ctx, row.Path, "rev-parse", "--verify", "@{upstream}"); err != nil {
				row.Status = "upstream gone"
			} else {
				counts, err := git(ctx, row.Path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
				if err != nil {
					return fail(err)
				}
				if _, err := fmt.Sscan(counts, &row.Ahead, &row.Behind); err != nil {
					return fail(err)
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
	args := []string{"diff", "--no-ext-diff", "--no-textconv", "--no-renames", "--numstat", "-z", "--ignore-submodules=none"}
	if unborn {
		args = append(args, "--cached")
	} else {
		args = append(args, "HEAD")
	}
	args = append(args, "--")
	stats, err := git(ctx, row.Path, args...)
	if err != nil {
		return fail(err)
	}
	for _, entry := range strings.Split(stats, "\x00") {
		if entry == "" {
			continue
		}
		fields := strings.SplitN(entry, "\t", 3)
		if len(fields) != 3 {
			return fail(fmt.Errorf("invalid Git numstat"))
		}
		if fields[0] == "-" {
			row.Binary++
			continue
		}
		add, e1 := strconv.Atoi(fields[0])
		del, e2 := strconv.Atoi(fields[1])
		if e1 != nil || e2 != nil {
			return fail(fmt.Errorf("invalid Git line counts"))
		}
		row.Added += add
		row.Deleted += del
	}
	return row
}
