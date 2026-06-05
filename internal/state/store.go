package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/spf13/afero"
)

// statePerm is the only permission a state file may have (§6.1).
const statePerm os.FileMode = 0o600

// DefaultPath returns the default state file location:
// $XDG_STATE_HOME/maintainer/fetch/state.json (fallback $HOME/.local/state).
func DefaultPath(getenv func(string) string, home string) string {
	base := getenv("XDG_STATE_HOME")
	if base == "" && home != "" {
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "maintainer", "fetch", "state.json")
}

// Locker is the advisory lock around a run (§6.3). The default implementation
// wraps gofrs/flock; tests inject a no-op.
type Locker interface {
	TryLock() (bool, error)
	Unlock() error
}

// Store loads and persists the state JSON document. The file content goes
// through afero (so round-trips are testable on a memory FS), while the lock
// uses a real path via the injected Locker.
type Store struct {
	fs   afero.Fs
	path string
	lock Locker
}

// NewStore builds a Store. A nil lock means no advisory locking (tests).
func NewStore(fs afero.Fs, path string, lock Locker) *Store {
	return &Store{fs: fs, path: path, lock: lock}
}

// Path returns the state file path.
func (s *Store) Path() string { return s.path }

// Lock acquires the advisory lock for the duration of a run. It returns a
// release func. A contended lock is a user error (the caller maps it to exit 2).
func (s *Store) Lock() (release func() error, err error) {
	if s.lock == nil {
		return func() error { return nil }, nil
	}
	ok, err := s.lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("acquire state lock: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("state file is locked by another maintainer fetch process")
	}
	return s.lock.Unlock, nil
}

// Load reads the state file. A missing file yields a fresh empty state. The
// 0600 permission guard and the schema-version check are enforced here.
func (s *Store) Load() (*State, error) {
	ok, err := afero.Exists(s.fs, s.path)
	if err != nil {
		return nil, fmt.Errorf("stat state file %q: %w", s.path, err)
	}
	if !ok {
		return New(), nil
	}
	info, err := s.fs.Stat(s.path)
	if err != nil {
		return nil, fmt.Errorf("stat state file %q: %w", s.path, err)
	}
	if perm := info.Mode().Perm(); perm&^statePerm != 0 {
		return nil, fmt.Errorf("state file %q permissions %#o are more permissive than %#o", s.path, perm, statePerm)
	}
	raw, err := afero.ReadFile(s.fs, s.path)
	if err != nil {
		return nil, fmt.Errorf("read state file %q: %w", s.path, err)
	}
	st := new(State)
	if err := json.Unmarshal(raw, st); err != nil {
		return nil, fmt.Errorf("parse state file %q: %w", s.path, err)
	}
	if st.Version != Version {
		return nil, fmt.Errorf("state file %q has unsupported version %d (this binary supports %d)", s.path, st.Version, Version)
	}
	if st.Profiles == nil {
		st.Profiles = map[string]time.Time{}
	}
	return st, nil
}

// Save writes the state atomically (temp + rename) with 0600 permissions,
// creating the parent directory if needed.
func (s *Store) Save(st *State) error {
	if st.Version == 0 {
		st.Version = Version
	}
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	raw = append(raw, '\n')

	dir := filepath.Dir(s.path)
	if err := s.fs.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state dir %q: %w", dir, err)
	}
	tmp := s.path + ".tmp"
	if err := afero.WriteFile(s.fs, tmp, raw, statePerm); err != nil {
		return fmt.Errorf("write state temp %q: %w", tmp, err)
	}
	if err := s.fs.Chmod(tmp, statePerm); err != nil {
		return fmt.Errorf("chmod state temp %q: %w", tmp, err)
	}
	if err := s.fs.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("rename state temp into place: %w", err)
	}
	return nil
}

// FileLock returns a Locker backed by gofrs/flock at "<path>.lock" (§6.3).
// The lock directory is created eagerly so a first run does not fail.
func FileLock(path string) (Locker, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create state dir %q: %w", dir, err)
	}
	return flock.New(path + ".lock"), nil
}
