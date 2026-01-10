---
code:
id: I_kwDOE2M9Zc5OQKzo
databaseId: 1312861416
number: 68
url: https://github.com/octomation/maintainer/issues/68
title: "github: contribution: invalid suggestion for edge case with zero"
labels:
  - "scope: code"
  - "scope: test"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-21T08:15:21Z
updatedAt: 2026-09-25T05:13:19Z
lastEditedAt: 2026-09-25T05:13:19Z
closedAt: 2022-07-22T18:18:53Z
---

# github: contribution: invalid suggestion for edge case with zero

Take days with zero contributions into account when choosing a date. Skipping such a day defeats the main purpose of suggest — to find the unfilled parts of the calendar rather than to grow already high counts.

The original scenario:

```bash
maintainer github contribution suggest --delta 2021
```

On the data in the report the command chose 12 September: `9 → 12`. The expected choice was 11 September, a Saturday with no activity: `0 → 7`. The relative values `-312d` and `-313d` refer to the moment of the original run, not to today.

The fix criterion: a day with a zero count participates in the choice on equal terms with the rest, including the last day of the week, and actual/target match the chosen date.
