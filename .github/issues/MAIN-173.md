---
code: MAIN-173
id: I_kwDOE2M9Zc8AAAABihz0-w
databaseId: 6612120827
number: 556
url: https://github.com/octomation/maintainer/issues/556
title: "service: hub: keep F, P, E and Escape working on the full issue page"
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
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2026-11-25T19:31:06Z
updatedAt: 2026-11-25T20:23:48Z
lastEditedAt: 2026-11-25T20:23:48Z
closedAt: 2026-11-25T20:08:47Z
---

# service: hub: keep F, P, E and Escape working on the full issue page

## Problem

With the full issue page open, F opens the Filter menu inside the hidden top
bar. After that P and E do nothing, and the first Escape closes the invisible
menu instead of returning to the list.

## PoC

1. Open `/inbox?task=<id>`.
2. Press F, then P, then Escape.
3. Today: no visible change, P opens nothing, and the URL keeps `?task=`.
   Expected: F does nothing, P opens its menu, and Escape returns to the list.

## Expected behavior

F opens the filters only while they are visible. On the full issue page, P, E
and Escape act as they do when F was not pressed.

## Acceptance criteria

- [x] A component test on `/inbox?task=<id>`: after `f` no popover exists, `p`
      opens its menu, and `Escape` removes `?task=` from the URL.
- [x] On a list page F still opens the Filter menu.
- [x] `make web-check` passes.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "F on the full issue page opens a hidden filter menu".
