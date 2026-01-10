---
code:
id: I_kwDOE2M9Zc5it3yU
databaseId: 1656192148
number: 126
url: https://github.com/octomation/maintainer/issues/126
title: "github: contribution: incorrect shows stats for future"
labels:
  - "help wanted"
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: minor"
  - "impact: low"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-04-05T19:49:00Z
updatedAt: 2026-09-25T05:09:59Z
lastEditedAt: 2026-09-25T05:09:59Z
closedAt: 2023-04-05T20:02:21Z
---

# github: contribution: incorrect shows stats for future

Do not present the future days of the last week as days with zero activity. In the lookup table that creates a false impression of gaps that are already available for analysis.

The original example:

```bash
maintainer github contribution lookup /-2
# In week #14, Thursday, Friday and Saturday show "-" instead of "?".
```

Expected: a past day without contributions is `-`; a day beyond the time boundary is `?`. The counts of existing contributions are preserved. A check with a fixed moment in the middle of a week is needed, not only with a completed year.

Related: the analogous case in suggest [[issue-75]], time-zone differences [[issue-65]].
