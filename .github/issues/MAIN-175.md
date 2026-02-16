---
code: MAIN-175
id: I_kwDOE2M9Zc8AAAABihz_YQ
databaseId: 6612123489
number: 558
url: https://github.com/octomation/maintainer/issues/558
title: "fetch: state: prune only the records of the selected owners and profiles"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: medium"
  - "scope: code"
  - "scope: docs"
  - "scope: test"
  - "severity: major"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-25T19:31:15Z
updatedAt: 2026-11-25T19:31:15Z
lastEditedAt: null
closedAt: null
---

# fetch: state: prune only the records of the selected owners and profiles

## Problem

`fetch state prune --owner a` forgets the missing records of every owner. The
help of `--owner` and `--profile` says that others are left alone.

## PoC

```bash
# a state with missing paths for owners a and b
maintainer fetch --config "$tmp/fetch.toml" state prune --owner a; echo "exit=$?"
# today:    "pruned 2 record(s)", exit=0, no record left
# expected: "pruned 1 record(s)", the record of b stays
```

## Expected behavior

`state prune` honours `--owner` and `--profile` as `fetch` does: it forgets only
the missing records of the selected owners and profiles. Without a selection, it
prunes every missing record, as today.

## Acceptance criteria

- [x] The PoC prunes the record of `a` and keeps the record of `b`.
- [x] A test covers `--owner`, `--profile` and both together.
- [x] `docs/content/fetch.md` says that `state prune` honours the selection.

## Out of scope

Which records count as missing: [[MAIN-137]] decides it.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "`state prune` ignores `--owner` and `--profile` and forgets records of other owners".
