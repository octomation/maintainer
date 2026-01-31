package skill

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// MarkerFile is the ownership marker that maintainer writes next to SKILL.md.
// CopyDigest leaves it out, so the marker doesn't change the digest.
const MarkerFile = ".maintainer-skill.json"

// Marker records what maintainer installed into a directory.
type Marker struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Digest  string `json:"digest"`
}

// Outcome tells what Export did.
type Outcome int

const (
	Installed Outcome = iota + 1
	Replaced
	Unchanged
)

// RefusedError explains why Export left a directory alone.
type RefusedError struct{ Reason string }

func (e *RefusedError) Error() string { return e.Reason }

// ReadMarker reads the ownership marker of an installed copy.
func ReadMarker(dir string) (Marker, error) {
	var m Marker
	data, err := os.ReadFile(filepath.Join(dir, MarkerFile))
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("%s: %w", MarkerFile, err)
	}
	return m, nil
}

// Export writes the skill from src into dir together with the marker. It
// writes a new or empty directory, and replaces only a copy it owns: one
// with a marker whose digest still matches the files, and only when replace
// is set. The same content is left as it is. Symbolic links and directories
// of other installers are never touched.
func Export(src fs.FS, dir string, marker Marker, replace bool) (Outcome, error) {
	refuse := func(format string, args ...any) (Outcome, error) {
		return 0, &RefusedError{Reason: fmt.Sprintf(format, args...)}
	}

	info, err := os.Lstat(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return Installed, write(src, dir, marker, false)
	case err != nil:
		return 0, err
	case info.Mode()&fs.ModeSymlink != 0:
		return refuse("%s is a symbolic link, likely from another installer; remove it or choose another directory", dir)
	case !info.IsDir():
		return refuse("%s is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	if len(entries) == 0 {
		return Installed, write(src, dir, marker, true)
	}

	owned, err := ReadMarker(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return refuse("%s was not installed by maintainer; remove it or choose another directory", dir)
	case err != nil:
		return refuse("%s has an unreadable marker: %v", dir, err)
	}
	current, err := CopyDigest(os.DirFS(dir))
	if err != nil || current != owned.Digest {
		return refuse("%s was changed after maintainer installed it; move it away and try again", dir)
	}
	if owned.Digest == marker.Digest {
		return Unchanged, nil
	}
	if !replace {
		return refuse("%s holds the skill %s; pass --replace to replace it with %s", dir, owned.Version, marker.Version)
	}
	return Replaced, write(src, dir, marker, true)
}

// write builds the copy in a sibling directory and swaps it in, so a failure
// leaves the previous copy in place.
func write(src fs.FS, dir string, marker Marker, exists bool) error {
	parent := filepath.Dir(dir)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(parent, ".maintainer-skill-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	if err := copyFS(src, tmp); err != nil {
		return err
	}
	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmp, MarkerFile), append(data, '\n'), 0o644); err != nil {
		return err
	}

	if !exists {
		return os.Rename(tmp, dir)
	}
	old := tmp + ".old"
	if err := os.Rename(dir, old); err != nil {
		return err
	}
	if err := os.Rename(tmp, dir); err != nil {
		return errors.Join(err, os.Rename(old, dir))
	}
	return os.RemoveAll(old)
}

func copyFS(src fs.FS, dst string) error {
	files, err := Files(src)
	if err != nil {
		return err
	}
	for _, name := range files {
		data, err := fs.ReadFile(src, name)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
