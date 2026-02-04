package github

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.octolab.org/safe"
	"go.octolab.org/unsafe"
)

// cacheVersion is the format of a cached year; a file of another one is
// not read, its months are fetched again.
const cacheVersion = 1

// DefaultCacheDir returns where the contribution events are cached:
// $XDG_CACHE_HOME/maintainer/contributions (fallback $HOME/.cache).
// It is empty, so nothing is cached, if there is neither.
func DefaultCacheDir(getenv func(string) string, home string) string {
	base := getenv("XDG_CACHE_HOME")
	if base == "" {
		if home == "" {
			return ""
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "maintainer", "contributions")
}

// cachedYear is the file of a year of a user in the cache.
type cachedYear struct {
	Version int                    `json:"version"`
	Months  map[string]monthEvents `json:"months"`
}

func (srv *Service) cachePath(user string, year int) string {
	if srv.cache == "" || user == "" || strings.ContainsAny(user, `/\.`) {
		return ""
	}
	return filepath.Join(srv.cache, strings.ToLower(user), fmt.Sprintf("%d.json", year))
}

// loadYear returns the cached months of the year by YYYY-MM. The cache only
// saves requests: a file that is missing, unreadable or of another format
// is no cache at all.
func (srv *Service) loadYear(user string, year int) map[string]monthEvents {
	path := srv.cachePath(user, year)
	if path == "" {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cached cachedYear
	if err := json.Unmarshal(raw, &cached); err != nil || cached.Version != cacheVersion {
		return nil
	}
	return cached.Months
}

// storeYear replaces the cached months of the year, as a whole, so a reader
// sees either the previous file or the new one. A failure leaves the cache as
// it was: the next run fetches the months again.
func (srv *Service) storeYear(user string, year int, months map[string]monthEvents) {
	path := srv.cachePath(user, year)
	if path == "" {
		return
	}
	raw, err := json.Marshal(cachedYear{Version: cacheVersion, Months: months})
	if err != nil {
		return
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	tmp, err := os.CreateTemp(dir, ".year-*.json")
	if err != nil {
		return
	}
	defer safe.Do(func() error { return os.Remove(tmp.Name()) }, unsafe.Ignore)
	if _, err := tmp.Write(raw); err != nil {
		safe.Close(tmp, unsafe.Ignore)
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}
	_ = os.Rename(tmp.Name(), path)
}
