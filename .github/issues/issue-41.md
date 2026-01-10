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
updatedAt: 2026-09-25T05:12:09Z
lastEditedAt: 2026-09-25T05:12:09Z
closedAt: 2022-05-21T12:10:11Z
---

# github: contribution: lookup forward/backward

Let the user choose the direction of the contributions view: around a date, forward, or backward. A symmetric window is inconvenient when only the subsequent activity, or only the recent history, is of interest.

The argument forms:

```bash
maintainer github contribution lookup 2021-01-01/3
maintainer github contribution lookup 2021-01-01/+2
maintainer github contribution lookup now/-2
```

`/3` requests a centred window; `/+2` the reference week and the two following it; `/-2` the reference week and the two before it. The future boundary limits the available part of the calendar. Here the sign changes the meaning of the selection, not merely how the number is written.

The original idea also proposed separate `forward` / `backward` words and optional autodetection of the direction; adopting them needs a scope decision.

Related: [[issue-43]], [[issue-60]].

<!-- 2022-05-21T12:10Z https://github.com/octomation/maintainer/issues/41#issuecomment-1133609152
https://github.com/octomation/maintainer/commit/377624544e97f2afd367a45ea2e3f37682983031
-->
