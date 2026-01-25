---
code: MAIN-75
id: I_kwDOE2M9Zc5OmBQ8
databaseId: 1318589500
number: 75
url: https://github.com/octomation/maintainer/issues/75
title: "github: contribution: suggest shows future"
issueType: null
labels:
  - "scope: code"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2022-07-26T18:07:42Z
updatedAt: 2026-09-25T05:13:38Z
lastEditedAt: 2026-09-25T05:13:38Z
closedAt: 2023-04-05T20:54:20Z
---

# github: contribution: suggest shows future

Distinguish the future from days without contributions in the suggest calendar, and do not show weeks that lie entirely in the future. Zeros in the future look like suitable gaps and mislead the user.

The original calls:

```bash
maintainer github contribution suggest --delta --target=10 2022-05-01/+12
maintainer github contribution suggest --delta --target=10 2022-07-24
```

At the time of the original report, 27 July and the days after it should have shown `?`, and week `#31` onwards should have been absent. A past day with a zero count stays `-`.

This concerns displaying the calendar: keeping the suggestion timestamp out of the future once the random offset is added is a separate concern.

Related: the random offset [[MAIN-118]].
