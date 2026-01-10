---
code:
id: I_kwDOE2M9Zc5MZ38c
databaseId: 1281851164
number: 63
url: https://github.com/octomation/maintainer/issues/63
title: "github: contribution: highlight suggested day"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-23T06:00:34Z
updatedAt: 2026-09-25T05:13:07Z
lastEditedAt: 2026-09-25T05:13:07Z
closedAt: 2023-04-05T20:27:04Z
---

# github: contribution: highlight suggested day

Highlight the suggested day in the calendar so the user can check the date and its surroundings visually — to verify correctness at a glance. A single suggestion line is not enough when the table holds several weeks with identical counts.

The intended rendering:

```text
Wednesday  3*  → the chosen day, which has three contributions
Thursday    *  → the chosen day, whose count is zero
```

In the detailed mode exactly the cell of the resulting date must be highlighted, and its count must not be lost. With `--short` the table is not printed and the machine-readable date stays available.

Related: [[issue-124]], [[issue-136]].
