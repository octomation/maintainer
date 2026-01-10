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
updatedAt: 2026-09-25T05:11:25Z
lastEditedAt: 2026-09-25T05:11:25Z
closedAt: 2022-05-10T07:16:56Z
---

# implement github contribution lookup command

Show GitHub contributions for a chosen interval in the terminal. The user must see the activity by day and by week in order to explore the history without switching to a browser.

The working example:

```bash
maintainer github contribution lookup 2021-02-01/3
```

A window around the given date is requested: the week containing it plus the neighbouring weeks. Table rows correspond to the days from Sunday to Saturday; the dates of the last column are printed on the right. A `-` marks a zero count, and `?` is used beyond the available end of the range.

Related: [[issue-284]].
