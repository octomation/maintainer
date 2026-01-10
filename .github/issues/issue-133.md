---
code:
id: I_kwDOE2M9Zc5jTtbY
databaseId: 1666111192
number: 133
url: https://github.com/octomation/maintainer/issues/133
title: "github: contribution: suggest for cmm"
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2023-04-13T10:02:01Z
updatedAt: 2026-09-25T05:10:16Z
lastEditedAt: 2026-09-25T05:10:16Z
closedAt:
---

# github: contribution: suggest for cmm

Do not crash when the author date of HEAD lies in the future relative to the current time. This happens when a Git wrapper is used again after a commit with a future timestamp.

The original scenario:

```text
git contrib … → creates a commit dated later than the present moment
git contrib … → recovered: assertion is not a true
                unexpected panic occurred
```

The expected behaviour is to limit the reference moment by the current time, as the original report proposed, or else to state clearly that there is no valid suggestion. An impossible range must not be constructed, and an empty date must not be handed to the next command.

Cases needed: a future time today, and a future calendar date. The detailed reproduction with clock times and a stack is [[issue-148]]; an excessive random offset is [[issue-193]].
