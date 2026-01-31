package state_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "go.octolab.org/toolset/maintainer/internal/state"
)

func TestStore_RoundTrip(t *testing.T) {
	fs := afero.NewMemMapFs()
	store := NewStore(fs, "/state/fetch/state.json", nil)

	st := New()
	st.Upsert(Record{ID: 1, OwnerLogin: "acme", Name: "svc", Path: "/work/acme/svc", FirstSeen: time.Unix(1, 0).UTC()})
	require.NoError(t, store.Save(st))

	// permissions are 0600.
	info, err := fs.Stat("/state/fetch/state.json")
	require.NoError(t, err)
	assert.Equal(t, "-rw-------", info.Mode().String())

	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, Version, loaded.Version)
	require.Len(t, loaded.Repos, 1)
	assert.Equal(t, int64(1), loaded.Repos[0].ID)
}

func TestStore_MissingFileIsEmpty(t *testing.T) {
	store := NewStore(afero.NewMemMapFs(), "/nope/state.json", nil)
	st, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, Version, st.Version)
	assert.Empty(t, st.Repos)
}

func TestStore_RejectsLoosePermissions(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/state.json", []byte(`{"version":1}`), 0o644))
	_, err := NewStore(fs, "/state.json", nil).Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permissive")
}

func TestStore_RejectsUnknownVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "/state.json", []byte(`{"version":2}`), 0o600))
	_, err := NewStore(fs, "/state.json", nil).Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

type fakeLocker struct {
	ok       bool
	unlocked bool
}

func (f *fakeLocker) TryLock() (bool, error) { return f.ok, nil }
func (f *fakeLocker) Unlock() error          { f.unlocked = true; return nil }

func TestStore_Lock(t *testing.T) {
	t.Run("acquired", func(t *testing.T) {
		lk := &fakeLocker{ok: true}
		store := NewStore(afero.NewMemMapFs(), "/s.json", lk)
		release, err := store.Lock()
		require.NoError(t, err)
		require.NoError(t, release())
		assert.True(t, lk.unlocked)
	})
	t.Run("contended fails fast", func(t *testing.T) {
		store := NewStore(afero.NewMemMapFs(), "/s.json", &fakeLocker{ok: false})
		_, err := store.Lock()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "locked by another")
	})
}

func TestState_Helpers(t *testing.T) {
	st := New()
	st.Upsert(Record{ID: 1, Name: "a"})
	st.Upsert(Record{ID: 1, Name: "a2"}) // replace
	st.Upsert(Record{ID: 2, Name: "b"})
	require.Len(t, st.Repos, 2)

	r, ok := st.ByID(1)
	require.True(t, ok)
	assert.Equal(t, "a2", r.Name)

	assert.True(t, st.Remove(2))
	assert.False(t, st.Remove(99))
	require.Len(t, st.Repos, 1)
}

func TestDefaultPath(t *testing.T) {
	xdg := func(k string) string { return map[string]string{"XDG_STATE_HOME": "/xdg"}[k] }
	assert.Equal(t, "/xdg/maintainer/fetch/state.json", DefaultPath(xdg, "/home/op"))
	none := func(string) string { return "" }
	assert.Equal(t, "/home/op/.local/state/maintainer/fetch/state.json", DefaultPath(none, "/home/op"))
}

// A save never writes through what already lies at the old temporary name or
// at the state path: neither a link nor a file is followed, truncated or
// chmodded, and the state path ends up a regular 0600 file.
func TestStore_SaveIgnoresPlantedPaths(t *testing.T) {
	for _, planted := range []string{"temp link", "temp file", "state link"} {
		t.Run(planted, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "state.json")
			victim := filepath.Join(t.TempDir(), "victim.txt")
			require.NoError(t, os.WriteFile(victim, []byte("precious\n"), 0o644))
			require.NoError(t, os.WriteFile(path, []byte(`{"version":1,"profiles":{},"repos":[]}`), 0o600))
			switch planted {
			case "temp link":
				require.NoError(t, os.Symlink(victim, path+".tmp"))
			case "temp file":
				victim = path + ".tmp"
				require.NoError(t, os.WriteFile(victim, []byte("precious\n"), 0o644))
			case "state link":
				require.NoError(t, os.Remove(path))
				require.NoError(t, os.Symlink(victim, path))
			}

			st := New()
			st.Upsert(Record{ID: 1, OwnerLogin: "acme", Name: "svc", Path: "/work/acme/svc"})
			require.NoError(t, NewStore(afero.NewOsFs(), path, nil).Save(st))

			raw, err := os.ReadFile(victim)
			require.NoError(t, err)
			assert.Equal(t, "precious\n", string(raw))
			info, err := os.Lstat(victim)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())

			info, err = os.Lstat(path)
			require.NoError(t, err)
			assert.True(t, info.Mode().IsRegular(), "the state path is %s", info.Mode())
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
			loaded, err := NewStore(afero.NewOsFs(), path, nil).Load()
			require.NoError(t, err)
			assert.Equal(t, st.Repos, loaded.Repos)

			leftovers, err := filepath.Glob(filepath.Join(dir, "state.json.*.tmp"))
			require.NoError(t, err)
			assert.Empty(t, leftovers)
		})
	}
}

// brokenFs fails a save midway: a write stops after half of the data, or the
// rename into place fails.
type brokenFs struct {
	afero.Fs
	write, rename bool
}

type halfFile struct{ afero.File }

func (f halfFile) Write(p []byte) (int, error) {
	n, _ := f.File.Write(p[:len(p)/2])
	return n, errors.New("injected write failure")
}

func (fs brokenFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	f, err := fs.Fs.OpenFile(name, flag, perm)
	if err != nil || !fs.write || flag&(os.O_WRONLY|os.O_RDWR) == 0 {
		return f, err
	}
	return halfFile{f}, nil
}

func (fs brokenFs) Rename(oldname, newname string) error {
	if fs.rename {
		return errors.New("injected rename failure")
	}
	return fs.Fs.Rename(oldname, newname)
}

func TestStore_FailedSaveKeepsPreviousState(t *testing.T) {
	for name, fs := range map[string]brokenFs{
		"write":  {Fs: afero.NewOsFs(), write: true},
		"rename": {Fs: afero.NewOsFs(), rename: true},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "state.json")
			previous := []byte(`{"version":1,"profiles":{},"repos":[{"id":7}]}` + "\n")
			require.NoError(t, os.WriteFile(path, previous, 0o600))

			st := New()
			st.Upsert(Record{ID: 1, OwnerLogin: "acme", Name: "svc"})
			require.Error(t, NewStore(fs, path, nil).Save(st))

			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, previous, raw)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)
			require.Len(t, entries, 1, "a failed save removes its temporary file")
			assert.Equal(t, "state.json", entries[0].Name())
		})
	}
}
