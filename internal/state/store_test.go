package state_test

import (
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
