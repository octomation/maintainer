---
id: 29
database_id: 1228627172
node_id: I_kwDOE2M9Zc5JO1zk
status: closed
title: "implement github contribution histogram command"
labels: ["scope: code"]
url: https://github.com/octomation/maintainer/issues/29
created_at: 2022-05-07T13:22:14Z
updated_at: 2022-06-15T10:15:48Z
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

**Current state:** the closed task is implemented. One `#` stands for one day, so a whole-year selection can be far too wide for a terminal; that is the separate improvement [#130](issue-130.md).
