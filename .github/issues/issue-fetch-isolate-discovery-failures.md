---
code:
id:
databaseId:
number:
url:
title: "fetch: discovery: isolate repository failures and respect rate limits"
labels:
  - "type: bug"
  - "scope: code"
  - "scope: test"
  - "severity: major"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-3]]"
state: OPEN
stateReason:
createdAt: 2026-09-26T09:26:00Z
updatedAt: 2026-09-26T09:26:00Z
lastEditedAt:
closedAt:
---

# fetch: discovery: isolate repository failures and respect rate limits

**Motivation:** a collection of 160 checkouts across eight owners must survive one organisation behind SAML SSO, one secondary rate limit and one DNS hiccup without leaving the user without a plan, and without the request pattern GitHub says can get an integration banned.

**Requirements**

- A 401, 403 or 451 while resolving one clone by name yields id 0 with a per-clone error, so that checkout becomes a conflict row and the rest of the run proceeds.
- RateGuard distinguishes three classes and treats nothing else as retryable:
  - primary limit — 403 or 429 with `x-ratelimit-remaining: 0`: one wait until `x-ratelimit-reset`, then one retry;
  - secondary limit — 403 or 429 with `Retry-After`, or without it when the body names the secondary rate limit: one wait of `Retry-After` or at least 60 seconds with backoff, then one retry;
  - transient failure — 502, 503, 504 and retryable transport errors: the existing bounded retries with backoff and jitter stay.
  An ordinary 403 (permissions, SSO) is never retried. When the required wait exceeds the guard's maximum wait or the remaining `--timeout` budget, the call fails with a message that names the wait; it never retries earlier. Permanent transport errors (DNS `NXDOMAIN`, TLS failures) are not retried.
- Tests with a fake clock and a fake transport: an ordinary 403; a headerless secondary limit; an oversized `Retry-After`; a primary lockout with the reset beyond the budget; a transient 503; a permanent transport failure; a 403 while resolving one clone with a plan for the rest.

**Evidence:** findings 11 (`internal/service/fetch/service.go:562`, `adopter.go:126`) and 12 (`internal/service/github/rate.go:94`, `rate.go:151`) of [the code review](../reviews/20260926T071417Z.md), reproduced with scratch probes, and C10a of [the review of 2026-07-09](../reviews/20260709T132530Z.md) (`rate.go:66`), still open on `932b9c8`. GitHub's documentation of primary and secondary limits is the reference for the classes above.

**PoC**

Runnable as table-driven cases in `internal/service/github/rate_test.go` with an injected clock, reconstructed from the review's probes:

```text
secondary 403, Retry-After: 30, then 200     today: attempts=1 waits=[] status=403        expected: waits=[30s], status=200
primary 429, reset in 30 min, budget 0       today: attempts=6, total wait 12.7s, 429     expected: one wait of 30 min (the guard's cap permitting), then retry
primary 429, reset in 30 min, budget 5 min   expected: error naming the 30 min wait, no request
ordinary 403 (SSO)                           expected: no retry, error after 1 attempt
503 then 200                                 expected: one backoff, status=200
dial tcp: no such host                       today: 6 attempts                             expected: 1 attempt
```

And a service test where resolving `sso-org/tool` answers 403: today `scan workspace …: 403 Forbidden` and an empty stdout; expected a conflict row for that checkout and a plan for the rest.

**Out of scope:** the GraphQL client of the contribution commands; body replay for retried POSTs (C10b), the shared jitter source (C10c) and the shadowed builtin (C10d), which stay recorded in the reviews.

Related: [[issue-fetch-honor-owners-profiles-filters]], [[issue-workspace-enforce-boundaries]].
