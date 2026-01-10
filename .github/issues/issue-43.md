---
code:
id: I_kwDOE2M9Zc5KuJto
databaseId: 1253612392
number: 43
url: https://github.com/octomation/maintainer/issues/43
title: "github: contribution: expand heat map by merging with neighbors"
labels: []
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-31T09:45:24Z
updatedAt: 2026-09-25T05:12:13Z
lastEditedAt: 2026-09-25T05:12:13Z
closedAt: 2022-05-31T14:21:48Z
---

# github: contribution: expand heat map by merging with neighbors

Show a continuous window of contributions across the calendar-year boundary. The user selects weeks around a date rather than a single year, so truncating at 31 December or 1 January makes the result incomplete.

The original examples:

```bash
maintainer github contribution lookup 2013-12-31/5
maintainer github contribution lookup 2014-01-01/5
```

Each call used to return only the part belonging to its own year and printed `?` where the neighbouring year's data was required. All requested weeks are expected, with the actual values from both years; the unavailable future stays a separate case. The original checklist named the two steps: drop the year trimming from the scope calculation, and let the heat map span several years.

Related: [[issue-77]], [[issue-284]].
