---
code:
id: I_kwDOE2M9Zc57TNqj
databaseId: 2068634275
number: 189
url: https://github.com/octomation/maintainer/issues/189
title: "github: contribution: suggest doesn't work properly in headless mode"
labels:
  - "scope: code"
  - "type: bug"
  - "severity: major"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: OPEN
stateReason:
createdAt: 2024-01-06T13:33:42Z
updatedAt: 2024-01-06T13:33:43Z
lastEditedAt:
closedAt:
---

# github: contribution: suggest doesn't work properly in headless mode

Use the same Git repository whether the date is chosen from inside its directory or with `GIT_DIR` set explicitly. This matters for scripts that work with a repository without entering its working tree.

The scenarios to compare:

```bash
# From the working tree:
maintainer github contribution suggest git/1
# From another directory:
GIT_DIR=/path/to/repo/.git maintainer github contribution suggest git/1
```

In the original report the first form chose 29 December 2023 (`45 → 50`), while the second chose the current day, 6 January 2024 (`13 → 15`). The likely effect is that the Git anchor is lost and the current time is used as a fallback.

**Code context:** HEAD detection looks for `.git` relative to the working directory; there is no explicit handling of `GIT_DIR`. The fix criterion is the same repository and the same lower bound under both ways of addressing it, with no silent switch to a different checkout. The exact time may still differ because of jitter. `git --git-dir=… log` is useful for confirming the expected HEAD; maintainer itself has no `--git-dir` flag today.
