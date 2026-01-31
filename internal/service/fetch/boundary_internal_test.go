package fetch

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitsvc "go.octolab.org/toolset/maintainer/internal/service/git"
	"go.octolab.org/toolset/maintainer/internal/service/workspace"
)

type recordingSync struct {
	gitsvc.GitSync
	cloned, moved []string
}

func (r *recordingSync) Clone(_ context.Context, opt gitsvc.CloneOptions) error {
	r.cloned = append(r.cloned, opt.Path)
	return nil
}

func (r *recordingSync) Move(_, to string) error {
	r.moved = append(r.moved, to)
	return nil
}

func TestBoundedSyncChecksAtTheMutationPoint(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	inner := new(recordingSync)
	sync := boundedSync{GitSync: inner, root: root}
	linked, occupied := filepath.Join(root, "acme/svc"), filepath.Join(root, "beta/svc")
	// Both destinations were clear when the plan was made.
	require.NoError(t, workspace.Destination(root, linked))
	require.NoError(t, workspace.Destination(root, occupied))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "acme")))
	require.NoError(t, os.MkdirAll(occupied, 0o755))

	ctx := context.Background()
	assert.ErrorIs(t, sync.Clone(ctx, gitsvc.CloneOptions{Path: linked}), workspace.ErrSymlinkParent)
	assert.ErrorIs(t, sync.Move(filepath.Join(root, "old/svc"), linked), workspace.ErrSymlinkParent)
	assert.ErrorIs(t, sync.Clone(ctx, gitsvc.CloneOptions{Path: occupied}), fs.ErrExist)
	assert.ErrorIs(t, sync.Move(filepath.Join(root, "old/svc"), occupied), fs.ErrExist)
	assert.ErrorContains(t, sync.Clone(ctx, gitsvc.CloneOptions{Path: filepath.Join(outside, "svc")}), "outside the workspace root")
	assert.Empty(t, inner.cloned)
	assert.Empty(t, inner.moved)

	fresh := filepath.Join(root, "gamma/svc")
	require.NoError(t, sync.Clone(ctx, gitsvc.CloneOptions{Path: fresh}))
	require.NoError(t, sync.Move(filepath.Join(root, "old/svc"), fresh))
	assert.Equal(t, []string{fresh}, inner.cloned)
	assert.Equal(t, []string{fresh}, inner.moved)
}
