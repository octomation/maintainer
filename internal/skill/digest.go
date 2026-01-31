package skill

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// DigestAlgorithm names the payload digest shared with the octolab/skills
// catalog: for every file in byte order of its relative POSIX path, sha256
// takes the path, NUL, the decimal size, NUL and the contents.
const DigestAlgorithm = "octolab-skill-sha256-v1"

// Files returns the skill's files as sorted relative paths. Symbolic links
// and dotfiles are errors: they don't survive every channel, and dotfiles
// are reserved for installers' ownership markers.
func Files(fsys fs.FS) ([]string, error) {
	return walk(fsys, false)
}

// walk lists the files of a skill or, when installed is set, of an installed
// copy: there the ownership marker and Finder's .DS_Store are left out, and
// any other dotfile counts as content, so a copy holding one is changed.
func walk(fsys fs.FS, installed bool) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case path == ".":
			return nil
		case installed && !d.IsDir() && (path == MarkerFile || d.Name() == ".DS_Store"):
			return nil
		case strings.HasPrefix(d.Name(), ".") && !installed:
			return fmt.Errorf("%s: dotfiles are not allowed in a skill", path)
		case d.Type()&fs.ModeSymlink != 0:
			return fmt.Errorf("%s: symbolic links are not allowed in a skill", path)
		case d.IsDir():
			return nil
		case !d.Type().IsRegular():
			return fmt.Errorf("%s: unsupported file type", path)
		}
		files = append(files, path)
		return nil
	})
	sort.Strings(files)
	return files, err
}

// Digest returns the DigestAlgorithm digest of the skill as "sha256:<hex>".
func Digest(fsys fs.FS) (string, error) {
	files, err := Files(fsys)
	if err != nil {
		return "", err
	}
	return digest(fsys, files)
}

// CopyDigest returns the digest of an installed copy, leaving out the
// ownership marker and .DS_Store.
func CopyDigest(fsys fs.FS) (string, error) {
	files, err := walk(fsys, true)
	if err != nil {
		return "", err
	}
	return digest(fsys, files)
}

func digest(fsys fs.FS, files []string) (string, error) {
	h := sha256.New()
	for _, path := range files {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return "", err
		}
		h.Write([]byte(path))
		h.Write([]byte{0})
		h.Write([]byte(strconv.Itoa(len(data))))
		h.Write([]byte{0})
		h.Write(data)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
