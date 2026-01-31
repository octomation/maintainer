package github

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.octolab.org/toolset/maintainer/internal/pkg/time/jitter"
)

// RateGuard is an http.RoundTripper that translates GitHub's rate-limit and
// transient error responses into a retry policy (§11). It tells three classes
// apart and retries nothing else:
//
//   - a primary limit, a 403 or 429 with X-RateLimit-Remaining: 0, waits once
//     until X-RateLimit-Reset and retries once;
//   - a secondary limit, a 403 or 429 with Retry-After, or without it when the
//     body names the secondary rate limit, waits once for Retry-After or at
//     least a minute with backoff and retries once;
//   - a transient failure, a 502, 503, 504 or a retryable transport error, is
//     retried a bounded number of times with exponential backoff and jitter.
//
// A rate-limit wait longer than the guard's maximum or than the time left
// before the request's deadline fails the call at once with an error that
// names the wait; it is never shortened into an early retry.
type RateGuard struct {
	base         http.RoundTripper
	lowWater     int
	maxRetries   int
	baseDelay    time.Duration
	maxDelay     time.Duration
	minSecondary time.Duration
	maxWait      time.Duration
	jitter       jitter.Transformation
	sleep        func(ctx context.Context, d time.Duration) error
	now          func() time.Time
}

// RateGuardOption customises a RateGuard (tests disable sleeping).
type RateGuardOption func(*RateGuard)

// WithMaxRetries sets the retry budget for transient failures (default 5).
func WithMaxRetries(n int) RateGuardOption { return func(g *RateGuard) { g.maxRetries = n } }

// WithMaxWait sets the longest single rate-limit wait (default one hour, the
// length of a primary rate-limit window).
func WithMaxWait(d time.Duration) RateGuardOption { return func(g *RateGuard) { g.maxWait = d } }

// WithSleeper overrides the wait primitive (tests inject a no-op).
func WithSleeper(fn func(ctx context.Context, d time.Duration) error) RateGuardOption {
	return func(g *RateGuard) { g.sleep = fn }
}

// WithClock overrides the clock that rate-limit resets and deadlines are
// measured against (tests inject a fake one).
func WithClock(now func() time.Time) RateGuardOption { return func(g *RateGuard) { g.now = now } }

// NewRateGuard wraps base with the default retry policy.
func NewRateGuard(base http.RoundTripper, opts ...RateGuardOption) *RateGuard {
	if base == nil {
		base = http.DefaultTransport
	}
	g := &RateGuard{
		base:         base,
		lowWater:     100,
		maxRetries:   5,
		baseDelay:    500 * time.Millisecond,
		maxDelay:     30 * time.Second,
		minSecondary: time.Minute,
		maxWait:      time.Hour,
		jitter:       jitter.FullRandom(),
		sleep:        sleepCtx,
		now:          time.Now,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// limit is the class of a response that may be retried.
type limit int

const (
	limitNone limit = iota
	limitPrimary
	limitSecondary
	limitTransient
)

func (l limit) String() string {
	if l == limitPrimary {
		return "primary"
	}
	return "secondary"
}

// RateLimitWaitError reports a rate-limit wait the guard refused to sleep
// through because it exceeds the guard's maximum or the request's deadline.
type RateLimitWaitError struct {
	Limit    string        // "primary" or "secondary"
	Wait     time.Duration // the wait GitHub requires before the next request
	Allowed  time.Duration // the longest wait the guard may take
	Deadline bool          // Allowed is the time left before the deadline
}

func (e *RateLimitWaitError) Error() string {
	bound := "the longest wait of " + e.Allowed.String()
	if e.Deadline {
		bound = "the remaining --timeout budget of " + e.Allowed.String()
	}
	return fmt.Sprintf("GitHub %s rate limit: the next request is allowed in %s, beyond %s",
		e.Limit, e.Wait.Round(time.Second), bound)
}

// RoundTrip implements http.RoundTripper. Discovery is GET-only in the PoC, so
// request bodies never need to be replayed.
func (g *RateGuard) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	limited, transient := false, 0
	for {
		resp, err := g.base.RoundTrip(req)
		if err != nil {
			if transient >= g.maxRetries || !retryableTransport(ctx, err) {
				return nil, err
			}
			if werr := g.sleep(ctx, g.backoff(transient)); werr != nil {
				return nil, werr
			}
			transient++
			continue
		}

		class, wait := g.classify(resp, transient)
		switch class {
		case limitNone:
			return resp, nil
		case limitTransient:
			if transient >= g.maxRetries {
				return resp, nil
			}
			transient++
		default:
			if limited {
				return resp, nil // one wait per rate limit; the caller sees the second refusal
			}
			limited = true
			if werr := g.affordable(ctx, class, wait); werr != nil {
				drain(resp)
				return nil, werr
			}
		}
		drain(resp)
		if werr := g.sleep(ctx, wait); werr != nil {
			return nil, werr
		}
	}
}

// classify decides whether resp warrants another attempt and how long to
// wait first.
func (g *RateGuard) classify(resp *http.Response, transient int) (limit, time.Duration) {
	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusTooManyRequests:
		after, hasAfter := retryAfter(resp, g.now())
		if remaining(resp) == 0 {
			wait, ok := g.untilReset(resp)
			if !ok {
				wait = g.minSecondary + g.backoff(0)
			}
			if hasAfter && after > wait {
				wait = after
			}
			return limitPrimary, wait
		}
		if hasAfter {
			return limitSecondary, after
		}
		if namesSecondaryLimit(resp) {
			return limitSecondary, g.minSecondary + g.backoff(0)
		}
		return limitNone, 0 // permissions, SSO, IP allowlists: never retried
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return limitTransient, g.backoff(transient)
	default:
		return limitNone, 0
	}
}

// affordable refuses a wait beyond the guard's maximum or the time left
// before the request's deadline.
func (g *RateGuard) affordable(ctx context.Context, class limit, wait time.Duration) error {
	if wait > g.maxWait {
		return &RateLimitWaitError{Limit: class.String(), Wait: wait, Allowed: g.maxWait}
	}
	if deadline, ok := ctx.Deadline(); ok {
		if left := deadline.Sub(g.now()); wait > left {
			return &RateLimitWaitError{Limit: class.String(), Wait: wait, Allowed: max(left, 0).Round(time.Second), Deadline: true}
		}
	}
	return nil
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

// untilReset returns how long to wait for the primary rate limit to reset:
// until one second after X-RateLimit-Reset, the first moment GitHub accepts
// the next request.
func (g *RateGuard) untilReset(resp *http.Response) (time.Duration, bool) {
	secs, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
	if err != nil {
		return 0, false
	}
	return max(time.Unix(secs, 0).Add(time.Second).Sub(g.now()), 0), true
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

// retryAfter parses Retry-After as delay seconds or as an HTTP date.
func retryAfter(resp *http.Response, now time.Time) (time.Duration, bool) {
	v := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.ParseInt(v, 10, 64); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second, true
	}
	if at, err := http.ParseTime(v); err == nil {
		return max(at.Sub(now), 0), true
	}
	return 0, false
}

// namesSecondaryLimit reports whether the error body says a secondary rate
// limit was exceeded. The body stays readable for the caller.
func namesSecondaryLimit(resp *http.Response) bool {
	if resp.Body == nil {
		return false
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	resp.Body = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(data), resp.Body), resp.Body}
	body := strings.ToLower(string(data))
	return strings.Contains(body, "secondary rate limit") ||
		strings.Contains(body, "secondary-rate-limits") ||
		strings.Contains(body, "abuse-rate-limits")
}

// retryableTransport reports whether a transport error may pass on its own:
// a timeout, a refused, reset, aborted or dropped connection, an unreachable
// network, a temporary name server failure. Anything else,
// a cancelled or expired context, a name that does not exist, a TLS alert or
// a certificate GitHub's client cannot verify included, is permanent:
// retrying it only delays the error.
func retryableTransport(ctx context.Context, err error) bool {
	if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	if dnsErr := (*net.DNSError)(nil); errors.As(err, &dnsErr) {
		return !dnsErr.IsNotFound && (dnsErr.IsTimeout || dnsErr.IsTemporary)
	}
	for _, transient := range []error{
		syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ECONNABORTED, syscall.EPIPE, syscall.ETIMEDOUT,
		syscall.ENETDOWN, syscall.ENETUNREACH, syscall.EHOSTUNREACH, io.EOF, io.ErrUnexpectedEOF,
	} {
		if errors.Is(err, transient) {
			return true
		}
	}
	// net/http reports a connection the server closed before answering with
	// an unexported sentinel; only its text identifies it.
	if strings.Contains(err.Error(), "http: server closed idle connection") {
		return true
	}
	var timeout interface{ Timeout() bool }
	return errors.As(err, &timeout) && timeout.Timeout()
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
