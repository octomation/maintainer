---
code:
id: I_kwDOE2M9Zc5t5-gy
databaseId: 1843914802
number: 155
url: https://github.com/octomation/maintainer/issues/155
title: "github: contribution: simplify arguments format"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-08-09T19:54:39Z
updatedAt: 2026-09-25T05:10:37Z
lastEditedAt: 2026-09-25T05:10:37Z
closedAt:
---

# github: contribution: simplify arguments format

Fix the lookup crash on a short argument and make the period formats predictable. The original title speaks about simplifying the input, but the [screenshot](https://github.com/octomation/maintainer/assets/1165416/01fa06e5-7a9f-42e2-b18e-688956177309) records a concrete, reproducible bug:

```bash
maintainer github contribution lookup 2022
# recovered: assertion is not a true
# unexpected panic occurred
```

The expectation is no panic for a year, a month, a day, or empty input. The supported forms must have a clear meaning and examples in the help, and invalid input must produce an ordinary diagnostic error. The calendar-month contract is specified separately in [[issue-45]]; simply accepting `YYYY-MM` must not be mistaken for implementing it. The difference from [[issue-121]]: there the parsing of `2022/+10` was refused, whereas here the window calculation itself breaks when the suffix is absent.
