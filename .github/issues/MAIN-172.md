---
code: MAIN-172
id: null
databaseId: null
number: null
url: null
title: "service: hub: show the repository tree once the catalog loads again"
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

# service: hub: show the repository tree once the catalog loads again

## Problem

When the first catalog request of the hub fails, for example while the service
restarts, the sidebar shows no Repositories tree. **Try again** brings the page
back, but the tree returns only after a full reload.

## PoC

1. Open `/tablo` while `/api/v0/catalog` fails, for example during a restart of
   the service.
2. After the service is up, press **Try again**.
3. Today: the page loads, and the sidebar has no Repositories tree. Expected: the
   tree is there.

## Expected behavior

The sidebar shows the Repositories tree as soon as a catalog request succeeds,
without a page reload.

## Acceptance criteria

- [ ] A component test renders the app at `/tablo` with a catalog request that
      fails until **Try again**: afterwards the Repositories tree is present.
- [ ] `make web-check` passes.

Source: [20261125T152232Z](../reviews/20261125T152232Z.md), "The sidebar repository tree never retries a failed catalog request".
