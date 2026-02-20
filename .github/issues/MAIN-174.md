---
code: MAIN-174
id: I_kwDOE2M9Zc8AAAABihz6mQ
databaseId: 6612122265
number: 557
url: https://github.com/octomation/maintainer/issues/557
title: "service: inbox: count only the selected issues in a group header"
issueType: Bug
labels:
  - "effort: medium"
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
createdAt: 2026-11-25T19:31:11Z
updatedAt: 2026-11-25T20:23:50Z
lastEditedAt: 2026-11-25T20:23:50Z
closedAt: 2026-11-25T20:07:36Z
---

# service: inbox: count only the selected issues in a group header

## Problem

With the inbox filtered by Status (To do) and grouped by priority, each group
header counts the issues of every status, so "High 4" heads one row.
`docs/content/service.md` says that the count of a group covers the selection.

## PoC

1. Open `/inbox`, set the Status filter to To do, and group by priority with
   **View**.
2. Today: a group with one row reads "High 4". Expected: "High 1".

## Expected behavior

The count in a group header of the inbox is the number of issues that match
every active filter, the Status filter included. It does not depend on how many
issues the page has loaded.

## Acceptance criteria

- [x] A test with issues in several statuses and a Status filter of To do gives
      each group the count of its To do issues.
- [x] The counts stay the same when a group holds more issues than the first
      page.
- [x] `go test ./internal/service/hub/...` and `make web-check` pass.

## Out of scope

The board `/tablo`, whose columns are the statuses.

Source: [20261125T152232Z](../reports/20261125T152232Z.md), "Inbox group counts ignore the Status filter of the list".
