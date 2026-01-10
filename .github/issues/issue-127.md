---
code:
id: I_kwDOE2M9Zc5it6ag
databaseId: 1656202912
number: 127
url: https://github.com/octomation/maintainer/issues/127
title: "github: contribution: resolve all related to the scope todos"
labels:
  - "scope: code"
  - "type: improvement"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-04-05T19:57:08Z
updatedAt: 2026-09-25T05:10:02Z
lastEditedAt: 2026-09-25T05:10:02Z
closedAt:
---

# github: contribution: resolve all related to the scope todos

Work through the remaining unfinished parts of the contribution commands in terms of their user-visible effect. The old list of line numbers goes stale quickly and does not explain what is supposed to get better.

Two directions:

- Align the diff output with the calendar and fix its edge cases; the user must read the sign, the date and the week of a change correctly ([[issue-70]]).
- Allow the working hours and the time zone of suggest to be configured, so the proposed time matches the user's routine.

A prototype of the configuration; the flag names are provisional:

```bash
maintainer github contribution suggest --hours=09:00-18:00 --timezone=Europe/Moscow git/+2
```

Related: the general refactoring [[issue-44]], the statistics behind `Stats: coming soon` [[issue-79]].
