package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
)

// EnvFetchConfig is the environment variable that points at a fetch config.
const EnvFetchConfig = "MAINTAINER_FETCH_CONFIG"

// fetchConfigNames are the file names of a fetch config, in lookup order.
var fetchConfigNames = []string{"fetch.toml", "fetch.yaml"}

// FetchConfigLookup carries the inputs needed to locate a fetch config file.
type FetchConfigLookup struct {
	// Explicit is the --config flag value; ExplicitSet records whether the
	// flag was provided at all (so --config="" can disable discovery).
	Explicit    string
	ExplicitSet bool
	// Getenv resolves environment variables (injected for testability).
	Getenv func(string) string
	// WorkDir is the current working directory. It is never a source of
	// configuration: Find only reports a fetch config found there.
	WorkDir string
	// Home is $HOME, used for the XDG fallback.
	Home string
	// NameIt tells, with %s for the path, how the caller names a config;
	// the default is "pass --config %s".
	NameIt string
	// Where names the folder of WorkDir in the notice; empty means
	// "the working directory".
	Where string
}

// Resolve returns the config path per the lookup order: --config, an existing
// file named by $MAINTAINER_FETCH_CONFIG, then the user folder. The second
// result reports whether file discovery is disabled (explicit --config="").
// An empty path with disabled=false means no config file was found, which is
// not an error — owners then must come from --owner.
func (l FetchConfigLookup) Resolve(fs afero.Fs) (path string, disabled bool) {
	getenv := l.getenv()

	// 1. explicit --config (--config="" disables discovery entirely).
	if l.ExplicitSet {
		if l.Explicit == "" {
			return "", true
		}
		return l.Explicit, false
	}

	// 2. MAINTAINER_FETCH_CONFIG env var.
	if env := getenv(EnvFetchConfig); env != "" {
		if exists(fs, env) {
			return env, false
		}
	}

	// 3. $XDG_CONFIG_HOME/maintainer/… (fallback $HOME/.config/maintainer/).
	if dir := l.UserDir(); dir != "" {
		for _, name := range fetchConfigNames {
			candidate := filepath.Join(dir, name)
			if exists(fs, candidate) {
				return candidate, false
			}
		}
	}

	return "", false
}

// UserDir returns the user folder of the configuration:
// $XDG_CONFIG_HOME/maintainer, else $HOME/.config/maintainer.
// It is empty when neither is known.
func (l FetchConfigLookup) UserDir() string {
	base := l.getenv()("XDG_CONFIG_HOME")
	if base == "" && l.Home != "" {
		base = filepath.Join(l.Home, ".config")
	}
	if base == "" {
		return ""
	}
	return filepath.Join(base, "maintainer")
}

// Find resolves the config like Resolve and checks the working directory,
// which is never read. When the config comes from the user folder, or none is
// found, a fetch config in the working directory gives a notice; when no other
// config is found, the run must stop with the same text as an error instead
// of the single-run mode. A config the run names, with --config or
// $MAINTAINER_FETCH_CONFIG, and --config="" give neither.
func (l FetchConfigLookup) Find(fs afero.Fs) (path, notice string, err error) {
	path, _ = l.Resolve(fs)
	env := l.getenv()(EnvFetchConfig)
	if l.ExplicitSet || (env != "" && path == env) || l.WorkDir == "" {
		return path, "", nil
	}
	local := ""
	for _, name := range fetchConfigNames {
		candidate := filepath.Join(l.WorkDir, name)
		if exists(fs, candidate) {
			local = candidate
			break
		}
	}
	if local == "" || (path != "" && l.same(path, local)) {
		return path, "", nil
	}
	target := filepath.Join("~", ".config", "maintainer", filepath.Base(local))
	if dir := l.UserDir(); dir != "" {
		target = filepath.Join(dir, filepath.Base(local))
	}
	nameIt := l.NameIt
	if nameIt == "" {
		nameIt = "pass --config %s"
	}
	where := l.Where
	if where == "" {
		where = "the working directory"
	}
	notice = fmt.Sprintf("%s in %s is not read; to use it, %s, set %s=%s, or move it to %s",
		filepath.Base(local), where, fmt.Sprintf(nameIt, local), EnvFetchConfig, local, target)
	if path == "" {
		return "", "", fmt.Errorf("no fetch config found: %s", notice)
	}
	return path, notice, nil
}

// same reports whether two paths, relative ones against WorkDir, are one file.
func (l FetchConfigLookup) same(a, b string) bool {
	abs := func(p string) string {
		if !filepath.IsAbs(p) {
			p = filepath.Join(l.WorkDir, p)
		}
		if resolved, err := filepath.EvalSymlinks(p); err == nil {
			return resolved
		}
		return filepath.Clean(p)
	}
	return abs(a) == abs(b)
}

func (l FetchConfigLookup) getenv() func(string) string {
	if l.Getenv == nil {
		return func(string) string { return "" }
	}
	return l.Getenv
}

func exists(fs afero.Fs, path string) bool {
	ok, err := afero.Exists(fs, path)
	return err == nil && ok
}
