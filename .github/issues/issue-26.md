---
code:
id: I_kwDOE2M9Zc5JOiQP
databaseId: 1228547087
number: 26
url: https://github.com/octomation/maintainer/issues/26
title: "implement github contribution lookup command"
labels:
  - "scope: docs"
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-07T06:48:18Z
updatedAt: 2022-06-15T10:15:47Z
lastEditedAt:
closedAt: 2022-05-10T07:16:56Z
---

# implement github contribution lookup command

Show GitHub contributions for a chosen interval in the terminal. The user must see the activity by day and by week in order to explore the history without switching to a browser.

The working example:

```bash
maintainer github contribution lookup 2021-02-01/3
```

A window around the given date is requested: the week containing it plus the neighbouring weeks. Table rows correspond to the days from Sunday to Saturday; the dates of the last column are printed on the right. A `-` marks a zero count, and `?` is used beyond the available end of the range.

**Current state:** the closed task is implemented in [lookup](../../internal/command/github/contribution/lookup.go). The original `--weeks=3` was replaced by `/3`. `now/3` anchors the window to the present moment; with an empty date the HEAD date of an available Git repository is used, otherwise the current time. The known week-numbering defect is described separately in [#284](issue-284.md).
