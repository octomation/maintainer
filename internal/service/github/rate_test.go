package github_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.octolab.org/toolset/maintainer/internal/service/github"
)

// step is one scripted outcome: a response, or a transport error.
type step struct {
	status  int
	body    string
	headers []string
	err     error
}

func answer(status int, body string, headers ...string) step {
	return step{status: status, body: body, headers: headers}
}

func failure(err error) step { return step{err: err} }

// scriptedTransport answers each attempt with the next scripted step; the
// last one repeats.
type scriptedTransport struct {
	steps    []step
	attempts int
}

func (s *scriptedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	next := s.steps[min(s.attempts, len(s.steps)-1)]
	s.attempts++
	if next.err != nil {
		return nil, next.err
	}
	h := http.Header{}
	for i := 0; i+1 < len(next.headers); i += 2 {
		h.Set(next.headers[i], next.headers[i+1])
	}
	return &http.Response{StatusCode: next.status, Header: h, Body: io.NopCloser(strings.NewReader(next.body)), Request: req}, nil
}

// fakeClock advances only when the guard sleeps.
type fakeClock struct {
	now   time.Time
	waits []time.Duration
}

func (c *fakeClock) Now() time.Time { return c.now }

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	c.waits = append(c.waits, d)
	c.now = c.now.Add(d)
	return ctx.Err()
}

func TestRateGuard(t *testing.T) {
	start := time.Unix(1_800_000_000, 0)
	reset := strconv.FormatInt(start.Add(30*time.Minute).Unix(), 10)
	const secondaryBody = `{"message":"You have exceeded a secondary rate limit. Please wait a few minutes before you try again.","documentation_url":"https://docs.github.com/rest/overview/rate-limits-for-the-rest-api#about-secondary-rate-limits"}`
	const ssoBody = `{"message":"Resource protected by organization SAML enforcement. You must grant your Personal Access token access to this organization.","documentation_url":"https://docs.github.com/articles/authenticating-to-a-github-organization-with-saml-single-sign-on/"}`
	ok := answer(http.StatusOK, `{}`)
	nxdomain := &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Err: "no such host", Name: "api.github.com", IsNotFound: true}}

	for _, tc := range []struct {
		name     string
		steps    []step
		budget   time.Duration
		attempts int
		status   int
		waits    func(*testing.T, []time.Duration)
		err      string
	}{
		{
			name:     "an ordinary 403 is not retried",
			steps:    []step{answer(http.StatusForbidden, ssoBody, "X-RateLimit-Remaining", "4999"), ok},
			attempts: 1, status: http.StatusForbidden,
			waits: func(t *testing.T, w []time.Duration) { assert.Empty(t, w) },
		},
		{
			name:     "a secondary 403 waits for Retry-After",
			steps:    []step{answer(http.StatusForbidden, secondaryBody, "Retry-After", "30"), ok},
			attempts: 2, status: http.StatusOK,
			waits: func(t *testing.T, w []time.Duration) { assert.Equal(t, []time.Duration{30 * time.Second}, w) },
		},
		{
			name:     "a headerless secondary limit waits at least a minute",
			steps:    []step{answer(http.StatusForbidden, secondaryBody), ok},
			attempts: 2, status: http.StatusOK,
			waits: func(t *testing.T, w []time.Duration) {
				require.Len(t, w, 1)
				assert.GreaterOrEqual(t, w[0], time.Minute)
				assert.Less(t, w[0], time.Minute+time.Second)
			},
		},
		{
			name:     "a secondary 429 waits once and returns the second refusal",
			steps:    []step{answer(http.StatusTooManyRequests, secondaryBody, "Retry-After", "5")},
			attempts: 2, status: http.StatusTooManyRequests,
			waits: func(t *testing.T, w []time.Duration) { assert.Equal(t, []time.Duration{5 * time.Second}, w) },
		},
		{
			name:     "an oversized Retry-After fails without a retry",
			steps:    []step{answer(http.StatusForbidden, secondaryBody, "Retry-After", "7200"), ok},
			attempts: 1,
			waits:    func(t *testing.T, w []time.Duration) { assert.Empty(t, w) },
			err:      "GitHub secondary rate limit: the next request is allowed in 2h0m0s, beyond the longest wait of 1h0m0s",
		},
		{
			name:     "a primary 429 waits until the reset without a budget",
			steps:    []step{answer(http.StatusTooManyRequests, `{}`, "X-RateLimit-Remaining", "0", "X-RateLimit-Reset", reset), ok},
			attempts: 2, status: http.StatusOK,
			waits: func(t *testing.T, w []time.Duration) {
				assert.Equal(t, []time.Duration{30*time.Minute + time.Second}, w)
			},
		},
		{
			name:     "a primary 403 with the reset beyond the budget fails without a retry",
			steps:    []step{answer(http.StatusForbidden, `{}`, "X-RateLimit-Remaining", "0", "X-RateLimit-Reset", reset), ok},
			budget:   5 * time.Minute,
			attempts: 1,
			waits:    func(t *testing.T, w []time.Duration) { assert.Empty(t, w) },
			err:      "GitHub primary rate limit: the next request is allowed in 30m1s, beyond the remaining --timeout budget of 5m0s",
		},
		{
			name:     "a transient 503 is retried with backoff",
			steps:    []step{answer(http.StatusServiceUnavailable, ``), ok},
			attempts: 2, status: http.StatusOK,
			waits: func(t *testing.T, w []time.Duration) {
				require.Len(t, w, 1)
				assert.Less(t, w[0], time.Second)
			},
		},
		{
			name:     "transient 502s exhaust the bounded retries",
			steps:    []step{answer(http.StatusBadGateway, ``)},
			attempts: 6, status: http.StatusBadGateway,
			waits: func(t *testing.T, w []time.Duration) { assert.Len(t, w, 5) },
		},
		{
			name:     "a reset connection is retried",
			steps:    []step{failure(&net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNRESET)}), ok},
			attempts: 2, status: http.StatusOK,
			waits: func(t *testing.T, w []time.Duration) { assert.Len(t, w, 1) },
		},
		{
			name:     "a name that does not exist is not retried",
			steps:    []step{failure(nxdomain), ok},
			attempts: 1,
			waits:    func(t *testing.T, w []time.Duration) { assert.Empty(t, w) },
			err:      "no such host",
		},
		{
			name:     "a TLS failure is not retried",
			steps:    []step{failure(x509.UnknownAuthorityError{}), ok},
			attempts: 1,
			waits:    func(t *testing.T, w []time.Duration) { assert.Empty(t, w) },
			err:      "certificate signed by unknown authority",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clock := &fakeClock{now: start}
			transport := &scriptedTransport{steps: tc.steps}
			guard := github.NewRateGuard(transport, github.WithClock(clock.Now), github.WithSleeper(clock.Sleep))

			ctx := context.Background()
			if tc.budget > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, start.Add(tc.budget))
				defer cancel()
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/acme/tool", nil)
			require.NoError(t, err)

			resp, err := guard.RoundTrip(req)
			assert.Equal(t, tc.attempts, transport.attempts)
			tc.waits(t, clock.waits)
			if tc.err != "" {
				assert.ErrorContains(t, err, tc.err)
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.status, resp.StatusCode)
			_ = resp.Body.Close()
		})
	}
}

func TestRateGuard_KeepsTheBodyOfARefusal(t *testing.T) {
	const body = `{"message":"Must have admin rights to Repository."}`
	transport := &scriptedTransport{steps: []step{answer(http.StatusForbidden, body)}}
	guard := github.NewRateGuard(transport, github.WithSleeper(func(context.Context, time.Duration) error {
		t.Fatal("an ordinary 403 must not wait")
		return nil
	}))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.github.com/repos/acme/tool", nil)
	require.NoError(t, err)
	resp, err := guard.RoundTrip(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, body, string(data))
}

// countingTransport counts the attempts that reach the network.
type countingTransport struct {
	base     http.RoundTripper
	attempts int
}

func (c *countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	c.attempts++
	return c.base.RoundTrip(req)
}

func TestRateGuard_RealTransportFailures(t *testing.T) {
	// A server that accepts the connection and drops it before answering.
	dropping, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = dropping.Close() }()
	go func() {
		for {
			conn, err := dropping.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	// A server that refuses the client's TLS version with an alert.
	modern := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	modern.TLS = &tls.Config{MinVersion: tls.VersionTLS13}
	modern.StartTLS()
	defer modern.Close()

	for _, tc := range []struct {
		name     string
		url      string
		attempts int
		err      string
	}{
		{"a dropped connection is retried", "http://" + dropping.Addr().String(), 6, "EOF|connection reset|server closed idle connection"},
		{"a TLS alert is not retried", modern.URL, 1, "protocol version not supported"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := &countingTransport{base: &http.Transport{
				TLSClientConfig:   &tls.Config{MaxVersion: tls.VersionTLS12, InsecureSkipVerify: true}, //nolint:gosec // a local test server
				DisableKeepAlives: true,
			}}
			guard := github.NewRateGuard(base, github.WithSleeper(func(context.Context, time.Duration) error { return nil }))
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tc.url, nil)
			require.NoError(t, err)
			resp, err := guard.RoundTrip(req)
			if resp != nil {
				_ = resp.Body.Close()
			}
			require.Error(t, err)
			assert.Regexp(t, tc.err, err.Error())
			assert.Equal(t, tc.attempts, base.attempts)
		})
	}
}
