---
code:
id: I_kwDOE2M9Zc5LHwB2
databaseId: 1260322934
number: 45
url: https://github.com/octomation/maintainer/issues/45
title: "github: contribution: extend support lookup for a month"
labels:
  - "scope: docs"
  - "scope: code"
  - "scope: test"
  - "type: feature"
  - "impact: medium"
  - "effort: easy"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2022-06-03T20:04:59Z
updatedAt: 2023-04-06T11:18:18Z
lastEditedAt:
closedAt:
---

# github: contribution: extend support lookup for a month

Support viewing contributions for a calendar month without counting weeks by hand. The user expects `2013-11` to mean November, including all of its days, rather than an arbitrary window around the first of the month. It is useful for exploration.

The proposed scenario:

```bash
maintainer github contribution lookup 2013-11
# The calendar for 1–30 November 2013.
```

How neighbouring days in partial weeks are shown has to be decided, the month boundaries have to be visible, and directed weekly requests must still work for arguments carrying `/…`.

**At present** the shared parser accepts `YYYY-MM` but passes the first day of the month into the weekly calculation; there is no separate monthly mode for lookup. Moreover, the suffix-less form runs into the defect [#155](issue-155.md). In `histogram` a calendar month already has its own meaning ([#29](issue-29.md)); that is a useful consistency reference, not proof that lookup implements it.
