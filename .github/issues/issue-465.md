---
code: null
id: I_kwDOE2M9Zc8AAAABToAC8Q
databaseId: 5611979505
number: 465
url: https://github.com/octomation/maintainer/issues/465
title: "ci/cd: verification: restore reliable CI and local gates"
issueType: null
labels:
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "scope: deps"
  - "scope: inventory"
  - "impact: high"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[milestone-3]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-09-28T08:58:35Z
updatedAt: 2026-10-19T06:30:38Z
lastEditedAt: null
closedAt: 2026-10-19T06:30:38Z
---

# ci/cd: verification: restore reliable CI and local gates

**Motivation:** the *Continuous integration* baseline on `main` is red, so it cannot serve as a clean acceptance gate for the v0.2.0 work, and `release-verify` shares the failing dependency gate, so no tag can be published. Locally, the default `make test` fails on its timeout and one of the two generation gates is always dirty, which teaches everyone to bypass the pre-push hook.

**Requirements**

- `TestInspectRenameAndConflict` runs its `git merge` with the same options and environment as `runGit` and asserts on `CONFLICT` in the merge output, not on any error.
- `golang.org/x/crypto` at v0.56.0 or newer with `go.sum` reconciled; the Git adapter tests rerun; `govulncheck ./...` clean.
- One contract for the committed mocks: `go-generate` produces the canonical committed form (it applies the same `goimports -local` step that `generate` applies), so `make generate git-check`, `make go-generate git-check`, `make verify` and `make release-verify` all leave a clean tree; the mocks are regenerated once.
- The local `TIMEOUT` default equals the CI value, so `make test` and the pre-push hook pass on a clean checkout.

**Evidence:** findings 1–4 of [the code review of 2026-09-26](../../.github/reviews/20260926T071417Z.md) — `internal/service/status/git_test.go:218`, `go.mod:81` (GO-2026-6354, GO-2026-6355 reachable from `internal/service/git/sync.go:111`), `Makefile:372` and `Makefile:420`, `Makefile:83`. The timeout is a local-only defect: CI already passes `TIMEOUT=2m`.

**PoC**

```bash
set -o pipefail
make test                                              # panic: test timed out after 1s … FAIL internal/service/status
govulncheck ./...                                      # exit 3: 2 vulnerabilities from 1 module
make generate git-check; make go-generate git-check    # one of the two fails on a clean tree
docker run --rm -u 1000:1000 -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomod \
  -v "$PWD":/src -w /src golang:1.27 \
  go test -run TestInspectRenameAndConflict ./internal/service/status/   # expected: 1, actual: 0
```

Expected after the fix: all four commands succeed on a clean checkout, and *Continuous integration* is green on the next push to `main`.

**Out of scope:** the tools-module scan (`make tools-check`), which fails by design on GO-2026-6225 and GO-2026-5932; every other finding of the review.

Related: [[issue-org-prepare-release-v020]], [[issue-175]].
