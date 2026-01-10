---
code:
id: I_kwDOE2M9Zc5L1CKb
databaseId: 1272193691
number: 57
url: https://github.com/octomation/maintainer/issues/57
title: "github: contribution: refactor logic of suggest command"
labels:
  - "scope: docs"
  - "scope: code"
milestone: "[[milestone-1]]"
state: CLOSED
stateReason: COMPLETED
createdAt: 2022-06-15T12:54:26Z
updatedAt: 2026-09-25T05:12:50Z
lastEditedAt: 2026-09-25T05:12:50Z
closedAt: 2022-06-15T14:21:11Z
---

# github: contribution: refactor logic of suggest command

Simplify how suggest connects to personal Git commands: choosing a valid date and presenting it should be maintainer's job, so the logic is not duplicated in shell scripts.

The working scenario:

```bash
maintainer github contribution suggest --short --delta git/+2
```

The stdout result is a relative date suitable for further processing; a normal run also shows the calendar. The Git context sets the lower bound of the search, so a new suggestion does not send the user arbitrarily back in history.

Related: [[issue-56]], [[issue-127]], the [historical git_commit](https://github.com/kamilsk/dotfiles/blob/d19edea4e4a08a325acece3ddb41f5c5312a0133/bin/git_commit#L40-L43), [dotfiles#320](https://github.com/kamilsk/dotfiles/issues/320).
