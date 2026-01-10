---
code:
id: I_kwDOE2M9Zc5LHwXR
databaseId: 1260324305
number: 46
url: https://github.com/octomation/maintainer/issues/46
title: "github: contribution: show datetime on the right side of lookup"
labels:
  - "scope: docs"
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-03T20:06:15Z
updatedAt: 2023-04-05T20:41:10Z
lastEditedAt:
closedAt: 2023-04-05T20:41:10Z
---

# github: contribution: show datetime on the right side of lookup

Add a column of dates to the right of the calendar so that a weekday row can be matched quickly to a concrete date. Week numbers without dates force the user to recount the calendar by hand.

An example of the expected fragment:

```text
Day / Week   #10   Date
Sunday        10   Mar 7
Monday        10   Mar 8
...
Saturday      10   Mar 13
```

For several weeks the column refers to the last, rightmost column of the calendar — the "last" dates of the requested window. Crossing a month or a year must not break the correspondence between a row and its date.

**Current state:** the task is closed; the `Date` column is produced by the shared [calendar view](../../internal/command/github/contribution/helper.go) and is used by lookup and by the detailed suggest. Fixing the numbering of the weeks themselves is handled separately in [#284](issue-284.md).
