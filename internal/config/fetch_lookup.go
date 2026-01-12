package config

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// EnvFetchConfig is the environment variable that points at a fetch config.
const EnvFetchConfig = "MAINTAINER_FETCH_CONFIG"

// FetchConfigLookup carries the inputs needed to locate a fetch config file.
type FetchConfigLookup struct {
	// Explicit is the --config flag value; ExplicitSet records whether the
	// flag was provided at all (so --config="" can disable discovery).
	Explicit    string
	ExplicitSet bool
	// Getenv resolves environment variables (injected for testability).
	Getenv func(string) string
	// WorkDir is the current working directory used for ./fetch.{toml,yaml}.
	WorkDir string
	// Home is $HOME, used for the XDG fallback.
	Home string
}

// Resolve returns the config path per the lookup order in §4.1. The second
// result reports whether file discovery is disabled (explicit --config="").
// An empty path with disabled=false means no config file was found, which is
// not an error — owners then must come from --owner (§4.1 "no config present").
func (l FetchConfigLookup) Resolve(fs afero.Fs) (path string, disabled bool) {
	getenv := l.Getenv
	if getenv == nil {
		getenv = func(string) string { return "" }
	}

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

	// 3. ./fetch.toml then ./fetch.yaml in the working directory.
	for _, name := range []string{"fetch.toml", "fetch.yaml"} {
		candidate := filepath.Join(l.WorkDir, name)
		if exists(fs, candidate) {
			return candidate, false
		}
	}

	// 4. $XDG_CONFIG_HOME/maintainer/… (fallback $HOME/.config/maintainer/).
	base := getenv("XDG_CONFIG_HOME")
	if base == "" && l.Home != "" {
		base = filepath.Join(l.Home, ".config")
	}
	if base != "" {
		for _, name := range []string{"fetch.toml", "fetch.yaml"} {
			candidate := filepath.Join(base, "maintainer", name)
			if exists(fs, candidate) {
				return candidate, false
			}
		}
	}

	return "", false
}

func exists(fs afero.Fs, path string) bool {
	ok, err := afero.Exists(fs, path)
	return err == nil && ok
}
