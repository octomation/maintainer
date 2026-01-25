---
code: MAIN-121
id: I_kwDOE2M9Zc6EHWe3
databaseId: 2216519607
number: 220
url: https://github.com/octomation/maintainer/issues/220
title: "github: contributions: GitHub changed a way to provide data"
issueType: null
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: critical"
  - "impact: high"
  - "effort: medium"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2024-03-30T15:09:41Z
updatedAt: 2026-09-25T05:11:16Z
lastEditedAt: 2026-09-25T05:11:16Z
closedAt: 2024-03-30T17:19:02Z
---

# github: contributions: GitHub changed a way to provide data

Restore contribution loading after GitHub moved to providing the calendar asynchronously. An ordinary profile page stopped being a sufficient source: the page HTML can load successfully and still contain none of the required cells.

The contract, illustrated:

```text
Request the chosen user and year
→ a response carrying the contributions calendar
→ the existing dates and counts in lookup/snapshot
```

The calendar content itself has to be fetched, and the same mechanism has to be used when the test data is refreshed. A successful HTTP response without the expected content must not be presented as a confirmed empty year.

Related: [[MAIN-115]] (moving the counts into a tooltip, a distinct change), [[MAIN-120]] (checking a real response on a schedule).

<!-- 2024-03-30T17:19Z https://github.com/octomation/maintainer/issues/220#issuecomment-2028320047
fixed by 650e1a8ce0824ada64d0446622f0a9ebbbe48297
-->
