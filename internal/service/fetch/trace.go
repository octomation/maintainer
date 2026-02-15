package fetch

import (
	"context"
	"time"

	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
)

// tracedSync writes a trace breadcrumb for each Git operation of a run: the
// checkouts the scan inspects and every clone, fetch, move and remote update,
// with its duration and error (-vvv).
type tracedSync struct {
	gitsvc.GitSync
	reporter *Reporter
}

func (t tracedSync) trace(op string, start time.Time, err error, args ...any) {
	args = append(args, "took", time.Since(start).Round(time.Millisecond))
	if err != nil {
		args = append(args, "error", err)
	}
	t.reporter.Tracef("git "+op, args...)
}

func (t tracedSync) Clone(ctx context.Context, opt gitsvc.CloneOptions) error {
	start := time.Now()
	err := t.GitSync.Clone(ctx, opt)
	t.trace("clone", start, err, "url", opt.URL, "path", opt.Path)
	return err
}

func (t tracedSync) Fetch(ctx context.Context, path string, auth gitsvc.Auth) error {
	start := time.Now()
	err := t.GitSync.Fetch(ctx, path, auth)
	t.trace("fetch", start, err, "path", path, "transport", auth.Transport)
	return err
}

func (t tracedSync) Move(from, to string) error {
	start := time.Now()
	err := t.GitSync.Move(from, to)
	t.trace("move", start, err, "from", from, "to", to)
	return err
}

func (t tracedSync) UpdateRemote(path, url string) error {
	start := time.Now()
	err := t.GitSync.UpdateRemote(path, url)
	t.trace("update-remote", start, err, "path", path, "url", url)
	return err
}

func (t tracedSync) RepairWorktrees(from, to string) []error {
	start := time.Now()
	errs := t.GitSync.RepairWorktrees(from, to)
	t.trace("repair-worktrees", start, nil, "from", from, "to", to, "errors", len(errs))
	return errs
}

func (t tracedSync) Inspect(path string) (gitsvc.CloneInfo, error) {
	start := time.Now()
	info, err := t.GitSync.Inspect(path)
	t.trace("inspect", start, err, "path", path)
	return info, err
}
