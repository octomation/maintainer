package web_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/status"
	"go.octolab.org/toolset/maintainer/internal/service/web"
)

func TestCheckoutsRefreshDuringCollection(t *testing.T) {
	started, release := make(chan struct{}, 8), make(chan struct{})
	checkouts := web.NewCheckouts(func(ctx context.Context) ([]status.Row, error) {
		started <- struct{}{}
		select {
		case <-release:
			return nil, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	checkouts.Start(ctx)
	waitStart := func(msg string) {
		t.Helper()
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			require.FailNow(t, msg)
		}
	}
	idle := func() bool { _, _, _, refreshing, _ := checkouts.Snapshot(); return !refreshing }

	checkouts.Refresh()
	waitStart("a refresh with no collection running starts one")
	for range 5 {
		checkouts.Refresh()
	}
	release <- struct{}{}
	waitStart("the refreshes during the collection run one more")
	release <- struct{}{}
	require.Eventually(t, idle, 5*time.Second, 10*time.Millisecond)
	assert.Never(t, func() bool { return len(started) > 0 }, 300*time.Millisecond, 10*time.Millisecond,
		"five refreshes during a collection run exactly one more")

	checkouts.Refresh()
	waitStart("a refresh after the collections starts one again")
	release <- struct{}{}
	require.Eventually(t, idle, 5*time.Second, 10*time.Millisecond)
	assert.Empty(t, started)
}
