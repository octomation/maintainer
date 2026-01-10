---
code:
id: I_kwDOE2M9Zc5h2Qws
databaseId: 1641614380
number: 122
url: https://github.com/octomation/maintainer/issues/122
title: "github: contribution: suggest works incorrectly after refactoring"
labels:
  - "scope: code"
  - "scope: test"
  - "type: bug"
  - "severity: major"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-03-27T08:11:43Z
updatedAt: 2023-04-05T19:42:34Z
lastEditedAt: 2023-04-05T19:41:45Z
closedAt: 2023-04-05T19:42:34Z
---

# github: contribution: suggest works incorrectly after refactoring

Reconcile the suggest table with the counts of the chosen day after the refactoring. In the original output the calendar showed existing activity while the suggestion line declared the same day empty; on top of that, Sunday's data was lost.

The example from [v0.1.0-rc10](https://github.com/octomation/maintainer/releases/tag/v0.1.0-rc10):

```text
HEAD anchor: 2022-07-26 08:57:56 +0300
Suggestion:  2022-07-26T09:39:14+03:00, 0 → 10
Table:       the corresponding Tuesday already has contributions.
```

The date, the highlighted cell and the actual value in the explanation are expected to refer to one calendar day, and Sunday must not be lost when the weeks are aligned. A full commit timestamp must not act as a different key from the calendar date.

**Current state:** the issue is closed, and both original sub-tasks — the empty Sunday and the `0 → 10` count — were marked done. The current model normalizes dates to UTC; a new case of a wrong count for today is handled in [#136](issue-136.md).
