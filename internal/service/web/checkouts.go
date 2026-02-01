package web

import (
	"context"
	"sync"
	"time"

	"go.octolab.org/toolset/maintainer/internal/service/status"
)

// Checkouts caches the rows of status.Collect: the first request and every
// refresh collect, at most one collection at a time, and refresh requests
// during a collection coalesce into one more. Filtering never collects.
type Checkouts struct {
	collect func(ctx context.Context) ([]status.Row, error)
	now     func() time.Time

	mu          sync.Mutex
	rows        []status.Row
	collectedAt time.Time
	err         string
	has         bool
	running     bool
	pending     bool
	wake        chan struct{}
}

// NewCheckouts returns a cache over collect.
func NewCheckouts(collect func(ctx context.Context) ([]status.Row, error)) *Checkouts {
	return &Checkouts{collect: collect, now: time.Now, wake: make(chan struct{}, 1)}
}

// Start serves the collection requests until ctx ends.
func (c *Checkouts) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-c.wake:
				c.run(ctx)
			}
		}
	}()
}

func (c *Checkouts) run(ctx context.Context) {
	for {
		c.mu.Lock()
		c.running, c.pending = true, false
		c.mu.Unlock()
		rows, err := c.collect(ctx)
		c.mu.Lock()
		c.running = false
		if err != nil {
			c.err = err.Error()
		} else {
			c.rows, c.collectedAt, c.err, c.has = rows, c.now().UTC(), "", true
		}
		again := c.pending
		c.mu.Unlock()
		if !again || ctx.Err() != nil {
			return
		}
	}
}

// Refresh asks for a collection; during one, it asks for one more after it.
func (c *Checkouts) Refresh() {
	c.mu.Lock()
	c.pending = true
	c.mu.Unlock()
	select {
	case c.wake <- struct{}{}:
	default:
	}
}

// Snapshot returns the cached rows and when they were collected. ok is false
// before the first collection; refreshing reports one in progress or asked
// for; failure is the error of the last collection.
func (c *Checkouts) Snapshot() (rows []status.Row, collectedAt time.Time, ok, refreshing bool, failure string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rows, c.collectedAt, c.has, c.running || c.pending, c.err
}

// SetCollectedAt replaces the time of the cached rows: fixtures pin it.
func (c *Checkouts) SetCollectedAt(at time.Time) {
	c.mu.Lock()
	c.collectedAt = at
	c.mu.Unlock()
}

// Ensure asks for the first collection when none happened or is in progress.
func (c *Checkouts) Ensure() {
	c.mu.Lock()
	needed := !c.has && !c.running && !c.pending
	c.mu.Unlock()
	if needed {
		c.Refresh()
	}
}
