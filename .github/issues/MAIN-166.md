---
code: MAIN-166
id: null
databaseId: null
number: null
url: null
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
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
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

- [ ] A test with a sparse checkout of a renamed tracked repository plans
      `move`, not `clone`.
- [ ] The same test with a partial clone plans `move`.
- [ ] A test with a checkout that cannot be read reports a conflict with its
      path, and the state record keeps the old path.

## Out of scope

Sparse or partial clones made by `fetch` itself: it keeps cloning full
checkouts.

Related: [[MAIN-150]].

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "A sparse or partial checkout is invisible to the scan, so a rename re-clones it".
