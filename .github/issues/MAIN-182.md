---
code: MAIN-182
id: null
databaseId: null
number: null
url: null
title: "service: hub: give every control a 44 px touch target below 600 px"
issueType: Bug
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
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

# service: hub: give every control a 44 px touch target below 600 px

## Problem

In a 390 px window, the Priority and Estimate triggers and the back button of
the breadcrumb are 28 px high. The segments of the view switch are 32×40 px
inside an 84 px frame, with a 20 px empty tail. `web/DESIGN.md` says that touch
targets are 44 px below 600 px.

## PoC

1. Open an issue page and `/tablo` in a 390 px window.
2. Measure `.prop__button`, `.issue-header__crumb--back` and
   `.view-switch__item` in the developer tools.
3. Today: 28, 28 and 32×40 px. Expected: at least 44×44 px each.

## Expected behavior

Below 600 px every interactive control of the hub offers a target of at least
44×44 px, as `web/DESIGN.md` states. The view switch has no empty tail. A
control can look smaller than its target.

## Acceptance criteria

- [ ] In headless Chrome at 390 px, the three controls of the PoC have targets
      of at least 44×44 px.
- [ ] A check over every interactive element of `/status`, `/inbox`, `/tablo`
      and the issue page at 390 px finds no target under 44 px.
- [ ] `npx @google/design.md lint web/DESIGN.md` reports no errors, and `make
      web-check` passes.

## Out of scope

The desktop layout.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "Below 600 px several controls stay under the 44 px touch size".
