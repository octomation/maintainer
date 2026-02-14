---
code: MAIN-181
id: null
databaseId: null
number: null
url: null
title: "service: hub: keep the Filter menu inside a 390 px window"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: medium"
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

# service: hub: keep the Filter menu inside a 390 px window

## Problem

In a 390 px window, the Filter menu on `/tablo` and `/status` spans x=-82 to
198. The search field, the icons and the start of each dimension name are cut
off.

## PoC

1. Open `/tablo` in a 390 px window.
2. Click **Filter**.
3. Today: the menu starts 82 px left of the screen. Expected: the whole menu is
   on screen.

## Expected behavior

At 390 px the Filter menu lies fully inside the viewport on every page of the
hub. At desktop width it opens where it opens today.

## Acceptance criteria

- [ ] In headless Chrome at 390 px, after a click on **Filter** on `/tablo`,
      `/status` and `/inbox`, the box of the menu has left ≥ 0 and right ≤ 390.
- [ ] `make web-check` passes, and the layout follows `web/DESIGN.md` in both
      themes.

## Out of scope

The touch size of the controls.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "At 390 px the Filter menu opens 82 px past the left screen edge".
