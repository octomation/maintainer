---
code: MAIN-73
id: I_kwDOE2M9Zc5Ogsu8
databaseId: 1317194684
number: 73
url: https://github.com/octomation/maintainer/issues/73
title: "github: contribution: suggest command has regression with weeks arg"
issueType: null
labels:
  - "scope: code"
assignees:
  - kamilsk
milestone: "[[MAIN-M1]]"
parent: null
issueFields: {}
state: CLOSED
stateReason: COMPLETED
duplicateOf: null
createdAt: 2022-07-25T18:16:08Z
updatedAt: 2026-09-25T05:13:33Z
lastEditedAt: 2026-09-25T05:13:33Z
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

Related: the direction contract [[MAIN-41]], the original Sunday selection defect [[MAIN-72]].
