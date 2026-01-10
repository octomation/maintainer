---
code:
id: I_kwDOE2M9Zc5NfBpZ
databaseId: 1299978841
number: 67
url: https://github.com/octomation/maintainer/issues/67
title: "github: contribution: suggest doesn't work on last week with Sunday"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-07-10T17:51:04Z
updatedAt: 2022-07-10T19:11:06Z
lastEditedAt:
closedAt: 2022-07-10T19:11:06Z
---

# github: contribution: suggest doesn't work on last week with Sunday

Fix suggest at the Sunday boundary of the available calendar: instead of a sensible date the command returned a zero date and an enormous negative delta. Such output must not be passed into a Git wrapper.

The original reproduction:

```bash
maintainer github contribution suggest --delta 2022-07-10
# Suggestion is 0001-01-01: -106751d, 0 -> 0
```

A valid day from the chosen period is expected, Sunday 10 July included, or else a clear message that there is no candidate. `0001-01-01` is not a user-facing suggestion, and the table must not jump to unrelated winter weeks.

**Current state:** the issue is closed; the current command checks for a zero date and returns `nothing to suggest`. That guards the output, while the correctness of the search at the boundaries needs verifying on its own. The missing Sunday in lookup is [#66](issue-66.md), and a future HEAD date is [#148](issue-148.md).

<!-- 2022-07-10T19:11Z https://github.com/octomation/maintainer/issues/67#issuecomment-1179782513
https://github.com/octomation/maintainer/commit/294366c4b28fb63f5c20fe3c6bdbd44ecfd559d2
-->
