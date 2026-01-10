---
code:
id: I_kwDOE2M9Zc5Ne8OC
databaseId: 1299956610
number: 66
url: https://github.com/octomation/maintainer/issues/66
title: "github: contribution: lookup doesn't show last week on Sunday"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-10T16:05:28Z
updatedAt: 2026-09-25T05:13:14Z
lastEditedAt: 2026-09-25T05:13:14Z
closedAt: 2022-07-10T19:10:52Z
---

# github: contribution: lookup doesn't show last week on Sunday

Do not lose the week that has just begun when lookup runs on a Sunday. The calendar uses Sunday-to-Saturday weeks, so the last day of a range can be the start of a new column.

The original example:

```bash
maintainer github contribution lookup /-2
# The range is announced as ending 2022-07-10, but there is no row for 10 July.
```

A new week column with Sunday 2022-07-10 is expected; when checked on that day, the following days must not look like past days without contributions. The range and the table must describe the same dates.

Related: suggest with a zero date [[issue-67]], panics on Sunday [[issue-123]], the wrong new-year week number [[issue-284]].

<!-- 2022-07-10T19:10Z https://github.com/octomation/maintainer/issues/66#issuecomment-1179782485
https://github.com/octomation/maintainer/commit/a33aefc1af8968bcc46bdb409164fab4117db3a8
-->
