---
code:
id: I_kwDOE2M9Zc5KJDqA
databaseId: 1243888256
number: 41
url: https://github.com/octomation/maintainer/issues/41
title: "github: contribution: lookup forward/backward"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-21T06:31:57Z
updatedAt: 2022-06-15T10:15:50Z
lastEditedAt:
closedAt: 2022-05-21T12:10:11Z
---

# github: contribution: lookup forward/backward

Let the user choose the direction of the contributions view: around a date, forward, or backward. A symmetric window is inconvenient when only the subsequent activity, or only the recent history, is of interest.

The current argument forms:

```bash
maintainer github contribution lookup 2021-01-01/3
maintainer github contribution lookup 2021-01-01/+2
maintainer github contribution lookup now/-2
```

`/3` requests a centred window; `/+2` the reference week and the two following it; `/-2` the reference week and the two before it. The future boundary limits the available part of the calendar. Here the sign changes the meaning of the selection, not merely how the number is written.

**Current state:** the task is closed and the direction is implemented in [date parsing](../../internal/command/github/contribution/helper.go) and the range calculation. The variant with separate `forward` / `backward` words from the original idea was not implemented, and neither was the optional autodetection of direction it mentioned. The year boundary is handled in [#43](issue-43.md); carrying the same syntax into suggest is [#60](issue-60.md).

<!-- 2022-05-21T12:10Z https://github.com/octomation/maintainer/issues/41#issuecomment-1133609152
https://github.com/octomation/maintainer/commit/377624544e97f2afd367a45ea2e3f37682983031
-->
