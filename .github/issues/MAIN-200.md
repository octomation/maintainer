---
code: MAIN-200
id: I_kwDOE2M9Zc8AAAABi08JPg
databaseId: 6632180030
number: 578
url: https://github.com/octomation/maintainer/issues/578
title: "service: hub: close the issue page without a duplicate history entry"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
assignees:
  - kamilsk
milestone: null
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-26T17:42:30Z
updatedAt: 2026-11-26T20:02:03Z
lastEditedAt: 2026-11-26T20:02:03Z
closedAt: 2026-11-26T19:47:33Z
---

# service: hub: close the issue page without a duplicate history entry

## Problem

The hub opens an issue page over the list, and closing it is meant to let the
browser's Forward reopen the issue, as the comment in `web/src/lib/task.ts`
says. The close replaces that entry instead, so the history grows by one, and
the browser's Back after a close shows the same list again.

## PoC

1. Open `/inbox`, open an issue, close it.
2. Press the browser's Back.
3. Today: `/inbox` again. Expected: the page before `/inbox`.

## Expected behavior

Closing the issue page leaves no history entry equal to the one before it. The
comment of the code states the behavior that the code has.

## Acceptance criteria

- [x] A test of open and close from `/inbox` ends with no two equal adjacent
      history entries, and Back leaves `/inbox`.
- [x] `make web-check` passes, and `web/dist` equals a fresh build.

## Out of scope

The layout of the issue page.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "Closing the issue page leaves a duplicate list entry in the history".
