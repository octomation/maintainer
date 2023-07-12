---
id: 58
database_id: 1273086903
node_id: I_kwDOE2M9Zc5L4cO3
status: closed
title: "github: contribution: add --auto flag to suggest command"
labels: [scope: code, type: feature, impact: high, effort: medium]
url: https://github.com/octomation/maintainer/issues/58
created_at: 2022-06-16T05:55:02Z
updated_at: 2023-03-26T20:47:06Z
---

# github: contribution: add --auto flag to suggest command

**Motivation:** it allows to simplify https://github.com/kamilsk/dotfiles/blob/c7c6f9f73d99710081f5894614709abeadd439c9/bin/git_commit#L41.

**Interface**

```bash
$ timestamp=$(maintainer github contribution suggest --short --auto)
# vs
# timestamp=$(maintainer github contribution suggest --short "$(git --no-pager log -1 --format="%as")")
```

**To do:**

- [ ] refactor `dotfiles`.
