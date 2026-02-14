---
code: MAIN-173
id: null
databaseId: null
number: null
url: null
title: "service: hub: keep F, P, E and Escape working on the full issue page"
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
state: OPEN
stateReason: null
duplicateOf: null
createdAt: null
updatedAt: null
lastEditedAt: null
closedAt: null
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

- [ ] A component test on `/inbox?task=<id>`: after `f` no popover exists, `p`
      opens its menu, and `Escape` removes `?task=` from the URL.
- [ ] On a list page F still opens the Filter menu.
- [ ] `make web-check` passes.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "F on the full issue page opens a hidden filter menu".
