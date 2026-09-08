package status

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/sync/errgroup"

	"go.octolab.org/toolset/maintainer/internal/config"
	fetchsvc "go.octolab.org/toolset/maintainer/internal/service/fetch"
	"go.octolab.org/toolset/maintainer/internal/service/github"
	"go.octolab.org/toolset/maintainer/internal/state"
)

// Collect combines disk discovery, remembered checkouts and explicit pins.
// State is only a snapshot loaded by the caller and is never mutated.
func Collect(ctx context.Context, cnf *config.Fetch, st *state.State, home, cwd string, owners []string, concurrency int) ([]Row, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("concurrency must be positive")
	}
	renderer, err := fetchsvc.NewPathRenderer(cnf.Defaults.Root, home, cwd)
	if err != nil {
		return nil, err
	}
	paths := fetchsvc.NewPathResolver(cnf, renderer)
	byPath := map[string]Row{}
	byName := map[string]state.Record{}
	pins := map[string]string{}
	pinnedPaths := map[string]bool{}
	add := func(path string, rec state.Record) {
		path = filepath.Clean(path)
		if _, exists := byPath[path]; exists && rec.ID == 0 {
			return
		}
		name := filepath.Base(path)
		if rec.OwnerLogin != "" {
			name = rec.OwnerLogin + "/" + rec.Name
		}
		byPath[path] = Row{ID: rec.ID, Repository: name, Path: path, DefaultBranch: rec.DefaultBranch}
	}
	for _, rec := range st.Repos {
		byName[strings.ToLower(rec.OwnerLogin+"/"+rec.Name)] = rec
		if cnf.Ignored(rec.ID, rec.OwnerLogin, rec.Name) {
			continue
		}
		snap := snapshot(rec)
		pin, err := paths.PinnedPath(snap, &rec)
		if err != nil {
			return nil, err
		}
		if pin != "" {
			pinnedPaths[pin] = true
			pins[strings.ToLower(rec.OwnerLogin+"/"+rec.Name)] = pin
			add(pin, rec)
		} else if rec.Path != "" {
			add(rec.Path, rec)
		}
	}
	// Rules without state still support literal paths and owner/name templates.
	for _, rule := range cnf.Repos {
		if rule.Ignore || rule.Path == "" {
			continue
		}
		rec := state.Record{ID: rule.Match.ID, OwnerLogin: rule.Match.Owner, Name: rule.Match.Name}
		if known, ok := st.ByID(rec.ID); rec.ID != 0 && ok {
			rec = *known
		}
		pin, err := paths.PinnedPath(snapshot(rec), &rec)
		if err != nil {
			return nil, err
		}
		add(pin, rec)
		pinnedPaths[pin] = true
		if rec.OwnerLogin != "" {
			pins[strings.ToLower(rec.OwnerLogin+"/"+rec.Name)] = pin
		}
	}
	err = filepath.WalkDir(renderer.Root(), func(path string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			add(path, state.Record{})
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan %s: %w", renderer.Root(), err)
	}
	rows := make([]Row, 0, len(byPath))
	for _, row := range byPath {
		row.Pinned = pinnedPaths[row.Path]
		rows = append(rows, row)
	}
	var group errgroup.Group
	group.SetLimit(concurrency)
	for i := range rows {
		group.Go(func() error {
			expected := rows[i].Repository
			rows[i] = Inspect(ctx, rows[i])
			if rows[i].ID != 0 && strings.Contains(expected, "/") && !strings.EqualFold(expected, rows[i].Repository) {
				rows[i].Error = fmt.Sprintf("origin is %s; expected %s from state (run fetch to verify identity)", rows[i].Repository, expected)
				rows[i].Repository, rows[i].Status = expected, "error"
			}
			return nil
		})
	}
	_ = group.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// A literal ID-only pin can learn its display identity from its origin.
	for _, row := range rows {
		if row.Pinned {
			pins[strings.ToLower(row.Repository)] = row.Path
		}
	}
	selected := make([]Row, 0, len(rows))
	for _, row := range rows {
		key := strings.ToLower(row.Repository)
		if rec, ok := byName[key]; ok && row.ID == 0 {
			row.ID = rec.ID
			if row.DefaultBranch == "" {
				row.DefaultBranch = rec.DefaultBranch
			}
		}
		owner, name, _ := strings.Cut(row.Repository, "/")
		if cnf.Ignored(row.ID, owner, name) {
			continue
		}
		if pin, ok := pins[key]; ok {
			if row.Path != pin {
				continue
			}
			row.Pinned = true
		}
		if len(owners) > 0 {
			match := false
			for _, filter := range owners {
				match = match || strings.EqualFold(owner, filter)
			}
			if !match {
				continue
			}
		}
		selected = append(selected, row)
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].Repository != selected[j].Repository {
			return selected[i].Repository < selected[j].Repository
		}
		return selected[i].Path < selected[j].Path
	})
	return selected, nil
}

func snapshot(rec state.Record) github.RepoSnapshot {
	return github.RepoSnapshot{ID: rec.ID, Owner: rec.OwnerLogin, Name: rec.Name,
		Visibility: github.Visibility(rec.Visibility), DefaultBranch: rec.DefaultBranch,
		IsFork: rec.IsFork, IsTemplate: rec.IsTemplate, IsArchived: rec.ArchivedOnGitHub}
}
