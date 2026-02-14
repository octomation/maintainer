package web_test

import (
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/web"
)

// The built application registers no service worker and keeps nothing in
// sessionStorage or a cookie: an origin that another program may serve later
// must find nothing of the user's there.
func TestDistStoresNothingOutsideTheTheme(t *testing.T) {
	scripts := 0
	require.NoError(t, fs.WalkDir(web.Dist(), ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(name) != ".js" {
			return err
		}
		scripts++
		raw, err := fs.ReadFile(web.Dist(), name)
		require.NoError(t, err)
		for _, api := range []string{"serviceWorker", "document.cookie"} {
			assert.NotContains(t, string(raw), api, name)
		}
		// sessionStorage is only cleared, on a plain address.
		assert.Equal(t, strings.Count(string(raw), "sessionStorage"), strings.Count(string(raw), "sessionStorage.clear()"), name)
		// The expanded owners of earlier releases are only removed.
		const legacy = "maintainer.sidebar.owners"
		assert.Equal(t, strings.Count(string(raw), legacy), strings.Count(string(raw), "localStorage.removeItem(`"+legacy+"`)"), name)
		return nil
	}))
	assert.Positive(t, scripts)
}
