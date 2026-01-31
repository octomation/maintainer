// Package workspace owns the common local discovery boundary for fetch/status.
package workspace

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.octolab.org/toolset/maintainer/internal/config"
	"go.octolab.org/toolset/maintainer/internal/pkg/pathpattern"
)

type Scope struct {
	Root     string
	pins     []string
	patterns []*pathpattern.Pattern
}

func New(cnf *config.Fetch, root, home string) (*Scope, error) {
	s := &Scope{Root: filepath.Clean(root)}
	w := cnf.WorkspaceConfig()
	p, err := pathpattern.Compile(w.Path, "")
	if err != nil {
		return nil, err
	}
	s.patterns = append(s.patterns, p)
	for _, o := range cnf.Owners {
		if o.Path == "" {
			continue
		}
		p, err := pathpattern.Compile(o.Path, o.Name)
		if err != nil {
			return nil, err
		}
		s.patterns = append(s.patterns, p)
	}
	for _, raw := range w.Pins {
		if strings.TrimSpace(raw) == "" || strings.ContainsAny(raw, "\x00\r\n") || strings.Contains(raw, "{{") {
			return nil, fmt.Errorf("workspace.pins must contain nonempty literal paths")
		}
		path := raw
		switch {
		case raw == "~" || strings.HasPrefix(raw, "~/"):
			if home == "" {
				return nil, fmt.Errorf("cannot expand pin %q without home", raw)
			}
			path = filepath.Join(home, strings.TrimPrefix(raw, "~"))
		case strings.HasPrefix(raw, "~"):
			return nil, fmt.Errorf("unsupported pin path %q", raw)
		case !filepath.IsAbs(raw):
			path = filepath.Join(root, raw)
		}
		s.pins = append(s.pins, filepath.Clean(path))
	}
	return s, nil
}

func Within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func (s *Scope) Pinned(path string) bool {
	for _, pin := range s.pins {
		if Within(pin, path) {
			return true
		}
	}
	return false
}

func (s *Scope) Managed(path string) bool {
	if !Within(s.Root, path) {
		return false
	}
	rel, _ := filepath.Rel(s.Root, path)
	for _, p := range s.patterns {
		if p.Match(rel) {
			return true
		}
	}
	return false
}

func (s *Scope) Contains(path string) bool { return s.Pinned(path) || s.Managed(path) }

func (s *Scope) descend(path string) bool {
	for _, pin := range s.pins {
		if Within(pin, path) || Within(path, pin) {
			return true
		}
	}
	if !Within(s.Root, path) {
		return false
	}
	rel, _ := filepath.Rel(s.Root, path)
	for _, p := range s.patterns {
		if p.Prefix(rel) {
			return true
		}
	}
	return false
}

// ErrUnmanagedCheckout names a Git repository at a prefix level of the
// layout: it is neither listed nor entered, and the walk continues below it.
var ErrUnmanagedCheckout = errors.New("a Git repository at an unmanaged level of the layout")

// Walk never inspects Git or traverses directories outside the configured
// boundary. add also receives missing pin roots so callers report them visibly.
// obstruct receives every path inside the boundary that the walk cannot scan,
// an unreadable directory or a Git repository at a prefix level of the layout,
// with the reason; the walk keeps every reachable result and continues.
func (s *Scope) Walk(ctx context.Context, add func(path string, pinned bool) error, obstruct func(path string, reason error) error) error {
	seen := map[string]bool{}
	obstructed := map[string]bool{}
	report := func(path string, reason error) error {
		if obstructed[path] {
			return nil
		}
		obstructed[path] = true
		return obstruct(path, reason)
	}
	roots := append([]string{s.Root}, s.pins...)
	for i, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if walkErr != nil {
				if os.IsNotExist(walkErr) && path == root {
					if i > 0 && !seen[path] {
						seen[path] = true
						return add(path, true)
					}
					return nil
				}
				// A directory that cannot be read hides only its own subtree.
				if err := report(path, walkErr); err != nil {
					return err
				}
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if seen[path] {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.IsDir() {
				if path == root {
					return fmt.Errorf("workspace path %s is not a directory (symlink roots are not followed)", path)
				}
				return nil
			}
			if path != root && d.Name() == ".git" {
				return filepath.SkipDir // Git metadata of a repository at a prefix level
			}
			if !s.descend(path) {
				return filepath.SkipDir
			}
			seen[path] = true
			if _, err := os.Lstat(filepath.Join(path, ".git")); err == nil {
				if s.Contains(path) {
					if err := add(path, s.Pinned(path)); err != nil {
						return err
					}
					return filepath.SkipDir
				}
				return report(path, ErrUnmanagedCheckout)
			} else if !os.IsNotExist(err) {
				if err := report(path, err); err != nil {
					return err
				}
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("scan workspace %s: %w", root, err)
		}
	}
	return nil
}

// Destination errors name an unsafe parent of a clone or move destination.
var (
	ErrSymlinkParent = errors.New("a parent directory inside the workspace root is a symbolic link")
	ErrFileParent    = errors.New("a parent path inside the workspace root is not a directory")
)

// Destination checks a clone or move destination right before it is created.
// The path must lie strictly inside root, and every existing component between
// them must be a real directory, so creating the missing ones never writes
// through a symbolic link. An existing destination is reported as fs.ErrExist.
func Destination(root, path string) error {
	root, path = filepath.Clean(root), filepath.Clean(path)
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || !Within(root, path) {
		return fmt.Errorf("destination %s is outside the workspace root %s", path, root)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	current := root
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		switch {
		case os.IsNotExist(err):
			return nil
		case err != nil:
			return err
		case i == len(parts)-1:
			return fmt.Errorf("destination %s: %w", current, fs.ErrExist)
		case info.Mode()&fs.ModeSymlink != 0:
			return fmt.Errorf("%w: %s", ErrSymlinkParent, current)
		case !info.IsDir():
			return fmt.Errorf("%w: %s", ErrFileParent, current)
		}
	}
	return nil
}
