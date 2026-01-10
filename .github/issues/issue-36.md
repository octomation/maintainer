---
code:
id: I_kwDOE2M9Zc5Jj_J6
databaseId: 1234170490
number: 36
url: https://github.com/octomation/maintainer/issues/36
title: "command: github contribution lookup doesn't work well with now()"
labels:
  - "scope: code"
  - "scope: test"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-12T15:38:44Z
updatedAt: 2026-09-25T05:11:55Z
lastEditedAt: 2026-09-25T05:11:55Z
closedAt: 2022-05-12T20:47:39Z
---

# command: github contribution lookup doesn't work well with now()

Fix viewing the calendar anchored to the current moment: in the original report empty cells were printed instead of the existing contributions. As a result the user saw a false picture of no activity.

The historical reproduction from May 2022:

```bash
maintainer github contribution lookup /2
# Weeks #17–#19: nearly every cell shows "-", although the GitHub calendar is populated.
```

The actual counts up to the current date are expected to be loaded and displayed, and future days must be distinguishable from days without activity. The [original expected view](https://user-images.githubusercontent.com/1165416/168114309-beb71f1a-c88f-4221-8cd8-1e56c2ab4453.png) is kept as evidence of the report.

A reproduction must fix the moment of execution and pass an explicit `now/3`, since an empty date may resolve from the Git history; re-running the old `/2` at a different time does not reproduce the historical output.

Related: [[issue-38]], [[issue-65]].
