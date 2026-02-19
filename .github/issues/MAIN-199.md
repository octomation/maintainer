---
code: MAIN-199
id: I_kwDOE2M9Zc8AAAABi08FcQ
databaseId: 6632179057
number: 577
url: https://github.com/octomation/maintainer/issues/577
title: "service: checkouts: run one more collection for refreshes during a collection"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: low"
  - "scope: code"
  - "scope: test"
  - "severity: minor"
  - "type: bug"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: OPEN
stateReason: null
duplicateOf: null
createdAt: 2026-11-26T17:42:26Z
updatedAt: 2026-11-26T17:43:07Z
lastEditedAt: 2026-11-26T17:43:07Z
closedAt: null
---

# service: checkouts: run one more collection for refreshes during a collection

## Problem

`internal/service/web/checkouts.go` says that refresh requests during a
collection coalesce into one more collection. Each **Refresh** pressed while
the hub collects the status of the checkouts can start a further Git pass over
every checkout, so five presses run three collections, and the page shows the
refresh for longer.

## PoC

1. Hold the first collection of the hub.
2. Press **Refresh** five times.
3. Today: three collections. Expected: two, the running one and one more.

## Expected behavior

Any number of refresh requests during a collection yields exactly one more
collection after it, as the comment of the code states.

## Acceptance criteria

- [x] A test holds a collection, sends several refresh requests and counts
      exactly one more collection.
- [x] A refresh request when no collection runs still starts one.

## Out of scope

How long one collection takes: [[MAIN-170]].

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "A refresh during a collection runs two more collections instead of one".
