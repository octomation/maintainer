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
updatedAt: 2023-04-05T20:27:04Z
lastEditedAt:
closedAt: 2023-04-05T20:27:04Z
---

# github: contribution: highlight suggested day

Highlight the suggested day in the calendar so the user can check the date and its surroundings visually — to verify correctness at a glance. A single suggestion line is not enough when the table holds several weeks with identical counts.

An example of the current rendering:

```text
Wednesday  3*  → the chosen day, which has three contributions
Thursday    *  → the chosen day, whose count is zero
```

In the detailed mode exactly the cell of the resulting date must be highlighted, and its count must not be lost. With `--short` the table is not printed and the machine-readable date stays available.

**Current state:** the issue is closed. [Suggest](../../internal/command/github/contribution/suggest.go) uses `*` rather than the decorative `★` of the original prototype. The highlight itself does not prove the date was chosen correctly; the table centring shift is described in [#124](issue-124.md) and today's defects in [#136](issue-136.md).
