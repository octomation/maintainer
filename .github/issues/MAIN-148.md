---
code: MAIN-148
id: I_kwDOE2M9Zc8AAAABYusFYQ
databaseId: 5954536801
number: 504
url: https://github.com/octomation/maintainer/issues/504
title: "fetch: clone: do not adopt an interrupted clone as a checkout"
issueType: Bug
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M2]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-10-21T04:49:51Z
updatedAt: 2026-10-21T10:23:38Z
lastEditedAt: null
closedAt: 2026-10-21T04:49:54Z
---

# fetch: clone: do not adopt an interrupted clone as a checkout

## Problem

A clone that fails after it started — cut by `--timeout`, a signal or a dropped connection — keeps its partial checkout for inspection, by design ([[MAIN-123]]). The next run finds a repository with a GitHub origin there and adopts it like any complete checkout: it becomes tracked with no checked-out commit, is fetched on every run, and `status` shows it as `unborn`, as if the repository on GitHub were empty.

## PoC

```bash
maintainer fetch --apply --timeout=5s     # a plan with one clone of a large repository; the clone is cut
maintainer fetch                          # today: adopt for the partial checkout — expected: a conflict naming it
maintainer status --format=json | jq -r '.[] | select(.status == "unborn") | .path'   # today: the partial clone
```

As a regression test: the real Git adapter cloning from a disposable local origin with commits, wrapped like `localOriginSync` in `service_test.go`, the context cancelled once the clone has started writing the target; then a second plan against unchanged inputs.

## Expected behavior

The next run neither adopts nor fetches an interrupted clone as if it were complete: it reports the path as an incomplete clone, or completes it, and says which. A clone of an empty upstream stays a valid unborn checkout ([[MAIN-124]]), and nothing is deleted.

## Acceptance criteria

- [ ] After a clone cut by `--timeout`, the next plan has no `adopt` and no `fetch` for that path, but an action that names the path as an incomplete clone.
- [ ] The tests of [[MAIN-124]] for an empty upstream still pass.
- [ ] The partial checkout is never removed recursively ([[MAIN-123]]).
- [ ] `fetch.md` says what happens after an interrupted clone.

## Out of scope

Resuming a partial transfer; retries within one run; deleting partial checkouts.

## Sources

- [[20261019T105848Z#F3. An interrupted clone is adopted as a checkout]]
