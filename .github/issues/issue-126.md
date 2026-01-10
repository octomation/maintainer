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
updatedAt: 2023-04-05T20:02:21Z
lastEditedAt:
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

**Current state:** the issue is closed; the current table respects the end of the available range. The analogous case in suggest is [#75](issue-75.md), and time-zone differences are [#65](issue-65.md).
