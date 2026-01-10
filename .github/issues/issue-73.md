---
code:
id: I_kwDOE2M9Zc5Ogsu8
databaseId: 1317194684
number: 73
url: https://github.com/octomation/maintainer/issues/73
title: "github: contribution: suggest command has regression with weeks arg"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-25T18:16:08Z
updatedAt: 2022-07-25T18:20:49Z
lastEditedAt:
closedAt: 2022-07-25T18:20:48Z
---

# github: contribution: suggest command has regression with weeks arg

Restore the size of the directed suggest window after the regression between v0.1.0-rc6 and v0.1.0-rc7. Fixing a calendar boundary must not change what the number of weeks means.

The original call:

```bash
maintainer github contribution suggest --delta 2022-05-01/+1
```

rc6 displayed two weeks and rc7 displayed one; in both outputs the wrong suggestion of 24 April remained. These are two independent symptoms: the width of the window and the choice of day.

For `/+1` the reference week plus one following week is expected, with Sunday in the right column and a date no earlier than 1 May. Verification has to compare identical data and identical input arguments, not just the visible number of columns.

**Current state:** the issue is closed. The direction contract is described in [#41](issue-41.md), and the original Sunday selection defect in [#72](issue-72.md).
