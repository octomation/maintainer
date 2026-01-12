package github

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/time/jitter"
)

// RateGuard is an http.RoundTripper that translates GitHub's rate-limit and
// transient error responses into a retry policy with exponential backoff and
// jitter (§11). Retry-After is honoured exactly when present; a primary limit
// (403 + X-RateLimit-Remaining: 0) waits until the reset window.
type RateGuard struct {
	base       http.RoundTripper
	lowWater   int
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
	jitter     jitter.Transformation
	sleep      func(ctx context.Context, d time.Duration) error
	now        func() time.Time
}

// RateGuardOption customises a RateGuard (tests disable sleeping).
type RateGuardOption func(*RateGuard)

// WithMaxRetries sets the retry budget (default 5).
func WithMaxRetries(n int) RateGuardOption { return func(g *RateGuard) { g.maxRetries = n } }

// WithSleeper overrides the wait primitive (tests inject a no-op).
func WithSleeper(fn func(ctx context.Context, d time.Duration) error) RateGuardOption {
	return func(g *RateGuard) { g.sleep = fn }
}

// NewRateGuard wraps base with the default retry policy.
func NewRateGuard(base http.RoundTripper, opts ...RateGuardOption) *RateGuard {
	if base == nil {
		base = http.DefaultTransport
	}
	g := &RateGuard{
		base:       base,
		lowWater:   100,
		maxRetries: 5,
		baseDelay:  500 * time.Millisecond,
		maxDelay:   30 * time.Second,
		jitter:     jitter.FullRandom(),
		sleep:      sleepCtx,
		now:        time.Now,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// RoundTrip implements http.RoundTripper. Discovery is GET-only in the PoC, so
// request bodies never need to be replayed.
func (g *RateGuard) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	for attempt := 0; ; attempt++ {
		resp, err := g.base.RoundTrip(req)
		if err != nil {
			if attempt < g.maxRetries {
				if werr := g.sleep(ctx, g.backoff(attempt)); werr != nil {
					return nil, werr
				}
				continue
			}
			return nil, err
		}

		wait, retry := g.shouldRetry(resp, attempt)
		if !retry {
			return resp, nil
		}
		drain(resp)
		if werr := g.sleep(ctx, wait); werr != nil {
			return nil, werr
		}
	}
}

// shouldRetry decides whether resp warrants another attempt and how long to
// wait first.
func (g *RateGuard) shouldRetry(resp *http.Response, attempt int) (time.Duration, bool) {
	if attempt >= g.maxRetries {
		return 0, false
	}
	switch {
	case resp.StatusCode == http.StatusForbidden && remaining(resp) == 0:
		return g.untilReset(resp), true
	case resp.StatusCode == http.StatusTooManyRequests:
		return retryAfter(resp, g.backoff(attempt)), true
	case resp.StatusCode == http.StatusBadGateway,
		resp.StatusCode == http.StatusServiceUnavailable,
		resp.StatusCode == http.StatusGatewayTimeout:
		return g.backoff(attempt), true
	default:
		return 0, false
	}
}

func (g *RateGuard) backoff(attempt int) time.Duration {
	d := g.baseDelay << attempt
	if d > g.maxDelay || d <= 0 {
		d = g.maxDelay
	}
	if g.jitter != nil {
		d = d/2 + g.jitter.Apply(d/2+1)
	}
	return d
}

// untilReset returns how long to wait for the primary rate limit to reset,
// capped at a single reset window so the guard never hangs indefinitely (§11).
func (g *RateGuard) untilReset(resp *http.Response) time.Duration {
	reset := resp.Header.Get("X-RateLimit-Reset")
	if reset == "" {
		return g.maxDelay
	}
	secs, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		return g.maxDelay
	}
	wait := time.Until(time.Unix(secs, 0))
	if wait < 0 {
		return 0
	}
	if cap := time.Hour; wait > cap {
		return cap
	}
	return wait
}

func remaining(resp *http.Response) int {
	v := resp.Header.Get("X-RateLimit-Remaining")
	if v == "" {
		return -1
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return -1
	}
	return n
}

func retryAfter(resp *http.Response, fallback time.Duration) time.Duration {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second
		}
	}
	return fallback
}

func drain(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	_ = resp.Body.Close()
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
