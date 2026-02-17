---
code: MAIN-181
id: I_kwDOE2M9Zc8AAAABih0Zfw
databaseId: 6612130175
number: 563
url: https://github.com/octomation/maintainer/issues/563
title: "service: hub: keep the Filter menu inside a 390 px window"
issueType: Bug
labels:
  - "effort: easy"
  - "impact: medium"
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
createdAt: 2026-11-25T19:31:38Z
updatedAt: 2026-11-25T20:24:08Z
lastEditedAt: 2026-11-25T20:24:08Z
closedAt: 2026-11-25T20:09:37Z
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

- [x] In headless Chrome at 390 px, after a click on **Filter** on `/tablo`,
      `/status` and `/inbox`, the box of the menu has left ≥ 0 and right ≤ 390.
- [x] `make web-check` passes, and the layout follows `web/DESIGN.md` in both
      themes.

## Out of scope

The touch size of the controls.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "At 390 px the Filter menu opens 82 px past the left screen edge".
