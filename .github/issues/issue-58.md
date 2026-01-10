---
code:
id: I_kwDOE2M9Zc5L4cO3
databaseId: 1273086903
number: 58
url: https://github.com/octomation/maintainer/issues/58
title: "github: contribution: add --auto flag to suggest command"
labels:
  - "scope: code"
  - "type: feature"
  - "impact: high"
  - "effort: medium"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-16T05:55:02Z
updatedAt: 2023-03-26T20:47:06Z
lastEditedAt: 2022-06-16T05:55:43Z
closedAt: 2023-03-26T20:47:05Z
---

# github: contribution: add --auto flag to suggest command

Take the date of the last commit automatically, so a Git wrapper does not have to read `git log` and pass a timestamp itself.

The original idea proposed `--auto`; the current equivalent is:

```bash
timestamp=$(maintainer github contribution suggest --short git/+2)
```

`git` explicitly selects the author date of HEAD in the available repository; an empty date uses the same fallback, and if the repository cannot be opened the current time is used. The user should get a date that respects the chosen history, without extra shell logic.

**Current state:** the issue is closed and there is no separate `--auto` flag. The original follow-up of adapting [dotfiles](https://github.com/kamilsk/dotfiles/blob/c7c6f9f73d99710081f5894614709abeadd439c9/bin/git_commit#L41) was left unchecked, and the local codebase cannot confirm its status. Working through `GIT_DIR` remains the separate defect [#189](issue-189.md), and a future HEAD date is [#133](issue-133.md).

<!-- 2022-07-23T19:32Z https://github.com/octomation/maintainer/issues/58#issuecomment-1193177422
add buffer: `maintainer github contribution suggest --short --auto --buffer=5m` - add 5m+- to the latest commit
-->
