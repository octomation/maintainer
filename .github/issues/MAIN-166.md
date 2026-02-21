---
code: MAIN-166
id: I_kwDOE2M9Zc8AAAABi07keg
databaseId: 6632170618
number: 569
url: https://github.com/octomation/maintainer/issues/569
title: "fetch: scan: recognise sparse and partial checkouts of tracked repositories"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: medium"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-26T17:41:51Z
updatedAt: 2026-11-27T18:40:30Z
lastEditedAt: 2026-11-27T18:40:30Z
closedAt: 2026-11-26T19:42:37Z
---

# fetch: scan: recognise sparse and partial checkouts of tracked repositories

## Problem

The scan skips a checkout that go-git cannot open, without a message: one with
`extensions.worktreeConfig` from `git sparse-checkout set`, or with
`extensions.partialClone`. After a rename on GitHub, `fetch` then clones the
repository a second time, and the new state record loses the old path.

## PoC

```bash
# a tracked checkout of acme/old; GitHub renamed the repository to acme/new
git -C "$checkout" sparse-checkout set --cone docs   # or a clone with --filter=blob:none
maintainer fetch
# today:    "clone acme/new", nothing about $checkout
# expected: "move" of $checkout to the new path, or a conflict at $checkout
```

## Expected behavior

- The scan recognises every usable checkout of a tracked repository, sparse
  checkouts and partial clones included.
- A checkout that the scan cannot read is reported with its path and the reason.
- A rename moves the existing checkout or reports a conflict for it. `fetch`
  never clones a tracked repository a second time.

## Acceptance criteria

- [x] A test with a sparse checkout of a renamed tracked repository plans
      `move`, not `clone`.
- [x] The same test with a partial clone plans `move`.
- [x] A test with a checkout that cannot be read reports a conflict with its
      path, and the state record keeps the old path.

## Out of scope

Sparse or partial clones made by `fetch` itself: it keeps cloning full
checkouts.

Related: [[MAIN-150]].

Source: [20261125T152232Z](../reports/20261125T152232Z.md), "A sparse or partial checkout is invisible to the scan, so a rename re-clones it".
