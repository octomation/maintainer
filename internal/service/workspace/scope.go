// Package workspace owns the common local discovery boundary for fetch/status.
package workspace

import (
	"context"
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
		if raw == "~" || strings.HasPrefix(raw, "~/") {
			if home == "" {
				return nil, fmt.Errorf("cannot expand pin %q without home", raw)
			}
			path = filepath.Join(home, strings.TrimPrefix(raw, "~"))
		} else if strings.HasPrefix(raw, "~") {
			return nil, fmt.Errorf("unsupported pin path %q", raw)
		} else if !filepath.IsAbs(raw) {
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

// Walk never inspects Git or traverses directories outside the configured
// boundary. add also receives missing pin roots so callers report them visibly.
func (s *Scope) Walk(ctx context.Context, add func(path string, pinned bool) error) error {
	seen := map[string]bool{}
	roots := append([]string{s.Root}, s.pins...)
	for i, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if seen[path] {
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if walkErr != nil {
				if os.IsNotExist(walkErr) && path == root {
					if i > 0 {
						seen[path] = true
						return add(path, true)
					}
					return nil
				}
				return walkErr
			}
			if !d.IsDir() {
				if path == root {
					return fmt.Errorf("workspace path %s is not a directory (symlink roots are not followed)", path)
				}
				return nil
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
				}
				return filepath.SkipDir
			} else if !os.IsNotExist(err) {
				return err
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("scan workspace %s: %w", root, err)
		}
	}
	return nil
}
