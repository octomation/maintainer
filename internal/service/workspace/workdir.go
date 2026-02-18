package workspace

import (
	"os"
	"path/filepath"
)

// WorkDir returns the working directory that a relative or implicit root
// resolves against. A working directory reached through a symlink is the
// folder the symlink names: the implicit root "." scans it. A root the user
// names as a symlink is still refused by the scan.
func WorkDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if info, err := os.Lstat(cwd); err == nil && info.Mode()&os.ModeSymlink != 0 {
		if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
			cwd = resolved
		}
	}
	return cwd, nil
}
