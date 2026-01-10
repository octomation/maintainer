---
code:
id: I_kwDOE2M9Zc5hW_Y_
databaseId: 1633416767
number: 119
url: https://github.com/octomation/maintainer/issues/119
title: "github: contribution: invalid suggestion for specific date"
labels:
  - "type: bug"
  - "severity: critical"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2023-03-21T08:18:46Z
updatedAt: 2023-03-25T20:26:23Z
lastEditedAt:
closedAt: 2023-03-25T20:26:23Z
---

# github: contribution: invalid suggestion for specific date

Do not choose a day earlier than the explicitly supplied date, even when a profitable gap exists earlier in the same week. The user sets the lower bound of the search, and calendar alignment must not override it.

The original example:

```bash
maintainer github contribution suggest --delta 2022-02-12
# Obtained: 2022-02-06, 6 → 10
# Expected on the data in the report: 2022-02-16, 5 → 6
```

The invalid past part of the week is expected to be skipped and the next suitable day found. Verification has to account for a saturated Saturday and the move into the following week; the actual/target values refer to the chosen date specifically.

**Current state:** the issue is closed. The [suggest tests](../../internal/model/github/contribution/suggest_test.go) include the case `issue#119: max Saturday`, and the command bounds the search by the reference moment. The test uses its own stored data set, not the original 2022 calendar.

Related: [#84](issue-84.md), [dotfiles#543](https://github.com/kamilsk/dotfiles/issues/543).
