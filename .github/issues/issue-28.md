---
code:
id: I_kwDOE2M9Zc5JOihI
databaseId: 1228548168
number: 28
url: https://github.com/octomation/maintainer/issues/28
title: "implement github contribution diff command"
labels:
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-05-07T06:54:20Z
updatedAt: 2022-06-15T10:15:48Z
lastEditedAt:
closedAt: 2022-06-03T18:06:34Z
---

# implement github contribution diff command

Compare two states of the contributions calendar and show the per-day changes. This makes the effect of work since a snapshot visible and reveals when GitHub recounts history.

The working forms:

```bash
maintainer github contribution diff before.json after.json
maintainer github contribution diff before.json 2021
```

The first argument is the base, the second is the new state; a year means loading data from GitHub, a path means reading a stored snapshot. The expected meaning of the output: 3 became 5 → `+2`; 5 became 3 → `-2`; no changes → an explicit message that there is no diff.

**Current state:** the command exists and the issue is closed. The original `--src`/`--dst`, and the later `--base`/`--head`, were replaced by positional arguments. The correctness of the sign and the remaining presentation problems are handled in the open [#70](issue-70.md). Snapshots are produced by the command from [#27](issue-27.md).
