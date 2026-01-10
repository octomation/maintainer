---
code:
id: I_kwDOE2M9Zc5OmBQ8
databaseId: 1318589500
number: 75
url: https://github.com/octomation/maintainer/issues/75
title: "github: contribution: suggest shows future"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-26T18:07:42Z
updatedAt: 2023-04-05T20:54:21Z
lastEditedAt:
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

**Current state:** the issue is closed; the current detailed output limits the range to the present moment and marks the days beyond it. That concerns displaying the calendar: the absence of future columns does not by itself guarantee that the suggestion timestamp will not move into the future once the random offset is added ([#193](issue-193.md)).
