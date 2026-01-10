---
code:
id: I_kwDOE2M9Zc5PUv6h
databaseId: 1330839201
number: 79
url: https://github.com/octomation/maintainer/issues/79
title: "github: contribution: extend suggest command by stats"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: medium"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-08-06T20:09:30Z
updatedAt: 2023-03-25T20:55:37Z
lastEditedAt:
closedAt:
---

# github: contribution: extend suggest command by stats

Show how much work is left to reach the target activity, alongside the suggested date. Today the actual/target of a single day is visible, but the size of the unfilled part of the week, month or year is not — the user cannot tell how many contributions they still have to make.

A prototype of the summary:

```text
Suggestion: the chosen day, 2 → 5
To target: 3 contributions this week, 15 this month, 558 this year
```

The numbers are illustrative. A single deficit formula and its calculation period have to be defined: count only the available days, do not treat the future as zeros, and explain the influence of `--target` and of future presets. Total contributions must not be equated with a number of commits, especially in the presence of mirrors ([#129](issue-129.md)).

**At present** the table prints `Stats: coming soon`; the statistics are not implemented. They should be useful in the detailed mode and must not pollute the date on stdout under `--short`. Related: the targets from [#128](issue-128.md), the activity distribution [#29](issue-29.md).
