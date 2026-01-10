---
code:
id: I_kwDOE2M9Zc5JO1zk
databaseId: 1228627172
number: 29
url: https://github.com/octomation/maintainer/issues/29
title: "implement github contribution histogram command"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-07T13:22:14Z
updatedAt: 2026-09-25T05:11:35Z
lastEditedAt: 2026-09-25T05:11:35Z
closedAt: 2022-05-21T15:56:08Z
---

# implement github contribution histogram command

Show the distribution of daily activity over a year, a month or a week. Unlike the calendar, a histogram answers "how many days had this many contributions" and helps to pick a reasonable target for a suggestion.

The working examples:

```bash
maintainer github contribution histogram 2021
maintainer github contribution histogram 2021-02
maintainer github contribution histogram 2021-02-01
```

How to read the result:

```text
  1 ##      → two days with one contribution
  2 #####   → five days with two contributions
```

With no argument the current week is shown. Zero values can be included with the `--with-zero` flag.

Related: [[issue-130]].
